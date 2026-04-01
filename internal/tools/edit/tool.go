package edit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the edit tool
const ToolName = "Edit"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for EditTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"file_path": map[string]any{
			"type":        "string",
			"description": "The absolute path to the file to modify",
		},
		"old_string": map[string]any{
			"type":        "string",
			"description": "The text to replace",
		},
		"new_string": map[string]any{
			"type":        "string",
			"description": "The replacement text",
		},
		"replace_all": map[string]any{
			"type":        "boolean",
			"description": "Replace all occurrences (default: false)",
		},
	},
	"required": []string{"file_path", "old_string", "new_string"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"filePath":    map[string]any{"type": "string"},
		"oldString":   map[string]any{"type": "string"},
		"newString":   map[string]any{"type": "string"},
		"replacements": map[string]any{"type": "integer"},
		"gitDiff":    map[string]any{"type": "string"},
	},
}

// Tool implements the Edit tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new EditTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Make edits to a file by replacing text",
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

// IsReadOnly always returns false
func (t *Tool) IsReadOnly(_ map[string]any) bool {
	return false
}

// IsDestructive returns false (edit is not inherently destructive)
func (t *Tool) IsDestructive(_ map[string]any) bool {
	return false
}

// IsConcurrencySafe returns false
func (t *Tool) IsConcurrencySafe(_ map[string]any) bool {
	return false
}

// Call performs the edit
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return &tools.ToolResult{Error: fmt.Errorf("file_path is required")}, nil
	}

	oldString, ok := input["old_string"].(string)
	if !ok || oldString == "" {
		return &tools.ToolResult{Error: fmt.Errorf("old_string is required")}, nil
	}

	newString, ok := input["new_string"].(string)
	if !ok {
		return &tools.ToolResult{Error: fmt.Errorf("new_string is required")}, nil
	}

	replaceAll := false
	if ra, ok := input["replace_all"].(bool); ok {
		replaceAll = ra
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &tools.ToolResult{Error: fmt.Errorf("file does not exist: %s", filePath)}, nil
		}
		return &tools.ToolResult{Error: fmt.Errorf("failed to read file: %w", err)}, nil
	}

	originalContent := string(content)
	var newContent string
	var replacements int

	if replaceAll {
		// Replace all occurrences
		newContent = strings.ReplaceAll(originalContent, oldString, newString)
		replacements = strings.Count(originalContent, oldString)
	} else {
		// Replace first occurrence only
		if idx := strings.Index(originalContent, oldString); idx >= 0 {
			newContent = originalContent[:idx] + newString + originalContent[idx+len(oldString):]
			replacements = 1
		} else {
			return &tools.ToolResult{Error: fmt.Errorf("old_string not found in file")}, nil
		}
	}

	// Generate diff
	diff := generateDiff(filePath, originalContent, newContent)

	// Write file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return &tools.ToolResult{Error: fmt.Errorf("failed to write file: %w", err)}, nil
	}

	return &tools.ToolResult{
		Data: map[string]any{
			"filePath":     filePath,
			"oldString":    oldString,
			"newString":    newString,
			"replacements": replacements,
			"gitDiff":      diff,
		},
	}, nil
}

// generateDiff creates a unified diff string
func generateDiff(filePath, oldContent, newContent string) string {
	// Simple unified diff format
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s\n", filePath))
	diff.WriteString(fmt.Sprintf("+++ %s\n", filePath))

	// Calculate line numbers (simplified - just show counts)
	oldCount := len(oldLines)
	newCount := len(newLines)

	diff.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", oldCount, newCount))

	// Show a simple representation of changes
	if oldCount <= newCount {
		for i := 0; i < oldCount && i < 5; i++ {
			diff.WriteString(fmt.Sprintf("-%s\n", oldLines[i]))
		}
		for i := 0; i < newCount && i < 5; i++ {
			diff.WriteString(fmt.Sprintf("+%s\n", newLines[i]))
		}
	} else {
		for i := 0; i < newCount && i < 5; i++ {
			diff.WriteString(fmt.Sprintf("-%s\n", oldLines[i]))
			diff.WriteString(fmt.Sprintf("+%s\n", newLines[i]))
		}
		if oldCount > 5 {
			diff.WriteString(fmt.Sprintf("... and %d more lines\n", oldCount-newCount))
		}
	}

	return diff.String()
}
