# ADR-007: OmniVoice Integration Strategy

## Status

**Implemented** - OmniMeet now uses omnivoice-core types directly.

## Date

2026-07-03

## Context

OmniMeet and OmniVoice serve complementary but distinct purposes:

| Library | Purpose | Owns |
|---------|---------|------|
| OmniVoice | Speech processing | TTS, STT, voice pipelines, audio codecs |
| OmniMeet | Real-time collaboration | Meetings, participants, tracks, events |

For AI agents to participate in meetings, both libraries must work together:

1. OmniMeet receives audio from meeting participants
2. OmniVoice transcribes audio (STT)
3. OmniAgent reasons about the transcript
4. OmniVoice synthesizes response (TTS)
5. OmniMeet publishes audio back to the meeting

The question is: **how should these libraries integrate?**

## Decision

OmniMeet will **depend on OmniVoice-Core** for STT/TTS capabilities, but will **not embed voice pipelines**. Instead, OmniMeet provides audio in a format compatible with OmniVoice, and higher-level orchestration (OmniAgent) connects them.

### Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        OmniAgent                            │
│                                                             │
│   ┌─────────────┐     ┌──────────────┐     ┌─────────────┐ │
│   │  OmniMeet   │────▶│  OmniVoice   │────▶│  OmniLLM    │ │
│   │ (audio in)  │     │    (STT)     │     │ (reasoning) │ │
│   └─────────────┘     └──────────────┘     └──────────────┘ │
│         ▲                                        │          │
│         │              ┌──────────────┐          │          │
│         │              │  OmniVoice   │◀─────────┘          │
│         └──────────────│    (TTS)     │                     │
│                        └──────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

### Audio Format Contract

OmniMeet will use a standard audio format compatible with OmniVoice:

```go
// OmniMeet AudioFrame
type AudioFrame struct {
    ParticipantID string
    Data          []byte      // PCM16 audio data
    SampleRate    int         // 16000, 24000, or 48000
    Channels      int         // 1 (mono) or 2 (stereo)
    Timestamp     time.Time
}
```

This maps directly to OmniVoice's expectations:

```go
// OmniVoice STT expects
type TranscriptionConfig struct {
    SampleRate int         // 16000 recommended for STT
    Channels   int         // 1 (mono) recommended
    Encoding   AudioEncoding // PCM16
}

// OmniVoice TTS produces
type SynthesisResult struct {
    Audio      []byte      // PCM16 or provider-specific format
    SampleRate int
    Format     AudioFormat
}
```

### Conversion Utilities

OmniMeet will provide conversion utilities for common scenarios:

```go
package audio

// Convert meeting audio to OmniVoice STT format
func ToSTTFormat(frame AudioFrame) ([]byte, stt.TranscriptionConfig) {
    config := stt.TranscriptionConfig{
        SampleRate: frame.SampleRate,
        Channels:   frame.Channels,
        Encoding:   stt.EncodingPCM16,
    }

    // Resample if necessary (e.g., 48kHz → 16kHz for STT)
    data := resampleIfNeeded(frame.Data, frame.SampleRate, 16000)

    return data, config
}

// Convert OmniVoice TTS output to meeting audio format
func FromTTSResult(result *tts.SynthesisResult, targetSampleRate int) AudioFrame {
    data := resampleIfNeeded(result.Audio, result.SampleRate, targetSampleRate)

    return AudioFrame{
        Data:       data,
        SampleRate: targetSampleRate,
        Channels:   1,
    }
}
```

### Helper: Voice-Enabled Agent Participant

OmniMeet will provide an optional helper that wraps the integration:

