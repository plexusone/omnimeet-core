# Testing

This guide covers testing OmniMeet applications, from unit tests to integration tests with live providers.

## Test Categories

| Category | Description | Requirements |
|----------|-------------|--------------|
| Unit | Test individual components | None |
| Integration | Test with real provider | API credentials |
| E2E | Full agent workflow | Provider + voice APIs |

## Unit Testing

### Mock Provider

```go
package mocks

import (
    "context"

    "github.com/plexusone/omnimeet-core/meeting"
    "github.com/plexusone/omnimeet-core/participant"
    "github.com/plexusone/omnimeet-core/provider"
    "github.com/plexusone/omnimeet-core/token"
)

type MockProvider struct {
    meetings     map[string]*meeting.Meeting
    CreateFunc   func(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error)
    GetFunc      func(ctx context.Context, id string) (*meeting.Meeting, error)
}

func NewMockProvider() *MockProvider {
    return &MockProvider{
        meetings: make(map[string]*meeting.Meeting),
    }
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error) {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, req)
    }

    mtg := &meeting.Meeting{
        ID:     "mock-" + req.Name,
        Name:   req.Name,
        Status: meeting.StatusActive,
    }
    m.meetings[mtg.ID] = mtg
    return mtg, nil
}

func (m *MockProvider) GetMeeting(ctx context.Context, id string) (*meeting.Meeting, error) {
    if m.GetFunc != nil {
        return m.GetFunc(ctx, id)
    }
    mtg, ok := m.meetings[id]
    if !ok {
        return nil, errors.New("meeting not found")
    }
    return mtg, nil
}

// ... implement other methods
```

### Mock Agent Participant

```go
type MockAgentParticipant struct {
    JoinedMeeting   string
    PublishedAudio  []provider.AudioFrame
    SentMessages    []provider.DataMessage

    joinedHandlers  []func(participant.Participant)
    leftHandlers    []func(participant.Participant)
}

func (m *MockAgentParticipant) JoinMeeting(ctx context.Context, meetingID string, tok *token.JoinToken) error {
    m.JoinedMeeting = meetingID
    return nil
}

func (m *MockAgentParticipant) PublishAudio(ctx context.Context, frame provider.AudioFrame) error {
    m.PublishedAudio = append(m.PublishedAudio, frame)
    return nil
}

func (m *MockAgentParticipant) OnParticipantJoined(handler func(participant.Participant)) {
    m.joinedHandlers = append(m.joinedHandlers, handler)
}

// Simulate a participant joining
func (m *MockAgentParticipant) SimulateJoin(p participant.Participant) {
    for _, h := range m.joinedHandlers {
        h(p)
    }
}
```

### Unit Test Example

```go
func TestMeetingSkillCreateMeeting(t *testing.T) {
    mockProvider := mocks.NewMockProvider()

    skill, err := agent.NewMeetingSkill(mockProvider, agent.SkillConfig{
        DefaultMeetingName: "Test Meeting",
    })
    if err != nil {
        t.Fatal(err)
    }

    // Find create tool
    var createTool agent.Tool
    for _, tool := range skill.Tools() {
        if tool.Name() == "create_meeting" {
            createTool = tool
            break
        }
    }

    // Execute
    result, err := createTool.Execute(context.Background(), []byte(`{"name": "My Meeting"}`))
    if err != nil {
        t.Fatal(err)
    }

    // Verify
    if !strings.Contains(result, "mock-My Meeting") {
        t.Errorf("expected meeting ID in result: %s", result)
    }
}
```

## Integration Testing

### Setup

Integration tests require real API credentials:

```go
//go:build integration

package omnimeet_test

import (
    "context"
    "os"
    "testing"

    "github.com/plexusone/omni-livekit/omnimeet"
)

func TestMain(m *testing.M) {
    // Skip if credentials not set
    if os.Getenv("LIVEKIT_API_KEY") == "" {
        os.Exit(0)
    }
    os.Exit(m.Run())
}

func newTestProvider(t *testing.T) *omnimeet.Provider {
    t.Helper()

    provider, err := omnimeet.NewProvider(omnimeet.Config{
        APIKey:    os.Getenv("LIVEKIT_API_KEY"),
        APISecret: os.Getenv("LIVEKIT_API_SECRET"),
        ServerURL: os.Getenv("LIVEKIT_URL"),
    })
    if err != nil {
        t.Fatal(err)
    }
    return provider
}
```

