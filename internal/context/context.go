package context

import (
	"context"
	"sync"
	"time"
)

// Context manages the application context
type Context struct {
	mu          sync.RWMutex
	values      map[string]any
	workingDir  string
	userHomeDir string
	startTime   time.Time
	cancelled   bool
}

// New creates a new context
func New() *Context {
	return &Context{
		values:    make(map[string]any),
		startTime: time.Now(),
	}
}

// WithValue adds a key-value pair to the context
func (c *Context) WithValue(key string, value any) *Context {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
	return c
}

// Value returns the value for a key
func (c *Context) Value(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	return v, ok
}

// SetWorkingDir sets the working directory
func (c *Context) SetWorkingDir(dir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workingDir = dir
}

// GetWorkingDir returns the working directory
func (c *Context) GetWorkingDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.workingDir
}

// SetUserHomeDir sets the user's home directory
func (c *Context) SetUserHomeDir(dir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userHomeDir = dir
}

// GetUserHomeDir returns the user's home directory
func (c *Context) GetUserHomeDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userHomeDir
}

// Uptime returns the uptime of the context
func (c *Context) Uptime() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.startTime)
}

// IsCancelled returns whether the context was cancelled
func (c *Context) IsCancelled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cancelled
}

// Cancel cancels the context
func (c *Context) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelled = true
}

// Keys returns all keys in the context
func (c *Context) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	return keys
}

// Clear clears all values from the context
func (c *Context) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = make(map[string]any)
}

// DefaultContext is the default global context
var DefaultContext = New()

// ContextKey is a custom context key type
type ContextKey string

// Context keys
const (
	KeyWorkingDir    ContextKey = "workingDir"
	KeyUserHomeDir   ContextKey = "userHomeDir"
	KeyProjectRoot   ContextKey = "projectRoot"
	KeyAPIKey        ContextKey = "apiKey"
	KeyModel         ContextKey = "model"
	KeyPermissionMode ContextKey = "permissionMode"
)

// WithContext returns a context with the given key-value pair
func WithContext(ctx context.Context, key ContextKey, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// FromContext returns the value for a key from a context
func FromContext(ctx context.Context, key ContextKey) (any, bool) {
	v := ctx.Value(key)
	return v, v != nil
}
