// Package voice provides integration between OmniMeet and OmniVoice.
//
// VoiceAgentParticipant wraps an AgentParticipant with OmniVoice STT/TTS
// capabilities, enabling AI agents to hear and speak in meetings.
//
// This package uses omnivoice-core types directly, allowing seamless
// switching between providers (Deepgram, OpenAI, ElevenLabs, Google, etc.).
package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/plexusone/omnimeet-core/participant"
	"github.com/plexusone/omnimeet-core/provider"
	"github.com/plexusone/omnivoice-core/stt"
	"github.com/plexusone/omnivoice-core/tts"
)

// TranscriptHandler is called when speech is transcribed.
type TranscriptHandler func(participantID, participantName, text string, isFinal bool)

// VoiceAgentParticipant wraps AgentParticipant with OmniVoice integration.
type VoiceAgentParticipant struct {
	participant   provider.AgentParticipant
	sttProvider   stt.Provider
	ttsProvider   tts.Provider
	sttConfig     stt.TranscriptionConfig
	ttsConfig     tts.SynthesisConfig
	onTranscript  TranscriptHandler

	// Active transcription streams per participant
	streams map[string]context.CancelFunc
	mu      sync.RWMutex
}

// Config configures the VoiceAgentParticipant.
type Config struct {
	// STTProvider is the speech-to-text provider (from omnivoice).
	STTProvider stt.Provider

	// TTSProvider is the text-to-speech provider (from omnivoice).
	TTSProvider tts.Provider

	// STTConfig configures speech-to-text.
	STTConfig stt.TranscriptionConfig

	// TTSConfig configures text-to-speech.
	TTSConfig tts.SynthesisConfig

	// OnTranscript is called when speech is transcribed.
	OnTranscript TranscriptHandler
}

// NewVoiceAgentParticipant creates a new VoiceAgentParticipant.
func NewVoiceAgentParticipant(participant provider.AgentParticipant, cfg Config) *VoiceAgentParticipant {
	// Set defaults
	if cfg.STTConfig.SampleRate == 0 {
		cfg.STTConfig.SampleRate = 16000
	}
	if cfg.STTConfig.Channels == 0 {
		cfg.STTConfig.Channels = 1
	}
	if cfg.TTSConfig.SampleRate == 0 {
		cfg.TTSConfig.SampleRate = 48000
	}

	return &VoiceAgentParticipant{
		participant:  participant,
		sttProvider:  cfg.STTProvider,
		ttsProvider:  cfg.TTSProvider,
		sttConfig:    cfg.STTConfig,
		ttsConfig:    cfg.TTSConfig,
		onTranscript: cfg.OnTranscript,
		streams:      make(map[string]context.CancelFunc),
	}
}

// Participant returns the underlying AgentParticipant.
func (v *VoiceAgentParticipant) Participant() provider.AgentParticipant {
	return v.participant
}

// StartTranscribing begins real-time transcription of a participant's audio.
func (v *VoiceAgentParticipant) StartTranscribing(ctx context.Context, participantID string) error {
	if v.sttProvider == nil {
		return fmt.Errorf("no STT provider configured")
	}

	// Check if already transcribing this participant
	v.mu.Lock()
	if _, exists := v.streams[participantID]; exists {
		v.mu.Unlock()
		return fmt.Errorf("already transcribing participant: %s", participantID)
	}

	// Create cancellable context for this stream
	streamCtx, cancel := context.WithCancel(ctx)
	v.streams[participantID] = cancel
	v.mu.Unlock()

	// Get participant name
	var participantName string
	if p := v.participant.GetParticipant(participantID); p != nil {
		participantName = p.Name
	}

	// Subscribe to audio
	audioCh, err := v.participant.SubscribeToAudio(streamCtx, participantID)
	if err != nil {
		v.mu.Lock()
		delete(v.streams, participantID)
		v.mu.Unlock()
		cancel()
		return fmt.Errorf("failed to subscribe to audio: %w", err)
	}

	// Start transcription goroutine
	go v.transcribeAudio(streamCtx, participantID, participantName, audioCh)

	return nil
}

