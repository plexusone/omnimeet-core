package dialin

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/plexusone/omnivoice-core/audio/format"
	"github.com/plexusone/omnivoice-core/callsystem"
	"github.com/plexusone/omnivoice-core/gateway"

	"github.com/plexusone/omnimeet-core/attendee"
)

// fakeCall is a gateway session in audio mode whose events and audio the
// test controls.
type fakeCall struct {
	id     string
	events chan gateway.Event
	audio  chan []byte

	mu        sync.Mutex
	sent      [][]byte
	closed    bool
	hungUp    bool
	closeOnce sync.Once
}

func newFakeCall() *fakeCall {
	return &fakeCall{id: "CA-test", events: make(chan gateway.Event, 16), audio: make(chan []byte, 16)}
}

func (c *fakeCall) ID() string              { return c.id }
func (c *fakeCall) From() string            { return "+15550000000" }
func (c *fakeCall) To() string              { return "+16695550100" }
func (c *fakeCall) Direction() string       { return "outbound" }
func (c *fakeCall) StartTime() time.Time    { return time.Time{} }
func (c *fakeCall) Duration() time.Duration { return 0 }
func (c *fakeCall) Events() <-chan gateway.Event {
	return c.events
}
func (c *fakeCall) Transcript() []gateway.Turn      { return nil }
func (c *fakeCall) Metrics() gateway.Metrics        { return gateway.Metrics{} }
func (c *fakeCall) SendText(string) error           { return nil }
func (c *fakeCall) Interrupt()                      {}
func (c *fakeCall) AudioFormat() format.AudioFormat { return format.Twilio }
func (c *fakeCall) ReceiveAudio() <-chan []byte     { return c.audio }
func (c *fakeCall) ClearAudio() error               { return nil }

func (c *fakeCall) SendAudio(audio []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return gateway.ErrSessionClosed
	}
	c.sent = append(c.sent, audio)
	return nil
}

// Close ends the call the way a gateway does: events and audio channels close.
func (c *fakeCall) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		close(c.events)
		close(c.audio)
	})
	return nil
}

// hangingCall also implements Hangup.
type hangingCall struct {
	*fakeCall
}

func (c hangingCall) Hangup(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hungUp = true
	return nil
}

// textCall is a session of a gateway that is not in audio mode.
type textCall struct {
	*fakeCall
}

func (c textCall) AudioFormat() {} // hides the AudioSession methods

// fakeGateway records outbound calls and returns session.
type fakeGateway struct {
	session gateway.Session
	to      string
	opts    callsystem.CallOptions
	calls   int
}

func (g *fakeGateway) Name() gateway.ProviderName                { return "fake" }
func (g *fakeGateway) Start(context.Context) error               { return nil }
func (g *fakeGateway) Stop() error                               { return nil }
func (g *fakeGateway) OnCall(gateway.CallHandler)                {}
func (g *fakeGateway) GetSession(string) (gateway.Session, bool) { return nil, false }
func (g *fakeGateway) ListSessions() []gateway.Session           { return nil }
func (g *fakeGateway) MakeCall(ctx context.Context, to string) (gateway.Session, error) {
	return g.MakeCallWithOptions(ctx, to)
}

func (g *fakeGateway) MakeCallWithOptions(_ context.Context, to string, opts ...callsystem.CallOption) (gateway.Session, error) {
	g.calls++
	g.to = to
	for _, opt := range opts {
		opt(&g.opts)
	}
	return g.session, nil
}

// noOptionsGateway does not implement gateway.OptionCaller.
type noOptionsGateway struct {
	gateway.Gateway
}

var testTiming = Timing{Answer: 10 * time.Millisecond, Between: 10 * time.Millisecond}

func meetTarget() attendee.JoinTarget {
	return attendee.JoinTarget{
		Platform: attendee.PlatformGoogleMeet,
		JoinURL:  "https://meet.google.com/abc-defg-hij",
		DialIn: &attendee.DialInDetails{
			Numbers: []attendee.PhoneNumber{
				{Number: "+442079460100", Country: "GB"},
				{Number: "+14135550123", Country: "US"},
			},
			ConferenceID: "123456789",
		},
	}
}

