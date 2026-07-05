// Package skill provides an omniskill-compatible meeting skill for OmniAgent.
//
// This skill enables AI agents to create, join, and manage meetings through
// the OmniAgent framework using any OmniMeet provider (LiveKit, Daily, etc.).
//
// Usage with OmniAgent:
//
//	import (
//	    "github.com/plexusone/omni-livekit/omnimeet"
//	    meetingskill "github.com/plexusone/omnimeet-core/skill"
//	)
//
//	// Create provider
//	provider, _ := omnimeet.NewProvider(config)
//
//	// Create skill
//	skill := meetingskill.New(provider, meetingskill.Config{
//	    DefaultAgentName: "AI Assistant",
//	})
//
//	// Register with OmniAgent
//	agent.New(config, agent.WithCompiledSkill(skill))
package skill

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
	"github.com/plexusone/omniskill/skill"
)

// MeetingSkill provides meeting capabilities to OmniAgent via omniskill.
type MeetingSkill struct {
	provider provider.MeetingProvider
	factory  provider.AgentParticipantFactory
	config   Config

	// Active sessions (agent participants in meetings)
	sessions map[string]*Session
	mu       sync.RWMutex

	// Event handlers
	onEvent func(Event)
}

// Config configures the MeetingSkill.
type Config struct {
	// DefaultMeetingName is used when creating meetings without a name.
	DefaultMeetingName string
	// DefaultAgentName is the default name for the agent participant.
	DefaultAgentName string
	// AutoJoinAsAgent controls whether the agent automatically joins created meetings.
	AutoJoinAsAgent bool
}

// Session represents an active meeting session.
type Session struct {
	Meeting      *meeting.Meeting
	Agent        provider.AgentParticipant
	JoinedAt     time.Time
	Participants []participant.Participant

	mu sync.RWMutex
}