// StopTranscribing stops transcription of a participant's audio.
func (v *VoiceAgentParticipant) StopTranscribing(participantID string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if cancel, exists := v.streams[participantID]; exists {
		cancel()
		delete(v.streams, participantID)
	}
}

// StartTranscribingAll begins transcription of all current participants.
func (v *VoiceAgentParticipant) StartTranscribingAll(ctx context.Context) error {
	for _, p := range v.participant.RemoteParticipants() {
		if p.Kind == participant.KindHuman && p.HasAudio() {
			if err := v.StartTranscribing(ctx, p.ID); err != nil {
				// Log but continue with other participants
				continue
			}
		}
	}
	return nil
}

// StopTranscribingAll stops all active transcriptions.
func (v *VoiceAgentParticipant) StopTranscribingAll() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for id, cancel := range v.streams {
		cancel()
		delete(v.streams, id)
	}
}

// Speak synthesizes text and publishes audio to the meeting.
func (v *VoiceAgentParticipant) Speak(ctx context.Context, text string) error {
	if v.ttsProvider == nil {
		return fmt.Errorf("no TTS provider configured")
	}

	// Synthesize speech
	result, err := v.ttsProvider.Synthesize(ctx, text, v.ttsConfig)
	if err != nil {
		return fmt.Errorf("failed to synthesize speech: %w", err)
	}

	// Convert sample rate if needed
	audio := result.Audio
	sampleRate := result.SampleRate
	if sampleRate == 0 {
		sampleRate = v.ttsConfig.SampleRate
	}

	// Resample to meeting audio format if needed (48kHz typical for WebRTC)
	targetSampleRate := 48000
	if sampleRate != targetSampleRate {
		audio = resampleAudio(audio, sampleRate, targetSampleRate, 1)
		sampleRate = targetSampleRate
	}

	// Publish audio frame
	return v.participant.PublishAudio(ctx, provider.AudioFrame{
		Data:       audio,
		SampleRate: sampleRate,
		Channels:   1,
		Timestamp:  time.Now(),
	})
}

// SpeakStream synthesizes text with streaming and publishes audio chunks.
func (v *VoiceAgentParticipant) SpeakStream(ctx context.Context, text string) error {
	// Try streaming synthesis
	chunks, err := v.ttsProvider.SynthesizeStream(ctx, text, v.ttsConfig)
	if err != nil {
		// Fall back to non-streaming
		return v.Speak(ctx, text)
	}

	targetSampleRate := 48000

	for chunk := range chunks {
		// Resample if needed
		audio := chunk.Audio
		if v.ttsConfig.SampleRate != targetSampleRate {
			audio = resampleAudio(chunk.Audio, v.ttsConfig.SampleRate, targetSampleRate, 1)
		}

		if err := v.participant.PublishAudio(ctx, provider.AudioFrame{
			Data:       audio,
			SampleRate: targetSampleRate,
			Channels:   1,
			Timestamp:  time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to publish audio: %w", err)
		}
	}

	return nil
}

// OnTranscript sets the transcript handler.
func (v *VoiceAgentParticipant) OnTranscript(handler TranscriptHandler) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.onTranscript = handler
}

