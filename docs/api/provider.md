# Provider Types

Package `provider` defines the core provider interfaces.

## MeetingProvider

Main interface for meeting management (control plane).

```go
type MeetingProvider interface {
    // Name returns the provider name (e.g., "livekit", "daily").
    Name() string

    // CreateMeeting creates a new meeting.
    CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error)

    // GetMeeting retrieves meeting information.
    GetMeeting(ctx context.Context, id string) (*meeting.Meeting, error)

    // ListMeetings returns all active meetings.
    ListMeetings(ctx context.Context) ([]*meeting.Meeting, error)

    // DeleteMeeting ends and removes a meeting.
    DeleteMeeting(ctx context.Context, id string) error

    // CreateJoinToken generates a token for joining a meeting.
    CreateJoinToken(ctx context.Context, req token.CreateRequest) (*token.JoinToken, error)

    // GetParticipants returns current participants in a meeting.
    GetParticipants(ctx context.Context, meetingID string) ([]participant.Participant, error)

    // RemoveParticipant kicks a participant from a meeting.
    RemoveParticipant(ctx context.Context, meetingID, participantID string) error
}
```

## AgentParticipantFactory

Interface for creating agent participants.

```go
type AgentParticipantFactory interface {
    CreateAgentParticipant(opts AgentParticipantOptions) (AgentParticipant, error)
}
```

## AgentParticipantOptions

Options for creating an agent participant.

```go
type AgentParticipantOptions struct {
    // AutoSubscribe automatically subscribes to new tracks.
    AutoSubscribe bool
}
```

## AgentParticipant

Interface for agents in meetings (media plane).

```go
type AgentParticipant interface {
    // Connection management
    JoinMeeting(ctx context.Context, meetingID string, tok *token.JoinToken) error
    LeaveMeeting(ctx context.Context) error
    ConnectionState() ConnectionState

    // Meeting state
    Meeting() *meeting.Meeting
    LocalParticipant() *participant.Participant
    RemoteParticipants() []participant.Participant
    GetParticipant(participantID string) *participant.Participant

    // Audio
    SubscribeToAudio(ctx context.Context, participantID string) (<-chan AudioFrame, error)
    SubscribeToAllAudio(ctx context.Context) (<-chan AudioFrame, error)
    PublishAudio(ctx context.Context, frame AudioFrame) error
    StartAudioTrack(ctx context.Context, opts AudioTrackOptions) (AudioWriter, error)
    StopAudioTrack(ctx context.Context) error

    // Tracks
    SubscribeToTrack(ctx context.Context, trackID string, opts track.SubscribeOptions) error
    UnsubscribeFromTrack(ctx context.Context, trackID string) error

    // Data messages
    SendDataMessage(ctx context.Context, msg DataMessage) error
    OnDataMessage(handler func(DataMessage))

    // Event handlers
    OnParticipantJoined(handler func(participant.Participant))
    OnParticipantLeft(handler func(participant.Participant))
    OnTrackPublished(handler func(participant.Participant, track.Track))
    OnTrackUnpublished(handler func(participant.Participant, track.Track))
    OnActiveSpeakerChanged(handler func([]participant.Participant))
    Events() <-chan event.Event
}
```

## ConnectionState

```go
type ConnectionState int

const (
    ConnectionStateDisconnected ConnectionState = iota
    ConnectionStateConnecting
    ConnectionStateConnected
    ConnectionStateReconnecting
)
```

## AudioFrame

Audio data from a participant.

```go
type AudioFrame struct {
    // Data is the raw audio data (PCM).
    Data []byte

    // ParticipantID is who sent this frame.
    ParticipantID string

    // ParticipantName is the sender's name.
    ParticipantName string

    // SampleRate is the audio sample rate.
    SampleRate int

    // Channels is the number of audio channels.
    Channels int

    // Timestamp is when this frame was captured.
    Timestamp time.Time
}
```

## AudioTrackOptions

Options for starting an audio track.

```go
type AudioTrackOptions struct {
    // SampleRate is the audio sample rate (e.g., 48000).
    SampleRate int

    // Channels is the number of channels (1 = mono, 2 = stereo).
    Channels int
}
```

## AudioWriter

Interface for writing audio frames.

```go
type AudioWriter interface {
    Write(frame AudioFrame) error
    Close() error
}
```

## DataMessage

Data sent between participants.

```go
type DataMessage struct {
    // Topic categorizes the message.
    Topic string

    // Data is the message payload.
    Data []byte

    // ParticipantID is the sender (when receiving).
    ParticipantID string

    // DestinationIDs limits recipients (empty = broadcast).
    DestinationIDs []string

    // Reliable ensures delivery (may be slower).
    Reliable bool
}
```

## Example Usage

```go
import "github.com/plexusone/omnimeet-core/provider"

// Type assertion to get factory
factory, ok := prov.(provider.AgentParticipantFactory)
if !ok {
    log.Fatal("provider does not support agent participation")
}

// Create agent
agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
    AutoSubscribe: true,
})

// Join and process audio
agent.JoinMeeting(ctx, meetingID, tok)

audioCh, _ := agent.SubscribeToAllAudio(ctx)
for frame := range audioCh {
    // Process audio
    processAudio(frame.Data, frame.SampleRate)
}
```
