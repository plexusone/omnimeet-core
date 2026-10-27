// Package dialintest simulates meetings on a phone line, so code that joins
// meetings by dial-in can be tested without a real meeting or telephony
// provider.
//
// A Gateway behaves like an omnivoice-core voice gateway in audio mode. Each
// Meeting answers one phone number and runs a phone menu like the platform's:
// it reads the keypad digits the caller sends, admits the caller only when the
// meeting ID and passcode are entered correctly, then plays a script of audio
// and records the audio the caller sends.
//
//	meeting := &dialintest.Meeting{
//		Platform:     attendee.PlatformZoom,
//		Number:       "+16695550100",
//		ConferenceID: "85012345678",
//		Passcode:     "482913",
//		Script:       [][]byte{greeting, discussion},
//	}
//	gw := dialintest.NewGateway(meeting)
//	joiner := dialin.NewJoiner(gw)
package dialintest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/plexusone/omnivoice-core/audio/format"
	"github.com/plexusone/omnivoice-core/callsystem"
	"github.com/plexusone/omnivoice-core/gateway"

	"github.com/plexusone/omnimeet-core/attendee"
)

// ErrNumberNotInService is returned when no simulated meeting answers the
// dialed number.
var ErrNumberNotInService = errors.New("number not in service")

// Meeting is a simulated meeting reachable by phone.
type Meeting struct {
	// Platform selects the phone menu: Zoom, Google Meet, or Teams.
	Platform attendee.Platform

	// Number is the dial-in number that reaches this meeting.
	Number string

	// ConferenceID is the meeting ID, PIN, or conference ID the menu expects.
	ConferenceID string

	// Passcode is the numeric phone passcode the menu expects, if any.
	Passcode string

	// Script is the meeting audio played to an admitted caller, in order.
	Script [][]byte

	// EndAfterScript ends the meeting once the script has played. Otherwise
	// the call stays up until the caller hangs up or End is called.
	EndAfterScript bool

	mu       sync.Mutex
	calls    []*Call
	received [][]byte
}

// Calls returns the calls placed to this meeting.
func (m *Meeting) Calls() []*Call {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*Call(nil), m.calls...)
}

// Received returns the audio callers sent into the meeting.
func (m *Meeting) Received() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]byte(nil), m.received...)
}

// End ends the meeting for every caller, as when the host ends it.
func (m *Meeting) End() {
	for _, c := range m.Calls() {
		c.end()
	}
}

// Gateway is a simulated voice gateway in audio mode. It implements
// gateway.Gateway and gateway.OptionCaller.
type Gateway struct {
	meetings map[string]*Meeting

	mu      sync.Mutex
	nextID  int
	handler gateway.CallHandler
}

var (
	_ gateway.Gateway      = (*Gateway)(nil)
	_ gateway.OptionCaller = (*Gateway)(nil)
)

// NewGateway returns a gateway whose calls reach the given meetings.
func NewGateway(meetings ...*Meeting) *Gateway {
	g := &Gateway{meetings: make(map[string]*Meeting, len(meetings))}
	for _, m := range meetings {
		g.meetings[m.Number] = m
	}
	return g
}

// Name returns "dialintest".
func (g *Gateway) Name() gateway.ProviderName { return "dialintest" }

// Start does nothing; the simulated gateway needs no server.
func (g *Gateway) Start(context.Context) error { return nil }

// Stop does nothing.
func (g *Gateway) Stop() error { return nil }

// OnCall stores the handler; the simulated gateway receives no calls.
func (g *Gateway) OnCall(handler gateway.CallHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.handler = handler
}

// GetSession reports no sessions; use Meeting.Calls instead.
func (g *Gateway) GetSession(string) (gateway.Session, bool) { return nil, false }

// ListSessions reports no sessions; use Meeting.Calls instead.
func (g *Gateway) ListSessions() []gateway.Session { return nil }

// MakeCall places a call without keypad digits.
func (g *Gateway) MakeCall(ctx context.Context, to string) (gateway.Session, error) {
	return g.MakeCallWithOptions(ctx, to)
}

// MakeCallWithOptions places a call to the meeting answering to. The call
// connects at once, runs the meeting's phone menu with the keypad digits
// from callsystem.WithSendDigits, and is admitted or hung up.
func (g *Gateway) MakeCallWithOptions(_ context.Context, to string, opts ...callsystem.CallOption) (gateway.Session, error) {
	var o callsystem.CallOptions
	for _, opt := range opts {
		opt(&o)
	}
	if err := callsystem.ValidateSendDigits(o.SendDigits); err != nil {
		return nil, err
	}
	m, ok := g.meetings[to]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNumberNotInService, to)
	}

	g.mu.Lock()
	g.nextID++
	id := fmt.Sprintf("SIM%04d", g.nextID)
	g.mu.Unlock()

	c := newCall(id, to, m, keysOnly(o.SendDigits))
	m.mu.Lock()
	m.calls = append(m.calls, c)
	m.mu.Unlock()

	go c.run()
	return c, nil
}

// keysOnly drops pause characters from a dial string.
func keysOnly(digits string) string {
	return strings.NewReplacer("w", "", "W", "").Replace(digits)
}

// Call is a simulated call to a meeting. It implements gateway.AudioSession.
type Call struct {
	id      string
	to      string
	meeting *Meeting
	keys    string
	started time.Time
	events  chan gateway.Event
	audio   chan []byte
	stop    chan struct{}

	mu       sync.Mutex
	admitted bool
	reason   string
	closed   bool
	hungUp   bool
	once     sync.Once
}

