# Voice Integration

This guide covers integrating OmniVoice for speech-to-text (STT) and text-to-speech (TTS) with OmniMeet agents.

## Overview

OmniMeet integrates directly with [OmniVoice](https://github.com/plexusone/omnivoice) to enable:

- **STT**: Convert participant audio to text
- **TTS**: Convert agent responses to speech
- **Real-time transcription**: Continuous speech recognition
- **Voice activity detection**: Detect when participants speak
- **Provider switching**: Seamlessly swap between Deepgram, OpenAI, ElevenLabs, etc.

## Voice Providers

OmniMeet uses OmniVoice's provider registry, supporting:

| Provider | STT | TTS | Notes |
|----------|-----|-----|-------|
| Deepgram | ✅ | ✅ | Real-time streaming |
| OpenAI | ✅ | ✅ | Whisper STT, multiple TTS voices |
| ElevenLabs | ❌ | ✅ | High-quality voices |
| Google Cloud | ✅ | ✅ | Enterprise |
| Twilio | ❌ | ✅ | Telephony-optimized |

## Quick Start with OmniVoice

### Installation

```bash
go get github.com/plexusone/omnivoice
go get github.com/plexusone/omnivoice/providers/all
```

### Basic Voice Agent

```go
import (
    "github.com/plexusone/omnimeet-core/voice"
    "github.com/plexusone/omnimeet-core/provider"
    "github.com/plexusone/omnivoice"
    _ "github.com/plexusone/omnivoice/providers/all" // Register all providers
)

// Get providers from OmniVoice registry
sttProvider, err := omnivoice.GetSTTProvider("deepgram",
    omnivoice.WithAPIKey(os.Getenv("DEEPGRAM_API_KEY")),
)
if err != nil {
    log.Fatal(err)
}

ttsProvider, err := omnivoice.GetTTSProvider("elevenlabs",
    omnivoice.WithAPIKey(os.Getenv("ELEVENLABS_API_KEY")),
)
if err != nil {
    log.Fatal(err)
}

// Create base agent
baseAgent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,
})

// Wrap with voice capabilities
voiceAgent := voice.NewVoiceAgentParticipant(baseAgent, voice.Config{
    STTProvider: sttProvider,
    TTSProvider: ttsProvider,
    STTConfig: stt.TranscriptionConfig{
        Language:   "en",
        SampleRate: 16000,
        Channels:   1,
    },
    TTSConfig: tts.SynthesisConfig{
        VoiceID:    "rachel", // ElevenLabs voice
        SampleRate: 48000,
    },
    OnTranscript: func(participantID, participantName, text string, isFinal bool) {
        if isFinal {
            log.Printf("%s: %s", participantName, text)
        }
    },
})
```

### Joining and Transcribing

```go
// Join meeting
voiceAgent.Participant().JoinMeeting(ctx, meetingID, token)

// Start transcribing a specific participant
voiceAgent.StartTranscribing(ctx, participantID)

// Or auto-transcribe all human participants
voiceAgent.StartTranscribingAll(ctx)

// Handle new participants
baseAgent.OnParticipantJoined(func(p participant.Participant) {
    if p.Kind == participant.KindHuman {
        voiceAgent.StartTranscribing(ctx, p.ID)
    }
})
```

### Speaking

```go
// Speak using TTS
err := voiceAgent.Speak(ctx, "Hello, I'm your AI assistant!")

// Stream speech for lower latency
err := voiceAgent.SpeakStream(ctx, "This streams audio as it's generated.")
```

## Switching Providers

One of the key benefits of the OmniVoice integration is seamless provider switching:

```go
// Switch to OpenAI for STT
sttProvider, _ := omnivoice.GetSTTProvider("openai",
    omnivoice.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
)

// Switch to Google for TTS
ttsProvider, _ := omnivoice.GetTTSProvider("google",
    omnivoice.WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
)

// Use the same VoiceAgentParticipant API
voiceAgent := voice.NewVoiceAgentParticipant(baseAgent, voice.Config{
    STTProvider: sttProvider,
    TTSProvider: ttsProvider,
    // ... same config structure
})
```

## Using VoiceMeetingSkill

For OmniAgent integration with the meeting skill:

```go
import (
    "github.com/plexusone/omnimeet-core/agent"
    "github.com/plexusone/omnivoice"
    "github.com/plexusone/omnivoice-core/stt"
    "github.com/plexusone/omnivoice-core/tts"
    _ "github.com/plexusone/omnivoice/providers/all"
)

// Get providers
sttProv, _ := omnivoice.GetSTTProvider("deepgram", omnivoice.WithAPIKey(deepgramKey))
ttsProv, _ := omnivoice.GetTTSProvider("elevenlabs", omnivoice.WithAPIKey(elevenLabsKey))

skill, err := agent.NewVoiceMeetingSkill(provider, agent.SkillConfig{
    DefaultMeetingName:   "AI Meeting",
    DefaultAgentName:     "Assistant",
    AutoJoinAsAgent:      true,
    TranscriptionEnabled: true,
}, agent.VoiceSkillConfig{
    STTProvider: sttProv,
    TTSProvider: ttsProv,
    STTConfig: stt.TranscriptionConfig{
        Language:   "en",
        SampleRate: 16000,
        Channels:   1,
    },
    TTSConfig: tts.SynthesisConfig{
        VoiceID:    "rachel",
        SampleRate: 48000,
    },
    OnTranscript: func(meetingID, participantID, participantName, text string, isFinal bool) {
        if isFinal {
            // Process with LLM
            response := processWithLLM(ctx, text)

            // Respond via TTS
            skill.SpeakInMeeting(ctx, meetingID, response)
        }
    },
})
```

## Real-Time Processing Pipeline

### Audio Pipeline

```
Participant Audio → Opus Decode → PCM → STT Provider
                                           ↓
                                     Transcription
                                           ↓
                                     LLM Processing
                                           ↓
Agent Audio ← Opus Encode ← PCM ← TTS Provider
```

### Code Example

```go
func processParticipantAudio(ctx context.Context, voiceAgent *voice.VoiceAgentParticipant, llm llmclient.Client) {
    // Audio is automatically routed to STT via VoiceAgentParticipant
    // OnTranscript callback handles results

    voiceAgent.OnTranscript(func(participantID, name, text string, isFinal bool) {
        if !isFinal {
            // Interim result - could show typing indicator
            return
        }

        // Final transcript - process with LLM
        go func() {
            response, err := llm.Complete(ctx, text)
            if err != nil {
                log.Printf("LLM error: %v", err)
                return
            }

            // Speak response
            if err := voiceAgent.Speak(ctx, response); err != nil {
                log.Printf("TTS error: %v", err)
            }
        }()
    })
}
```

## Audio Configuration

### STT Configuration (from omnivoice-core)

```go
import "github.com/plexusone/omnivoice-core/stt"

config := stt.TranscriptionConfig{
    Language:       "en",        // ISO 639-1 code
    Model:          "nova-2",    // Provider-specific model
    SampleRate:     16000,       // Audio sample rate
    Channels:       1,           // Mono or stereo
    Punctuation:    true,        // Auto-punctuation
    WordTimestamps: true,        // Word-level timing
}
```

### TTS Configuration (from omnivoice-core)

```go
import "github.com/plexusone/omnivoice-core/tts"

config := tts.SynthesisConfig{
    VoiceID:      "alloy",       // Voice identifier
    Model:        "tts-1-hd",    // Provider-specific model
    SampleRate:   48000,         // Output sample rate
    OutputFormat: "pcm",         // Audio format
    Speed:        1.0,           // Speech speed (0.5-2.0)
}
```

## Transcript Management

### Getting Transcripts

```go
// Via MeetingSkill
skill.GetMeetingTranscript(ctx, meetingID)

// Via VoiceMeetingSkill session
session := skill.GetSession(meetingID)
for _, entry := range session.Transcript {
    fmt.Printf("[%s] %s: %s\n",
        entry.Timestamp.Format(time.RFC3339),
        entry.ParticipantName,
        entry.Text,
    )
}
```

### Transcript Entry

```go
type TranscriptEntry struct {
    ParticipantID   string
    ParticipantName string
    Text            string
    Timestamp       time.Time
    IsFinal         bool
}
```

## Performance Considerations

1. **Buffer audio** - Use adequately sized buffers for smooth processing
2. **Concurrent processing** - Run STT/TTS in goroutines
3. **Rate limiting** - Respect provider rate limits
4. **Caching** - Cache TTS for repeated phrases
5. **Compression** - Use Opus for efficient audio transmission
6. **Sample rate** - STT typically works best at 16kHz, meetings use 48kHz

## Provider-Specific Tips

### Deepgram

```go
sttProvider, _ := omnivoice.GetSTTProvider("deepgram",
    omnivoice.WithAPIKey(key),
    omnivoice.WithModel("nova-2"),           // Best accuracy
    omnivoice.WithOption("smart_format", true), // Smart formatting
)
```

### OpenAI

```go
// STT (Whisper)
sttProvider, _ := omnivoice.GetSTTProvider("openai",
    omnivoice.WithAPIKey(key),
)

// TTS
ttsProvider, _ := omnivoice.GetTTSProvider("openai",
    omnivoice.WithAPIKey(key),
    omnivoice.WithVoice("alloy"), // alloy, echo, fable, onyx, nova, shimmer
)
```

### ElevenLabs

```go
ttsProvider, _ := omnivoice.GetTTSProvider("elevenlabs",
    omnivoice.WithAPIKey(key),
    omnivoice.WithVoice("rachel"),
    omnivoice.WithModel("eleven_turbo_v2"), // Lower latency
)
```

## Error Handling

```go
voiceAgent := voice.NewVoiceAgentParticipant(baseAgent, voice.Config{
    // ...
})

// Handle transcription errors in the callback
voiceAgent.OnTranscript(func(participantID, name, text string, isFinal bool) {
    // Process transcript...
})

// For TTS errors
if err := voiceAgent.Speak(ctx, text); err != nil {
    if errors.Is(err, context.Canceled) {
        // Agent left meeting
        return
    }
    log.Printf("TTS error: %v", err)
    // Fallback or retry logic
}
```

## Next Steps

- [API Reference: Voice](../api/voice.md) - Detailed API documentation
- [Testing](testing.md) - Write tests with mock providers
- [Provider Implementation](provider-implementation.md) - Create custom providers
- [OmniVoice Documentation](https://github.com/plexusone/omnivoice) - Full OmniVoice docs
