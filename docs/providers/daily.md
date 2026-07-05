# Daily Provider

The Daily provider implements OmniMeet interfaces for [Daily](https://www.daily.co/), an easy-to-use video API platform.

!!! warning "Coming Soon"
    The Daily provider is planned for Phase 3 and is not yet implemented.

## Planned Installation

```bash
go get github.com/plexusone/omni-daily
```

## Configuration

```go
import "github.com/plexusone/omni-daily/omnimeet"

provider, err := omnimeet.NewProvider(omnimeet.Config{
    APIKey: os.Getenv("DAILY_API_KEY"),
})
```

### Configuration Options

| Option | Description | Required |
|--------|-------------|----------|
| `APIKey` | Daily API key | Yes |
| `Domain` | Custom domain (optional) | No |

## Daily.co Setup

1. Sign up at [Daily.co](https://www.daily.co/)
2. Get your API key from the dashboard
3. Set the environment variable

```bash
export DAILY_API_KEY=your-api-key
```

## Mapping to Daily Concepts

| OmniMeet | Daily |
|----------|-------|
| Meeting | Room |
| Participant | Participant |
| Track | Track |
| Join Token | Meeting Token |

## Planned Features

### Meeting Management

```go
// Create meeting (Daily Room)
m, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{
    Name: "team-standup",
    Metadata: map[string]string{
        "team": "engineering",
    },
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
        Name: "Alice",
        Kind: participant.KindHuman,
    },
})
```

### Agent Participation

Agent participation will use Daily's REST API and WebRTC:

```go
factory := provider.(provider.AgentParticipantFactory)

agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,
})

agent.JoinMeeting(ctx, meetingID, tok)
```

## Dependencies

The Daily provider will depend on:

- `github.com/plexusone/daily-go` - Daily REST API client
- WebRTC libraries for media

## Implementation Status

| Feature | Status |
|---------|--------|
| Meeting CRUD | Planned |
| Token generation | Planned |
| Participant listing | Planned |
| Agent participation | Planned |
| Audio subscription | Planned |
| Audio publishing | Planned |
| Data messages | Planned |

## Comparison with LiveKit

| Feature | LiveKit | Daily |
|---------|---------|-------|
| Self-hosting | Yes | No |
| Ease of use | Medium | High |
| Media control | Full | Limited |
| Pricing | Usage-based | Per-participant-minute |
| Best for | Custom deployments | Quick integration |

## Contributing

The Daily provider implementation is tracked in the [ROADMAP](../specs/ROADMAP.md). Contributions are welcome!
