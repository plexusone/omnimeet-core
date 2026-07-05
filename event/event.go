// Package event provides event types for OmniMeet.
package event

import (
	"time"

	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/track"
)

// Type represents the type of event.
type Type string

const (
	// Meeting lifecycle events
	TypeMeetingCreated Type = "meeting.created"
	TypeMeetingStarted Type = "meeting.started"
	TypeMeetingEnded   Type = "meeting.ended"

	// Participant lifecycle events
	TypeParticipantJoined  Type = "participant.joined"
	TypeParticipantLeft    Type = "participant.left"
	TypeParticipantUpdated Type = "participant.updated"

	// Track events
	TypeTrackPublished    Type = "track.published"
	TypeTrackUnpublished  Type = "track.unpublished"
	TypeTrackMuted        Type = "track.muted"
	TypeTrackUnmuted      Type = "track.unmuted"
	TypeTrackSubscribed   Type = "track.subscribed"
	TypeTrackUnsubscribed Type = "track.unsubscribed"

	// Speaking events
	TypeActiveSpeakerChanged Type = "active_speaker.changed"
	TypeSpeakingStarted      Type = "speaking.started"
	TypeSpeakingStopped      Type = "speaking.stopped"

	// Recording events
	TypeRecordingStarted Type = "recording.started"
	TypeRecordingStopped Type = "recording.stopped"
	TypeRecordingFailed  Type = "recording.failed"

	// Transcript events
	TypeTranscriptUpdated   Type = "transcript.updated"
	TypeTranscriptFinalized Type = "transcript.finalized"

	// Data events
	TypeDataMessageReceived Type = "data_message.received"

	// Connection events
	TypeConnectionQualityChanged Type = "connection.quality_changed"
	TypeReconnecting             Type = "connection.reconnecting"
	TypeReconnected              Type = "connection.reconnected"
	TypeDisconnected             Type = "connection.disconnected"
)

// Event represents an event that occurred in a meeting.
type Event struct {
	// ID is a unique identifier for this event.
	ID string `json:"id"`

	// Type is the type of event.
	Type Type `json:"type"`

	// MeetingID is the ID of the meeting where the event occurred.
	MeetingID string `json:"meeting_id"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Data contains event-specific data.
	Data any `json:"data,omitempty"`
}

// Handler is a function that handles events.
type Handler func(event Event) error

// ParticipantData contains data for participant events.
type ParticipantData struct {
	Participant participant.Participant `json:"participant"`
}

// TrackData contains data for track events.
type TrackData struct {
	Participant participant.Participant `json:"participant"`
	Track       track.Track             `json:"track"`
}

// ActiveSpeakerData contains data for active speaker events.
type ActiveSpeakerData struct {
	// Speakers is the list of currently speaking participants, ordered by volume.
	Speakers []participant.Participant `json:"speakers"`
}

// SpeakingData contains data for speaking start/stop events.
type SpeakingData struct {
	Participant participant.Participant `json:"participant"`
}

// RecordingData contains data for recording events.
type RecordingData struct {
	RecordingID string `json:"recording_id"`
	MeetingID   string `json:"meeting_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

// TranscriptData contains data for transcript events.
type TranscriptData struct {
	ParticipantID   string    `json:"participant_id"`
	ParticipantName string    `json:"participant_name"`
	Text            string    `json:"text"`
	IsFinal         bool      `json:"is_final"`
	Confidence      float64   `json:"confidence,omitempty"`
	Language        string    `json:"language,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time,omitempty"`
}

// DataMessageData contains data for data message events.
type DataMessageData struct {
	From      participant.Participant `json:"from"`
	Topic     string                  `json:"topic,omitempty"`
	Payload   []byte                  `json:"payload"`
	Reliable  bool                    `json:"reliable"`
	Timestamp time.Time               `json:"timestamp"`
}

// ConnectionQualityData contains data for connection quality events.
type ConnectionQualityData struct {
	Participant participant.Participant           `json:"participant"`
	Quality     participant.ConnectionQuality     `json:"quality"`
	Previous    participant.ConnectionQuality     `json:"previous,omitempty"`
}

// MeetingEndedData contains data for meeting ended events.
type MeetingEndedData struct {
	Reason          string        `json:"reason"`
	Duration        time.Duration `json:"duration"`
	ParticipantCount int          `json:"participant_count"`
}
