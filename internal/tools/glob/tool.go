package glob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the glob tool
const ToolName = "Glob"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for GlobTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"pattern": map[string]any{
			"type":        "string",
			"description": "The glob pattern to match files",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "Directory to search in (default: current directory)",
		},
		"include_hidden": map[string]any{
			"type":        "boolean",
			"description": "Include hidden files (default: false)",
		},
		"max_results": map[string]any{
			"type":        "integer",
			"description": "Maximum number of results to return",
		},
	},
	"required": []string{"pattern"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"files":  map[string]any{"type": "array"},
		"count":  map[string]any{"type": "integer"},
	},
}

// Tool implements the Glob tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new GlobTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Find files matching a glob pattern",
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

// IsReadOnly returns true
func (t *Tool) IsReadOnly(_ map[string]any) bool {
	return true
}

// IsConcurrencySafe returns true
func (t *Tool) IsConcurrencySafe(_ map[string]any) bool {
	return true
}

// Call performs the glob search
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return &tools.ToolResult{Error: fmt.Errorf("pattern is required")}, nil
	}

	path := tc.WorkingDirectory
	if p, ok := input["path"].(string); ok && p != "" {
		path = p
	}

	includeHidden := false
	if h, ok := input["include_hidden"].(bool); ok {
		includeHidden = h
	}

	maxResults := 1000
	if mr, ok := input["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Expand pattern to absolute if needed
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(path, pattern)
	}

	// Find matches
	var matches []string
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden files if not including them
		if !includeHidden && strings.HasPrefix(filepath.Base(walkPath), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if matches pattern
		matched, err := filepath.Match(pattern, walkPath)
		if err != nil {
			return nil
		}
		if matched {
			matches = append(matches, walkPath)
			if len(matches) >= maxResults {
				return filepath.SkipAll
			}
		}

		return nil
	})

	result := &tools.ToolResult{
		Data: map[string]any{
			"files": matches,
			"count": len(matches),
		},
	}

	if err != nil && err != filepath.SkipAll && err != filepath.SkipDir {
		result.Error = err
	}

	return result, nil
}
