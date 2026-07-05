# OmniMeet

**Unified abstraction for real-time collaboration platforms**

OmniMeet enables AI agents to participate in meetings as first-class participants alongside humans, abstracting over LiveKit, Daily, Zoom, Google Meet, Microsoft Teams, and other real-time collaboration platforms.

## Why OmniMeet?

- 🔌 **Provider Agnostic**: Write code once, deploy on any supported platform
- 🤖 **Agent-First Design**: Built for AI agents, not retrofitted
- 🎙️ **Voice Integration**: Seamless STT/TTS with [OmniVoice](https://github.com/plexusone/omnivoice) - switch between Deepgram, OpenAI, ElevenLabs, and more
- 🛠️ **OmniAgent Skills**: Ready-to-use tools for meeting management

## Quick Example

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

## Packages

| Package | Description |
|---------|-------------|
| `omnimeet-core` | Core interfaces and types |
| `omni-livekit` | LiveKit provider implementation |
| `omni-daily` | Daily provider implementation (planned) |
| `daily-go` | Daily REST API client |

## Current Status

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 (V0.1) | Complete | Core interfaces + LiveKit provider |
| Phase 2 (V0.2) | Complete | Agent participation + Voice |
| Phase 3 (V0.3) | Planned | Daily provider |
| Phase 4 (V0.4) | Planned | Recording & transcripts |
| Phase 5 (V0.5) | Planned | Frontend SDK |
| Phase 6 (V1.0) | Planned | Production release |

## Next Steps

- [Installation](getting-started/installation.md) - Set up OmniMeet in your project
- [Quick Start](getting-started/quickstart.md) - Create your first meeting
- [Agent Participation](guides/agent-participation.md) - Join meetings as an AI agent
