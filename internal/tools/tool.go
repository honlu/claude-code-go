package tools

import (
	"context"
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Data          any
	Error         error
	NewMessages   []Message // optional additional messages
	IsImage       bool
	PersistedFile string   // for large outputs
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// ToolProgress reports progress for long-running tool operations
type ToolProgress func(progress ToolProgressData)

// ToolProgressData contains progress information
type ToolProgressData interface{}

// BashProgress is progress data for BashTool
type BashProgress struct {
	Type             string `json:"type"`
	Output           string `json:"output"`
	FullOutput       string `json:"fullOutput"`
	ElapsedSeconds  int    `json:"elapsedTimeSeconds"`
	TotalLines       int    `json:"totalLines"`
	TotalBytes       int    `json:"totalBytes"`
	TaskId           string `json:"taskId,omitempty"`
	TimeoutMs        int    `json:"timeoutMs,omitempty"`
}

// ToolContext provides execution context to tools
type ToolContext struct {
	WorkingDirectory string
	AbortSignal      <-chan struct{}
	AppState         any // *state.AppState - interface to avoid cycle
	OnProgress       ToolProgress
	CanUseTool       CanUseToolFunc
}

// CanUseToolFunc checks if a tool can be used
type CanUseToolFunc func(toolName string, input map[string]any) (PermissionResult, error)

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Behavior     string // "allow", "deny", "prompt"
	Message      string
	Suggestions  []PermissionSuggestion
	UpdatedInput map[string]any
}

// PermissionSuggestion suggests a permission rule to add
type PermissionSuggestion struct {
	Type       string // "addRules"
	Rule       PermissionRule
	Behavior   string
	Destination string // "localSettings"
}

// PermissionRule defines a permission rule
type PermissionRule struct {
	ToolName     string
	RuleContent  string
}

// Tool is the core interface all tools implement
type Tool interface {
	// Basic info
	Name() string
	Description(input map[string]any) string

	// Schema
	InputSchema() any  // JSON Schema representation
	OutputSchema() any

	// Execution
	Call(ctx context.Context, args map[string]any, tc ToolContext) (*ToolResult, error)

	// Classification
	IsConcurrencySafe(input map[string]any) bool
	IsReadOnly(input map[string]any) bool
	IsDestructive(input map[string]any) bool
	IsOpenWorld(input map[string]any) bool

	// Permissions
	CheckPermissions(ctx context.Context, input map[string]any, tc ToolContext) (PermissionResult, error)

	// UI (optional)
	UserFacingName() string
	RenderToolUseMessage(args map[string]any) string
	RenderToolResultMessage(result *ToolResult) string

	// MCP info
	IsMCP() bool
	MCPInfo() (serverName, toolName string)
}

// BaseTool provides default implementations for Tool interface
type BaseTool struct {
	NameVal        string
	DescriptionVal string
}

func (b *BaseTool) Name() string                    { return b.NameVal }
func (b *BaseTool) Description(_ map[string]any) string { return b.DescriptionVal }
func (b *BaseTool) OutputSchema() any                { return nil }
func (b *BaseTool) IsConcurrencySafe(_ map[string]any) bool { return false }
func (b *BaseTool) IsReadOnly(_ map[string]any) bool   { return false }
func (b *BaseTool) IsDestructive(_ map[string]any) bool { return false }
func (b *BaseTool) IsOpenWorld(_ map[string]any) bool   { return false }
func (b *BaseTool) UserFacingName() string            { return b.NameVal }
func (b *BaseTool) RenderToolUseMessage(_ map[string]any) string { return "" }
func (b *BaseTool) RenderToolResultMessage(_ *ToolResult) string { return "" }
func (b *BaseTool) IsMCP() bool                       { return false }
func (b *BaseTool) MCPInfo() (string, string)         { return "", "" }
func (b *BaseTool) CheckPermissions(_ context.Context, _ map[string]any, _ ToolContext) (PermissionResult, error) {
	return PermissionResult{Behavior: "allow"}, nil
}

// PermissionCheckers provides permission checking for tools
func DefaultCheckPermissions(_ context.Context, _ map[string]any, _ ToolContext) (PermissionResult, error) {
	return PermissionResult{Behavior: "allow"}, nil
}