### Meeting Lifecycle Test

```go
func TestMeetingLifecycle(t *testing.T) {
    ctx := context.Background()
    provider := newTestProvider(t)

    // Create
    m, err := provider.CreateMeeting(ctx, meeting.CreateRequest{
        Name: "Integration Test",
    })
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        provider.DeleteMeeting(ctx, m.ID)
    })

    // Get
    retrieved, err := provider.GetMeeting(ctx, m.ID)
    if err != nil {
        t.Fatal(err)
    }
    if retrieved.Name != m.Name {
        t.Errorf("expected name %q, got %q", m.Name, retrieved.Name)
    }

    // List
    meetings, err := provider.ListMeetings(ctx)
    if err != nil {
        t.Fatal(err)
    }
    found := false
    for _, mtg := range meetings {
        if mtg.ID == m.ID {
            found = true
            break
        }
    }
    if !found {
        t.Error("meeting not found in list")
    }
}
```

### Agent Join Test

```go
func TestAgentJoinMeeting(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    provider := newTestProvider(t)

    // Create meeting
    m, err := provider.CreateMeeting(ctx, meeting.CreateRequest{
        Name: "Agent Test",
    })
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        provider.DeleteMeeting(context.Background(), m.ID)
    })

    // Create agent
    factory := provider.(provider.AgentParticipantFactory)
    agent, err := factory.CreateAgentParticipant(provider.AgentParticipantOptions{
        AutoSubscribe: true,
    })
    if err != nil {
        t.Fatal(err)
    }

    // Generate token
    tok, err := provider.CreateJoinToken(ctx, token.CreateRequest{
        MeetingID: m.ID,
        Participant: participant.Info{
            Name: "Test Agent",
            Kind: participant.KindAgent,
        },
    })
    if err != nil {
        t.Fatal(err)
    }

    // Join
    err = agent.JoinMeeting(ctx, m.ID, tok)
    if err != nil {
        t.Fatal(err)
    }
    defer agent.LeaveMeeting(ctx)

    // Verify connection
    state := agent.ConnectionState()
    if state != provider.ConnectionStateConnected {
        t.Errorf("expected Connected, got %v", state)
    }

    // Verify local participant
    local := agent.LocalParticipant()
    if local == nil {
        t.Fatal("expected local participant")
    }
    if local.Name != "Test Agent" {
        t.Errorf("expected name 'Test Agent', got %q", local.Name)
    }
}
```

### Multiple Participants Test

```go
func TestMultipleParticipants(t *testing.T) {
    ctx := context.Background()
    provider := newTestProvider(t)
    factory := provider.(provider.AgentParticipantFactory)

    // Create meeting
    m, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{Name: "Multi Test"})
    t.Cleanup(func() { provider.DeleteMeeting(context.Background(), m.ID) })

    // Create and join multiple agents
    agents := make([]provider.AgentParticipant, 3)
    for i := 0; i < 3; i++ {
        agent, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{})
        tok, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
            MeetingID: m.ID,
            Participant: participant.Info{
                Name: fmt.Sprintf("Agent-%d", i),
                Kind: participant.KindAgent,
            },
        })
        agent.JoinMeeting(ctx, m.ID, tok)
        agents[i] = agent
    }
    t.Cleanup(func() {
        for _, a := range agents {
            a.LeaveMeeting(context.Background())
        }
    })

    // Wait for connections
    time.Sleep(2 * time.Second)

    // Each agent should see the others
    for i, agent := range agents {
        remotes := agent.RemoteParticipants()
        if len(remotes) != 2 {
            t.Errorf("Agent-%d: expected 2 remote participants, got %d", i, len(remotes))
        }
    }
}
```

### Data Message Test

