# OmniAgent Integration

This guide covers integrating OmniMeet with OmniAgent to enable AI agents to manage meetings through conversation.

## Overview

The `omnimeet-core/skill` package provides a MeetingSkill that implements the `omniskill/skill.Skill` interface. This allows OmniAgent to:

- Create and manage meetings
- Generate join links for participants
- Join meetings as an AI agent
- Speak in meetings using TTS
- List participants and meeting status

## Installation

```go
import (
    "github.com/plexusone/omniagent/agent"
    "github.com/plexusone/omni-livekit/omnimeet"
    meetingskill "github.com/plexusone/omnimeet-core/skill"
)
```

## Quick Start

```go
// 1. Create meeting provider
provider, err := omnimeet.NewProvider(omnimeet.Config{
    APIKey:    os.Getenv("LIVEKIT_API_KEY"),
    APISecret: os.Getenv("LIVEKIT_API_SECRET"),
    ServerURL: os.Getenv("LIVEKIT_URL"),
})
if err != nil {
    log.Fatal(err)
}

// 2. Create meeting skill
skill := meetingskill.New(provider, meetingskill.Config{
    DefaultAgentName:   "AI Assistant",
    DefaultMeetingName: "AI Meeting",
    AutoJoinAsAgent:    true,
})

// 3. Register with OmniAgent
agent, err := agent.New(agentConfig,
    agent.WithCompiledSkill(skill),
)
if err != nil {
    log.Fatal(err)
}

// 4. Agent can now handle meeting requests!
```

## Configuration

```go
type Config struct {
    // DefaultMeetingName is used when creating meetings without a name.
    DefaultMeetingName string

    // DefaultAgentName is the agent's display name in meetings.
    DefaultAgentName string

    // AutoJoinAsAgent automatically joins meetings after creation.
    AutoJoinAsAgent bool
}
```

## Available Tools

The MeetingSkill provides 9 tools:

| Tool | Description |
|------|-------------|
| `create_meeting` | Create a new meeting room |
| `get_meeting` | Get meeting details by ID |
| `list_meetings` | List all active meetings |
| `end_meeting` | End a meeting |
| `join_meeting` | Join as AI agent |
| `leave_meeting` | Leave meeting |
| `get_join_link` | Generate join link for participant |
| `list_participants` | List meeting participants |
| `speak_in_meeting` | Speak text using TTS |

## Example Conversations

### Creating a Meeting

```
User: "Create a meeting for our team standup"

Agent: [calls create_meeting with name="Team Standup"]
       "I've created a meeting called 'Team Standup'.
        Join here: https://meet.example.com/abc123"
```

### Inviting Participants

```
User: "Generate a join link for Alice"

Agent: [calls get_join_link with participant_name="Alice"]
       "Here's the join link for Alice:
        https://meet.example.com/abc123?token=xyz"
```

### Agent Participation

```
User: "Join the meeting and introduce yourself"

Agent: [calls join_meeting]
       [calls speak_in_meeting with text="Hello everyone, I'm your AI assistant..."]
       "I've joined the meeting and introduced myself."
```

### Checking Status

```
User: "Who's in the meeting?"

Agent: [calls list_participants]
       "There are 3 participants in the meeting:
        - Alice (Human)
        - Bob (Human)
        - AI Assistant (Agent)"
```

## Event Handling

Listen for meeting events:

```go
skill.OnEvent(func(e meetingskill.Event) {
    switch e.Type {
    case "meeting_created":
        log.Printf("Meeting created: %s", e.MeetingID)
    case "participant_joined":
        log.Printf("Participant joined: %v", e.Data)
    case "agent_joined":
        log.Printf("Agent joined meeting: %s", e.MeetingID)
    }
})
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

## Provider Flexibility

The skill works with any OmniMeet provider:

```go
// LiveKit
livekitProvider, _ := livekit.NewProvider(livekitConfig)
skill := meetingskill.New(livekitProvider, config)

// Daily (when available)
dailyProvider, _ := daily.NewProvider(dailyConfig)
skill := meetingskill.New(dailyProvider, config)

// Same skill interface, different providers
```

## Accessing Sessions

Get active meeting sessions:

```go
// Get specific session
session := skill.GetSession(meetingID)
if session != nil {
    fmt.Printf("Joined at: %s\n", session.JoinedAt)
    fmt.Printf("Participants: %d\n", len(session.Participants))
}

// Access underlying provider
provider := skill.Provider()
```

## Voice Integration

For voice capabilities, use a provider that supports agent participation:

```go
// The skill automatically uses voice if the agent participant supports it
session := skill.GetSession(meetingID)
if session != nil && session.Agent != nil {
    // If using VoiceAgentParticipant from omni-livekit
    if vap, ok := session.Agent.(interface{ Speak(context.Context, string) error }); ok {
        vap.Speak(ctx, "Hello from the agent!")
    }
}
```

## Best Practices

1. **Set meaningful defaults** - Configure `DefaultAgentName` appropriately
2. **Handle events** - Listen for participant events for context awareness
3. **Clean up** - Call `skill.Close()` on shutdown to leave all meetings
4. **Error handling** - Tools return errors that agents can report to users
5. **Provider choice** - Use LiveKit for self-hosted, Daily for quick setup

## Next Steps

- [Agent Participation](agent-participation.md) - Deep dive into agent features
- [Voice Integration](voice-integration.md) - Add STT/TTS capabilities
- [LiveKit Provider](../providers/livekit.md) - LiveKit-specific features
