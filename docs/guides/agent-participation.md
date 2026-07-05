# Agent Participation

This guide covers how AI agents join and participate in meetings using OmniMeet.

## Overview

OmniMeet is designed with AI agents as first-class participants. Agents can:

- Join meetings alongside humans
- Listen to audio from participants
- Speak using synthesized voice
- Send and receive data messages
- React to meeting events

## Creating an Agent Participant

### Basic Setup

```go
import (
    "github.com/plexusone/omni-livekit/omnimeet"
    "github.com/plexusone/omnimeet-core/provider"
)

// Create provider
prov, _ := omnimeet.NewProvider(omnimeet.Config{...})

// Cast to factory
factory := prov.(provider.AgentParticipantFactory)

// Create agent with options
agent, err := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,  // Automatically subscribe to new tracks
})
```

### Agent Options

| Option | Description |
|--------|-------------|
| `AutoSubscribe` | Automatically subscribe to audio/video tracks |

## Joining a Meeting

```go
import (
    "github.com/plexusone/omnimeet-core/token"
    "github.com/plexusone/omnimeet-core/participant"
)

// Generate token for agent
tok, err := prov.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: meetingID,
    Participant: participant.Info{
        Name:     "AI Assistant",
        Kind:     participant.KindAgent,
        Identity: "agent-001",
        Metadata: map[string]string{
            "type": "assistant",
            "language": "en",
        },
    },
})

// Join meeting
err = agent.JoinMeeting(ctx, meetingID, tok)
```

## Event Handling

### Participant Events

```go
// Participant joined
agent.OnParticipantJoined(func(p participant.Participant) {
    log.Printf("Participant joined: %s (kind: %s)", p.Name, p.Kind)

    // Welcome new humans
    if p.Kind == participant.KindHuman {
        go speakWelcome(ctx, agent, p.Name)
    }
})

// Participant left
agent.OnParticipantLeft(func(p participant.Participant) {
    log.Printf("Participant left: %s", p.Name)
})

// Active speaker changed
agent.OnActiveSpeakerChanged(func(speakers []participant.Participant) {
    if len(speakers) > 0 {
        log.Printf("Active speaker: %s", speakers[0].Name)
    }
})
```

### Track Events

```go
import "github.com/plexusone/omnimeet-core/track"

// Track published
agent.OnTrackPublished(func(p participant.Participant, t track.Track) {
    log.Printf("%s published %s track", p.Name, t.Kind)
})

// Track unpublished
agent.OnTrackUnpublished(func(p participant.Participant, t track.Track) {
    log.Printf("%s unpublished %s track", p.Name, t.Kind)
})
```

### Data Messages

```go
agent.OnDataMessage(func(msg provider.DataMessage) {
    log.Printf("Data from %s: %s", msg.ParticipantID, string(msg.Data))

    // Handle commands
    if string(msg.Data) == "mute" {
        // Handle mute command
    }
})
```

## Audio Handling

### Receiving Audio

```go
// Subscribe to specific participant
audioCh, err := agent.SubscribeToAudio(ctx, participantID)

// Or subscribe to all participants
audioCh, err := agent.SubscribeToAllAudio(ctx)

// Process audio frames
for frame := range audioCh {
    log.Printf("Audio from %s: %d bytes, %d Hz",
        frame.ParticipantID,
        len(frame.Data),
        frame.SampleRate,
    )

    // Send to STT provider
    transcript, _ := sttProvider.Transcribe(ctx, frame.Data, sttConfig)
}
```

### Audio Frame Format

```go
type AudioFrame struct {
    Data          []byte    // PCM audio data
    ParticipantID string    // Source participant
    SampleRate    int       // Sample rate (e.g., 48000)
    Channels      int       // Number of channels (1 or 2)
    Timestamp     time.Time // Capture timestamp
}
```

### Sending Audio

