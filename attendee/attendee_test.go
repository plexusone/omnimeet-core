package attendee

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/plexusone/omnivoice-core/audio/format"
)

func TestStateTerminal(t *testing.T) {
	terminal := map[State]bool{
		StateJoining:   false,
		StateConnected: false,
		StateWaiting:   false,
		StateAdmitted:  false,
		StateLeft:      true,
		StateRemoved:   true,
		StateEnded:     true,
		StateFailed:    true,
	}
	for s, want := range terminal {
		if got := s.Terminal(); got != want {
			t.Errorf("%s.Terminal() = %v, want %v", s, got, want)
		}
	}
}

func TestPlatformString(t *testing.T) {
	if got := PlatformUnknown.String(); got != "unknown" {
		t.Errorf("PlatformUnknown.String() = %q, want %q", got, "unknown")
	}
	if got := PlatformGoogleMeet.String(); got != "google_meet" {
		t.Errorf("PlatformGoogleMeet.String() = %q, want %q", got, "google_meet")
	}
}

// audioOnlySession is a dial-in style session: audio in and out, nothing else.
type audioOnlySession struct {
	in  *bytes.Reader
	out bytes.Buffer
}

func (s *audioOnlySession) ID() string           { return "s1" }
func (s *audioOnlySession) Target() JoinTarget   { return JoinTarget{Platform: PlatformZoom} }
func (s *audioOnlySession) State() State         { return StateAdmitted }
func (s *audioOnlySession) Events() <-chan Event { return nil }
func (s *audioOnlySession) Leave(context.Context) error {
	return nil
}
func (s *audioOnlySession) Capabilities() Capabilities {
	return Capabilities{ReceiveAudio: true, SendAudio: true}
}
func (s *audioOnlySession) ReceiveFormat() format.AudioFormat { return format.Twilio }
func (s *audioOnlySession) ReceiveAudio() io.Reader           { return s.in }
func (s *audioOnlySession) SendFormat() format.AudioFormat    { return format.Twilio }
func (s *audioOnlySession) SendAudio() io.Writer              { return &s.out }

func TestOptionalInterfacesMatchCapabilities(t *testing.T) {
	var sess Session = &audioOnlySession{in: bytes.NewReader([]byte{1, 2, 3})}
	caps := sess.Capabilities()

	_, isReceiver := sess.(AudioReceiver)
	_, isSender := sess.(AudioSender)
	_, isRecorder := sess.(Recorder)
	_, isLister := sess.(ParticipantLister)

	if caps.ReceiveAudio != isReceiver {
		t.Errorf("ReceiveAudio = %v but AudioReceiver implemented = %v", caps.ReceiveAudio, isReceiver)
	}
	if caps.SendAudio != isSender {
		t.Errorf("SendAudio = %v but AudioSender implemented = %v", caps.SendAudio, isSender)
	}
	if caps.Recording != isRecorder {
		t.Errorf("Recording = %v but Recorder implemented = %v", caps.Recording, isRecorder)
	}
	if caps.Participants != isLister {
		t.Errorf("Participants = %v but ParticipantLister implemented = %v", caps.Participants, isLister)
	}
}
