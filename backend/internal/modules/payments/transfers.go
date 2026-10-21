package payments

import (
	"context"
	"net/url"
	"strings"
)

// Transfers send money out of TechShop's provider balance to a bank account (seller payouts).
// Paystack Transfers: create a recipient once per account, then initiate with our own unique
// reference, so a retry with the same reference can never pay twice; the outcome is confirmed
// by verifying the reference (webhooks only prompt a re-check).
type Transfers interface {
	CreateRecipient(ctx context.Context, name, accountNumber, bankCode string) (string, error)
	Initiate(ctx context.Context, r TransferRequest) (TransferResult, error)
	VerifyTransfer(ctx context.Context, reference string) (TransferResult, error)
}

// TransferRequest is one payout transfer.
type TransferRequest struct {
	Reference  string // lowercase letters, digits and hyphens; at least 16 characters (Paystack)
	AmountKobo int64
	Recipient  string
	Reason     string
}

// TransferResult is the provider's answer: pending, success, failed or reversed.
type TransferResult struct {
	Status string
	Code   string // the provider's transfer code
	Reason string
}

// Transfers returns the live Paystack adapter, or the development fake with PAYMENTS_FAKE.
func (s *Service) Transfers() Transfers {
	if s.d.Cfg.PaymentsFake {
		return fakeTransfers{}
	}
	return NewPaystack(s.d.Cfg.PaystackSecretKey)
}

// fakeTransfers pays instantly in development. Accounts whose number starts with 8 are refused
// by the "bank", so the failure path can be tested.
type fakeTransfers struct{}

func (fakeTransfers) CreateRecipient(_ context.Context, _, accountNumber, _ string) (string, error) {
	if strings.HasPrefix(accountNumber, "8") {
		return "RCP_DEV_REFUSE", nil
	}
	return "RCP_DEV_" + accountNumber[len(accountNumber)-4:], nil
}

func (fakeTransfers) Initiate(_ context.Context, r TransferRequest) (TransferResult, error) {
	if r.Recipient == "RCP_DEV_REFUSE" {
		return TransferResult{Status: "failed", Code: "TRF_DEV_" + r.Reference[:8], Reason: "The bank refused the transfer (development)."}, nil
	}
	return TransferResult{Status: "success", Code: "TRF_DEV_" + r.Reference[:8]}, nil
}

func (fakeTransfers) VerifyTransfer(_ context.Context, reference string) (TransferResult, error) {
	return TransferResult{Status: "success", Code: "TRF_DEV_" + reference[:8]}, nil
}

// CreateRecipient implements Transfers (POST /transferrecipient, type nuban).
func (p *Paystack) CreateRecipient(ctx context.Context, name, accountNumber, bankCode string) (string, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			RecipientCode string `json:"recipient_code"`
		} `json:"data"`
	}
	body := map[string]any{"type": "nuban", "name": name, "account_number": accountNumber, "bank_code": bankCode, "currency": "NGN"}
	if err := p.call(ctx, "POST", "/transferrecipient", body, &out); err != nil {
		return "", err
	}
	if !out.Status || out.Data.RecipientCode == "" {
		return "", providerError("paystack recipient: " + out.Message)
	}
	return out.Data.RecipientCode, nil
}

// Initiate implements Transfers (POST /transfer from the balance). A refused request is
// checked against the reference first: after a timeout the transfer may already exist.
func (p *Paystack) Initiate(ctx context.Context, r TransferRequest) (TransferResult, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status       string `json:"status"`
			TransferCode string `json:"transfer_code"`
		} `json:"data"`
	}
	body := map[string]any{"source": "balance", "amount": r.AmountKobo, "recipient": r.Recipient, "reference": r.Reference, "reason": r.Reason}
	if err := p.call(ctx, "POST", "/transfer", body, &out); err != nil {
		return TransferResult{}, err
	}
	if !out.Status {
		if v, err := p.VerifyTransfer(ctx, r.Reference); err == nil && v.Status != "" {
			return v, nil
		}
		return TransferResult{Status: "failed", Reason: out.Message}, nil
	}
	return transferResult(out.Data.Status, out.Data.TransferCode, ""), nil
}

// VerifyTransfer implements Transfers (GET /transfer/verify/{reference}).
func (p *Paystack) VerifyTransfer(ctx context.Context, reference string) (TransferResult, error) {
	var out struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status       string `json:"status"`
			TransferCode string `json:"transfer_code"`
			Reason       string `json:"reason"`
		} `json:"data"`
	}
	if err := p.call(ctx, "GET", "/transfer/verify/"+url.PathEscape(reference), nil, &out); err != nil {
		return TransferResult{}, err
	}
	if !out.Status {
		return TransferResult{}, providerError("paystack verify transfer: " + out.Message)
	}
	return transferResult(out.Data.Status, out.Data.TransferCode, ""), nil
}

// transferResult maps Paystack's transfer statuses onto ours. "otp" means the account still
// requires an OTP for transfers (finance must disable it for API payouts): it stays pending
// until someone acts, and the sweep keeps checking.
func transferResult(status, code, reason string) TransferResult {
	switch status {
	case "success", "failed", "reversed":
		return TransferResult{Status: status, Code: code, Reason: reason}
	case "abandoned", "rejected", "blocked":
		return TransferResult{Status: "failed", Code: code, Reason: "Transfer " + status + " by the provider."}
	default: // pending, processing, queued, received, otp
		return TransferResult{Status: "pending", Code: code}
	}
}

type providerError string

func (e providerError) Error() string { return string(e) }
