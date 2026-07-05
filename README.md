# OmniMeet Core

[![Go Reference](https://pkg.go.dev/badge/github.com/plexusone/omnimeet-core.svg)](https://pkg.go.dev/github.com/plexusone/omnimeet-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/plexusone/omnimeet-core)](https://goreportcard.com/report/github.com/plexusone/omnimeet-core)

**Unified abstraction for real-time collaboration platforms**

OmniMeet enables AI agents to participate in meetings as first-class participants alongside humans, abstracting over LiveKit, Daily, Zoom, Google Meet, Microsoft Teams, and other real-time collaboration platforms.

## Features

- **Provider Agnostic** - Write code once, deploy on any supported platform
- **Agent-First Design** - Built for AI agents, not retrofitted
- **Voice Integration** - Seamless STT/TTS with [OmniVoice](https://github.com/plexusone/omnivoice)
- **OmniAgent Skills** - Ready-to-use tools for meeting management
- **Real-time Events** - Participant joins, track publications, active speakers

## Installation

```bash
go get github.com/plexusone/omnimeet-core
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/plexusone/omnimeet-core/meeting"
    "github.com/plexusone/omni-livekit/omnimeet"
)

func main() {
    ctx := context.Background()

    // Create LiveKit provider
    provider, _ := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    "your-api-key",
        APISecret: "your-api-secret",
        ServerURL: "wss://your-livekit-server",
    })

    // Create a meeting
    m, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{
        Name: "Team Standup",
    })

    log.Printf("Meeting created: %s", m.ID)
}
```

## Packages

| Package | Description |
|---------|-------------|
| `meeting` | Meeting state and lifecycle |
| `participant` | Participant info and roles |
| `track` | Audio/video track types |
| `event` | Real-time meeting events |
| `token` | Authentication tokens |
| `provider` | Provider interface and client |
| `agent` | Agent skill interfaces |
| `skill` | Meeting skill for OmniAgent |
| `voice` | Voice agent types |
| `recording` | Recording state |
| `transcript` | Transcription types |

## Provider Implementations

| Provider | Package | Status |
|----------|---------|--------|
| LiveKit | [omni-livekit](https://github.com/plexusone/omni-livekit) | Available |
| Daily | omni-daily | Planned |
| Zoom | omni-zoom | Planned |

## Architecture

OmniMeet is part of the PlexusOne OmniAgent ecosystem:

```
OmniAgent (orchestration)
     |
     +-- OmniLLM    (reasoning)
     +-- OmniChat   (async messaging)
     +-- OmniVoice  (speech: TTS/STT)
     +-- OmniMemory (conversation memory)
     +-- OmniMeet   (real-time collaboration)  <-- This library
             |
             +-- omni-livekit
             +-- omni-daily
             +-- omni-zoom
```

## Documentation

- [Getting Started](https://plexusone.github.io/omnimeet-core/getting-started/overview/)
- [API Reference](https://plexusone.github.io/omnimeet-core/api/overview/)
- [Agent Participation Guide](https://plexusone.github.io/omnimeet-core/guides/agent-participation/)
- [Voice Integration](https://plexusone.github.io/omnimeet-core/guides/voice-integration/)

## Requirements

- Go 1.21 or later

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
