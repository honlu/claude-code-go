package tools

import (
	"fmt"
	"sync"
)

// Registry holds all registered tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Names returns the names of all registered tools
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetOrError retrieves a tool or returns an error
func (r *Registry) GetOrError(name string) (Tool, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetRegistry returns the global tool registry
func GetRegistry() *Registry {
	return globalRegistry
}

// RegisterTool registers a tool with the global registry
func RegisterTool(tool Tool) {
	globalRegistry.Register(tool)
}

// GetTool retrieves a tool from the global registry
func GetTool(name string) (Tool, bool) {
	return globalRegistry.Get(name)
}

// ListTools returns all tools from the global registry
func ListTools() []Tool {
	return globalRegistry.List()
}
