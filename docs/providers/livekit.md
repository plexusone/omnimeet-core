# LiveKit Provider

The LiveKit provider implements OmniMeet interfaces for [LiveKit](https://livekit.io/), a high-performance WebRTC infrastructure.

## Installation

```bash
go get github.com/plexusone/omni-livekit
```

## Configuration

```go
import "github.com/plexusone/omni-livekit/omnimeet"

provider, err := omnimeet.NewProvider(omnimeet.Config{
    APIKey:    os.Getenv("LIVEKIT_API_KEY"),
    APISecret: os.Getenv("LIVEKIT_API_SECRET"),
    ServerURL: os.Getenv("LIVEKIT_URL"),
})
```

### Configuration Options

| Option | Description | Required |
|--------|-------------|----------|
| `APIKey` | LiveKit API key | Yes |
| `APISecret` | LiveKit API secret | Yes |
| `ServerURL` | LiveKit server URL (ws:// or wss://) | Yes |

## LiveKit Cloud

1. Sign up at [LiveKit Cloud](https://cloud.livekit.io/)
2. Create a new project
3. Get your credentials from the project settings

```bash
export LIVEKIT_URL=wss://your-project.livekit.cloud
export LIVEKIT_API_KEY=your-api-key
export LIVEKIT_API_SECRET=your-api-secret
```

## Local Development

Run LiveKit locally with Docker:

```bash
docker run --rm -p 7880:7880 -p 7881:7881 -p 7882:7882/udp \
    livekit/livekit-server \
    --dev \
    --bind 0.0.0.0
```

```bash
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=devkey
export LIVEKIT_API_SECRET=secret
```

## Mapping to LiveKit Concepts

| OmniMeet | LiveKit |
|----------|---------|
| Meeting | Room |
| Participant | Participant |
| Track | Track |
| Agent | Agent (via livekit-server-sdk-go) |

## Features

### Meeting Management

```go
// Create meeting (LiveKit Room)
m, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{
    Name: "team-standup",
    Metadata: map[string]string{
        "team": "engineering",
    },
    EmptyTimeout: 5 * time.Minute,  // Maps to LiveKit's empty_timeout
})

// List meetings (active rooms)
meetings, _ := provider.ListMeetings(ctx)

// Delete meeting (delete room)
provider.DeleteMeeting(ctx, meetingID)
```

### Token Generation

```go
tok, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: meetingID,
    Participant: participant.Info{
        Name:     "Alice",
        Identity: "alice@example.com",
        Kind:     participant.KindHuman,
    },
    // Token expiry (default: 1 hour)
    TTL: 2 * time.Hour,
})

// tok.Token contains the JWT
// tok.JoinURL contains the full join URL with token
```

### Agent Participation

```go
factory := provider.(provider.AgentParticipantFactory)

agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,
})

// Join uses LiveKit's room client SDK
agent.JoinMeeting(ctx, meetingID, tok)

// Audio uses WebRTC tracks
audioCh, _ := agent.SubscribeToAllAudio(ctx)
for frame := range audioCh {
    // frame.Data is Opus-decoded PCM (with CGO)
    // or raw Opus packets (without CGO)
}

// Publish audio
writer, _ := agent.StartAudioTrack(ctx, provider.AudioTrackOptions{
    SampleRate: 48000,
    Channels:   1,
})
writer.Write(audioFrame)
```

### Data Channels

```go
// Send reliable data (SCTP)
agent.SendDataMessage(ctx, provider.DataMessage{
    Topic:    "chat",
    Data:     []byte("Hello!"),
    Reliable: true,
})

// Send unreliable data (faster, may drop)
agent.SendDataMessage(ctx, provider.DataMessage{
    Topic:    "cursor",
    Data:     positionData,
    Reliable: false,
})
```

## Audio Pipeline

### With CGO (Recommended)

When built with CGO and the `opus` build tag, the provider:

1. Decodes incoming Opus audio to PCM
2. Encodes outgoing PCM audio to Opus

```bash
# macOS
brew install opus libsoxr pkg-config
CGO_ENABLED=1 go build -tags=cgo,opus ./...

# Linux
apt-get install libopus-dev libsoxr-dev pkg-config
CGO_ENABLED=1 go build -tags=cgo,opus ./...
```

### Without CGO

Without CGO, audio frames contain raw data without Opus processing:

```bash
CGO_ENABLED=0 go build ./...
```

This may work for some use cases but is not recommended for production voice applications.

## LiveKit-Specific Extensions

The provider exposes LiveKit-specific features via type assertion:

```go
if lkProvider, ok := provider.(*omnimeet.Provider); ok {
    // Access LiveKit room service directly
    roomService := lkProvider.RoomService()

    // Use LiveKit-specific features
    roomService.MutePublishedTrack(ctx, &livekit.MuteRoomTrackRequest{
        Room:     meetingID,
        Identity: participantID,
        TrackSid: trackID,
        Muted:    true,
    })
}
```

## Error Handling

LiveKit-specific errors are wrapped in OmniMeet errors:

```go
err := provider.CreateMeeting(ctx, req)
if err != nil {
    if errors.Is(err, omnimeet.ErrInvalidToken) {
        // Invalid API credentials
    }
    // Other errors
}
```

## Testing

Run integration tests:

```bash
source .envrc
go test -v -tags=integration ./omnimeet/...
```

## Performance Considerations

1. **Connection pooling** - The provider reuses the room service client
2. **Audio buffers** - Use adequately sized channels for audio (100+ frames)
3. **Reconnection** - LiveKit handles reconnection automatically
4. **Regional deployment** - Use LiveKit Cloud regions close to users

## Limitations

- Video tracks are not fully supported (audio-focused)
- SIP integration not exposed through OmniMeet
- Egress/Ingress require direct LiveKit API access
