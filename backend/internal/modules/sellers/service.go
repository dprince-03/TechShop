// Package sellers runs marketplace seller onboarding: opening a shop, identity and business
// verification (KYC), payout bank accounts, and the staff KYC review. Plans: trust-safety.md §2,
// payments-finance.md §9, identity-access.md §7.3.
//
// Seller requests run inside inShop, which checks the caller's role in the shop and sets
// app.seller_id for the transaction, so row-level security (migration 00025) also keeps every
// query to that shop's rows.
package sellers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/modules/notify"
	"github.com/dprince-03/techshop/backend/internal/modules/payments"
	"github.com/dprince-03/techshop/backend/internal/platform/httpx"
	"github.com/dprince-03/techshop/backend/internal/platform/jobs"
	"github.com/dprince-03/techshop/backend/internal/platform/kit"
	"github.com/dprince-03/techshop/backend/internal/platform/outbox"
	"github.com/dprince-03/techshop/backend/internal/platform/uow"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// newAccountHold is how long a new payout account waits before payouts may use it (account-
// takeover protection, payments-finance.md §9).
const newAccountHold = 24 * time.Hour

// Roles allowed to act, per area (identity-access.md §7.3).
var (
	anyRole     = []string{"owner", "manager", "staff", "finance"}
	ownerOnly   = []string{"owner"}
	moneyRoles  = []string{"owner", "finance"}
	reviewables = []string{"submitted", "in_review"}
)

// Banks is the name-enquiry port, implemented by the payments module.
type Banks interface {
	ResolveAccount(ctx context.Context, bankCode, accountNumber, claimedName string) (string, error)
	Banks(ctx context.Context) ([]payments.Bank, error)
}

// Service is the sellers module.
type Service struct {
	d      *kit.Deps
	banks  Banks
	notify *notify.Service
}

// New creates the module.
func New(d *kit.Deps, banks Banks, n *notify.Service) *Service {
	return &Service{d: d, banks: banks, notify: n}
}

// Name implements kit.Module.
func (s *Service) Name() string { return "sellers" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// ---- Opening a shop

// RegisterInput is what a new seller fills in.
type RegisterInput struct {
	Type        string `json:"type" enum:"business,individual"`
	DisplayName string `json:"displayName" minLength:"3" maxLength:"60"`
	StateCode   string `json:"stateCode" pattern:"^[A-Z]{2}$"`
	City        string `json:"city" minLength:"2" maxLength:"60"`
}

// RegisterShop opens a shop owned by the caller, waiting for KYC. One open shop per owner.
func (s *Service) RegisterShop(ctx context.Context, userID uuid.UUID, in RegisterInput) (Shop, error) {
	base := slugify(in.DisplayName)
	if base == "" {
		return Shop{}, httpx.Invalid("invalid_name", "Use letters or numbers in the shop name.")
	}
	var out Shop
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		n, err := tx.Q.SellersCountOwnedByUser(ctx, &userID)
		if err != nil {
			return err
		}
		if n > 0 {
			return httpx.Conflict("shop_exists", "You already have a shop. Manage it in Seller Centre.")
		}
		slug, err := freeSlug(ctx, tx.Q, base)
		if err != nil {
			return err
		}
		state, city := in.StateCode, strings.TrimSpace(in.City)
		shop, err := tx.Q.SellersRegisterShop(ctx, store.SellersRegisterShopParams{Type: in.Type, DisplayName: strings.TrimSpace(in.DisplayName),
			Slug: slug, OwnerUserID: &userID, StateCode: &state, City: &city})
		if err != nil {
			return httpx.DB(err, "shop")
		}
		if err := tx.Q.SellersAddMember(ctx, store.SellersAddMemberParams{SellerID: shop.ID, UserID: userID, Role: "owner"}); err != nil {
			return err
		}
		out = shopView(shop, "owner")
		return tx.Audit("seller.registered", "seller", shop.ID.String(), map[string]string{"type": in.Type, "slug": slug})
	})
	return out, err
}

