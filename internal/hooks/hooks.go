package hooks

import (
	"context"
	"fmt"
	"sync"
)

// HookType represents the type of hook
type HookType string

const (
	HookBeforeStart     HookType = "before_start"
	HookAfterStart      HookType = "after_start"
	HookBeforeQuery     HookType = "before_query"
	HookAfterQuery      HookType = "after_query"
	HookBeforeToolCall  HookType = "before_tool_call"
	HookAfterToolCall   HookType = "after_tool_call"
	HookBeforeExit      HookType = "before_exit"
	HookAfterExit       HookType = "after_exit"
	HookOnError         HookType = "on_error"
	HookOnToolResult    HookType = "on_tool_result"
)

// Hook represents a hook function
type Hook func(ctx context.Context, event HookEvent) error

// HookEvent represents the event data passed to hooks
type HookEvent struct {
	Type      HookType
	Data      map[string]any
	Timestamp int64
}

// HookManager manages lifecycle hooks
type HookManager struct {
	mu    sync.RWMutex
	hooks map[HookType][]Hook
}

// NewHookManager creates a new hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Register registers a hook for a specific event type
func (hm *HookManager) Register(eventType HookType, hook Hook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hooks[eventType] = append(hm.hooks[eventType], hook)
}

// Unregister removes a hook
func (hm *HookManager) Unregister(eventType HookType, hook Hook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hooks := hm.hooks[eventType]
	for i, h := range hooks {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", hook) {
			hm.hooks[eventType] = append(hooks[:i], hooks[i+1:]...)
			return
		}
	}
}

// Emit emits an event to all registered hooks
func (hm *HookManager) Emit(ctx context.Context, eventType HookType, data map[string]any) error {
	hm.mu.RLock()
	hooks := make([]Hook, len(hm.hooks[eventType]))
	copy(hooks, hm.hooks[eventType])
	hm.mu.RUnlock()

	event := HookEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: nowUnix(),
	}

	for _, hook := range hooks {
		if err := hook(ctx, event); err != nil {
			return fmt.Errorf("hook error (%s): %w", eventType, err)
		}
	}
	return nil
}

// Clear clears all hooks
func (hm *HookManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hooks = make(map[HookType][]Hook)
}

// List returns all registered hooks
func (hm *HookManager) List() map[HookType]int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	result := make(map[HookType]int)
	for k, v := range hm.hooks {
		result[k] = len(v)
	}
	return result
}

// DefaultHookManager is the global hook manager
var DefaultHookManager = NewHookManager()

// Register is a convenience function using the default manager
func Register(eventType HookType, hook Hook) {
	DefaultHookManager.Register(eventType, hook)
}

// Emit is a convenience function using the default manager
func Emit(ctx context.Context, eventType HookType, data map[string]any) error {
	return DefaultHookManager.Emit(ctx, eventType, data)
}

// nowUnix returns the current Unix timestamp
func nowUnix() int64 {
	return int64(0) // Will be replaced with actual implementation
}
