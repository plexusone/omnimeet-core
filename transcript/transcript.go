// Package transcript provides transcript types for OmniMeet.
package transcript

import (
	"time"
)

// Transcript represents a meeting transcript.
type Transcript struct {
	// ID is the unique identifier for this transcript.
	ID string `json:"id"`

	// MeetingID is the ID of the meeting.
	MeetingID string `json:"meeting_id"`

	// Segments contains the transcript segments in order.
	Segments []Segment `json:"segments"`

	// Language is the detected or configured language.
	Language string `json:"language,omitempty"`

	// Duration is the total duration of the transcript.
	Duration time.Duration `json:"duration,omitempty"`

	// CreatedAt is when the transcript was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the transcript was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// Segment represents a segment of a transcript.
type Segment struct {
	// ID is the unique identifier for this segment.
	ID string `json:"id"`

	// ParticipantID is the ID of the speaker.
	ParticipantID string `json:"participant_id"`

	// ParticipantName is the name of the speaker.
	ParticipantName string `json:"participant_name"`

	// Text is the transcribed text.
	Text string `json:"text"`

	// StartTime is when this segment started (relative to meeting start).
	StartTime time.Duration `json:"start_time"`

	// EndTime is when this segment ended (relative to meeting start).
	EndTime time.Duration `json:"end_time"`

	// Confidence is the confidence score (0-1).
	Confidence float64 `json:"confidence,omitempty"`

	// Language is the detected language for this segment.
	Language string `json:"language,omitempty"`

	// Words contains word-level timing (if available).
	Words []Word `json:"words,omitempty"`

	// IsFinal indicates whether this is a final transcript.
	IsFinal bool `json:"is_final"`
}

// Word represents a word with timing information.
type Word struct {
	// Text is the word.
	Text string `json:"text"`

	// StartTime is when the word started.
	StartTime time.Duration `json:"start_time"`

	// EndTime is when the word ended.
	EndTime time.Duration `json:"end_time"`

	// Confidence is the confidence score (0-1).
	Confidence float64 `json:"confidence,omitempty"`
}

// Text returns the full text of the transcript.
func (t *Transcript) Text() string {
	var result string
	for i, seg := range t.Segments {
		if i > 0 {
			result += " "
		}
		result += seg.Text
	}
	return result
}

// ByParticipant returns segments grouped by participant.
func (t *Transcript) ByParticipant() map[string][]Segment {
	result := make(map[string][]Segment)
	for _, seg := range t.Segments {
		result[seg.ParticipantID] = append(result[seg.ParticipantID], seg)
	}
	return result
}

// Config contains configuration for transcription.
type Config struct {
	// Enabled enables transcription.
	Enabled bool `json:"enabled"`

	// Language specifies the expected language (BCP-47).
	Language string `json:"language,omitempty"`

	// Provider specifies the STT provider (uses OmniVoice).
	Provider string `json:"provider,omitempty"`

	// Model specifies the STT model.
	Model string `json:"model,omitempty"`

	// Punctuation enables automatic punctuation.
	Punctuation bool `json:"punctuation,omitempty"`

	// WordTimestamps enables word-level timestamps.
	WordTimestamps bool `json:"word_timestamps,omitempty"`

	// Diarization enables speaker diarization.
	Diarization bool `json:"diarization,omitempty"`
}
