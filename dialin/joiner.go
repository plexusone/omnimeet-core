package dialin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/plexusone/omnivoice-core/callsystem"
	"github.com/plexusone/omnivoice-core/gateway"

	"github.com/plexusone/omnimeet-core/attendee"
)

// ErrAudioModeRequired is returned by Joiner.Join when the voice gateway does
// not run in gateway.PipelineModeAudio, so the meeting's audio would go to
// the gateway's own voice pipeline instead of the caller.
var ErrAudioModeRequired = errors.New("dial-in joiner requires a voice gateway in audio mode")

// defaultKeyDuration allows for each key tone when estimating how long the
// keypad sequence takes.
const defaultKeyDuration = 200 * time.Millisecond

// Joiner joins meetings by phone through an omnivoice-core voice gateway
// running in gateway.PipelineModeAudio, such as the omni-twilio or
// omni-telnyx gateway. The gateway must already be started so that it can
// receive the provider's webhooks and media stream.
//
// The meeting's audio arrives as one mixed stream of all participants, so
// sessions implement attendee.AudioReceiver and attendee.AudioSender only.
type Joiner struct {
	gw          gateway.Gateway
	country     string
	timing      Timing
	keyDuration time.Duration
}

var _ attendee.Joiner = (*Joiner)(nil)

// JoinerOption configures a Joiner.
type JoinerOption func(*Joiner)

// WithCountry picks dial-in numbers for an ISO 3166-1 alpha-2 country code,
// such as "US". By default the first number in the invite is used.
func WithCountry(country string) JoinerOption {
	return func(j *Joiner) {
		j.country = country
	}
}

// WithTiming sets the keypad pauses. The default is DefaultTiming.
func WithTiming(timing Timing) JoinerOption {
	return func(j *Joiner) {
		j.timing = timing
	}
}

// NewJoiner returns a Joiner that places calls through gw.
func NewJoiner(gw gateway.Gateway, opts ...JoinerOption) *Joiner {
	j := &Joiner{gw: gw, timing: DefaultTiming, keyDuration: defaultKeyDuration}
	for _, opt := range opts {
		opt(j)
	}
	return j
}

// Join dials the meeting's phone number and enters its keypad sequence once
// the call is answered. It returns when the call is placed; watch the
// session's events for StateConnected and StateAdmitted.
//
// Phone dial-in cannot see whether the meeting admitted the caller. The
// session reports StateAdmitted once the call is connected and the keypad
// sequence has had time to finish; a waiting room is not detected.
func (j *Joiner) Join(ctx context.Context, target attendee.JoinTarget) (attendee.Session, error) {
	dial, err := Choose(target, j.country)
	if err != nil {
		return nil, err
	}
	steps, err := Sequence(dial, j.timing)
	if err != nil {
		return nil, err
	}

	placed := time.Now()
	gs, err := gateway.MakeCallWithOptions(ctx, j.gw, dial.Number,
		callsystem.WithSendDigits(Render(steps)))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", dial.Number, err)
	}

	audio, ok := gs.(gateway.AudioSession)
	if !ok {
		if err := hangUp(ctx, gs); err != nil {
			return nil, fmt.Errorf("%w; ending the call: %w", ErrAudioModeRequired, err)
		}
		return nil, ErrAudioModeRequired
	}

	s := newPhoneSession(target, audio, placed.Add(sequenceDuration(steps, j.keyDuration)))
	go s.watch()
	return s, nil
}

// sequenceDuration estimates how long the keypad sequence takes to enter.
func sequenceDuration(steps []Step, keyDuration time.Duration) time.Duration {
	var d time.Duration
	for _, s := range steps {
		d += s.Pause + time.Duration(len(s.Digits))*keyDuration
	}
	return d
}

// hanger is implemented by gateway sessions that can end the call itself.
// Closing a session only stops its media stream, which on some providers
// leaves the call connected.
type hanger interface {
	Hangup(ctx context.Context) error
}

// hangUp ends the call when the session supports it, and closes the session.
func hangUp(ctx context.Context, gs gateway.Session) error {
	var hangupErr error
	if h, ok := gs.(hanger); ok {
		hangupErr = h.Hangup(ctx)
	}
	return errors.Join(hangupErr, gs.Close())
}
