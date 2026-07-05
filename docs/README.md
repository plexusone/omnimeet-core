# OmniMeet Documentation

This directory contains documentation for PlexusOne OmniMeet.

## Building the Docs

This documentation uses [MkDocs](https://www.mkdocs.org/) with the Material theme.

### Prerequisites

```bash
pip install mkdocs mkdocs-material mkdocs-minify-plugin
```

### Serve Locally

```bash
cd /path/to/omnimeet-core
mkdocs serve
```

Then open http://127.0.0.1:8000 in your browser.

### Build Static Site

```bash
mkdocs build
```

Output will be in the `site/` directory.

## Contents

```
docs/
├── index.md                    # Home page
├── getting-started/
│   ├── overview.md            # Core concepts
│   ├── installation.md        # Setup instructions
│   └── quickstart.md          # First steps
├── guides/
│   ├── provider-implementation.md   # Implementing providers
│   ├── agent-participation.md       # Agent features
│   ├── voice-integration.md         # STT/TTS integration
│   └── testing.md                   # Testing guide
├── api/
│   ├── overview.md            # API reference overview
│   ├── meeting.md             # Meeting types
│   ├── participant.md         # Participant types
│   ├── track.md               # Track types
│   ├── provider.md            # Provider interfaces
│   ├── agent.md               # Agent skill types
│   └── voice.md               # Voice integration types
├── providers/
│   ├── livekit.md             # LiveKit provider
│   └── daily.md               # Daily provider (planned)
├── adr/                       # Architecture Decision Records
│   ├── README.md
│   └── 001-008 ADRs
└── specs/
    └── ROADMAP.md             # Project roadmap
```

## Quick Links

- **[Home](index.md)** — Overview and quick start
- **[Installation](getting-started/installation.md)** — Setup instructions
- **[Quick Start](getting-started/quickstart.md)** — Create your first meeting
- **[Agent Participation](guides/agent-participation.md)** — Join meetings as an AI agent
- **[ROADMAP](specs/ROADMAP.md)** — Project vision and phased plan
- **[ADR Index](adr/README.md)** — Architecture Decision Records

## Development Status

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 (V0.1) | Complete | Core interfaces + LiveKit provider |
| Phase 2 (V0.2) | Complete | Agent participation + Voice |
| Phase 3 (V0.3) | Planned | Daily provider |
| Phase 4 (V0.4) | Planned | Recording & transcripts |
| Phase 5 (V0.5) | Planned | Frontend SDK |
| Phase 6 (V1.0) | Planned | Production release |

## Related Projects

- [omniagent](https://github.com/plexusone/omniagent) — AI agent runtime
- [omnivoice](https://github.com/plexusone/omnivoice) — Voice provider registry (Deepgram, OpenAI, ElevenLabs, etc.)
- [omnivoice-core](https://github.com/plexusone/omnivoice-core) — Voice/speech interfaces and types
- [omnillm-core](https://github.com/plexusone/omnillm-core) — LLM abstraction
- [omnichat](https://github.com/plexusone/omnichat) — Messaging abstraction