// freeSlug returns base, or base-2 … base-9, whichever is free; then a random suffix.
func freeSlug(ctx context.Context, q *store.Queries, base string) (string, error) {
	for i := 1; i <= 9; i++ {
		slug := base
		if i > 1 {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		taken, err := q.SellersSlugTaken(ctx, slug)
		if err != nil || !taken {
			return slug, err
		}
	}
	return base + "-" + strings.ToLower(crypto.RandomToken(3)), nil
}

// MyShops lists the shops the caller belongs to.
func (s *Service) MyShops(ctx context.Context, userID uuid.UUID) ([]Shop, error) {
	rows, err := s.d.Q.SellersShopsForUser(ctx, userID)
	out := make([]Shop, len(rows))
	for i, r := range rows {
		out[i] = Shop{ID: r.ID, Type: r.Type, DisplayName: r.DisplayName, Slug: r.Slug, Status: r.Status,
			StateCode: r.StateCode, City: r.City, Role: r.MemberRole, CreatedAt: r.CreatedAt}
	}
	return out, err
}

// inShop runs fn in a transaction scoped to one shop: the caller must hold one of roles, and
// app.seller_id is set so row-level security hides every other shop. A shop the caller doesn't
// belong to is reported as not found (its existence isn't revealed).
func (s *Service) inShop(ctx context.Context, shopID uuid.UUID, roles []string, fn func(tx *uow.Tx, shop store.SellersSeller, role string) error) error {
	userID := httpx.MustPrincipal(ctx).UserID
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.PgTx.Exec(ctx, "select set_config('app.seller_id', $1, true)", shopID.String()); err != nil {
			return err
		}
		role, err := tx.Q.SellersMemberRole(ctx, store.SellersMemberRoleParams{SellerID: shopID, UserID: userID})
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound("Shop")
		}
		if err != nil {
			return err
		}
		if !slices.Contains(roles, role) {
			return httpx.Forbidden("Your role in this shop can't do that.")
		}
		shop, err := tx.Q.SellersGet(ctx, shopID)
		if err != nil {
			return httpx.DB(err, "Shop")
		}
		if shop.Status == "closed" {
			return httpx.NotFound("Shop")
		}
		return fn(tx, shop, role)
	})
}

// legalHolder is the name payout accounts must belong to: the registered business name, or
// the owner's own name for individuals.
func legalHolder(ctx context.Context, q *store.Queries, shop store.SellersSeller) (string, error) {
	if shop.Type != "individual" || shop.OwnerUserID == nil {
		return shop.DisplayName, nil
	}
	u, err := q.IdentityGetUser(ctx, *shop.OwnerUserID)
	if err != nil {
		return "", err
	}
	return u.FirstName + " " + u.LastName, nil
}

// Onboarding returns the shop with its KYC and bank accounts, and the next step to take.
func (s *Service) Onboarding(ctx context.Context, shopID uuid.UUID) (Onboarding, error) {
	var out Onboarding
	err := s.inShop(ctx, shopID, anyRole, func(tx *uow.Tx, shop store.SellersSeller, role string) error {
		out.Shop = shopView(shop, role)
		k, err := tx.Q.SellersLatestKyc(ctx, shopID)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return err
		default:
			docs, err := tx.Q.SellersKycDocuments(ctx, k.ID)
			if err != nil {
				return err
			}
			v := kycView(k, docs)
			out.Kyc = &v
		}
		if out.BankAccounts, err = s.bankViews(ctx, tx.Q, shop); err != nil {
			return err
		}
		out.NextStep = nextStep(out.Kyc, len(out.BankAccounts))
		return nil
	})
	return out, err
}

func nextStep(k *Kyc, banks int) string {
	switch {
	case k == nil:
		return "submit_kyc"
	case k.Status == "rejected" || k.Status == "needs_more_info":
		return "fix_kyc"
	case banks == 0:
		return "add_bank_account"
	case k.Status == "approved":
		return "done"
	default:
		return "await_review"
	}
}

func (s *Service) bankViews(ctx context.Context, q *store.Queries, shop store.SellersSeller) ([]BankAccount, error) {
	rows, err := q.SellersBankAccounts(ctx, shop.ID)
	if err != nil {
		return nil, err
	}
	holder, err := legalHolder(ctx, q, shop)
	if err != nil {
		return nil, err
	}
	out := make([]BankAccount, len(rows))
	for i, b := range rows {
		out[i] = bankView(b, holder)
	}
	return out, nil
}

