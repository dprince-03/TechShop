// Package validate holds Nigerian and device validators shared by every module:
// phone numbers (same rules as @techshop/api-client), NUBAN, IMEI and VIN.
package validate

import (
	"regexp"
	"strings"
)

var nonDigits = regexp.MustCompile(`[^\d]`)

// NormalisePhone converts a Nigerian mobile number to E.164 (+234…). It accepts
// 0803…, 803…, 234803… and +234803…; it returns "" if the number isn't a Nigerian mobile.
func NormalisePhone(raw string) string {
	d := nonDigits.ReplaceAllString(strings.TrimSpace(raw), "")
	var local string
	switch {
	case len(d) == 13 && strings.HasPrefix(d, "234"):
		local = d[3:]
	case len(d) == 11 && strings.HasPrefix(d, "0"):
		local = d[1:]
	case len(d) == 10:
		local = d
	default:
		return ""
	}
	// Mobile ranges start 7, 8 or 9 followed by 0 or 1 (e.g. 70x, 80x, 81x, 90x, 91x).
	if !regexp.MustCompile(`^[789][01]\d{8}$`).MatchString(local) {
		return ""
	}
	return "+234" + local
}

// MaskPhone shows a phone as +234803***4567.
func MaskPhone(e164 string) string {
	if len(e164) < 10 {
		return "***"
	}
	return e164[:7] + "***" + e164[len(e164)-4:]
}

// IMEI reports whether s is a 15-digit IMEI with a valid Luhn check digit.
func IMEI(s string) bool {
	if len(s) != 15 || nonDigits.MatchString(s) {
		return false
	}
	return luhn(s)
}

func luhn(s string) bool {
	sum := 0
	for i := 0; i < len(s); i++ {
		d := int(s[len(s)-1-i] - '0')
		if i%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// NUBAN reports whether account is a 10-digit Nigerian account number. When a 3-digit
// CBN bank code is given, the NUBAN check digit is also verified.
func NUBAN(bankCode, account string) bool {
	if len(account) != 10 || nonDigits.MatchString(account) {
		return false
	}
	if len(bankCode) != 3 || nonDigits.MatchString(bankCode) {
		return true // newer 5/6-digit codes use a different scheme; length check only
	}
	weights := []int{3, 7, 3, 3, 7, 3, 3, 7, 3, 3, 7, 3}
	digits := bankCode + account[:9]
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(digits[i]-'0') * weights[i]
	}
	check := (10 - sum%10) % 10
	return check == int(account[9]-'0')
}

var vinChars = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

// VIN reports whether v is a 17-character VIN (letters I, O, Q excluded).
func VIN(v string) bool { return vinChars.MatchString(strings.ToUpper(v)) }

var states = map[string]bool{"AB": true, "AD": true, "AK": true, "AN": true, "BA": true, "BY": true, "BE": true, "BO": true, "CR": true,
	"DE": true, "EB": true, "ED": true, "EK": true, "EN": true, "FC": true, "GO": true, "IM": true, "JI": true, "KD": true, "KN": true,
	"KT": true, "KE": true, "KO": true, "KW": true, "LA": true, "NA": true, "NI": true, "OG": true, "ON": true, "OS": true, "OY": true,
	"PL": true, "RI": true, "SO": true, "TA": true, "YO": true, "ZA": true}

// StateCode reports whether code is one of the 36 states or the FCT.
func StateCode(code string) bool { return states[code] }
