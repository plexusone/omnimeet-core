// Package attendee defines how agents attend meetings hosted by others.
//
// The provider package covers meetings the application hosts: it creates the
// room and mints join tokens. Meetings on Zoom, Google Meet, Microsoft Teams,
// or Webex are hosted by someone else, so there is no token to mint. An agent
// joins them through a connector, such as phone dial-in or a browser, that
// implements Joiner.
//
// Connectors differ in what they can do. A dial-in connection carries mixed
// audio but no names or chat; a browser connection can show a display name and
// read speaker names from captions. Every Session reports its Capabilities and
// implements the optional interfaces that match them (AudioReceiver,
// AudioSender, Recorder, ParticipantLister), so callers can adapt:
//
//	sess, err := joiner.Join(ctx, target)
//	if err != nil {
//		return err
//	}
//	if rx, ok := sess.(attendee.AudioReceiver); ok {
//		go transcribe(rx.ReceiveFormat(), rx.ReceiveAudio())
//	}
package attendee

import "context"

// Platform identifies the meeting service that hosts a meeting.
type Platform string

const (
	// PlatformUnknown is a meeting whose service could not be identified.
	PlatformUnknown Platform = ""
	// PlatformZoom is Zoom.
	PlatformZoom Platform = "zoom"
	// PlatformGoogleMeet is Google Meet.
	PlatformGoogleMeet Platform = "google_meet"
	// PlatformTeams is Microsoft Teams.
	PlatformTeams Platform = "teams"
	// PlatformWebex is Cisco Webex.
	PlatformWebex Platform = "webex"
)

// String returns the platform name, or "unknown" for PlatformUnknown.
func (p Platform) String() string {
	if p == PlatformUnknown {
		return "unknown"
	}
	return string(p)
}

// JoinTarget describes a meeting hosted by someone else. Fields that a
// connector does not need may be empty: dial-in uses DialIn, a browser
// connector uses JoinURL.
type JoinTarget struct {
	// Platform is the service hosting the meeting.
	Platform Platform `json:"platform"`

	// JoinURL is the link to join the meeting in a browser or app.
	JoinURL string `json:"join_url,omitempty"`

	// MeetingID is the platform's meeting identifier as shown in the invite.
	MeetingID string `json:"meeting_id,omitempty"`

	// Passcode is the meeting passcode for joining by link or app. Phone
	// passcodes, which may differ, are in DialIn.
	Passcode string `json:"passcode,omitempty"`

	// DisplayName is the name the agent joins with, on connectors that can
	// show one.
	DisplayName string `json:"display_name,omitempty"`

	// DialIn holds phone dial-in details, or nil when the meeting has none.
	DialIn *DialInDetails `json:"dial_in,omitempty"`
}

// DialInDetails holds what a phone caller needs to join a meeting.
type DialInDetails struct {
	// Numbers are the dial-in phone numbers listed in the invite.
	Numbers []PhoneNumber `json:"numbers"`

	// ConferenceID is the number entered on the keypad to select the
	// meeting: Zoom's meeting ID, Google Meet's PIN, or Teams' phone
	// conference ID. Digits only.
	ConferenceID string `json:"conference_id"`

	// Passcode is the numeric passcode entered on the keypad, when the
	// meeting requires one. Digits only.
	Passcode string `json:"passcode,omitempty"`
}

// PhoneNumber is one dial-in number.
type PhoneNumber struct {
	// Number is in E.164 form, for example "+14155550123".
	Number string `json:"number"`

	// Country is the ISO 3166-1 alpha-2 code the number serves, for example
	// "US", or empty when unknown.
	Country string `json:"country,omitempty"`

	// Label is the location text from the invite, for example "San Jose".
	Label string `json:"label,omitempty"`

	// TollFree reports whether the invite marks the number as toll-free.
	TollFree bool `json:"toll_free,omitempty"`
}

// Joiner joins meetings hosted by others. Each connector (dial-in, browser,
// native SDK) provides one.
type Joiner interface {
	// Join starts joining the meeting and returns once the session exists.
	// Joining continues in the background: watch Session.Events for the
	// session to reach StateAdmitted, or a terminal state.
	Join(ctx context.Context, target JoinTarget) (Session, error)
}