// ---- KYC

// KycInput is a verification submission.
type KycInput struct {
	IDType    string        `json:"idType" enum:"nin,nin_slip,passport,drivers_licence,voters_card"`
	IDNumber  string        `json:"idNumber" minLength:"6" maxLength:"30" doc:"Encrypted at rest; only the last 4 digits are ever shown"`
	CACNumber string        `json:"cacRcNumber,omitempty" maxLength:"20" doc:"Required for business sellers"`
	TIN       string        `json:"tin,omitempty" maxLength:"20"`
	Documents []KycDocument `json:"documents" minItems:"1" maxItems:"6" doc:"Files uploaded with purpose kyc_document"`
}

// SubmitKyc records a verification submission for review. Allowed first time, or after a
// rejection or a request for more information.
func (s *Service) SubmitKyc(ctx context.Context, shopID uuid.UUID, in KycInput) (Kyc, error) {
	userID := httpx.MustPrincipal(ctx).UserID
	number := normaliseID(in.IDNumber)
	if !validIDNumber(in.IDType, number) {
		return Kyc{}, httpx.Invalid("invalid_id_number", "That ID number doesn't look right. A NIN is 11 digits.")
	}
	var out Kyc
	err := s.inShop(ctx, shopID, ownerOnly, func(tx *uow.Tx, shop store.SellersSeller, _ string) error {
		if shop.Type == "business" && strings.TrimSpace(in.CACNumber) == "" {
			return httpx.Invalid("cac_required", "Registered businesses must give their CAC number.")
		}
		prev, err := tx.Q.SellersLatestKyc(ctx, shopID)
		if err == nil && !slices.Contains([]string{"rejected", "needs_more_info"}, prev.Status) {
			return httpx.Conflict("kyc_in_progress", "Your verification is already "+strings.ReplaceAll(prev.Status, "_", " ")+".")
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err := s.checkDocuments(ctx, tx.Q, userID, shop.Type, in.Documents); err != nil {
			return err
		}
		sealed, err := s.d.Sealer.Seal([]byte(number), "sellers.seller_kyc_submissions:"+shopID.String())
		if err != nil {
			return err
		}
		k, err := tx.Q.SellersInsertKyc(ctx, store.SellersInsertKycParams{SellerID: shopID, IDType: in.IDType, IDNumberEncrypted: sealed,
			IDNumberHmac: s.idIndex(in.IDType, number), IDNumberLast4: last4(number), CacRcNumber: optional(in.CACNumber), Tin: optional(in.TIN)})
		if err != nil {
			return httpx.DB(err, "KYC submission")
		}
		docs := make([]store.SellersKycDocumentsRow, len(in.Documents))
		for i, d := range in.Documents {
			if err := tx.Q.SellersInsertKycDocument(ctx, store.SellersInsertKycDocumentParams{SubmissionID: k.ID, Kind: d.Kind, FileID: d.FileID}); err != nil {
				return err
			}
			docs[i] = store.SellersKycDocumentsRow{Kind: d.Kind, FileID: d.FileID}
		}
		out = kycView(k, docs)
		return tx.Audit("seller.kyc_submitted", "seller", shopID.String(), map[string]any{"submission": k.ID, "idType": in.IDType, "documents": len(in.Documents)})
	})
	return out, err
}

// checkDocuments makes sure every file is the caller's own KYC upload, each kind appears once,
// and the required kinds are present.
func (s *Service) checkDocuments(ctx context.Context, q *store.Queries, userID uuid.UUID, sellerType string, docs []KycDocument) error {
	seen := map[string]bool{}
	for _, d := range docs {
		if seen[d.Kind] {
			return httpx.Invalid("duplicate_document", "Upload each kind of document once.")
		}
		seen[d.Kind] = true
		f, err := q.PlatformGetFile(ctx, d.FileID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && (f.UploadedBy == nil || *f.UploadedBy != userID)) {
			return httpx.Invalid("unknown_file", "One of the documents wasn't found. Upload it again.")
		}
		if err != nil {
			return err
		}
		if f.Purpose != "kyc_document" || f.Status != "ready" {
			return httpx.Invalid("invalid_file", "One of the documents isn't a finished KYC upload.")
		}
	}
	for _, kind := range requiredDocuments(sellerType) {
		if !seen[kind] {
			return httpx.Invalid("missing_document", "Missing document: "+strings.ReplaceAll(kind, "_", " ")+".")
		}
	}
	return nil
}

// idIndex is the blind index that finds the same identity number across shops. NIN and NIN
// slip are one number, so they share a family.
func (s *Service) idIndex(idType, number string) []byte {
	family := idType
	if idType == "nin_slip" {
		family = "nin"
	}
	return crypto.HMAC(s.d.Cfg.BlindIndexKey, "kyc:"+family+":"+number)
}

func (s *Service) accountIndex(bankCode, number string) []byte {
	return crypto.HMAC(s.d.Cfg.BlindIndexKey, "nuban:"+bankCode+":"+number)
}

func optional(s string) *string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return &s
}

