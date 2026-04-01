package bridge

import (
	"fmt"
	"sync"
)

// Bridge provides platform-specific functionality
type Bridge interface {
	// Platform returns the platform name
	Platform() string

	// GetEnv gets an environment variable
	GetEnv(key string) (string, bool)

	// SetEnv sets an environment variable
	SetEnv(key, value string) error

	// Getwd returns the current working directory
	Getwd() (string, error)

	// HomeDir returns the user's home directory
	HomeDir() (string, error)

	// TempDir returns the temp directory
	TempDir() (string, error)

	// Open opens a URL or file
	Open(target string) error

	// Exists checks if a path exists
	Exists(path string) bool

	// IsAbs returns true if the path is absolute
	IsAbs(path string) bool

	// Abs returns the absolute path
	Abs(path string) (string, error)
}

// Manager manages bridges
type Manager struct {
	mu      sync.RWMutex
	bridges map[string]Bridge
	defaultBridge Bridge
}

// NewManager creates a new bridge manager
func NewManager() *Manager {
	return &Manager{
		bridges: make(map[string]Bridge),
	}
}

// Register registers a bridge
func (m *Manager) Register(name string, bridge Bridge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bridges[name] = bridge
	if m.defaultBridge == nil {
		m.defaultBridge = bridge
	}
}

// SetDefault sets the default bridge
func (m *Manager) SetDefault(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bridge, ok := m.bridges[name]
	if !ok {
		return fmt.Errorf("bridge not found: %s", name)
	}
	m.defaultBridge = bridge
	return nil
}

// Get returns a bridge by name
func (m *Manager) Get(name string) (Bridge, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.bridges[name]
	return b, ok
}

// Default returns the default bridge
func (m *Manager) Default() Bridge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultBridge
}

// DefaultManager is the global bridge manager
var DefaultManager = NewManager()

// Register is a convenience function using the default manager
func Register(name string, bridge Bridge) {
	DefaultManager.Register(name, bridge)
}

// Default is a convenience function using the default manager
func Default() Bridge {
	return DefaultManager.Default()
}
