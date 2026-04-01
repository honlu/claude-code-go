package tasks

import (
	"context"
	"sync"
	"time"
)

// Status represents the status of a task
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Task represents a task to be executed
type Task struct {
	ID          string
	Name        string
	Description string
	Status      Status
	Progress    int
	Result      any
	Error       error
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Metadata    map[string]any
	mu          sync.RWMutex
}

// NewTask creates a new task
func NewTask(id, name, description string) *Task {
	return &Task{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      StatusPending,
		Progress:    0,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]any),
	}
}

// Start starts the task
func (t *Task) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = StatusRunning
	t.StartedAt = time.Now()
}

// Complete completes the task with a result
func (t *Task) Complete(result any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = StatusCompleted
	t.Result = result
	t.CompletedAt = time.Now()
}

// Fail fails the task with an error
func (t *Task) Fail(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = StatusFailed
	t.Error = err
	t.CompletedAt = time.Now()
}

// Cancel cancels the task
func (t *Task) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == StatusPending || t.Status == StatusRunning {
		t.Status = StatusCancelled
		t.CompletedAt = time.Now()
	}
}

// SetProgress sets the task progress (0-100)
func (t *Task) SetProgress(progress int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	t.Progress = progress
}

// SetMetadata sets a metadata value
func (t *Task) SetMetadata(key string, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (t *Task) GetMetadata(key string) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.Metadata[key]
	return v, ok
}

// Duration returns the task duration
func (t *Task) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.StartedAt.IsZero() {
		return 0
	}
	end := t.CompletedAt
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(t.StartedAt)
}

// IsCompleted returns true if the task is completed
func (t *Task) IsCompleted() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == StatusCompleted
}

// IsFailed returns true if the task has failed
func (t *Task) IsFailed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == StatusFailed
}

// IsRunning returns true if the task is running
func (t *Task) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == StatusRunning
}

// TaskManager manages tasks
type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

// CreateTask creates a new task
func (tm *TaskManager) CreateTask(id, name, description string) *Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task := NewTask(id, name, description)
	tm.tasks[id] = task
	return task
}

// GetTask gets a task by ID
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.tasks[id]
	return t, ok
}

// ListTasks lists all tasks
func (tm *TaskManager) ListTasks() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tasks := make([]*Task, 0, len(tm.tasks))
	for _, t := range tm.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// RemoveTask removes a task
func (tm *TaskManager) RemoveTask(id string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tasks, id)
}

// ClearCompleted clears all completed tasks
func (tm *TaskManager) ClearCompleted() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for id, task := range tm.tasks {
		if task.Status == StatusCompleted || task.Status == StatusFailed || task.Status == StatusCancelled {
			delete(tm.tasks, id)
		}
	}
}

// RunTask runs a task with the given function
func (tm *TaskManager) RunTask(ctx context.Context, id, name, description string, fn func(*Task) error) (*Task, error) {
	task := tm.CreateTask(id, name, description)
	task.Start()

	err := fn(task)
	if err != nil {
		task.Fail(err)
		return task, err
	}

	task.Complete(nil)
	return task, nil
}

// DefaultTaskManager is the global task manager
var DefaultTaskManager = NewTaskManager()

// CreateTask is a convenience function using the default manager
func CreateTask(id, name, description string) *Task {
	return DefaultTaskManager.CreateTask(id, name, description)
}

// GetTask is a convenience function using the default manager
func GetTask(id string) (*Task, bool) {
	return DefaultTaskManager.GetTask(id)
}
