package aftersales

// Trade-ins: an online estimate from today's new price, appraisal at a store (IMEI checked by
// the risk engine's trade-in checkpoint), an offer the customer accepts or declines within 7
// days, then payout as store credit (at once) or by bank transfer (finance pays and records it).
// An accepted device goes to the workshop for refurbishment before it can be sold.
// Plan: payments-finance.md §11; docs/plan.md 2026-10-23.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/ledger"
	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/risk"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/money"
	"github.com/dprince-03/techshop/backend/pkg/validate"
)

// Ledger accounts for trade-ins (additions to the payments-finance.md §4.1 chart, logged in
// docs/plan.md). The ledger doesn't value stock, so a bought-in device is a purchase expense.
const (
	expenseTradeIns  = "expense:trade_in_purchases"
	tradeInsPayable  = "liability:trade_in_payable"
	tradeInFinancial = "trade_in"
)

// TradeInInput is a customer's trade-in request.
type TradeInInput struct {
	DeviceType string     `json:"deviceType" enum:"phone,laptop,tablet,console,smartwatch"`
	VariantID  *uuid.UUID `json:"variantId,omitempty" doc:"The matching catalogue model, for an instant estimate"`
	Brand      string     `json:"brand,omitempty" maxLength:"60" doc:"When the model isn't in the list"`
	Model      string     `json:"model,omitempty" maxLength:"80"`
	Storage    string     `json:"storage,omitempty" maxLength:"20"`
	Condition  string     `json:"condition" enum:"like_new,good,fair,faulty"`
	IMEI       string     `json:"imei,omitempty" maxLength:"20" doc:"Phones and tablets: dial *#06#"`
}

// CreateTradeIn records a trade-in. A catalogue model gets its estimate at once; anything else
// waits for staff to price it.
func (s *Service) CreateTradeIn(ctx context.Context, userID uuid.UUID, in TradeInInput) (TradeIn, error) {
	imei := strings.ReplaceAll(strings.TrimSpace(in.IMEI), " ", "")
	if imei != "" && !validate.IMEI(imei) {
		return TradeIn{}, httpx.Invalid("imei_invalid", "That IMEI doesn't look right. Dial *#06# and copy all 15 digits.")
	}
	brand, model := strings.TrimSpace(in.Brand), strings.TrimSpace(in.Model)
	status, estimate := "submitted", (*int64)(nil)
	if in.VariantID != nil {
		v, err := s.d.Q.TradeInVariantPrice(ctx, *in.VariantID)
		if err != nil {
			return TradeIn{}, httpx.Invalid("unknown_model", "Choose a model from the list, or type it in.")
		}
		if !inCategories(in.DeviceType, v.CategorySlug, v.ParentCategorySlug) {
			return TradeIn{}, httpx.Invalid("wrong_type", v.ProductName+" isn't a "+in.DeviceType+".")
		}
		brand, model = v.Brand, v.ProductName+" "+v.VariantName
		if e := tradeInEstimate(v.NewPriceKobo, in.Condition); e > 0 {
			status, estimate = "estimated", &e
		}
	}
	if brand == "" || model == "" {
		return TradeIn{}, httpx.Invalid("model_required", "Tell us the brand and model.")
	}
	var ref string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var imeiP *string
		if imei != "" {
			imeiP = &imei
		}
		t, err := tx.Q.TradeInCreate(ctx, store.TradeInCreateParams{UserID: userID, DeviceType: in.DeviceType, Brand: brand, Model: model, Storage: optional(in.Storage),
			DeclaredCondition: in.Condition, Imei: imeiP, VariantID: in.VariantID, EstimateKobo: estimate, Status: status})
		if err != nil {
			return httpx.DB(err, "trade-in")
		}
		ref = t.Reference
		return tx.Audit("trade_in.created", "trade_in", t.ID.String(), map[string]any{"estimate": estimate})
	})
	if err != nil {
		return TradeIn{}, err
	}
	return s.tradeInView(ctx, ref)
}

func inCategories(deviceType, cat, parent string) bool {
	cats := tradeInCategories[deviceType]
	return slices.Contains(cats, cat) || slices.Contains(cats, parent)
}

