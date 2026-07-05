// Package track provides the Track type and related types for OmniMeet.
package track

import (
	"time"
)

// Kind represents the type of media track.
type Kind string

const (
	// KindAudio represents an audio track (microphone, etc.).
	KindAudio Kind = "audio"
	// KindVideo represents a video track (camera, etc.).
	KindVideo Kind = "video"
	// KindScreenShare represents a screen share track.
	KindScreenShare Kind = "screen_share"
	// KindScreenShareAudio represents audio from a screen share.
	KindScreenShareAudio Kind = "screen_share_audio"
	// KindData represents a data channel.
	KindData Kind = "data"
)

// Source represents the source of a track.
type Source string

const (
	// SourceMicrophone indicates audio from a microphone.
	SourceMicrophone Source = "microphone"
	// SourceCamera indicates video from a camera.
	SourceCamera Source = "camera"
	// SourceScreen indicates content from screen sharing.
	SourceScreen Source = "screen"
	// SourceApplication indicates content from a specific application.
	SourceApplication Source = "application"
	// SourceUnknown indicates an unknown source.
	SourceUnknown Source = "unknown"
)

// Track represents a media stream published by a participant.
//
// Tracks can be audio, video, screen share, or data channels. Each track
// has a kind, source, and mute state.
type Track struct {
	// ID is the unique identifier for this track.
	ID string `json:"id"`

	// ParticipantID is the ID of the participant who published this track.
	ParticipantID string `json:"participant_id"`

	// Kind indicates the type of track.
	Kind Kind `json:"kind"`

	// Source indicates where the track originated.
	Source Source `json:"source"`

	// Name is an optional name for the track.
	Name string `json:"name,omitempty"`

	// Muted indicates whether the track is currently muted.
	Muted bool `json:"muted"`

	// PublishedAt is when the track was published.
	PublishedAt time.Time `json:"published_at"`

	// UnpublishedAt is when the track was unpublished (nil if still active).
	UnpublishedAt *time.Time `json:"unpublished_at,omitempty"`

	// Width is the video width in pixels (0 for non-video tracks).
	Width int `json:"width,omitempty"`

	// Height is the video height in pixels (0 for non-video tracks).
	Height int `json:"height,omitempty"`

	// Simulcast indicates whether simulcast is enabled for this track.
	Simulcast bool `json:"simulcast,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsActive returns true if the track is still published.
func (t *Track) IsActive() bool {
	return t.UnpublishedAt == nil
}

// IsAudio returns true if this is an audio track.
func (t *Track) IsAudio() bool {
	return t.Kind == KindAudio || t.Kind == KindScreenShareAudio
}

// IsVideo returns true if this is a video track.
func (t *Track) IsVideo() bool {
	return t.Kind == KindVideo || t.Kind == KindScreenShare
}

// PublishRequest contains options for publishing a track.
type PublishRequest struct {
	// Kind is the type of track to publish.
	Kind Kind `json:"kind"`

	// Source is where the track originates.
	Source Source `json:"source"`

	// Name is an optional name for the track.
	Name string `json:"name,omitempty"`

	// Simulcast enables simulcast for video tracks.
	Simulcast bool `json:"simulcast,omitempty"`

	// Width is the video width (for video tracks).
	Width int `json:"width,omitempty"`

	// Height is the video height (for video tracks).
	Height int `json:"height,omitempty"`

	// Metadata contains arbitrary key-value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SubscribeOptions contains options for subscribing to a track.
type SubscribeOptions struct {
	// Quality specifies the desired quality for video tracks.
	Quality VideoQuality `json:"quality,omitempty"`
}

// VideoQuality represents the quality level for video tracks.
type VideoQuality string

const (
	VideoQualityLow    VideoQuality = "low"
	VideoQualityMedium VideoQuality = "medium"
	VideoQualityHigh   VideoQuality = "high"
	VideoQualityOff    VideoQuality = "off"
)
