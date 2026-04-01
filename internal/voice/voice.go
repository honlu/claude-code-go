package voice

import (
	"context"
	"fmt"
	"sync"
)

// VoiceService represents a voice service
type VoiceService interface {
	Name() string
	Initialize(ctx context.Context) error
	Shutdown() error
	IsEnabled() bool
}

// STTService represents a speech-to-text service
type STTService interface {
	VoiceService
	Transcribe(ctx context.Context, audio []byte) (string, error)
}

// TTSService represents a text-to-speech service
type TTSService interface {
	VoiceService
	Speak(ctx context.Context, text string) error
}

// VoiceManager manages voice services
type VoiceManager struct {
	mu         sync.RWMutex
	stt        STTService
	tts        TTSService
	enabled    bool
}

// NewManager creates a new voice manager
func NewManager() *VoiceManager {
	return &VoiceManager{
		enabled: false,
	}
}

// SetSTT sets the speech-to-text service
func (vm *VoiceManager) SetSTT(service STTService) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if err := service.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize STT: %w", err)
	}

	vm.stt = service
	vm.enabled = true
	return nil
}

// SetTTS sets the text-to-speech service
func (vm *VoiceManager) SetTTS(service TTSService) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if err := service.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize TTS: %w", err)
	}

	vm.tts = service
	vm.enabled = true
	return nil
}

// Transcribe transcribes audio to text
func (vm *VoiceManager) Transcribe(ctx context.Context, audio []byte) (string, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.stt == nil {
		return "", fmt.Errorf("STT service not configured")
	}
	return vm.stt.Transcribe(ctx, audio)
}

// Speak speaks text
func (vm *VoiceManager) Speak(ctx context.Context, text string) error {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if vm.tts == nil {
		return fmt.Errorf("TTS service not configured")
	}
	return vm.tts.Speak(ctx, text)
}

// IsEnabled returns whether voice is enabled
func (vm *VoiceManager) IsEnabled() bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.enabled
}

// DefaultManager is the global voice manager
var DefaultManager = NewManager()

// SetSTT is a convenience function using the default manager
func SetSTT(service STTService) error {
	return DefaultManager.SetSTT(service)
}

// SetTTS is a convenience function using the default manager
func SetTTS(service TTSService) error {
	return DefaultManager.SetTTS(service)
}

// Transcribe is a convenience function using the default manager
func Transcribe(ctx context.Context, audio []byte) (string, error) {
	return DefaultManager.Transcribe(ctx, audio)
}

// Speak is a convenience function using the default manager
func Speak(ctx context.Context, text string) error {
	return DefaultManager.Speak(ctx, text)
}

// IsEnabled is a convenience function using the default manager
func IsEnabled() bool {
	return DefaultManager.IsEnabled()
}
