// Package invite reads meeting invitations from Zoom, Google Meet, Microsoft
// Teams, and Webex and returns the meeting as an attendee.JoinTarget: the join
// link, meeting ID and passcode, and phone dial-in details.
//
// Parse takes invitation text, such as an email body or a calendar event
// description. ParseICS takes an iCalendar (.ics) file. Pass the result to
// dialin.Choose to dial in, which returns dialin.ErrNoDialIn when the invite
// has no phone option.
package invite

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/plexusone/omnimeet-core/attendee"
)

// ErrNoMeeting is returned when the text has neither a recognizable meeting
// link nor phone dial-in details.
var ErrNoMeeting = errors.New("no meeting found in invite")

// Parse extracts the meeting from invitation text.
func Parse(text string) (attendee.JoinTarget, error) {
	text = clean(text)

	jt := attendee.JoinTarget{}
	jt.Platform, jt.JoinURL = findJoinURL(text)
	if jt.Platform == attendee.PlatformUnknown {
		jt.Platform = platformFromText(text)
	}

	switch jt.Platform {
	case attendee.PlatformZoom:
		parseZoom(text, &jt)
	case attendee.PlatformGoogleMeet:
		parseGoogleMeet(text, &jt)
	case attendee.PlatformTeams:
		parseTeams(text, &jt)
	case attendee.PlatformWebex:
		parseWebex(text, &jt)
	}

	if jt.JoinURL == "" && jt.DialIn == nil {
		return attendee.JoinTarget{}, ErrNoMeeting
	}
	return jt, nil
}

// invisible characters that calendar and mail clients insert around phone
// numbers and PINs (bidirectional marks and isolates, zero-width spaces).
var invisibleReplacer = strings.NewReplacer(
	"\u200b", "", "\u200e", "", "\u200f", "",
	"\u202a", "", "\u202b", "", "\u202c", "", "\u202d", "", "\u202e", "",
	"\u2066", "", "\u2067", "", "\u2068", "", "\u2069", "",
	"\u00a0", " ", "\r\n", "\n", "\r", "\n",
)

func clean(text string) string {
	return invisibleReplacer.Replace(text)
}

var urlPattern = regexp.MustCompile(`https?://[^\s<>"'\])]+`)

// findJoinURL returns the first link that joins a meeting, skipping help and
// "find a local number" links.
func findJoinURL(text string) (attendee.Platform, string) {
	for _, raw := range urlPattern.FindAllString(text, -1) {
		raw = strings.TrimRight(raw, ".,;")
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		if p := joinPlatform(u); p != attendee.PlatformUnknown {
			return p, raw
		}
	}
	return attendee.PlatformUnknown, ""
}

var meetCodePattern = regexp.MustCompile(`^/[a-z]{3}-[a-z]{4}-[a-z]{3}$`)

func joinPlatform(u *url.URL) attendee.Platform {
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "zoom.us" || strings.HasSuffix(host, ".zoom.us") ||
		host == "zoomgov.com" || strings.HasSuffix(host, ".zoomgov.com"):
		for _, prefix := range []string{"/j/", "/w/", "/my/", "/s/"} {
			if strings.HasPrefix(u.Path, prefix) {
				return attendee.PlatformZoom
			}
		}
	case host == "meet.google.com":
		if meetCodePattern.MatchString(u.Path) {
			return attendee.PlatformGoogleMeet
		}
	case host == "teams.microsoft.com" || host == "teams.live.com":
		if strings.Contains(u.Path, "/meetup-join/") || strings.HasPrefix(u.Path, "/meet/") {
			return attendee.PlatformTeams
		}
	case strings.HasSuffix(host, ".webex.com"):
		if strings.Contains(u.Path, "/j.php") || strings.HasPrefix(u.Path, "/meet/") ||
			strings.Contains(u.RawQuery, "MTID=") {
			return attendee.PlatformWebex
		}
	}
	return attendee.PlatformUnknown
}

// platformFromText identifies the platform from dial-in wording when the text
// has no join link.
func platformFromText(text string) attendee.Platform {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "phone conference id"):
		return attendee.PlatformTeams
	case strings.Contains(lower, "google meet") || strings.Contains(lower, "pin:"):
		return attendee.PlatformGoogleMeet
	case strings.Contains(lower, "zoom"):
		return attendee.PlatformZoom
	case strings.Contains(lower, "webex"):
		return attendee.PlatformWebex
	}
	return attendee.PlatformUnknown
}

