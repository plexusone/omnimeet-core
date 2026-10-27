package invite

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/plexusone/omnimeet-core/attendee"
	"github.com/plexusone/omnimeet-core/dialin"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

func TestParseZoom(t *testing.T) {
	jt, err := Parse(readTestdata(t, "zoom.txt"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if jt.Platform != attendee.PlatformZoom {
		t.Errorf("Platform = %q, want zoom", jt.Platform)
	}
	if jt.JoinURL != "https://us02web.zoom.us/j/85012345678?pwd=AbCdEfGhIjKlMnOp.1" {
		t.Errorf("JoinURL = %q", jt.JoinURL)
	}
	if jt.MeetingID != "85012345678" {
		t.Errorf("MeetingID = %q, want 85012345678", jt.MeetingID)
	}
	if jt.Passcode != "Rv7xQ2" {
		t.Errorf("Passcode = %q, want the web passcode Rv7xQ2", jt.Passcode)
	}
	if jt.DialIn == nil {
		t.Fatal("DialIn = nil")
	}
	if jt.DialIn.ConferenceID != "85012345678" || jt.DialIn.Passcode != "482913" {
		t.Errorf("DialIn conference ID/passcode = %q/%q, want 85012345678/482913",
			jt.DialIn.ConferenceID, jt.DialIn.Passcode)
	}

	want := []attendee.PhoneNumber{
		{Number: "+16695550100", Country: "US", Label: "San Jose"},
		{Number: "+19295550100", Country: "US", Label: "New York"},
		{Number: "+18335550100", Country: "US", TollFree: true},
		{Number: "+442079460100", Country: "GB"},
	}
	if !reflect.DeepEqual(jt.DialIn.Numbers, want) {
		t.Errorf("Numbers =\n%+v\nwant\n%+v", jt.DialIn.Numbers, want)
	}
}

func TestParseGoogleMeet(t *testing.T) {
	jt, err := Parse(readTestdata(t, "google_meet.txt"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	assertGoogleMeet(t, jt)
}

func TestParseICSGoogleMeet(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "google_meet.ics"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}()

	jt, err := ParseICS(f)
	if err != nil {
		t.Fatalf("ParseICS() error: %v", err)
	}
	assertGoogleMeet(t, jt)
}

func assertGoogleMeet(t *testing.T, jt attendee.JoinTarget) {
	t.Helper()
	if jt.Platform != attendee.PlatformGoogleMeet {
		t.Errorf("Platform = %q, want google_meet", jt.Platform)
	}
	if jt.JoinURL != "https://meet.google.com/abc-defg-hij" || jt.MeetingID != "abc-defg-hij" {
		t.Errorf("JoinURL/MeetingID = %q/%q", jt.JoinURL, jt.MeetingID)
	}
	want := &attendee.DialInDetails{
		Numbers:      []attendee.PhoneNumber{{Number: "+14135550123", Country: "US"}},
		ConferenceID: "123456789",
	}
	if !reflect.DeepEqual(jt.DialIn, want) {
		t.Errorf("DialIn = %+v, want %+v", jt.DialIn, want)
	}
}

func TestParseTeams(t *testing.T) {
	jt, err := Parse(readTestdata(t, "teams.txt"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if jt.Platform != attendee.PlatformTeams {
		t.Errorf("Platform = %q, want teams", jt.Platform)
	}
	if !strings.HasPrefix(jt.JoinURL, "https://teams.microsoft.com/l/meetup-join/") {
		t.Errorf("JoinURL = %q", jt.JoinURL)
	}
	if jt.MeetingID != "123456789012" || jt.Passcode != "a1B2c3" {
		t.Errorf("MeetingID/Passcode = %q/%q", jt.MeetingID, jt.Passcode)
	}
	if jt.DialIn == nil {
		t.Fatal("DialIn = nil")
	}
	if jt.DialIn.ConferenceID != "987654321" {
		t.Errorf("ConferenceID = %q, want the phone conference ID 987654321", jt.DialIn.ConferenceID)
	}
	want := []attendee.PhoneNumber{
		{Number: "+13235550123", Country: "US", Label: "Los Angeles"},
		{Number: "+14165550142", Country: "CA", Label: "Toronto"},
	}
	if !reflect.DeepEqual(jt.DialIn.Numbers, want) {
		t.Errorf("Numbers =\n%+v\nwant\n%+v", jt.DialIn.Numbers, want)
	}
}

// TestParsedInvitesDial checks that each parsed invite yields a keypad
// sequence for its platform.
func TestParsedInvitesDial(t *testing.T) {
	tests := map[string]string{
		"zoom.txt":        "WWWW85012345678#WWWWWWWW*482913#",
		"google_meet.txt": "WWWW123456789#",
		"teams.txt":       "WWWW987654321#",
	}
	for file, want := range tests {
		t.Run(file, func(t *testing.T) {
			jt, err := Parse(readTestdata(t, file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			target, err := dialin.Choose(jt, "US")
			if err != nil {
				t.Fatalf("Choose() error: %v", err)
			}
			got, err := dialin.DialString(target)
			if err != nil {
				t.Fatalf("DialString() error: %v", err)
			}
			if got != want {
				t.Fatalf("DialString() = %q, want %q", got, want)
			}
		})
	}
}

func TestParseLinkWithoutDialIn(t *testing.T) {
	jt, err := Parse("Join: https://meet.google.com/xyz-abcd-efg")
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if jt.Platform != attendee.PlatformGoogleMeet || jt.DialIn != nil {
		t.Fatalf("Parse() = %+v, want Google Meet without dial-in", jt)
	}
	if _, err := dialin.Choose(jt, "US"); !errors.Is(err, dialin.ErrNoDialIn) {
		t.Fatalf("Choose() error = %v, want ErrNoDialIn", err)
	}
}

func TestParseNoMeeting(t *testing.T) {
	inputs := []string{
		"",
		"Lunch at noon. Call me at +1 415 555 0100.",
		"Help: https://support.google.com/meet and https://us02web.zoom.us/u/abc",
	}
	for _, in := range inputs {
		if _, err := Parse(in); !errors.Is(err, ErrNoMeeting) {
			t.Errorf("Parse(%q) error = %v, want ErrNoMeeting", in, err)
		}
	}
}

func TestParseICSNoEvent(t *testing.T) {
	_, err := ParseICS(strings.NewReader("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n"))
	if !errors.Is(err, ErrNoMeeting) {
		t.Fatalf("ParseICS() error = %v, want ErrNoMeeting", err)
	}
}

func TestSplitProperty(t *testing.T) {
	tests := []struct {
		line, name, value string
	}{
		{"DESCRIPTION:hello", "DESCRIPTION", "hello"},
		{"description;LANGUAGE=en-US:hi: there", "DESCRIPTION", "hi: there"},
		{`ATTENDEE;CN="Doe: Jane":mailto:jane@example.com`, "ATTENDEE", "mailto:jane@example.com"},
	}
	for _, tt := range tests {
		name, value, ok := splitProperty(tt.line)
		if !ok || name != tt.name || value != tt.value {
			t.Errorf("splitProperty(%q) = %q, %q, %v", tt.line, name, value, ok)
		}
	}
}
