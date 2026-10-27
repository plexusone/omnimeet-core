# ROADMAP — External Meeting Attendance — Dial-in, Browser, and Zoom SDK Connectors

**Initiative:** `INIT-OMNIMEET-001`
**Repository:** `github.com/plexusone/omnimeet-core`

Commits implementing an item carry `Refs: RMI-<REPOSLUG>-<NNN>`.

## Phase 1 — Dial-in Foundations

**Theme:** Interfaces, invite parsing, and keypad dialing across the voice stack

- [x] `RMI-OMNIMEETCORE-001` Attendee session interfaces: `Joiner`, `Session`, optional audio, recorder, and participant interfaces, `Capabilities`, and states that separate connected from admitted
- [x] `RMI-OMNIMEETCORE-002` Package `dialin`: `DialInTarget` and keypad templates for Zoom, Google Meet, and Teams
  - Depends on: `RMI-OMNIMEETCORE-001`
- [x] `RMI-OMNIMEETCORE-003` Invite parser for Zoom, Google Meet, and Teams invitations and `.ics` files
  - Depends on: `RMI-OMNIMEETCORE-002`
- [x] `RMI-OMNIVOICECORE-001` Keypad-digits option for outbound calls, and call options on `Gateway.MakeCall`
- [x] `RMI-OMNITWILIO-001` Send keypad digits on outbound calls with `<Play digits>` before the media stream
  - Depends on: `RMI-OMNIVOICECORE-001`
- [x] `RMI-OMNITELNYX-001` Send keypad digits on outbound calls through call control after answer
  - Depends on: `RMI-OMNIVOICECORE-001`

## Phase 2 — Dial-in Attendance

**Theme:** OmniAgent joins any meeting with dial-in and summarizes it

- [x] `RMI-OMNIVOICECORE-002` Audio pipeline mode and `AudioSession` interface: gateway sessions that hand raw call audio to the application instead of running STT, LLM, and TTS
- [x] `RMI-OMNITWILIO-002` Audio pipeline mode in the Twilio gateway: sessions implement `AudioSession` over Media Streams
  - Depends on: `RMI-OMNIVOICECORE-002`
- [x] `RMI-OMNITELNYX-002` Audio pipeline mode in the Telnyx gateway: sessions implement `AudioSession` over Media Streaming
  - Depends on: `RMI-OMNIVOICECORE-002`
- [x] `RMI-OMNIMEETCORE-004` Phone dial-in joiner over an audio-mode voice gateway, exposing the meeting's mixed audio as one `KindSIP` participant
  - Depends on: `RMI-OMNIMEETCORE-002`
  - Depends on: `RMI-OMNIVOICECORE-002`
  - Depends on: `RMI-OMNITWILIO-002`
- [x] `RMI-OMNIMEETCORE-005` Simulated meeting phone menu for tests, and `examples/dialin`
  - Depends on: `RMI-OMNIMEETCORE-004`
- [ ] `RMI-OMNIAGENT-411` Join meetings by phone from an invite: tool and CLI command, announce on admission, muted unless addressed, restricted tools, summary with decisions and action items
  - Depends on: `RMI-OMNIMEETCORE-003`
  - Depends on: `RMI-OMNIMEETCORE-004`

## Phase 3 — Browser Connector

**Theme:** Named, visible participant in Google Meet and the Zoom web client

- [ ] `RMI-OMNIMEETCORE-006` Browser joiner core in a separate provider module: browser lifecycle, display name, call audio in and out, waiting rooms, removal and meeting-end detection
  - Depends on: `RMI-OMNIMEETCORE-001`
- [ ] `RMI-OMNIMEETCORE-007` Speaker-attributed transcript events and the `SpeakerNames` capability
  - Depends on: `RMI-OMNIMEETCORE-001`
- [ ] `RMI-OMNIMEETCORE-008` Google Meet adapter with optional signed-in bot account and caption-based speaker names
  - Depends on: `RMI-OMNIMEETCORE-006`
  - Depends on: `RMI-OMNIMEETCORE-007`
- [ ] `RMI-OMNIMEETCORE-009` Zoom web-client adapter, recording whether the web client is reliable for bots
  - Depends on: `RMI-OMNIMEETCORE-006`
  - Depends on: `RMI-OMNIMEETCORE-007`
- [ ] `RMI-OMNIAGENT-412` Choose the browser connector where the platform is supported, with dial-in fallback; speaker names in summaries
  - Depends on: `RMI-OMNIAGENT-411`
  - Depends on: `RMI-OMNIMEETCORE-008`

## Phase 4 — Zoom Meeting SDK

**Theme:** Native Zoom attendance, only if the Zoom web-client adapter proves unreliable

- [ ] `RMI-OMNIMEETCORE-010` Zoom Meeting SDK joiner in a separate provider module, with app JWT for the app owner's meetings
  - Depends on: `RMI-OMNIMEETCORE-009`
- [ ] `RMI-OMNIMEETCORE-011` On-behalf-of authorization for joining other accounts' Zoom meetings
  - Depends on: `RMI-OMNIMEETCORE-010`
