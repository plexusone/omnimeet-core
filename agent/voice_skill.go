// Package agent provides OmniAgent integration for OmniMeet.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/plexusone/omnimeet-core/event"
	"github.com/plexusone/omnimeet-core/meeting"
	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/provider"
	"github.com/plexusone/omnimeet-core/token"
	"github.com/plexusone/omnimeet-core/track"
	"github.com/plexusone/omnimeet-core/voice"
	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/tts"
)

// VoiceMeetingSkill extends MeetingSkill with voice capabilities.
// It integrates OmniVoice STT/TTS for real-time speech processing.
type VoiceMeetingSkill struct {
	*MeetingSkill
	voiceConfig VoiceSkillConfig
}

// VoiceSkillConfig configures voice integration.
type VoiceSkillConfig struct {
	// STTProvider is the speech-to-text provider (from omnivoice-core).
	STTProvider stt.Provider
	// TTSProvider is the text-to-speech provider (from omnivoice-core).
	TTSProvider tts.Provider
	// STTConfig configures speech-to-text (from omnivoice-core).
	STTConfig stt.TranscriptionConfig
	// TTSConfig configures text-to-speech (from omnivoice-core).
	TTSConfig tts.SynthesisConfig
	// OnTranscript is called when speech is transcribed.
	OnTranscript func(meetingID, participantID, participantName, text string, isFinal bool)
}

// NewVoiceMeetingSkill creates a new VoiceMeetingSkill.
func NewVoiceMeetingSkill(prov provider.MeetingProvider, cfg SkillConfig, voiceCfg VoiceSkillConfig) (*VoiceMeetingSkill, error) {
	baseSkill, err := NewMeetingSkill(prov, cfg)
	if err != nil {
		return nil, err
	}

	return &VoiceMeetingSkill{
		MeetingSkill: baseSkill,
		voiceConfig:  voiceCfg,
	}, nil
}

// Description returns the skill description.
func (s *VoiceMeetingSkill) Description() string {
	return "Manage real-time meetings with voice capabilities. Create meetings, join as an AI agent, listen to participants via STT, and respond with TTS."
}

