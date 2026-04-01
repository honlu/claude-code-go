package bash

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ai/claude-code/internal/tools"
)

// ToolName is the name of the bash tool
const ToolName = "Bash"

func init() {
	tools.RegisterTool(NewTool())
}

// InputSchema defines the input schema for BashTool
var InputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"command": map[string]any{
			"type":        "string",
			"description": "The shell command to execute",
		},
		"description": map[string]any{
			"type":        "string",
			"description": "Optional description of what this command does",
		},
		"timeout": map[string]any{
			"type":        "number",
			"description": "Timeout in milliseconds (default: 60000)",
		},
		"workingDirectory": map[string]any{
			"type":        "string",
			"description": "Working directory to execute the command in",
		},
	},
	"required": []string{"command"},
}

// OutputSchema defines the output schema
var OutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"stdout": map[string]any{"type": "string"},
		"stderr": map[string]any{"type": "string"},
		"exitCode": map[string]any{"type": "integer"},
		"interrupted": map[string]any{"type": "boolean"},
		"elapsedMs": map[string]any{"type": "integer"},
	},
}

// Tool implements the Bash tool
type Tool struct {
	*tools.BaseTool
}

// NewTool creates a new BashTool
func NewTool() *Tool {
	return &Tool{
		BaseTool: &tools.BaseTool{
			NameVal:        ToolName,
			DescriptionVal: "Execute a shell command in the terminal",
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

// IsReadOnly returns whether this command is read-only
func (t *Tool) IsReadOnly(input map[string]any) bool {
	cmd := t.getCommand(input)
	return isReadOnlyCommand(cmd)
}

// IsConcurrencySafe returns whether this tool can run concurrently
func (t *Tool) IsConcurrencySafe(_ map[string]any) bool {
	return false // Shell commands should not run concurrently by default
}

// getCommand extracts the command from input
func (t *Tool) getCommand(input map[string]any) string {
	if cmd, ok := input["command"].(string); ok {
		return cmd
	}
	return ""
}

// Call executes the shell command
func (t *Tool) Call(ctx context.Context, input map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
	// Extract command
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return &tools.ToolResult{Error: fmt.Errorf("command is required")}, nil
	}

	// Extract optional parameters
	timeout := int64(60000) // default 60 seconds
	if to, ok := input["timeout"].(float64); ok {
		timeout = int64(to)
	}

	workingDir := tc.WorkingDirectory
	if wd, ok := input["workingDirectory"].(string); ok && wd != "" {
		workingDir = wd
	}

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Prepare command
	var cmd *exec.Cmd
	var shell, flag string

	// Detect shell
	switch {
	case strings.Contains(command, "&&") || strings.Contains(command, "||"):
		shell, flag = "/bin/bash", "-c"
	case strings.Contains(command, "|"):
		shell, flag = "/bin/bash", "-c"
	default:
		shell, flag = "/bin/bash", "-c"
	}

	cmd = exec.CommandContext(cmdCtx, shell, flag, command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	// Build result
	result := &tools.ToolResult{
		Data: map[string]any{
			"stdout":    stdout.String(),
			"stderr":    stderr.String(),
			"exitCode":  0,
			"interrupted": false,
			"elapsedMs": elapsed.Milliseconds(),
		},
	}

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			result.Data.(map[string]any)["interrupted"] = true
			result.Data.(map[string]any)["stderr"] = fmt.Sprintf("Command timed out after %d ms\n%s", timeout, stderr.String())
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			result.Data.(map[string]any)["exitCode"] = exitErr.ExitCode()
		} else {
			result.Error = err
		}
	}

	return result, nil
}

// CheckPermissions checks if the command is allowed
func (t *Tool) CheckPermissions(ctx context.Context, input map[string]any, tc tools.ToolContext) (tools.PermissionResult, error) {
	command := t.getCommand(input)
	if command == "" {
		return tools.PermissionResult{Behavior: "deny", Message: "No command provided"}, nil
	}

	// Basic security checks
	if isDangerousCommand(command) {
		return tools.PermissionResult{
			Behavior: "deny",
			Message:  fmt.Sprintf("Command '%s' is not allowed for security reasons", truncateCommand(command)),
		}, nil
	}

	return tools.PermissionResult{Behavior: "allow"}, nil
}

// isReadOnlyCommand checks if a command is read-only
func isReadOnlyCommand(cmd string) bool {
	readOnlyPrefixes := []string{
		"ls", "cat", "head", "tail", "grep", "rg", "find", "stat", "file",
		"wc", "du", "tree", "echo", "pwd", "which", "whereis",
	}

	cmd = strings.TrimSpace(cmd)
	// Remove leading whitespace and possible command operators
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	baseCmd := strings.ToLower(parts[0])
	for _, prefix := range readOnlyPrefixes {
		if baseCmd == prefix || strings.HasPrefix(cmd, prefix+" ") || strings.HasPrefix(cmd, prefix+"\t") {
			return true
		}
	}
	return false
}

// isDangerousCommand checks for potentially dangerous commands
func isDangerousCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf /", "rm -rf /*", "mkfs", "dd if=",
		":(){:|:&};:", // fork bomb
	}
	cmd = strings.TrimSpace(cmd)
	for _, d := range dangerous {
		if cmd == d || strings.Contains(cmd, d) {
			return true
		}
	}
	return false
}

// truncateCommand truncates a command for display
func truncateCommand(cmd string) string {
	if len(cmd) > 50 {
		return cmd[:50] + "..."
	}
	return cmd
}

// BashSearchCommands is the set of commands considered as searches
var BashSearchCommands = map[string]bool{
	"find": true, "grep": true, "rg": true, "ag": true,
	"ack": true, "locate": true, "which": true, "whereis": true,
}

// BashReadCommands is the set of commands considered as reads
var BashReadCommands = map[string]bool{
	"cat": true, "head": true, "tail": true, "less": true, "more": true,
	"wc": true, "stat": true, "file": true, "strings": true,
	"jq": true, "awk": true, "cut": true, "sort": true, "uniq": true, "tr": true,
}

// BashListCommands is the set of commands that list directories
var BashListCommands = map[string]bool{
	"ls": true, "tree": true, "du": true,
}

// IsSearchOrReadCommand determines if a bash command is a search/read operation
func IsSearchOrReadCommand(command string) (isSearch, isRead, isList bool) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false, false, false
	}

	baseCmd := strings.ToLower(parts[0])
	isSearch = BashSearchCommands[baseCmd]
	isRead = BashReadCommands[baseCmd]
	isList = BashListCommands[baseCmd]
	return
}
