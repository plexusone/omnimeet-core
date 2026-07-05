# ADR-003: LiveKit as First Provider

## Status

Accepted

## Date

2026-07-03

## Context

OmniMeet needs to implement multiple providers to validate the abstraction. The choice of first provider significantly impacts the initial interface design.

Candidates considered:

| Provider | Go Support | Self-Hosted | Agent SDK | Complexity |
|----------|-----------|-------------|-----------|------------|
| LiveKit | First-class | Yes | LiveKit Agents | Medium |
| Daily | REST only | No | Python (Pipecat) | Medium |
| Zoom | REST only | No | Meeting SDK | High |
| Twilio Video | REST + SDK | No | No | Medium |
| Agora | SDK | No | No | High |

## Decision

We will implement **LiveKit** as the first provider.

### Rationale

1. **Go-native ecosystem**
   - LiveKit server is written in Go
   - `livekit/server-sdk-go` provides first-class Go support
   - LiveKit Agents framework supports Go agents
   - Natural fit for PlexusOne's Go-first architecture

2. **Clean abstraction model**
   - Room/Participant/Track maps directly to our Meeting/Participant/Track
   - Event-driven architecture aligns with our event model
   - Webhook support for server-side event handling

3. **Agent participation**
   - LiveKit Agents provides a proven pattern for AI meeting participants
   - Go implementation means no Python sidecar required
   - Real-time audio in/out with low latency

4. **Self-hosted option**
   - Can run LiveKit server locally for development
   - No vendor lock-in for production
   - Easier debugging and testing

5. **Active ecosystem**
   - Strong open-source community
   - Frequent releases and updates
   - Good documentation

## Consequences

### Positive

- **Fast iteration** — Go-native means less friction during development
- **Full control** — Can implement agent participation entirely in Go
- **Debugging** — Self-hosted server enables detailed debugging
- **Performance** — No language boundary overhead

### Negative

- **Abstraction bias** — Initial interfaces may be influenced by LiveKit's model
- **Second provider validation** — Must verify abstraction works with Daily's different model

### Mitigations

- Define interfaces based on cross-provider analysis before implementing
- Implement Daily immediately after LiveKit to validate abstraction
- Review interfaces after Daily implementation and refactor if needed

## Implementation Notes

### Dependencies

```go
require (
    github.com/livekit/server-sdk-go v1.x.x
    github.com/livekit/protocol v1.x.x
)
```

### Key LiveKit Concepts to Map

| LiveKit | OmniMeet |
|---------|----------|
| `Room` | `Meeting` |
| `ParticipantInfo` | `Participant` |
| `TrackInfo` | `Track` |
| `RoomServiceClient` | `MeetingProvider` |
| `Room.Join()` | `AgentParticipant.JoinMeeting()` |
| `Participant.PublishTrack()` | `AgentParticipant.PublishAudio()` |

### Agent Participation Strategy

Use LiveKit's agent framework rather than raw WebRTC:

```go
// LiveKit Agent approach (recommended)
agent := livekit.NewAgent(config)
agent.OnTrackSubscribed(handleAudio)
agent.PublishTrack(audioTrack)
```

## Alternatives Considered

### Daily First

Daily has a strong REST API and Pipecat ecosystem, but lacks a Go SDK. Would require:

- Building `daily-go` REST client first
- Using Python sidecar for agent participation
- More complex deployment

Rejected as first provider, but selected as second provider (see ADR-004).

### Zoom First

Zoom is strategically important but has high complexity:

- OAuth2 authentication required
- Meeting SDK for bot attendance
- Complex permission model
- No self-hosted option

Rejected as first provider; better suited for V2 after abstraction is proven.

## References

- [LiveKit Documentation](https://docs.livekit.io/)
- [LiveKit Server SDK Go](https://github.com/livekit/server-sdk-go)
- [LiveKit Agents](https://docs.livekit.io/agents/)
