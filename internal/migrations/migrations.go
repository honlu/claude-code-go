package migrations

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID        string
	Name      string
	Up        func(ctx context.Context) error
	Down      func(ctx context.Context) error
	AppliedAt time.Time
}

// Manager manages migrations
type Manager struct {
	mu         sync.RWMutex
	migrations []*Migration
	applied    map[string]bool
}

// NewManager creates a new migration manager
func NewManager() *Manager {
	return &Manager{
		migrations: make([]*Migration, 0),
		applied:    make(map[string]bool),
	}
}

// Register registers a migration
func (m *Manager) Register(migration *Migration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.migrations = append(m.migrations, migration)
}

// Up applies all pending migrations
func (m *Manager) Up(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for _, migration := range m.migrations {
		if m.applied[migration.ID] {
			continue
		}

		if err := migration.Up(ctx); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		migration.AppliedAt = time.Now()
		m.applied[migration.ID] = true
		lastErr = nil
	}

	return lastErr
}

// Down rolls back the last applied migration
func (m *Manager) Down(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the last applied migration
	var last *Migration
	for _, migration := range m.migrations {
		if m.applied[migration.ID] {
			last = migration
		}
	}

	if last == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	if err := last.Down(ctx); err != nil {
		return fmt.Errorf("rollback of %s failed: %w", last.ID, err)
	}

	delete(m.applied, last.ID)
	return nil
}

// Status returns the status of all migrations
func (m *Manager) Status() []*MigrationStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make([]*MigrationStatus, len(m.migrations))
	for i, migration := range m.migrations {
		status[i] = &MigrationStatus{
			ID:        migration.ID,
			Name:      migration.Name,
			Applied:   m.applied[migration.ID],
			AppliedAt: migration.AppliedAt,
		}
	}
	return status
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	ID        string
	Name      string
	Applied   bool
	AppliedAt time.Time
}

// DefaultManager is the global migration manager
var DefaultManager = NewManager()

// Register registers a migration with the default manager
func Register(migration *Migration) {
	DefaultManager.Register(migration)
}

// Up applies all pending migrations
func Up(ctx context.Context) error {
	return DefaultManager.Up(ctx)
}

// Down rolls back the last migration
func Down(ctx context.Context) error {
	return DefaultManager.Down(ctx)
}

// Status returns the status of all migrations
func Status() []*MigrationStatus {
	return DefaultManager.Status()
}
