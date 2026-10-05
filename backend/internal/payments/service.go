// Package payments takes money and gives it back (docs/payments-finance.md §2–§3, §7):
// payment intents, provider checkout, signed webhooks stored once, verify-before-trust,
// the expiry and sweep jobs, late payments and refunds with dual control.
package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/ledger"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/sales"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/money"
)

// ExcessPrefix marks refunds of money that never belonged to an order total (overpayments and
// late payments). That money already sits in liability:customer_refunds, so approving such a
// refund posts no approve journal.
const ExcessPrefix = "Excess payment: "

// breakerCooldown: an open circuit hides the provider at checkout, then lets one try through.
const breakerCooldown = time.Minute

// Service is the payments module.
type Service struct {
	d         *kit.Deps
	sales     *sales.Service
	ledger    *ledger.Service
	notify    *notify.Service
	providers map[string]Provider
}

// New creates the payments service and its provider adapters (fakes unless PAYMENTS_FAKE=false).
func New(d *kit.Deps, sl *sales.Service, led *ledger.Service, n *notify.Service) *Service {
	s := &Service{d: d, sales: sl, ledger: led, notify: n, providers: map[string]Provider{}}
	for _, name := range d.Cfg.PaymentsEnabledProviders {
		if d.Cfg.PaymentsFake {
			s.providers[name] = NewFake(name, s.fakeSecret(name), d.Cfg.AppBaseURL, s.fakeState)
		} else if name == "paystack" {
			s.providers[name] = NewPaystack(d.Cfg.PaystackSecretKey)
		}
	}
	sl.Pay = s
	return s
}

// Name implements kit.Module.
func (s *Service) Name() string { return "payments" }

// fakeSecret is the per-provider webhook secret of the development fakes (derived, never configured).
func (s *Service) fakeSecret(provider string) []byte {
	m := hmac.New(sha256.New, []byte(s.d.Cfg.TokenHMACKey))
	m.Write([]byte("fake-webhook:" + provider))
	return []byte(hex.EncodeToString(m.Sum(nil)))
}

// fakeState answers Verify for the fakes from the last signed charge event they sent.
func (s *Service) fakeState(ctx context.Context, provider, reference string) (VerifyResult, error) {
	payload, err := s.d.Q.PaymentsFakeLatestCharge(ctx, store.PaymentsFakeLatestChargeParams{Provider: provider, Reference: reference})
	if errors.Is(err, pgx.ErrNoRows) {
		return VerifyResult{Status: "pending"}, nil
	}
	if err != nil {
		return VerifyResult{}, err
	}
	var e paystackEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		return VerifyResult{}, err
	}
	return VerifyResult{Status: e.Data.Status, AmountKobo: e.Data.Amount, Currency: e.Data.Currency, Channel: e.Data.Channel, FeesKobo: e.Data.Fees, PaidAt: e.Data.PaidAt}, nil
}

func (s *Service) provider(name string) (Provider, error) {
	p, ok := s.providers[name]
	if !ok {
		return nil, httpx.Invalid("provider_unavailable", "That payment method isn't available.")
	}
	return p, nil
}

// Providers implements sales.PaymentPort: enabled providers whose circuit isn't open.
func (s *Service) Providers(ctx context.Context) []string {
	open := map[string]bool{}
	if hs, err := s.d.Q.PaymentsGetHealth(ctx); err == nil {
		for _, h := range hs {
			if h.BreakerState == "open" && h.LastFailureAt != nil && time.Since(*h.LastFailureAt) < breakerCooldown {
				open[h.Provider] = true
			}
		}
	}
	out := []string{}
	for _, name := range s.d.Cfg.PaymentsEnabledProviders {
		if _, ok := s.providers[name]; ok && !open[name] {
			out = append(out, name)
		}
	}
	return out
}

// Intent implements sales.PaymentPort: a payment row in the checkout transaction.
func (s *Service) Intent(ctx context.Context, tx *uow.Tx, o store.SalesOrder, provider string) (store.PaymentsPayment, error) {
	n, err := tx.Q.PaymentsCountForOrder(ctx, &o.ID)
	if err != nil {
		return store.PaymentsPayment{}, err
	}
	p, err := tx.Q.PaymentsCreate(ctx, store.PaymentsCreateParams{OrderID: &o.ID, Provider: provider,
		IdempotencyKey: fmt.Sprintf("order:%s:%d", o.ID, n+1), AmountKobo: o.TotalKobo, ExpiresAt: o.PaymentDueAt})
	if err != nil {
		return p, httpx.DB(err, "payment")
	}
	return p, nil
}

