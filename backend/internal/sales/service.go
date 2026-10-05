// Package sales owns carts, pricing, checkout, orders and their fulfilments
// (docs/orders-fulfilment.md, docs/backend.md §3.1). Payments plug in through PaymentPort so
// the dependency points one way: payments → sales.
package sales

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/inventory"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/ledger"
	"github.com/dprince-03/techshop/backend/internal/marketing"
	"github.com/dprince-03/techshop/backend/internal/notify"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/ratelimit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
	"github.com/dprince-03/techshop/backend/pkg/money"
)

// Seller SLA windows after payment (docs/orders-fulfilment.md §4).
const (
	sellerAcceptWindow = 24 * time.Hour
	sellerPackWindow   = 48 * time.Hour
)

// checkoutLimit stops scripted checkouts (flash-sale abuse): per customer per 10 minutes.
var checkoutLimit = ratelimit.Rule{Name: "checkout", Max: 10, Window: 10 * time.Minute}

// PaymentAction tells the client how to pay.
type PaymentAction struct {
	PaymentID   uuid.UUID        `json:"paymentId"`
	Provider    string           `json:"provider"`
	Reference   string           `json:"reference,omitempty"`
	Status      string           `json:"status"`
	CheckoutURL string           `json:"checkoutUrl,omitempty" doc:"Redirect the customer here (card, USSD, wallet)"`
	Transfer    *TransferAccount `json:"transfer,omitempty" doc:"Pay-by-transfer account for this order"`
	ExpiresAt   *time.Time       `json:"expiresAt,omitempty"`
	Message     string           `json:"message,omitempty"`
}

