package grep

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the grep tool
const ToolName = "Grep"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for GrepTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"pattern": map[string]any{
			"type":        "string",
			"description": "The regular expression pattern to search for",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "File or directory to search in",
		},
		"glob": map[string]any{
			"type":        "string",
			"description": "Glob pattern to filter files",
		},
		"output_mode": map[string]any{
			"type":        "string",
			"enum":        []string{"content", "files_with_matches", "count"},
			"description": "Output format: content, files_with_matches, or count",
		},
		"-n": map[string]any{
			"type":        "boolean",
			"description": "Show line numbers",
		},
		"-C": map[string]any{
			"type":        "integer",
			"description": "Show N lines of context",
		},
		"-i": map[string]any{
			"type":        "boolean",
			"description": "Case insensitive search",
		},
		"head_limit": map[string]any{
			"type":        "integer",
			"description": "Limit number of results",
		},
	},
	"required": []string{"pattern"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"content":     map[string]any{"type": "string"},
		"filenames":   map[string]any{"type": "array"},
		"numFiles":    map[string]any{"type": "integer"},
		"numMatches":  map[string]any{"type": "integer"},
		"mode":        map[string]any{"type": "string"},
	},
}

// Tool implements the Grep tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new GrepTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Search file contents using regular expressions",
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

// Call performs the grep search
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return &tools.ToolResult{Error: fmt.Errorf("pattern is required")}, nil
	}

	path := "."
	if p, ok := input["path"].(string); ok && p != "" {
		path = p
	}

	outputMode := "content"
	if mode, ok := input["output_mode"].(string); ok && mode != "" {
		outputMode = mode
	}

	// Build rg command
	args := []string{"--json"}

	// Add flags
	if lineNum, ok := input["-n"]; ok && lineNum == true {
		args = append(args, "-n")
	}

	if caseInsensitive, ok := input["-i"]; ok && caseInsensitive == true {
		args = append(args, "-i")
	}

	if context, ok := input["-C"].(float64); ok {
		args = append(args, fmt.Sprintf("-C%d", int(context)))
	}

	if glob, ok := input["glob"].(string); ok && glob != "" {
		args = append(args, "--glob", glob)
	}

	if limit, ok := input["head_limit"].(float64); ok && limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	args = append(args, pattern)
	args = append(args, path)

	// Execute ripgrep
	cmd := exec.CommandContext(ctx, "rg", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &tools.ToolResult{
		Data: map[string]any{
			"content":   stdout.String(),
			"stderr":    stderr.String(),
			"mode":      outputMode,
			"numFiles":  0,
			"numMatches": 0,
		},
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means no matches found, which is not an error
			if exitErr.ExitCode() == 1 {
				return result, nil
			}
		}
		// For other errors, still return what we have
		result.Data.(map[string]any)["error"] = err.Error()
	}

	// Parse output to get filenames and counts
	content := stdout.String()
	if content != "" {
		lines := strings.Split(strings.TrimSpace(content), "\n")
		files := make(map[string]bool)
		for _, line := range lines {
			if line != "" {
				files[path] = true // simplified - actual implementation would parse JSON
			}
		}
		result.Data.(map[string]any)["numFiles"] = len(files)
		result.Data.(map[string]any)["numMatches"] = len(lines)
	}

	return result, nil
}