func newTestJoiner(gw gateway.Gateway) *Joiner {
	j := NewJoiner(gw, WithCountry("US"), WithTiming(testTiming))
	j.keyDuration = 0
	return j
}

// waitState waits for the session to report want.
func waitState(t *testing.T, s attendee.Session, want attendee.State) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s.State() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("state = %s, want %s", s.State(), want)
}

// collectStates reads state changes until the events channel closes.
func collectStates(t *testing.T, s attendee.Session) []attendee.State {
	t.Helper()
	var states []attendee.State
	timeout := time.After(2 * time.Second)
	for {
		select {
		case ev, ok := <-s.Events():
			if !ok {
				return states
			}
			if ev.Type == attendee.EventStateChanged {
				states = append(states, ev.State)
			}
		case <-timeout:
			t.Fatalf("events not closed; states so far %v", states)
		}
	}
}

func TestJoinDialsNumberWithKeypadDigits(t *testing.T) {
	gw := &fakeGateway{session: newFakeCall()}
	s, err := newTestJoiner(gw).Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}

	if gw.to != "+14135550123" {
		t.Errorf("dialed %q, want the US number", gw.to)
	}
	want := Render([]Step{{Pause: testTiming.Answer, Digits: "123456789#"}})
	if gw.opts.SendDigits != want {
		t.Errorf("SendDigits = %q, want %q", gw.opts.SendDigits, want)
	}
	if s.State() != attendee.StateJoining {
		t.Errorf("state after Join = %s, want joining", s.State())
	}
	if s.ID() != "CA-test" || s.Target().JoinURL != "https://meet.google.com/abc-defg-hij" {
		t.Errorf("ID/Target = %q/%+v", s.ID(), s.Target())
	}
	caps := s.Capabilities()
	if !caps.ReceiveAudio || !caps.SendAudio || caps.DisplayName || caps.SpeakerNames || caps.Participants {
		t.Errorf("Capabilities() = %+v, want audio only", caps)
	}
}

func TestSessionLifecycleWhenMeetingEnds(t *testing.T) {
	call := newFakeCall()
	s, err := newTestJoiner(&fakeGateway{session: call}).Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}

	call.events <- gateway.Event{Type: gateway.EventSessionStarted}
	waitState(t, s, attendee.StateAdmitted)

	call.events <- gateway.Event{Type: gateway.EventSessionEnded}
	states := collectStates(t, s)
	want := []attendee.State{attendee.StateJoining, attendee.StateConnected, attendee.StateAdmitted, attendee.StateEnded}
	if len(states) != len(want) {
		t.Fatalf("states = %v, want %v", states, want)
	}
	for i := range want {
		if states[i] != want[i] {
			t.Fatalf("states = %v, want %v", states, want)
		}
	}
}

func TestAdmittedWaitsForKeypadSequence(t *testing.T) {
	call := newFakeCall()
	j := newTestJoiner(&fakeGateway{session: call})
	j.timing = Timing{Answer: 300 * time.Millisecond}
	s, err := j.Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}

	call.events <- gateway.Event{Type: gateway.EventSessionStarted}
	waitState(t, s, attendee.StateConnected)
	time.Sleep(50 * time.Millisecond)
	if s.State() != attendee.StateConnected {
		t.Fatalf("state = %s before the keypad sequence finished, want connected", s.State())
	}
	waitState(t, s, attendee.StateAdmitted)
}

