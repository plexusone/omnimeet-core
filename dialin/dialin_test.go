package dialin

import (
	"errors"
	"testing"
	"time"

	"github.com/plexusone/omnivoice-core/callsystem"

	"github.com/plexusone/omnimeet-core/attendee"
)

func joinTarget(numbers ...attendee.PhoneNumber) attendee.JoinTarget {
	return attendee.JoinTarget{
		Platform: attendee.PlatformZoom,
		DialIn: &attendee.DialInDetails{
			Numbers:      numbers,
			ConferenceID: "85012345678",
			Passcode:     "123456",
		},
	}
}

func TestChoose(t *testing.T) {
	usTollFree := attendee.PhoneNumber{Number: "+18005550100", Country: "US", TollFree: true}
	usToll := attendee.PhoneNumber{Number: "+14155550100", Country: "US"}
	gb := attendee.PhoneNumber{Number: "+442079460000", Country: "GB"}

	tests := []struct {
		name    string
		target  attendee.JoinTarget
		country string
		want    string
		wantErr error
	}{
		{name: "first number without country", target: joinTarget(gb, usToll), want: "+442079460000"},
		{name: "toll number preferred", target: joinTarget(usTollFree, gb, usToll), country: "US", want: "+14155550100"},
		{name: "toll-free when only option", target: joinTarget(gb, usTollFree), country: "us", want: "+18005550100"},
		{name: "no number for country", target: joinTarget(gb), country: "US", wantErr: ErrNoNumberForCountry},
		{name: "no dial-in", target: attendee.JoinTarget{Platform: attendee.PlatformZoom}, wantErr: ErrNoDialIn},
		{name: "no numbers", target: joinTarget(), wantErr: ErrNoDialIn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Choose(tt.target, tt.country)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Choose() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Choose() error: %v", err)
			}
			if got.Number != tt.want {
				t.Fatalf("Number = %q, want %q", got.Number, tt.want)
			}
			if got.ConferenceID != "85012345678" || got.Passcode != "123456" || got.Platform != tt.target.Platform {
				t.Fatalf("Choose() = %+v, lost meeting details", got)
			}
		})
	}
}

func TestTargetValidate(t *testing.T) {
	valid := Target{Platform: attendee.PlatformZoom, Number: "+14155550100", ConferenceID: "85012345678", Passcode: "123456"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid target: %v", err)
	}

	invalid := []Target{
		{Number: "4155550100", ConferenceID: "1"},
		{Number: "+1 415 555 0100", ConferenceID: "1"},
		{Number: "+14155550100", ConferenceID: ""},
		{Number: "+14155550100", ConferenceID: "850 123"},
		{Number: "+14155550100", ConferenceID: "1", Passcode: "abc123"},
	}
	for _, tgt := range invalid {
		if err := tgt.Validate(); err == nil {
			t.Errorf("Validate(%+v) = nil, want error", tgt)
		}
	}
}

func TestDialString(t *testing.T) {
	tests := []struct {
		name   string
		target Target
		want   string
	}{
		{
			name:   "zoom with passcode",
			target: Target{Platform: attendee.PlatformZoom, Number: "+14155550100", ConferenceID: "85012345678", Passcode: "123456"},
			want:   "WWWW85012345678#WWWWWWWW*123456#",
		},
		{
			name:   "zoom without passcode",
			target: Target{Platform: attendee.PlatformZoom, Number: "+14155550100", ConferenceID: "85012345678"},
			want:   "WWWW85012345678#WWWWWWWW#",
		},
		{
			name:   "google meet",
			target: Target{Platform: attendee.PlatformGoogleMeet, Number: "+14135550100", ConferenceID: "123456789"},
			want:   "WWWW123456789#",
		},
		{
			name:   "teams",
			target: Target{Platform: attendee.PlatformTeams, Number: "+13235550100", ConferenceID: "123456789"},
			want:   "WWWW123456789#",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DialString(tt.target)
			if err != nil {
				t.Fatalf("DialString() error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("DialString() = %q, want %q", got, tt.want)
			}
			if err := callsystem.ValidateSendDigits(got); err != nil {
				t.Fatalf("dial string not accepted by callsystem.WithSendDigits: %v", err)
			}
		})
	}
}

func TestSequenceErrors(t *testing.T) {
	webex := Target{Platform: attendee.PlatformWebex, Number: "+14155550100", ConferenceID: "123456789"}
	if _, err := Sequence(webex, DefaultTiming); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("webex error = %v, want ErrUnsupportedPlatform", err)
	}

	meetWithPasscode := Target{Platform: attendee.PlatformGoogleMeet, Number: "+14135550100", ConferenceID: "123456789", Passcode: "1"}
	if _, err := Sequence(meetWithPasscode, DefaultTiming); err == nil {
		t.Fatal("Google Meet with a phone passcode returned nil error")
	}

	invalid := Target{Platform: attendee.PlatformZoom, Number: "not-a-number", ConferenceID: "1"}
	if _, err := Sequence(invalid, DefaultTiming); err == nil {
		t.Fatal("invalid target returned nil error")
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		steps []Step
		want  string
	}{
		{steps: nil, want: ""},
		{steps: []Step{{Digits: "1#"}}, want: "1#"},
		{steps: []Step{{Pause: 500 * time.Millisecond, Digits: "1"}}, want: "w1"},
		{steps: []Step{{Pause: 1500 * time.Millisecond, Digits: "1"}}, want: "Ww1"},
		{steps: []Step{{Pause: 1200 * time.Millisecond, Digits: "1"}}, want: "Ww1"},
		{steps: []Step{{Pause: time.Millisecond, Digits: "1"}}, want: "w1"},
		{steps: []Step{{Pause: 2 * time.Second, Digits: "1#"}, {Pause: time.Second, Digits: "#"}}, want: "WW1#W#"},
	}
	for _, tt := range tests {
		if got := Render(tt.steps); got != tt.want {
			t.Errorf("Render(%+v) = %q, want %q", tt.steps, got, tt.want)
		}
	}
}
