package write

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the write tool
const ToolName = "Write"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for WriteTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The absolute path to the file to create",
		},
		"content": map[string]any{
			"type":        "string",
			"description": "The content to write to the file",
		},
	},
	"required": []string{"file_path", "content"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"filePath":   map[string]any{"type": "string"},
		"bytes":      map[string]any{"type": "integer"},
		"created":    map[string]any{"type": "boolean"},
	},
}

// Tool implements the Write tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new WriteTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Create or overwrite a file with content",
		},
	}
}

// InputSchema returns the input schema
func (t *Tool) InputSchema() any {
	return InputSchema
}

// OutputSchema returns the output schema
func (t *Tool) OutputSchema() any {
	return OutputSchema
}

// IsReadOnly returns false
func (t *Tool) IsReadOnly(_ map[string]any) bool {
	return false
}

// IsDestructive returns false (overwrite is not considered destructive)
func (t *Tool) IsDestructive(_ map[string]any) bool {
	return false
}

// IsConcurrencySafe returns false
func (t *Tool) IsConcurrencySafe(_ map[string]any) bool {
	return false
}

// Call creates or overwrites the file
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return &tools.ToolResult{Error: fmt.Errorf("file_path is required")}, nil
	}

	content, ok := input["content"].(string)
	if !ok && content != "" {
		// Content can be any type, convert to string
		content = fmt.Sprintf("%v", input["content"])
	}

	// Check if file already exists
	exists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &tools.ToolResult{Error: fmt.Errorf("failed to create directory: %w", err)}, nil
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return &tools.ToolResult{Error: fmt.Errorf("failed to write file: %w", err)}, nil
	}

	return &tools.ToolResult{
		Data: map[string]any{
			"filePath": filePath,
			"bytes":    len(content),
			"created":  !exists,
		},
	}, nil
}
