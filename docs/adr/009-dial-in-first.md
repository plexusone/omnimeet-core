# ADR-009: Dial-in First for External Meetings

## Status

Accepted

## Date

2026-10-27

## Context

OmniMeet's providers so far host meetings the application controls: the provider creates the room and mints a join token. Agents also need to attend meetings hosted by others on Zoom, Google Meet, Microsoft Teams, and Webex. There is no join token to mint, and native bot access differs on each platform:

| Platform | Native bot path | Constraint |
|----------|-----------------|------------|
| Zoom | Meeting SDK | Joining meetings outside the app owner's account requires an on-behalf-of token from a user who is in the meeting |
| Google Meet | Meet Media API | Not open to new developers; no general bot API |
| Microsoft Teams | Graph calling bots | Raw media access requires .NET on Windows |
| Webex | Webex SDKs | Separate integration per platform |

Nearly every one of these meetings offers phone dial-in: a number plus a meeting ID and passcode entered on the keypad. Dial-in works the same way on every platform and needs no platform app registration.

Dial-in has limits. The agent receives one mixed audio stream at telephone quality, appears as a phone number rather than a name, and cannot read chat, see video, or tell speakers apart without diarization. Dial-in availability also depends on the host's plan and region.

## Decision

OmniMeet attends external meetings through connectors that all implement one attendee session interface and report their capabilities. Connectors are added in this order:

1. **Phone dial-in** for every platform, using the existing `omnivoice-core` call system with per-platform keypad templates.
2. **Browser connector** for Google Meet, then the Zoom web client. A headless browser joins as a visible participant with a display name and reads live captions for speaker names.
3. **Zoom Meeting SDK**, only if the Zoom web client proves unreliable in the browser connector.

Teams and Webex stay on dial-in until either becomes common enough to need a native connector.

The attendee interfaces are separate from `MeetingProvider` and `AgentParticipant`, which remain the API for hosted meetings.

## Consequences

### Positive

- One code path reaches nearly all external meetings from the first phase
- No platform app review or registration before the first release
- Reuses the telephony stack (`omnivoice-core`, `omni-twilio`, `omni-telnyx`) instead of a new media pipeline
- The capability model lets agents adapt when a connection lacks names, chat, or video

### Negative

- Dial-in transcripts lack speaker names and have lower accuracy than native audio
- The agent appears as a phone number, so it should announce itself when admitted
- The browser connector needs ongoing maintenance as meeting web UIs change, and runs one browser per meeting
- Dial-in is unavailable on some host plans and regions

## Alternatives Considered

### Native SDK per platform first

Rejected as the starting point. It requires an app registration and review per platform, Google Meet has no open bot API, and Teams raw media requires .NET on Windows. Native connectors remain an option where they add names, chat, or video.

### Third-party meeting bot service

A hosted bot API covers all platforms with names and speaker attribution, but adds a usage-priced dependency outside the Go stack. It remains a fallback if the browser connector proves too costly to maintain.

### LiveKit SIP outbound

LiveKit can dial a meeting's phone number into a LiveKit room, which would let existing LiveKit agents attend unchanged. It requires a LiveKit SIP service and trunk. It is compatible with this decision as an alternative transport for the dial-in connector.

## References

- [ROADMAP Phases 3–5](../specs/ROADMAP.md#phase-3-dial-in-attendance)
- [ADR-005: Agent Participation Architecture](005-agent-participation.md)
- [ADR-007: OmniVoice Integration Strategy](007-omnivoice-integration.md)