var _ gateway.AudioSession = (*Call)(nil)

func newCall(id, to string, m *Meeting, keys string) *Call {
	return &Call{
		id:      id,
		to:      to,
		meeting: m,
		keys:    keys,
		started: time.Now(),
		events:  make(chan gateway.Event, 16),
		audio:   make(chan []byte, len(m.Script)+1),
		stop:    make(chan struct{}),
	}
}

// Keys returns the keypad keys the caller entered, without pauses.
func (c *Call) Keys() string { return c.keys }

// Admitted reports whether the phone menu admitted the caller, and if not,
// why it hung up.
func (c *Call) Admitted() (bool, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.admitted, c.reason
}

// HungUp reports whether the caller hung up with Hangup.
func (c *Call) HungUp() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hungUp
}

// run connects the call, runs the phone menu, and plays the script. It is
// the only sender on the audio channel and closes it when the call ends.
func (c *Call) run() {
	defer close(c.audio)
	c.emit(gateway.EventSessionStarted, nil)

	reason := menu(c.meeting, c.keys)
	c.mu.Lock()
	c.admitted = reason == ""
	c.reason = reason
	c.mu.Unlock()

	if reason != "" {
		c.end()
		return
	}
	for _, chunk := range c.meeting.Script {
		select {
		case c.audio <- chunk:
		case <-c.stop:
			return
		}
	}
	if c.meeting.EndAfterScript {
		c.end()
		return
	}
	<-c.stop
}

// menu runs the platform's phone menu on keys and returns why the caller was
// hung up, or "" when admitted.
func menu(m *Meeting, keys string) string {
	entry, rest, ok := strings.Cut(keys, "#")
	if !ok {
		return "no # after the meeting ID"
	}
	if entry != m.ConferenceID {
		return fmt.Sprintf("wrong meeting ID %q", entry)
	}

	switch m.Platform {
	case attendee.PlatformZoom:
		// Participant ID prompt: "#" skips it, "*" starts the passcode.
		switch {
		case strings.HasPrefix(rest, "*"):
			pass, after, ok := strings.Cut(rest[1:], "#")
			if !ok || pass != m.Passcode {
				return fmt.Sprintf("wrong passcode %q", pass)
			}
			rest = after
		case strings.HasPrefix(rest, "#"):
			rest = rest[1:]
			if m.Passcode != "" {
				pass, after, ok := strings.Cut(rest, "#")
				if !ok || pass != m.Passcode {
					return "passcode required"
				}
				rest = after
			}
		default:
			return "no response at the participant ID prompt"
		}
	case attendee.PlatformGoogleMeet, attendee.PlatformTeams:
	default:
		return fmt.Sprintf("no phone menu for platform %q", m.Platform)
	}

	if rest != "" {
		return fmt.Sprintf("unexpected keys %q after joining", rest)
	}
	return ""
}

func (c *Call) emit(t gateway.EventType, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	select {
	case c.events <- gateway.Event{Type: t, Timestamp: time.Now(), Data: data}:
	default:
	}
}

// end ends the call from the meeting's side.
func (c *Call) end() {
	c.emit(gateway.EventSessionEnded, nil)
	c.close()
}

// ID returns the call ID.
func (c *Call) ID() string { return c.id }

// From returns a fixed caller number.
func (c *Call) From() string { return "+15550000000" }

// To returns the dialed number.
func (c *Call) To() string { return c.to }

// Direction returns "outbound".
func (c *Call) Direction() string { return "outbound" }

// StartTime returns when the call was placed.
func (c *Call) StartTime() time.Time { return c.started }

// Duration returns how long the call has lasted.
func (c *Call) Duration() time.Duration { return time.Since(c.started) }

// Events returns the call's gateway events.
func (c *Call) Events() <-chan gateway.Event { return c.events }

// Transcript returns nil; audio mode keeps no transcript.
func (c *Call) Transcript() []gateway.Turn { return nil }

// Metrics returns empty metrics.
func (c *Call) Metrics() gateway.Metrics { return gateway.Metrics{} }

// SendText is not supported in audio mode and does nothing.
func (c *Call) SendText(string) error { return nil }

// Interrupt does nothing in audio mode.
func (c *Call) Interrupt() {}

// AudioFormat returns mu-law 8 kHz mono, as telephony gateways do.
func (c *Call) AudioFormat() format.AudioFormat { return format.Twilio }

// ReceiveAudio returns the meeting's script audio.
func (c *Call) ReceiveAudio() <-chan []byte { return c.audio }

// SendAudio records audio into the meeting.
func (c *Call) SendAudio(audio []byte) error {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return gateway.ErrSessionClosed
	}
	c.meeting.mu.Lock()
	c.meeting.received = append(c.meeting.received, append([]byte(nil), audio...))
	c.meeting.mu.Unlock()
	return nil
}

// ClearAudio does nothing; audio is recorded as soon as it is sent.
func (c *Call) ClearAudio() error { return nil }

// Hangup ends the call from the caller's side.
func (c *Call) Hangup(context.Context) error {
	c.mu.Lock()
	c.hungUp = true
	c.mu.Unlock()
	c.close()
	return nil
}

// Close ends the call. The events channel closes at once and the audio
// channel once the meeting stops playing.
func (c *Call) Close() error {
	c.close()
	return nil
}

func (c *Call) close() {
	c.once.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.events)
		c.mu.Unlock()
		close(c.stop)
	})
}
