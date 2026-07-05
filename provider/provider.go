// Package provider defines the core provider interfaces for OmniMeet.
package provider

import (
	"context"

	"github.com/plexusone/omnimeet-core/event"
	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/recording"
	"github.com/plexusone/omnimeet-core/token"
)

// MeetingProvider is the core interface that all meeting providers must implement.
//
// MeetingProvider handles the control plane operations: creating meetings,
// generating join tokens, managing participants, and handling recordings.
// For AI agent participation (media plane), see AgentParticipant.
type MeetingProvider interface {
	// Name returns the provider name (e.g., "livekit", "daily").
	Name() string

	// Meeting lifecycle
	// CreateMeeting creates a new meeting.
	CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error)

	// GetMeeting retrieves a meeting by ID.
	GetMeeting(ctx context.Context, meetingID string) (*meeting.Meeting, error)

	// ListMeetings returns a list of meetings.
	ListMeetings(ctx context.Context, opts meeting.ListOptions) ([]meeting.Meeting, error)

	// EndMeeting ends a meeting.
	EndMeeting(ctx context.Context, meetingID string) error

	// DeleteMeeting deletes a meeting (may fail if active).
	DeleteMeeting(ctx context.Context, meetingID string) error

	// Participant management
	// CreateJoinToken generates a token for a participant to join a meeting.
	CreateJoinToken(ctx context.Context, req token.CreateRequest) (*token.JoinToken, error)

	// ListParticipants returns the current participants in a meeting.
	ListParticipants(ctx context.Context, meetingID string) ([]participant.Participant, error)

	// GetParticipant retrieves a specific participant.
	GetParticipant(ctx context.Context, meetingID, participantID string) (*participant.Participant, error)

	// RemoveParticipant removes a participant from a meeting.
	RemoveParticipant(ctx context.Context, meetingID, participantID string) error

	// UpdateParticipant updates participant metadata or permissions.
	UpdateParticipant(ctx context.Context, meetingID, participantID string, update ParticipantUpdate) error

	// Events
	// OnEvent registers a handler for meeting events.
	// Events are delivered via webhooks or polling, depending on the provider.
	OnEvent(handler event.Handler)

	// Lifecycle
	// Close releases any resources held by the provider.
	Close() error
}

// ParticipantUpdate contains fields to update on a participant.
type ParticipantUpdate struct {
	// Name updates the participant's display name.
	Name *string `json:"name,omitempty"`

	// Metadata updates the participant's metadata (merged with existing).
	Metadata map[string]string `json:"metadata,omitempty"`

	// Permissions updates the participant's permissions.
	Permissions *participant.Permissions `json:"permissions,omitempty"`
}

// RecordingProvider is an optional interface for providers that support recording.
type RecordingProvider interface {
	MeetingProvider

	// StartRecording starts recording a meeting.
	StartRecording(ctx context.Context, meetingID string, opts recording.Options) (*recording.Recording, error)

	// StopRecording stops recording a meeting.
	StopRecording(ctx context.Context, meetingID string) (*recording.Recording, error)

	// GetRecording retrieves a recording by ID.
	GetRecording(ctx context.Context, recordingID string) (*recording.Recording, error)

	// ListRecordings returns recordings for a meeting.
	ListRecordings(ctx context.Context, meetingID string) ([]recording.Recording, error)
}

// WebhookHandler is an optional interface for providers that support webhooks.
type WebhookHandler interface {
	// HandleWebhook processes an incoming webhook request.
	// Returns the parsed events.
	HandleWebhook(ctx context.Context, body []byte, headers map[string]string) ([]event.Event, error)

	// ValidateWebhook validates the webhook signature.
	ValidateWebhook(body []byte, headers map[string]string, secret string) error
}

// Named is a common interface for types that have a name.
type Named interface {
	Name() string
}

// ensure MeetingProvider implements Named
var _ Named = (MeetingProvider)(nil)
