# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for OmniMeet.

ADRs document significant architectural decisions made during the design and development of OmniMeet. Each ADR describes the context, decision, consequences, and alternatives considered.

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](001-meeting-abstraction.md) | Meeting as Primary Abstraction | Accepted | 2026-07-03 |
| [002](002-provider-pattern.md) | Provider Pattern Alignment with PlexusOne | Accepted | 2026-07-03 |
| [003](003-livekit-first.md) | LiveKit as First Provider | Accepted | 2026-07-03 |
| [004](004-daily-second.md) | Daily as Second Provider | Accepted | 2026-07-03 |
| [005](005-agent-participation.md) | Agent Participation Architecture | Accepted | 2026-07-03 |
| [006](006-event-model.md) | Event-Driven Architecture | Accepted | 2026-07-03 |
| [007](007-omnivoice-integration.md) | OmniVoice Integration Strategy | Accepted | 2026-07-03 |
| [008](008-frontend-sdk.md) | Frontend SDK Architecture | Proposed | 2026-07-03 |

## ADR Status Definitions

- **Proposed** — Under discussion, not yet accepted
- **Accepted** — Decision made and documented
- **Deprecated** — No longer applies, replaced by another ADR
- **Superseded** — Replaced by a newer ADR (linked in document)

## Creating New ADRs

When creating a new ADR:

1. Copy the template below
2. Use the next sequential number (e.g., `009-topic.md`)
3. Fill in all sections
4. Submit for review

### ADR Template

```markdown
# ADR-NNN: Title

## Status

Proposed | Accepted | Deprecated | Superseded by [ADR-XXX](XXX-topic.md)

## Date

YYYY-MM-DD

## Context

What is the issue that we're seeing that is motivating this decision?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

### Positive

What becomes easier or possible as a result of this change?

### Negative

What becomes more difficult as a result of this change?

### Neutral

Any other impacts.

## Alternatives Considered

What other options were considered and why were they rejected?

## References

Links to relevant documentation, issues, or discussions.
```

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
