// Package omnimeet provides a unified abstraction for real-time collaboration platforms.
package omnimeet

import (
	"github.com/plexusone/omnimeet-core/event"
	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/provider"
	"github.com/plexusone/omnimeet-core/recording"
	"github.com/plexusone/omnimeet-core/token"
	"github.com/plexusone/omnimeet-core/track"
	"github.com/plexusone/omnimeet-core/transcript"
	"github.com/plexusone/omnimeet-core/voice"
	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/tts"
)

// Re-export core types for convenience.
// Users can import just "github.com/plexusone/omnimeet-core" and access all types.

// Meeting types
type (
	// Meeting represents a live collaborative session.
	Meeting = meeting.Meeting
	// MeetingStatus represents the current state of a meeting.
	MeetingStatus = meeting.Status
	// CreateMeetingRequest contains the parameters for creating a new meeting.
	CreateMeetingRequest = meeting.CreateRequest
	// ListMeetingsOptions contains options for listing meetings.
	ListMeetingsOptions = meeting.ListOptions
)

// Meeting status constants
const (
	MeetingStatusPending = meeting.StatusPending
	MeetingStatusActive  = meeting.StatusActive
	MeetingStatusEnded   = meeting.StatusEnded
)

// Participant types
type (
	// Participant represents an entity that joins a meeting.
	Participant = participant.Participant
	// ParticipantKind represents the type of participant.
	ParticipantKind = participant.Kind
	// ParticipantInfo contains information for creating or identifying a participant.
	ParticipantInfo = participant.Info
	// ParticipantPermissions defines what a participant is allowed to do.
	ParticipantPermissions = participant.Permissions
	// ConnectionQuality indicates the quality of a participant's connection.
	ConnectionQuality = participant.ConnectionQuality
)

// Participant kind constants
const (
	ParticipantKindHuman    = participant.KindHuman
	ParticipantKindAgent    = participant.KindAgent
	ParticipantKindRecorder = participant.KindRecorder
	ParticipantKindObserver = participant.KindObserver
	ParticipantKindSIP      = participant.KindSIP
)

// Connection quality constants
const (
	ConnectionQualityUnknown   = participant.ConnectionQualityUnknown
	ConnectionQualityExcellent = participant.ConnectionQualityExcellent
	ConnectionQualityGood      = participant.ConnectionQualityGood
	ConnectionQualityPoor      = participant.ConnectionQualityPoor
	ConnectionQualityLost      = participant.ConnectionQualityLost
)

// Track types
type (
	// Track represents a media stream published by a participant.
	Track = track.Track
	// TrackKind represents the type of media track.
	TrackKind = track.Kind
	// TrackSource represents the source of a track.
	TrackSource = track.Source
	// VideoQuality represents the quality level for video tracks.
	VideoQuality = track.VideoQuality
)

// Track kind constants
const (
	TrackKindAudio            = track.KindAudio
	TrackKindVideo            = track.KindVideo
	TrackKindScreenShare      = track.KindScreenShare
	TrackKindScreenShareAudio = track.KindScreenShareAudio
	TrackKindData             = track.KindData
)

// Track source constants
const (
	TrackSourceMicrophone  = track.SourceMicrophone
	TrackSourceCamera      = track.SourceCamera
	TrackSourceScreen      = track.SourceScreen
	TrackSourceApplication = track.SourceApplication
	TrackSourceUnknown     = track.SourceUnknown
)

// Token types
type (
	// JoinToken represents an access token for joining a meeting.
	JoinToken = token.JoinToken
	// CreateJoinTokenRequest contains the parameters for creating a join token.
	CreateJoinTokenRequest = token.CreateRequest
)

// Event types
type (
	// Event represents an event that occurred in a meeting.
	Event = event.Event
	// EventType represents the type of event.
	EventType = event.Type
	// EventHandler is a function that handles events.
	EventHandler = event.Handler
)

// Event type constants
const (
	EventMeetingCreated          = event.TypeMeetingCreated
	EventMeetingStarted          = event.TypeMeetingStarted
	EventMeetingEnded            = event.TypeMeetingEnded
	EventParticipantJoined       = event.TypeParticipantJoined
	EventParticipantLeft         = event.TypeParticipantLeft
	EventParticipantUpdated      = event.TypeParticipantUpdated
	EventTrackPublished          = event.TypeTrackPublished
	EventTrackUnpublished        = event.TypeTrackUnpublished
	EventTrackMuted              = event.TypeTrackMuted
	EventTrackUnmuted            = event.TypeTrackUnmuted
	EventActiveSpeakerChanged    = event.TypeActiveSpeakerChanged
	EventRecordingStarted        = event.TypeRecordingStarted
	EventRecordingStopped        = event.TypeRecordingStopped
	EventTranscriptUpdated       = event.TypeTranscriptUpdated
	EventDataMessageReceived     = event.TypeDataMessageReceived
	EventConnectionQualityChanged = event.TypeConnectionQualityChanged
)

