package constants

// Version is the current version
const Version = "0.1.0-dev"

// AppName is the application name
const AppName = "Claude Code"

// Description is the application description
const Description = "AI-powered coding assistant"

// DefaultModel is the default model
const DefaultModel = "claude-sonnet-4-6"

// MaxTokens is the default max tokens
const MaxTokens = 4096

// MaxConcurrentTools is the max concurrent tool calls
const MaxConcurrentTools = 10

// DefaultTimeout is the default timeout
const DefaultTimeoutSeconds = 300

// MaxToolLoopIterations is the max iterations for tool loops
const MaxToolLoopIterations = 20

// Environment variables
const (
	EnvAPIKey           = "ANTHROPIC_API_KEY"
	EnvAPIURL           = "ANTHROPIC_API_URL"
	EnvUseBedrock       = "CLAUDE_CODE_USE_BEDROCK"
	EnvUseFoundry       = "CLAUDE_CODE_USE_FOUNDRY"
	EnvUseVertex        = "CLAUDE_CODE_USE_VERTEX"
	EnvCustomHeaders    = "ANTHROPIC_CUSTOM_HEADERS"
	EnvPermissionMode   = "CLAUDE_CODE_PERMISSION_MODE"
	EnvDangerouslySkip  = "CLAUDE_CODE_DANGEROUSLY_SKIP_PERMISSIONS"
)

// File paths
const (
	ConfigDirName     = ".claude"
	ConfigFileName    = "config.json"
	StateFileName     = "state.json"
	HistoryFileName   = "history.json"
	MemoryDirName     = "memory"
	CacheDirName      = "cache"
)

// Permissions
const (
	PermissionAsk    = "ask"
	PermissionAccept  = "accept"
	PermissionDeny    = "deny"
	PermissionApprove = "approve"
)

// Tool names
const (
	ToolBash  = "Bash"
	ToolRead  = "Read"
	ToolEdit  = "Edit"
	ToolWrite = "Write"
	ToolGrep  = "Grep"
	ToolGlob  = "Glob"
)

// Error codes
const (
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeRateLimit      = "RATE_LIMIT"
	ErrCodeInvalidRequest = "INVALID_REQUEST"
	ErrCodeInternalError  = "INTERNAL_ERROR"
)

// Message types
const (
	MsgTypeText     = "text"
	MsgTypeToolUse  = "tool_use"
	MsgTypeToolResult = "tool_result"
)

// Stream event types
const (
	EventContentBlockStart = "content_block_start"
	EventContentBlockDelta = "content_block_delta"
	EventContentBlockStop  = "content_block_stop"
	EventMessageStart      = "message_start"
	EventMessageDelta      = "message_delta"
	EventMessageStop      = "message_stop"
)

// Stop reasons
const (
	StopReasonEndTurn   = "end_turn"
	StopReasonToolUse   = "tool_use"
	StopReasonMaxTokens = "max_tokens"
)