// joinAsAgent joins a meeting with voice capabilities.
func (s *VoiceMeetingSkill) joinAsAgentWithVoice(ctx context.Context, meetingID string) error {
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
	baseAgent, err := s.factory.CreateAgentParticipant(provider.AgentParticipantOptions{
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
	if err := baseAgent.JoinMeeting(ctx, meetingID, tok); err != nil {
		return fmt.Errorf("failed to join meeting: %w", err)
	}

	// Wrap with VoiceAgentParticipant
	voiceAgent := voice.NewVoiceAgentParticipant(baseAgent, voice.Config{
		STTProvider: s.voiceConfig.STTProvider,
		TTSProvider: s.voiceConfig.TTSProvider,
		STTConfig:   s.voiceConfig.STTConfig,
		TTSConfig:   s.voiceConfig.TTSConfig,
		OnTranscript: func(participantID, participantName, text string, isFinal bool) {
			// Store in transcript
			s.mu.RLock()
			session, ok := s.sessions[meetingID]
			s.mu.RUnlock()

			if ok && isFinal {
				session.mu.Lock()
				session.Transcript = append(session.Transcript, TranscriptEntry{
					ParticipantID:   participantID,
					ParticipantName: participantName,
					Text:            text,
					Timestamp:       time.Now(),
					IsFinal:         isFinal,
				})
				session.mu.Unlock()
			}

			// Call external handler if provided
			if s.voiceConfig.OnTranscript != nil {
				s.voiceConfig.OnTranscript(meetingID, participantID, participantName, text, isFinal)
			}

			// Emit event
			s.emitEvent(MeetingEvent{
				Type:      "transcript_updated",
				MeetingID: meetingID,
				Timestamp: time.Now(),
				Data: TranscriptEntry{
					ParticipantID:   participantID,
					ParticipantName: participantName,
					Text:            text,
					Timestamp:       time.Now(),
					IsFinal:         isFinal,
				},
			})
		},
	})

	// Create session with voice-enabled agent
	session := &MeetingSession{
		Meeting:  m,
		Agent:    voiceAgentWrapper{voiceAgent},
		JoinedAt: time.Now(),
	}

	// Set up event handlers
	baseAgent.OnParticipantJoined(func(p participant.Participant) {
		session.mu.Lock()
		session.Participants = append(session.Participants, p)
		session.mu.Unlock()

		// Auto-start transcription for human participants
		if p.Kind == participant.KindHuman && s.config.TranscriptionEnabled {
			go func() {
				// Best effort - ignore errors as participant may leave before transcription starts
				_ = voiceAgent.StartTranscribing(ctx, p.ID)
			}()
		}

		s.emitEvent(MeetingEvent{
			Type:      "participant_joined",
			MeetingID: meetingID,
			Timestamp: time.Now(),
			Data:      p,
		})
	})

	baseAgent.OnParticipantLeft(func(p participant.Participant) {
		// Stop transcription
		voiceAgent.StopTranscribing(p.ID)

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

// Tools returns the skill's tools with voice-specific overrides.
func (s *VoiceMeetingSkill) Tools() []Tool {
	// Get base tools
	baseTools := s.MeetingSkill.Tools()

	// Find and replace join_meeting with voice-enabled version
	for i, tool := range baseTools {
		if tool.Name() == "join_meeting" {
			baseTools[i] = &joinMeetingVoiceTool{skill: s}
		}
	}

	return baseTools
}

type joinMeetingVoiceTool struct {
	skill *VoiceMeetingSkill
}

func (t *joinMeetingVoiceTool) Name() string { return "join_meeting" }
func (t *joinMeetingVoiceTool) Description() string {
	return "Join a meeting as an AI agent with voice capabilities (STT/TTS)."
}
func (t *joinMeetingVoiceTool) Parameters() map[string]any {
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

func (t *joinMeetingVoiceTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", err
	}

	if t.skill.factory == nil {
		return "", fmt.Errorf("agent participation not supported by this provider")
	}

	if err := t.skill.joinAsAgentWithVoice(ctx, params.MeetingID); err != nil {
		return "", err
	}

	return `{"success": true, "message": "Joined meeting as AI agent with voice capabilities"}`, nil
}

// voiceAgentWrapper wraps VoiceAgentParticipant to implement provider.AgentParticipant.
type voiceAgentWrapper struct {
	*voice.VoiceAgentParticipant
}

// JoinMeeting delegates to the underlying participant.
func (w voiceAgentWrapper) JoinMeeting(ctx context.Context, meetingID string, tok *token.JoinToken) error {
	return w.Participant().JoinMeeting(ctx, meetingID, tok)
}

// LeaveMeeting stops all transcriptions and leaves.
func (w voiceAgentWrapper) LeaveMeeting(ctx context.Context) error {
	w.StopTranscribingAll()
	return w.Participant().LeaveMeeting(ctx)
}

// Speak synthesizes text and publishes audio.
func (w voiceAgentWrapper) Speak(ctx context.Context, text string) error {
	return w.VoiceAgentParticipant.Speak(ctx, text)
}

// SubscribeToAudio delegates to the underlying participant.
func (w voiceAgentWrapper) SubscribeToAudio(ctx context.Context, participantID string) (<-chan provider.AudioFrame, error) {
	return w.Participant().SubscribeToAudio(ctx, participantID)
}

// SubscribeToAllAudio delegates to the underlying participant.
func (w voiceAgentWrapper) SubscribeToAllAudio(ctx context.Context) (<-chan provider.AudioFrame, error) {
	return w.Participant().SubscribeToAllAudio(ctx)
}

// PublishAudio delegates to the underlying participant.
func (w voiceAgentWrapper) PublishAudio(ctx context.Context, frame provider.AudioFrame) error {
	return w.Participant().PublishAudio(ctx, frame)
}

// StartAudioTrack delegates to the underlying participant.
func (w voiceAgentWrapper) StartAudioTrack(ctx context.Context, opts provider.AudioTrackOptions) (provider.AudioWriter, error) {
	return w.Participant().StartAudioTrack(ctx, opts)
}

// StopAudioTrack delegates to the underlying participant.
func (w voiceAgentWrapper) StopAudioTrack(ctx context.Context) error {
	return w.Participant().StopAudioTrack(ctx)
}

// SubscribeToTrack delegates to the underlying participant.
func (w voiceAgentWrapper) SubscribeToTrack(ctx context.Context, trackID string, opts track.SubscribeOptions) error {
	return w.Participant().SubscribeToTrack(ctx, trackID, opts)
}

// UnsubscribeFromTrack delegates to the underlying participant.
func (w voiceAgentWrapper) UnsubscribeFromTrack(ctx context.Context, trackID string) error {
	return w.Participant().UnsubscribeFromTrack(ctx, trackID)
}

// SendDataMessage delegates to the underlying participant.
func (w voiceAgentWrapper) SendDataMessage(ctx context.Context, msg provider.DataMessage) error {
	return w.Participant().SendDataMessage(ctx, msg)
}

// OnDataMessage delegates to the underlying participant.
func (w voiceAgentWrapper) OnDataMessage(handler func(provider.DataMessage)) {
	w.Participant().OnDataMessage(handler)
}

// OnParticipantJoined delegates to the underlying participant.
func (w voiceAgentWrapper) OnParticipantJoined(handler func(participant.Participant)) {
	w.Participant().OnParticipantJoined(handler)
}

// OnParticipantLeft delegates to the underlying participant.
func (w voiceAgentWrapper) OnParticipantLeft(handler func(participant.Participant)) {
	w.Participant().OnParticipantLeft(handler)
}

// OnTrackPublished delegates to the underlying participant.
func (w voiceAgentWrapper) OnTrackPublished(handler func(participant.Participant, track.Track)) {
	w.Participant().OnTrackPublished(handler)
}

// OnTrackUnpublished delegates to the underlying participant.
func (w voiceAgentWrapper) OnTrackUnpublished(handler func(participant.Participant, track.Track)) {
	w.Participant().OnTrackUnpublished(handler)
}

// OnActiveSpeakerChanged delegates to the underlying participant.
func (w voiceAgentWrapper) OnActiveSpeakerChanged(handler func([]participant.Participant)) {
	w.Participant().OnActiveSpeakerChanged(handler)
}

// Meeting delegates to the underlying participant.
func (w voiceAgentWrapper) Meeting() *meeting.Meeting {
	return w.Participant().Meeting()
}

// LocalParticipant delegates to the underlying participant.
func (w voiceAgentWrapper) LocalParticipant() *participant.Participant {
	return w.Participant().LocalParticipant()
}

// RemoteParticipants delegates to the underlying participant.
func (w voiceAgentWrapper) RemoteParticipants() []participant.Participant {
	return w.Participant().RemoteParticipants()
}

// GetParticipant delegates to the underlying participant.
func (w voiceAgentWrapper) GetParticipant(participantID string) *participant.Participant {
	return w.Participant().GetParticipant(participantID)
}

// ConnectionState delegates to the underlying participant.
func (w voiceAgentWrapper) ConnectionState() provider.ConnectionState {
	return w.Participant().ConnectionState()
}

// Events delegates to the underlying participant.
func (w voiceAgentWrapper) Events() <-chan event.Event {
	return w.Participant().Events()
}