// Recording types
type (
	// Recording represents a meeting recording.
	Recording = recording.Recording
	// RecordingStatus represents the status of a recording.
	RecordingStatus = recording.Status
	// RecordingOptions contains options for starting a recording.
	RecordingOptions = recording.Options
)

// Recording status constants
const (
	RecordingStatusPending    = recording.StatusPending
	RecordingStatusRecording  = recording.StatusRecording
	RecordingStatusProcessing = recording.StatusProcessing
	RecordingStatusCompleted  = recording.StatusCompleted
	RecordingStatusFailed     = recording.StatusFailed
)

// Transcript types
type (
	// Transcript represents a meeting transcript.
	Transcript = transcript.Transcript
	// TranscriptSegment represents a segment of a transcript.
	TranscriptSegment = transcript.Segment
	// TranscriptWord represents a word with timing information.
	TranscriptWord = transcript.Word
	// TranscriptConfig contains configuration for transcription.
	TranscriptConfig = transcript.Config
)

// Provider types
type (
	// MeetingProvider is the core interface that all meeting providers must implement.
	MeetingProvider = provider.MeetingProvider
	// RecordingProvider is an optional interface for providers that support recording.
	RecordingProvider = provider.RecordingProvider
	// WebhookHandler is an optional interface for providers that support webhooks.
	WebhookHandler = provider.WebhookHandler
	// AgentParticipant represents an AI agent's participation in a meeting.
	AgentParticipant = provider.AgentParticipant
	// AgentParticipantFactory creates AgentParticipant instances.
	AgentParticipantFactory = provider.AgentParticipantFactory
	// AgentParticipantOptions contains options for creating an AgentParticipant.
	AgentParticipantOptions = provider.AgentParticipantOptions
	// AudioFrame represents a frame of audio data.
	AudioFrame = provider.AudioFrame
	// AudioConfig specifies the audio configuration for an agent.
	AudioConfig = provider.AudioConfig
	// DataMessage represents a data message sent between participants.
	DataMessage = provider.DataMessage
	// ConnectionState represents the connection state.
	ConnectionState = provider.ConnectionState
	// ProviderClient manages multiple meeting providers.
	ProviderClient = provider.Client[provider.MeetingProvider]
)

// Connection state constants
const (
	ConnectionStateDisconnected = provider.ConnectionStateDisconnected
	ConnectionStateConnecting   = provider.ConnectionStateConnecting
	ConnectionStateConnected    = provider.ConnectionStateConnected
	ConnectionStateReconnecting = provider.ConnectionStateReconnecting
	ConnectionStateFailed       = provider.ConnectionStateFailed
)

// Helper functions

// DefaultAudioConfig returns the default audio configuration.
var DefaultAudioConfig = provider.DefaultAudioConfig

// DefaultHumanPermissions returns the default permissions for a human participant.
var DefaultHumanPermissions = participant.DefaultHumanPermissions

// DefaultAgentPermissions returns the default permissions for an AI agent.
var DefaultAgentPermissions = participant.DefaultAgentPermissions

// DefaultObserverPermissions returns the default permissions for an observer.
var DefaultObserverPermissions = participant.DefaultObserverPermissions

// DefaultRecorderPermissions returns the default permissions for a recorder.
var DefaultRecorderPermissions = participant.DefaultRecorderPermissions

// NewProviderClient creates a new provider client.
func NewProviderClient(primary MeetingProvider, fallbacks ...MeetingProvider) *ProviderClient {
	return provider.NewClient(primary, fallbacks...)
}

// Voice integration types
type (
	// VoiceAgentParticipant wraps AgentParticipant with OmniVoice integration.
	VoiceAgentParticipant = voice.VoiceAgentParticipant
	// VoiceConfig configures the VoiceAgentParticipant.
	VoiceConfig = voice.Config
	// TranscriptHandler is called when speech is transcribed.
	TranscriptHandler = voice.TranscriptHandler
	// STTProvider is the interface for speech-to-text providers (from omnivoice-core).
	STTProvider = stt.Provider
	// TTSProvider is the interface for text-to-speech providers (from omnivoice-core).
	TTSProvider = tts.Provider
	// STTConfig configures speech-to-text (from omnivoice-core).
	STTConfig = stt.TranscriptionConfig
	// TTSConfig configures text-to-speech (from omnivoice-core).
	TTSConfig = tts.SynthesisConfig
)

// NewVoiceAgentParticipant creates a new VoiceAgentParticipant.
var NewVoiceAgentParticipant = voice.NewVoiceAgentParticipant

// ToSTTFormat converts a meeting AudioFrame to OmniVoice STT format.
var ToSTTFormat = voice.ToSTTFormat

// FromTTSResult converts an OmniVoice TTS result to a meeting AudioFrame.
var FromTTSResult = voice.FromTTSResult