```go
// Option 1: Simple frame-by-frame
agent.PublishAudio(ctx, provider.AudioFrame{
    Data:       audioData,
    SampleRate: 48000,
    Channels:   1,
})

// Option 2: Audio track writer (recommended for TTS)
writer, err := agent.StartAudioTrack(ctx, provider.AudioTrackOptions{
    SampleRate: 48000,
    Channels:   1,
})

// Write TTS output
for _, chunk := range ttsChunks {
    writer.Write(provider.AudioFrame{
        Data:       chunk,
        SampleRate: 48000,
        Channels:   1,
    })
}

// Stop when done
agent.StopAudioTrack(ctx)
```

## Data Messages

### Sending Messages

```go
// Send to all participants
agent.SendDataMessage(ctx, provider.DataMessage{
    Topic: "chat",
    Data:  []byte(`{"type": "typing", "agent": "assistant"}`),
})

// Send to specific participant
agent.SendDataMessage(ctx, provider.DataMessage{
    Topic:            "chat",
    Data:             []byte("Hello!"),
    DestinationIDs:   []string{participantID},
})
```

### Message Topics

Common topics for agent communication:

| Topic | Use Case |
|-------|----------|
| `chat` | Text chat messages |
| `transcription` | Real-time transcription |
| `agent-status` | Agent state updates |
| `command` | Control commands |

## Using MeetingSkill

For OmniAgent integration, use the pre-built MeetingSkill:

```go
import "github.com/plexusone/omnimeet-core/agent"

// Create skill
skill, err := agent.NewMeetingSkill(prov, agent.SkillConfig{
    DefaultMeetingName: "AI Meeting",
    DefaultAgentName:   "Assistant",
    AutoJoinAsAgent:    true,
})

// Get available tools
for _, tool := range skill.Tools() {
    log.Printf("Tool: %s - %s", tool.Name(), tool.Description())
}

// Tools available:
// - create_meeting
// - get_meeting
// - list_meetings
// - end_meeting
// - join_meeting
// - leave_meeting
// - get_join_link
// - list_participants
// - speak_in_meeting
// - get_meeting_transcript
```

### With Voice Integration

```go
import "github.com/plexusone/omnimeet-core/voice"

// Create voice-enabled skill
voiceSkill, err := agent.NewVoiceMeetingSkill(prov, skillConfig, agent.VoiceSkillConfig{
    STTProvider: deepgramProvider,
    TTSProvider: elevenLabsProvider,
    STTConfig: voice.STTConfig{
        Language:   "en",
        SampleRate: 16000,
    },
    TTSConfig: voice.TTSConfig{
        Voice:      "alloy",
        SampleRate: 48000,
    },
    OnTranscript: func(meetingID, participantID, name, text string, isFinal bool) {
        if isFinal {
            log.Printf("[%s] %s: %s", meetingID, name, text)
            // Process with LLM and respond
        }
    },
})
```

## Querying Meeting State

```go
// Get current meeting
meeting := agent.Meeting()
log.Printf("In meeting: %s", meeting.Name)

// Get local participant info
local := agent.LocalParticipant()
log.Printf("I am: %s", local.Name)

// Get remote participants
for _, p := range agent.RemoteParticipants() {
    log.Printf("Participant: %s (%s)", p.Name, p.Kind)
}

// Get specific participant
p := agent.GetParticipant(participantID)

// Check connection state
state := agent.ConnectionState()
// States: Disconnected, Connecting, Connected, Reconnecting
```

## Graceful Shutdown

```go
// Handle signals
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigCh
    log.Println("Shutting down...")

    // Leave meeting gracefully
    agent.LeaveMeeting(ctx)

    cancel()
}()
```

## Best Practices

1. **Handle reconnection** - Agents may disconnect; implement retry logic
2. **Buffer audio** - Use channels with buffers for audio processing
3. **Respect rate limits** - Don't publish audio too frequently
4. **Clean up resources** - Always call `LeaveMeeting` on shutdown
5. **Log events** - Track participant activity for debugging
6. **Use timeouts** - Context with timeout for all operations

## Next Steps

- [Voice Integration](voice-integration.md) - Add STT/TTS
- [Testing](testing.md) - Write integration tests
