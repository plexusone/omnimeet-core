// Package meeting provides the core Meeting type and related types for OmniMeet.
package meeting

import (
	"time"
)

// Status represents the current state of a meeting.
type Status string

const (
	// StatusPending indicates the meeting has been created but not yet started.
	StatusPending Status = "pending"
	// StatusActive indicates the meeting is currently in progress.
	StatusActive Status = "active"
	// StatusEnded indicates the meeting has concluded.
	StatusEnded Status = "ended"
)

// Meeting represents a live collaborative session.
//
// A Meeting is the primary abstraction in OmniMeet, representing a real-time
// collaboration space that can contain humans, AI agents, recorders, and other
// participants. This maps to provider-specific concepts like LiveKit Rooms,
// Daily Rooms, or Zoom Meetings.
type Meeting struct {
	// ID is the unique identifier for this meeting.
	ID string `json:"id"`

	// Name is the human-readable name of the meeting.
	Name string `json:"name"`

	// Status is the current state of the meeting.
	Status Status `json:"status"`

	// Provider is the name of the provider hosting this meeting (e.g., "livekit", "daily").
	Provider string `json:"provider"`

	// ParticipantCount is the current number of participants in the meeting.
	ParticipantCount int `json:"participant_count"`

	// MaxParticipants is the maximum allowed participants (0 = unlimited).
	MaxParticipants int `json:"max_participants,omitempty"`

	// CreatedAt is when the meeting was created.
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when the first participant joined (nil if not started).
	StartedAt *time.Time `json:"started_at,omitempty"`

	// EndedAt is when the meeting ended (nil if still active).
	EndedAt *time.Time `json:"ended_at,omitempty"`

	// Metadata contains arbitrary key-value pairs for application-specific data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// RecordingEnabled indicates whether recording is enabled for this meeting.
	RecordingEnabled bool `json:"recording_enabled,omitempty"`

	// TranscriptionEnabled indicates whether transcription is enabled for this meeting.
	TranscriptionEnabled bool `json:"transcription_enabled,omitempty"`

	// JoinURL is the URL for participants to join the meeting (if available).
	JoinURL string `json:"join_url,omitempty"`
}

// IsActive returns true if the meeting is currently active.
func (m *Meeting) IsActive() bool {
	return m.Status == StatusActive
}

// Duration returns the duration of the meeting.
// Returns 0 if the meeting hasn't started yet.
func (m *Meeting) Duration() time.Duration {
	if m.StartedAt == nil {
		return 0
	}
	if m.EndedAt != nil {
		return m.EndedAt.Sub(*m.StartedAt)
	}
	return time.Since(*m.StartedAt)
}

// CreateRequest contains the parameters for creating a new meeting.
type CreateRequest struct {
	// Name is the human-readable name of the meeting.
	Name string `json:"name"`

	// MaxParticipants limits the number of participants (0 = unlimited).
	MaxParticipants int `json:"max_participants,omitempty"`

	// RecordingEnabled enables recording for this meeting.
	RecordingEnabled bool `json:"recording_enabled,omitempty"`

	// TranscriptionEnabled enables transcription for this meeting.
	TranscriptionEnabled bool `json:"transcription_enabled,omitempty"`

	// Metadata contains arbitrary key-value pairs for application-specific data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// EmptyTimeout is how long to wait before closing an empty meeting (0 = default).
	EmptyTimeout time.Duration `json:"empty_timeout,omitempty"`

	// MaxDuration is the maximum duration of the meeting (0 = unlimited).
	MaxDuration time.Duration `json:"max_duration,omitempty"`

	// Extensions contains provider-specific configuration.
	// Keys should be namespaced (e.g., "livekit.egress_config").
	Extensions map[string]any `json:"extensions,omitempty"`
}

// ListOptions contains options for listing meetings.
type ListOptions struct {
	// Status filters meetings by status.
	Status *Status `json:"status,omitempty"`

	// Limit is the maximum number of meetings to return.
	Limit int `json:"limit,omitempty"`

	// Offset is the number of meetings to skip.
	Offset int `json:"offset,omitempty"`

	// IncludeEnded includes ended meetings in the results.
	IncludeEnded bool `json:"include_ended,omitempty"`
}
