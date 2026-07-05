// Package provider defines the core provider interfaces for OmniMeet.
package provider

import (
	"context"
	"time"

	"github.com/plexusone/omnimeet-core/event"
	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/token"
	"github.com/plexusone/omnimeet-core/track"
)

// AgentParticipant represents an AI agent's participation in a meeting.
//
// While MeetingProvider handles control plane operations, AgentParticipant
// handles media plane operations: joining meetings, subscribing to audio,
// publishing audio responses, and receiving real-time events.
//
// Not all providers support AgentParticipant. Use SupportsAgentParticipation()
// to check if a provider supports this interface.
type AgentParticipant interface {
	// Join/leave
	// JoinMeeting joins the meeting as an agent participant.
	JoinMeeting(ctx context.Context, meetingID string, token *token.JoinToken) error

	// LeaveMeeting leaves the current meeting.
	LeaveMeeting(ctx context.Context) error

	// Audio handling
	// SubscribeToAudio subscribes to a specific participant's audio.
	SubscribeToAudio(ctx context.Context, participantID string) (<-chan AudioFrame, error)

	// SubscribeToAllAudio subscribes to all participants' audio (mixed or separate).
	SubscribeToAllAudio(ctx context.Context) (<-chan AudioFrame, error)

	// PublishAudio publishes an audio frame to the meeting.
	PublishAudio(ctx context.Context, frame AudioFrame) error

	// StartAudioTrack starts publishing audio and returns a writer for continuous streaming.
	StartAudioTrack(ctx context.Context, opts AudioTrackOptions) (AudioWriter, error)

	// StopAudioTrack stops publishing audio.
	StopAudioTrack(ctx context.Context) error

	// Track subscription
	// SubscribeToTrack subscribes to a specific track.
	SubscribeToTrack(ctx context.Context, trackID string, opts track.SubscribeOptions) error

	// UnsubscribeFromTrack unsubscribes from a track.
	UnsubscribeFromTrack(ctx context.Context, trackID string) error

	// Data messages
	// SendDataMessage sends a data message to all or specific participants.
	SendDataMessage(ctx context.Context, msg DataMessage) error

	// OnDataMessage registers a handler for incoming data messages.
	OnDataMessage(handler func(DataMessage))

	// Events
	// Events returns a channel of real-time events.
	Events() <-chan event.Event

	// OnParticipantJoined registers a handler for participant join events.
	OnParticipantJoined(handler func(participant.Participant))

	// OnParticipantLeft registers a handler for participant leave events.
	OnParticipantLeft(handler func(participant.Participant))

	// OnTrackPublished registers a handler for track publish events.
	OnTrackPublished(handler func(participant.Participant, track.Track))

	// OnTrackUnpublished registers a handler for track unpublish events.
	OnTrackUnpublished(handler func(participant.Participant, track.Track))

	// OnActiveSpeakerChanged registers a handler for active speaker changes.
	OnActiveSpeakerChanged(handler func([]participant.Participant))

	// State
	// Meeting returns the current meeting.
	Meeting() *meeting.Meeting

	// LocalParticipant returns the local (agent) participant.
	LocalParticipant() *participant.Participant

	// RemoteParticipants returns all remote participants.
	RemoteParticipants() []participant.Participant

	// GetParticipant returns a specific participant by ID.
	GetParticipant(participantID string) *participant.Participant

	// ConnectionState returns the current connection state.
	ConnectionState() ConnectionState
}

// AgentParticipantFactory creates AgentParticipant instances.
type AgentParticipantFactory interface {
	// CreateAgentParticipant creates a new AgentParticipant.
	CreateAgentParticipant(opts AgentParticipantOptions) (AgentParticipant, error)

	// SupportsAgentParticipation returns true if this provider supports agent participation.
	SupportsAgentParticipation() bool
}

// AgentParticipantOptions contains options for creating an AgentParticipant.
type AgentParticipantOptions struct {
	// Logger is an optional logger.
	// Logger *slog.Logger

	// AutoSubscribe automatically subscribes to all tracks.
	AutoSubscribe bool

	// AudioConfig specifies the audio configuration.
	AudioConfig AudioConfig
}

// AudioFrame represents a frame of audio data.
type AudioFrame struct {
	// ParticipantID is the ID of the participant who produced this audio.
	// Empty for outgoing audio.
	ParticipantID string

	// ParticipantName is the name of the participant.
	ParticipantName string

	// Data is the raw audio data (PCM16 format).
	Data []byte

	// SampleRate is the sample rate in Hz (e.g., 16000, 24000, 48000).
	SampleRate int

	// Channels is the number of audio channels (1 = mono, 2 = stereo).
	Channels int

	// Timestamp is when this frame was captured.
	Timestamp time.Time

	// SequenceNumber is the sequence number for ordering.
	SequenceNumber uint64
}

// AudioConfig specifies the audio configuration for an agent.
type AudioConfig struct {
	// SampleRate is the sample rate in Hz.
	SampleRate int

	// Channels is the number of channels.
	Channels int

	// FrameDuration is the duration of each audio frame.
	FrameDuration time.Duration
}

// DefaultAudioConfig returns the default audio configuration.
func DefaultAudioConfig() AudioConfig {
	return AudioConfig{
		SampleRate:    48000,
		Channels:      1,
		FrameDuration: 20 * time.Millisecond,
	}
}

// AudioTrackOptions contains options for starting an audio track.
type AudioTrackOptions struct {
	// Name is an optional name for the track.
	Name string

	// SampleRate is the sample rate in Hz.
	SampleRate int

	// Channels is the number of channels.
	Channels int
}

// AudioWriter is used to write continuous audio to a track.
type AudioWriter interface {
	// Write writes audio data to the track.
	Write(data []byte) (int, error)

	// Close stops writing and releases resources.
	Close() error
}

// DataMessage represents a data message sent between participants.
type DataMessage struct {
	// Topic is an optional topic/channel for the message.
	Topic string

	// Payload is the message payload.
	Payload []byte

	// DestinationIDs limits delivery to specific participants.
	// Empty means broadcast to all.
	DestinationIDs []string

	// Reliable indicates whether reliable delivery is required.
	Reliable bool

	// From is the sender (populated on receive).
	From *participant.Participant

	// Timestamp is when the message was sent.
	Timestamp time.Time
}

// ConnectionState represents the connection state.
type ConnectionState string

const (
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateConnecting   ConnectionState = "connecting"
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateReconnecting ConnectionState = "reconnecting"
	ConnectionStateFailed       ConnectionState = "failed"
)
