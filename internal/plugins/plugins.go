package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin represents a plugin
type Plugin interface {
	Name() string
	Version() string
	Initialize(ctx context.Context) error
	Shutdown() error
}

// Loader loads plugins
type Loader struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewLoader creates a new plugin loader
func NewLoader() *Loader {
	return &Loader{
		plugins: make(map[string]Plugin),
	}
}

// Load loads a plugin from a file
func (l *Loader) Load(path string) (Plugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Open the plugin file
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look up the plugin symbol
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export Plugin symbol: %w", err)
	}

	// Assert the plugin interface
	pl, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement Plugin interface")
	}

	l.plugins[pl.Name()] = pl
	return pl, nil
}

// Unload unloads a plugin
func (l *Loader) Unload(name string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	p, ok := l.plugins[name]
	if !ok {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if err := p.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugin: %w", err)
	}

	delete(l.plugins, name)
	return nil
}

// Get returns a plugin by name
func (l *Loader) Get(name string) (Plugin, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	p, ok := l.plugins[name]
	return p, ok
}

// List returns all loaded plugins
func (l *Loader) List() []Plugin {
	l.mu.RLock()
	defer l.mu.RUnlock()
	plugins := make([]Plugin, 0, len(l.plugins))
	for _, p := range l.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// ScanDir scans a directory for plugins
func (l *Loader) ScanDir(dir string) ([]Plugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".so" && ext != ".dylib" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		p, err := l.Load(path)
		if err != nil {
			// Log error but continue
			continue
		}
		plugins = append(plugins, p)
	}

	return plugins, nil
}

// DefaultLoader is the global plugin loader
var DefaultLoader = NewLoader()

// Load is a convenience function using the default loader
func Load(path string) (Plugin, error) {
	return DefaultLoader.Load(path)
}

// Get is a convenience function using the default loader
func Get(name string) (Plugin, bool) {
	return DefaultLoader.Get(name)
}

// List is a convenience function using the default loader
func List() []Plugin {
	return DefaultLoader.List()
}
