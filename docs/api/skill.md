# Skill Types

Package `skill` provides an omniskill-compatible meeting skill for OmniAgent integration.

## MeetingSkill

The main skill type implementing `omniskill/skill.Skill`.

```go
type MeetingSkill struct {
    // ... internal fields
}

func New(prov provider.MeetingProvider, cfg Config) *MeetingSkill
```

### Config

```go
type Config struct {
    // DefaultMeetingName is used when creating meetings without a name.
    DefaultMeetingName string

    // DefaultAgentName is the default name for the agent participant.
    DefaultAgentName string

    // AutoJoinAsAgent controls whether the agent automatically joins created meetings.
    AutoJoinAsAgent bool
}
```

### Methods

```go
// Name returns the skill name ("meeting").
func (s *MeetingSkill) Name() string

// Description returns the skill description.
func (s *MeetingSkill) Description() string

// Tools returns the skill's tools (9 tools).
func (s *MeetingSkill) Tools() []skill.Tool

// Init initializes the skill.
func (s *MeetingSkill) Init(ctx context.Context) error

// Close releases resources and leaves all meetings.
func (s *MeetingSkill) Close() error

// OnEvent sets the event handler.
func (s *MeetingSkill) OnEvent(handler func(Event))

// GetSession returns an active meeting session.
func (s *MeetingSkill) GetSession(meetingID string) *Session

// Provider returns the underlying meeting provider.
func (s *MeetingSkill) Provider() provider.MeetingProvider
```

### Available Tools

| Tool | Parameters | Description |
|------|------------|-------------|
| `create_meeting` | `name`, `max_participants` | Create a new meeting |
| `get_meeting` | `meeting_id` | Get meeting details |
| `list_meetings` | - | List active meetings |
| `end_meeting` | `meeting_id` | End a meeting |
| `join_meeting` | `meeting_id` | Join as AI agent |
| `leave_meeting` | `meeting_id` | Leave meeting |
| `get_join_link` | `meeting_id`, `participant_name` | Generate join link |
| `list_participants` | `meeting_id` | List participants |
| `speak_in_meeting` | `meeting_id`, `text` | Speak using TTS |

## Session

Represents an active meeting session.

```go
type Session struct {
    // Meeting is the meeting info.
    Meeting *meeting.Meeting

    // Agent is the agent participant (if joined).
    Agent provider.AgentParticipant

    // JoinedAt is when the agent joined.
    JoinedAt time.Time

    // Participants lists current participants.
    Participants []participant.Participant
}
```

## Event

Events emitted by the skill.

```go
type Event struct {
    // Type is the event type.
    Type string

    // MeetingID is the affected meeting.
    MeetingID string

    // Timestamp is when the event occurred.
    Timestamp time.Time

    // Data contains event-specific data.
    Data any
}
```

Event types:

| Type | Data | Description |
|------|------|-------------|
| `meeting_created` | `*meeting.Meeting` | Meeting was created |
| `meeting_ended` | - | Meeting ended |
| `agent_joined` | - | Agent joined meeting |
| `agent_left` | - | Agent left meeting |
| `participant_joined` | `participant.Participant` | Participant joined |
| `participant_left` | `participant.Participant` | Participant left |

## Example Usage

```go
import (
    "github.com/plexusone/omniagent/agent"
    "github.com/plexusone/omni-livekit/omnimeet"
    meetingskill "github.com/plexusone/omnimeet-core/skill"
)

// Create provider
provider, _ := omnimeet.NewProvider(omnimeet.Config{
    APIKey:    os.Getenv("LIVEKIT_API_KEY"),
    APISecret: os.Getenv("LIVEKIT_API_SECRET"),
    ServerURL: os.Getenv("LIVEKIT_URL"),
})

// Create skill
skill := meetingskill.New(provider, meetingskill.Config{
    DefaultAgentName:   "AI Assistant",
    DefaultMeetingName: "AI Meeting",
    AutoJoinAsAgent:    true,
})

// Handle events
skill.OnEvent(func(e meetingskill.Event) {
    log.Printf("Event: %s in %s", e.Type, e.MeetingID)
})

// Register with OmniAgent
agent, _ := agent.New(config,
    agent.WithCompiledSkill(skill),
)

// Skill is now available for the agent to use
// Agent can call tools like create_meeting, join_meeting, etc.
```

## Interface Compliance

The MeetingSkill implements `skill.Skill` from omniskill:

```go
var _ skill.Skill = (*MeetingSkill)(nil)
```

This allows it to be used with:

- OmniAgent's `WithCompiledSkill()` option
- Any system that supports omniskill-compatible skills
