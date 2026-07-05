# Implementing a Provider

This guide explains how to implement a new OmniMeet provider for a platform not yet supported.

## Overview

A provider implements the OmniMeet interfaces for a specific platform. The main interfaces are:

1. **MeetingProvider** - Meeting lifecycle management
2. **AgentParticipantFactory** - Creates agent participants
3. **AgentParticipant** - Agent in a meeting (media plane)

## Project Structure

Create a new Go module for your provider:

```
omni-yourplatform/
├── go.mod
├── omnimeet/
│   ├── provider.go      # MeetingProvider implementation
│   ├── agent.go         # AgentParticipant implementation
│   ├── audio.go         # Audio handling (optional)
│   └── provider_test.go # Tests
└── README.md
```

## Step 1: Implement MeetingProvider

The `MeetingProvider` interface handles meeting lifecycle:

```go
package omnimeet

import (
    "context"

    "github.com/plexusone/omnimeet-core/meeting"
    "github.com/plexusone/omnimeet-core/participant"
    "github.com/plexusone/omnimeet-core/provider"
    "github.com/plexusone/omnimeet-core/token"
)

type Config struct {
    APIKey    string
    APISecret string
    ServerURL string
}

type Provider struct {
    config Config
    // Platform-specific client
}

func NewProvider(cfg Config) (*Provider, error) {
    // Initialize platform client
    return &Provider{config: cfg}, nil
}

func (p *Provider) Name() string {
    return "yourplatform"
}

// CreateMeeting creates a new meeting on the platform
func (p *Provider) CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error) {
    // Call platform API to create meeting/room
    // Map response to meeting.Meeting
}

// GetMeeting retrieves meeting info
func (p *Provider) GetMeeting(ctx context.Context, id string) (*meeting.Meeting, error) {
    // Call platform API
}

// ListMeetings returns active meetings
func (p *Provider) ListMeetings(ctx context.Context) ([]*meeting.Meeting, error) {
    // Call platform API
}

// DeleteMeeting ends and removes a meeting
func (p *Provider) DeleteMeeting(ctx context.Context, id string) error {
    // Call platform API
}

// CreateJoinToken generates a token for joining
func (p *Provider) CreateJoinToken(ctx context.Context, req token.CreateRequest) (*token.JoinToken, error) {
    // Generate platform-specific token
    // Include participant metadata
}

// GetParticipants returns current participants
func (p *Provider) GetParticipants(ctx context.Context, meetingID string) ([]participant.Participant, error) {
    // Call platform API
}

// RemoveParticipant kicks a participant
func (p *Provider) RemoveParticipant(ctx context.Context, meetingID, participantID string) error {
    // Call platform API
}
```

## Step 2: Implement AgentParticipantFactory

The factory creates agent participants:

```go
// Ensure Provider implements the factory interface
var _ provider.AgentParticipantFactory = (*Provider)(nil)

func (p *Provider) CreateAgentParticipant(opts provider.AgentParticipantOptions) (provider.AgentParticipant, error) {
    return &AgentParticipant{
        provider:      p,
        autoSubscribe: opts.AutoSubscribe,
        eventHandlers: make(map[string][]func(any)),
    }, nil
}
```

## Step 3: Implement AgentParticipant

The `AgentParticipant` is the media-plane interface for agents:

