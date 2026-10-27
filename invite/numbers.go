package invite

import (
	"regexp"
	"strings"

	"github.com/plexusone/omnimeet-core/attendee"
)

// phonePattern matches an international number written with spaces, dots,
// hyphens, or parentheses, such as "+1 669 555 0100" or "+1 (323) 555-0123".
var phonePattern = regexp.MustCompile(`\+\d[\d \t().\-]{5,}\d`)

var (
	tapConferenceIDPattern = regexp.MustCompile(`^,+(\d+)#`)
	tapPasscodePattern     = regexp.MustCompile(`^,+\d+#,+\*(\d+)#`)
	leadingCountryPattern  = regexp.MustCompile(`^\s*([A-Z]{2})\b`)
	trailingCountryPattern = regexp.MustCompile(`\(([A-Z]{2})\)\s*$`)
	parenLabelPattern      = regexp.MustCompile(`\(([^)]+)\)`)
)

// oneTap holds the keypad entries embedded in a dial-in line such as
// "+16695550100,,85012345678#,,,,*123456#".
type oneTap struct {
	conferenceID string
	passcode     string
}

// findNumbers returns the dial-in numbers in text, in order and without
// duplicates, and the keypad entries from the first one-tap line.
func findNumbers(text string) ([]attendee.PhoneNumber, oneTap) {
	var (
		numbers []attendee.PhoneNumber
		seen    = map[string]bool{}
		tap     oneTap
	)
	for line := range strings.SplitSeq(text, "\n") {
		if isURLLine(line) {
			continue
		}
		for _, loc := range phonePattern.FindAllStringIndex(line, -1) {
			e164 := "+" + digitsOnly(line[loc[0]:loc[1]])
			if len(e164) < 9 || len(e164) > 16 {
				continue
			}
			before, rest := line[:loc[0]], line[loc[1]:]

			if tap.conferenceID == "" {
				if m := tapConferenceIDPattern.FindStringSubmatch(rest); m != nil {
					tap.conferenceID = m[1]
					if p := tapPasscodePattern.FindStringSubmatch(rest); p != nil {
						tap.passcode = p[1]
					}
				}
			}
			if seen[e164] {
				continue
			}
			seen[e164] = true

			rest = strings.TrimLeft(rest, ",0123456789#*")
			numbers = append(numbers, attendee.PhoneNumber{
				Number:   e164,
				Country:  country(e164, before, rest),
				Label:    label(rest),
				TollFree: tollFree(e164, rest),
			})
		}
	}
	return numbers, tap
}

// isURLLine reports whether line is only a link, whose digits are not a
// phone number.
func isURLLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")
}

// country returns the ISO 3166-1 alpha-2 code for a number: a code written
// next to it ("(US) +1 ..." or "+1 ... US"), a country name after it, or its
// calling code.
func country(e164, before, rest string) string {
	if m := trailingCountryPattern.FindStringSubmatch(before); m != nil {
		return m[1]
	}
	if m := leadingCountryPattern.FindStringSubmatch(rest); m != nil {
		return m[1]
	}
	lower := strings.ToLower(rest)
	for _, c := range countryNames {
		if strings.Contains(lower, c.name) {
			return c.code
		}
	}
	for _, c := range callingCodes {
		if strings.HasPrefix(e164, "+"+c.prefix) {
			return c.code
		}
	}
	return ""
}

// label returns the location written after a number: "(San Jose)" in Zoom
// invites, or the text after the country in Teams invites.
func label(rest string) string {
	if m := parenLabelPattern.FindStringSubmatch(rest); m != nil {
		return strings.TrimSpace(m[1])
	}
	if i := strings.LastIndex(rest, ","); i >= 0 {
		return strings.TrimSpace(rest[i+1:])
	}
	return ""
}

var nanpTollFreePrefixes = []string{"+1800", "+1833", "+1844", "+1855", "+1866", "+1877", "+1888"}

func tollFree(e164, rest string) bool {
	lower := strings.ToLower(rest)
	if strings.Contains(lower, "toll-free") || strings.Contains(lower, "toll free") {
		return true
	}
	for _, p := range nanpTollFreePrefixes {
		if strings.HasPrefix(e164, p) {
			return true
		}
	}
	return false
}

// countryNames maps country names used in Teams invites to ISO codes.
var countryNames = []struct{ name, code string }{
	{"united states", "US"},
	{"canada", "CA"},
	{"united kingdom", "GB"},
	{"australia", "AU"},
	{"germany", "DE"},
	{"france", "FR"},
	{"ireland", "IE"},
	{"india", "IN"},
	{"japan", "JP"},
	{"singapore", "SG"},
	{"netherlands", "NL"},
}

// callingCodes maps calling codes to the most likely ISO code. Longer codes
// come first so that "+353" is not read as "+3". "+1" is shared by the United
// States and Canada and is read as US.
var callingCodes = []struct{ prefix, code string }{
	{"353", "IE"}, {"852", "HK"}, {"972", "IL"},
	{"27", "ZA"}, {"31", "NL"}, {"33", "FR"}, {"34", "ES"}, {"39", "IT"},
	{"41", "CH"}, {"44", "GB"}, {"46", "SE"}, {"49", "DE"}, {"52", "MX"},
	{"55", "BR"}, {"61", "AU"}, {"64", "NZ"}, {"65", "SG"}, {"81", "JP"},
	{"82", "KR"}, {"86", "CN"}, {"91", "IN"},
	{"1", "US"},
}
