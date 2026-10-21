package sellers

import (
	"regexp"
	"strings"
	"unicode"
)

// Validation and matching rules for seller onboarding (trust-safety.md §2). Pure functions, so
// they are unit-tested without a database.

var (
	nubanPattern  = regexp.MustCompile(`^[0-9]{10}$`)
	ninPattern    = regexp.MustCompile(`^[0-9]{11}$`)
	idPattern     = regexp.MustCompile(`^[A-Z0-9]{6,20}$`)
	bankCodeRegex = regexp.MustCompile(`^[0-9]{3,6}$`)
	slugJunk      = regexp.MustCompile(`[^a-z0-9]+`)
)

// validIDNumber checks the shape of an identity number for its type. A NIN is 11 digits; other
// documents vary, so they only need 6–20 letters and digits. The provider or reviewer checks the
// number itself.
func validIDNumber(idType, number string) bool {
	switch idType {
	case "nin", "nin_slip":
		return ninPattern.MatchString(number)
	default:
		return idPattern.MatchString(number)
	}
}

// normaliseID strips spaces and dashes and upper-cases, so "1234 5678 901" and "12345678901"
// encrypt and hash the same way (the blind index must find the same ID however it was typed).
func normaliseID(s string) string {
	return strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(s)))
}

// last4 is the display tail of an ID or account number.
func last4(s string) string { return s[len(s)-4:] }

// requiredDocuments lists the uploads a submission needs: proof of identity and a selfie for
// everyone, and the CAC certificate for registered businesses.
func requiredDocuments(sellerType string) []string {
	if sellerType == "business" {
		return []string{"id_front", "selfie", "cac_certificate"}
	}
	return []string{"id_front", "selfie"}
}

// slugify turns a shop name into a URL slug ("Ikeja Gadget Hub!" → "ikeja-gadget-hub").
func slugify(name string) string {
	s := slugJunk.ReplaceAllString(strings.ToLower(name), "-")
	s = strings.Trim(s, "-")
	if len(s) > 48 {
		s = strings.TrimRight(s[:48], "-")
	}
	return s
}

// Name match results for a bank name enquiry against the expected holder.
const (
	NameMatch    = "match"
	NamePartial  = "partial"
	NameMismatch = "mismatch"
	NameUnknown  = "unknown"
)

// nameMatch compares the bank's account name with the expected holder (the owner's legal name
// for individuals, the shop's registered name for businesses). Banks reorder names and add
// middle names, so it compares word sets: every word of the shorter name present = match,
// at least two shared words = partial. Reviewers decide on partial and mismatch.
func nameMatch(bankName, expected string) string {
	a, b := nameWords(bankName), nameWords(expected)
	if len(a) == 0 || len(b) == 0 {
		return NameUnknown
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	inB := map[string]bool{}
	for _, w := range b {
		inB[w] = true
	}
	shared := 0
	for _, w := range a {
		if inB[w] {
			shared++
		}
	}
	switch {
	case shared == len(a):
		return NameMatch
	case shared >= 2:
		return NamePartial
	default:
		return NameMismatch
	}
}

// Words that carry no identity: honorifics and company suffixes.
var nameNoise = map[string]bool{"MR": true, "MRS": true, "MS": true, "MISS": true, "DR": true, "CHIEF": true,
	"LTD": true, "LIMITED": true, "PLC": true, "ENTERPRISES": true, "ENTERPRISE": true, "VENTURES": true, "NIG": true, "NIGERIA": true}

func nameWords(s string) []string {
	f := strings.FieldsFunc(strings.ToUpper(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	out := f[:0]
	for _, w := range f {
		if !nameNoise[w] {
			out = append(out, w)
		}
	}
	return out
}