```go
type AgentParticipant struct {
    provider      *Provider
    autoSubscribe bool
    meetingID     string
    localInfo     *participant.Participant

    // Event handlers
    onJoined  []func(participant.Participant)
    onLeft    []func(participant.Participant)

    // Platform-specific connection
    conn      *platformConnection
}

// JoinMeeting connects to the meeting
func (a *AgentParticipant) JoinMeeting(ctx context.Context, meetingID string, tok *token.JoinToken) error {
    a.meetingID = meetingID

    // Connect using platform SDK
    // a.conn = platform.Connect(tok.Token, ...)

    // Set up event listeners
    // a.conn.OnParticipantJoined(...)

    return nil
}

// LeaveMeeting disconnects from the meeting
func (a *AgentParticipant) LeaveMeeting(ctx context.Context) error {
    // Disconnect from platform
    return nil
}

// SubscribeToAudio subscribes to a participant's audio
func (a *AgentParticipant) SubscribeToAudio(ctx context.Context, participantID string) (<-chan provider.AudioFrame, error) {
    ch := make(chan provider.AudioFrame, 100)

    // Subscribe to participant's audio track
    // Convert platform audio to provider.AudioFrame

    return ch, nil
}

// SubscribeToAllAudio subscribes to all participants' audio
func (a *AgentParticipant) SubscribeToAllAudio(ctx context.Context) (<-chan provider.AudioFrame, error) {
    ch := make(chan provider.AudioFrame, 100)

    // Subscribe to all audio tracks

    return ch, nil
}

// PublishAudio sends an audio frame
func (a *AgentParticipant) PublishAudio(ctx context.Context, frame provider.AudioFrame) error {
    // Convert to platform format and send
    return nil
}

// StartAudioTrack begins publishing audio
func (a *AgentParticipant) StartAudioTrack(ctx context.Context, opts provider.AudioTrackOptions) (provider.AudioWriter, error) {
    // Create and publish audio track
    return &audioWriter{}, nil
}

// StopAudioTrack stops publishing audio
func (a *AgentParticipant) StopAudioTrack(ctx context.Context) error {
    return nil
}

// SendDataMessage sends a data message
func (a *AgentParticipant) SendDataMessage(ctx context.Context, msg provider.DataMessage) error {
    // Send via platform data channel
    return nil
}

// Event handler setters
func (a *AgentParticipant) OnParticipantJoined(handler func(participant.Participant)) {
    a.onJoined = append(a.onJoined, handler)
}

func (a *AgentParticipant) OnParticipantLeft(handler func(participant.Participant)) {
    a.onLeft = append(a.onLeft, handler)
}

// ... other event handlers
```

## Step 4: Register Provider (Optional)

If using the registry pattern:

```go
func init() {
    omnimeet.RegisterProvider("yourplatform", func(cfg map[string]any) (provider.MeetingProvider, error) {
        return NewProvider(Config{
            APIKey:    cfg["api_key"].(string),
            APISecret: cfg["api_secret"].(string),
            ServerURL: cfg["server_url"].(string),
        })
    })
}
```

## Audio Handling

Audio is the most complex part. Key considerations:

### Audio Frame Format

```go
type AudioFrame struct {
    Data          []byte    // Raw PCM or encoded audio
    ParticipantID string    // Who sent this
    SampleRate    int       // e.g., 48000
    Channels      int       // 1 (mono) or 2 (stereo)
    Timestamp     time.Time // When captured
}
```

### Opus Encoding/Decoding

Most WebRTC platforms use Opus. Consider:

1. **CGO with libopus** - Best quality, requires CGO
2. **Pure Go decoder** - No CGO, limited features
3. **Raw passthrough** - Let downstream handle decoding

See `omni-livekit/omnimeet/audio.go` for an example using CGO.

## Testing

Write integration tests:

```go
//go:build integration

package omnimeet_test

import (
    "context"
    "os"
    "testing"

    "github.com/plexusone/omni-yourplatform/omnimeet"
    "github.com/plexusone/omnimeet-core/meeting"
)

func TestCreateMeeting(t *testing.T) {
    provider, err := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    os.Getenv("YOURPLATFORM_API_KEY"),
        // ...
    })
    if err != nil {
        t.Fatal(err)
    }

    m, err := provider.CreateMeeting(context.Background(), meeting.CreateRequest{
        Name: "Test Meeting",
    })
    if err != nil {
        t.Fatal(err)
    }

    if m.ID == "" {
        t.Error("expected meeting ID")
    }

    // Cleanup
    provider.DeleteMeeting(context.Background(), m.ID)
}
```

## Platform-Specific Considerations

### LiveKit

- WebRTC-based, high performance
- Use `livekit-server-sdk-go` for server API
- Use `livekit/protocol` for types
- Audio via Opus

### Daily

- REST API for room management
- WebRTC for media
- Use `daily-go` client library
- Simpler API than LiveKit

### Zoom

- OAuth2 authentication
- Zoom SDK for native client
- Complex participant model
- Consider using Zoom Apps

## Next Steps

- Review `omni-livekit` for a complete reference implementation
- Add provider-specific features as extensions
- Submit a PR to register your provider
