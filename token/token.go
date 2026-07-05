// Package token provides join token types for OmniMeet.
package token

import (
	"time"

	"github.com/plexusone/omnimeet-core/participant"
)

// JoinToken represents an access token for joining a meeting.
type JoinToken struct {
	// Token is the actual token string.
	Token string `json:"token"`

	// MeetingID is the ID of the meeting this token is for.
	MeetingID string `json:"meeting_id"`

	// ParticipantIdentity is the identity of the participant.
	ParticipantIdentity string `json:"participant_identity"`

	// ParticipantName is the display name of the participant.
	ParticipantName string `json:"participant_name"`

	// ExpiresAt is when the token expires.
	ExpiresAt time.Time `json:"expires_at"`

	// JoinURL is the full URL to join the meeting (if available).
	JoinURL string `json:"join_url,omitempty"`

	// Metadata contains additional token metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsExpired returns true if the token has expired.
func (t *JoinToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TTL returns the remaining time until the token expires.
func (t *JoinToken) TTL() time.Duration {
	return time.Until(t.ExpiresAt)
}

// CreateRequest contains the parameters for creating a join token.
type CreateRequest struct {
	// MeetingID is the ID of the meeting to join.
	MeetingID string `json:"meeting_id"`

	// Participant contains information about the participant.
	Participant participant.Info `json:"participant"`

	// TTL is how long the token should be valid (0 = default).
	TTL time.Duration `json:"ttl,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Extensions contains provider-specific configuration.
	Extensions map[string]any `json:"extensions,omitempty"`
}

// DefaultTTL is the default token validity period.
const DefaultTTL = 24 * time.Hour