// transcribeAudio processes audio frames and calls the transcript handler.
func (v *VoiceAgentParticipant) transcribeAudio(ctx context.Context, participantID, participantName string, audioCh <-chan provider.AudioFrame) {
	defer func() {
		v.mu.Lock()
		delete(v.streams, participantID)
		v.mu.Unlock()
	}()

	// Check if we have streaming STT
	streamingSTT, hasStreaming := v.sttProvider.(stt.StreamingProvider)

	if hasStreaming {
		// Use streaming transcription
		writer, eventCh, err := streamingSTT.TranscribeStream(ctx, v.sttConfig)
		if err != nil {
			return
		}
		defer func() { _ = writer.Close() }()

		// Forward audio to STT
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case frame, ok := <-audioCh:
					if !ok {
						return
					}
					// Resample to STT format if needed
					audio := frame.Data
					if frame.SampleRate != v.sttConfig.SampleRate {
						audio = resampleAudio(frame.Data, frame.SampleRate, v.sttConfig.SampleRate, frame.Channels)
					}
					_, _ = writer.Write(audio)
				}
			}
		}()

		// Process transcription events
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-eventCh:
				if !ok {
					return
				}
				if evt.Type == stt.EventTranscript {
					v.mu.RLock()
					handler := v.onTranscript
					v.mu.RUnlock()
					if handler != nil {
						handler(participantID, participantName, evt.Transcript, evt.IsFinal)
					}
				}
			}
		}
	} else {
		// Use batch transcription with buffering
		var audioBuffer []byte
		ticker := time.NewTicker(500 * time.Millisecond) // Transcribe every 500ms
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-audioCh:
				if !ok {
					return
				}
				// Resample to STT format if needed
				audio := frame.Data
				if frame.SampleRate != v.sttConfig.SampleRate {
					audio = resampleAudio(frame.Data, frame.SampleRate, v.sttConfig.SampleRate, frame.Channels)
				}
				audioBuffer = append(audioBuffer, audio...)
			case <-ticker.C:
				if len(audioBuffer) > 0 {
					// Transcribe buffered audio
					result, err := v.sttProvider.Transcribe(ctx, audioBuffer, v.sttConfig)
					audioBuffer = nil

					if err == nil && result.Text != "" {
						v.mu.RLock()
						handler := v.onTranscript
						v.mu.RUnlock()
						if handler != nil {
							handler(participantID, participantName, result.Text, true)
						}
					}
				}
			}
		}
	}
}

// resampleAudio resamples PCM16 audio from one sample rate to another.
// This is a simple linear interpolation - production code should use
// a proper resampling library like libsoxr.
func resampleAudio(input []byte, fromRate, toRate, channels int) []byte {
	if fromRate == toRate {
		return input
	}

	// Simple linear interpolation resampling
	// Each sample is 2 bytes (int16)
	samplesIn := len(input) / 2
	ratio := float64(toRate) / float64(fromRate)
	samplesOut := int(float64(samplesIn) * ratio)

	output := make([]byte, samplesOut*2)

	for i := 0; i < samplesOut; i++ {
		srcPos := float64(i) / ratio
		srcIdx := int(srcPos) * 2

		if srcIdx+1 >= len(input) {
			srcIdx = len(input) - 2
		}

		// Simple copy (proper implementation would interpolate)
		output[i*2] = input[srcIdx]
		output[i*2+1] = input[srcIdx+1]
	}

	return output
}

// ToSTTFormat converts a meeting AudioFrame to OmniVoice STT format.
func ToSTTFormat(frame provider.AudioFrame, targetSampleRate int) ([]byte, stt.TranscriptionConfig) {
	config := stt.TranscriptionConfig{
		SampleRate: targetSampleRate,
		Channels:   1,
	}

	data := frame.Data
	if frame.SampleRate != targetSampleRate {
		data = resampleAudio(frame.Data, frame.SampleRate, targetSampleRate, frame.Channels)
	}

	return data, config
}

// FromTTSResult converts an OmniVoice TTS result to a meeting AudioFrame.
func FromTTSResult(result *tts.SynthesisResult, targetSampleRate int) provider.AudioFrame {
	audio := result.Audio
	sampleRate := result.SampleRate

	if sampleRate != targetSampleRate {
		audio = resampleAudio(audio, sampleRate, targetSampleRate, 1)
		sampleRate = targetSampleRate
	}

	return provider.AudioFrame{
		Data:       audio,
		SampleRate: sampleRate,
		Channels:   1,
		Timestamp:  time.Now(),
	}
}
