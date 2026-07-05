# ADR-004: Daily as Second Provider

## Status

Accepted

## Date

2026-07-03

## Context

After implementing the first provider (LiveKit), we need a second provider to validate that the `omnimeet-core` abstraction is truly provider-neutral.

The second provider should:

1. Be conceptually similar enough to map cleanly
2. Be different enough to stress-test the abstraction
3. Have production-quality APIs
4. Support AI agent use cases

Candidates:

| Provider | Similarity to LiveKit | API Quality | Agent Support | Validation Value |
|----------|----------------------|-------------|---------------|------------------|
| Daily | High (Room model) | Excellent REST | Pipecat/Python | High |
| Zoom | Medium (Meeting model) | Good REST | Meeting SDK | Medium |
| Agora | Medium (Channel model) | Good SDK | Limited | Medium |
| Twilio Video | High (Room model) | Good REST | Limited | Medium |

## Decision

We will implement **Daily** as the second provider.

### Rationale

1. **Similar enough to map**
   - Daily uses Room/Participant model like LiveKit
   - REST API covers meeting lifecycle, tokens, recordings
   - Webhook events map to our event model

2. **Different enough to validate**
   - No first-class Go SDK (REST-only from Go)
   - Agent participation via `daily-python` or browser
   - Different operational model (SaaS-only vs self-hosted)
   - Forces us to ensure abstraction isn't LiveKit-biased

3. **Production-quality APIs**
   - Well-documented REST API
   - Stable and reliable
   - Used in production by major companies

4. **Pipecat ecosystem**
   - Daily maintains Pipecat (open-source voice agent framework)
   - Strong AI agent use case support
   - Validates that OmniMeet can work with existing AI ecosystems

5. **Different SDK strategy tests abstraction**
   - LiveKit: Go-native SDK
   - Daily: REST API + frontend SDKs
   - If abstraction works for both, it's likely provider-neutral

## Consequences

### Positive

- **Abstraction validation** — Two different provider models proves abstraction is sound
- **REST API experience** — Informs how future providers might be implemented
- **Pipecat compatibility** — Validates interop with existing voice agent frameworks

### Negative

- **No Go media SDK** — Agent participation may require Python sidecar or browser
- **Additional dependency** — Need to build `daily-go` REST client

### Mitigations

- Build `daily-go` as a standalone package (useful beyond OmniMeet)
- For V1, focus on control plane (meetings, tokens, recordings)
- Agent participation can use:
  - LiveKit for Go-native agents
  - Daily for browser-based agents
  - Python sidecar if Daily-native agent needed

## Implementation Notes

### daily-go Package

Create a standalone Go client for Daily's REST API:

```
github.com/plexusone/daily-go/
├── client.go       # HTTP client
├── rooms.go        # Rooms API
├── tokens.go       # Meeting tokens API
├── meetings.go     # Active meetings API
├── recordings.go   # Recordings API
├── transcripts.go  # Transcripts API
├── webhooks.go     # Webhook handling
└── types.go        # Request/response types
```

### V1 Scope for omni-daily

Control plane only (no media plane):

```go
type DailyProvider struct {
    client *daily.Client
}

func (p *DailyProvider) CreateMeeting(ctx, req) (*Meeting, error)
func (p *DailyProvider) GetMeeting(ctx, id) (*Meeting, error)
func (p *DailyProvider) EndMeeting(ctx, id) error
func (p *DailyProvider) CreateJoinToken(ctx, req) (*JoinToken, error)
func (p *DailyProvider) ListParticipants(ctx, id) ([]Participant, error)
func (p *DailyProvider) StartRecording(ctx, id, opts) (*Recording, error)
func (p *DailyProvider) StopRecording(ctx, id) (*Recording, error)
```

### Agent Participation Strategy

For Daily, agent participation is deferred to V2 or handled via:

1. **Browser-based agent** — Agent runs in headless browser with `daily-js`
2. **Python sidecar** — Use `daily-python` for media participation
3. **Hybrid** — Control plane in Go, media in Python subprocess

This is acceptable because:

- LiveKit provides full Go-native agent participation
- Daily's strength is the control plane and frontend SDKs
- OmniMeet abstraction is validated at the interface level

## Alternatives Considered

### Zoom Second

Zoom is strategically important but adds different concerns:

- OAuth2 authentication complexity
- SaaS scheduling workflows
- Bot attendance via Meeting SDK

Better as third provider after core abstraction is proven.

### Agora Second

Agora has strong scale but:

- Channel-based model is more different from Room model
- Less AI agent ecosystem
- Would be a good stress-test but not ideal for validation

Better as fourth provider for scale testing.

## References

- [Daily REST API Documentation](https://docs.daily.co/reference/rest-api)
- [Daily Python SDK](https://docs.daily.co/reference/daily-python)
- [Pipecat Framework](https://github.com/pipecat-ai/pipecat)