// ---- Bank accounts

// BankInput is a new payout account.
type BankInput struct {
	BankCode      string `json:"bankCode" pattern:"^[0-9]{3,6}$"`
	AccountNumber string `json:"accountNumber" pattern:"^[0-9]{10}$" doc:"10-digit NUBAN; encrypted at rest"`
	AccountName   string `json:"accountName" minLength:"3" maxLength:"100"`
}

// AddBankAccount verifies the account with the bank (name enquiry) and saves it. The first
// account becomes the default. Payouts can't use a new account for 24 hours.
func (s *Service) AddBankAccount(ctx context.Context, shopID uuid.UUID, in BankInput) (BankAccount, error) {
	if !nubanPattern.MatchString(in.AccountNumber) || !bankCodeRegex.MatchString(in.BankCode) {
		return BankAccount{}, httpx.Invalid("invalid_account", "Enter the 10-digit account number and pick your bank.")
	}
	index := s.accountIndex(in.BankCode, in.AccountNumber)
	// Check access and duplicates first, then ask the bank outside any transaction (no network
	// calls while holding database locks), then save.
	if err := s.inShop(ctx, shopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		exists, err := tx.Q.SellersBankAccountExists(ctx, store.SellersBankAccountExistsParams{SellerID: shopID, AccountNumberHmac: index})
		if err == nil && exists {
			return httpx.Conflict("account_exists", "This account is already saved.")
		}
		return err
	}); err != nil {
		return BankAccount{}, err
	}
	resolved, err := s.banks.ResolveAccount(ctx, in.BankCode, in.AccountNumber, in.AccountName)
	if errors.Is(err, payments.ErrAccountNotFound) {
		return BankAccount{}, httpx.Invalid("account_not_found", "The bank couldn't find that account. Check the number and the bank.")
	}
	if err != nil {
		s.d.Logger.Warn("bank name enquiry failed", "error", err)
		return BankAccount{}, httpx.E(503, "bank_unavailable", "We couldn't reach the bank to check the account. Try again shortly.")
	}

	var out BankAccount
	err = s.inShop(ctx, shopID, moneyRoles, func(tx *uow.Tx, shop store.SellersSeller, _ string) error {
		sealed, err := s.d.Sealer.Seal([]byte(in.AccountNumber), "sellers.seller_bank_accounts:"+shopID.String())
		if err != nil {
			return err
		}
		existing, err := tx.Q.SellersBankAccounts(ctx, shopID)
		if err != nil {
			return err
		}
		now, provider := time.Now(), "paystack"
		b, err := tx.Q.SellersInsertBankAccount(ctx, store.SellersInsertBankAccountParams{SellerID: shopID, BankCode: in.BankCode,
			AccountName: strings.TrimSpace(in.AccountName), ResolvedAccountName: &resolved, AccountNumberEncrypted: sealed,
			AccountNumberHmac: index, AccountNumberLast4: last4(in.AccountNumber), Provider: &provider, VerifiedAt: &now, IsDefault: len(existing) == 0})
		if err != nil {
			return httpx.DB(err, "bank account")
		}
		holder, err := legalHolder(ctx, tx.Q, shop)
		if err != nil {
			return err
		}
		out = bankView(b, holder)
		return tx.Audit("seller.bank_account_added", "seller", shopID.String(), map[string]any{"account": b.ID, "last4": b.AccountNumberLast4, "nameMatch": out.NameMatch})
	})
	return out, err
}