// EstimateTradeIn prices a trade-in staff couldn't price automatically.
func (s *Service) EstimateTradeIn(ctx context.Context, ref string, amount int64, variantID *uuid.UUID) (TradeIn, error) {
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.Status != "submitted" && t.Status != "estimated" {
			return httpx.Conflict("bad_transition", "This trade-in is past the estimate stage.")
		}
		if _, err := tx.Q.TradeInSetEstimate(ctx, store.TradeInSetEstimateParams{ID: t.ID, EstimateKobo: &amount, VariantID: variantID}); err != nil {
			return httpx.DB(err, "model")
		}
		return s.tellTradeIn(ctx, tx, t, "estimated at "+money.Format(amount)+". Bring it to a TechShop store for inspection")
	})
}

// CancelTradeIn withdraws a trade-in before the device is handed in.
func (s *Service) CancelTradeIn(ctx context.Context, userID uuid.UUID, ref string) (TradeIn, error) {
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.UserID != userID {
			return httpx.NotFound("Trade-in")
		}
		if t.Status != "submitted" && t.Status != "estimated" {
			return httpx.Conflict("bad_transition", "The device is already with us. Decline the offer instead.")
		}
		_, err := tx.Q.TradeInSetStatus(ctx, store.TradeInSetStatusParams{ID: t.ID, Status: "cancelled"})
		return err
	})
}

// ReceiveInput checks a device in at a store.
type ReceiveInput struct {
	StoreID uuid.UUID `json:"storeId"`
	IMEI    string    `json:"imei,omitempty" maxLength:"20" doc:"Required for phones and tablets"`
	Serial  string    `json:"serial,omitempty" maxLength:"40" doc:"Laptops, consoles, watches"`
}

// ReceiveTradeIn takes the device at a store. The trade-in checkpoint checks the IMEI (check
// digit, stolen list) and the person's recent trade-ins; a block stops it and opens a case.
func (s *Service) ReceiveTradeIn(ctx context.Context, ref string, in ReceiveInput) (TradeIn, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	blocked := false
	imei := strings.ReplaceAll(strings.TrimSpace(in.IMEI), " ", "")
	serial := strings.TrimSpace(in.Serial)
	_, err := s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.Status != "estimated" {
			if t.Status == "submitted" {
				return httpx.Conflict("needs_estimate", "Give it an estimate first.")
			}
			return httpx.Conflict("bad_transition", "This trade-in is already "+humanise(t.Status)+".")
		}
		needsIMEI := t.DeviceType == "phone" || t.DeviceType == "tablet"
		if needsIMEI && imei == "" {
			return httpx.Invalid("imei_required", "Scan the IMEI (*#06#) of the phone in hand.")
		}
		if !needsIMEI && serial == "" {
			return httpx.Invalid("serial_required", "Record the serial number on the device.")
		}
		if t.Imei != nil && imei != "" && *t.Imei != imei {
			return httpx.Invalid("imei_mismatch", "This isn't the phone in the request (the IMEI differs). Check with the customer.")
		}
		f, err := s.tradeInFeatures(ctx, tx, t.UserID, imei)
		if err != nil {
			return err
		}
		dec, err := s.risk.Evaluate(ctx, tx, risk.Check{Checkpoint: "trade_in", SubjectType: "trade_in", SubjectID: t.ID, UserID: &t.UserID, Features: f,
			MoneyKobo: deref(t.EstimateKobo), CaseKind: "trade_in", CaseTitle: "Trade-in " + t.Reference})
		if err != nil {
			return err
		}
		if dec.Outcome == "block" {
			blocked = true // commit the block and the case, then refuse below
			note := "Blocked by the risk check (" + dec.CaseRef + ")."
			_, err := tx.Q.TradeInBlock(ctx, store.TradeInBlockParams{ID: t.ID, InspectionNotes: &note, Imei: optional(imei), Serial: optional(serial)})
			return err
		}
		if _, err := tx.Q.TradeInReceive(ctx, store.TradeInReceiveParams{ID: t.ID, StoreID: &in.StoreID, ReceivedBy: &actor, Imei: optional(imei), Serial: optional(serial)}); err != nil {
			return httpx.DB(err, "store")
		}
		if dec.Held() || dec.Outcome == "step_up" {
			return tx.Audit("trade_in.flagged", "trade_in", t.ID.String(), map[string]any{"score": dec.Score, "case": dec.CaseRef})
		}
		return nil
	})
	if err == nil && blocked {
		return TradeIn{}, httpx.Conflict("device_blocked", "We can't accept this device (it failed the IMEI or stolen-device check). A risk case was opened; keep the device and call your manager.")
	}
	if err != nil {
		return TradeIn{}, err
	}
	return s.tradeInView(ctx, ref)
}

