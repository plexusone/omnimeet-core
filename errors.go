// Package omnimeet provides a unified abstraction for real-time collaboration platforms.
package omnimeet

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrMeetingNotFound is returned when a meeting cannot be found.
	ErrMeetingNotFound = errors.New("meeting not found")

	// ErrMeetingEnded is returned when trying to operate on an ended meeting.
	ErrMeetingEnded = errors.New("meeting has ended")

	// ErrMeetingFull is returned when a meeting has reached its participant limit.
	ErrMeetingFull = errors.New("meeting is full")

	// ErrParticipantNotFound is returned when a participant cannot be found.
	ErrParticipantNotFound = errors.New("participant not found")

	// ErrParticipantAlreadyJoined is returned when a participant tries to join twice.
	ErrParticipantAlreadyJoined = errors.New("participant already joined")

	// ErrTrackNotFound is returned when a track cannot be found.
	ErrTrackNotFound = errors.New("track not found")

	// ErrTrackAlreadyPublished is returned when trying to publish a duplicate track.
	ErrTrackAlreadyPublished = errors.New("track already published")

	// ErrNotConnected is returned when an operation requires a connection.
	ErrNotConnected = errors.New("not connected to meeting")

	// ErrAlreadyConnected is returned when trying to connect while already connected.
	ErrAlreadyConnected = errors.New("already connected to meeting")

	// ErrConnectionFailed is returned when a connection attempt fails.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrTokenExpired is returned when a join token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenInvalid is returned when a join token is invalid.
	ErrTokenInvalid = errors.New("token invalid")

	// ErrPermissionDenied is returned when an operation is not permitted.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRecordingNotFound is returned when a recording cannot be found.
	ErrRecordingNotFound = errors.New("recording not found")

	// ErrRecordingAlreadyStarted is returned when trying to start a second recording.
	ErrRecordingAlreadyStarted = errors.New("recording already started")

	// ErrRecordingNotStarted is returned when trying to stop a non-existent recording.
	ErrRecordingNotStarted = errors.New("recording not started")

	// ErrRecordingNotSupported is returned when recording is not supported.
	ErrRecordingNotSupported = errors.New("recording not supported by provider")

	// ErrAgentParticipationNotSupported is returned when agent participation is not supported.
	ErrAgentParticipationNotSupported = errors.New("agent participation not supported by provider")

	// ErrProviderNotFound is returned when a provider cannot be found in the registry.
	ErrProviderNotFound = errors.New("provider not found")

	// ErrProviderAlreadyRegistered is returned when trying to register a duplicate provider.
	ErrProviderAlreadyRegistered = errors.New("provider already registered")

	// ErrInvalidConfiguration is returned when configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrWebhookValidationFailed is returned when webhook signature validation fails.
	ErrWebhookValidationFailed = errors.New("webhook validation failed")

	// ErrRateLimited is returned when rate limited by the provider.
	ErrRateLimited = errors.New("rate limited")

	// ErrQuotaExceeded is returned when quota is exceeded.
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")
)

// ProviderError wraps an error with provider context.
type ProviderError struct {
	Provider string
	Op       string
	Err      error
}

// Error returns the error message.
func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Provider, e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, op string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Op:       op,
		Err:      err,
	}
}

// MeetingError wraps an error with meeting context.
type MeetingError struct {
	MeetingID string
	Op        string
	Err       error
}

// Error returns the error message.
func (e *MeetingError) Error() string {
	return fmt.Sprintf("meeting %s: %s: %v", e.MeetingID, e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *MeetingError) Unwrap() error {
	return e.Err
}

// NewMeetingError creates a new MeetingError.
func NewMeetingError(meetingID, op string, err error) *MeetingError {
	return &MeetingError{
		MeetingID: meetingID,
		Op:        op,
		Err:       err,
	}
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable errors
	if errors.Is(err, ErrRateLimited) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrConnectionFailed) {
		return true
	}

	return false
}

// IsPermanent returns true if the error is permanent (not retryable).
func IsPermanent(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific permanent errors
	if errors.Is(err, ErrMeetingNotFound) ||
		errors.Is(err, ErrMeetingEnded) ||
		errors.Is(err, ErrParticipantNotFound) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrTokenInvalid) ||
		errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrProviderNotFound) ||
		errors.Is(err, ErrInvalidConfiguration) ||
		errors.Is(err, ErrRecordingNotSupported) ||
		errors.Is(err, ErrAgentParticipationNotSupported) {
		return true
	}

	return false
}
