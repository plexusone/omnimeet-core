# Voice Types

Package `voice` provides OmniVoice integration for STT/TTS in meetings.

!!! info "OmniVoice Integration"
    OmniMeet uses [omnivoice-core](https://github.com/plexusone/omnivoice-core) types directly.
    STT and TTS providers come from the OmniVoice registry, enabling seamless switching between
    Deepgram, OpenAI, ElevenLabs, Google, and more.

## Provider Interfaces

OmniMeet uses omnivoice-core's provider interfaces directly:

### stt.Provider

From `github.com/plexusone/omnivoice-core/stt`:

```go
type Provider interface {
    // Name returns the provider name.
    Name() string

    // Transcribe converts audio to text (batch mode).
    Transcribe(ctx context.Context, audio []byte, config TranscriptionConfig) (*TranscriptionResult, error)

    // TranscribeFile transcribes audio from a file path.
    TranscribeFile(ctx context.Context, filePath string, config TranscriptionConfig) (*TranscriptionResult, error)

    // TranscribeURL transcribes audio from a URL.
    TranscribeURL(ctx context.Context, url string, config TranscriptionConfig) (*TranscriptionResult, error)
}
```

### tts.Provider

From `github.com/plexusone/omnivoice-core/tts`:

```go
type Provider interface {
    // Name returns the provider name.
    Name() string

    // Synthesize converts text to speech and returns audio data.
    Synthesize(ctx context.Context, text string, config SynthesisConfig) (*SynthesisResult, error)

    // SynthesizeStream converts text to speech with streaming output.
    SynthesizeStream(ctx context.Context, text string, config SynthesisConfig) (<-chan StreamChunk, error)

    // ListVoices returns available voices from this provider.
    ListVoices(ctx context.Context) ([]Voice, error)

    // GetVoice returns a specific voice by ID.
    GetVoice(ctx context.Context, voiceID string) (*Voice, error)
}
```

## Configuration Types

### stt.TranscriptionConfig

From `github.com/plexusone/omnivoice-core/stt`:

```go
type TranscriptionConfig struct {
    // Language is the ISO 639-1 language code (e.g., "en", "es").
    Language string

    // Model is the provider-specific model identifier.
    Model string

    // SampleRate is the audio sample rate in Hz.
    SampleRate int

    // Channels is the number of audio channels (1=mono, 2=stereo).
    Channels int

    // Punctuation enables automatic punctuation.
    Punctuation bool

    // WordTimestamps enables word-level timing information.
    WordTimestamps bool

    // Encoding specifies the audio encoding (e.g., "pcm16", "opus").
    Encoding AudioEncoding
}
```

### tts.SynthesisConfig

From `github.com/plexusone/omnivoice-core/tts`:

```go
type SynthesisConfig struct {
    // VoiceID is the voice to use for synthesis.
    VoiceID string

    // Model is the provider-specific model identifier (optional).
    Model string

    // OutputFormat specifies the audio format ("mp3", "pcm", "wav", "opus").
    OutputFormat string

    // SampleRate is the audio sample rate in Hz (e.g., 22050, 44100).
    SampleRate int

    // Speed is the speech speed multiplier (1.0 = normal).
    Speed float64
}
```

## Result Types

### stt.TranscriptionResult

```go
type TranscriptionResult struct {
    // Text is the full transcription text.
    Text string

    // Segments contains segment-level details.
    Segments []Segment

    // Language is the detected language.
    Language string

    // LanguageConfidence is the confidence in language detection.
    LanguageConfidence float64

    // Duration is the audio duration.
    Duration time.Duration
}
```

### tts.SynthesisResult

```go
type SynthesisResult struct {
    // Audio is the synthesized audio data.
    Audio []byte

    // SampleRate is the audio sample rate.
    SampleRate int

    // Duration is the audio duration.
    Duration time.Duration
}
```

## VoiceAgentParticipant

Wraps an AgentParticipant with voice capabilities.

```go
type VoiceAgentParticipant struct {
    // ... internal fields
}

func NewVoiceAgentParticipant(agent provider.AgentParticipant, config Config) *VoiceAgentParticipant
```

### Config

```go
type Config struct {
    // STTProvider is the speech-to-text provider (from omnivoice-core).
    STTProvider stt.Provider

    // TTSProvider is the text-to-speech provider (from omnivoice-core).
    TTSProvider tts.Provider

    // STTConfig configures speech-to-text.
    STTConfig stt.TranscriptionConfig

    // TTSConfig configures text-to-speech.
    TTSConfig tts.SynthesisConfig

    // OnTranscript is called when speech is transcribed.
    OnTranscript TranscriptHandler
}

// TranscriptHandler is called when speech is transcribed.
type TranscriptHandler func(participantID, participantName, text string, isFinal bool)
```

### Methods

```go
// Participant returns the underlying AgentParticipant.
func (v *VoiceAgentParticipant) Participant() provider.AgentParticipant

// StartTranscribing begins transcription for a participant.
func (v *VoiceAgentParticipant) StartTranscribing(ctx context.Context, participantID string) error

// StopTranscribing stops transcription for a participant.
func (v *VoiceAgentParticipant) StopTranscribing(participantID string)

// StartTranscribingAll starts transcription for all human participants.
func (v *VoiceAgentParticipant) StartTranscribingAll(ctx context.Context) error

// StopTranscribingAll stops all active transcriptions.
func (v *VoiceAgentParticipant) StopTranscribingAll()

// Speak synthesizes text and publishes audio to the meeting.
func (v *VoiceAgentParticipant) Speak(ctx context.Context, text string) error

// SpeakStream synthesizes text with streaming and publishes audio chunks.
func (v *VoiceAgentParticipant) SpeakStream(ctx context.Context, text string) error

// OnTranscript sets the transcript handler.
func (v *VoiceAgentParticipant) OnTranscript(handler TranscriptHandler)
```

## Convenience Type Aliases

OmniMeet re-exports omnivoice-core types for convenience:

```go
import omnimeet "github.com/plexusone/omnimeet-core"

// These are aliases to omnivoice-core types
type STTProvider = stt.Provider
type TTSProvider = tts.Provider
type STTConfig = stt.TranscriptionConfig
type TTSConfig = tts.SynthesisConfig
```

## Utility Functions

### ToSTTFormat

Converts a meeting AudioFrame to OmniVoice STT format.

```go
func ToSTTFormat(frame provider.AudioFrame, targetSampleRate int) ([]byte, stt.TranscriptionConfig)
```

### FromTTSResult

Converts an OmniVoice TTS result to a meeting AudioFrame.

```go
func FromTTSResult(result *tts.SynthesisResult, targetSampleRate int) provider.AudioFrame
```

## Example Usage

```go
import (
    "github.com/plexusone/omnimeet-core/voice"
    "github.com/plexusone/omnivoice"
    "github.com/plexusone/omnivoice-core/stt"
    "github.com/plexusone/omnivoice-core/tts"
    _ "github.com/plexusone/omnivoice/providers/all"
)

// Get providers from OmniVoice registry
sttProvider, _ := omnivoice.GetSTTProvider("deepgram",
    omnivoice.WithAPIKey(os.Getenv("DEEPGRAM_API_KEY")),
)
ttsProvider, _ := omnivoice.GetTTSProvider("elevenlabs",
    omnivoice.WithAPIKey(os.Getenv("ELEVENLABS_API_KEY")),
)

// Wrap agent with voice
voiceAgent := voice.NewVoiceAgentParticipant(agent, voice.Config{
    STTProvider: sttProvider,
    TTSProvider: ttsProvider,
    STTConfig: stt.TranscriptionConfig{
        Language:   "en",
        SampleRate: 16000,
        Channels:   1,
    },
    TTSConfig: tts.SynthesisConfig{
        VoiceID:    "rachel",
        SampleRate: 48000,
        Speed:      1.0,
    },
    OnTranscript: func(participantID, name, text string, isFinal bool) {
        if isFinal {
            log.Printf("%s: %s", name, text)
        }
    },
})

// Join and transcribe
voiceAgent.Participant().JoinMeeting(ctx, meetingID, tok)
voiceAgent.StartTranscribingAll(ctx)

// Speak
voiceAgent.Speak(ctx, "Hello, how can I help you today?")
```

## Streaming STT

For real-time transcription, providers can implement `stt.StreamingProvider`:

```go
type StreamingProvider interface {
    Provider

    // TranscribeStream starts a streaming transcription session.
    TranscribeStream(ctx context.Context, config TranscriptionConfig) (io.WriteCloser, <-chan StreamEvent, error)
}

type StreamEvent struct {
    Type       StreamEventType
    Transcript string
    IsFinal    bool
    Segment    *Segment
    Error      error
}
```

VoiceAgentParticipant automatically uses streaming when available.

## See Also

- [Voice Integration Guide](../guides/voice-integration.md) - Detailed integration guide
- [OmniVoice Documentation](https://github.com/plexusone/omnivoice) - Full provider documentation
- [omnivoice-core](https://github.com/plexusone/omnivoice-core) - Core types and interfaces