func (s *Service) tradeInFeatures(ctx context.Context, tx *uow.Tx, userID uuid.UUID, imei string) (risk.Features, error) {
	age, err := tx.Q.TradeInAccountAgeDays(ctx, userID)
	if err != nil {
		return nil, err
	}
	recent, err := tx.Q.TradeInRecent(ctx, userID)
	if err != nil {
		return nil, err
	}
	f := risk.Features{"account_age_days": age, "trade_ins_7d": float64(recent)}
	if imei != "" {
		st, err := risk.CheckIMEI(ctx, tx.Q, imei)
		if err != nil {
			return nil, err
		}
		f["imei_luhn_ok"], f["imei_blocked"] = b2f(st.LuhnOK), b2f(st.Blocked)
	}
	return f, nil
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func deref(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// InspectTradeInInput is the appraisal.
type InspectTradeInInput struct {
	Condition string     `json:"condition" enum:"like_new,good,fair,faulty"`
	VariantID *uuid.UUID `json:"variantId,omitempty" doc:"The catalogue model, if the request didn't have one (needed to stock it)"`
	Offer     *int64     `json:"offer,omitempty" minimum:"0" doc:"Kobo. Leave out to price from the catalogue and the grade"`
	Notes     string     `json:"notes,omitempty" maxLength:"1000" doc:"What was checked: screen, battery, iCloud/Google lock, accessories"`
}

// InspectTradeIn grades the device and sends the offer (valid 7 days).
func (s *Service) InspectTradeIn(ctx context.Context, ref string, in InspectTradeInInput) (TradeIn, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.Status != "awaiting_inspection" {
			return httpx.Conflict("bad_transition", "Receive the device before inspecting it.")
		}
		variant := t.VariantID
		if in.VariantID != nil {
			variant = in.VariantID
		}
		if variant == nil {
			return httpx.Invalid("variant_required", "Match the device to a catalogue model so it can be stocked after refurbishment.")
		}
		offer := in.Offer
		if offer == nil {
			v, err := tx.Q.TradeInVariantPrice(ctx, *variant)
			if err != nil {
				return httpx.Invalid("unknown_model", "That catalogue model doesn't exist.")
			}
			o := tradeInEstimate(v.NewPriceKobo, in.Condition)
			if o == 0 {
				return httpx.Invalid("offer_required", "There's no live price for this model; enter the offer.")
			}
			offer = &o
		}
		expires := time.Now().Add(tradeInOfferWindow)
		if _, err := tx.Q.TradeInOffer(ctx, store.TradeInOfferParams{ID: t.ID, InspectedCondition: &in.Condition, OfferKobo: offer, InspectionNotes: optional(in.Notes),
			InspectedBy: &actor, OfferExpiresAt: &expires, VariantID: in.VariantID}); err != nil {
			return httpx.DB(err, "model")
		}
		return s.tellTradeIn(ctx, tx, t, "our offer is "+money.Format(*offer)+". Accept or decline within 7 days")
	})
}

// AcceptInput is the customer's yes, with how to be paid.
type AcceptInput struct {
	PayoutMethod  string `json:"payoutMethod" enum:"store_credit,bank"`
	BankCode      string `json:"bankCode,omitempty" maxLength:"10"`
	AccountNumber string `json:"accountNumber,omitempty" pattern:"^[0-9]{10}$"`
	AccountName   string `json:"accountName,omitempty" maxLength:"120" doc:"As the bank has it; checked with a name enquiry"`
}