// reference is the merchant reference sent to providers (unique per payment).
func reference(id uuid.UUID) string {
	return "TSP" + strings.ToUpper(strings.ReplaceAll(id.String(), "-", "")[:20])
}

// Open implements sales.PaymentPort: initialise with the provider (network, after commit).
func (s *Service) Open(ctx context.Context, paymentID uuid.UUID) (sales.PaymentAction, error) {
	p, err := s.d.Q.PaymentsGet(ctx, paymentID)
	if err != nil {
		return sales.PaymentAction{}, err
	}
	prov, err := s.provider(p.Provider)
	if err != nil {
		return sales.PaymentAction{}, err
	}
	o, err := s.d.Q.PaymentsOrderWithCustomer(ctx, *p.OrderID)
	if err != nil {
		return sales.PaymentAction{}, err
	}
	email := ""
	if o.CustomerEmail != nil {
		email = *o.CustomerEmail
	} else if o.CustomerPhone != nil {
		// Providers require an email; customers who sign in by phone get a non-deliverable alias.
		email = strings.TrimPrefix(*o.CustomerPhone, "+") + "@customers.techshop.ng"
	}
	phone := ""
	if o.CustomerPhone != nil {
		phone = *o.CustomerPhone
	}
	ref := reference(p.ID)
	transfer := p.Provider == "moniepoint"
	res, err := prov.Initialize(ctx, InitRequest{Reference: ref, AmountKobo: p.AmountKobo, Email: email, Phone: phone,
		CallbackURL: s.d.Cfg.PaymentCallbackURL + "?order=" + o.OrderNumber, Transfer: transfer, ExpiresAt: deref(p.ExpiresAt)})
	if err != nil {
		_, _ = s.d.Q.PaymentsRecordFailure(ctx, p.Provider)
		_, _ = s.d.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: p.ID, Status: "failed", FailureReason: ptr("provider initialise failed")})
		return sales.PaymentAction{}, err
	}
	_ = s.d.Q.PaymentsRecordSuccess(ctx, p.Provider)
	act := sales.PaymentAction{PaymentID: p.ID, Provider: p.Provider, Reference: res.Reference, Status: "pending", CheckoutURL: res.CheckoutURL, ExpiresAt: p.ExpiresAt}
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var ch *string
		if transfer {
			ch = ptr("bank_transfer")
		}
		if _, err := tx.Q.PaymentsSetOpened(ctx, store.PaymentsSetOpenedParams{ID: p.ID, ProviderReference: &res.Reference, CheckoutUrl: &res.CheckoutURL, Channel: ch}); err != nil {
			return err
		}
		if res.AccountNumber != "" && p.ExpiresAt != nil {
			va, err := tx.Q.PaymentsCreateVirtualAccount(ctx, store.PaymentsCreateVirtualAccountParams{Provider: p.Provider, AccountNumber: res.AccountNumber, BankName: res.BankName,
				OwnerType: "order", OrderID: p.OrderID, ExpectedKobo: &p.AmountKobo, ExpiresAt: p.ExpiresAt})
			if err != nil {
				return httpx.DB(err, "transfer account")
			}
			act.Transfer = &sales.TransferAccount{AccountNumber: va.AccountNumber, BankName: va.BankName, AmountKobo: p.AmountKobo, ExpiresAt: *p.ExpiresAt}
		}
		return nil
	})
	return act, err
}

