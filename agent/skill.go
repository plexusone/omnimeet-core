// Package agent provides OmniAgent integration for OmniMeet.
//
// This package provides a meeting skill that enables AI agents to create,
// join, and manage meetings through the OmniAgent framework.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/provider"
	"github.com/plexusone/omnimeet-core/token"
)

// MeetingSkill provides meeting capabilities to OmniAgent.
type MeetingSkill struct {
	provider provider.MeetingProvider
	factory  provider.AgentParticipantFactory
	config   SkillConfig

	// Active sessions (agent participants in meetings)
	sessions map[string]*MeetingSession
	mu       sync.RWMutex

	// Event handlers
	onMeetingEvent func(event MeetingEvent)
}

// SkillConfig configures the MeetingSkill.
type SkillConfig struct {
	// DefaultMeetingName is used when creating meetings without a name.
	DefaultMeetingName string
	// DefaultAgentName is the default name for the agent participant.
	DefaultAgentName string
	// AutoJoinAsAgent controls whether the agent automatically joins created meetings.
	AutoJoinAsAgent bool
	// TranscriptionEnabled enables real-time transcription.
	TranscriptionEnabled bool
}

// MeetingSession represents an active meeting session.
type MeetingSession struct {
	Meeting      *meeting.Meeting
	Agent        provider.AgentParticipant
	JoinedAt     time.Time
	Transcript   []TranscriptEntry
	Participants []participant.Participant

	mu sync.RWMutex
}

// TranscriptEntry represents a transcript segment.
type TranscriptEntry struct {
	ParticipantID   string    `json:"participant_id"`
	ParticipantName string    `json:"participant_name"`
	Text            string    `json:"text"`
	Timestamp       time.Time `json:"timestamp"`
	IsFinal         bool      `json:"is_final"`
}