// AcceptTradeIn takes the offer. Store credit is paid at once; a bank payout is checked with the
// bank and queued for finance. Either way the device goes to the workshop for refurbishment.
func (s *Service) AcceptTradeIn(ctx context.Context, userID uuid.UUID, ref string, in AcceptInput) (TradeIn, error) {
	var holder string
	if in.PayoutMethod == "bank" {
		if in.BankCode == "" || in.AccountNumber == "" {
			return TradeIn{}, httpx.Invalid("bank_required", "Choose your bank and enter the 10-digit account number.")
		}
		name, err := s.pay.ResolveAccount(ctx, in.BankCode, in.AccountNumber, in.AccountName)
		if err != nil {
			return TradeIn{}, err
		}
		holder = name // name enquiry is a network call, so it happens before the transaction
	}
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.UserID != userID {
			return httpx.NotFound("Trade-in")
		}
		if t.Status != "offer_sent" {
			return httpx.Conflict("bad_transition", "There's no open offer on this trade-in.")
		}
		if t.OfferExpiresAt != nil && time.Now().After(*t.OfferExpiresAt) {
			return httpx.Conflict("offer_expired", "This offer has expired. Ask the store for a new one.")
		}
		p := store.TradeInDecideParams{ID: t.ID, Status: "accepted", PayoutMethod: &in.PayoutMethod}
		if in.PayoutMethod == "bank" {
			sealed, err := s.d.Sealer.Seal([]byte(in.AccountNumber), "aftersales.trade_ins:"+t.ID.String())
			if err != nil {
				return err
			}
			last4 := in.AccountNumber[6:]
			p.BankCode, p.AccountNumberEncrypted, p.AccountNumberLast4, p.AccountName = &in.BankCode, sealed, &last4, &holder
		}
		t, err := tx.Q.TradeInDecide(ctx, p)
		if err != nil {
			return httpx.DB(err, "trade-in")
		}
		if err := s.stockForRefurb(ctx, tx, t); err != nil {
			return err
		}
		offer := *t.OfferKobo
		if in.PayoutMethod == "store_credit" {
			if _, _, err := s.ledger.Post(ctx, tx, ledger.Journal{Key: "trade_in:" + t.ID.String() + ":paid", Kind: tradeInFinancial, RefType: "trade_in", RefID: t.ID,
				Memo: "Trade-in " + t.Reference + " as store credit", Lines: []ledger.Line{{Code: expenseTradeIns, Amount: offer}, {Code: ledger.StoreCredit(t.UserID), Amount: -offer}}}); err != nil {
				return err
			}
			ref := "store-credit"
			if _, err := tx.Q.TradeInPaid(ctx, store.TradeInPaidParams{ID: t.ID, PayoutReference: &ref}); err != nil {
				return err
			}
			return s.tellTradeIn(ctx, tx, t, money.Format(offer)+" added to your store credit")
		}
		if _, _, err := s.ledger.Post(ctx, tx, ledger.Journal{Key: "trade_in:" + t.ID.String() + ":accepted", Kind: tradeInFinancial, RefType: "trade_in", RefID: t.ID,
			Memo: "Trade-in " + t.Reference + " owed by bank transfer", Lines: []ledger.Line{{Code: expenseTradeIns, Amount: offer}, {Code: tradeInsPayable, Amount: -offer}}}); err != nil {
			return err
		}
		return s.tellTradeIn(ctx, tx, t, "accepted. We'll send "+money.Format(offer)+" to your account ending "+*p.AccountNumberLast4+" within 2 working days")
	})
}

// stockForRefurb records the accepted device as a unit awaiting refurbishment, with a workshop
// job; it joins sellable stock (as refurbished) only when that job is done.
func (s *Service) stockForRefurb(ctx context.Context, tx *uow.Tx, t store.AftersalesTradeIn) error {
	if t.VariantID == nil || t.StoreID == nil {
		return httpx.Conflict("not_stockable", "Match the device to a catalogue model before accepting.")
	}
	owner, err := tx.Q.TradeInFirstPartySeller(ctx)
	if err != nil {
		return err
	}
	u, err := tx.Q.TradeInCreateUnit(ctx, store.TradeInCreateUnitParams{VariantID: *t.VariantID, Imei: t.Imei, SerialNumber: t.SerialNumber, WarehouseID: t.StoreID,
		OwnerSellerID: owner, CostKobo: t.OfferKobo})
	if err != nil {
		return httpx.DB(err, "device")
	}
	diag := fmt.Sprintf("Refurbish trade-in %s (graded %s)", t.Reference, humanise(deref2(t.InspectedCondition)))
	job, err := tx.Q.TradeInCreateRefurb(ctx, store.TradeInCreateRefurbParams{DeviceUnitID: &u.ID, Diagnosis: &diag})
	if err != nil {
		return err
	}
	return tx.Q.TradeInLinkUnit(ctx, store.TradeInLinkUnitParams{ID: t.ID, DeviceUnitID: &u.ID, RepairJobID: &job.ID})
}