// Retry opens a new payment for an unpaid order (e.g. with another method).
func (s *Service) Retry(ctx context.Context, userID uuid.UUID, number, provider string) (sales.PaymentAction, error) {
	if !contains(s.Providers(ctx), provider) {
		return sales.PaymentAction{}, httpx.Invalid("provider_unavailable", "That payment method isn't available right now; please choose another.")
	}
	var id uuid.UUID
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		o, err := tx.Q.SalesCustomerOrderByNumber(ctx, store.SalesCustomerOrderByNumberParams{OrderNumber: number, CustomerUserID: &userID})
		if err != nil {
			return httpx.NotFound("Order")
		}
		o, err = tx.Q.SalesGetOrderForUpdate(ctx, o.ID)
		if err != nil {
			return err
		}
		if o.Status != "pending_payment" || o.PaymentDueAt == nil || time.Now().After(*o.PaymentDueAt) {
			return httpx.Conflict("not_payable", "This order can't be paid any more.")
		}
		p, err := s.Intent(ctx, tx, o, provider)
		id = p.ID
		return err
	})
	if err != nil {
		return sales.PaymentAction{}, err
	}
	return s.Open(ctx, id)
}

// ---- Webhooks

// ReceiveWebhook checks the signature over the raw body, stores the event once and queues its
// processing. Duplicates and replays are acknowledged without doing anything.
func (s *Service) ReceiveWebhook(ctx context.Context, provider string, body []byte, signature string) (duplicate bool, err error) {
	prov, ok := s.providers[provider]
	if !ok {
		return false, httpx.NotFound("Provider")
	}
	if signature == "" || !prov.VerifySignature(body, signature) {
		s.d.Logger.Warn("webhook signature rejected", "provider", provider)
		return false, httpx.E(401, "bad_signature", "Invalid webhook signature.")
	}
	ev, err := prov.ParseWebhook(body)
	if err != nil {
		return false, httpx.Invalid("bad_payload", "Unrecognised webhook payload.")
	}
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		var paymentID *uuid.UUID
		if p, err := tx.Q.PaymentsGetByReference(ctx, store.PaymentsGetByReferenceParams{Provider: provider, ProviderReference: &ev.Reference}); err == nil {
			paymentID = &p.ID
		}
		id, err := tx.Q.PaymentsInsertEvent(ctx, store.PaymentsInsertEventParams{Provider: provider, ProviderEventID: ev.ID, EventType: ev.Type,
			PaymentID: paymentID, Payload: body, SignatureValid: true})
		if errors.Is(err, pgx.ErrNoRows) {
			duplicate = true
			return nil
		}
		if err != nil {
			return err
		}
		tx.Enqueue(ProcessEventArgs{EventID: id}, critical())
		return nil
	})
	return duplicate, err
}

// ProcessEvent handles one stored webhook: charge events re-verify the payment.
func (s *Service) ProcessEvent(ctx context.Context, eventID uuid.UUID) error {
	ev, err := s.d.Q.PaymentsGetEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if ev.ProcessedAt != nil {
		return nil
	}
	if strings.HasPrefix(ev.EventType, "charge.") && ev.PaymentID != nil {
		if err := s.Reconcile(ctx, *ev.PaymentID); err != nil {
			return err
		}
	}
	return s.d.Q.PaymentsMarkEventProcessed(ctx, store.PaymentsMarkEventProcessedParams{ID: ev.ID, PaymentID: ev.PaymentID})
}

// Reconcile asks the provider for the truth about a payment and applies it. Safe to call any
// number of times: state only moves forward and the sale journal is keyed by payment.
func (s *Service) Reconcile(ctx context.Context, paymentID uuid.UUID) error {
	p, err := s.d.Q.PaymentsGet(ctx, paymentID)
	if err != nil {
		return err
	}
	if p.ProviderReference == nil || terminal(p.Status) {
		return nil
	}
	prov, err := s.provider(p.Provider)
	if err != nil {
		return err
	}
	v, err := prov.Verify(ctx, *p.ProviderReference) // network: outside the transaction
	if err != nil {
		_, _ = s.d.Q.PaymentsRecordFailure(ctx, p.Provider)
		return err
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error { return s.apply(ctx, tx, paymentID, v) })
}

func terminal(status string) bool {
	return status == "succeeded" || status == "failed" || status == "cancelled" || status == "reversed" || status == "pending_review"
}

