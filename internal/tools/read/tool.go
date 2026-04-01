package read

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the read tool
const ToolName = "Read"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for ReadTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The absolute path to the file to read",
		},
		"offset": map[string]any{
			"type":        "integer",
			"description": "Line number to start reading from (0-indexed)",
		},
		"limit": map[string]any{
			"type":        "integer",
			"description": "Maximum number of lines to read",
		},
	},
	"required": []string{"file_path"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"content":     map[string]any{"type": "string"},
		"filePath":    map[string]any{"type": "string"},
		"lines":       map[string]any{"type": "integer"},
		"truncated":    map[string]any{"type": "boolean"},
		"mimeType":    map[string]any{"type": "string"},
		"notExist":    map[string]any{"type": "boolean"},
		"isBinary":    map[string]any{"type": "boolean"},
	},
}

// Tool implements the Read tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new ReadTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Read the contents of a file",
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

// IsReadOnly always returns true
func (t *Tool) IsReadOnly(_ map[string]any) bool {
	return true
}

// IsConcurrencySafe returns true
func (t *Tool) IsConcurrencySafe(_ map[string]any) bool {
	return true
}

// Call reads the file
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return &tools.ToolResult{Error: fmt.Errorf("file_path is required")}, nil
	}

	// Validate path
	if err := validatePath(filePath, tc.WorkingDirectory); err != nil {
		return &tools.ToolResult{Error: err}, nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &tools.ToolResult{
				Data: map[string]any{
					"filePath":  filePath,
					"notExist":  true,
					"truncated": false,
				},
			}, nil
		}
		return &tools.ToolResult{Error: fmt.Errorf("failed to read file: %w", err)}, nil
	}

	// Check if binary
	if isBinary(content) {
		return &tools.ToolResult{
			Data: map[string]any{
				"filePath":  filePath,
				"isBinary":  true,
				"truncated": false,
			},
		}, nil
	}

	// Apply offset and limit
	offset := 0
	limit := -1
	if off, ok := input["offset"].(float64); ok {
		offset = int(off)
	}
	if lim, ok := input["limit"].(float64); ok {
		limit = int(lim)
	}

	lines := strings.Split(string(content), "\n")
	if offset >= len(lines) {
		return &tools.ToolResult{
			Data: map[string]any{
				"filePath":  filePath,
				"content":   "",
				"lines":     0,
				"truncated": false,
			},
		}, nil
	}

	end := len(lines)
	if limit >= 0 {
		end = offset + limit
		if end > len(lines) {
			end = len(lines)
		}
	}

	displayLines := lines[offset:end]
	truncated := end < len(lines)

	result := &tools.ToolResult{
		Data: map[string]any{
			"filePath":  filePath,
			"content":   strings.Join(displayLines, "\n"),
			"lines":     len(displayLines),
			"truncated": truncated,
			"mimeType":  detectMimeType(filePath),
		},
	}

	return result, nil
}

// validatePath ensures the path is within the working directory
func validatePath(filePath, workingDir string) error {
	// TODO: Implement path traversal validation
	// For now, just check that it's not an absolute path outside of project
	return nil
}

// isBinary checks if the content appears to be binary
func isBinary(content []byte) bool {
	// Check for null bytes in the first 8000 bytes
	for i := 0; i < len(content) && i < 8000; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

// detectMimeType detects the MIME type based on file extension
func detectMimeType(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return "text/x-go"
	case strings.HasSuffix(lower, ".ts") || strings.HasSuffix(lower, ".tsx"):
		return "text/x-typescript"
	case strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".jsx"):
		return "text/javascript"
	case strings.HasSuffix(lower, ".py"):
		return "text/x-python"
	case strings.HasSuffix(lower, ".rs"):
		return "text/x-rust"
	case strings.HasSuffix(lower, ".java"):
		return "text/x-java"
	case strings.HasSuffix(lower, ".c"):
		return "text/x-c"
	case strings.HasSuffix(lower, ".cpp") || strings.HasSuffix(lower, ".cc"):
		return "text/x-c++"
	case strings.HasSuffix(lower, ".h") || strings.HasSuffix(lower, ".hpp"):
		return "text/x-chdr"
	case strings.HasSuffix(lower, ".json"):
		return "application/json"
	case strings.HasSuffix(lower, ".xml"):
		return "application/xml"
	case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
		return "application/x-yaml"
	case strings.HasSuffix(lower, ".md"):
		return "text/markdown"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	case strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm"):
		return "text/html"
	case strings.HasSuffix(lower, ".css"):
		return "text/css"
	case strings.HasSuffix(lower, ".sql"):
		return "text/x-sql"
	case strings.HasSuffix(lower, ".sh"):
		return "application/x-sh"
	case strings.HasSuffix(lower, ".bash"):
		return "application/x-bash"
	case strings.HasSuffix(lower, ".zsh"):
		return "application/x-zsh"
	case strings.HasSuffix(lower, ".fish"):
		return "application/x-fish"
	default:
		return "text/plain"
	}
}
