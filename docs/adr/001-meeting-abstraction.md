# ADR-001: Meeting as Primary Abstraction

## Status

Accepted

## Date

2026-07-03

## Context

PlexusOne needs a unified abstraction for real-time collaboration platforms. Several naming options were considered:

- **OmniVideo** — Too narrow; video is just one capability
- **OmniRoom** — Exposes LiveKit's internal terminology; doesn't fit Zoom/Teams
- **OmniSession** — Too generic; conflicts with auth/login sessions
- **OmniConference** — Enterprise-sounding; doesn't fit 1:1 AI conversations
- **OmniLive** — Too broad; could overlap with OmniVoice or future live features
- **OmniMeeting** / **OmniMeet** — Matches the actual abstraction

## Decision

We will use **Meeting** as the primary abstraction, with the package named **OmniMeet**.

A Meeting represents:

- A live collaborative session
- Containing participants (humans, agents, recorders, observers)
- With media streams (audio, video, screen share, data)
- Supporting real-time events and state

This maps cleanly across providers:

| OmniMeet | LiveKit | Daily | Zoom | Teams | Slack |
|----------|---------|-------|------|-------|-------|
| Meeting | Room | Room | Meeting | Meeting | Huddle |
| Participant | Participant | Participant | Participant | Participant | Participant |
| Track | Track | Track | Stream | Stream | N/A |

## Consequences

### Positive

- **Clear mental model** — Developers immediately understand what a Meeting is
- **Provider-neutral** — Doesn't expose implementation details of any provider
- **Extensible** — Meetings can include humans, AI agents, recorders, etc.
- **Consistent with industry** — Matches Zoom, Teams, Meet terminology

### Negative

- **LiveKit terminology mismatch** — LiveKit uses "Room" internally; our mapping layer handles this
- **Name collision** — "OmniMeeting" is used by at least one AI meeting product; we differentiate with "PlexusOne OmniMeet"

### Neutral

- The shorter form "OmniMeet" is used for package names (`omnimeet-core`, `omnimeet`)
- Documentation uses "PlexusOne OmniMeet" for disambiguation

## Alternatives Considered

### OmniRoom

Rejected because "Room" is LiveKit-specific. Zoom and Teams use "Meeting", and exposing provider terminology in the abstraction would bias the design.

### OmniVideo

Rejected because video is just one media type. Audio-only meetings, SIP calls, and data-only sessions are all valid use cases.

### OmniCollab

Considered for its future-proof scope, but rejected because it doesn't immediately imply real-time media or meetings.

## References

- [IDEATION_CHAT.md](../../IDEATION_CHAT.md) — Original naming discussion
- [LiveKit Rooms Documentation](https://docs.livekit.io/intro/basics/rooms-participants-tracks/)
