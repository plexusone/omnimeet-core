// Package provider defines the core provider interfaces for OmniMeet.
package provider

import (
	"sync"
)

// Client manages multiple meeting providers with fallback support.
//
// This is a generic implementation similar to other PlexusOne libraries,
// allowing for multi-provider setups with automatic fallback.
type Client[T Named] struct {
	primary   T
	fallbacks []T
	all       []T
	byName    map[string]T
	mu        sync.RWMutex
}

// NewClient creates a new Client with the given primary provider.
func NewClient[T Named](primary T, fallbacks ...T) *Client[T] {
	c := &Client[T]{
		primary:   primary,
		fallbacks: fallbacks,
		byName:    make(map[string]T),
	}

	// Build lookup maps
	c.all = append([]T{primary}, fallbacks...)
	for _, p := range c.all {
		c.byName[p.Name()] = p
	}

	return c
}

// Primary returns the primary provider.
func (c *Client[T]) Primary() T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.primary
}

// Fallbacks returns the fallback providers.
func (c *Client[T]) Fallbacks() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fallbacks
}

// All returns all providers (primary + fallbacks).
func (c *Client[T]) All() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.all
}

// Provider returns a provider by name.
func (c *Client[T]) Provider(name string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.byName[name]
	return p, ok
}

// SetPrimary sets the primary provider.
func (c *Client[T]) SetPrimary(primary T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.primary = primary
	c.rebuildMaps()
}

// AddFallback adds a fallback provider.
func (c *Client[T]) AddFallback(fallback T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fallbacks = append(c.fallbacks, fallback)
	c.rebuildMaps()
}

// rebuildMaps rebuilds the internal lookup maps.
// Must be called with lock held.
func (c *Client[T]) rebuildMaps() {
	c.all = append([]T{c.primary}, c.fallbacks...)
	c.byName = make(map[string]T)
	for _, p := range c.all {
		c.byName[p.Name()] = p
	}
}

// Names returns the names of all providers.
func (c *Client[T]) Names() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, len(c.all))
	for i, p := range c.all {
		names[i] = p.Name()
	}
	return names
}
