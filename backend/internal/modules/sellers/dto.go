package sellers

import (
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// Response types. ID and account numbers only ever leave as their last 4 digits; the encrypted
// values and blind indexes never do.

// Shop is a seller account as its own members see it.
type Shop struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type" enum:"business,individual,first_party"`
	DisplayName string    `json:"displayName"`
	Slug        string    `json:"slug"`
	Status      string    `json:"status" enum:"pending_kyc,active,suspended,closed"`
	StateCode   *string   `json:"stateCode"`
	City        *string   `json:"city"`
	Role        string    `json:"role,omitempty" doc:"Your role in this shop" enum:"owner,manager,staff,finance"`
	CreatedAt   time.Time `json:"createdAt"`
}

func shopView(s store.SellersSeller, role string) Shop {
	return Shop{ID: s.ID, Type: s.Type, DisplayName: s.DisplayName, Slug: s.Slug, Status: s.Status,
		StateCode: s.StateCode, City: s.City, Role: role, CreatedAt: s.CreatedAt}
}

// KycDocument is one uploaded file of a submission (download via /files/{id}/url).
type KycDocument struct {
	Kind   string    `json:"kind" enum:"id_front,id_back,cac_certificate,utility_bill,selfie"`
	FileID uuid.UUID `json:"fileId"`
}

// Kyc is a verification submission.
type Kyc struct {
	ID          uuid.UUID     `json:"id"`
	Status      string        `json:"status" enum:"submitted,in_review,approved,rejected,needs_more_info"`
	IDType      string        `json:"idType"`
	IDLast4     string        `json:"idNumberLast4"`
	CACNumber   *string       `json:"cacRcNumber"`
	TIN         *string       `json:"tin"`
	Note        *string       `json:"note" doc:"The reviewer's reason (rejected) or request (needs_more_info)"`
	Documents   []KycDocument `json:"documents"`
	SubmittedAt time.Time     `json:"submittedAt"`
	ReviewedAt  *time.Time    `json:"reviewedAt"`
}

func kycView(k store.SellersSellerKycSubmission, docs []store.SellersKycDocumentsRow) Kyc {
	out := Kyc{ID: k.ID, Status: k.Status, IDType: k.IDType, IDLast4: k.IDNumberLast4, CACNumber: k.CacRcNumber, TIN: k.Tin,
		Note: k.RejectionReason, SubmittedAt: k.CreatedAt, ReviewedAt: k.ReviewedAt, Documents: make([]KycDocument, len(docs))}
	for i, d := range docs {
		out.Documents[i] = KycDocument{Kind: d.Kind, FileID: d.FileID}
	}
	return out
}

// BankAccount is a payout account.
type BankAccount struct {
	ID                  uuid.UUID `json:"id"`
	BankCode            string    `json:"bankCode"`
	AccountName         string    `json:"accountName" doc:"The name the seller typed"`
	ResolvedAccountName *string   `json:"resolvedAccountName" doc:"The name the bank holds"`
	Last4               string    `json:"accountNumberLast4"`
	NameMatch           string    `json:"nameMatch" enum:"match,partial,mismatch,unknown" doc:"Bank name vs the shop's legal holder"`
	Verified            bool      `json:"verified"`
	IsDefault           bool      `json:"isDefault"`
	PayoutsFrom         time.Time `json:"payoutsFrom" doc:"New accounts are held for 24 hours before payouts can use them"`
	CreatedAt           time.Time `json:"createdAt"`
}

func bankView(b store.SellersSellerBankAccount, holder string) BankAccount {
	match := NameUnknown
	if b.ResolvedAccountName != nil {
		match = nameMatch(*b.ResolvedAccountName, holder)
	}
	return BankAccount{ID: b.ID, BankCode: b.BankCode, AccountName: b.AccountName, ResolvedAccountName: b.ResolvedAccountName,
		Last4: b.AccountNumberLast4, NameMatch: match, Verified: b.VerifiedAt != nil, IsDefault: b.IsDefault,
		PayoutsFrom: b.CreatedAt.Add(newAccountHold), CreatedAt: b.CreatedAt}
}

// Onboarding is everything the Seller Centre checklist needs in one call.
type Onboarding struct {
	Shop         Shop          `json:"shop"`
	Kyc          *Kyc          `json:"kyc"`
	BankAccounts []BankAccount `json:"bankAccounts"`
	NextStep     string        `json:"nextStep" enum:"submit_kyc,add_bank_account,await_review,fix_kyc,done" doc:"What the seller should do next"`
}

// VendorRow is one line of the staff vendor list.
type VendorRow struct {
	ID             uuid.UUID  `json:"id"`
	Type           string     `json:"type"`
	DisplayName    string     `json:"displayName"`
	Slug           string     `json:"slug"`
	Status         string     `json:"status"`
	City           *string    `json:"city"`
	StateCode      *string    `json:"stateCode"`
	KycStatus      *string    `json:"kycStatus"`
	KycSubmittedAt *time.Time `json:"kycSubmittedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func vendorRow(r store.SellersStaffListRow) VendorRow {
	return VendorRow{ID: r.ID, Type: r.Type, DisplayName: r.DisplayName, Slug: r.Slug, Status: r.Status, City: r.City,
		StateCode: r.StateCode, KycStatus: r.KycStatus, KycSubmittedAt: r.KycSubmittedAt, CreatedAt: r.CreatedAt}
}

// Member is a person acting for a shop (names only; contact details stay masked).
type Member struct {
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// StaffBankAccount adds the duplicate signal reviewers need.
type StaffBankAccount struct {
	BankAccount
	OtherSellersWithAccount int64 `json:"otherSellersWithAccount" doc:"Other shops using the same account number: a banned seller returning?"`
}

// Vendor is the KYC reviewer's view of a shop.
type Vendor struct {
	Shop                     Shop               `json:"shop"`
	LegalHolder              string             `json:"legalHolder" doc:"Name bank accounts are compared with"`
	Members                  []Member           `json:"members"`
	Kyc                      *Kyc               `json:"kyc"`
	OtherSellersWithIDNumber int64              `json:"otherSellersWithIdNumber" doc:"Other shops that submitted the same ID number"`
	BankAccounts             []StaffBankAccount `json:"bankAccounts"`
}
