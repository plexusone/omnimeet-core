package dialintest_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plexusone/omnivoice-core/callsystem"

	"github.com/plexusone/omnimeet-core/attendee"
	"github.com/plexusone/omnimeet-core/dialin"
	"github.com/plexusone/omnimeet-core/dialin/dialintest"
	"github.com/plexusone/omnimeet-core/invite"
)

var fastTiming = dialin.Timing{Answer: time.Millisecond, Between: time.Millisecond}

func parseFixture(t *testing.T, name string) attendee.JoinTarget {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("..", "..", "invite", "testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	jt, err := invite.Parse(string(text))
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return jt
}

func waitState(t *testing.T, s attendee.Session, want attendee.State) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s.State() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("state = %s, want %s", s.State(), want)
}

// TestJoinFromInvites parses each invite fixture, joins its meeting by
// phone, and checks that the simulated phone menu admits the caller.
func TestJoinFromInvites(t *testing.T) {
	tests := []struct {
		fixture string
		meeting *dialintest.Meeting
	}{
		{"zoom.txt", &dialintest.Meeting{
			Platform: attendee.PlatformZoom, Number: "+16695550100",
			ConferenceID: "85012345678", Passcode: "482913",
		}},
		{"google_meet.txt", &dialintest.Meeting{
			Platform: attendee.PlatformGoogleMeet, Number: "+14135550123", ConferenceID: "123456789",
		}},
		{"teams.txt", &dialintest.Meeting{
			Platform: attendee.PlatformTeams, Number: "+13235550123", ConferenceID: "987654321",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			tt.meeting.Script = [][]byte{{1, 2}, {3}}
			gw := dialintest.NewGateway(tt.meeting)
			joiner := dialin.NewJoiner(gw, dialin.WithCountry("US"), dialin.WithTiming(fastTiming), dialin.WithKeyDuration(0))

			s, err := joiner.Join(context.Background(), parseFixture(t, tt.fixture))
			if err != nil {
				t.Fatalf("Join() error: %v", err)
			}
			waitState(t, s, attendee.StateAdmitted)

			calls := tt.meeting.Calls()
			if len(calls) != 1 {
				t.Fatalf("calls = %d, want 1", len(calls))
			}
			if ok, reason := calls[0].Admitted(); !ok {
				t.Fatalf("phone menu rejected keys %q: %s", calls[0].Keys(), reason)
			}

			// The agent hears the meeting and can speak into it.
			buf := make([]byte, 3)
			if _, err := io.ReadFull(s.(attendee.AudioReceiver).ReceiveAudio(), buf); err != nil {
				t.Fatalf("read meeting audio: %v", err)
			}
			if !bytes.Equal(buf, []byte{1, 2, 3}) {
				t.Fatalf("meeting audio = %v, want [1 2 3]", buf)
			}
			if _, err := s.(attendee.AudioSender).SendAudio().Write([]byte{7}); err != nil {
				t.Fatalf("send audio: %v", err)
			}
			if got := tt.meeting.Received(); len(got) != 1 || !bytes.Equal(got[0], []byte{7}) {
				t.Fatalf("meeting received %v, want [[7]]", got)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.Leave(ctx); err != nil {
				t.Fatalf("Leave() error: %v", err)
			}
			if s.State() != attendee.StateLeft || !calls[0].HungUp() {
				t.Fatalf("state=%s hungUp=%v, want left and hung up", s.State(), calls[0].HungUp())
			}
		})
	}
}

func TestWrongPasscodeIsHungUp(t *testing.T) {
	meeting := &dialintest.Meeting{
		Platform: attendee.PlatformZoom, Number: "+16695550100",
		ConferenceID: "85012345678", Passcode: "000000",
	}
	joiner := dialin.NewJoiner(dialintest.NewGateway(meeting), dialin.WithCountry("US"), dialin.WithTiming(fastTiming), dialin.WithKeyDuration(0))

	s, err := joiner.Join(context.Background(), parseFixture(t, "zoom.txt"))
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}
	waitState(t, s, attendee.StateEnded)
	if ok, reason := meeting.Calls()[0].Admitted(); ok || reason == "" {
		t.Fatalf("Admitted() = %v, %q; want rejected with a reason", ok, reason)
	}
}

