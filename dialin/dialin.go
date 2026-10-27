// Package dialin turns a meeting's phone dial-in details into a number to call
// and the keypad tones to enter once the call is answered.
//
// The tones follow each platform's phone menu:
//
//   - Zoom: meeting ID and #, then * and the passcode and # at the
//     participant ID prompt, or # alone to skip it. This mirrors the
//     "one tap mobile" sequence in Zoom invites.
//   - Google Meet: PIN and #.
//   - Microsoft Teams: phone conference ID and #.
//
// DialString renders the tones in the dial-string notation that
// omnivoice-core's callsystem.WithSendDigits accepts:
//
//	target, err := dialin.Choose(joinTarget, "US")
//	if err != nil {
//		return err
//	}
//	digits, err := dialin.DialString(target)
//	if err != nil {
//		return err
//	}
//	call, err := calls.MakeCall(ctx, target.Number, callsystem.WithSendDigits(digits))
package dialin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/plexusone/omnimeet-core/attendee"
)

var (
	// ErrNoDialIn is returned when a meeting has no phone dial-in.
	ErrNoDialIn = errors.New("meeting has no phone dial-in")

	// ErrNoNumberForCountry is returned when a meeting has dial-in numbers,
	// but none for the requested country.
	ErrNoNumberForCountry = errors.New("meeting has no dial-in number for the country")

	// ErrUnsupportedPlatform is returned for platforms without a keypad
	// template.
	ErrUnsupportedPlatform = errors.New("no dial-in keypad template for platform")
)

// Target is one dial-in number plus the keypad entries for the meeting.
type Target struct {
	// Platform selects the keypad template.
	Platform attendee.Platform `json:"platform"`

	// Number is the phone number to call, in E.164 form.
	Number string `json:"number"`

	// ConferenceID is entered first: Zoom's meeting ID, Google Meet's PIN, or
	// Teams' phone conference ID. Digits only.
	ConferenceID string `json:"conference_id"`

	// Passcode is the numeric phone passcode, if the meeting has one.
	Passcode string `json:"passcode,omitempty"`
}

var (
	e164Pattern   = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
	digitsPattern = regexp.MustCompile(`^[0-9]+$`)
)

// Validate reports whether the target can be dialed.
func (t Target) Validate() error {
	if !e164Pattern.MatchString(t.Number) {
		return fmt.Errorf("dial-in number %q is not in E.164 form", t.Number)
	}
	if !digitsPattern.MatchString(t.ConferenceID) {
		return fmt.Errorf("conference ID %q must be digits only", t.ConferenceID)
	}
	if t.Passcode != "" && !digitsPattern.MatchString(t.Passcode) {
		return errors.New("phone passcode must be digits only")
	}
	return nil
}

// Choose picks the number to call for a meeting. With a country code such as
// "US", it picks that country's first number, preferring toll numbers over
// toll-free ones, and returns ErrNoNumberForCountry when there is none. With
// an empty country, it picks the first number listed.
func Choose(jt attendee.JoinTarget, country string) (Target, error) {
	if jt.DialIn == nil || len(jt.DialIn.Numbers) == 0 || jt.DialIn.ConferenceID == "" {
		return Target{}, ErrNoDialIn
	}

	number, ok := chooseNumber(jt.DialIn.Numbers, country)
	if !ok {
		return Target{}, fmt.Errorf("%w: %s", ErrNoNumberForCountry, country)
	}

	t := Target{
		Platform:     jt.Platform,
		Number:       number.Number,
		ConferenceID: jt.DialIn.ConferenceID,
		Passcode:     jt.DialIn.Passcode,
	}
	if err := t.Validate(); err != nil {
		return Target{}, err
	}
	return t, nil
}

func chooseNumber(numbers []attendee.PhoneNumber, country string) (attendee.PhoneNumber, bool) {
	if country == "" {
		return numbers[0], true
	}
	var tollFree *attendee.PhoneNumber
	for i, n := range numbers {
		if !strings.EqualFold(n.Country, country) {
			continue
		}
		if !n.TollFree {
			return n, true
		}
		if tollFree == nil {
			tollFree = &numbers[i]
		}
	}
	if tollFree != nil {
		return *tollFree, true
	}
	return attendee.PhoneNumber{}, false
}
