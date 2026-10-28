# OmniMeet Core

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omnimeet-core/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnimeet-core
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnimeet-core
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omnimeet-core
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomnimeet-core
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omnimeet-core
 [repo-url]: https://github.com/plexusone/omnimeet-core
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnimeet-core/blob/main/LICENSE

**Unified abstraction for real-time collaboration platforms**

OmniMeet enables AI agents to participate in meetings as first-class participants alongside humans, abstracting over LiveKit, Daily, Zoom, Google Meet, Microsoft Teams, and other real-time collaboration platforms.

## Features

- 🔌 **Provider Agnostic** - Write code once, deploy on any supported platform
- 🤖 **Agent-First Design** - Built for AI agents, not retrofitted
- 🎙️ **Voice Integration** - Seamless STT/TTS with [OmniVoice](https://github.com/plexusone/omnivoice)
- 🛠️ **OmniAgent Skills** - Ready-to-use tools for meeting management
- ⚡ **Real-time Events** - Participant joins, track publications, active speakers

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
| `attendee` | Sessions for attending meetings hosted by others |
| `dialin` | Join meetings by phone: dial-in numbers, keypad sequences, and the dial-in joiner |
| `dialin/dialintest` | Simulated meetings for testing dial-in |
| `invite` | Parse meeting invitations and `.ics` files |

## Attending External Meetings

Agents can also attend meetings hosted by others on Zoom, Google Meet, and Microsoft Teams. Phone dial-in works with any meeting that lists a dial-in number: parse the invite, and the dial-in joiner calls the meeting, enters the meeting ID and passcode as keypad tones, and hands the agent the meeting's audio.

```go
jt, err := invite.Parse(invitationText) // or invite.ParseICS for .ics files
if err != nil {
    return err
}

// gw is an omnivoice-core voice gateway in audio mode, for example from
// omni-twilio or omni-telnyx with Mode: gateway.PipelineModeAudio.
joiner := dialin.NewJoiner(gw, dialin.WithCountry("US"))
sess, err := joiner.Join(ctx, jt)
if err != nil {
    return err // dialin.ErrNoDialIn when the meeting has no phone option
}

audio := sess.(attendee.AudioReceiver).ReceiveAudio() // mixed meeting audio
speaker := sess.(attendee.AudioSender).SendAudio()    // played into the meeting
defer sess.Leave(ctx)
```

The session reports its state through `Events()`: joining, connected, admitted, and left or ended. Phone dial-in cannot see whether the meeting admitted the caller, so the session reports admitted once the keypad sequence has had time to finish.

For tests, `dialin/dialintest` simulates meetings with Zoom, Google Meet, and Teams phone menus, so code can join, listen, speak, and leave without a real meeting. [examples/dialin](examples/dialin/main.go) runs the whole flow against a simulated meeting. See [ADR-009](docs/adr/009-dial-in-first.md) for the connector plan.

To attend from the command line, [OmniAgent](https://github.com/plexusone/omniagent) builds on these packages: `omniagent meeting join --invite meeting.ics` dials in, prints a live transcript, and writes meeting notes with a summary. See its [Attending Meetings by Phone](https://github.com/plexusone/omniagent/blob/main/docs/guides/meetings.md) guide.

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

- Go 1.26 or later

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
