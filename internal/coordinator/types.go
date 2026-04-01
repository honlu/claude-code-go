package coordinator

import "time"

// CoordinatorMode represents the mode of operation
type CoordinatorMode string

const (
	ModeNormal     CoordinatorMode = "normal"
	ModeCoordinator CoordinatorMode = "coordinator"
)

// AgentState represents the state of an agent
type AgentState string

const (
	AgentStatePending   AgentState = "pending"
	AgentStateRunning   AgentState = "running"
	AgentStateComplete  AgentState = "complete"
	AgentStateFailed    AgentState = "failed"
	AgentStateStopped   AgentState = "stopped"
)

// Agent represents a worker agent
type Agent struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	State     AgentState      `json:"state"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	ParentID  string          `json:"parentId,omitempty"`
	Tools     []string        `json:"tools,omitempty"`
	Model     string          `json:"model,omitempty"`
	Error     string          `json:"error,omitempty"`
	Result    *AgentResult    `json:"result,omitempty"`
}

// AgentResult represents the result of an agent
type AgentResult struct {
	Output    string   `json:"output"`
	Messages  int      `json:"messages"`
	ToolCalls int      `json:"toolCalls"`
	Duration  int64    `json:"durationMs"`
}

// TaskNotification represents a notification from an agent task
type TaskNotification struct {
	Type      string `json:"type"`
	TaskID    string `json:"taskId"`
	Status    string `json:"status"` // "completed", "failed", "stopped"
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	AgentID   string `json:"agentId,omitempty"`
}

// CoordinatorConfig represents coordinator configuration
type CoordinatorConfig struct {
	// Whether coordinator mode is enabled
	Enabled bool

	// Tools available to workers
	WorkerTools []string

	// MCP server names available to workers
	MCPServers []string

	// Scratchpad directory for cross-worker communication
	ScratchpadDir string

	// Simple mode (reduced tool set)
	SimpleMode bool
}

// IsCoordinatorModeEnabled returns whether coordinator mode is enabled
func IsCoordinatorModeEnabled() bool {
	return false // TODO: Read from environment
}

// GetCoordinatorConfig returns the coordinator configuration
func GetCoordinatorConfig() *CoordinatorConfig {
	// TODO: Load from environment and settings
	return &CoordinatorConfig{
		Enabled:     false,
		WorkerTools:  []string{},
		MCPServers:   []string{},
		SimpleMode:  false,
	}
}