// Event represents a meeting event.
type Event struct {
	Type      string    `json:"type"`
	MeetingID string    `json:"meeting_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// New creates a new MeetingSkill.
func New(prov provider.MeetingProvider, cfg Config) *MeetingSkill {
	if cfg.DefaultMeetingName == "" {
		cfg.DefaultMeetingName = "AI Meeting"
	}
	if cfg.DefaultAgentName == "" {
		cfg.DefaultAgentName = "AI Assistant"
	}

	s := &MeetingSkill{
		provider: prov,
		config:   cfg,
		sessions: make(map[string]*Session),
	}

	// Check if provider supports agent participation
	if factory, ok := prov.(provider.AgentParticipantFactory); ok {
		s.factory = factory
	}

	return s
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
func (s *MeetingSkill) Tools() []skill.Tool {
	return []skill.Tool{
		s.createMeetingTool(),
		s.getMeetingTool(),
		s.listMeetingsTool(),
		s.endMeetingTool(),
		s.joinMeetingTool(),
		s.leaveMeetingTool(),
		s.getJoinLinkTool(),
		s.listParticipantsTool(),
		s.speakTool(),
	}
}

// Init initializes the skill.
func (s *MeetingSkill) Init(ctx context.Context) error {
	return nil
}

// Close closes the skill and leaves all meetings.
func (s *MeetingSkill) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.sessions {
		if session.Agent != nil {
			_ = session.Agent.LeaveMeeting(context.Background())
		}
	}
	s.sessions = make(map[string]*Session)

	return nil
}

// OnEvent sets the event handler.
func (s *MeetingSkill) OnEvent(handler func(Event)) {
	s.onEvent = handler
}

// GetSession returns an active meeting session.
func (s *MeetingSkill) GetSession(meetingID string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[meetingID]
}

// Provider returns the underlying meeting provider.
func (s *MeetingSkill) Provider() provider.MeetingProvider {
	return s.provider
}

// --- Tool implementations ---

func (s *MeetingSkill) createMeetingTool() skill.Tool {
	return skill.NewTool(
		"create_meeting",
		"Create a new meeting room. Returns the meeting ID and join link.",
		map[string]skill.Parameter{
			"name": {
				Type:        "string",
				Description: "Name of the meeting",
			},
			"max_participants": {
				Type:        "integer",
				Description: "Maximum number of participants (optional)",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			name, _ := params["name"].(string)
			if name == "" {
				name = fmt.Sprintf("%s-%d", s.config.DefaultMeetingName, time.Now().Unix())
			}

			var maxParticipants int
			if v, ok := params["max_participants"].(float64); ok {
				maxParticipants = int(v)
			}

			m, err := s.provider.CreateMeeting(ctx, meeting.CreateRequest{
				Name:            name,
				MaxParticipants: maxParticipants,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create meeting: %w", err)
			}

			// Generate join link
			tok, err := s.provider.CreateJoinToken(ctx, token.CreateRequest{
				MeetingID: m.ID,
				Participant: participant.Info{
					Name: "Guest",
					Kind: participant.KindHuman,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to generate join link: %w", err)
			}

			result := map[string]any{
				"meeting_id": m.ID,
				"name":       m.Name,
				"join_url":   tok.JoinURL,
				"status":     string(m.Status),
			}

			// Auto-join if configured
			if s.config.AutoJoinAsAgent && s.factory != nil {
				if err := s.joinAsAgent(ctx, m.ID); err != nil {
					result["agent_join_error"] = err.Error()
				} else {
					result["agent_joined"] = true
				}
			}

			s.emitEvent(Event{
				Type:      "meeting_created",
				MeetingID: m.ID,
				Timestamp: time.Now(),
				Data:      m,
			})

			return result, nil
		},
	)
}

func (s *MeetingSkill) getMeetingTool() skill.Tool {
	return skill.NewTool(
		"get_meeting",
		"Get details about a specific meeting by ID.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			if meetingID == "" {
				return nil, fmt.Errorf("meeting_id is required")
			}

			m, err := s.provider.GetMeeting(ctx, meetingID)
			if err != nil {
				return nil, fmt.Errorf("failed to get meeting: %w", err)
			}

			return m, nil
		},
	)
}

func (s *MeetingSkill) listMeetingsTool() skill.Tool {
	return skill.NewTool(
		"list_meetings",
		"List all active meetings.",
		map[string]skill.Parameter{},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetings, err := s.provider.ListMeetings(ctx, meeting.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list meetings: %w", err)
			}

			return meetings, nil
		},
	)
}

func (s *MeetingSkill) endMeetingTool() skill.Tool {
	return skill.NewTool(
		"end_meeting",
		"End a meeting and disconnect all participants.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID to end",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			if meetingID == "" {
				return nil, fmt.Errorf("meeting_id is required")
			}

			// Leave the meeting first if we're in it
			s.mu.Lock()
			if session, ok := s.sessions[meetingID]; ok {
				if session.Agent != nil {
					_ = session.Agent.LeaveMeeting(ctx)
				}
				delete(s.sessions, meetingID)
			}
			s.mu.Unlock()

			if err := s.provider.EndMeeting(ctx, meetingID); err != nil {
				return nil, fmt.Errorf("failed to end meeting: %w", err)
			}

			s.emitEvent(Event{
				Type:      "meeting_ended",
				MeetingID: meetingID,
				Timestamp: time.Now(),
			})

			return map[string]any{"success": true}, nil
		},
	)
}

func (s *MeetingSkill) joinMeetingTool() skill.Tool {
	return skill.NewTool(
		"join_meeting",
		"Join a meeting as an AI agent to listen and speak.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID to join",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			if meetingID == "" {
				return nil, fmt.Errorf("meeting_id is required")
			}

			if s.factory == nil {
				return nil, fmt.Errorf("agent participation not supported by this provider")
			}

			if err := s.joinAsAgent(ctx, meetingID); err != nil {
				return nil, err
			}

			return map[string]any{
				"success": true,
				"message": "Joined meeting as AI agent",
			}, nil
		},
	)
}

func (s *MeetingSkill) leaveMeetingTool() skill.Tool {
	return skill.NewTool(
		"leave_meeting",
		"Leave a meeting the agent is currently in.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID to leave",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			if meetingID == "" {
				return nil, fmt.Errorf("meeting_id is required")
			}

			s.mu.Lock()
			session, ok := s.sessions[meetingID]
			if ok {
				delete(s.sessions, meetingID)
			}
			s.mu.Unlock()

			if !ok {
				return nil, fmt.Errorf("not in meeting: %s", meetingID)
			}

			if session.Agent != nil {
				if err := session.Agent.LeaveMeeting(ctx); err != nil {
					return nil, fmt.Errorf("failed to leave meeting: %w", err)
				}
			}

			s.emitEvent(Event{
				Type:      "agent_left",
				MeetingID: meetingID,
				Timestamp: time.Now(),
			})

			return map[string]any{"success": true}, nil
		},
	)
}

func (s *MeetingSkill) getJoinLinkTool() skill.Tool {
	return skill.NewTool(
		"get_join_link",
		"Generate a join link for a participant to join a meeting.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID",
				Required:    true,
			},
			"participant_name": {
				Type:        "string",
				Description: "Name of the participant",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			participantName, _ := params["participant_name"].(string)

			if meetingID == "" || participantName == "" {
				return nil, fmt.Errorf("meeting_id and participant_name are required")
			}

			tok, err := s.provider.CreateJoinToken(ctx, token.CreateRequest{
				MeetingID: meetingID,
				Participant: participant.Info{
					Name: participantName,
					Kind: participant.KindHuman,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to generate join link: %w", err)
			}

			return map[string]any{
				"join_url": tok.JoinURL,
				"token":    tok.Token,
			}, nil
		},
	)
}

func (s *MeetingSkill) listParticipantsTool() skill.Tool {
	return skill.NewTool(
		"list_participants",
		"List all participants in a meeting.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			if meetingID == "" {
				return nil, fmt.Errorf("meeting_id is required")
			}

			participants, err := s.provider.ListParticipants(ctx, meetingID)
			if err != nil {
				return nil, fmt.Errorf("failed to list participants: %w", err)
			}

			return participants, nil
		},
	)
}

func (s *MeetingSkill) speakTool() skill.Tool {
	return skill.NewTool(
		"speak_in_meeting",
		"Speak text in a meeting using text-to-speech. The agent must be in the meeting.",
		map[string]skill.Parameter{
			"meeting_id": {
				Type:        "string",
				Description: "The meeting ID",
				Required:    true,
			},
			"text": {
				Type:        "string",
				Description: "The text to speak",
				Required:    true,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			meetingID, _ := params["meeting_id"].(string)
			text, _ := params["text"].(string)

			if meetingID == "" || text == "" {
				return nil, fmt.Errorf("meeting_id and text are required")
			}

			s.mu.RLock()
			session, ok := s.sessions[meetingID]
			s.mu.RUnlock()

			if !ok {
				return nil, fmt.Errorf("not in meeting: %s", meetingID)
			}

			// Check if the session has a VoiceAgentParticipant
			vap, ok := session.Agent.(interface {
				Speak(ctx context.Context, text string) error
			})
			if !ok {
				return nil, fmt.Errorf("voice not enabled for this session")
			}

			if err := vap.Speak(ctx, text); err != nil {
				return nil, fmt.Errorf("failed to speak: %w", err)
			}

			return map[string]any{"success": true}, nil
		},
	)
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
	session := &Session{
		Meeting:  m,
		Agent:    agent,
		JoinedAt: time.Now(),
	}

	// Set up event handlers
	agent.OnParticipantJoined(func(p participant.Participant) {
		session.mu.Lock()
		session.Participants = append(session.Participants, p)
		session.mu.Unlock()

		s.emitEvent(Event{
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

		s.emitEvent(Event{
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

	s.emitEvent(Event{
		Type:      "agent_joined",
		MeetingID: meetingID,
		Timestamp: time.Now(),
	})

	return nil
}

func (s *MeetingSkill) emitEvent(event Event) {
	if s.onEvent != nil {
		s.onEvent(event)
	}
}

// MarshalJSON implements json.Marshaler for tool results.
func (s *Session) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(map[string]any{
		"meeting":      s.Meeting,
		"joined_at":    s.JoinedAt,
		"participants": s.Participants,
	})
}

// Ensure MeetingSkill implements skill.Skill.
var _ skill.Skill = (*MeetingSkill)(nil)
