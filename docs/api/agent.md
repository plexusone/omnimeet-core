# Agent Types

Package `agent` provides OmniAgent skill types for meeting management.

## MeetingSkill

Provides meeting management tools for OmniAgent.

```go
type MeetingSkill struct {
    // ... internal fields
}

func NewMeetingSkill(prov provider.MeetingProvider, cfg SkillConfig) (*MeetingSkill, error)
```

### SkillConfig

```go
type SkillConfig struct {
    // DefaultMeetingName is used when no name is specified.
    DefaultMeetingName string

    // DefaultAgentName is the agent's display name.
    DefaultAgentName string

    // AutoJoinAsAgent automatically joins meetings after creation.
    AutoJoinAsAgent bool

    // TranscriptionEnabled enables automatic transcription.
    TranscriptionEnabled bool
}
```

### Methods

```go
// Name returns the skill name.
func (s *MeetingSkill) Name() string

// Description returns the skill description.
func (s *MeetingSkill) Description() string

// Tools returns available tools.
func (s *MeetingSkill) Tools() []Tool

// Init initializes the skill.
func (s *MeetingSkill) Init(ctx context.Context) error

// Close cleans up resources.
func (s *MeetingSkill) Close() error

// OnMeetingEvent registers an event handler.
func (s *MeetingSkill) OnMeetingEvent(handler func(MeetingEvent))
```

### Available Tools

| Tool | Description |
|------|-------------|
| `create_meeting` | Create a new meeting |
| `get_meeting` | Get meeting information |
| `list_meetings` | List active meetings |
| `end_meeting` | End a meeting |
| `join_meeting` | Join a meeting as agent |
| `leave_meeting` | Leave a meeting |
| `get_join_link` | Generate a join URL |
| `list_participants` | List meeting participants |
| `speak_in_meeting` | Speak using TTS |
| `get_meeting_transcript` | Get meeting transcript |

## VoiceMeetingSkill

Extends MeetingSkill with voice capabilities.

```go
type VoiceMeetingSkill struct {
    *MeetingSkill
    // ... voice config
}

func NewVoiceMeetingSkill(prov provider.MeetingProvider, cfg SkillConfig, voiceCfg VoiceSkillConfig) (*VoiceMeetingSkill, error)
```

### VoiceSkillConfig

```go
type VoiceSkillConfig struct {
    // STTProvider is the speech-to-text provider.
    STTProvider voice.STTProvider

    // TTSProvider is the text-to-speech provider.
    TTSProvider voice.TTSProvider

    // STTConfig configures speech-to-text.
    STTConfig voice.STTConfig

    // TTSConfig configures text-to-speech.
    TTSConfig voice.TTSConfig

    // OnTranscript is called when speech is transcribed.
    OnTranscript func(meetingID, participantID, participantName, text string, isFinal bool)
}
```

## Tool

Interface for skill tools.

```go
type Tool interface {
    // Name returns the tool name.
    Name() string

    // Description returns a description for the LLM.
    Description() string

    // Parameters returns the JSON Schema for parameters.
    Parameters() map[string]any

    // Execute runs the tool with JSON arguments.
    Execute(ctx context.Context, args json.RawMessage) (string, error)
}
```

## MeetingSession

Represents an active agent session in a meeting.

```go
type MeetingSession struct {
    // Meeting is the meeting info.
    Meeting *meeting.Meeting

    // Agent is the agent participant.
    Agent provider.AgentParticipant

    // JoinedAt is when the agent joined.
    JoinedAt time.Time

    // Participants lists current participants.
    Participants []participant.Participant

    // Transcript contains transcribed speech.
    Transcript []TranscriptEntry
}
```

## TranscriptEntry

```go
type TranscriptEntry struct {
    ParticipantID   string
    ParticipantName string
    Text            string
    Timestamp       time.Time
    IsFinal         bool
}
```

## MeetingEvent

Events emitted by the skill.

```go
type MeetingEvent struct {
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
| `meeting_ended` | `string` (meeting ID) | Meeting ended |
| `agent_joined` | - | Agent joined meeting |
| `agent_left` | - | Agent left meeting |
| `participant_joined` | `participant.Participant` | Participant joined |
| `participant_left` | `participant.Participant` | Participant left |
| `transcript_updated` | `TranscriptEntry` | New transcription |

## Example Usage

```go
import "github.com/plexusone/omnimeet-core/agent"

// Create skill
skill, _ := agent.NewMeetingSkill(provider, agent.SkillConfig{
    DefaultMeetingName: "AI Meeting",
    DefaultAgentName:   "Assistant",
    AutoJoinAsAgent:    true,
})
defer skill.Close()

// Initialize
skill.Init(ctx)

// Handle events
skill.OnMeetingEvent(func(e agent.MeetingEvent) {
    log.Printf("Event: %s in %s", e.Type, e.MeetingID)
})

// Use tools
createTool := findTool(skill.Tools(), "create_meeting")
result, _ := createTool.Execute(ctx, []byte(`{"name": "Team Meeting"}`))
log.Printf("Created: %s", result)
```