// SetDefaultBank makes an account the one payouts go to.
func (s *Service) SetDefaultBank(ctx context.Context, shopID, accountID uuid.UUID) error {
	return s.inShop(ctx, shopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		if err := tx.Q.SellersClearDefaultBank(ctx, shopID); err != nil {
			return err
		}
		n, err := tx.Q.SellersSetDefaultBank(ctx, store.SellersSetDefaultBankParams{ID: accountID, SellerID: shopID})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("Bank account")
		}
		return tx.Audit("seller.bank_account_default", "seller", shopID.String(), map[string]any{"account": accountID})
	})
}

// DeleteBank removes an account that isn't the default.
func (s *Service) DeleteBank(ctx context.Context, shopID, accountID uuid.UUID) error {
	return s.inShop(ctx, shopID, moneyRoles, func(tx *uow.Tx, _ store.SellersSeller, _ string) error {
		n, err := tx.Q.SellersDeleteBankAccount(ctx, store.SellersDeleteBankAccountParams{ID: accountID, SellerID: shopID})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.Conflict("cannot_delete", "That account doesn't exist or is your default. Choose another default first.")
		}
		return tx.Audit("seller.bank_account_removed", "seller", shopID.String(), map[string]any{"account": accountID})
	})
}

// ---- Staff: vendor review

// Vendor is the reviewer's full view of a shop.
func (s *Service) Vendor(ctx context.Context, shopID uuid.UUID) (Vendor, error) {
	q := s.d.Q
	shop, err := q.SellersGet(ctx, shopID)
	if err != nil {
		return Vendor{}, httpx.DB(err, "Seller")
	}
	if shop.Type == "first_party" {
		return Vendor{}, httpx.NotFound("Seller")
	}
	out := Vendor{Shop: shopView(shop, "")}
	if out.LegalHolder, err = legalHolder(ctx, q, shop); err != nil {
		return Vendor{}, err
	}
	members, err := q.SellersMembers(ctx, shopID)
	if err != nil {
		return Vendor{}, err
	}
	out.Members = make([]Member, len(members))
	for i, m := range members {
		out.Members[i] = Member{UserID: m.UserID, Name: m.FirstName + " " + m.LastName, Role: m.Role, CreatedAt: m.CreatedAt}
	}
	k, err := q.SellersLatestKyc(ctx, shopID)
	if err == nil {
		docs, err := q.SellersKycDocuments(ctx, k.ID)
		if err != nil {
			return Vendor{}, err
		}
		v := kycView(k, docs)
		out.Kyc = &v
		if out.OtherSellersWithIDNumber, err = q.SellersOtherSellersWithIDNumber(ctx, store.SellersOtherSellersWithIDNumberParams{IDNumberHmac: k.IDNumberHmac, SellerID: shopID}); err != nil {
			return Vendor{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Vendor{}, err
	}
	banks, err := q.SellersBankAccounts(ctx, shopID)
	if err != nil {
		return Vendor{}, err
	}
	out.BankAccounts = make([]StaffBankAccount, len(banks))
	for i, b := range banks {
		n, err := q.SellersOtherSellersWithAccount(ctx, store.SellersOtherSellersWithAccountParams{AccountNumberHmac: b.AccountNumberHmac, SellerID: shopID})
		if err != nil {
			return Vendor{}, err
		}
		out.BankAccounts[i] = StaffBankAccount{BankAccount: bankView(b, out.LegalHolder), OtherSellersWithAccount: n}
	}
	return out, nil
}

// ReviewInput is a reviewer's decision.
type ReviewInput struct {
	Decision string `json:"decision" enum:"approved,rejected,needs_more_info"`
	Note     string `json:"note,omitempty" maxLength:"500" doc:"Required unless approving; the seller sees it"`
}

// ReviewKyc records a decision. The reviewer can't be a member of the shop (separation of
// duties); approval needs a verified bank account and activates the shop.
func (s *Service) ReviewKyc(ctx context.Context, submissionID uuid.UUID, in ReviewInput) (Kyc, error) {
	reviewer := httpx.MustPrincipal(ctx).UserID
	note := optional(in.Note)
	if in.Decision != "approved" && note == nil {
		return Kyc{}, httpx.Invalid("note_required", "Tell the seller why, so they can fix it.")
	}
	var out Kyc
	var owner store.IdentityUser
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		k, err := tx.Q.SellersGetKycForUpdate(ctx, submissionID)
		if err != nil {
			return httpx.DB(err, "KYC submission")
		}
		if !slices.Contains(reviewables, k.Status) {
			return httpx.Conflict("bad_transition", "This submission was already "+strings.ReplaceAll(k.Status, "_", " ")+".")
		}
		members, err := tx.Q.SellersMembers(ctx, k.SellerID)
		if err != nil {
			return err
		}
		for _, m := range members {
			if m.UserID == reviewer {
				return httpx.E(403, "separation_of_duties", "You can't review a shop you belong to.")
			}
		}
		shop, err := tx.Q.SellersGet(ctx, k.SellerID)
		if err != nil {
			return err
		}
		if in.Decision == "approved" {
			banks, err := tx.Q.SellersBankAccounts(ctx, k.SellerID)
			if err != nil {
				return err
			}
			if !slices.ContainsFunc(banks, func(b store.SellersSellerBankAccount) bool { return b.VerifiedAt != nil }) {
				return httpx.Conflict("bank_account_missing", "The seller needs a verified bank account before approval.")
			}
		}
		var reason *string
		if in.Decision != "approved" {
			reason = note
		}
		k, err = tx.Q.SellersReviewKyc(ctx, store.SellersReviewKycParams{ID: k.ID, Status: in.Decision, ReviewedBy: &reviewer, RejectionReason: reason})
		if err != nil {
			return err
		}
		if in.Decision == "approved" && shop.Status == "pending_kyc" {
			if _, err := tx.Q.SellersSetStatus(ctx, store.SellersSetStatusParams{ID: shop.ID, Status: "active"}); err != nil {
				return err
			}
		}
		docs, err := tx.Q.SellersKycDocuments(ctx, k.ID)
		if err != nil {
			return err
		}
		out = kycView(k, docs)
		if err := tx.Audit("seller.kyc_reviewed", "seller", shop.ID.String(), map[string]any{"submission": k.ID, "decision": in.Decision}); err != nil {
			return err
		}
		if shop.OwnerUserID == nil {
			return nil
		}
		if owner, err = tx.Q.IdentityGetUser(ctx, *shop.OwnerUserID); err != nil {
			return err
		}
		return s.tellSeller(ctx, tx, owner, k.ID, in.Decision, note)
	})
	return out, err
}

