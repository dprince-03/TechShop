package payments

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// bankListTTL is how long the provider's bank list is reused. It changes a few times a year;
// fetching it on every Seller Centre page would mean several paged provider calls each time.
const bankListTTL = 24 * time.Hour

type bankCache struct {
	mu      sync.Mutex
	banks   []Bank
	fetched time.Time
}

// ErrAccountNotFound means the bank has no account with that number (name enquiry failed).
var ErrAccountNotFound = errors.New("bank account not found")

// Bank is one entry in the bank list a seller picks from.
type Bank struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ResolveAccount runs a name enquiry: the name the bank holds for this account number. Sellers'
// payout accounts are checked this way before first use (payments-finance.md §9).
//
// With PAYMENTS_FAKE (development) there is no bank to ask, so the fake echoes claimedName,
// except for test numbers: ones starting with 9 return a different name (a mismatch) and
// 0000000000 is not found.
func (s *Service) ResolveAccount(ctx context.Context, bankCode, accountNumber, claimedName string) (string, error) {
	if s.d.Cfg.PaymentsFake {
		switch {
		case accountNumber == "0000000000":
			return "", ErrAccountNotFound
		case strings.HasPrefix(accountNumber, "9"):
			return "ADEWALE TESTING OKONKWO", nil
		default:
			return strings.ToUpper(strings.TrimSpace(claimedName)), nil
		}
	}
	return NewPaystack(s.d.Cfg.PaystackSecretKey).resolveAccount(ctx, bankCode, accountNumber)
}

// Banks lists the banks payouts can go to.
func (s *Service) Banks(ctx context.Context) ([]Bank, error) {
	if s.d.Cfg.PaymentsFake {
		// Development only: a short list so forms work without a provider.
		return []Bank{{"044", "Access Bank (dev)"}, {"058", "Guaranty Trust Bank (dev)"}, {"057", "Zenith Bank (dev)"}}, nil
	}
	s.banks.mu.Lock()
	defer s.banks.mu.Unlock()
	if s.banks.banks != nil && time.Since(s.banks.fetched) < bankListTTL {
		return s.banks.banks, nil
	}
	banks, err := NewPaystack(s.d.Cfg.PaystackSecretKey).banks(ctx)
	if err != nil {
		if s.banks.banks != nil {
			return s.banks.banks, nil // a stale list beats no list while the provider is down
		}
		return nil, err
	}
	s.banks.banks, s.banks.fetched = banks, time.Now()
	return banks, nil
}

// resolveAccount calls Paystack's name enquiry (GET /bank/resolve).
func (p *Paystack) resolveAccount(ctx context.Context, bankCode, accountNumber string) (string, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AccountName string `json:"account_name"`
		} `json:"data"`
	}
	q := url.Values{"account_number": {accountNumber}, "bank_code": {bankCode}}
	if err := p.call(ctx, "GET", "/bank/resolve?"+q.Encode(), nil, &out); err != nil {
		return "", err
	}
	if !out.Status || out.Data.AccountName == "" {
		return "", fmt.Errorf("%w: %s", ErrAccountNotFound, out.Message)
	}
	return out.Data.AccountName, nil
}

// banks calls Paystack's bank list (GET /bank), following the page cursor until the end.
func (p *Paystack) banks(ctx context.Context) ([]Bank, error) {
	var banks []Bank
	next := ""
	for page := 0; page < 20; page++ { // bounded: Nigeria has a few hundred banks, 100 per page
		var out struct {
			Status bool `json:"status"`
			Data   []struct {
				Name   string `json:"name"`
				Code   string `json:"code"`
				Active bool   `json:"active"`
			} `json:"data"`
			Meta struct {
				Next string `json:"next"`
			} `json:"meta"`
		}
		q := url.Values{"country": {"nigeria"}, "perPage": {"100"}, "use_cursor": {"true"}}
		if next != "" {
			q.Set("next", next)
		}
		if err := p.call(ctx, "GET", "/bank?"+q.Encode(), nil, &out); err != nil {
			return nil, err
		}
		if !out.Status {
			return nil, errors.New("paystack: bank list unavailable")
		}
		for _, b := range out.Data {
			if b.Active {
				banks = append(banks, Bank{Code: b.Code, Name: b.Name})
			}
		}
		if out.Meta.Next == "" {
			return banks, nil
		}
		next = out.Meta.Next
	}
	return banks, nil
}
