package buddy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Buddy represents a buddy agent
type Buddy struct {
	mu       sync.RWMutex
	id       string
	name     string
	status   BuddyStatus
	tasks    []*BuddyTask
	skills   []string
	config   *BuddyConfig
}

// BuddyStatus represents the status of a buddy
type BuddyStatus string

const (
	StatusIdle       BuddyStatus = "idle"
	StatusWorking    BuddyStatus = "working"
	StatusThinking   BuddyStatus = "thinking"
	StatusListening  BuddyStatus = "listening"
	StatusSpeaking   BuddyStatus = "speaking"
)

// BuddyTask represents a task for a buddy
type BuddyTask struct {
	ID          string
	Description string
	Status      string
	Result      any
	CreatedAt   time.Time
	CompletedAt time.Time
}

// BuddyConfig represents buddy configuration
type BuddyConfig struct {
	Name        string
	Model       string
	VoiceEnabled bool
	Personality string
}

// NewBuddy creates a new buddy
func NewBuddy(id string, config *BuddyConfig) *Buddy {
	return &Buddy{
		id:     id,
		name:   config.Name,
		status: StatusIdle,
		tasks:  make([]*BuddyTask, 0),
		skills: make([]string, 0),
		config: config,
	}
}

// ID returns the buddy's ID
func (b *Buddy) ID() string {
	return b.id
}

// Name returns the buddy's name
func (b *Buddy) Name() string {
	return b.name
}

// Status returns the buddy's status
func (b *Buddy) Status() BuddyStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// SetStatus sets the buddy's status
func (b *Buddy) SetStatus(status BuddyStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

// AddTask adds a task to the buddy
func (b *Buddy) AddTask(task *BuddyTask) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tasks = append(b.tasks, task)
}

// CompleteTask marks a task as complete
func (b *Buddy) CompleteTask(taskID string, result any) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, task := range b.tasks {
		if task.ID == taskID {
			task.Status = "completed"
			task.Result = result
			task.CompletedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", taskID)
}

// AddSkill adds a skill to the buddy
func (b *Buddy) AddSkill(skill string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.skills = append(b.skills, skill)
}

// HasSkill checks if the buddy has a skill
func (b *Buddy) HasSkill(skill string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.skills {
		if s == skill {
			return true
		}
	}
	return false
}

// Listen listens for user input
func (b *Buddy) Listen(ctx context.Context) error {
	b.SetStatus(StatusListening)
	defer b.SetStatus(StatusIdle)

	// Placeholder for actual listening logic
	return nil
}

// Speak speaks the given text
func (b *Buddy) Speak(ctx context.Context, text string) error {
	b.SetStatus(StatusSpeaking)
	defer b.SetStatus(StatusIdle)

	// Placeholder for actual speaking logic
	fmt.Println(text)
	return nil
}

// Think processes information
func (b *Buddy) Think(ctx context.Context, input string) (string, error) {
	b.SetStatus(StatusThinking)
	defer b.SetStatus(StatusIdle)

	// Placeholder for actual thinking logic
	return "", nil
}

// Manager manages buddy instances
type Manager struct {
	mu      sync.RWMutex
	buddies map[string]*Buddy
}

// NewManager creates a new buddy manager
func NewManager() *Manager {
	return &Manager{
		buddies: make(map[string]*Buddy),
	}
}

// CreateBuddy creates a new buddy
func (m *Manager) CreateBuddy(id string, config *BuddyConfig) *Buddy {
	m.mu.Lock()
	defer m.mu.Unlock()

	buddy := NewBuddy(id, config)
	m.buddies[id] = buddy
	return buddy
}

// GetBuddy returns a buddy by ID
func (m *Manager) GetBuddy(id string) (*Buddy, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.buddies[id]
	return b, ok
}

// RemoveBuddy removes a buddy
func (m *Manager) RemoveBuddy(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.buddies, id)
}

// ListBuddies returns all buddies
func (m *Manager) ListBuddies() []*Buddy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	buddies := make([]*Buddy, 0, len(m.buddies))
	for _, b := range m.buddies {
		buddies = append(buddies, b)
	}
	return buddies
}

// DefaultManager is the global buddy manager
var DefaultManager = NewManager()

// CreateBuddy is a convenience function using the default manager
func CreateBuddy(id string, config *BuddyConfig) *Buddy {
	return DefaultManager.CreateBuddy(id, config)
}

// GetBuddy is a convenience function using the default manager
func GetBuddy(id string) (*Buddy, bool) {
	return DefaultManager.GetBuddy(id)
}
