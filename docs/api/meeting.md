# Meeting Types

Package `meeting` defines types for meeting lifecycle.

## Meeting

Represents a live collaborative session.

```go
type Meeting struct {
    // ID is the unique identifier for the meeting.
    ID string

    // Name is the human-readable meeting name.
    Name string

    // Status is the current meeting status.
    Status MeetingStatus

    // CreatedAt is when the meeting was created.
    CreatedAt time.Time

    // StartedAt is when the first participant joined.
    StartedAt *time.Time

    // EndedAt is when the meeting ended.
    EndedAt *time.Time

    // Metadata is arbitrary key-value data.
    Metadata map[string]string
}
```

## MeetingStatus

```go
type MeetingStatus string

const (
    StatusScheduled MeetingStatus = "scheduled"  // Created but not started
    StatusActive    MeetingStatus = "active"     // In progress
    StatusEnded     MeetingStatus = "ended"      // Completed
)
```

## CreateRequest

Request to create a new meeting.

```go
type CreateRequest struct {
    // Name is the meeting name (required).
    Name string

    // Metadata is optional key-value data.
    Metadata map[string]string

    // MaxParticipants limits the number of participants (0 = unlimited).
    MaxParticipants int

    // EmptyTimeout is the duration to wait before closing an empty meeting.
    EmptyTimeout time.Duration
}
```

## Example Usage

```go
import "github.com/plexusone/omnimeet-core/meeting"

// Create meeting
m, err := provider.CreateMeeting(ctx, meeting.CreateRequest{
    Name: "Team Standup",
    Metadata: map[string]string{
        "team": "engineering",
        "recurring": "true",
    },
    MaxParticipants: 10,
    EmptyTimeout:    5 * time.Minute,
})

// Check status
if m.Status == meeting.StatusActive {
    log.Printf("Meeting %s is active", m.Name)
}
```