```go
package agent

// VoiceAgentParticipant wraps AgentParticipant with OmniVoice integration
type VoiceAgentParticipant struct {
    participant  AgentParticipant
    sttProvider  stt.Provider
    ttsProvider  tts.Provider
    onTranscript func(participantID string, transcript string)
}

func NewVoiceAgentParticipant(
    participant AgentParticipant,
    sttProvider stt.Provider,
    ttsProvider tts.Provider,
) *VoiceAgentParticipant {
    return &VoiceAgentParticipant{
        participant: participant,
        sttProvider: sttProvider,
        ttsProvider: ttsProvider,
    }
}

// StartTranscribing begins real-time transcription of a participant
func (v *VoiceAgentParticipant) StartTranscribing(ctx context.Context, participantID string) error {
    audioCh, err := v.participant.SubscribeToAudio(ctx, participantID)
    if err != nil {
        return err
    }

    go func() {
        for frame := range audioCh {
            data, config := audio.ToSTTFormat(frame)
            result, err := v.sttProvider.Transcribe(ctx, data, config)
            if err != nil {
                continue
            }
            if v.onTranscript != nil {
                v.onTranscript(participantID, result.Text)
            }
        }
    }()

    return nil
}

// Speak synthesizes text and publishes audio to the meeting
func (v *VoiceAgentParticipant) Speak(ctx context.Context, text string) error {
    config := tts.SynthesisConfig{
        Voice:      "default",
        SampleRate: 48000, // Meeting standard
    }

    result, err := v.ttsProvider.Synthesize(ctx, text, config)
    if err != nil {
        return err
    }

    frame := audio.FromTTSResult(result, 48000)
    return v.participant.PublishAudio(ctx, frame)
}

// OnTranscript registers callback for transcription updates
func (v *VoiceAgentParticipant) OnTranscript(handler func(participantID string, transcript string)) {
    v.onTranscript = handler
}
```

### Usage in OmniAgent

```go
// In OmniAgent meeting skill
func (s *MeetingSkill) JoinMeeting(ctx context.Context, meetingID string) error {
    // Get providers
    meetingProvider := omnimeet.GetMeetingProvider("livekit")
    sttProvider := omnivoice.GetSTTProvider("deepgram")
    ttsProvider := omnivoice.GetTTSProvider("elevenlabs")

    // Create agent participant
    participant := meetingProvider.CreateAgentParticipant()

    // Wrap with voice capabilities
    voiceAgent := agent.NewVoiceAgentParticipant(participant, sttProvider, ttsProvider)

    // Join meeting
    token, _ := meetingProvider.CreateJoinToken(ctx, ...)
    participant.JoinMeeting(ctx, meetingID, token)

    // Handle transcriptions
    voiceAgent.OnTranscript(func(participantID, transcript string) {
        // Process with OmniAgent
        response := s.agent.Process(ctx, sessionID, transcript)

        // Speak response
        voiceAgent.Speak(ctx, response)
    })

    // Start transcribing all participants
    for _, p := range participant.RemoteParticipants() {
        voiceAgent.StartTranscribing(ctx, p.ID)
    }

    return nil
}
```

## Consequences

### Positive

- **Clean separation** — OmniMeet handles meetings, OmniVoice handles speech
- **Flexibility** — Users can choose any STT/TTS provider
- **Testability** — Each layer can be tested independently
- **Reusability** — OmniVoice works beyond meetings (phone calls, voice notes)

### Negative

- **Integration code** — Users must connect OmniMeet and OmniVoice
- **Format conversion** — May need audio resampling between systems

### Mitigations

- `VoiceAgentParticipant` helper handles common integration patterns
- `audio` package provides conversion utilities
- OmniAgent's meeting skill encapsulates the full integration

## Alternatives Considered

### OmniMeet Embeds OmniVoice

Have OmniMeet directly depend on and embed OmniVoice pipelines.

Rejected because:

- Tight coupling reduces flexibility
- OmniVoice is useful outside meetings
- Users may want different STT/TTS providers for different participants

### OmniVoice Owns Meeting Audio

Have OmniVoice's gateway handle meeting audio.

Rejected because:

- OmniVoice gateway is designed for telephony
- Meetings have multi-participant, multi-track complexity
- Different lifecycle and event models

### No Integration Helper

Let users integrate manually without helpers.

Rejected because:

- Common patterns should be easy
- Audio format conversion is error-prone
- Better DX with helpers

## References

- [OmniVoice-Core STT](https://github.com/plexusone/omnivoice-core/tree/main/stt)
- [OmniVoice-Core TTS](https://github.com/plexusone/omnivoice-core/tree/main/tts)
- [OmniVoice-Core Audio](https://github.com/plexusone/omnivoice-core/tree/main/audio)