// MeetingEvent represents an event during a meeting.
type MeetingEvent struct {
	Type      string    `json:"type"`
	MeetingID string    `json:"meeting_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// NewMeetingSkill creates a new MeetingSkill.
func NewMeetingSkill(prov provider.MeetingProvider, cfg SkillConfig) (*MeetingSkill, error) {
	if prov == nil {
		return nil, fmt.Errorf("provider is required")
	}

	// Set defaults
	if cfg.DefaultMeetingName == "" {
		cfg.DefaultMeetingName = "AI Meeting"
	}
	if cfg.DefaultAgentName == "" {
		cfg.DefaultAgentName = "AI Assistant"
	}

	skill := &MeetingSkill{
		provider: prov,
		config:   cfg,
		sessions: make(map[string]*MeetingSession),
	}

	// Check if provider supports agent participation
	if factory, ok := prov.(provider.AgentParticipantFactory); ok {
		skill.factory = factory
	}

	return skill, nil
}

// Name returns the skill name.
func (s *MeetingSkill) Name() string {
	return "meeting"
}

// Description returns the skill description.
func (s *MeetingSkill) Description() string {
	return "Manage real-time meetings and video conferences. Create meetings, invite participants, join as an AI agent, and interact via voice."
}

// Tools returns the skill's tools.
func (s *MeetingSkill) Tools() []Tool {
	return []Tool{
		&createMeetingTool{skill: s},
		&getMeetingTool{skill: s},
		&listMeetingsTool{skill: s},
		&endMeetingTool{skill: s},
		&joinMeetingTool{skill: s},
		&leaveMeetingTool{skill: s},
		&getJoinLinkTool{skill: s},
		&listParticipantsTool{skill: s},
		&speakTool{skill: s},
		&getTranscriptTool{skill: s},
	}
}

// Init initializes the skill.
func (s *MeetingSkill) Init(ctx context.Context) error {
	return nil
}

// Close closes the skill.
func (s *MeetingSkill) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Leave all active meetings
	for _, session := range s.sessions {
		if session.Agent != nil {
			_ = session.Agent.LeaveMeeting(context.Background())
		}
	}
	s.sessions = make(map[string]*MeetingSession)

	return nil
}

// OnMeetingEvent sets the event handler.
func (s *MeetingSkill) OnMeetingEvent(handler func(MeetingEvent)) {
	s.onMeetingEvent = handler
}

// GetSession returns an active meeting session.
func (s *MeetingSkill) GetSession(meetingID string) *MeetingSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[meetingID]
}

// Tool interface for OmniAgent integration.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

// --- Tool implementations ---

type createMeetingTool struct {
	skill *MeetingSkill
}

func (t *createMeetingTool) Name() string { return "create_meeting" }
func (t *createMeetingTool) Description() string {
	return "Create a new meeting room. Returns the meeting ID and join link."
}
func (t *createMeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the meeting",
			},
			"max_participants": map[string]any{
				"type":        "integer",
				"description": "Maximum number of participants (optional)",
			},
		},
	}
}

func (t *createMeetingTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Name            string `json:"name"`
		MaxParticipants int    `json:"max_participants"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	name := params.Name
	if name == "" {
		name = fmt.Sprintf("%s-%d", t.skill.config.DefaultMeetingName, time.Now().Unix())
	}

	m, err := t.skill.provider.CreateMeeting(ctx, meeting.CreateRequest{
		Name:            name,
		MaxParticipants: params.MaxParticipants,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create meeting: %w", err)
	}

	// Generate join link
	tok, err := t.skill.provider.CreateJoinToken(ctx, token.CreateRequest{
		MeetingID: m.ID,
		Participant: participant.Info{
			Name: "Guest",
			Kind: participant.KindHuman,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate join link: %w", err)
	}

	result := map[string]any{
		"meeting_id": m.ID,
		"name":       m.Name,
		"join_url":   tok.JoinURL,
		"status":     string(m.Status),
	}

	// Auto-join if configured
	if t.skill.config.AutoJoinAsAgent && t.skill.factory != nil {
		if err := t.skill.joinAsAgent(ctx, m.ID); err != nil {
			result["agent_join_error"] = err.Error()
		} else {
			result["agent_joined"] = true
		}
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}

type getMeetingTool struct {
	skill *MeetingSkill
}

func (t *getMeetingTool) Name() string { return "get_meeting" }
func (t *getMeetingTool) Description() string {
	return "Get details about a specific meeting by ID."
}
func (t *getMeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *getMeetingTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	m, err := t.skill.provider.GetMeeting(ctx, params.MeetingID)
	if err != nil {
		return "", fmt.Errorf("failed to get meeting: %w", err)
	}

	data, _ := json.Marshal(m)
	return string(data), nil
}

type listMeetingsTool struct {
	skill *MeetingSkill
}

func (t *listMeetingsTool) Name() string { return "list_meetings" }
func (t *listMeetingsTool) Description() string {
	return "List all active meetings."
}
func (t *listMeetingsTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *listMeetingsTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	meetings, err := t.skill.provider.ListMeetings(ctx, meeting.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list meetings: %w", err)
	}

	data, _ := json.Marshal(meetings)
	return string(data), nil
}

type endMeetingTool struct {
	skill *MeetingSkill
}

func (t *endMeetingTool) Name() string { return "end_meeting" }
func (t *endMeetingTool) Description() string {
	return "End a meeting and disconnect all participants."
}
func (t *endMeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID to end",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *endMeetingTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	// Leave the meeting first if we're in it
	t.skill.mu.Lock()
	if session, ok := t.skill.sessions[params.MeetingID]; ok {
		if session.Agent != nil {
			_ = session.Agent.LeaveMeeting(ctx)
		}
		delete(t.skill.sessions, params.MeetingID)
	}
	t.skill.mu.Unlock()

	if err := t.skill.provider.EndMeeting(ctx, params.MeetingID); err != nil {
		return "", fmt.Errorf("failed to end meeting: %w", err)
	}

	return `{"success": true}`, nil
}

type joinMeetingTool struct {
	skill *MeetingSkill
}

func (t *joinMeetingTool) Name() string { return "join_meeting" }
func (t *joinMeetingTool) Description() string {
	return "Join a meeting as an AI agent to listen and speak."
}
func (t *joinMeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID to join",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *joinMeetingTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	if t.skill.factory == nil {
		return "", fmt.Errorf("agent participation not supported by this provider")
	}

	if err := t.skill.joinAsAgent(ctx, params.MeetingID); err != nil {
		return "", err
	}

	return `{"success": true, "message": "Joined meeting as AI agent"}`, nil
}

type leaveMeetingTool struct {
	skill *MeetingSkill
}

func (t *leaveMeetingTool) Name() string { return "leave_meeting" }
func (t *leaveMeetingTool) Description() string {
	return "Leave a meeting the agent is currently in."
}
func (t *leaveMeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID to leave",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *leaveMeetingTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	t.skill.mu.Lock()
	session, ok := t.skill.sessions[params.MeetingID]
	if ok {
		delete(t.skill.sessions, params.MeetingID)
	}
	t.skill.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("not in meeting: %s", params.MeetingID)
	}

	if session.Agent != nil {
		if err := session.Agent.LeaveMeeting(ctx); err != nil {
			return "", fmt.Errorf("failed to leave meeting: %w", err)
		}
	}

	return `{"success": true}`, nil
}

type getJoinLinkTool struct {
	skill *MeetingSkill
}

func (t *getJoinLinkTool) Name() string { return "get_join_link" }
func (t *getJoinLinkTool) Description() string {
	return "Generate a join link for a participant to join a meeting."
}
func (t *getJoinLinkTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID",
			},
			"participant_name": map[string]any{
				"type":        "string",
				"description": "Name of the participant",
			},
		},
		"required": []string{"meeting_id", "participant_name"},
	}
}

