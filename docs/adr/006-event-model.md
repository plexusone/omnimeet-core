# ADR-006: Event-Driven Architecture

## Status

Accepted

## Date

2026-07-03

## Context

Real-time collaboration platforms are inherently event-driven:

- Participants join and leave
- Tracks are published and unpublished
- Recordings start and stop
- Transcripts update in real-time

OmniMeet must expose these events in a provider-neutral way while supporting both:

1. **Server-side event handling** — Webhooks for backend integration
2. **Client-side event handling** — Real-time updates in agent participants

## Decision

We will implement a unified event model with support for multiple delivery mechanisms.

### Event Types

```go
type EventType string

const (
    // Meeting lifecycle
    EventMeetingCreated       EventType = "meeting.created"
    EventMeetingStarted       EventType = "meeting.started"
    EventMeetingEnded         EventType = "meeting.ended"

    // Participant lifecycle
    EventParticipantJoined    EventType = "participant.joined"
    EventParticipantLeft      EventType = "participant.left"
    EventParticipantUpdated   EventType = "participant.updated"

    // Track events
    EventTrackPublished       EventType = "track.published"
    EventTrackUnpublished     EventType = "track.unpublished"
    EventTrackMuted           EventType = "track.muted"
    EventTrackUnmuted         EventType = "track.unmuted"
    EventTrackSubscribed      EventType = "track.subscribed"
    EventTrackUnsubscribed    EventType = "track.unsubscribed"

    // Recording events
    EventRecordingStarted     EventType = "recording.started"
    EventRecordingStopped     EventType = "recording.stopped"
    EventRecordingFailed      EventType = "recording.failed"

    // Transcript events
    EventTranscriptUpdated    EventType = "transcript.updated"
    EventTranscriptFinalized  EventType = "transcript.finalized"

    // Data events
    EventDataMessageReceived  EventType = "data_message.received"

    // Connection events
    EventConnectionQualityChanged EventType = "connection.quality_changed"
    EventReconnecting            EventType = "connection.reconnecting"
    EventReconnected             EventType = "connection.reconnected"
)
```

### Event Structure

```go
type Event struct {
    ID        string            `json:"id"`
    Type      EventType         `json:"type"`
    MeetingID string            `json:"meeting_id"`
    Timestamp time.Time         `json:"timestamp"`
    Data      any               `json:"data"`
}

// Event-specific data types
type ParticipantJoinedData struct {
    Participant Participant `json:"participant"`
}

type TrackPublishedData struct {
    Participant Participant `json:"participant"`
    Track       Track       `json:"track"`
}

type TranscriptUpdatedData struct {
    ParticipantID string    `json:"participant_id"`
    Text          string    `json:"text"`
    IsFinal       bool      `json:"is_final"`
    Confidence    float64   `json:"confidence"`
    Timestamp     time.Time `json:"timestamp"`
}
```

### Event Delivery Mechanisms

#### 1. Callback Registration (Provider Interface)

```go
type EventHandler func(ctx context.Context, event Event) error

type MeetingProvider interface {
    // Register event handler for server-side events
    OnEvent(handler EventHandler)

    // ... other methods
}
```

#### 2. Channel-Based (Agent Participant)

```go
type AgentParticipant interface {
    // Real-time event channel
    Events() <-chan Event

    // Typed handlers for common events
    OnParticipantJoined(handler func(Participant))
    OnParticipantLeft(handler func(Participant))
    OnTrackPublished(handler func(Track))

    // ... other methods
}
```

#### 3. Webhook Handler (Server Integration)

```go
// WebhookHandler processes provider webhooks
type WebhookHandler interface {
    // HandleWebhook processes incoming webhook and returns normalized events
    HandleWebhook(r *http.Request) ([]Event, error)

    // ValidateSignature verifies webhook authenticity
    ValidateSignature(r *http.Request, secret string) error
}

// Provider-specific implementations
type LiveKitWebhookHandler struct{}
type DailyWebhookHandler struct{}
```

### Provider Event Mapping

| OmniMeet Event | LiveKit | Daily |
|----------------|---------|-------|
| `participant.joined` | `ParticipantJoined` | `participant-joined` |
| `participant.left` | `ParticipantLeft` | `participant-left` |
| `track.published` | `TrackPublished` | `track-started` |
| `track.unpublished` | `TrackUnpublished` | `track-stopped` |
| `recording.started` | `EgressStarted` | `recording-started` |
| `recording.stopped` | `EgressEnded` | `recording-stopped` |

### Usage Patterns

#### Server-Side (Webhooks)

```go
// Configure webhook endpoint
meetingProvider := omnimeet.GetMeetingProvider("livekit")
meetingProvider.OnEvent(func(ctx context.Context, event Event) error {
    switch event.Type {
    case omnimeet.EventParticipantJoined:
        data := event.Data.(ParticipantJoinedData)
        log.Printf("Participant joined: %s", data.Participant.Name)
    case omnimeet.EventMeetingEnded:
        log.Printf("Meeting ended: %s", event.MeetingID)
    }
    return nil
})

// HTTP handler for provider webhooks
http.HandleFunc("/webhooks/livekit", func(w http.ResponseWriter, r *http.Request) {
    events, err := meetingProvider.WebhookHandler().HandleWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    for _, event := range events {
        meetingProvider.EmitEvent(event)
    }
    w.WriteHeader(http.StatusOK)
})
```

#### Agent-Side (Real-Time)

```go
agentParticipant.OnParticipantJoined(func(p Participant) {
    log.Printf("Participant joined: %s", p.Name)
    if p.Kind == ParticipantKindHuman {
        // Greet the human
        audioGreeting := omnivoice.Synthesize(ctx, "Hello! How can I help you?")
        agentParticipant.PublishAudio(ctx, audioGreeting)
    }
})

agentParticipant.OnTrackPublished(func(t Track) {
    if t.Kind == TrackKindAudio {
        // Start transcribing this participant's audio
        go transcribeParticipant(ctx, t.ParticipantID)
    }
})
```

## Consequences

### Positive

- **Unified model** — Same event types across all providers
- **Multiple delivery mechanisms** — Webhooks, callbacks, channels
- **Type safety** — Strongly typed event data
- **Provider mapping** — Clear mapping from provider events to OmniMeet events

### Negative

- **Event translation overhead** — Must map provider-specific events
- **Eventual consistency** — Webhooks may arrive after real-time events

### Mitigations

- Event mapping is one-time effort per provider
- Document expected event ordering and timing
- Provide idempotency support via event IDs

## Alternatives Considered

### Only Webhooks

Rejected because agent participants need real-time events, not HTTP callbacks.

### Only Callbacks

Rejected because server-side integrations benefit from webhook-based event delivery.

### Generic Event Bus

Considered using a generic pub/sub system, but rejected because:

- Adds external dependency
- Overkill for single-meeting context
- Provider-native event mechanisms are sufficient

## References

- [LiveKit Webhooks](https://docs.livekit.io/realtime/server/webhooks/)
- [Daily Webhooks](https://docs.daily.co/reference/webhooks)
- [OmniVoice Event Model](https://github.com/plexusone/omnivoice-core/tree/main/observability)
