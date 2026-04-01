package assistant

import (
	"context"
	"fmt"
	"sync"

	"github.com/ai/claude-code/internal/query"
	"github.com/ai/claude-code/internal/tools"
)

// Assistant represents the main assistant
type Assistant struct {
	mu           sync.RWMutex
	name         string
	model        string
	queryEngine  *query.QueryEngine
	toolExecutor *tools.ToolExecutor
	history      []Message
}

// Message represents a message in the assistant
type Message struct {
	Role    string
	Content string
}

// New creates a new assistant
func New(name, model string) *Assistant {
	return &Assistant{
		name:    name,
		model:   model,
		history: make([]Message, 0),
	}
}

// SetQueryEngine sets the query engine
func (a *Assistant) SetQueryEngine(qe *query.QueryEngine) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.queryEngine = qe
}

// SetToolExecutor sets the tool executor
func (a *Assistant) SetToolExecutor(te *tools.ToolExecutor) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.toolExecutor = te
}

// Query sends a query to the assistant
func (a *Assistant) Query(ctx context.Context, input string) (*Response, error) {
	a.mu.Lock()
	a.history = append(a.history, Message{Role: "user", Content: input})
	a.mu.Unlock()

	if a.queryEngine != nil {
		if err := a.queryEngine.Query(ctx, input); err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		return &Response{Content: "Query processed"}, nil
	}

	return nil, fmt.Errorf("query engine not configured")
}

// QueryStream sends a query with streaming response
func (a *Assistant) QueryStream(ctx context.Context, input string, onChunk func(string) error) error {
	a.mu.Lock()
	a.history = append(a.history, Message{Role: "user", Content: input})
	a.mu.Unlock()

	if a.queryEngine != nil {
		return a.queryEngine.QueryStream(ctx, input, onChunk)
	}

	return fmt.Errorf("query engine not configured")
}

// GetHistory returns the message history
func (a *Assistant) GetHistory() []Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	history := make([]Message, len(a.history))
	copy(history, a.history)
	return history
}

// ClearHistory clears the message history
func (a *Assistant) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = make([]Message, 0)
}

// AddMessage adds a message to history
func (a *Assistant) AddMessage(role, content string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = append(a.history, Message{Role: role, Content: content})
}

// Response represents a response from the assistant
type Response struct {
	Content   string
	ToolCalls []ToolCall
	Error     error
}

// ToolCall represents a tool call from the assistant
type ToolCall struct {
	Name      string
	Arguments map[string]any
	Result    any
}

// Handler handles assistant requests
type Handler struct {
	mu         sync.RWMutex
	assistants map[string]*Assistant
}

// NewHandler creates a new assistant handler
func NewHandler() *Handler {
	return &Handler{
		assistants: make(map[string]*Assistant),
	}
}

// CreateAssistant creates a new assistant
func (h *Handler) CreateAssistant(id, name, model string) *Assistant {
	h.mu.Lock()
	defer h.mu.Unlock()

	a := New(name, model)
	h.assistants[id] = a
	return a
}

// GetAssistant returns an assistant by ID
func (h *Handler) GetAssistant(id string) (*Assistant, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	a, ok := h.assistants[id]
	return a, ok
}

// RemoveAssistant removes an assistant
func (h *Handler) RemoveAssistant(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.assistants, id)
}

// DefaultHandler is the global assistant handler
var DefaultHandler = NewHandler()

// CreateAssistant is a convenience function using the default handler
func CreateAssistant(id, name, model string) *Assistant {
	return DefaultHandler.CreateAssistant(id, name, model)
}

// GetAssistant is a convenience function using the default handler
func GetAssistant(id string) (*Assistant, bool) {
	return DefaultHandler.GetAssistant(id)
}

// Config represents assistant configuration
type Config struct {
	Name        string
	Model       string
	APIKey      string
	Tools       []string
	Permissions string
}

// NewAssistantFromConfig creates an assistant from config
func NewAssistantFromConfig(config *Config) (*Assistant, error) {
	a := New(config.Name, config.Model)

	if config.APIKey != "" {
		qe := query.NewQueryEngine(
			query.WithAPIKey(config.APIKey),
			query.WithModel(config.Model),
		)
		a.SetQueryEngine(qe)
	}

	return a, nil
}