func (t *getJoinLinkTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID       string `json:"meeting_id"`
		ParticipantName string `json:"participant_name"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	tok, err := t.skill.provider.CreateJoinToken(ctx, token.CreateRequest{
		MeetingID: params.MeetingID,
		Participant: participant.Info{
			Name: params.ParticipantName,
			Kind: participant.KindHuman,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate join link: %w", err)
	}

	result := map[string]any{
		"join_url": tok.JoinURL,
		"token":    tok.Token,
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

type listParticipantsTool struct {
	skill *MeetingSkill
}

func (t *listParticipantsTool) Name() string { return "list_participants" }
func (t *listParticipantsTool) Description() string {
	return "List all participants in a meeting."
}
func (t *listParticipantsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *listParticipantsTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	participants, err := t.skill.provider.ListParticipants(ctx, params.MeetingID)
	if err != nil {
		return "", fmt.Errorf("failed to list participants: %w", err)
	}

	data, _ := json.Marshal(participants)
	return string(data), nil
}

type speakTool struct {
	skill *MeetingSkill
}

func (t *speakTool) Name() string { return "speak_in_meeting" }
func (t *speakTool) Description() string {
	return "Speak text in a meeting using text-to-speech. The agent must be in the meeting."
}
func (t *speakTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID",
			},
			"text": map[string]any{
				"type":        "string",
				"description": "The text to speak",
			},
		},
		"required": []string{"meeting_id", "text"},
	}
}

func (t *speakTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
		Text      string `json:"text"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	t.skill.mu.RLock()
	session, ok := t.skill.sessions[params.MeetingID]
	t.skill.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("not in meeting: %s", params.MeetingID)
	}

	// Check if the session has a VoiceAgentParticipant
	vap, ok := session.Agent.(interface {
		Speak(ctx context.Context, text string) error
	})
	if !ok {
		return "", fmt.Errorf("voice not enabled for this session")
	}

	if err := vap.Speak(ctx, params.Text); err != nil {
		return "", fmt.Errorf("failed to speak: %w", err)
	}

	return `{"success": true}`, nil
}

type getTranscriptTool struct {
	skill *MeetingSkill
}

func (t *getTranscriptTool) Name() string { return "get_meeting_transcript" }
func (t *getTranscriptTool) Description() string {
	return "Get the transcript of a meeting the agent is in."
}
func (t *getTranscriptTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"meeting_id": map[string]any{
				"type":        "string",
				"description": "The meeting ID",
			},
		},
		"required": []string{"meeting_id"},
	}
}

func (t *getTranscriptTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	t.skill.mu.RLock()
	session, ok := t.skill.sessions[params.MeetingID]
	t.skill.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("not in meeting: %s", params.MeetingID)
	}

	session.mu.RLock()
	transcript := make([]TranscriptEntry, len(session.Transcript))
	copy(transcript, session.Transcript)
	session.mu.RUnlock()

	data, _ := json.Marshal(transcript)
	return string(data), nil
}

// --- Internal methods ---

func (s *MeetingSkill) joinAsAgent(ctx context.Context, meetingID string) error {
	// Check if already in meeting
	s.mu.RLock()
	if _, ok := s.sessions[meetingID]; ok {
		s.mu.RUnlock()
		return fmt.Errorf("already in meeting: %s", meetingID)
	}
	s.mu.RUnlock()

	// Get meeting info
	m, err := s.provider.GetMeeting(ctx, meetingID)
	if err != nil {
		return fmt.Errorf("failed to get meeting: %w", err)
	}

	// Create agent participant
	agent, err := s.factory.CreateAgentParticipant(provider.AgentParticipantOptions{
		AutoSubscribe: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create agent participant: %w", err)
	}

	// Generate token for agent
	tok, err := s.provider.CreateJoinToken(ctx, token.CreateRequest{
		MeetingID: meetingID,
		Participant: participant.Info{
			Name:     s.config.DefaultAgentName,
			Kind:     participant.KindAgent,
			Identity: fmt.Sprintf("agent-%d", time.Now().UnixNano()),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to generate agent token: %w", err)
	}

	// Join the meeting
	if err := agent.JoinMeeting(ctx, meetingID, tok); err != nil {
		return fmt.Errorf("failed to join meeting: %w", err)
	}

	// Create session
	session := &MeetingSession{
		Meeting:  m,
		Agent:    agent,
		JoinedAt: time.Now(),
	}

	// Set up event handlers
	agent.OnParticipantJoined(func(p participant.Participant) {
		session.mu.Lock()
		session.Participants = append(session.Participants, p)
		session.mu.Unlock()

		s.emitEvent(MeetingEvent{
			Type:      "participant_joined",
			MeetingID: meetingID,
			Timestamp: time.Now(),
			Data:      p,
		})
	})

	agent.OnParticipantLeft(func(p participant.Participant) {
		session.mu.Lock()
		for i, existing := range session.Participants {
			if existing.ID == p.ID {
				session.Participants = append(session.Participants[:i], session.Participants[i+1:]...)
				break
			}
		}
		session.mu.Unlock()

		s.emitEvent(MeetingEvent{
			Type:      "participant_left",
			MeetingID: meetingID,
			Timestamp: time.Now(),
			Data:      p,
		})
	})

	// Store session
	s.mu.Lock()
	s.sessions[meetingID] = session
	s.mu.Unlock()

	s.emitEvent(MeetingEvent{
		Type:      "agent_joined",
		MeetingID: meetingID,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *MeetingSkill) emitEvent(event MeetingEvent) {
	if s.onMeetingEvent != nil {
		s.onMeetingEvent(event)
	}
}
