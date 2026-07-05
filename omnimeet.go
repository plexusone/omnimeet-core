// Package omnimeet provides a unified abstraction for real-time collaboration platforms.
//
// OmniMeet enables AI agents to participate in meetings alongside humans,
// supporting providers like LiveKit, Daily, Zoom, Google Meet, and more.
//
// # Core Concepts
//
//   - Meeting: A live collaborative session containing participants and media streams
//   - Participant: An entity in a meeting (human, AI agent, recorder, observer)
//   - Track: A media stream (audio, video, screen share, data)
//   - Event: Real-time events (join, leave, track published, etc.)
//
// # Architecture
//
// OmniMeet follows the PlexusOne provider pattern:
//
//   - omnimeet-core: Core interfaces and types (this package)
//   - omni-livekit: LiveKit provider implementation
//   - omni-daily: Daily provider implementation
//   - omnimeet: Bundle package with common providers
//
// # Basic Usage
//
//	import (
//	    "github.com/plexusone/omnimeet"
//	    _ "github.com/plexusone/omni-livekit" // Register provider
//	)
//
//	func main() {
//	    provider, _ := omnimeet.GetMeetingProvider("livekit",
//	        omnimeet.WithAPIKey("..."),
//	        omnimeet.WithAPISecret("..."),
//	        omnimeet.WithServerURL("wss://your-livekit-server.com"),
//	    )
//
//	    meeting, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{
//	        Name: "Team Standup",
//	    })
//
//	    token, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
//	        MeetingID: meeting.ID,
//	        Participant: participant.Info{
//	            Name: "AI Assistant",
//	            Kind: participant.KindAgent,
//	        },
//	    })
//
//	    fmt.Printf("Join URL: %s\n", token.JoinURL)
//	}
//
// # Agent Participation
//
// For AI agents to actively participate in meetings (receiving/publishing audio),
// use the AgentParticipant interface:
//
//	factory := provider.(provider.AgentParticipantFactory)
//	agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{})
//
//	agent.JoinMeeting(ctx, meeting.ID, token)
//	defer agent.LeaveMeeting(ctx)
//
//	audioCh, _ := agent.SubscribeToAllAudio(ctx)
//	for frame := range audioCh {
//	    // Process audio with OmniVoice STT
//	    // Generate response with OmniAgent
//	    // Publish response with agent.PublishAudio()
//	}
//
// # Integration with PlexusOne
//
// OmniMeet integrates with other PlexusOne libraries:
//
//   - OmniVoice: STT/TTS for speech processing
//   - OmniAgent: AI agent orchestration
//   - OmniChat: Meeting invites via messaging
//   - OmniMemory: Storing meeting transcripts
//
// See https://github.com/plexusone/omnimeet-core for documentation.
package omnimeet

// Version is the current version of omnimeet-core.
const Version = "0.1.0"
