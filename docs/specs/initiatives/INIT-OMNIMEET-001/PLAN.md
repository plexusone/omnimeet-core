# PLAN — External Meeting Attendance — Dial-in, Browser, and Zoom SDK Connectors

**Initiative:** `INIT-OMNIMEET-001`
**Repository:** `github.com/plexusone/omnimeet-core`

1. **Dial-in foundations.** Attendee interfaces, keypad templates, and the
   invite parser in `omnimeet-core`; the keypad-digits call option in
   `omnivoice-core`, implemented by `omni-twilio` and `omni-telnyx`.
2. **Dial-in attendance.** An audio pipeline mode in `omnivoice-core`,
   implemented by the `omni-twilio` and `omni-telnyx` gateways, so calls hand
   raw audio to the application and can be hung up; then the phone joiner and
   simulated meeting in `omnimeet-core`.
3. **OmniAgent meeting attendance.** A meeting attendant library in
   `omniagent` that announces itself, transcribes, and summarizes; then
   answering when addressed, the `omniagent meeting join` CLI command, and
   agent tools. Validate against real Zoom, Meet, and Teams meetings before
   Phase 4.
4. **Browser connector.** Browser core and caption-based speaker names, then
   the Google Meet adapter, then the Zoom web-client adapter; `omniagent`
   picks the browser connector where supported.
5. **Zoom Meeting SDK.** Start only if the Zoom web-client adapter proves
   unreliable.

Each phase ends with a release of the touched modules, so `omniagent` can
depend on tagged versions without `replace` directives.
