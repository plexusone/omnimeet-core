package attendee

import (
	"context"
	"io"
	"time"

	"github.com/plexusone/omnivoice-core/audio/format"

	"github.com/plexusone/omnimeet-core/participant"
)

// State is where a session is in joining and leaving a meeting.
type State string

const (
	// StateJoining means the connector is placing the call or opening the
	// meeting page.
	StateJoining State = "joining"
	// StateConnected means the connection is up but the agent is not in the
	// meeting yet, for example while entering keypad digits.
	StateConnected State = "connected"
	// StateWaiting means the agent is in a waiting room or the host has not
	// started the meeting.
	StateWaiting State = "waiting"
	// StateAdmitted means the agent is in the meeting.
	StateAdmitted State = "admitted"
	// StateLeft means the agent left the meeting.
	StateLeft State = "left"
	// StateRemoved means a host removed the agent or denied admission.
	StateRemoved State = "removed"
	// StateEnded means the meeting ended.
	StateEnded State = "ended"
	// StateFailed means joining failed or the connection dropped.
	StateFailed State = "failed"
)

// Terminal reports whether the session is over and will not change state
// again.
func (s State) Terminal() bool {
	switch s {
	case StateLeft, StateRemoved, StateEnded, StateFailed:
		return true
	default:
		return false
	}
}

// Capabilities describes what a session supports. A connector sets only what
// the connection actually provides; for example, dial-in sets ReceiveAudio and
// SendAudio only.
type Capabilities struct {
	// ReceiveAudio means the session implements AudioReceiver.
	ReceiveAudio bool `json:"receive_audio"`
	// SendAudio means the session implements AudioSender.
	SendAudio bool `json:"send_audio"`
	// DisplayName means other participants see JoinTarget.DisplayName.
	DisplayName bool `json:"display_name"`
	// SpeakerNames means events or transcripts name who is speaking.
	SpeakerNames bool `json:"speaker_names"`
	// Participants means the session implements ParticipantLister.
	Participants bool `json:"participants"`
	// Recording means the session implements Recorder.
	Recording bool `json:"recording"`
	// Chat means the session can read and send meeting chat.
	Chat bool `json:"chat"`
	// Video means the session can receive video.
	Video bool `json:"video"`
}

// Session is an agent's presence in one meeting.
type Session interface {
	// ID returns the session identifier, unique within the connector.
	ID() string

	// Target returns the meeting this session joined.
	Target() JoinTarget

	// Capabilities returns what this session supports.
	Capabilities() Capabilities

	// State returns the current state.
	State() State

	// Events returns session events. The channel is closed after the
	// session reaches a terminal state.
	Events() <-chan Event

	// Leave leaves the meeting. Leaving a session that is already in a
	// terminal state returns nil.
	Leave(ctx context.Context) error
}

// AudioReceiver is implemented by sessions that hear the meeting.
type AudioReceiver interface {
	// ReceiveFormat returns the format of the audio from ReceiveAudio.
	ReceiveFormat() format.AudioFormat

	// ReceiveAudio returns the meeting's audio. On dial-in it is a single
	// mixed stream of all participants. Reads return io.EOF after the
	// session ends.
	ReceiveAudio() io.Reader
}

// AudioSender is implemented by sessions that can speak into the meeting.
type AudioSender interface {
	// SendFormat returns the format SendAudio expects.
	SendFormat() format.AudioFormat

	// SendAudio returns a writer whose audio is played into the meeting.
	SendAudio() io.Writer
}

// Recorder is implemented by sessions that can control the meeting's own
// recording. Check recording consent before calling StartRecording.
type Recorder interface {
	// StartRecording starts recording the meeting.
	StartRecording(ctx context.Context) error

	// StopRecording stops recording the meeting.
	StopRecording(ctx context.Context) error
}

// ParticipantLister is implemented by sessions that can see who is in the
// meeting.
type ParticipantLister interface {
	// Participants returns the people and agents currently in the meeting.
	Participants(ctx context.Context) ([]participant.Participant, error)
}

// EventType identifies a session event.
type EventType string

const (
	// EventStateChanged reports a new State in Event.State.
	EventStateChanged EventType = "attendee.state_changed"
	// EventError reports a non-fatal problem in Event.Err, for example a
	// failed keypad entry that the connector will retry.
	EventError EventType = "attendee.error"
)

// Event is a session event.
type Event struct {
	// Type is the event type.
	Type EventType `json:"type"`

	// SessionID is the session that produced the event.
	SessionID string `json:"session_id"`

	// State is the new state, for EventStateChanged.
	State State `json:"state,omitempty"`

	// Err is the problem, for EventError, and the reason for StateFailed.
	Err error `json:"-"`

	// Time is when the event happened.
	Time time.Time `json:"time"`
}
