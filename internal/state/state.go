package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State represents the application state
type State struct {
	mu            sync.RWMutex
	data          map[string]any
	filePath      string
	dirty         bool
	lastPersisted time.Time
}

// New creates a new state
func New(filePath string) (*State, error) {
	s := &State{
		data:     make(map[string]any),
		filePath: filePath,
	}

	// Load existing state if file exists
	if filePath != "" {
		if err := s.Load(); err != nil {
			// Ignore load errors, start fresh
		}
	}

	return s, nil
}

// Get returns a value from state
func (s *State) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set sets a value in state
func (s *State) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.dirty = true
}

// Delete deletes a value from state
func (s *State) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.dirty = true
}

// Keys returns all keys in state
func (s *State) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Clear clears all state
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
	s.dirty = true
}

// Save saves state to file
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return err
	}

	s.dirty = false
	s.lastPersisted = time.Now()
	return nil
}

// Load loads state from file
func (s *State) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	s.data = make(map[string]any)
	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}

	s.dirty = false
	return nil
}

// IsDirty returns whether the state has unsaved changes
func (s *State) IsDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// LastPersisted returns the last time state was persisted
func (s *State) LastPersisted() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastPersisted
}

// AutoSave starts a goroutine that auto-saves the state
func (s *State) AutoSave(interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if s.IsDirty() {
					s.Save()
				}
			case <-stopCh:
				s.Save()
				return
			}
		}
	}()
}

// History represents a history entry
type HistoryEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	Success   bool      `json:"success"`
}

// History manages command history
type History struct {
	mu      sync.RWMutex
	entries []HistoryEntry
	maxSize int
}

// NewHistory creates a new history
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &History{
		entries: make([]HistoryEntry, 0),
		maxSize: maxSize,
	}
}

// Add adds an entry to history
func (h *History) Add(entry HistoryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, entry)

	// Trim if over max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
}

// Entries returns all history entries
func (h *History) Entries() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]HistoryEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// Search searches history for entries matching a query
func (h *History) Search(query string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []HistoryEntry
	for _, entry := range h.entries {
		if contains(entry.Command, query) || contains(entry.Input, query) {
			results = append(results, entry)
		}
	}
	return results
}

// Clear clears history
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = h.entries[:0]
}

// contains is a simple substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
