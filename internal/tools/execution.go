package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ToolExecutor executes tools with permission checking and progress reporting
type ToolExecutor struct {
	registry   *Registry
	timeout    time.Duration
	semaphore chan struct{}
	mu         sync.RWMutex
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(registry *Registry) *ToolExecutor {
	return &ToolExecutor{
		registry:   registry,
		timeout:    5 * time.Minute,
		semaphore:  make(chan struct{}, 10), // Max 10 concurrent tool calls
	}
}

// Execute executes a tool by name with the given arguments
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, args map[string]any, tc ToolContext) (*ToolResult, error) {
	return e.ExecuteWithProgress(ctx, toolName, args, tc, nil)
}

// ExecuteWithProgress executes a tool with progress reporting
func (e *ToolExecutor) ExecuteWithProgress(ctx context.Context, toolName string, args map[string]any, tc ToolContext, onProgress ToolProgress) (*ToolResult, error) {
	// Get tool from registry
	tool, ok := e.registry.Get(toolName)
	if !ok {
		return &ToolResult{Error: fmt.Errorf("tool not found: %s", toolName)}, nil
	}

	// Check concurrency
	if !tool.IsConcurrencySafe(args) {
		select {
		case e.semaphore <- struct{}{}:
			defer func() { <-e.semaphore }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Apply timeout to context
	if tc.AbortSignal == nil {
		abortChan := make(chan struct{})
		defer close(abortChan)
		tc.AbortSignal = abortChan
	}

	// Create progress callback if provided
	var progressCallback ToolProgress
	if onProgress != nil {
		progressCallback = onProgress
	}

	// Check permissions
	permResult, err := tool.CheckPermissions(ctx, args, tc)
	if err != nil {
		return &ToolResult{Error: fmt.Errorf("permission check failed: %w", err)}, nil
	}

	if permResult.Behavior == "deny" {
		return &ToolResult{Error: fmt.Errorf("permission denied: %s", permResult.Message)}, nil
	}

	// Apply updated input from permission check
	if permResult.UpdatedInput != nil {
		args = permResult.UpdatedInput
	}

	// Create tool context with progress
	execCtx := ToolContext{
		WorkingDirectory: tc.WorkingDirectory,
		AbortSignal:      tc.AbortSignal,
		OnProgress:       progressCallback,
		CanUseTool:       tc.CanUseTool,
	}

	// Execute tool with timeout
	resultCh := make(chan *ToolResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := tool.Call(ctx, args, execCtx)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return &ToolResult{Error: err}, nil
	case result := <-resultCh:
		return result, nil
	}
}

// ValidateInput validates tool input against schema
func (e *ToolExecutor) ValidateInput(tool Tool, args map[string]any) error {
	// Basic validation: check required fields from schema
	schema := tool.InputSchema()
	if schema == nil {
		return nil
	}

	// If schema is a map, validate it as JSON Schema-like
	if schemaMap, ok := schema.(map[string]any); ok {
		if props, ok := schemaMap["properties"].(map[string]any); ok {
			// Check required fields
			if required, ok := schemaMap["required"].([]any); ok {
				for _, req := range required {
					if reqStr, ok := req.(string); ok {
						if _, exists := args[reqStr]; !exists {
							return fmt.Errorf("missing required field: %s", reqStr)
						}
					}
				}
			}

			// Validate field types
			for key, value := range args {
				if prop, ok := props[key].(map[string]any); ok {
					if expectedType, ok := prop["type"].(string); ok {
						if err := validateType(value, expectedType); err != nil {
							return fmt.Errorf("field %s: %w", key, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// validateType checks if a value matches the expected JSON Schema type
func validateType(value any, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number", "integer":
		switch value.(type) {
		case float64, int, int64:
			// Numeric types are valid
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// SetTimeout sets the default timeout for tool execution
func (e *ToolExecutor) SetTimeout(timeout time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timeout = timeout
}

// GetTimeout gets the default timeout
func (e *ToolExecutor) GetTimeout() time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.timeout
}

// ExecuteMultiple executes multiple tools in parallel
func (e *ToolExecutor) ExecuteMultiple(ctx context.Context, calls []ToolCall, tc ToolContext) ([]*ToolResult, error) {
	results := make([]*ToolResult, len(calls))
	errors := make([]error, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		// Only execute concurrently safe tools in parallel
		tool, ok := e.registry.Get(call.ToolName)
		if !ok {
			results[i] = &ToolResult{Error: fmt.Errorf("tool not found: %s", call.ToolName)}
			continue
		}

		if tool.IsConcurrencySafe(call.Args) {
			wg.Add(1)
			go func(idx int, c ToolCall) {
				defer wg.Done()
				results[idx], errors[idx] = e.Execute(ctx, c.ToolName, c.Args, tc)
			}(i, call)
		} else {
			results[i], errors[i] = e.Execute(ctx, call.ToolName, call.Args, tc)
		}
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// ToolCall represents a tool call request
type ToolCall struct {
	ToolName string
	Args     map[string]any
}

// NewToolCall creates a new tool call
func NewToolCall(toolName string, args map[string]any) ToolCall {
	return ToolCall{
		ToolName: toolName,
		Args:     args,
	}
}
