// Package participant provides the Participant type and related types for OmniMeet.
package participant

import (
	"time"

	"github.com/plexusone/omnimeet-core/track"
)

// Kind represents the type of participant.
type Kind string

const (
	// KindHuman represents a human participant.
	KindHuman Kind = "human"
	// KindAgent represents an AI agent participant.
	KindAgent Kind = "agent"
	// KindRecorder represents a recording bot.
	KindRecorder Kind = "recorder"
	// KindObserver represents an observer (can see but not be seen).
	KindObserver Kind = "observer"
	// KindSIP represents a participant connected via SIP/PSTN.
	KindSIP Kind = "sip"
)

// Participant represents an entity that joins a meeting.
//
// Participants can be humans, AI agents, recording bots, observers, or
// SIP/phone callers. Each participant can publish and subscribe to tracks.
type Participant struct {
	// ID is the unique identifier for this participant within the meeting.
	ID string `json:"id"`

	// MeetingID is the ID of the meeting this participant belongs to.
	MeetingID string `json:"meeting_id"`

	// Kind indicates the type of participant.
	Kind Kind `json:"kind"`

	// Name is the display name of the participant.
	Name string `json:"name"`

	// Identity is the provider-specific identity string.
	Identity string `json:"identity"`

	// JoinedAt is when the participant joined the meeting.
	JoinedAt time.Time `json:"joined_at"`

	// LeftAt is when the participant left the meeting (nil if still present).
	LeftAt *time.Time `json:"left_at,omitempty"`

	// Tracks are the media tracks published by this participant.
	Tracks []track.Track `json:"tracks,omitempty"`

	// IsSpeaking indicates whether the participant is currently speaking.
	IsSpeaking bool `json:"is_speaking,omitempty"`

	// ConnectionQuality indicates the participant's connection quality.
	ConnectionQuality ConnectionQuality `json:"connection_quality,omitempty"`

	// Metadata contains arbitrary key-value pairs for application-specific data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Permissions defines what this participant is allowed to do.
	Permissions *Permissions `json:"permissions,omitempty"`
}

// IsActive returns true if the participant is still in the meeting.
func (p *Participant) IsActive() bool {
	return p.LeftAt == nil
}

// Duration returns how long the participant has been in the meeting.
func (p *Participant) Duration() time.Duration {
	if p.LeftAt != nil {
		return p.LeftAt.Sub(p.JoinedAt)
	}
	return time.Since(p.JoinedAt)
}

// HasAudio returns true if the participant has published an audio track.
func (p *Participant) HasAudio() bool {
	for _, t := range p.Tracks {
		if t.Kind == track.KindAudio && !t.Muted {
			return true
		}
	}
	return false
}

// HasVideo returns true if the participant has published a video track.
func (p *Participant) HasVideo() bool {
	for _, t := range p.Tracks {
		if t.Kind == track.KindVideo && !t.Muted {
			return true
		}
	}
	return false
}

// ConnectionQuality indicates the quality of a participant's connection.
type ConnectionQuality string

const (
	ConnectionQualityUnknown   ConnectionQuality = "unknown"
	ConnectionQualityExcellent ConnectionQuality = "excellent"
	ConnectionQualityGood      ConnectionQuality = "good"
	ConnectionQualityPoor      ConnectionQuality = "poor"
	ConnectionQualityLost      ConnectionQuality = "lost"
)

// Permissions defines what a participant is allowed to do in a meeting.
type Permissions struct {
	// CanPublish allows the participant to publish tracks.
	CanPublish bool `json:"can_publish"`
	// CanSubscribe allows the participant to subscribe to tracks.
	CanSubscribe bool `json:"can_subscribe"`
	// CanPublishData allows the participant to send data messages.
	CanPublishData bool `json:"can_publish_data"`
	// CanUpdateMetadata allows the participant to update their own metadata.
	CanUpdateMetadata bool `json:"can_update_metadata"`
	// Hidden makes the participant invisible to others.
	Hidden bool `json:"hidden,omitempty"`
	// Recorder indicates this participant is a recorder.
	Recorder bool `json:"recorder,omitempty"`
}

// Info contains information for creating or identifying a participant.
type Info struct {
	// Name is the display name of the participant.
	Name string `json:"name"`

	// Kind indicates the type of participant.
	Kind Kind `json:"kind"`

	// Identity is an optional identity string (provider-specific).
	Identity string `json:"identity,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Permissions defines what the participant is allowed to do.
	Permissions *Permissions `json:"permissions,omitempty"`
}

// DefaultHumanPermissions returns the default permissions for a human participant.
func DefaultHumanPermissions() *Permissions {
	return &Permissions{
		CanPublish:        true,
		CanSubscribe:      true,
		CanPublishData:    true,
		CanUpdateMetadata: true,
	}
}

// DefaultAgentPermissions returns the default permissions for an AI agent.
func DefaultAgentPermissions() *Permissions {
	return &Permissions{
		CanPublish:        true,
		CanSubscribe:      true,
		CanPublishData:    true,
		CanUpdateMetadata: true,
	}
}

// DefaultObserverPermissions returns the default permissions for an observer.
func DefaultObserverPermissions() *Permissions {
	return &Permissions{
		CanPublish:     false,
		CanSubscribe:   true,
		CanPublishData: false,
		Hidden:         true,
	}
}

// DefaultRecorderPermissions returns the default permissions for a recorder.
func DefaultRecorderPermissions() *Permissions {
	return &Permissions{
		CanPublish:     false,
		CanSubscribe:   true,
		CanPublishData: false,
		Hidden:         true,
		Recorder:       true,
	}
}
