# Participant Types

Package `participant` defines types for meeting participants.

## Participant

Represents an entity in a meeting.

```go
type Participant struct {
    // ID is the unique identifier.
    ID string

    // Name is the display name.
    Name string

    // Identity is the user identity (email, user ID, etc.).
    Identity string

    // Kind is the participant type.
    Kind ParticipantKind

    // JoinedAt is when the participant joined.
    JoinedAt time.Time

    // Metadata is arbitrary key-value data.
    Metadata map[string]string

    // Tracks lists the participant's published tracks.
    Tracks []track.Track
}
```

## ParticipantKind

```go
type ParticipantKind string

const (
    // KindHuman is a real user.
    KindHuman ParticipantKind = "human"

    // KindAgent is an AI agent.
    KindAgent ParticipantKind = "agent"

    // KindRecorder is a system recorder.
    KindRecorder ParticipantKind = "recorder"

    // KindObserver is a read-only participant.
    KindObserver ParticipantKind = "observer"
)
```

## Info

Used when creating tokens or joining meetings.

```go
type Info struct {
    // Name is the display name.
    Name string

    // Identity is the user identity.
    Identity string

    // Kind is the participant type.
    Kind ParticipantKind

    // Metadata is optional key-value data.
    Metadata map[string]string
}
```

## Example Usage

```go
import "github.com/plexusone/omnimeet-core/participant"

// Create token for human
tok, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: meetingID,
    Participant: participant.Info{
        Name:     "Alice",
        Identity: "alice@example.com",
        Kind:     participant.KindHuman,
        Metadata: map[string]string{
            "role": "presenter",
        },
    },
})

// Create token for agent
agentTok, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: meetingID,
    Participant: participant.Info{
        Name:     "AI Assistant",
        Identity: "agent-001",
        Kind:     participant.KindAgent,
    },
})

// Check participant kind
for _, p := range agent.RemoteParticipants() {
    switch p.Kind {
    case participant.KindHuman:
        // Process human participant
    case participant.KindAgent:
        // Another agent
    }
}
```
