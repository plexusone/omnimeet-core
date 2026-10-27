# PRD — External Meeting Attendance — Dial-in, Browser, and Zoom SDK Connectors

**Initiative:** `INIT-OMNIMEET-001`
**Repository:** `github.com/plexusone/omnimeet-core`

## Problem

OmniAgent can join LiveKit rooms, but most meetings people attend are hosted
by someone else on Zoom, Google Meet, Microsoft Teams, or Webex. OmniMeet's
provider model assumes the application hosts the meeting and mints the join
token, so it cannot reach those meetings.

Native bot access differs on every platform. Zoom's Meeting SDK needs an
on-behalf-of token from a user in the meeting to join other accounts'
meetings, Google Meet has no open bot API, and Teams raw media needs .NET on
Windows. Almost every one of these meetings offers phone dial-in, which works
the same way everywhere and needs no app registration.

## Goals

- An agent joins any meeting that offers dial-in, given its invite, and
  produces a transcript and summary.
- An agent joins Google Meet and Zoom as a visible participant with a display
  name, and attributes what was said to named speakers.
- Agents learn what each connection supports and adapt (for example, audio
  but no speaker names, or no chat).

## Non-Goals

- Native Teams and Webex connectors. Both are attended by dial-in.
- Hosting meetings on Zoom, Meet, or Teams. `MeetingProvider` covers hosted
  meetings on LiveKit and Daily.
- Calendar sync and scheduled attendance. Those come after joining works.
- Video understanding and screen-share capture.

## Requirements

1. **Attendee sessions.** One interface for joining external meetings, with
   optional capabilities and states that separate "call connected" from
   "admitted to the meeting".
2. **Dial-in.** Parse dial-in details from Zoom, Meet, and Teams invites and
   `.ics` files; dial and enter the meeting ID and passcode through Twilio or
   Telnyx; expose the meeting's audio.
3. **Browser.** Join Google Meet and the Zoom web client with a display name;
   read live captions for speaker names.
4. **Zoom SDK.** Only if the Zoom web client proves unreliable: join through
   the Meeting SDK, with on-behalf-of authorization for other accounts.
5. **OmniAgent.** A tool and CLI command to join a meeting from an invite,
   staying muted unless addressed, with a restricted tool set while in a
   meeting, and a summary with decisions and action items afterwards.

## Success Criteria

- The agent joins a real Zoom, Google Meet, and Teams meeting by dial-in from
  its invite, without manual keypad entry.
- Google Meet and Zoom transcripts from the browser connector name the
  speakers.
- The agent announces itself when admitted and never speaks unless addressed.

## Risks

- Dial-in availability depends on the host's plan and region; invites without
  dial-in fail with a clear error.
- Phone audio is low quality and mixed, so dial-in transcripts lack names.
- Meeting web UIs change, so the browser connector needs ongoing upkeep.
- Spoken instructions from participants are untrusted input to the agent.
