# API Reference Overview

This section documents the core types and interfaces in OmniMeet.

## Package Structure

```
github.com/plexusone/omnimeet-core
├── meeting/      # Meeting types
├── participant/  # Participant types
├── track/        # Track types
├── token/        # Token types
├── event/        # Event types
├── provider/     # Provider interfaces
├── skill/        # OmniAgent skill (omniskill-compatible)
├── agent/        # Agent types (internal)
├── voice/        # Voice integration types
├── recording/    # Recording types
├── transcript/   # Transcript types
└── errors.go     # Error types
```

## Core Interfaces

### MeetingProvider

The main interface for meeting management:

```go
type MeetingProvider interface {
    Name() string

    // Meeting lifecycle
    CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error)
    GetMeeting(ctx context.Context, id string) (*meeting.Meeting, error)
    ListMeetings(ctx context.Context) ([]*meeting.Meeting, error)
    DeleteMeeting(ctx context.Context, id string) error

    // Tokens
    CreateJoinToken(ctx context.Context, req token.CreateRequest) (*token.JoinToken, error)

    // Participants
    GetParticipants(ctx context.Context, meetingID string) ([]participant.Participant, error)
    RemoveParticipant(ctx context.Context, meetingID, participantID string) error
}
```

### AgentParticipantFactory

Creates agent participants:

```go
type AgentParticipantFactory interface {
    CreateAgentParticipant(opts AgentParticipantOptions) (AgentParticipant, error)
}
```

### AgentParticipant

Media-plane interface for agents:

```go
type AgentParticipant interface {
    // Connection
    JoinMeeting(ctx context.Context, meetingID string, tok *token.JoinToken) error
    LeaveMeeting(ctx context.Context) error
    ConnectionState() ConnectionState

    // Meeting info
    Meeting() *meeting.Meeting
    LocalParticipant() *participant.Participant
    RemoteParticipants() []participant.Participant
    GetParticipant(participantID string) *participant.Participant

    // Audio
    SubscribeToAudio(ctx context.Context, participantID string) (<-chan AudioFrame, error)
    SubscribeToAllAudio(ctx context.Context) (<-chan AudioFrame, error)
    PublishAudio(ctx context.Context, frame AudioFrame) error
    StartAudioTrack(ctx context.Context, opts AudioTrackOptions) (AudioWriter, error)
    StopAudioTrack(ctx context.Context) error

    // Tracks
    SubscribeToTrack(ctx context.Context, trackID string, opts track.SubscribeOptions) error
    UnsubscribeFromTrack(ctx context.Context, trackID string) error

    // Data
    SendDataMessage(ctx context.Context, msg DataMessage) error
    OnDataMessage(handler func(DataMessage))

    // Events
    OnParticipantJoined(handler func(participant.Participant))
    OnParticipantLeft(handler func(participant.Participant))
    OnTrackPublished(handler func(participant.Participant, track.Track))
    OnTrackUnpublished(handler func(participant.Participant, track.Track))
    OnActiveSpeakerChanged(handler func([]participant.Participant))
    Events() <-chan event.Event
}
```

## Type Categories

| Category | Package | Description |
|----------|---------|-------------|
| [Meeting](meeting.md) | `meeting` | Meeting lifecycle and metadata |
| [Participant](participant.md) | `participant` | Participant info and kinds |
| [Track](track.md) | `track` | Media tracks |
| [Provider](provider.md) | `provider` | Provider interfaces and types |
| [Skill](skill.md) | `skill` | OmniAgent skill (omniskill-compatible) |
| [Agent](agent.md) | `agent` | Internal agent types |
| [Voice](voice.md) | `voice` | STT/TTS integration |

## Import Paths

```go
import (
    "github.com/plexusone/omnimeet-core/meeting"
    "github.com/plexusone/omnimeet-core/participant"
    "github.com/plexusone/omnimeet-core/track"
    "github.com/plexusone/omnimeet-core/token"
    "github.com/plexusone/omnimeet-core/event"
    "github.com/plexusone/omnimeet-core/provider"
    "github.com/plexusone/omnimeet-core/agent"
    "github.com/plexusone/omnimeet-core/voice"
)
```

## Error Handling

OmniMeet defines standard errors in `errors.go`:

```go
var (
    ErrMeetingNotFound      = errors.New("meeting not found")
    ErrParticipantNotFound  = errors.New("participant not found")
    ErrNotConnected         = errors.New("not connected to meeting")
    ErrAlreadyConnected     = errors.New("already connected to meeting")
    ErrInvalidToken         = errors.New("invalid join token")
    ErrProviderNotSupported = errors.New("provider does not support this operation")
)
```

Use `errors.Is()` for checking:

```go
if errors.Is(err, omnimeet.ErrMeetingNotFound) {
    // Handle not found
}
```
