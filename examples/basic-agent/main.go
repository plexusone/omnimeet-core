// Example: basic-agent demonstrates an AI agent joining a LiveKit meeting.
//
// This example shows how to:
// 1. Create a meeting using OmniMeet
// 2. Generate join tokens for human and agent participants
// 3. Have an agent join and listen to participant audio
// 4. Process audio with OmniVoice STT and respond with TTS
//
// Prerequisites:
// - LiveKit server running (local or cloud)
// - Set environment variables: LIVEKIT_API_KEY, LIVEKIT_API_SECRET, LIVEKIT_URL
//
// Usage:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	omnimeet "github.com/plexusone/omnimeet-core"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/voice"
	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/tts"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Get configuration from environment
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	serverURL := os.Getenv("LIVEKIT_URL")

	if apiKey == "" || apiSecret == "" || serverURL == "" {
		log.Fatal("Please set LIVEKIT_API_KEY, LIVEKIT_API_SECRET, and LIVEKIT_URL environment variables")
	}

	// Get meeting provider
	// Note: In a real application, you would import omni-livekit/omnimeet to register the provider
	// import _ "github.com/plexusone/omni-livekit/omnimeet"
	provider, err := omnimeet.GetMeetingProvider("livekit",
		omnimeet.WithAPIKey(apiKey),
		omnimeet.WithAPISecret(apiSecret),
		omnimeet.WithServerURL(serverURL),
	)
	if err != nil {
		log.Fatalf("Failed to get meeting provider: %v", err)
	}
	defer func() { _ = provider.Close() }()

	// Create a meeting
	meeting, err := provider.CreateMeeting(ctx, omnimeet.CreateMeetingRequest{
		Name: fmt.Sprintf("demo-meeting-%d", time.Now().Unix()),
		Metadata: map[string]string{
			"created_by": "basic-agent-example",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create meeting: %v", err)
	}
	//nolint:gosec // G706: values from internal API, not user input
	log.Printf("Created meeting: %s", meeting.ID)

	// Generate token for human participant
	humanToken, err := provider.CreateJoinToken(ctx, omnimeet.CreateJoinTokenRequest{
		MeetingID: meeting.ID,
		Participant: omnimeet.ParticipantInfo{
			Name:     "Human User",
			Kind:     omnimeet.ParticipantKindHuman,
			Identity: "human-user",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create human token: %v", err)
	}
	//nolint:gosec // G706: values from internal API, not user input
	log.Printf("Human join URL: %s", humanToken.JoinURL)

	// Generate token for AI agent
	agentToken, err := provider.CreateJoinToken(ctx, omnimeet.CreateJoinTokenRequest{
		MeetingID: meeting.ID,
		Participant: omnimeet.ParticipantInfo{
			Name:     "AI Assistant",
			Kind:     omnimeet.ParticipantKindAgent,
			Identity: "ai-assistant",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent token: %v", err)
	}

	// Check if provider supports agent participation
	factory, ok := provider.(omnimeet.AgentParticipantFactory)
	if !ok {
		log.Fatal("Provider does not support agent participation")
	}

	// Create agent participant
	agent, err := factory.CreateAgentParticipant(omnimeet.AgentParticipantOptions{
		AutoSubscribe: true,
	})
	if err != nil {
		log.Fatalf("Failed to create agent participant: %v", err)
	}

	// Join the meeting
	if err := agent.JoinMeeting(ctx, meeting.ID, agentToken); err != nil {
		log.Fatalf("Failed to join meeting: %v", err)
	}
	defer func() { _ = agent.LeaveMeeting(ctx) }()
	log.Println("Agent joined the meeting")

	// Set up event handlers
	agent.OnParticipantJoined(func(p participant.Participant) {
		log.Printf("Participant joined: %s (%s)", p.Name, p.Kind)

		// Start transcribing if it's a human
		if p.Kind == participant.KindHuman {
			log.Printf("Would start transcribing audio from: %s", p.Name)
		}
	})

	agent.OnParticipantLeft(func(p participant.Participant) {
		log.Printf("Participant left: %s", p.Name)
	})

	agent.OnActiveSpeakerChanged(func(speakers []participant.Participant) {
		if len(speakers) > 0 {
			log.Printf("Active speaker: %s", speakers[0].Name)
		}
	})

	// Create voice-enabled agent (with mock providers for demo)
	// In a real application, use omnivoice providers:
	//   sttProv, _ := omnivoice.GetSTTProvider("deepgram", omnivoice.WithAPIKey(deepgramKey))
	//   ttsProv, _ := omnivoice.GetTTSProvider("elevenlabs", omnivoice.WithAPIKey(elevenlabsKey))
	voiceAgent := voice.NewVoiceAgentParticipant(agent, voice.Config{
		STTProvider: &mockSTTProvider{},
		TTSProvider: &mockTTSProvider{},
		OnTranscript: func(participantID, participantName, text string, isFinal bool) {
			if isFinal {
				log.Printf("[%s]: %s", participantName, text)

				// In a real application, you would:
				// 1. Send text to OmniAgent for processing
				// 2. Get response
				// 3. Speak response using TTS
				//
				// response := omniagent.Process(ctx, sessionID, text)
				// voiceAgent.Speak(ctx, response)
			}
		},
	})

	// Wait for participants
	log.Println("Waiting for participants...")
	//nolint:gosec // G706: values from internal API, not user input
	log.Printf("Share this URL with participants: %s", humanToken.JoinURL)

	// Process events until shutdown
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, leaving meeting")
			voiceAgent.StopTranscribingAll()
			return
		case evt := <-agent.Events():
			log.Printf("Event: %s", evt.Type)
		}
	}
}

// mockSTTProvider is a placeholder STT provider for demonstration.
// In a real application, use omnivoice providers like Deepgram or OpenAI.
type mockSTTProvider struct{}

func (m *mockSTTProvider) Name() string { return "mock" }

func (m *mockSTTProvider) Transcribe(ctx context.Context, audio []byte, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	// In a real implementation, this would call an actual STT service
	return &stt.TranscriptionResult{
		Text: "[transcribed audio]",
	}, nil
}

func (m *mockSTTProvider) TranscribeFile(ctx context.Context, filePath string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	return &stt.TranscriptionResult{Text: "[file transcription]"}, nil
}

func (m *mockSTTProvider) TranscribeURL(ctx context.Context, url string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	return &stt.TranscriptionResult{Text: "[url transcription]"}, nil
}

// mockTTSProvider is a placeholder TTS provider for demonstration.
// In a real application, use omnivoice providers like ElevenLabs or OpenAI.
type mockTTSProvider struct{}

func (m *mockTTSProvider) Name() string { return "mock" }

func (m *mockTTSProvider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
	// In a real implementation, this would call an actual TTS service
	// and return actual audio bytes
	return &tts.SynthesisResult{
		Audio:      make([]byte, 0), // Empty for mock
		SampleRate: 48000,
	}, nil
}

func (m *mockTTSProvider) SynthesizeStream(ctx context.Context, text string, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	ch := make(chan tts.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *mockTTSProvider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	return []tts.Voice{{ID: "default", Name: "Default Voice"}}, nil
}

func (m *mockTTSProvider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
	return &tts.Voice{ID: voiceID, Name: "Mock Voice"}, nil
}
