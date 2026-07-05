// Package omnimeet provides a unified abstraction for real-time collaboration platforms.
package omnimeet

import (
	"fmt"
	"sync"

	"github.com/plexusone/omnimeet-core/provider"
)

// Priority levels for provider registration.
const (
	// PriorityThin is for minimal implementations (e.g., stdlib-only).
	PriorityThin = 0
	// PriorityThick is for full SDK implementations.
	PriorityThick = 10
)

// ProviderFactory is a function that creates a MeetingProvider.
type ProviderFactory func(config map[string]any) (provider.MeetingProvider, error)

// registeredProvider holds a provider factory and its priority.
type registeredProvider struct {
	factory  ProviderFactory
	priority int
}

var (
	// meetingProviders holds registered meeting provider factories.
	meetingProviders = make(map[string]registeredProvider)
	meetingMu        sync.RWMutex
)

// RegisterMeetingProvider registers a meeting provider factory.
//
// If a provider with the same name is already registered, it will be replaced
// only if the new provider has equal or higher priority.
func RegisterMeetingProvider(name string, factory ProviderFactory, priority int) {
	meetingMu.Lock()
	defer meetingMu.Unlock()

	existing, exists := meetingProviders[name]
	if exists && existing.priority > priority {
		// Don't replace higher priority provider
		return
	}

	meetingProviders[name] = registeredProvider{
		factory:  factory,
		priority: priority,
	}
}

// GetMeetingProvider returns a meeting provider by name.
func GetMeetingProvider(name string, opts ...ProviderOption) (provider.MeetingProvider, error) {
	meetingMu.RLock()
	reg, exists := meetingProviders[name]
	meetingMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	// Apply options
	cfg := &providerConfig{
		config: make(map[string]any),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return reg.factory(cfg.config)
}

// ListMeetingProviders returns the names of all registered meeting providers.
func ListMeetingProviders() []string {
	meetingMu.RLock()
	defer meetingMu.RUnlock()

	names := make([]string, 0, len(meetingProviders))
	for name := range meetingProviders {
		names = append(names, name)
	}
	return names
}

// HasMeetingProvider returns true if a meeting provider is registered.
func HasMeetingProvider(name string) bool {
	meetingMu.RLock()
	defer meetingMu.RUnlock()
	_, exists := meetingProviders[name]
	return exists
}

// UnregisterMeetingProvider removes a meeting provider from the registry.
// This is primarily useful for testing.
func UnregisterMeetingProvider(name string) {
	meetingMu.Lock()
	defer meetingMu.Unlock()
	delete(meetingProviders, name)
}

// ClearMeetingProviders removes all meeting providers from the registry.
// This is primarily useful for testing.
func ClearMeetingProviders() {
	meetingMu.Lock()
	defer meetingMu.Unlock()
	meetingProviders = make(map[string]registeredProvider)
}

// providerConfig holds configuration for provider creation.
type providerConfig struct {
	config map[string]any
}

// ProviderOption is an option for provider creation.
type ProviderOption func(*providerConfig)

// WithConfig sets configuration values for the provider.
func WithConfig(config map[string]any) ProviderOption {
	return func(c *providerConfig) {
		for k, v := range config {
			c.config[k] = v
		}
	}
}

// WithAPIKey sets the API key for the provider.
func WithAPIKey(apiKey string) ProviderOption {
	return func(c *providerConfig) {
		c.config["api_key"] = apiKey
	}
}

// WithAPISecret sets the API secret for the provider.
func WithAPISecret(apiSecret string) ProviderOption {
	return func(c *providerConfig) {
		c.config["api_secret"] = apiSecret
	}
}

// WithServerURL sets the server URL for the provider.
func WithServerURL(url string) ProviderOption {
	return func(c *providerConfig) {
		c.config["server_url"] = url
	}
}

// WithWebhookSecret sets the webhook secret for signature validation.
func WithWebhookSecret(secret string) ProviderOption {
	return func(c *providerConfig) {
		c.config["webhook_secret"] = secret
	}
}
