# ADR-005: Agent Participation Architecture

## Status

Accepted

## Date

2026-07-03

## Context

OmniMeet's primary use case is enabling AI agents to participate in meetings alongside humans. This requires:

1. Joining meetings as a participant
2. Receiving audio from other participants
3. Publishing audio responses
4. Handling real-time events
5. Integrating with OmniVoice (STT/TTS) and OmniAgent (reasoning)

The architecture must support different provider capabilities:

| Provider | Go Agent Support | Agent Strategy |
|----------|-----------------|----------------|
| LiveKit | Full (LiveKit Agents) | Native Go |
| Daily | None (Python only) | Python sidecar or browser |
| Zoom | Limited (Meeting SDK) | Bot process |

## Decision

We will define a separate **`AgentParticipant`** interface that represents an AI agent's participation in a meeting, distinct from the control-plane **`MeetingProvider`** interface.

### Interface Design

```go
// MeetingProvider handles control plane operations
type MeetingProvider interface {
    CreateMeeting(ctx, req) (*Meeting, error)
    CreateJoinToken(ctx, req) (*JoinToken, error)
    // ... other control plane operations
}

// AgentParticipant handles media plane operations for AI agents
type AgentParticipant interface {
    // Join/leave
    JoinMeeting(ctx context.Context, meetingID string, token *JoinToken) error
    LeaveMeeting(ctx context.Context) error

    // Audio handling
    SubscribeToAudio(ctx context.Context, participantID string) (<-chan AudioFrame, error)
    SubscribeToAllAudio(ctx context.Context) (<-chan AudioFrame, error)
    PublishAudio(ctx context.Context, frame AudioFrame) error

    // Events
    OnParticipantJoined(handler func(Participant))
    OnParticipantLeft(handler func(Participant))
    OnTrackPublished(handler func(Track))
    OnTrackUnpublished(handler func(Track))

    // Data messages
    SendDataMessage(ctx context.Context, msg DataMessage) error
    OnDataMessage(handler func(DataMessage))

    // State
    Meeting() *Meeting
    LocalParticipant() *Participant
    RemoteParticipants() []Participant
}
```

### AudioFrame Type

```go
type AudioFrame struct {
    ParticipantID string
    Data          []byte      // PCM16 audio data
    SampleRate    int         // e.g., 48000
    Channels      int         // e.g., 1 (mono)
    Timestamp     time.Time
}
```

### Provider Implementation Strategies

**LiveKit (Go-native):**

```go
type LiveKitAgentParticipant struct {
    room       *lksdk.Room
    localParticipant *lksdk.LocalParticipant
}

func (a *LiveKitAgentParticipant) JoinMeeting(ctx, meetingID, token) error {
    a.room = lksdk.ConnectToRoom(url, token)
    return nil
}

func (a *LiveKitAgentParticipant) SubscribeToAudio(ctx, participantID) (<-chan AudioFrame, error) {
    // Subscribe to participant's audio track
    // Return channel that emits PCM16 frames
}

func (a *LiveKitAgentParticipant) PublishAudio(ctx, frame) error {
    // Publish audio frame to room
}
```

**Daily (Python sidecar or deferred):**

For V1, Daily provider focuses on control plane. Agent participation options:

1. **Defer to LiveKit** — Use LiveKit for agent participation, Daily for meetings that don't need agents
2. **Python subprocess** — Spawn `daily-python` process, communicate via IPC
3. **Browser automation** — Run headless browser with `daily-js`

### Integration Pattern

```go
// Full agent-in-meeting flow
func RunAgentInMeeting(ctx context.Context, meetingID string) error {
    // 1. Get meeting provider
    meetingProvider := omnimeet.GetMeetingProvider("livekit")

    // 2. Create join token for agent
    token, err := meetingProvider.CreateJoinToken(ctx, JoinTokenRequest{
        MeetingID: meetingID,
        Participant: ParticipantInfo{
            Name: "AI Assistant",
            Kind: ParticipantKindAgent,
        },
    })

    // 3. Create agent participant
    agentParticipant := meetingProvider.CreateAgentParticipant()
    if err := agentParticipant.JoinMeeting(ctx, meetingID, token); err != nil {
        return err
    }
    defer agentParticipant.LeaveMeeting(ctx)

    // 4. Subscribe to all audio
    audioCh, err := agentParticipant.SubscribeToAllAudio(ctx)
    if err != nil {
        return err
    }

    // 5. Process audio with OmniVoice + OmniAgent
    for frame := range audioCh {
        // STT
        transcript, err := omnivoice.TranscribeStream(ctx, frame.Data)
        if err != nil {
            continue
        }

        // Agent reasoning
        response, err := omniagent.Process(ctx, sessionID, transcript)
        if err != nil {
            continue
        }

        // TTS
        audioResponse, err := omnivoice.Synthesize(ctx, response)
        if err != nil {
            continue
        }

        // Publish response
        agentParticipant.PublishAudio(ctx, AudioFrame{
            Data:       audioResponse,
            SampleRate: 48000,
            Channels:   1,
        })
    }

    return nil
}
```

## Consequences

### Positive

- **Clean separation** — Control plane vs media plane clearly separated
- **Provider flexibility** — Each provider implements what it can
- **OmniVoice integration** — AudioFrame design aligns with OmniVoice pipeline
- **Testability** — Can mock AgentParticipant for testing

### Negative

- **Not all providers support full agent participation** — Daily requires workarounds
- **Additional interface complexity** — Two interfaces instead of one

### Mitigations

- Document which providers support AgentParticipant
- Provide helper functions for common patterns
- Consider future `AgentParticipantAdapter` for providers without native support

## Alternatives Considered

### Single Interface

Combine control plane and agent participation in one `MeetingProvider` interface.

Rejected because:

- Not all providers support agent participation
- Separation of concerns is cleaner
- Easier to test control plane separately

### OmniVoice Gateway Integration

Have OmniMeet delegate to OmniVoice's gateway for audio handling.

Rejected because:

- OmniVoice gateway is designed for telephony, not meetings
- Meeting audio has different characteristics (multiple participants, tracks)
- Better to have OmniMeet handle meeting-specific audio, use OmniVoice for STT/TTS

## References

- [LiveKit Go SDK](https://github.com/livekit/server-sdk-go)
- [OmniVoice-Core Pipeline](https://github.com/plexusone/omnivoice-core/tree/main/pipeline)
