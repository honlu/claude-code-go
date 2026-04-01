package query

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ai/claude-code/internal/api"
	"github.com/ai/claude-code/internal/bootstrap"
	"github.com/ai/claude-code/internal/permissions"
	"github.com/ai/claude-code/internal/tools"
)

// QueryEngine is the main query processing engine
type QueryEngine struct {
	apiClient    *api.Client
	toolExecutor *tools.ToolExecutor
	permissions  *permissions.PermissionEngine
	messages     []api.Message
	model        string
	maxRetries   int
	maxIterations int
	mu           sync.RWMutex
}

// Option configures the query engine
type Option func(*QueryEngine)

// WithAPIKey sets the API key
func WithAPIKey(key string) Option {
	return func(q *QueryEngine) {
		if q.apiClient != nil {
			return
		}
		client, err := api.NewClient(api.ProviderDirect, api.WithAPIKey(key))
		if err == nil {
			q.apiClient = client
		}
	}
}

// WithModel sets the model
func WithModel(model string) Option {
	return func(q *QueryEngine) {
		q.model = model
	}
}

// WithMaxRetries sets max retries
func WithMaxRetries(n int) Option {
	return func(q *QueryEngine) {
		q.maxRetries = n
	}
}

// WithMaxIterations sets max tool call iterations
func WithMaxIterations(n int) Option {
	return func(q *QueryEngine) {
		q.maxIterations = n
	}
}

