# Quick Start

This guide walks you through creating a meeting, generating join tokens, and having an AI agent join.

## Prerequisites

- OmniMeet installed ([Installation](installation.md))
- LiveKit credentials configured
- Environment variables set

## Step 1: Create a Provider

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/plexusone/omni-livekit/omnimeet"
)

func main() {
    ctx := context.Background()

    // Create provider from environment
    provider, err := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    os.Getenv("LIVEKIT_API_KEY"),
        APISecret: os.Getenv("LIVEKIT_API_SECRET"),
        ServerURL: os.Getenv("LIVEKIT_URL"),
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Provider: %s", provider.Name())
}
```

## Step 2: Create a Meeting

```go
import "github.com/plexusone/omnimeet-core/meeting"

// Create a new meeting
m, err := provider.CreateMeeting(ctx, meeting.CreateRequest{
    Name: "Team Standup",
    Metadata: map[string]string{
        "team": "engineering",
    },
})
if err != nil {
    log.Fatal(err)
}

log.Printf("Meeting ID: %s", m.ID)
log.Printf("Meeting Name: %s", m.Name)
```

## Step 3: Generate Join Tokens

```go
import (
    "github.com/plexusone/omnimeet-core/token"
    "github.com/plexusone/omnimeet-core/participant"
)

// Generate token for a human participant
humanToken, err := provider.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: m.ID,
    Participant: participant.Info{
        Name:     "Alice",
        Kind:     participant.KindHuman,
        Identity: "alice@example.com",
    },
})
if err != nil {
    log.Fatal(err)
}

log.Printf("Human join URL: %s", humanToken.JoinURL)

// Generate token for an AI agent
agentToken, err := provider.CreateJoinToken(ctx, token.CreateRequest{
    MeetingID: m.ID,
    Participant: participant.Info{
        Name:     "AI Assistant",
        Kind:     participant.KindAgent,
        Identity: "ai-assistant",
    },
})
if err != nil {
    log.Fatal(err)
}

log.Printf("Agent token: %s", agentToken.Token)
```

## Step 4: Join as an Agent

```go
import "github.com/plexusone/omnimeet-core/provider"

// Get agent factory
factory := provider.(provider.AgentParticipantFactory)

// Create agent participant
agent, err := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,
})
if err != nil {
    log.Fatal(err)
}

// Set up event handlers
agent.OnParticipantJoined(func(p participant.Participant) {
    log.Printf("Participant joined: %s", p.Name)
})

agent.OnParticipantLeft(func(p participant.Participant) {
    log.Printf("Participant left: %s", p.Name)
})

// Join the meeting
err = agent.JoinMeeting(ctx, m.ID, agentToken)
if err != nil {
    log.Fatal(err)
}

log.Println("Agent joined the meeting")
```

## Step 5: Listen to Audio

```go
// Subscribe to all audio
audioCh, err := agent.SubscribeToAllAudio(ctx)
if err != nil {
    log.Fatal(err)
}

go func() {
    for frame := range audioCh {
        // Process audio frame
        // In a real app, send to STT provider
        log.Printf("Received audio from %s: %d bytes",
            frame.ParticipantID, len(frame.Data))
    }
}()
```

## Step 6: Add Voice (Optional)

For a complete voice agent with STT/TTS, use [OmniVoice](https://github.com/plexusone/omnivoice):

```go
import (
    "github.com/plexusone/omnimeet-core/voice"
    "github.com/plexusone/omnivoice"
    "github.com/plexusone/omnivoice-core/stt"
    "github.com/plexusone/omnivoice-core/tts"
    _ "github.com/plexusone/omnivoice/providers/all"
)

// Get STT/TTS providers
sttProv, _ := omnivoice.GetSTTProvider("deepgram",
    omnivoice.WithAPIKey(os.Getenv("DEEPGRAM_API_KEY")),
)
ttsProv, _ := omnivoice.GetTTSProvider("openai",
    omnivoice.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
)

// Wrap agent with voice
voiceAgent := voice.NewVoiceAgentParticipant(agent, voice.Config{
    STTProvider: sttProv,
    TTSProvider: ttsProv,
    STTConfig: stt.TranscriptionConfig{
        Language:   "en",
        SampleRate: 16000,
    },
    TTSConfig: tts.SynthesisConfig{
        VoiceID:    "alloy",
        SampleRate: 48000,
    },
    OnTranscript: func(participantID, name, text string, isFinal bool) {
        if isFinal {
            log.Printf("%s: %s", name, text)
            // Process with LLM and respond
            voiceAgent.Speak(ctx, "I heard you say: "+text)
        }
    },
})

// Start transcribing all participants
voiceAgent.StartTranscribingAll(ctx)
```

See [Voice Integration](../guides/voice-integration.md) for the full guide.

## Complete Example

Here's a complete example putting it all together:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/plexusone/omni-livekit/omnimeet"
    "github.com/plexusone/omnimeet-core/meeting"
    "github.com/plexusone/omnimeet-core/participant"
    "github.com/plexusone/omnimeet-core/provider"
    "github.com/plexusone/omnimeet-core/token"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    // Create provider
    prov, _ := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    os.Getenv("LIVEKIT_API_KEY"),
        APISecret: os.Getenv("LIVEKIT_API_SECRET"),
        ServerURL: os.Getenv("LIVEKIT_URL"),
    })

    // Create meeting
    m, _ := prov.CreateMeeting(ctx, meeting.CreateRequest{
        Name: "Quick Start Demo",
    })
    log.Printf("Meeting: %s", m.ID)

    // Generate agent token
    tok, _ := prov.CreateJoinToken(ctx, token.CreateRequest{
        MeetingID: m.ID,
        Participant: participant.Info{
            Name: "AI Assistant",
            Kind: participant.KindAgent,
        },
    })

    // Create and join as agent
    factory := prov.(provider.AgentParticipantFactory)
    agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
        AutoSubscribe: true,
    })

    agent.OnParticipantJoined(func(p participant.Participant) {
        log.Printf("+ %s joined", p.Name)
    })

    agent.OnParticipantLeft(func(p participant.Participant) {
        log.Printf("- %s left", p.Name)
    })

    agent.JoinMeeting(ctx, m.ID, tok)
    log.Println("Agent in meeting. Press Ctrl+C to exit.")

    <-ctx.Done()
    agent.LeaveMeeting(ctx)
    log.Println("Done")
}
```

## Next Steps

- [Agent Participation](../guides/agent-participation.md) - Deep dive into agent features
- [Voice Integration](../guides/voice-integration.md) - Add STT/TTS capabilities
- [Testing](../guides/testing.md) - Write integration tests