var (
	zoomMeetingIDPattern = regexp.MustCompile(`(?i)meeting id:\s*([\d ]+\d)`)
	zoomURLIDPattern     = regexp.MustCompile(`/j/(\d+)`)
	passcodePattern      = regexp.MustCompile(`(?i)passcode:\s*(\S+)`)
	digitsOnlyPattern    = regexp.MustCompile(`^\d+$`)
)

func parseZoom(text string, jt *attendee.JoinTarget) {
	if m := zoomMeetingIDPattern.FindStringSubmatch(text); m != nil {
		jt.MeetingID = digitsOnly(m[1])
	} else if m := zoomURLIDPattern.FindStringSubmatch(jt.JoinURL); m != nil {
		jt.MeetingID = m[1]
	}

	// The first passcode is for joining by link; it may contain letters.
	// The phone passcode is numeric, from a one-tap line or a numeric
	// "Passcode:" in the dial-in section.
	phonePasscode := ""
	for i, m := range passcodePattern.FindAllStringSubmatch(text, -1) {
		if i == 0 {
			jt.Passcode = m[1]
		}
		if phonePasscode == "" && digitsOnlyPattern.MatchString(m[1]) {
			phonePasscode = m[1]
		}
	}

	numbers, tap := findNumbers(text)
	if tap.passcode != "" {
		phonePasscode = tap.passcode
	}
	conferenceID := jt.MeetingID
	if conferenceID == "" {
		conferenceID = tap.conferenceID
	}
	setDialIn(jt, numbers, conferenceID, phonePasscode)
}

var (
	meetPINPattern    = regexp.MustCompile(`(?i)pin:\s*([\d ]+\d)\s*#?`)
	meetTelPINPattern = regexp.MustCompile(`[?&]pin=(\d+)`)
)

func parseGoogleMeet(text string, jt *attendee.JoinTarget) {
	if jt.JoinURL != "" {
		if u, err := url.Parse(jt.JoinURL); err == nil {
			jt.MeetingID = strings.TrimPrefix(u.Path, "/")
		}
	}

	pin := ""
	if m := meetPINPattern.FindStringSubmatch(text); m != nil {
		pin = digitsOnly(m[1])
	} else if m := meetTelPINPattern.FindStringSubmatch(text); m != nil {
		pin = m[1]
	}

	numbers, _ := findNumbers(text)
	setDialIn(jt, numbers, pin, "")
}

var teamsConferenceIDPattern = regexp.MustCompile(`(?i)phone conference id:\s*([\d ]+\d)\s*#?`)

func parseTeams(text string, jt *attendee.JoinTarget) {
	if m := zoomMeetingIDPattern.FindStringSubmatch(text); m != nil {
		jt.MeetingID = digitsOnly(m[1])
	}
	if m := passcodePattern.FindStringSubmatch(text); m != nil {
		jt.Passcode = m[1]
	}

	numbers, tap := findNumbers(text)
	conferenceID := tap.conferenceID
	if m := teamsConferenceIDPattern.FindStringSubmatch(text); m != nil {
		conferenceID = digitsOnly(m[1])
	}
	setDialIn(jt, numbers, conferenceID, "")
}

var webexAccessCodePattern = regexp.MustCompile(`(?i)(?:access code|meeting number)[^:\n]*:\s*([\d ]+\d)`)

func parseWebex(text string, jt *attendee.JoinTarget) {
	numbers, tap := findNumbers(text)
	conferenceID := tap.conferenceID
	if m := webexAccessCodePattern.FindStringSubmatch(text); m != nil {
		conferenceID = digitsOnly(m[1])
		jt.MeetingID = conferenceID
	}
	setDialIn(jt, numbers, conferenceID, "")
}

func setDialIn(jt *attendee.JoinTarget, numbers []attendee.PhoneNumber, conferenceID, passcode string) {
	if len(numbers) == 0 || conferenceID == "" {
		return
	}
	jt.DialIn = &attendee.DialInDetails{
		Numbers:      numbers,
		ConferenceID: conferenceID,
		Passcode:     passcode,
	}
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