// apply moves a payment forward according to the provider's verified answer.
func (s *Service) apply(ctx context.Context, tx *uow.Tx, paymentID uuid.UUID, v VerifyResult) error {
	p, err := tx.Q.PaymentsGetForUpdate(ctx, paymentID)
	if err != nil {
		return err
	}
	if terminal(p.Status) {
		return nil // a late "failed" after "succeeded" is recorded (event row) and ignored
	}
	switch v.Status {
	case "success":
	case "failed", "abandoned", "reversed":
		_, err := tx.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: p.ID, Status: "failed", ProviderStatus: &v.Status, FailureReason: ptr("provider reported " + v.Status)})
		return err
	default:
		return nil // still pending at the provider
	}
	if v.Currency != "" && v.Currency != "NGN" {
		return s.review(ctx, tx, p, v, "currency "+v.Currency)
	}
	excess := v.AmountKobo - p.AmountKobo
	if excess < 0 {
		if derefStr(p.Channel) == "bank_transfer" {
			return s.underpaid(ctx, tx, p, v)
		}
		return s.review(ctx, tx, p, v, "amount "+money.Format(v.AmountKobo)+" ≠ expected "+money.Format(p.AmountKobo))
	}
	if excess > 0 && derefStr(p.Channel) != "bank_transfer" {
		return s.review(ctx, tx, p, v, "amount "+money.Format(v.AmountKobo)+" ≠ expected "+money.Format(p.AmountKobo))
	}
	p, err = tx.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: p.ID, Status: "succeeded", ProviderStatus: &v.Status, FeesKobo: nonZero(v.FeesKobo)})
	if err != nil {
		return err
	}
	if va, err := tx.Q.PaymentsVirtualAccountForOrder(ctx, p.OrderID); err == nil {
		if err := tx.Q.PaymentsSetVirtualAccountStatus(ctx, store.PaymentsSetVirtualAccountStatusParams{ID: va.ID, Status: "paid", ReceivedKobo: v.AmountKobo}); err != nil {
			return err
		}
	}
	paid, err := s.sales.MarkPaid(ctx, tx, *p.OrderID, p, v.AmountKobo)
	if err != nil {
		return err
	}
	if paid {
		if excess > 0 {
			// Overpaid transfer: the order is paid; the excess is refunded (§4.3).
			return s.requestExcessRefund(ctx, tx, p, excess, "transfer overpaid by "+money.Format(excess))
		}
		return nil
	}
	return s.latePayment(ctx, tx, p, v.AmountKobo)
}

// review parks a payment for a human: amounts or currency don't match (never auto-accepted).
func (s *Service) review(ctx context.Context, tx *uow.Tx, p store.PaymentsPayment, v VerifyResult, why string) error {
	if _, err := tx.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: p.ID, Status: "pending_review", ProviderStatus: &v.Status}); err != nil {
		return err
	}
	_, err := tx.Q.PaymentsCreateException(ctx, store.PaymentsCreateExceptionParams{Kind: "amount_mismatch", Provider: &p.Provider, Reference: p.ProviderReference,
		AmountKobo: &v.AmountKobo, Detail: "Payment " + p.ID.String() + ": " + why})
	return err
}

// underpaid: a transfer short of the total keeps the order pending; the customer is told the
// remainder (docs/payments-finance.md §4.3). Expiry refunds the partial amount.
func (s *Service) underpaid(ctx context.Context, tx *uow.Tx, p store.PaymentsPayment, v VerifyResult) error {
	va, err := tx.Q.PaymentsVirtualAccountForOrder(ctx, p.OrderID)
	if err == nil {
		if err := tx.Q.PaymentsSetVirtualAccountStatus(ctx, store.PaymentsSetVirtualAccountStatusParams{ID: va.ID, Status: "active", ReceivedKobo: v.AmountKobo}); err != nil {
			return err
		}
	}
	o, err := tx.Q.SalesGetOrder(ctx, *p.OrderID)
	if err != nil {
		return err
	}
	if o.CustomerUserID == nil {
		return nil
	}
	return s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, To: derefStr(o.ContactPhone), Channel: "sms", Category: "orders", Template: "payment_partial",
		Data: map[string]any{"order": o.OrderNumber, "received": money.Format(v.AmountKobo), "remaining": money.Format(p.AmountKobo - v.AmountKobo)},
		Key:  "payment:" + p.ID.String() + ":partial:" + fmt.Sprint(v.AmountKobo)})
}

