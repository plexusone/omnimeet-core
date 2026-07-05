# ADR-002: Provider Pattern Alignment with PlexusOne

## Status

Accepted

## Date

2026-07-03

## Context

PlexusOne has established a consistent provider pattern across its libraries:

- **`-core` packages** define interfaces and types
- **`omni-{provider}` packages** implement provider-specific logic
- **Bundle packages** aggregate common providers with auto-registration
- **Registry pattern** enables runtime provider discovery

Examples:

```
omnillm-core    → omni-anthropic, omni-openai    → omnillm
omnivoice-core  → omni-deepgram, omni-elevenlabs → omnivoice
omnistorage-core → omni-sqlite, omni-postgres    → omnistorage
```

OmniMeet must follow this pattern for consistency.

## Decision

We will structure OmniMeet following the established PlexusOne provider pattern:

```
omnimeet-core   → omni-livekit, omni-daily, omni-zoom → omnimeet
```

### Package Responsibilities

**`omnimeet-core`:**
- Defines `MeetingProvider` interface
- Defines core types: `Meeting`, `Participant`, `Track`, `Event`
- Provides `provider.Client[T]` for multi-provider management
- Provides registry for factory-based instantiation
- No external dependencies beyond stdlib

**`omni-{provider}`:**
- Implements `MeetingProvider` for specific provider
- Depends on provider SDK (e.g., `livekit/server-sdk-go`)
- Registers via `init()` with priority system
- Handles provider-specific configuration

**`omnimeet`:**
- Imports and re-exports `omnimeet-core`
- Imports common providers to trigger auto-registration
- Provides "batteries-included" experience

### Registry Pattern

```go
// In omnimeet-core/registry.go
func RegisterMeetingProvider(name string, factory ProviderFactory, priority int)
func GetMeetingProvider(name string, opts ...ProviderOption) (MeetingProvider, error)
func ListMeetingProviders() []string
func HasMeetingProvider(name string) bool
```

```go
// In omni-livekit/init.go
func init() {
    omnimeet.RegisterMeetingProvider("livekit", NewProvider, omnimeet.PriorityThick)
}
```

### Priority System

- `PriorityThin` (0): Minimal implementations (stdlib-only)
- `PriorityThick` (10): Full SDK implementations
- Higher priority overrides lower priority

## Consequences

### Positive

- **Consistency** — Developers familiar with OmniVoice or OmniLLM immediately understand OmniMeet
- **Modularity** — Users import only the providers they need
- **Extensibility** — Third parties can add providers without forking
- **Testability** — Mock providers can be registered for testing

### Negative

- **Boilerplate** — Each provider requires registration code
- **Complexity** — Factory pattern adds indirection

### Neutral

- Provider-specific configuration uses `Extensions map[string]any` pattern
- Multi-provider fallback support via `provider.Client[T]`

## Alternatives Considered

### Direct Instantiation (No Registry)

```go
provider := livekit.NewProvider(config)
```

Rejected because it prevents runtime provider selection and doesn't support the "batteries-included" bundle pattern.

### Interface-Only (No Factory)

Rejected because it doesn't support configuration-driven provider selection, which is important for OmniAgent integration.

## References

- [OmniVoice-Core Registry](https://github.com/plexusone/omnivoice-core/blob/main/registry.go)
- [OmniLLM Provider Pattern](https://github.com/plexusone/omnillm-core)