// tellSeller emails the owner the decision (in the same transaction, via the outbox).
func (s *Service) tellSeller(ctx context.Context, tx *uow.Tx, owner store.IdentityUser, submissionID uuid.UUID, decision string, note *string) error {
	if owner.Email == nil {
		return nil
	}
	result := map[string]string{"approved": "approved", "rejected": "not approved", "needs_more_info": "put on hold: we need more information"}[decision]
	text := "You can start listing products in Seller Centre."
	if note != nil {
		text = *note
	}
	return s.notify.Send(ctx, tx, notify.Msg{UserID: &owner.ID, To: *owner.Email, Channel: "email", Category: notify.CatSeller,
		Template: "seller_kyc_result", Data: map[string]any{"result": result, "note": text}, Key: "kyc:" + submissionID.String() + ":" + decision})
}

// RevealID decrypts a submission's full ID number for a reviewer who gives a reason. The guard
// requires a recent authenticator check (vendors.reveal_id is sensitive); every reveal is audited.
func (s *Service) RevealID(ctx context.Context, submissionID uuid.UUID, reason string) (string, error) {
	var number string
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		k, err := tx.Q.SellersGetKyc(ctx, submissionID)
		if err != nil {
			return httpx.DB(err, "KYC submission")
		}
		plain, err := s.d.Sealer.Open(k.IDNumberEncrypted, "sellers.seller_kyc_submissions:"+k.SellerID.String())
		if err != nil {
			return err
		}
		number = string(plain)
		return tx.Audit("seller.kyc_id_revealed", "seller", k.SellerID.String(), map[string]any{"submission": k.ID, "reason": reason})
	})
	return number, err
}