// latePayment: money arrived for an order that was already cancelled. PLACEHOLDER policy
// (LATE_PAYMENT_POLICY, owner decision #57): reinstate if all stock can be reserved again,
// otherwise refund automatically. Either way it is audited.
func (s *Service) latePayment(ctx context.Context, tx *uow.Tx, p store.PaymentsPayment, paid int64) error {
	if err := tx.Audit("payment.late", "payment", p.ID.String(), map[string]any{"order": p.OrderID, "amountKobo": paid, "policy": s.d.Cfg.LatePaymentPolicy}); err != nil {
		return err
	}
	if s.d.Cfg.LatePaymentPolicy == "reinstate_or_refund" {
		ok, err := s.sales.Reinstate(ctx, tx, *p.OrderID, p, paid)
		if err != nil || ok {
			return err
		}
	}
	// Not reinstated: book the money as owed back to the customer, then ask for the refund.
	if err := s.ledger.Sale(ctx, tx, p.ID, *p.OrderID, p.Provider, paid, 0, paid, nil); err != nil {
		return err
	}
	return s.requestExcessRefund(ctx, tx, p, paid, "late payment for a cancelled order")
}

// requestExcessRefund raises a system refund request (requested_by null) for money already
// held in liability:customer_refunds. A finance officer still approves it.
func (s *Service) requestExcessRefund(ctx context.Context, tx *uow.Tx, p store.PaymentsPayment, amount int64, why string) error {
	r, err := tx.Q.PaymentsCreateRefund(ctx, store.PaymentsCreateRefundParams{PaymentID: p.ID, OrderID: *p.OrderID, AmountKobo: amount, Method: "original", Reason: ExcessPrefix + why})
	if err != nil {
		return httpx.DB(err, "refund")
	}
	tx.Emit("refund", r.ID, "refund.requested", map[string]any{"refundId": r.ID, "orderId": p.OrderID, "amountKobo": amount, "system": true})
	return nil
}

// ---- Expiry and sweep jobs

// ExpireOrders cancels unpaid orders past their payment window — after asking the provider
// first, so a paid-but-unreported order is never cancelled.
func (s *Service) ExpireOrders(ctx context.Context) (int, error) {
	orders, err := s.d.Q.SalesUnpaidExpiredOrders(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, o := range orders {
		pays, err := s.d.Q.PaymentsForOrder(ctx, &o.ID)
		if err != nil {
			return n, err
		}
		verifyFailed := false
		for _, p := range pays {
			if terminal(p.Status) {
				continue
			}
			if err := s.Reconcile(ctx, p.ID); err != nil {
				s.d.Logger.Warn("expiry: provider verify failed; will retry", "order", o.OrderNumber, "err", err)
				verifyFailed = true
			}
		}
		if verifyFailed {
			continue // never cancel without the provider's answer
		}
		err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
			cur, err := tx.Q.SalesGetOrderForUpdate(ctx, o.ID)
			if err != nil || cur.Status != "pending_payment" {
				return err // paid meanwhile
			}
			pays, err := tx.Q.PaymentsForOrder(ctx, &o.ID)
			if err != nil {
				return err
			}
			for _, p := range pays {
				if p.Status == "initiated" || p.Status == "pending" {
					if _, err := tx.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: p.ID, Status: "cancelled", FailureReason: ptr("payment window expired")}); err != nil {
						return err
					}
				}
			}
			if va, err := tx.Q.PaymentsVirtualAccountForOrder(ctx, &o.ID); err == nil {
				if err := tx.Q.PaymentsSetVirtualAccountStatus(ctx, store.PaymentsSetVirtualAccountStatusParams{ID: va.ID, Status: "expired", ReceivedKobo: va.ReceivedKobo}); err != nil {
					return err
				}
				if va.ReceivedKobo > 0 {
					// Partial transfer then expiry: the partial amount goes back (§4.3).
					for _, p := range pays {
						if derefStr(p.Channel) == "bank_transfer" {
							if err := s.ledger.Sale(ctx, tx, p.ID, o.ID, p.Provider, va.ReceivedKobo, 0, va.ReceivedKobo, nil); err != nil {
								return err
							}
							if err := s.requestExcessRefund(ctx, tx, p, va.ReceivedKobo, "partial transfer before the order expired"); err != nil {
								return err
							}
							break
						}
					}
				}
			}
			n++
			return s.sales.ExpireUnpaid(ctx, tx, o.ID)
		})
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Sweep re-verifies payments still pending after 10 minutes, so a missed webhook never strands an order.
func (s *Service) Sweep(ctx context.Context) error {
	pays, err := s.d.Q.PaymentsStalePending(ctx, time.Now().Add(-10*time.Minute))
	if err != nil {
		return err
	}
	for _, p := range pays {
		if err := s.Reconcile(ctx, p.ID); err != nil {
			s.d.Logger.Warn("sweep: verify failed", "payment", p.ID, "err", err)
		}
	}
	return nil
}

