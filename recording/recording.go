// Package recording provides recording types for OmniMeet.
package recording

import (
	"time"
)

// Status represents the status of a recording.
type Status string

const (
	StatusPending    Status = "pending"
	StatusRecording  Status = "recording"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Recording represents a meeting recording.
type Recording struct {
	// ID is the unique identifier for this recording.
	ID string `json:"id"`

	// MeetingID is the ID of the meeting being recorded.
	MeetingID string `json:"meeting_id"`

	// Status is the current status of the recording.
	Status Status `json:"status"`

	// StartedAt is when the recording started.
	StartedAt time.Time `json:"started_at"`

	// EndedAt is when the recording ended (nil if still recording).
	EndedAt *time.Time `json:"ended_at,omitempty"`

	// Duration is the duration of the recording.
	Duration time.Duration `json:"duration,omitempty"`

	// Size is the size of the recording in bytes.
	Size int64 `json:"size,omitempty"`

	// URL is the download URL for the recording (if available).
	URL string `json:"url,omitempty"`

	// Format is the recording format (e.g., "mp4", "webm").
	Format string `json:"format,omitempty"`

	// Error contains error details if the recording failed.
	Error string `json:"error,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsActive returns true if the recording is still in progress.
func (r *Recording) IsActive() bool {
	return r.Status == StatusPending || r.Status == StatusRecording
}

// Options contains options for starting a recording.
type Options struct {
	// Format specifies the output format (e.g., "mp4", "webm").
	Format string `json:"format,omitempty"`

	// AudioOnly records only audio.
	AudioOnly bool `json:"audio_only,omitempty"`

	// Layout specifies the video layout for composite recordings.
	Layout Layout `json:"layout,omitempty"`

	// OutputURL specifies where to upload the recording.
	OutputURL string `json:"output_url,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Extensions contains provider-specific configuration.
	Extensions map[string]any `json:"extensions,omitempty"`
}

// Layout specifies the video layout for composite recordings.
type Layout string

const (
	LayoutGrid       Layout = "grid"
	LayoutSpeaker    Layout = "speaker"
	LayoutSideBySide Layout = "side_by_side"
)