func TestSessionAudio(t *testing.T) {
	call := newFakeCall()
	s, err := newTestJoiner(&fakeGateway{session: call}).Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}
	rx, ok := s.(attendee.AudioReceiver)
	if !ok {
		t.Fatal("session does not implement AudioReceiver")
	}
	tx, ok := s.(attendee.AudioSender)
	if !ok {
		t.Fatal("session does not implement AudioSender")
	}
	if rx.ReceiveFormat() != format.Twilio || tx.SendFormat() != format.Twilio {
		t.Fatalf("formats = %+v / %+v", rx.ReceiveFormat(), tx.SendFormat())
	}

	if _, err := tx.SendAudio().Write([]byte{9, 8}); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	call.mu.Lock()
	if len(call.sent) != 1 || !bytes.Equal(call.sent[0], []byte{9, 8}) {
		t.Fatalf("sent = %v", call.sent)
	}
	call.mu.Unlock()

	call.audio <- []byte{1, 2, 3}
	call.audio <- []byte{4}
	if err := call.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	got, err := io.ReadAll(rx.ReceiveAudio())
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if !bytes.Equal(got, []byte{1, 2, 3, 4}) {
		t.Fatalf("received %v, want [1 2 3 4]", got)
	}

	if _, err := tx.SendAudio().Write([]byte{1}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Write after call ended = %v, want io.ErrClosedPipe", err)
	}
}

func TestLeaveHangsUp(t *testing.T) {
	call := newFakeCall()
	s, err := newTestJoiner(&fakeGateway{session: hangingCall{call}}).Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}
	call.events <- gateway.Event{Type: gateway.EventSessionStarted}
	waitState(t, s, attendee.StateAdmitted)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Leave(ctx); err != nil {
		t.Fatalf("Leave() error: %v", err)
	}
	if s.State() != attendee.StateLeft {
		t.Fatalf("state after Leave = %s, want left", s.State())
	}
	call.mu.Lock()
	if !call.hungUp || !call.closed {
		t.Fatalf("hungUp=%v closed=%v, want both", call.hungUp, call.closed)
	}
	call.mu.Unlock()
	if err := s.Leave(ctx); err != nil {
		t.Fatalf("second Leave() error: %v", err)
	}
}

func TestLeaveWithoutHangupClosesSession(t *testing.T) {
	call := newFakeCall()
	s, err := newTestJoiner(&fakeGateway{session: call}).Join(context.Background(), meetTarget())
	if err != nil {
		t.Fatalf("Join() error: %v", err)
	}
	if err := s.Leave(context.Background()); err != nil {
		t.Fatalf("Leave() error: %v", err)
	}
	if s.State() != attendee.StateLeft || !call.closed {
		t.Fatalf("state=%s closed=%v, want left and closed", s.State(), call.closed)
	}
}

func TestJoinRequiresAudioMode(t *testing.T) {
	call := newFakeCall()
	_, err := newTestJoiner(&fakeGateway{session: textCall{call}}).Join(context.Background(), meetTarget())
	if !errors.Is(err, ErrAudioModeRequired) {
		t.Fatalf("Join() error = %v, want ErrAudioModeRequired", err)
	}
	if !call.closed {
		t.Fatal("call left open after rejecting a non-audio gateway")
	}
}

func TestJoinErrorsBeforeDialing(t *testing.T) {
	gw := &fakeGateway{session: newFakeCall()}
	noDialIn := attendee.JoinTarget{Platform: attendee.PlatformGoogleMeet, JoinURL: "https://meet.google.com/abc-defg-hij"}
	if _, err := newTestJoiner(gw).Join(context.Background(), noDialIn); !errors.Is(err, ErrNoDialIn) {
		t.Fatalf("Join() without dial-in = %v, want ErrNoDialIn", err)
	}

	webex := meetTarget()
	webex.Platform = attendee.PlatformWebex
	if _, err := newTestJoiner(gw).Join(context.Background(), webex); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Join() for Webex = %v, want ErrUnsupportedPlatform", err)
	}
	if gw.calls != 0 {
		t.Fatalf("placed %d calls despite errors", gw.calls)
	}

	_, err := newTestJoiner(noOptionsGateway{gw}).Join(context.Background(), meetTarget())
	if !errors.Is(err, gateway.ErrCallOptionsUnsupported) {
		t.Fatalf("Join() through a gateway without call options = %v, want ErrCallOptionsUnsupported", err)
	}
}
