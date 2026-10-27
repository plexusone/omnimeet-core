package dialin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/plexusone/omnivoice-core/audio/format"
	"github.com/plexusone/omnivoice-core/gateway"

	"github.com/plexusone/omnimeet-core/attendee"
)

// eventBuffer is the number of session events held for a slow reader before
// further events are dropped. State always reports the current state.
const eventBuffer = 32

// phoneSession is a meeting attended by phone. It implements
// attendee.Session, attendee.AudioReceiver, and attendee.AudioSender.
type phoneSession struct {
	target  attendee.JoinTarget
	audio   gateway.AudioSession
	admitAt time.Time
	reader  *chunkReader
	events  chan attendee.Event
	done    chan struct{}

	mu      sync.Mutex
	state   attendee.State
	leaving bool
}

var (
	_ attendee.Session       = (*phoneSession)(nil)
	_ attendee.AudioReceiver = (*phoneSession)(nil)
	_ attendee.AudioSender   = (*phoneSession)(nil)
)

func newPhoneSession(target attendee.JoinTarget, audio gateway.AudioSession, admitAt time.Time) *phoneSession {
	s := &phoneSession{
		target:  target,
		audio:   audio,
		admitAt: admitAt,
		reader:  &chunkReader{ch: audio.ReceiveAudio()},
		events:  make(chan attendee.Event, eventBuffer),
		done:    make(chan struct{}),
		state:   attendee.StateJoining,
	}
	s.emit(attendee.Event{Type: attendee.EventStateChanged, State: attendee.StateJoining})
	return s
}

// ID returns the gateway session ID, which is the provider's call ID.
func (s *phoneSession) ID() string { return s.audio.ID() }

// Target returns the meeting this session joined.
func (s *phoneSession) Target() attendee.JoinTarget { return s.target }

// Capabilities reports audio in and out only: phone dial-in carries no
// display name, speaker names, participant list, chat, or video.
func (s *phoneSession) Capabilities() attendee.Capabilities {
	return attendee.Capabilities{ReceiveAudio: true, SendAudio: true}
}

// State returns the current state.
func (s *phoneSession) State() attendee.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Events returns session events; the channel is closed after the session
// reaches a terminal state.
func (s *phoneSession) Events() <-chan attendee.Event { return s.events }

// Leave hangs up and waits for the session to end, or for ctx to be done.
func (s *phoneSession) Leave(ctx context.Context) error {
	s.mu.Lock()
	if s.state.Terminal() {
		s.mu.Unlock()
		return nil
	}
	s.leaving = true
	s.mu.Unlock()

	if err := hangUp(ctx, s.audio); err != nil {
		return fmt.Errorf("leave meeting: %w", err)
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReceiveFormat returns the telephony audio format, such as mu-law 8 kHz.
func (s *phoneSession) ReceiveFormat() format.AudioFormat { return s.audio.AudioFormat() }

// ReceiveAudio returns the meeting's mixed audio. It supports a single
// reader and returns io.EOF after the call ends.
func (s *phoneSession) ReceiveAudio() io.Reader { return s.reader }

// SendFormat returns the format SendAudio expects.
func (s *phoneSession) SendFormat() format.AudioFormat { return s.audio.AudioFormat() }

// SendAudio returns a writer whose audio is played into the meeting.
func (s *phoneSession) SendAudio() io.Writer { return audioWriter{s.audio} }

// watch follows the gateway session's events until the call ends.
func (s *phoneSession) watch() {
	defer s.finish()

	var (
		timer *time.Timer
		admit <-chan time.Time
	)
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		select {
		case ev, ok := <-s.audio.Events():
			if !ok {
				return
			}
			switch ev.Type {
			case gateway.EventSessionStarted:
				s.setState(attendee.StateConnected)
				wait := time.Until(s.admitAt)
				if wait <= 0 {
					s.setState(attendee.StateAdmitted)
					continue
				}
				timer = time.NewTimer(wait)
				admit = timer.C
			case gateway.EventSessionEnded:
				return
			case gateway.EventError:
				err, ok := ev.Data.(error)
				if !ok {
					err = fmt.Errorf("gateway error: %v", ev.Data)
				}
				s.emit(attendee.Event{Type: attendee.EventError, Err: err})
			}
		case <-admit:
			admit = nil
			if s.State() == attendee.StateConnected {
				s.setState(attendee.StateAdmitted)
			}
		}
	}
}

// finish moves to a terminal state, Left after Leave and Ended otherwise,
// and closes the events channel.
func (s *phoneSession) finish() {
	s.mu.Lock()
	final := attendee.StateEnded
	if s.leaving {
		final = attendee.StateLeft
	}
	s.state = final
	s.mu.Unlock()

	s.emit(attendee.Event{Type: attendee.EventStateChanged, State: final})
	close(s.events)
	close(s.done)
}

func (s *phoneSession) setState(state attendee.State) {
	s.mu.Lock()
	if s.state == state || s.state.Terminal() {
		s.mu.Unlock()
		return
	}
	s.state = state
	s.mu.Unlock()
	s.emit(attendee.Event{Type: attendee.EventStateChanged, State: state})
}

// emit sends an event without blocking; events are dropped when the buffer
// is full. Only the constructor and the watch goroutine emit, so emit never
// races with closing the channel.
func (s *phoneSession) emit(ev attendee.Event) {
	ev.SessionID = s.audio.ID()
	ev.Time = time.Now()
	select {
	case s.events <- ev:
	default:
	}
}

// chunkReader reads audio chunks from a channel as a byte stream.
type chunkReader struct {
	ch  <-chan []byte
	buf []byte
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(r.buf) == 0 {
		chunk, ok := <-r.ch
		if !ok {
			return 0, io.EOF
		}
		r.buf = chunk
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// audioWriter plays written audio into the call.
type audioWriter struct {
	audio gateway.AudioSession
}

func (w audioWriter) Write(p []byte) (int, error) {
	if err := w.audio.SendAudio(append([]byte(nil), p...)); err != nil {
		if errors.Is(err, gateway.ErrSessionClosed) {
			return 0, fmt.Errorf("%w: %w", io.ErrClosedPipe, err)
		}
		return 0, err
	}
	return len(p), nil
}