// ---- Refunds (dual control: requester ≠ approver; approval needs step-up)

// RequestRefund opens a refund for an order. requestedBy nil = an automatic (system) request.
func (s *Service) RequestRefund(ctx context.Context, tx *uow.Tx, orderID uuid.UUID, amount int64, reason, method string, requestedBy *uuid.UUID) (store.PaymentsRefund, error) {
	pay, err := tx.Q.PaymentsSucceededForOrder(ctx, &orderID)
	if err != nil {
		return store.PaymentsRefund{}, httpx.Conflict("not_paid", "This order has no captured payment to refund.")
	}
	already, err := tx.Q.PaymentsRefundedTotal(ctx, pay.ID)
	if err != nil {
		return store.PaymentsRefund{}, err
	}
	if amount <= 0 || amount > pay.AmountKobo-already {
		return store.PaymentsRefund{}, httpx.Invalid("refund_too_large", "The refund can be at most "+money.Format(pay.AmountKobo-already)+" (captured minus already refunded).")
	}
	r, err := tx.Q.PaymentsCreateRefund(ctx, store.PaymentsCreateRefundParams{PaymentID: pay.ID, OrderID: orderID, AmountKobo: amount, Method: method, Reason: reason, RequestedBy: requestedBy})
	if err != nil {
		return r, httpx.DB(err, "refund")
	}
	tx.Emit("refund", r.ID, "refund.requested", map[string]any{"refundId": r.ID, "orderId": orderID, "amountKobo": amount, "system": requestedBy == nil})
	return r, tx.Audit("refund.requested", "refund", r.ID.String(), map[string]any{"orderId": orderID, "amountKobo": amount, "reason": reason, "method": method})
}

// Approve approves a refund: approver ≠ requester; above the threshold needs refunds.approve.
func (s *Service) Approve(ctx context.Context, refundID, approver uuid.UUID) (store.PaymentsRefund, error) {
	var out store.PaymentsRefund
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.PaymentsGetRefundForUpdate(ctx, refundID)
		if err != nil {
			return httpx.NotFound("Refund")
		}
		if r.Status != "requested" {
			return httpx.Conflict("not_pending", "This refund is not waiting for approval.")
		}
		if r.RequestedBy != nil && *r.RequestedBy == approver {
			return httpx.E(403, "separation_of_duties", "You requested this refund, so someone else must approve it.")
		}
		if r.AmountKobo > s.d.Cfg.RefundApprovalThreshold {
			ok, err := s.d.RBAC.HasPermission(ctx, approver, "refunds.approve")
			if err != nil {
				return err
			}
			if !ok {
				return httpx.Forbidden("Refunds above " + money.Format(s.d.Cfg.RefundApprovalThreshold) + " need a finance manager.")
			}
		}
		out, err = tx.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "approved", ApprovedBy: &approver})
		if err != nil {
			return httpx.DB(err, "refund")
		}
		if !strings.HasPrefix(r.Reason, ExcessPrefix) {
			if err := s.ledger.RefundApprove(ctx, tx, r.ID, r.OrderID, r.PaymentID, r.AmountKobo, &approver); err != nil {
				return err
			}
		}
		tx.Enqueue(ExecuteRefundArgs{RefundID: r.ID}, finance())
		return tx.Audit("refund.approved", "refund", r.ID.String(), map[string]any{"amountKobo": r.AmountKobo})
	})
	return out, err
}

