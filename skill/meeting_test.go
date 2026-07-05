package skill

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/plexusone/omnimeet-core/event"
	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/provider"
	"github.com/plexusone/omnimeet-core/token"
)

// mockMeetingProvider implements provider.MeetingProvider for unit tests.
type mockMeetingProvider struct {
	meetings map[string]*meeting.Meeting
}

func newMockProvider() *mockMeetingProvider {
	return &mockMeetingProvider{
		meetings: make(map[string]*meeting.Meeting),
	}
}

func (m *mockMeetingProvider) Name() string { return "mock" }

func (m *mockMeetingProvider) CreateMeeting(ctx context.Context, req meeting.CreateRequest) (*meeting.Meeting, error) {
	mtg := &meeting.Meeting{
		ID:        fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		Name:      req.Name,
		Status:    meeting.StatusActive,
		CreatedAt: time.Now(),
	}
	m.meetings[mtg.ID] = mtg
	return mtg, nil
}

func (m *mockMeetingProvider) GetMeeting(ctx context.Context, id string) (*meeting.Meeting, error) {
	mtg, ok := m.meetings[id]
	if !ok {
		return nil, fmt.Errorf("meeting not found: %s", id)
	}
	return mtg, nil
}

func (m *mockMeetingProvider) ListMeetings(ctx context.Context, opts meeting.ListOptions) ([]meeting.Meeting, error) {
	var result []meeting.Meeting
	for _, mtg := range m.meetings {
		result = append(result, *mtg)
	}
	return result, nil
}

func (m *mockMeetingProvider) EndMeeting(ctx context.Context, id string) error {
	delete(m.meetings, id)
	return nil
}

func (m *mockMeetingProvider) DeleteMeeting(ctx context.Context, id string) error {
	delete(m.meetings, id)
	return nil
}

func (m *mockMeetingProvider) CreateJoinToken(ctx context.Context, req token.CreateRequest) (*token.JoinToken, error) {
	return &token.JoinToken{
		Token:   "mock-token-" + req.MeetingID,
		JoinURL: "https://mock.example.com/join/" + req.MeetingID,
	}, nil
}

func (m *mockMeetingProvider) ListParticipants(ctx context.Context, meetingID string) ([]participant.Participant, error) {
	return []participant.Participant{}, nil
}

func (m *mockMeetingProvider) GetParticipant(ctx context.Context, meetingID, participantID string) (*participant.Participant, error) {
	return nil, fmt.Errorf("participant not found")
}

func (m *mockMeetingProvider) RemoveParticipant(ctx context.Context, meetingID, participantID string) error {
	return nil
}

func (m *mockMeetingProvider) UpdateParticipant(ctx context.Context, meetingID, participantID string, update provider.ParticipantUpdate) error {
	return nil
}

func (m *mockMeetingProvider) OnEvent(handler event.Handler) {}

func (m *mockMeetingProvider) Close() error { return nil }

func TestSkillTools(t *testing.T) {
	skill := New(nil, Config{
		DefaultAgentName:   "Test Agent",
		DefaultMeetingName: "Test Meeting",
	})

	tools := skill.Tools()
	if len(tools) != 9 {
		t.Errorf("expected 9 tools, got %d", len(tools))
	}

	expectedTools := []string{
		"create_meeting",
		"get_meeting",
		"list_meetings",
		"end_meeting",
		"join_meeting",
		"leave_meeting",
		"get_join_link",
		"list_participants",
		"speak_in_meeting",
	}

	for i, name := range expectedTools {
		if tools[i].Name() != name {
			t.Errorf("tool %d: expected %q, got %q", i, name, tools[i].Name())
		}
		if tools[i].Description() == "" {
			t.Errorf("tool %s has empty description", name)
		}
	}
}

func TestSkillLifecycle(t *testing.T) {
	skill := New(nil, Config{})

	if err := skill.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := skill.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestSkillName(t *testing.T) {
	skill := New(nil, Config{})

	if skill.Name() != "meeting" {
		t.Errorf("expected name 'meeting', got %q", skill.Name())
	}

	if skill.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestSkillEvents(t *testing.T) {
	skill := New(nil, Config{})

	var receivedEvent *Event
	skill.OnEvent(func(e Event) {
		receivedEvent = &e
	})

	skill.emitEvent(Event{
		Type:      "test_event",
		MeetingID: "test-123",
		Timestamp: time.Now(),
	})

	if receivedEvent == nil {
		t.Fatal("expected event to be received")
	}

	if receivedEvent.Type != "test_event" {
		t.Errorf("expected type 'test_event', got %q", receivedEvent.Type)
	}

	if receivedEvent.MeetingID != "test-123" {
		t.Errorf("expected meeting ID 'test-123', got %q", receivedEvent.MeetingID)
	}
}

func TestCreateMeetingTool(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{
		DefaultMeetingName: "Default Meeting",
	})

	ctx := context.Background()

	var createTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "create_meeting" {
			createTool = tool
			break
		}
	}

	if createTool == nil {
		t.Fatal("create_meeting tool not found")
	}

	result, err := createTool.Call(ctx, map[string]any{
		"name": "Test Meeting",
	})
	if err != nil {
		t.Fatalf("create_meeting failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if resultMap["name"] != "Test Meeting" {
		t.Errorf("expected name 'Test Meeting', got %v", resultMap["name"])
	}

	if resultMap["meeting_id"] == nil {
		t.Error("expected meeting_id in result")
	}

	if resultMap["join_url"] == nil {
		t.Error("expected join_url in result")
	}
}

func TestGetMeetingTool(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	mtg, err := mockProv.CreateMeeting(ctx, meeting.CreateRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create meeting: %v", err)
	}

	var getTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "get_meeting" {
			getTool = tool
			break
		}
	}

	result, err := getTool.Call(ctx, map[string]any{
		"meeting_id": mtg.ID,
	})
	if err != nil {
		t.Fatalf("get_meeting failed: %v", err)
	}

	resultMtg, ok := result.(*meeting.Meeting)
	if !ok {
		t.Fatalf("expected *meeting.Meeting, got %T", result)
	}

	if resultMtg.ID != mtg.ID {
		t.Errorf("expected ID %q, got %q", mtg.ID, resultMtg.ID)
	}
}