func TestMeetingEnds(t *testing.T) {
	meeting := &dialintest.Meeting{
		Platform: attendee.PlatformGoogleMeet, Number: "+14135550123", ConferenceID: "123456789",
		Script: [][]byte{{5}}, EndAfterScript: true,
	}
	joiner := dialin.NewJoiner(dialintest.NewGateway(meeting), dialin.WithCountry("US"), dialin.WithTiming(fastTiming), dialin.WithKeyDuration(0))

	s, err := joiner.Join(context.Background(), parseFixture(t, "google_meet.txt"))
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}
	audio, err := io.ReadAll(s.(attendee.AudioReceiver).ReceiveAudio())
	if err != nil {
		t.Fatalf("read meeting audio: %v", err)
	}
	if !bytes.Equal(audio, []byte{5}) {
		t.Fatalf("meeting audio = %v, want [5]", audio)
	}
	waitState(t, s, attendee.StateEnded)
}

func TestMenus(t *testing.T) {
	tests := []struct {
		name    string
		meeting *dialintest.Meeting
		keys    string
		admit   bool
	}{
		{"zoom passcode at participant prompt", &dialintest.Meeting{Platform: attendee.PlatformZoom, ConferenceID: "1", Passcode: "22"}, "1#*22#", true},
		{"zoom skip then passcode", &dialintest.Meeting{Platform: attendee.PlatformZoom, ConferenceID: "1", Passcode: "22"}, "1##22#", true},
		{"zoom no passcode", &dialintest.Meeting{Platform: attendee.PlatformZoom, ConferenceID: "1"}, "1##", true},
		{"zoom missing passcode", &dialintest.Meeting{Platform: attendee.PlatformZoom, ConferenceID: "1", Passcode: "22"}, "1##", false},
		{"zoom no participant response", &dialintest.Meeting{Platform: attendee.PlatformZoom, ConferenceID: "1"}, "1#", false},
		{"wrong meeting ID", &dialintest.Meeting{Platform: attendee.PlatformTeams, ConferenceID: "1"}, "2#", false},
		{"no pound", &dialintest.Meeting{Platform: attendee.PlatformGoogleMeet, ConferenceID: "1"}, "1", false},
		{"extra keys", &dialintest.Meeting{Platform: attendee.PlatformGoogleMeet, ConferenceID: "1"}, "1#9", false},
		{"pauses ignored", &dialintest.Meeting{Platform: attendee.PlatformTeams, ConferenceID: "12"}, "WW1w2#", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.meeting
			m.Number = "+15550000001"
			gw := dialintest.NewGateway(m)
			sess, err := gw.MakeCallWithOptions(context.Background(), m.Number, callsystem.WithSendDigits(tt.keys))
			if err != nil {
				t.Fatalf("MakeCallWithOptions() error: %v", err)
			}
			call := sess.(*dialintest.Call)
			deadline := time.Now().Add(2 * time.Second)
			for {
				ok, reason := call.Admitted()
				if ok || reason != "" {
					if ok != tt.admit {
						t.Fatalf("admitted = %v (%s), want %v", ok, reason, tt.admit)
					}
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("phone menu did not finish")
				}
				time.Sleep(time.Millisecond)
			}
			if err := call.Close(); err != nil {
				t.Fatalf("Close() error: %v", err)
			}
		})
	}
}

func TestUnknownNumber(t *testing.T) {
	gw := dialintest.NewGateway()
	if _, err := gw.MakeCall(context.Background(), "+15550000009"); !errors.Is(err, dialintest.ErrNumberNotInService) {
		t.Fatalf("MakeCall() error = %v, want ErrNumberNotInService", err)
	}
}
