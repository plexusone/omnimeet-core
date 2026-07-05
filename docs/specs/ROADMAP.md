# OmniMeet Roadmap

> **PlexusOne OmniMeet** — AI-native meeting and real-time collaboration abstraction for Go.

This document outlines the vision, architecture, and phased implementation plan for OmniMeet, the real-time collaboration layer of the PlexusOne ecosystem.

## Table of Contents

- [Vision](#vision)
- [Goals](#goals)
- [Non-Goals (V1)](#non-goals-v1)
- [Architecture Overview](#architecture-overview)
- [Core Abstractions](#core-abstractions)
- [Provider Strategy](#provider-strategy)
- [Integration with PlexusOne Ecosystem](#integration-with-plexusone-ecosystem)
- [Phased Roadmap](#phased-roadmap)
- [ADR Index](#adr-index)

---

## Vision

OmniMeet provides a unified abstraction for real-time collaboration platforms—LiveKit, Daily, Zoom, Google Meet, Microsoft Teams, Discord Voice, and Slack Huddles—enabling AI agents to participate in meetings as first-class participants alongside humans.

**Key insight:** A meeting is not primarily about video. It's a **live collaborative session** containing participants, media streams, shared context, and real-time events. Video is just one capability.

```
OmniAgent
     │
     ├── OmniLLM      (reasoning)
     ├── OmniChat     (async messaging)
     ├── OmniVoice    (speech: TTS/STT)
     ├── OmniMemory   (conversation memory)
     └── OmniMeet     (real-time collaboration)  ← NEW
             │
             ├── omni-livekit
             ├── omni-daily
             └── omni-zoom
```

---

## Goals

### V1 Goals

1. **Meeting lifecycle management** — Create, get, list, and end meetings
2. **Participant management** — Join/leave humans and AI agents
3. **Audio-first** — Real-time audio conversation between humans and agents
4. **Event-driven** — Rich event stream (join, leave, track published, transcript updated)
5. **Provider-neutral** — Same interface works across LiveKit, Daily, and eventually Zoom
6. **Agent participation** — AI agents join meetings and interact via OmniVoice/OmniLLM
7. **Transcript capture** — Real-time transcription via OmniVoice STT
8. **Recording hooks** — Start/stop recording via provider APIs

### V1.5 Goals

1. **Video support** — Optional video tracks for participants
2. **Screen share metadata** — Detect and expose screen share events
3. **Data channels** — In-meeting messaging and data exchange
4. **Invite integration** — Generate join links, integrate with OmniChat for delivery

### V2 Goals

1. **Scheduling** — Calendar integration (Google, Outlook, CalDAV)
2. **Recurring meetings** — Series management
3. **Breakout rooms** — Sub-meetings within larger meetings
4. **Meeting bots** — Attend existing Zoom/Teams/Meet meetings as a bot

---

## Non-Goals (V1)

The following are explicitly out of scope for V1:

- Full calendar/scheduling system (→ future OmniCalendar)
- Recurring meeting management
- Complex RSVP workflows
- Breakout rooms
- Whiteboards / collaborative documents
- Deep video understanding / computer vision
- Full Zoom/Teams SaaS bot attendance (V2)
- WebRTC implementation details (delegated to providers)

---

## Architecture Overview

### Package Structure

```
github.com/plexusone/omnimeet-core/
├── meeting/           # Meeting type and lifecycle
├── participant/       # Participant types and management
├── track/             # Media tracks (audio, video, data)
├── recording/         # Recording abstraction
├── transcript/        # Transcript types (integrates with omnivoice)
├── event/             # Meeting events
├── provider/          # Provider interface and multi-provider client
├── registry/          # Provider factory registration
├── config/            # Configuration types
├── errors/            # Domain errors
└── docs/
    ├── specs/         # Specifications and roadmap
    └── adr/           # Architecture Decision Records

github.com/plexusone/omni-livekit/    # LiveKit provider
github.com/plexusone/omni-daily/      # Daily provider (depends on daily-go)
github.com/plexusone/daily-go/        # Daily REST API client
github.com/plexusone/omnimeet/        # Bundle with common providers
```

### Dependency Graph

```
omnimeet (bundle)
    │
    ├── omnimeet-core (interfaces)
    │
    ├── omni-livekit
    │       └── livekit/server-sdk-go
    │
    └── omni-daily
            └── daily-go (REST client)
                    └── Daily REST API
```

---

## Core Abstractions

### Meeting

The primary abstraction representing a live collaborative session.

```go
type Meeting struct {
    ID           string
    Name         string
    Status       MeetingStatus      // pending, active, ended
    Provider     string             // "livekit", "daily", "zoom"
    Participants []Participant
    CreatedAt    time.Time
    StartedAt    *time.Time
    EndedAt      *time.Time
    Metadata     map[string]string
}

type MeetingStatus string

const (
    MeetingStatusPending MeetingStatus = "pending"
    MeetingStatusActive  MeetingStatus = "active"
    MeetingStatusEnded   MeetingStatus = "ended"
)
```

### Participant

An entity that joins a meeting—human, AI agent, recorder, or observer.

```go
type Participant struct {
    ID         string
    MeetingID  string
    Kind       ParticipantKind    // human, agent, recorder, observer, sip
    Name       string
    Identity   string             // Provider-specific identity
    JoinedAt   time.Time
    LeftAt     *time.Time
    Tracks     []Track
    Metadata   map[string]string
}

type ParticipantKind string

const (
    ParticipantKindHuman    ParticipantKind = "human"
    ParticipantKindAgent    ParticipantKind = "agent"
    ParticipantKindRecorder ParticipantKind = "recorder"
    ParticipantKindObserver ParticipantKind = "observer"
    ParticipantKindSIP      ParticipantKind = "sip"
)
```

### Track

A media stream published by a participant.

```go
type Track struct {
    ID            string
    ParticipantID string
    Kind          TrackKind          // audio, video, screen_share, data
    Source        TrackSource        // microphone, camera, screen, unknown
    Muted         bool
    Metadata      map[string]string
}

type TrackKind string

const (
    TrackKindAudio       TrackKind = "audio"
    TrackKindVideo       TrackKind = "video"
    TrackKindScreenShare TrackKind = "screen_share"
    TrackKindData        TrackKind = "data"
)
```

### MeetingProvider Interface

The core provider interface that all implementations must satisfy.

```go
type MeetingProvider interface {
    // Provider identification
    Name() string

    // Meeting lifecycle
    CreateMeeting(ctx context.Context, req CreateMeetingRequest) (*Meeting, error)
    GetMeeting(ctx context.Context, meetingID string) (*Meeting, error)
    ListMeetings(ctx context.Context, opts ListMeetingsOptions) ([]Meeting, error)
    EndMeeting(ctx context.Context, meetingID string) error

    // Participant management
    CreateJoinToken(ctx context.Context, req JoinTokenRequest) (*JoinToken, error)
    RemoveParticipant(ctx context.Context, meetingID, participantID string) error
    ListParticipants(ctx context.Context, meetingID string) ([]Participant, error)

    // Events
    OnEvent(handler EventHandler)

    // Recording (optional capability)
    StartRecording(ctx context.Context, meetingID string, opts RecordingOptions) (*Recording, error)
    StopRecording(ctx context.Context, meetingID string) (*Recording, error)

    // Lifecycle
    Close() error
}
```

### AgentParticipant Interface

For AI agents that actively participate in meetings.

```go
type AgentParticipant interface {
    // Join as an agent participant
    JoinMeeting(ctx context.Context, meetingID string, token *JoinToken) error

    // Audio handling
    SubscribeToAudio(ctx context.Context, participantID string) (<-chan []byte, error)
    PublishAudio(ctx context.Context, audio []byte) error

    // Events
    OnTrackPublished(handler TrackHandler)
    OnParticipantJoined(handler ParticipantHandler)
    OnParticipantLeft(handler ParticipantHandler)

    // Data messages
    SendDataMessage(ctx context.Context, msg DataMessage) error
    OnDataMessage(handler DataMessageHandler)

    // Lifecycle
    LeaveMeeting(ctx context.Context) error
}
```

### Events

Rich event stream for meeting lifecycle.

```go
type Event struct {
    Type      EventType
    MeetingID string
    Timestamp time.Time
    Data      any
}

type EventType string

const (
    EventMeetingStarted       EventType = "meeting.started"
    EventMeetingEnded         EventType = "meeting.ended"
    EventParticipantJoined    EventType = "participant.joined"
    EventParticipantLeft      EventType = "participant.left"
    EventTrackPublished       EventType = "track.published"
    EventTrackUnpublished     EventType = "track.unpublished"
    EventTrackMuted           EventType = "track.muted"
    EventTrackUnmuted         EventType = "track.unmuted"
    EventTranscriptUpdated    EventType = "transcript.updated"
    EventRecordingStarted     EventType = "recording.started"
    EventRecordingStopped     EventType = "recording.stopped"
    EventDataMessageReceived  EventType = "data_message.received"
)
```

---

## Provider Strategy

### Provider Implementation Order

| Order | Provider | Rationale |
|-------|----------|-----------|
| 1 | **LiveKit** | Go-native, WebRTC SFU, strong agent SDK, ideal for Go-first architecture |
| 2 | **Daily** | REST API + SDKs, different model validates abstraction, Pipecat ecosystem |
| 3 | **Zoom** | SaaS integration, different concerns (scheduling, invitations, OAuth) |
| 4 | **Agora** | Large-scale RTC, stress-tests abstraction at scale |

### Provider Comparison

| Capability | LiveKit | Daily | Zoom |
|------------|---------|-------|------|
| Core model | Room/Participant/Track | Room/Participant | Meeting/Participant |
| Go SDK | First-class | REST API (no Go SDK) | REST API |
| Agent participation | LiveKit Agents (Go) | daily-python sidecar | Meeting SDK |
| Self-hosted | Yes | No (SaaS) | No (SaaS) |
| Recording | Egress API | REST API | Cloud recording |
| SIP/PSTN | Built-in | Built-in | Add-on |

### Implementation Approach

**omni-livekit:**
- Full Go implementation using `livekit/server-sdk-go`
- Agent participation via LiveKit Agents framework
- Control plane + media plane in Go

**omni-daily:**
- Create `daily-go` REST client for control plane
- Use `daily-js` for browser participants (frontend)
- Python sidecar optional for server-side agent participation

**omni-zoom (future):**
- REST API for meeting management
- Meeting SDK for bot attendance
- OAuth2 for authentication

---

## Integration with PlexusOne Ecosystem

### OmniVoice Integration

OmniMeet uses OmniVoice for speech processing within meetings.

```
Meeting
    │
    ├── Human speaks
    │       │
    │       ▼
    │   Agent subscribes to audio
    │       │
    │       ▼
    │   OmniVoice STT
    │       │
    │       ▼
    │   OmniLLM (reasoning)
    │       │
    │       ▼
    │   OmniVoice TTS
    │       │
    │       ▼
    │   Agent publishes audio
    │       │
    │       ▼
    └── Human hears response
```

### OmniAgent Integration

OmniAgent orchestrates agent behavior during meetings.

```go
// Agent joining a meeting
agent := omniagent.New(config)
meetingProvider := omnimeet.GetProvider("livekit")

meeting, _ := meetingProvider.CreateMeeting(ctx, omnimeet.CreateMeetingRequest{
    Name: "Sales Demo",
})

token, _ := meetingProvider.CreateJoinToken(ctx, omnimeet.JoinTokenRequest{
    MeetingID:   meeting.ID,
    Participant: omnimeet.ParticipantInfo{
        Name: "AI Sales Assistant",
        Kind: omnimeet.ParticipantKindAgent,
    },
})

// Agent joins and processes audio
agentParticipant := meetingProvider.CreateAgentParticipant()
agentParticipant.JoinMeeting(ctx, meeting.ID, token)

audioStream, _ := agentParticipant.SubscribeToAudio(ctx, humanParticipantID)
for audio := range audioStream {
    // STT → OmniAgent → TTS → PublishAudio
    transcript := omnivoice.Transcribe(ctx, audio)
    response := agent.Process(ctx, sessionID, transcript)
    audioResponse := omnivoice.Synthesize(ctx, response)
    agentParticipant.PublishAudio(ctx, audioResponse)
}
```

### OmniChat Integration

Meeting invites can be delivered via OmniChat.

```go
meeting, _ := meetingProvider.CreateMeeting(ctx, req)
joinLink := meetingProvider.GetJoinLink(meeting.ID)

// Send invite via Slack
slackProvider := omnichat.GetProvider("slack")
slackProvider.Send(ctx, channelID, omnichat.OutgoingMessage{
    Content: fmt.Sprintf("Join the meeting: %s", joinLink),
})
```

### OmniMemory Integration

Meeting context and transcripts flow into memory.

```go
// After meeting ends
transcript := meetingProvider.GetTranscript(ctx, meetingID)
memory.Store(ctx, omnimemory.StoreRequest{
    Content:  transcript.Text,
    Metadata: map[string]string{
        "type":       "meeting_transcript",
        "meeting_id": meetingID,
    },
})
```

---

## Phased Roadmap

### Phase 1: Foundation (V0.1) - COMPLETED

**Goal:** Core interfaces and LiveKit provider

**Deliverables:**

- [x] `omnimeet-core` package structure
- [x] Core types: `Meeting`, `Participant`, `Track`, `Event`
- [x] `MeetingProvider` interface
- [x] `AgentParticipant` interface
- [x] Provider registry system
- [x] `omni-livekit` implementation
  - [x] Meeting lifecycle (create, get, list, end)
  - [x] Join token generation
  - [x] Participant management
  - [x] Event streaming via callbacks
- [x] Basic example: human + agent in LiveKit room
- [ ] Webhook handler implementation

**Dependencies:**
- `livekit/server-sdk-go`

**Completed:** 2026-07-03

### Phase 2: Agent Participation (V0.2) - COMPLETED

**Goal:** Full agent participation in LiveKit meetings

**Deliverables:**

- [x] LiveKit agent participant implementation
  - [x] Join meeting as agent
  - [x] Subscribe to participant audio (interface)
  - [x] Publish audio responses (interface)
  - [x] Data message send/receive
- [x] Audio pipeline implementation
  - [x] Opus decoding for incoming audio
  - [x] Opus encoding for outgoing audio
  - [x] PCM16 frame handling
- [x] OmniVoice integration
  - [x] STT pipeline for incoming audio
  - [x] TTS pipeline for outgoing audio
  - [x] VoiceAgentParticipant helper
- [x] OmniAgent integration
  - [x] Agent processes meeting conversation
  - [x] Tool calling during meetings (MeetingSkill with 10 tools)
  - [x] VoiceMeetingSkill for STT/TTS integration
- [x] Demo: "Talk to an OmniAgent in a LiveKit room"
  - [x] examples/omniagent-meeting demo application

**Dependencies:**
- `omnivoice-core` (STT/TTS)
- `omniagent` (agent runtime)

**Completed:** 2026-07-03

### Phase 3: Daily Provider (V0.3)

**Goal:** Second provider validates abstraction

**Deliverables:**

- [ ] `daily-go` REST API client
  - [ ] Rooms API
  - [ ] Meeting tokens API
  - [ ] Participants API
  - [ ] Recordings API
  - [ ] Webhooks
- [ ] `omni-daily` provider implementation
  - [ ] Meeting lifecycle
  - [ ] Join token generation
  - [ ] Participant management
  - [ ] Event streaming
- [ ] Validate `omnimeet-core` abstraction against two providers
- [ ] Demo: Same agent code works with LiveKit and Daily

**Dependencies:**
- Daily REST API

### Phase 4: Recording & Transcripts (V0.4)

**Goal:** Recording and transcript capture

**Deliverables:**

- [ ] Recording abstraction
  - [ ] Start/stop recording
  - [ ] Recording status events
  - [ ] Recording retrieval
- [ ] Transcript integration
  - [ ] Real-time transcript via OmniVoice STT
  - [ ] Store transcript via OmniStorage
  - [ ] Transcript events
- [ ] LiveKit Egress integration
- [ ] Daily recording API integration

### Phase 5: Frontend SDK (V0.5)

**Goal:** TypeScript/React SDK for browser participants

**Deliverables:**

- [ ] `@plexusone/omnimeet-js` - Core TypeScript library
- [ ] `@plexusone/omnimeet-react` - React components
  - [ ] `<MeetingRoom>` component
  - [ ] `<ParticipantTile>` component
  - [ ] `<Controls>` component (mute, leave, screen share)
  - [ ] `<Transcript>` component
- [ ] Provider adapters
  - [ ] LiveKit React integration
  - [ ] Daily React integration
- [ ] OmniAgent embedded UI integration

### Phase 6: Bundle & Polish (V1.0)

**Goal:** Production-ready release

**Deliverables:**

- [ ] `omnimeet` bundle package
- [ ] Comprehensive documentation
- [ ] Example applications
- [ ] CI/CD pipeline
- [ ] Performance benchmarks
- [ ] Security review

---

## ADR Index

Architecture Decision Records are stored in `/docs/adr/`.

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](../adr/001-meeting-abstraction.md) | Meeting as Primary Abstraction | Accepted |
| [ADR-002](../adr/002-provider-pattern.md) | Provider Pattern Alignment with PlexusOne | Accepted |
| [ADR-003](../adr/003-livekit-first.md) | LiveKit as First Provider | Accepted |
| [ADR-004](../adr/004-daily-second.md) | Daily as Second Provider | Accepted |
| [ADR-005](../adr/005-agent-participation.md) | Agent Participation Architecture | Accepted |
| [ADR-006](../adr/006-event-model.md) | Event-Driven Architecture | Accepted |
| [ADR-007](../adr/007-omnivoice-integration.md) | OmniVoice Integration Strategy | Accepted |
| [ADR-008](../adr/008-frontend-sdk.md) | Frontend SDK Architecture | Proposed |

---

## Success Criteria

### V1.0 Success Criteria

1. **Abstraction validation:** Same OmniAgent code works with both LiveKit and Daily
2. **Latency:** Agent response time < 500ms (excluding LLM latency)
3. **Reliability:** 99.9% uptime for meeting sessions
4. **Developer experience:** Create meeting + agent join in < 20 lines of code
5. **Documentation:** Complete API reference and examples

### Demo Scenario

```
1. Human joins LiveKit room via browser
2. AI agent joins same room
3. Human speaks: "What's on my calendar today?"
4. Agent hears via OmniMeet → OmniVoice STT
5. Agent reasons via OmniLLM with calendar tool
6. Agent responds via OmniVoice TTS → OmniMeet
7. Human hears: "You have a team standup at 10am and..."
8. Transcript is captured and stored
9. Meeting ends, recording is available
```

---

## References

- [IDEATION_CHAT.md](../../IDEATION_CHAT.md) - Original ideation discussion
- [LiveKit Documentation](https://docs.livekit.io/)
- [Daily Documentation](https://docs.daily.co/)
- [PlexusOne OmniVoice-Core](https://github.com/plexusone/omnivoice-core)
- [PlexusOne OmniAgent](https://github.com/plexusone/omniagent)