func deref2(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// DeclineTradeIn turns the offer down; the device waits at the store for collection.
func (s *Service) DeclineTradeIn(ctx context.Context, userID uuid.UUID, ref string) (TradeIn, error) {
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.UserID != userID {
			return httpx.NotFound("Trade-in")
		}
		if t.Status != "offer_sent" {
			return httpx.Conflict("bad_transition", "There's no open offer on this trade-in.")
		}
		_, err := tx.Q.TradeInDecide(ctx, store.TradeInDecideParams{ID: t.ID, Status: "declined"})
		return err
	})
}

// MarkTradeInPaid records finance's bank transfer for an accepted trade-in.
func (s *Service) MarkTradeInPaid(ctx context.Context, ref, reference string) (TradeIn, error) {
	actor := httpx.MustPrincipal(ctx).UserID
	if strings.TrimSpace(reference) == "" {
		return TradeIn{}, httpx.Invalid("reference_required", "Enter the bank transfer reference.")
	}
	return s.onTradeIn(ctx, ref, func(tx *uow.Tx, t store.AftersalesTradeIn) error {
		if t.Status != "accepted" || t.PayoutMethod == nil || *t.PayoutMethod != "bank" {
			return httpx.Conflict("bad_transition", "Only accepted bank payouts can be marked paid.")
		}
		if t.InspectedBy != nil && *t.InspectedBy == actor {
			return httpx.Forbidden("Whoever priced the device can't also pay it out.")
		}
		offer := *t.OfferKobo
		if _, _, err := s.ledger.Post(ctx, tx, ledger.Journal{Key: "trade_in:" + t.ID.String() + ":paid", Kind: tradeInFinancial, RefType: "trade_in", RefID: t.ID, PostedBy: &actor,
			Memo: "Trade-in " + t.Reference + " paid, ref " + reference, Lines: []ledger.Line{{Code: tradeInsPayable, Amount: offer}, {Code: ledger.BankOperating, Amount: -offer}}}); err != nil {
			return err
		}
		if _, err := tx.Q.TradeInPaid(ctx, store.TradeInPaidParams{ID: t.ID, PaidBy: &actor, PayoutReference: &reference}); err != nil {
			return err
		}
		if err := tx.Audit("trade_in.paid", "trade_in", t.ID.String(), map[string]any{"amountKobo": offer, "reference": reference}); err != nil {
			return err
		}
		return s.tellTradeIn(ctx, tx, t, money.Format(offer)+" sent to your bank account")
	})
}

// ExpireTradeInOffers lapses offers not answered in 7 days.
func (s *Service) ExpireTradeInOffers(ctx context.Context) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		rows, err := tx.Q.TradeInExpireOffers(ctx)
		if err != nil {
			return err
		}
		for _, r := range rows {
			if err := s.tellTradeIn(ctx, tx, store.AftersalesTradeIn{Reference: r.Reference, UserID: r.UserID}, "the offer expired. Collect your device from the store, or ask for a new offer"); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) onTradeIn(ctx context.Context, ref string, fn func(*uow.Tx, store.AftersalesTradeIn) error) (TradeIn, error) {
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		t, err := tx.Q.TradeInForUpdate(ctx, ref)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Trade-in")
		}
		if err != nil {
			return err
		}
		return fn(tx, t)
	})
	if err != nil {
		return TradeIn{}, err
	}
	return s.tradeInView(ctx, ref)
}

func (s *Service) tellTradeIn(ctx context.Context, tx *uow.Tx, t store.AftersalesTradeIn, what string) error {
	data := map[string]any{"ref": t.Reference, "status": what}
	for _, ch := range []string{"push", "in_app"} {
		if err := s.notify.Send(ctx, tx, notify.Msg{UserID: &t.UserID, Channel: ch, Category: notify.CatOrders, Template: "trade_in_update", Data: data,
			Key: "trade_in:" + t.Reference + ":" + what + ":" + ch}); err != nil {
			return err
		}
	}
	return nil
}
