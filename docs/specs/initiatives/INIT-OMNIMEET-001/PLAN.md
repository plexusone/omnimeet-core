# PLAN — External Meeting Attendance — Dial-in, Browser, and Zoom SDK Connectors

**Initiative:** `INIT-OMNIMEET-001`
**Repository:** `github.com/plexusone/omnimeet-core`

1. **Dial-in foundations.** Attendee interfaces, keypad templates, and the
   invite parser in `omnimeet-core`; the keypad-digits call option in
   `omnivoice-core`, implemented by `omni-twilio` and `omni-telnyx`.
2. **Dial-in attendance.** The phone joiner and simulated meeting in
   `omnimeet-core`, then the join-meeting tool and CLI in `omniagent`.
   Validate against real Zoom, Meet, and Teams meetings before Phase 3.
3. **Browser connector.** Browser core and caption-based speaker names, then
   the Google Meet adapter, then the Zoom web-client adapter; `omniagent`
   picks the browser connector where supported.
4. **Zoom Meeting SDK.** Start only if the Zoom web-client adapter proves
   unreliable.

Each phase ends with a release of the touched modules, so `omniagent` can
depend on tagged versions without `replace` directives.