// Reject closes a refund request without paying it.
func (s *Service) Reject(ctx context.Context, refundID, actor uuid.UUID, note string) (store.PaymentsRefund, error) {
	var out store.PaymentsRefund
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		r, err := tx.Q.PaymentsGetRefundForUpdate(ctx, refundID)
		if err != nil {
			return httpx.NotFound("Refund")
		}
		if r.Status != "requested" {
			return httpx.Conflict("not_pending", "This refund is not waiting for approval.")
		}
		out, err = tx.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "rejected"})
		if err != nil {
			return err
		}
		return tx.Audit("refund.rejected", "refund", r.ID.String(), map[string]any{"note": note, "by": actor})
	})
	return out, err
}

// Execute pays an approved refund through the provider (or as store credit).
func (s *Service) Execute(ctx context.Context, refundID uuid.UUID) error {
	r, err := s.d.Q.PaymentsGetRefund(ctx, refundID)
	if err != nil {
		return err
	}
	if r.Status != "approved" && r.Status != "processing" {
		return nil
	}
	pay, err := s.d.Q.PaymentsGet(ctx, r.PaymentID)
	if err != nil {
		return err
	}
	o, err := s.d.Q.SalesGetOrder(ctx, r.OrderID)
	if err != nil {
		return err
	}
	var res RefundResult
	var storeCreditFor *uuid.UUID
	switch {
	case r.Method == "store_credit" && o.CustomerUserID != nil:
		storeCreditFor = o.CustomerUserID
		res = RefundResult{Status: "processed", Reference: "store-credit"}
	default:
		prov, err := s.provider(pay.Provider)
		if err != nil {
			return err
		}
		if _, err := s.d.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "processing"}); err != nil {
			return err
		}
		res, err = prov.Refund(ctx, derefStr(pay.ProviderReference), r.AmountKobo)
		if err != nil {
			return err // River retries with backoff
		}
	}
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		switch res.Status {
		case "failed":
			_, err := tx.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "failed", FailureReason: ptr(res.Reason)})
			return err
		case "pending":
			_, err := tx.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "processing", ProviderReference: &res.Reference})
			return err // completed by the provider's refund webhook
		}
		return s.completeRefund(ctx, tx, r, pay, res.Reference, storeCreditFor)
	})
}

// completeRefund books a confirmed refund and tells the customer.
func (s *Service) completeRefund(ctx context.Context, tx *uow.Tx, r store.PaymentsRefund, pay store.PaymentsPayment, ref string, storeCreditFor *uuid.UUID) error {
	if _, err := tx.Q.PaymentsSetRefundStatus(ctx, store.PaymentsSetRefundStatusParams{ID: r.ID, Status: "succeeded", ProviderReference: &ref}); err != nil {
		return err
	}
	if err := s.ledger.RefundPaid(ctx, tx, r.ID, pay.Provider, r.AmountKobo, storeCreditFor); err != nil {
		return err
	}
	refunds, err := tx.Q.PaymentsOrderRefunds(ctx, r.OrderID)
	if err != nil {
		return err
	}
	var total int64
	for _, x := range refunds {
		if (x.Status == "succeeded" || x.ID == r.ID) && !strings.HasPrefix(x.Reason, ExcessPrefix) {
			total += x.AmountKobo
		}
	}
	if !strings.HasPrefix(r.Reason, ExcessPrefix) {
		if err := s.sales.SetRefundStatus(ctx, tx, r.OrderID, total); err != nil {
			return err
		}
		if total >= pay.AmountKobo {
			if _, err := tx.Q.PaymentsSetStatus(ctx, store.PaymentsSetStatusParams{ID: pay.ID, Status: "reversed"}); err != nil {
				return err
			}
		}
	}
	o, err := tx.Q.SalesGetOrder(ctx, r.OrderID)
	if err != nil {
		return err
	}
	tx.Emit("refund", r.ID, "refund.succeeded", map[string]any{"refundId": r.ID, "orderId": r.OrderID, "amountKobo": r.AmountKobo})
	if o.CustomerUserID != nil {
		return s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: "push", Category: "orders", Template: "refund_processed",
			Data: map[string]any{"order": o.OrderNumber, "amount": money.Format(r.AmountKobo)}, Key: "refund:" + r.ID.String() + ":push"})
	}
	return nil
}

func ptr[T any](v T) *T { return &v }

func deref(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
