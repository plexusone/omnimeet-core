# TRD — External Meeting Attendance — Dial-in, Browser, and Zoom SDK Connectors

**Initiative:** `INIT-OMNIMEET-001`
**Repository:** `github.com/plexusone/omnimeet-core`

The decision and alternatives are recorded in
[ADR-009](../../../adr/009-dial-in-first.md).

## Attendee Session Interfaces

New interfaces in `omnimeet-core`, separate from `MeetingProvider` and
`AgentParticipant`, which stay the API for hosted meetings:

- `Joiner.Join(ctx, JoinTarget) (Session, error)`
- `Session`: `ID`, `Events`, `State`, `Leave`
- Optional, discovered by type assertion: `AudioIn`, `AudioOut`, `Recorder`,
  `ParticipantLister`
- `Capabilities`: audio in, audio out, display name, speaker names, chat,
  video
- States: `dialing`, `connected`, `waiting` (waiting room or host absent),
  `admitted`, `left`, `removed`, `ended`, `failed`

Audio uses `omnivoice-core` types; `omnimeet-core` adds no audio format of its
own.

## Dial-in

- `dialin.DialInTarget`: number, meeting ID, passcode, platform.
- Keypad templates per platform, as digit and pause steps:
  - Zoom: meeting ID `#`, participant ID skipped with `#`, passcode `#`
  - Google Meet: PIN `#`
  - Teams: conference ID `#`
- Invite parser: Zoom, Meet, and Teams invitation text and `.ics`
  descriptions; picks a number by country; typed error when no dial-in exists.
- `omnivoice-core`: a keypad-digits call option, and options on
  `Gateway.MakeCall`.
- `omni-twilio`: digits sent with `<Play digits>` before `<Connect><Stream>`,
  which avoids the 32-character limit of the `SendDigits` call parameter.
- `omni-telnyx`: DTMF sent through call control after answer.
- Gateways run a conversation pipeline (STT, LLM, TTS, or a realtime model)
  on every call, which suits one-to-one calls but not meetings. An audio
  pipeline mode in `omnivoice-core` (`PipelineModeAudio`) skips the pipeline;
  its sessions implement `AudioSession`, which hands the call's audio to the
  application and plays the application's audio into the call. `omni-twilio`
  and `omni-telnyx` implement it over their media streams.
- The joiner places the call with `gateway.MakeCallWithOptions` and the keypad
  digits, takes the session's `AudioSession`, exposes the meeting's mixed audio
  as one `KindSIP` participant, and reports audio-only capabilities.

## Browser Connector

- Separate provider module so `omnimeet-core` has no browser dependency;
  candidate base: `w3pilot`. The module and its repository are created in
  Phase 3; its RMIs stay on `omnimeet-core` until then.
- One headless browser per meeting: join with a display name, capture call
  audio, play agent audio into the call, handle waiting rooms, detect removal
  and meeting end.
- Speaker names come from live captions and are attached to transcript
  events.
- Google Meet supports an optional signed-in bot account to avoid guest
  admission.

## Zoom Meeting SDK

Built only if the Zoom web-client adapter proves unreliable. Linux Meeting
SDK bot in a separate module; app JWT for the app owner's meetings, and an
OAuth flow for on-behalf-of tokens for other accounts' meetings, which
require the authorizing user to be present.

## OmniAgent Integration

- Tool and CLI command: join from an invite or `.ics` file, leave, status.
- Connector choice: browser when the platform is supported, dial-in
  otherwise.
- Announces itself when admitted; stays muted unless addressed by name.
- A restricted tool set while in a meeting, since speech from participants is
  untrusted input.
- Transcript, summary, decisions, and action items at meeting end.

## Testing

- Unit tests for keypad templates and the invite parser.
- A simulated meeting phone menu that checks the keypad sequence and plays
  scripted audio, so the dial-in joiner is tested without a real meeting.
- Manual acceptance against real Zoom, Meet, and Teams meetings.