```go
func TestDataMessages(t *testing.T) {
    ctx := context.Background()
    provider := newTestProvider(t)
    factory := provider.(provider.AgentParticipantFactory)

    // Create meeting with two agents
    m, _ := provider.CreateMeeting(ctx, meeting.CreateRequest{Name: "Data Test"})
    t.Cleanup(func() { provider.DeleteMeeting(context.Background(), m.ID) })

    agent1, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{})
    agent2, _ := factory.CreateAgentParticipant(provider.AgentParticipantOptions{})

    tok1, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
        MeetingID:   m.ID,
        Participant: participant.Info{Name: "Agent1", Kind: participant.KindAgent},
    })
    tok2, _ := provider.CreateJoinToken(ctx, token.CreateRequest{
        MeetingID:   m.ID,
        Participant: participant.Info{Name: "Agent2", Kind: participant.KindAgent},
    })

    agent1.JoinMeeting(ctx, m.ID, tok1)
    agent2.JoinMeeting(ctx, m.ID, tok2)
    defer agent1.LeaveMeeting(ctx)
    defer agent2.LeaveMeeting(ctx)

    time.Sleep(2 * time.Second)

    // Set up receiver
    received := make(chan provider.DataMessage, 1)
    agent2.OnDataMessage(func(msg provider.DataMessage) {
        received <- msg
    })

    // Send message
    agent1.SendDataMessage(ctx, provider.DataMessage{
        Topic: "test",
        Data:  []byte("hello"),
    })

    // Wait for receipt
    select {
    case msg := <-received:
        if string(msg.Data) != "hello" {
            t.Errorf("expected 'hello', got %q", string(msg.Data))
        }
    case <-time.After(5 * time.Second):
        t.Error("timeout waiting for message")
    }
}
```

## Running Tests

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
# Set credentials
source .envrc

# Run integration tests
go test -v -tags=integration ./omnimeet/...

# Run specific test
go test -v -tags=integration -run TestAgentJoinMeeting ./omnimeet/...
```

### With Race Detection

```bash
go test -race -tags=integration ./...
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Mock Voice Providers

```go
type MockSTTProvider struct {
    TranscribeFunc func(ctx context.Context, audio []byte, config voice.STTConfig) (*voice.TranscriptionResult, error)
}

func (m *MockSTTProvider) Name() string { return "mock-stt" }

func (m *MockSTTProvider) Transcribe(ctx context.Context, audio []byte, config voice.STTConfig) (*voice.TranscriptionResult, error) {
    if m.TranscribeFunc != nil {
        return m.TranscribeFunc(ctx, audio, config)
    }
    return &voice.TranscriptionResult{
        Text:       "[mock transcription]",
        Confidence: 0.95,
        IsFinal:    true,
    }, nil
}

type MockTTSProvider struct {
    SynthesizeFunc func(ctx context.Context, text string, config voice.TTSConfig) (*voice.SynthesisResult, error)
}

func (m *MockTTSProvider) Name() string { return "mock-tts" }

func (m *MockTTSProvider) Synthesize(ctx context.Context, text string, config voice.TTSConfig) (*voice.SynthesisResult, error) {
    if m.SynthesizeFunc != nil {
        return m.SynthesizeFunc(ctx, text, config)
    }
    return &voice.SynthesisResult{
        Audio:      make([]byte, 1000), // Fake audio
        SampleRate: config.SampleRate,
        Format:     "pcm16",
    }, nil
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./...

  integration:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v -tags=integration ./omnimeet/...
        env:
          LIVEKIT_API_KEY: ${{ secrets.LIVEKIT_API_KEY }}
          LIVEKIT_API_SECRET: ${{ secrets.LIVEKIT_API_SECRET }}
          LIVEKIT_URL: ${{ secrets.LIVEKIT_URL }}
```

## Best Practices

1. **Isolate tests** - Each test creates its own meeting
2. **Clean up** - Use `t.Cleanup()` to delete meetings
3. **Timeouts** - Use context with timeout
4. **Skip gracefully** - Skip integration tests when credentials unavailable
5. **Parallel safety** - Use unique meeting names
6. **Wait for async** - Allow time for WebRTC connections