// WithPermissions sets the permission engine
func WithPermissions(p *permissions.PermissionEngine) Option {
	return func(q *QueryEngine) {
		q.permissions = p
	}
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(opts ...Option) *QueryEngine {
	q := &QueryEngine{
		messages:     []api.Message{},
		model:        "claude-sonnet-4-6",
		maxRetries:   3,
		maxIterations: 20,
		permissions:  permissions.DefaultEngine,
	}

	registry := tools.GetRegistry()
	q.toolExecutor = tools.NewToolExecutor(registry)

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// Query processes a user query with full tool loop
func (q *QueryEngine) Query(ctx context.Context, input string) error {
	// Add user message
	q.mu.Lock()
	q.messages = append(q.messages, api.Message{
		Role:    "user",
		Content: input,
	})
	q.mu.Unlock()

	iteration := 0

	for iteration < q.maxIterations {
		iteration++

		// Build request
		req := &api.MessageRequest{
			Model:     q.model,
			Messages:  q.messages,
			MaxTokens: 4096,
		}

		// Get tools and add to request
		toolList := tools.ListTools()
		if len(toolList) > 0 {
			req.Tools = q.buildToolsInput(toolList)
		}

		// Call API
		resp, err := q.apiClient.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("API call failed: %w", err)
		}

		// Process response
		hasToolCalls, err := q.processResponse(resp)
		if err != nil {
			return err
		}

		// If no tool calls, we're done
		if !hasToolCalls {
			return nil
		}
	}

	return fmt.Errorf("max iterations (%d) exceeded", q.maxIterations)
}

// QueryStream processes a user query with streaming
func (q *QueryEngine) QueryStream(ctx context.Context, input string, onChunk func(string) error) error {
	// Add user message
	q.mu.Lock()
	q.messages = append(q.messages, api.Message{
		Role:    "user",
		Content: input,
	})
	q.mu.Unlock()

	iteration := 0

	for iteration < q.maxIterations {
		iteration++

		// Build request
		req := &api.MessageRequest{
			Model:     q.model,
			Messages:  q.messages,
			MaxTokens: 4096,
			Stream:    true,
		}

		// Get tools and add to request
		toolList := tools.ListTools()
		if len(toolList) > 0 {
			req.Tools = q.buildToolsInput(toolList)
		}

		// Start streaming
		events, err := q.apiClient.CreateMessageStream(ctx, req)
		if err != nil {
			return fmt.Errorf("streaming failed: %w", err)
		}

		// Process stream events
		var toolCalls []api.ContentBlock
		var content strings.Builder
		var stopReason string

		for event := range events {
			switch event.Type {
			case "content_block_start":
				if block, ok := event.Data.(map[string]any); ok {
					if blockType, _ := block["type"].(string); blockType == "tool_use" {
						// This is a tool use block
						toolCalls = append(toolCalls, api.ContentBlock{
							Type: "tool_use",
						})
					}
				}
			case "content_block_delta":
				if delta, ok := event.Data.(map[string]any); ok {
					if deltaType, _ := delta["type"].(string); deltaType == "text_delta" {
						if text, _ := delta["text"].(string); text != "" {
							content.WriteString(text)
							if onChunk != nil {
								if err := onChunk(text); err != nil {
									return err
								}
							}
						}
					}
				}
			case "message_delta":
				if delta, ok := event.Data.(map[string]any); ok {
					if stop, ok := delta["stop_reason"].(string); ok {
						stopReason = stop
					}
				}
			case "message_stop":
				// Store the complete text message
				textContent := content.String()
				if textContent != "" {
					q.mu.Lock()
					q.messages = append(q.messages, api.Message{
						Role:    "assistant",
						Content: textContent,
					})
					q.mu.Unlock()
				}

				// Handle tool calls if present
				if len(toolCalls) > 0 {
					if err := q.executeToolCalls(toolCalls); err != nil {
						return err
					}
					// Continue to next iteration
					continue
				}
				return nil
			}
		}

		// If we got here without message_stop, check stop_reason
		if stopReason == "tool_use" {
			continue
		}
		break
	}

	return nil
}

// buildToolsInput builds the tools input for API request
func (q *QueryEngine) buildToolsInput(toolList []tools.Tool) []api.Tool {
	result := make([]api.Tool, 0, len(toolList))
	for _, tool := range toolList {
		result = append(result, api.Tool{
			Name:        tool.Name(),
			Description: tool.Description(nil),
			InputSchema: q.convertInputSchema(tool.InputSchema()),
		})
	}
	return result
}

// convertInputSchema converts a tool's input schema to API's InputSchema
func (q *QueryEngine) convertInputSchema(schema any) *api.InputSchema {
	if schema == nil {
		return nil
	}

	// If it's already a map, convert it
	if m, ok := schema.(map[string]any); ok {
		result := &api.InputSchema{}
		if t, ok := m["type"].(string); ok {
			result.Type = t
		}
		if props, ok := m["properties"].(map[string]any); ok {
			result.Properties = props
		}
		if req, ok := m["required"].([]any); ok {
			result.Required = make([]string, len(req))
			for i, v := range req {
				if s, ok := v.(string); ok {
					result.Required[i] = s
				}
			}
		}
		return result
	}

	return nil
}

// processResponse processes an API response
func (q *QueryEngine) processResponse(resp *api.MessageResponse) (bool, error) {
	if resp == nil {
		return false, fmt.Errorf("nil response")
	}

	var textContent string
	var hasToolCalls bool

	// Process content blocks
	for _, block := range resp.Content {
		if block.Type == "text" {
			textContent += block.Text
		} else if block.Type == "tool_use" {
			hasToolCalls = true
		}
	}

	// Add assistant message if there's text
	if textContent != "" {
		q.mu.Lock()
		q.messages = append(q.messages, api.Message{
			Role:    "assistant",
			Content: textContent,
		})
		q.mu.Unlock()
	}

	// Handle tool calls
	if hasToolCalls {
		if err := q.executeToolCalls(resp.Content); err != nil {
			return false, err
		}
	}

	return hasToolCalls, nil
}

// executeToolCalls executes tool calls from content blocks
func (q *QueryEngine) executeToolCalls(content []api.ContentBlock) error {
	ctx := context.Background()
	tc := tools.ToolContext{
		WorkingDirectory: bootstrap.GetCwd(),
	}

	var toolResults []api.Message

	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}

		// Extract tool call info
		toolName := block.Name
		toolArgs := block.Input

		// Execute tool
		result, err := q.toolExecutor.Execute(ctx, toolName, toolArgs, tc)
		if err != nil {
			toolResults = append(toolResults, api.Message{
				Role: "user",
				Content: fmt.Sprintf("[Error: %s]", err.Error()),
			})
			continue
		}

		// Format result
		var resultContent string
		if result.Error != nil {
			resultContent = fmt.Sprintf("[Error: %s]", result.Error.Error())
		} else if result.Data != nil {
			if s, ok := result.Data.(string); ok {
				resultContent = s
			} else {
				resultContent = fmt.Sprintf("%v", result.Data)
			}
		} else {
			resultContent = "[Tool completed successfully]"
		}

		// Add tool result as user message (required for continued conversation)
		toolResults = append(toolResults, api.Message{
			Role: "user",
			Content: resultContent,
		})
	}

	// Add all tool results to messages
	if len(toolResults) > 0 {
		q.mu.Lock()
		q.messages = append(q.messages, toolResults...)
		q.mu.Unlock()
	}

	return nil
}

// GetMessages returns the message history
func (q *QueryEngine) GetMessages() []api.Message {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.messages
}

// ClearMessages clears the message history
func (q *QueryEngine) ClearMessages() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = []api.Message{}
}

// SetModel sets the model
func (q *QueryEngine) SetModel(model string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.model = model
}

// GetModel gets the current model
func (q *QueryEngine) GetModel() string {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.model
}

// GetWorkingDirectory returns the current working directory
func GetWorkingDirectory() string {
	return bootstrap.GetCwd()
}
