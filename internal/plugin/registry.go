package plugin

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrPluginNotFound is returned when a requested plugin isn't registered.
	ErrPluginNotFound = errors.New("plugin not found")
)

// Registry manages the available plugins in the system.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry creates a new empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry. Returns an error if a plugin with
// the same name is already registered.
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}

	r.plugins[name] = p
	return nil
}

// Get retrieves a plugin by its name. Returns ErrPluginNotFound if it doesn't exist.
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotFound, name)
	}

	return p, nil
}

// List returns all registered plugins.
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}
