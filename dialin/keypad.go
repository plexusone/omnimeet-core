package dialin

import (
	"fmt"
	"strings"
	"time"

	"github.com/plexusone/omnimeet-core/attendee"
)

// Step is one keypad action: wait for Pause, then press Digits.
type Step struct {
	// Pause is how long to wait before pressing Digits.
	Pause time.Duration `json:"pause"`

	// Digits are the keys to press: 0-9, *, and #.
	Digits string `json:"digits"`
}

// Timing sets the waits in a keypad sequence. Phone menus play a greeting
// before the first prompt and a new prompt after each entry; entering digits
// too early can be missed, too late can time out.
type Timing struct {
	// Answer is the wait after the call is answered, before the first entry.
	Answer time.Duration `json:"answer"`

	// Between is the wait between entries.
	Between time.Duration `json:"between"`
}

// DefaultTiming matches the pauses in Zoom's "one tap mobile" links: two
// dialer pauses before the meeting ID and four before the passcode, at two
// seconds each. Real calls may need tuning per platform.
var DefaultTiming = Timing{
	Answer:  4 * time.Second,
	Between: 8 * time.Second,
}

// Sequence returns the keypad steps that join the target's meeting.
func Sequence(t Target, timing Timing) ([]Step, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}

	switch t.Platform {
	case attendee.PlatformZoom:
		// After the meeting ID, Zoom asks for a participant ID. "*" followed
		// by the passcode answers with the passcode; "#" alone skips it.
		second := "#"
		if t.Passcode != "" {
			second = "*" + t.Passcode + "#"
		}
		return []Step{
			{Pause: timing.Answer, Digits: t.ConferenceID + "#"},
			{Pause: timing.Between, Digits: second},
		}, nil
	case attendee.PlatformGoogleMeet, attendee.PlatformTeams:
		if t.Passcode != "" {
			return nil, fmt.Errorf("%s dial-in takes no phone passcode", t.Platform)
		}
		return []Step{
			{Pause: timing.Answer, Digits: t.ConferenceID + "#"},
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, t.Platform)
	}
}

// DialString returns the target's keypad sequence with DefaultTiming in
// dial-string notation, ready for callsystem.WithSendDigits.
func DialString(t Target) (string, error) {
	steps, err := Sequence(t, DefaultTiming)
	if err != nil {
		return "", err
	}
	return Render(steps), nil
}

// Render converts steps to dial-string notation: W waits one second, w half a
// second, and pauses round up to the next half second.
func Render(steps []Step) string {
	var b strings.Builder
	for _, s := range steps {
		halves := int((s.Pause + 500*time.Millisecond - 1) / (500 * time.Millisecond))
		b.WriteString(strings.Repeat("W", halves/2))
		if halves%2 == 1 {
			b.WriteByte('w')
		}
		b.WriteString(s.Digits)
	}
	return b.String()
}