func TestListMeetingsTool(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	// Create meetings with unique IDs
	mtg1, err := mockProv.CreateMeeting(ctx, meeting.CreateRequest{Name: "Meeting 1"})
	if err != nil {
		t.Fatalf("failed to create meeting 1: %v", err)
	}
	time.Sleep(time.Millisecond) // Ensure different timestamps
	mtg2, err := mockProv.CreateMeeting(ctx, meeting.CreateRequest{Name: "Meeting 2"})
	if err != nil {
		t.Fatalf("failed to create meeting 2: %v", err)
	}

	// Verify both were created
	if mtg1.ID == mtg2.ID {
		t.Fatal("meetings should have different IDs")
	}

	var listTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "list_meetings" {
			listTool = tool
			break
		}
	}

	result, err := listTool.Call(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("list_meetings failed: %v", err)
	}

	meetings, ok := result.([]meeting.Meeting)
	if !ok {
		t.Fatalf("expected []meeting.Meeting, got %T", result)
	}

	if len(meetings) != 2 {
		t.Errorf("expected 2 meetings, got %d", len(meetings))
	}
}

func TestEndMeetingTool(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	mtg, err := mockProv.CreateMeeting(ctx, meeting.CreateRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create meeting: %v", err)
	}

	var endTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "end_meeting" {
			endTool = tool
			break
		}
	}

	result, err := endTool.Call(ctx, map[string]any{
		"meeting_id": mtg.ID,
	})
	if err != nil {
		t.Fatalf("end_meeting failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if resultMap["success"] != true {
		t.Error("expected success: true")
	}

	_, err = mockProv.GetMeeting(ctx, mtg.ID)
	if err == nil {
		t.Error("expected meeting to be deleted")
	}
}

func TestGetJoinLinkTool(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	mtg, err := mockProv.CreateMeeting(ctx, meeting.CreateRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("failed to create meeting: %v", err)
	}

	var linkTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "get_join_link" {
			linkTool = tool
			break
		}
	}

	result, err := linkTool.Call(ctx, map[string]any{
		"meeting_id":       mtg.ID,
		"participant_name": "Alice",
	})
	if err != nil {
		t.Fatalf("get_join_link failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if resultMap["join_url"] == nil {
		t.Error("expected join_url in result")
	}

	if resultMap["token"] == nil {
		t.Error("expected token in result")
	}
}

func TestJoinMeetingToolWithoutFactory(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	var joinTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "join_meeting" {
			joinTool = tool
			break
		}
	}

	_, err := joinTool.Call(ctx, map[string]any{
		"meeting_id": "test-123",
	})
	if err == nil {
		t.Error("expected error when provider doesn't support agent participation")
	}
}

func TestSpeakToolNotInMeeting(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	var speakTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "speak_in_meeting" {
			speakTool = tool
			break
		}
	}

	_, err := speakTool.Call(ctx, map[string]any{
		"meeting_id": "test-123",
		"text":       "Hello",
	})
	if err == nil {
		t.Error("expected error when not in meeting")
	}
}

func TestLeaveMeetingToolNotInMeeting(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	ctx := context.Background()

	var leaveTool interface {
		Call(context.Context, map[string]any) (any, error)
	}
	for _, tool := range skill.Tools() {
		if tool.Name() == "leave_meeting" {
			leaveTool = tool
			break
		}
	}

	_, err := leaveTool.Call(ctx, map[string]any{
		"meeting_id": "test-123",
	})
	if err == nil {
		t.Error("expected error when not in meeting")
	}
}

func TestDefaultConfig(t *testing.T) {
	skill := New(nil, Config{})

	if skill.config.DefaultMeetingName != "AI Meeting" {
		t.Errorf("expected default meeting name 'AI Meeting', got %q", skill.config.DefaultMeetingName)
	}

	if skill.config.DefaultAgentName != "AI Assistant" {
		t.Errorf("expected default agent name 'AI Assistant', got %q", skill.config.DefaultAgentName)
	}
}

func TestGetSession(t *testing.T) {
	skill := New(nil, Config{})

	// No session should exist
	session := skill.GetSession("nonexistent")
	if session != nil {
		t.Error("expected nil session for nonexistent meeting")
	}
}

func TestProviderAccessor(t *testing.T) {
	mockProv := newMockProvider()
	skill := New(mockProv, Config{})

	if skill.Provider() != mockProv {
		t.Error("expected Provider() to return the provider")
	}
}