// TransferAccount is a one-time account number for an order.
type TransferAccount struct {
	AccountNumber string    `json:"accountNumber"`
	BankName      string    `json:"bankName"`
	AmountKobo    int64     `json:"amountKobo"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

// PaymentPort is implemented by the payments module.
type PaymentPort interface {
	// Intent creates the payment row inside the checkout transaction (no network).
	Intent(ctx context.Context, tx *uow.Tx, order store.SalesOrder, provider string) (store.PaymentsPayment, error)
	// Open calls the provider after commit and returns what the client needs to pay.
	Open(ctx context.Context, paymentID uuid.UUID) (PaymentAction, error)
	// Providers lists the methods checkout may offer right now (health-aware).
	Providers(ctx context.Context) []string
}

// Service is the sales module.
type Service struct {
	d      *kit.Deps
	inv    *inventory.Service
	mk     *marketing.Service
	ledger *ledger.Service
	notify *notify.Service
	Pay    PaymentPort
	// Delivery prices delivery from zones (set by logistics; nil = flat fees).
	Delivery DeliveryQuoter
}

// New creates the sales service.
func New(d *kit.Deps, inv *inventory.Service, mk *marketing.Service, led *ledger.Service, n *notify.Service) *Service {
	return &Service{d: d, inv: inv, mk: mk, ledger: led, notify: n}
}

// Name implements kit.Module.
func (s *Service) Name() string { return "sales" }

// ---- Carts

// CartRef identifies the caller's cart: a signed-in user, or a guest token (X-Cart-Token).
type CartRef struct {
	UserID *uuid.UUID
	Token  string
}

func tokenHash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

func newToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// findCart returns the active cart for ref (pgx.ErrNoRows when there is none).
func findCart(ctx context.Context, q *store.Queries, ref CartRef) (store.SalesCart, error) {
	if ref.UserID != nil {
		return q.SalesGetUserCart(ctx, ref.UserID)
	}
	if ref.Token == "" {
		return store.SalesCart{}, pgx.ErrNoRows
	}
	return q.SalesGetGuestCart(ctx, tokenHash(ref.Token))
}

// ensureCart finds or creates the cart; a new guest cart gets a fresh token.
func (s *Service) ensureCart(ctx context.Context, tx *uow.Tx, ref *CartRef) (store.SalesCart, error) {
	c, err := findCart(ctx, tx.Q, *ref)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		return c, err
	}
	if ref.UserID != nil {
		return tx.Q.SalesCreateCart(ctx, store.SalesCreateCartParams{UserID: ref.UserID})
	}
	ref.Token = newToken()
	return tx.Q.SalesCreateCart(ctx, store.SalesCreateCartParams{SessionTokenHash: tokenHash(ref.Token)})
}

// CartView is the cart page: priced lines plus the guest token (only when newly issued).
type CartView struct {
	Token string `json:"cartToken,omitempty" doc:"Guest cart token; send it back in X-Cart-Token"`
	*Quote
}

// ViewCart prices the caller's cart (empty quote when there is no cart).
func (s *Service) ViewCart(ctx context.Context, ref CartRef, channel, couponCode string) (CartView, error) {
	q := s.d.Q
	c, err := findCart(ctx, q, ref)
	var rows []store.SalesCartLinesRow
	if err == nil {
		if rows, err = q.SalesCartLines(ctx, c.ID); err != nil {
			return CartView{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CartView{}, err
	}
	state := ""
	if ref.UserID != nil {
		if addrs, err := q.IdentityListAddresses(ctx, ref.UserID); err == nil && len(addrs) > 0 {
			state = addrs[0].StateCode // default address first
		}
	}
	qt, err := s.Price(ctx, q, rows, PriceInput{UserID: ref.UserID, Channel: channel, ShipState: state, CouponCode: couponCode})
	return CartView{Quote: qt}, err
}

// AddItem adds (or with replace=true sets) a listing's quantity in the cart.
func (s *Service) AddItem(ctx context.Context, ref *CartRef, listingID uuid.UUID, qty int32, replace bool) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := s.ensureCart(ctx, tx, ref)
		if err != nil {
			return err
		}
		l, err := tx.Q.CatalogGetListing(ctx, listingID)
		if err != nil {
			return httpx.NotFound("Listing")
		}
		if l.Status != "active" {
			return httpx.Conflict("unavailable", "This item isn't available to buy right now.")
		}
		item, err := tx.Q.SalesUpsertCartItem(ctx, store.SalesUpsertCartItemParams{CartID: c.ID, ListingID: listingID, Quantity: qty, Replace: replace})
		if err != nil {
			return httpx.DB(err, "cart item")
		}
		if item.Quantity > 50 {
			return httpx.Invalid("quantity_too_large", "You can buy up to 50 of one item per order; contact our wholesale team for more.")
		}
		return tx.Q.SalesTouchCart(ctx, c.ID)
	})
}

// SetQuantity changes one line's quantity (0 removes it).
func (s *Service) SetQuantity(ctx context.Context, ref CartRef, itemID uuid.UUID, qty int32) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := findCart(ctx, tx.Q, ref)
		if err != nil {
			return httpx.NotFound("Cart")
		}
		var n int64
		if qty == 0 {
			n, err = tx.Q.SalesDeleteCartItem(ctx, store.SalesDeleteCartItemParams{ID: itemID, CartID: c.ID})
		} else {
			n, err = tx.Q.SalesSetCartItemQuantity(ctx, store.SalesSetCartItemQuantityParams{ID: itemID, CartID: c.ID, Quantity: qty})
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("Cart item")
		}
		return tx.Q.SalesTouchCart(ctx, c.ID)
	})
}

// Merge moves a guest cart into the signed-in user's cart (larger quantity wins per listing).
func (s *Service) Merge(ctx context.Context, userID uuid.UUID, guestToken string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		guest, err := tx.Q.SalesGetGuestCart(ctx, tokenHash(guestToken))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // nothing to merge (already merged, or expired)
		}
		if err != nil {
			return err
		}
		ref := CartRef{UserID: &userID}
		mine, err := s.ensureCart(ctx, tx, &ref)
		if err != nil {
			return err
		}
		if err := tx.Q.SalesMergeCartItems(ctx, store.SalesMergeCartItemsParams{Target: mine.ID, Source: guest.ID}); err != nil {
			return err
		}
		if err := tx.Q.SalesSetCartStatus(ctx, store.SalesSetCartStatusParams{ID: guest.ID, Status: "merged"}); err != nil {
			return err
		}
		return tx.Q.SalesTouchCart(ctx, mine.ID)
	})
}

// ---- Checkout

// CheckoutInput is the confirmed checkout.
type CheckoutInput struct {
	UserID     uuid.UUID
	Channel    string
	AddressID  uuid.UUID
	CouponCode string
	PriceHash  string
	Provider   string
}

// Placed is the result of placing an order.
type Placed struct {
	OrderID      uuid.UUID     `json:"orderId"`
	OrderNumber  string        `json:"orderNumber"`
	TotalKobo    int64         `json:"totalKobo"`
	PaymentDueAt time.Time     `json:"paymentDueAt"`
	Payment      PaymentAction `json:"payment"`
}

// Preview prices the signed-in customer's cart for a delivery address.
func (s *Service) Preview(ctx context.Context, userID uuid.UUID, channel string, addressID *uuid.UUID, coupon string) (*Quote, error) {
	q := s.d.Q
	state, lga := "", ""
	if addressID != nil {
		a, err := q.IdentityGetAddress(ctx, store.IdentityGetAddressParams{ID: *addressID, UserID: &userID})
		if err != nil {
			return nil, httpx.Invalid("bad_address", "Choose one of your saved addresses.")
		}
		state, lga = a.StateCode, deref(a.Lga)
	}
	c, err := q.SalesGetUserCart(ctx, &userID)
	if err != nil {
		return nil, httpx.Invalid("cart_empty", "Your cart is empty.")
	}
	rows, err := q.SalesCartLines(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, httpx.Invalid("cart_empty", "Your cart is empty.")
	}
	return s.Price(ctx, q, rows, PriceInput{UserID: &userID, Channel: channel, ShipState: state, ShipLGA: lga, CouponCode: coupon})
}

// PlaceOrder runs checkout in one transaction (reprice, order + fulfilments + line snapshots,
// stock reservation, flash claims, coupon, payment intent, outbox), then opens the payment
// with the provider after commit. A changed price fails with 409 price_changed.
func (s *Service) PlaceOrder(ctx context.Context, in CheckoutInput) (*Placed, error) {
	if ok, err := s.d.Limiter.Allow(ctx, checkoutLimit, in.UserID.String()); err == nil && !ok {
		return nil, httpx.E(429, "rate_limited", "Too many checkout attempts. Please wait a few minutes.")
	}
	if !contains(s.Pay.Providers(ctx), in.Provider) {
		return nil, httpx.Invalid("provider_unavailable", "That payment method isn't available right now; please choose another.")
	}
	var out Placed
	var paymentID uuid.UUID
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		cart, err := tx.Q.SalesGetUserCart(ctx, &in.UserID)
		if err != nil {
			return httpx.Invalid("cart_empty", "Your cart is empty.")
		}
		rows, err := tx.Q.SalesCartLines(ctx, cart.ID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return httpx.Invalid("cart_empty", "Your cart is empty.")
		}
		addr, err := tx.Q.IdentityGetAddress(ctx, store.IdentityGetAddressParams{ID: in.AddressID, UserID: &in.UserID})
		if err != nil {
			return httpx.Invalid("bad_address", "Choose one of your saved addresses.")
		}
		qt, err := s.Price(ctx, tx.Q, rows, PriceInput{UserID: &in.UserID, Channel: in.Channel, ShipState: addr.StateCode, ShipLGA: deref(addr.Lga), CouponCode: in.CouponCode})
		if err != nil {
			return err
		}
		if !qt.Purchasable {
			// Stock taken by a concurrent checkout surfaces as out_of_stock (docs/orders-fulfilment.md §3).
			for _, l := range qt.Lines {
				if l.Problem == "sold_out" || l.Problem == "not_enough_stock" {
					return httpx.Conflict("out_of_stock", l.ProductName+" just sold out.")
				}
			}
			return httpx.Conflict("cart_has_problems", "Some items can't be bought as they are: "+join(qt.Problems))
		}
		if qt.PriceHash != in.PriceHash {
			return httpx.Conflict("price_changed", "Prices or availability changed since you reviewed your order. Please check the new total.")
		}

		due := time.Now().Add(s.d.Cfg.OrderPaymentTTL)
		shipTo, _ := json.Marshal(map[string]any{"recipientName": addr.RecipientName, "phone": addr.Phone, "line1": addr.Line1, "line2": addr.Line2,
			"landmark": addr.Landmark, "city": addr.City, "lga": addr.Lga, "stateCode": addr.StateCode, "latitude": addr.Latitude, "longitude": addr.Longitude})
		phone := addr.Phone
		hash := qt.PriceHash
		order, err := tx.Q.SalesCreateOrder(ctx, store.SalesCreateOrderParams{Channel: in.Channel, CustomerUserID: &in.UserID, Status: "pending_payment",
			SubtotalKobo: qt.ItemsKobo - qt.VATKobo, DeliveryFeeKobo: qt.DeliveryKobo, DiscountKobo: qt.DiscountKobo, VatKobo: qt.VATKobo, TotalKobo: qt.TotalKobo,
			ShipTo: shipTo, ContactPhone: &phone, PriceHash: &hash, PaymentDueAt: &due})
		if err != nil {
			return httpx.DB(err, "order")
		}
		var reserve []inventory.ReserveLine
		for _, g := range qt.Groups {
			f, err := tx.Q.SalesCreateFulfilment(ctx, store.SalesCreateFulfilmentParams{OrderID: order.ID, SellerID: g.SellerID, FulfilledBy: g.FulfilledBy,
				Method: "delivery", DeliveryFeeKobo: g.DeliveryFeeKobo})
			if err != nil {
				return httpx.DB(err, "fulfilment")
			}
			for _, i := range g.Lines {
				l := qt.Lines[i]
				ol, err := tx.Q.SalesCreateLine(ctx, store.SalesCreateLineParams{OrderID: order.ID, FulfilmentID: f.ID, ListingID: l.ListingID, SellerID: l.SellerID,
					VariantID: l.VariantID, ProductName: l.ProductName, VariantName: l.VariantName, Sku: l.SKU, Condition: l.Condition, Quantity: l.Quantity,
					UnitPriceKobo: l.UnitPriceKobo, DiscountKobo: l.DiscountKobo, LineTotalKobo: l.LineTotalKobo, VatKobo: l.VATKobo, CommissionBps: l.commissionBps})
				if err != nil {
					return httpx.DB(err, "order line")
				}
				reserve = append(reserve, inventory.ReserveLine{OrderID: order.ID, OrderLineID: ol.ID, ListingID: l.ListingID, VariantID: l.VariantID,
					OwnerSellerID: l.SellerID, Condition: l.Condition, Quantity: l.Quantity, FulfilledBy: l.FulfilledBy})
				if l.flashPromo != nil {
					if err := s.mk.ClaimFlash(ctx, tx, *l.flashPromo, l.ListingID, in.UserID, order.ID, l.Quantity, l.flashPerUser); err != nil {
						return err
					}
				}
			}
		}
		// Sorted by listing (Price sorts lines) so concurrent checkouts lock rows in the same order.
		if err := s.inv.Reserve(ctx, tx, reserve, due); err != nil {
			return err
		}
		if qt.coupon != nil {
			saved := qt.CouponKobo
			if qt.coupon.Promotion.Kind == "free_delivery" {
				saved = qt.deliverySaved
			}
			if err := s.mk.Redeem(ctx, tx, qt.coupon, order.ID, in.UserID, saved); err != nil {
				return err
			}
		}
		if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: order.ID, ToStatus: "pending_payment", ActorUserID: &in.UserID}); err != nil {
			return err
		}
		pay, err := s.Pay.Intent(ctx, tx, order, in.Provider)
		if err != nil {
			return err
		}
		paymentID = pay.ID
		if err := tx.Q.SalesSetCartStatus(ctx, store.SalesSetCartStatusParams{ID: cart.ID, Status: "converted"}); err != nil {
			return err
		}
		tx.Emit("order", order.ID, "order.placed", map[string]any{"orderId": order.ID, "orderNumber": order.OrderNumber, "userId": in.UserID, "totalKobo": order.TotalKobo})
		out = Placed{OrderID: order.ID, OrderNumber: order.OrderNumber, TotalKobo: order.TotalKobo, PaymentDueAt: due}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Network call after commit: a provider failure leaves the order payable via retry.
	act, err := s.Pay.Open(ctx, paymentID)
	if err != nil {
		s.d.Logger.Warn("payment open failed", "order", out.OrderNumber, "err", err)
		act = PaymentAction{PaymentID: paymentID, Provider: in.Provider, Status: "failed",
			Message: "We couldn't reach the payment provider. Try again or choose another method — your items are held until " + out.PaymentDueAt.In(lagos).Format("15:04") + "."}
	}
	out.Payment = act
	return &out, nil
}

var lagos = time.FixedZone("WAT", 3600)

// ---- Payment outcomes (called by the payments module inside its transaction)

// MarkPaid moves a pending order to paid: reservations become sales, seller SLAs start, the
// sale journal is posted, and the customer is told. Returns false when the order wasn't
// pending (the caller decides what a late payment means).
func (s *Service) MarkPaid(ctx context.Context, tx *uow.Tx, orderID uuid.UUID, pay store.PaymentsPayment, paid int64) (bool, error) {
	o, err := tx.Q.SalesGetOrderForUpdate(ctx, orderID)
	if err != nil {
		return false, err
	}
	if o.Status != "pending_payment" {
		return false, nil
	}
	if err := s.inv.Commit(ctx, tx, o.ID); err != nil {
		return false, err
	}
	if _, err := tx.Q.SalesSetOrderStatus(ctx, store.SalesSetOrderStatusParams{ID: o.ID, Status: "paid"}); err != nil {
		return false, err
	}
	if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FromStatus: ptr("pending_payment"), ToStatus: "paid", Note: ptr("Payment " + pay.Provider)}); err != nil {
		return false, err
	}
	return true, s.afterPaid(ctx, tx, o, pay, paid)
}

// afterPaid does what every newly paid order needs (also used when a late payment reinstates one).
func (s *Service) afterPaid(ctx context.Context, tx *uow.Tx, o store.SalesOrder, pay store.PaymentsPayment, paid int64) error {
	fs, err := tx.Q.SalesOrderFulfilments(ctx, o.ID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, f := range fs {
		if f.FulfilledBy == "seller" {
			accept, pack := now.Add(sellerAcceptWindow), now.Add(sellerPackWindow)
			if _, err := tx.Q.SalesSetFulfilmentStatus(ctx, store.SalesSetFulfilmentStatusParams{ID: f.ID, Status: "pending", AcceptBy: &accept, PackBy: &pack}); err != nil {
				return err
			}
		}
	}
	lines, err := tx.Q.SalesOrderLines(ctx, o.ID)
	if err != nil {
		return err
	}
	sellerType := map[uuid.UUID]string{}
	for _, f := range fs {
		sellerType[f.SellerID] = f.SellerType
	}
	var sl []ledger.SaleLine
	for _, l := range lines {
		gross := l.UnitPriceKobo * int64(l.Quantity)
		x := ledger.SaleLine{SellerID: l.SellerID, FirstParty: sellerType[l.SellerID] == "first_party", Gross: gross, Discount: l.DiscountKobo, VAT: l.VatKobo}
		if !x.FirstParty && l.DiscountKobo > 0 {
			sellerFunded, err := tx.Q.SalesSellerFundedAt(ctx, store.SalesSellerFundedAtParams{ListingID: &l.ListingID, StartsAt: o.PlacedAt})
			if err != nil {
				return err
			}
			if !sellerFunded {
				x.TechShopFunded = l.DiscountKobo
			}
		}
		sl = append(sl, x)
	}
	if err := s.ledger.Sale(ctx, tx, pay.ID, o.ID, pay.Provider, paid, o.DeliveryFeeKobo, paid-o.TotalKobo, sl); err != nil {
		return err
	}
	tx.Emit("order", o.ID, "order.paid", map[string]any{"orderId": o.ID, "orderNumber": o.OrderNumber, "userId": o.CustomerUserID, "totalKobo": o.TotalKobo})
	tx.Notify("realtime", map[string]any{"topic": "order:" + o.ID.String(), "event": "status", "data": map[string]any{"status": "paid"}})
	if o.CustomerUserID != nil {
		if err := s.mk.GoalReached(ctx, tx, *o.CustomerUserID, "order_placed"); err != nil {
			return err
		}
		data := map[string]any{"order": o.OrderNumber, "total": money.Format(o.TotalKobo)}
		for _, ch := range []string{"push", "in_app"} {
			if err := s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: ch, Category: "orders", Template: "order_confirmed", Data: data,
				Key: "order:" + o.OrderNumber + ":paid:" + ch}); err != nil {
				return err
			}
		}
	}
	return nil
}

// Reinstate re-reserves a cancelled order's stock for a late payment (LATE_PAYMENT_POLICY
// PLACEHOLDER "reinstate_or_refund", owner decision #57). False = stock is gone; refund instead.
func (s *Service) Reinstate(ctx context.Context, tx *uow.Tx, orderID uuid.UUID, pay store.PaymentsPayment, paid int64) (bool, error) {
	o, err := tx.Q.SalesGetOrderForUpdate(ctx, orderID)
	if err != nil || o.Status != "cancelled" {
		return false, err
	}
	lines, err := tx.Q.SalesOrderLines(ctx, o.ID)
	if err != nil {
		return false, err
	}
	fs, err := tx.Q.SalesOrderFulfilments(ctx, o.ID)
	if err != nil {
		return false, err
	}
	var reserve []inventory.ReserveLine
	for _, l := range lines {
		lst, err := tx.Q.CatalogGetListing(ctx, l.ListingID)
		if err != nil {
			return false, err
		}
		reserve = append(reserve, inventory.ReserveLine{OrderID: o.ID, OrderLineID: l.ID, ListingID: l.ListingID, VariantID: l.VariantID, OwnerSellerID: l.SellerID,
			Condition: l.Condition, Quantity: l.Quantity, FulfilledBy: lst.FulfilledBy})
	}
	// Try in a savepoint: if the stock is gone, the outer transaction carries on with a refund.
	err = tx.Savepoint(ctx, func(sub *uow.Tx) error {
		if err := s.inv.Reserve(ctx, sub, reserve, time.Now().Add(time.Hour)); err != nil {
			return err
		}
		return s.inv.Commit(ctx, sub, o.ID)
	})
	var prob *httpx.Problem
	if errors.As(err, &prob) && prob.Status == 409 {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, f := range fs {
		if _, err := tx.Q.SalesSetFulfilmentStatus(ctx, store.SalesSetFulfilmentStatusParams{ID: f.ID, Status: "pending"}); err != nil {
			return false, err
		}
	}
	if _, err := tx.Q.SalesSetOrderStatus(ctx, store.SalesSetOrderStatusParams{ID: o.ID, Status: "paid"}); err != nil {
		return false, err
	}
	if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FromStatus: ptr("cancelled"), ToStatus: "paid", Note: ptr("Late payment: order reinstated")}); err != nil {
		return false, err
	}
	if err := tx.Audit("order.reinstated", "order", o.ID.String(), map[string]any{"payment": pay.ID}); err != nil {
		return false, err
	}
	return true, s.afterPaid(ctx, tx, o, pay, paid)
}

// ExpireUnpaid cancels an unpaid order after the provider confirmed nothing was paid: stock,
// flash units and the coupon go back.
func (s *Service) ExpireUnpaid(ctx context.Context, tx *uow.Tx, orderID uuid.UUID) error {
	o, err := tx.Q.SalesGetOrderForUpdate(ctx, orderID)
	if err != nil || o.Status != "pending_payment" {
		return err
	}
	return s.cancelWhole(ctx, tx, o, "Payment not received in time", nil, false)
}

// cancelWhole cancels every open fulfilment of an order. paid=true restocks committed stock;
// unpaid orders release their reservations.
func (s *Service) cancelWhole(ctx context.Context, tx *uow.Tx, o store.SalesOrder, reason string, actor *uuid.UUID, paid bool) error {
	if paid {
		lines, err := tx.Q.SalesOrderLines(ctx, o.ID)
		if err != nil {
			return err
		}
		for _, l := range lines {
			if l.CancelledAt == nil {
				if err := s.restockLine(ctx, tx, l, actor); err != nil {
					return err
				}
			}
		}
	} else if err := s.inv.Release(ctx, tx, o.ID); err != nil {
		return err
	}
	if err := s.mk.ReleaseOrder(ctx, tx, o.ID); err != nil {
		return err
	}
	fs, err := tx.Q.SalesOrderFulfilments(ctx, o.ID)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.Status == "cancelled" {
			continue
		}
		if _, err := tx.Q.SalesSetFulfilmentStatus(ctx, store.SalesSetFulfilmentStatusParams{ID: f.ID, Status: "cancelled", Reason: &reason}); err != nil {
			return err
		}
	}
	if _, err := tx.Q.SalesSetOrderStatus(ctx, store.SalesSetOrderStatusParams{ID: o.ID, Status: "cancelled"}); err != nil {
		return err
	}
	if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FromStatus: &o.Status, ToStatus: "cancelled", ActorUserID: actor, Note: &reason}); err != nil {
		return err
	}
	tx.Emit("order", o.ID, "order.cancelled", map[string]any{"orderId": o.ID, "orderNumber": o.OrderNumber, "paid": paid, "reason": reason, "refundKobo": o.TotalKobo})
	tx.Notify("realtime", map[string]any{"topic": "order:" + o.ID.String(), "event": "status", "data": map[string]any{"status": "cancelled"}})
	if o.CustomerUserID != nil {
		return s.notify.Send(ctx, tx, notify.Msg{UserID: o.CustomerUserID, Channel: "push", Category: "orders", Template: "order_cancelled",
			Data: map[string]any{"order": o.OrderNumber, "reason": reason}, Key: "order:" + o.OrderNumber + ":cancelled:push"})
	}
	return nil
}

// restockLine puts a paid (committed) line's stock back where it came from.
func (s *Service) restockLine(ctx context.Context, tx *uow.Tx, l store.SalesOrderLine, actor *uuid.UUID) error {
	rs, err := tx.Q.SalesLineReservations(ctx, l.ID)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if r.Status != "committed" {
			continue
		}
		if r.WarehouseID != nil {
			if _, err := s.inv.Apply(ctx, tx, inventory.Move{WarehouseID: *r.WarehouseID, VariantID: r.VariantID, Condition: r.Condition, OwnerSellerID: r.OwnerSellerID,
				Delta: r.Quantity, Reason: "adjustment", RefType: "order_cancellation", RefID: &l.OrderID, Actor: actor}); err != nil {
				return err
			}
		} else if err := tx.Q.InventoryReleaseSellerStock(ctx, store.InventoryReleaseSellerStockParams{Qty: r.Quantity, ListingID: r.ListingID}); err != nil {
			return err
		}
		if err := tx.Q.InventorySetReservationStatus(ctx, store.InventorySetReservationStatusParams{ID: r.ID, Status: "released"}); err != nil {
			return err
		}
	}
	return nil
}

// cancellable: before a seller accepts or a warehouse starts picking (docs/orders-fulfilment.md §6).
func cancellable(status string) bool { return status == "pending" }

// CancelOrder is the customer (or staff) cancelling a whole order. Paid orders get an
// automatic refund request (the payments module listens for order.cancelled with paid=true).
func (s *Service) CancelOrder(ctx context.Context, o store.SalesOrder, actor uuid.UUID, reason string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		o, err := tx.Q.SalesGetOrderForUpdate(ctx, o.ID)
		if err != nil {
			return err
		}
		switch o.Status {
		case "pending_payment":
			return s.cancelWhole(ctx, tx, o, reason, &actor, false)
		case "paid", "processing":
			fs, err := tx.Q.SalesOrderFulfilments(ctx, o.ID)
			if err != nil {
				return err
			}
			for _, f := range fs {
				if f.Status != "cancelled" && !cancellable(f.Status) {
					return httpx.Conflict("too_late_to_cancel", "Part of this order is already being prepared. Contact support, or return it after delivery.")
				}
			}
			return s.cancelWhole(ctx, tx, o, reason, &actor, true)
		default:
			return httpx.Conflict("too_late_to_cancel", "This order can no longer be cancelled.")
		}
	})
}

// CancelLine cancels one line of a paid order before it's prepared; the line (and the
// fulfilment's delivery fee if it was its last line) is refunded.
func (s *Service) CancelLine(ctx context.Context, o store.SalesOrder, lineID, actor uuid.UUID, reason string) error {
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		o, err := tx.Q.SalesGetOrderForUpdate(ctx, o.ID)
		if err != nil {
			return err
		}
		if o.Status == "pending_payment" {
			return httpx.Conflict("cancel_whole_order", "This order isn't paid yet; cancel the whole order instead.")
		}
		line, err := tx.Q.SalesCancelLine(ctx, store.SalesCancelLineParams{ID: lineID, OrderID: o.ID})
		if err != nil {
			return httpx.NotFound("Open order line")
		}
		f, err := tx.Q.SalesGetFulfilmentForUpdate(ctx, line.FulfilmentID)
		if err != nil {
			return err
		}
		if !cancellable(f.Status) {
			return httpx.Conflict("too_late_to_cancel", "This item is already being prepared.")
		}
		if err := s.restockLine(ctx, tx, line, &actor); err != nil {
			return err
		}
		refund := line.LineTotalKobo
		open := 0
		lines, err := tx.Q.SalesFulfilmentLines(ctx, f.ID)
		if err != nil {
			return err
		}
		for _, l := range lines {
			if l.CancelledAt == nil {
				open++
			}
		}
		if open == 0 {
			refund += f.DeliveryFeeKobo
			if _, err := tx.Q.SalesSetFulfilmentStatus(ctx, store.SalesSetFulfilmentStatusParams{ID: f.ID, Status: "cancelled", Reason: &reason}); err != nil {
				return err
			}
		}
		if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FulfilmentID: &f.ID, FromStatus: &f.Status, ToStatus: "line_cancelled",
			ActorUserID: &actor, Note: ptr(line.ProductName + ": " + reason)}); err != nil {
			return err
		}
		if _, err := s.Derive(ctx, tx, o.ID); err != nil {
			return err
		}
		tx.Emit("order", o.ID, "order.line_cancelled", map[string]any{"orderId": o.ID, "orderNumber": o.OrderNumber, "lineId": line.ID, "refundKobo": refund, "reason": reason})
		return nil
	})
}

// Derive recomputes an order's status from its fulfilments (docs/orders-fulfilment.md §2.1).
func (s *Service) Derive(ctx context.Context, tx *uow.Tx, orderID uuid.UUID) (string, error) {
	o, err := tx.Q.SalesGetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	if o.Status == "pending_payment" || o.Status == "refunded" || o.Status == "partially_refunded" {
		return o.Status, nil
	}
	fs, err := tx.Q.SalesOrderFulfilments(ctx, orderID)
	if err != nil {
		return "", err
	}
	var active, shipped, delivered, started int
	for _, f := range fs {
		switch f.Status {
		case "cancelled":
			continue
		case "handed_over", "in_transit", "failed", "returned":
			shipped++
		case "delivered":
			delivered++
			shipped++
		case "accepted", "picking", "packed":
			started++
		}
		active++
	}
	next := "paid"
	switch {
	case active == 0:
		next = "cancelled"
	case delivered == active:
		next = "delivered"
	case shipped == active:
		next = "shipped"
	case shipped > 0:
		next = "partially_shipped"
	case started > 0:
		next = "processing"
	}
	if next == o.Status {
		return next, nil
	}
	if _, err := tx.Q.SalesSetOrderStatus(ctx, store.SalesSetOrderStatusParams{ID: o.ID, Status: next}); err != nil {
		return "", err
	}
	if err := tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FromStatus: &o.Status, ToStatus: next}); err != nil {
		return "", err
	}
	tx.Notify("realtime", map[string]any{"topic": "order:" + o.ID.String(), "event": "status", "data": map[string]any{"status": next}})
	return next, nil
}

// SetRefundStatus marks an order refunded or partially refunded (called by payments).
func (s *Service) SetRefundStatus(ctx context.Context, tx *uow.Tx, orderID uuid.UUID, refundedTotal int64) error {
	o, err := tx.Q.SalesGetOrderForUpdate(ctx, orderID)
	if err != nil {
		return err
	}
	next := "partially_refunded"
	if refundedTotal >= o.TotalKobo {
		next = "refunded"
	}
	if o.Status == "cancelled" && next == "partially_refunded" {
		return nil // a cancelled order keeps its status until fully refunded
	}
	if o.Status == next {
		return nil
	}
	if _, err := tx.Q.SalesSetOrderStatus(ctx, store.SalesSetOrderStatusParams{ID: o.ID, Status: next}); err != nil {
		return err
	}
	return tx.Q.SalesInsertHistory(ctx, store.SalesInsertHistoryParams{OrderID: o.ID, FromStatus: &o.Status, ToStatus: next})
}

func ptr[T any](v T) *T { return &v }

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func join(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += "; "
		}
		out += x
	}
	return out
}

// Jobs implements kit.Module. Seller SLA breach jobs arrive with seller fulfilment (M4).
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}
