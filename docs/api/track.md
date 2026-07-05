# Track Types

Package `track` defines types for media tracks.

## Track

Represents a media stream published by a participant.

```go
type Track struct {
    // ID is the unique track identifier.
    ID string

    // Kind is the track type.
    Kind TrackKind

    // ParticipantID is who published this track.
    ParticipantID string

    // Name is an optional track name.
    Name string

    // Muted indicates if the track is muted.
    Muted bool

    // Metadata is arbitrary key-value data.
    Metadata map[string]string
}
```

## TrackKind

```go
type TrackKind string

const (
    // KindAudio is a microphone or synthesized audio track.
    KindAudio TrackKind = "audio"

    // KindVideo is a camera track.
    KindVideo TrackKind = "video"

    // KindScreenShare is a screen sharing track.
    KindScreenShare TrackKind = "screenshare"

    // KindData is a data channel track.
    KindData TrackKind = "data"
)
```

## SubscribeOptions

Options for subscribing to a track.

```go
type SubscribeOptions struct {
    // Quality specifies the desired quality.
    Quality TrackQuality

    // Adaptive enables adaptive quality.
    Adaptive bool
}
```

## TrackQuality

```go
type TrackQuality string

const (
    QualityLow    TrackQuality = "low"
    QualityMedium TrackQuality = "medium"
    QualityHigh   TrackQuality = "high"
)
```

## Example Usage

```go
import "github.com/plexusone/omnimeet-core/track"

// Handle track published event
agent.OnTrackPublished(func(p participant.Participant, t track.Track) {
    log.Printf("%s published %s track", p.Name, t.Kind)

    switch t.Kind {
    case track.KindAudio:
        // Subscribe to audio
        agent.SubscribeToTrack(ctx, t.ID, track.SubscribeOptions{})

    case track.KindVideo:
        // Subscribe to video with quality preference
        agent.SubscribeToTrack(ctx, t.ID, track.SubscribeOptions{
            Quality:  track.QualityMedium,
            Adaptive: true,
        })

    case track.KindScreenShare:
        // Handle screen share
    }
})

// Check if track is muted
for _, t := range participant.Tracks {
    if t.Kind == track.KindAudio && t.Muted {
        log.Printf("%s has muted audio", participant.Name)
    }
}
```
