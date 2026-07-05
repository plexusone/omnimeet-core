# Overview

OmniMeet provides a unified abstraction layer for real-time collaboration platforms, enabling AI agents to join meetings and interact with participants through voice and data channels.

## Core Concepts

### Meeting

A **Meeting** is a live collaborative session containing participants and media streams. It represents the fundamental unit of real-time collaboration.

```go
type Meeting struct {
    ID          string
    Name        string
    Status      MeetingStatus  // Scheduled, Active, Ended
    CreatedAt   time.Time
    StartedAt   *time.Time
    EndedAt     *time.Time
    Metadata    map[string]string
}
```

### Participant

A **Participant** is an entity in a meeting. Participants can be:

- **Human**: Real users joining via web or native clients
- **Agent**: AI agents joining programmatically
- **Recorder**: System participants recording the meeting
- **Observer**: Read-only participants

```go
type Participant struct {
    ID       string
    Name     string
    Kind     ParticipantKind  // Human, Agent, Recorder, Observer
    JoinedAt time.Time
    Metadata map[string]string
}
```

### Track

A **Track** is a media stream published by a participant:

- **Audio**: Microphone or synthesized speech
- **Video**: Camera or screen share
- **Data**: Arbitrary data messages

```go
type Track struct {
    ID            string
    Kind          TrackKind  // Audio, Video, Data
    ParticipantID string
    Muted         bool
}
```

### Provider

A **Provider** implements the OmniMeet interfaces for a specific platform:

- **LiveKit**: High-performance WebRTC infrastructure
- **Daily**: Easy-to-use video API platform
- **Zoom**: Enterprise video conferencing (planned)

## Provider Pattern

OmniMeet follows the PlexusOne provider pattern, where:

1. Core interfaces are defined in `omnimeet-core`
2. Implementations live in separate packages (`omni-livekit`, `omni-daily`)
3. Providers register themselves via `init()` functions
4. Applications retrieve providers by name

```go
// Registration (in provider package)
func init() {
    omnimeet.RegisterProvider("livekit", factory)
}

// Usage (in application)
provider, err := omnimeet.GetMeetingProvider("livekit")
```

## Agent Participation

The key differentiator of OmniMeet is its agent-first design. AI agents can:

1. **Join meetings** as full participants
2. **Subscribe to audio** from other participants
3. **Publish audio** (synthesized speech via TTS)
4. **Send/receive data messages**
5. **React to events** (participant join/leave, active speaker changes)

```go
// Create agent participant
agent, _ := factory.CreateAgentParticipant(AgentParticipantOptions{
    AutoSubscribe: true,
})

// Join meeting
agent.JoinMeeting(ctx, meetingID, token)

// Listen to participants
audioCh, _ := agent.SubscribeToAllAudio(ctx)
for frame := range audioCh {
    // Process audio (send to STT)
}

// Speak in meeting
agent.Speak(ctx, "Hello everyone!")
```

## Voice Integration

OmniMeet integrates with OmniVoice for speech processing:

- **STT (Speech-to-Text)**: Convert participant audio to text
- **TTS (Text-to-Speech)**: Convert agent responses to audio

The `VoiceAgentParticipant` wrapper provides automatic transcription:

```go
voiceAgent := voice.NewVoiceAgentParticipant(baseAgent, voice.Config{
    STTProvider: deepgramProvider,
    TTSProvider: elevenLabsProvider,
    OnTranscript: func(participantID, name, text string, isFinal bool) {
        log.Printf("%s said: %s", name, text)
    },
})
```

## Next Steps

- [Installation](installation.md) - Set up OmniMeet
- [Quick Start](quickstart.md) - Create your first meeting
