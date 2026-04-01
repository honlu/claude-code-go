package coordinator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AgentManager manages agent lifecycle
type AgentManager struct {
	coordinator *Coordinator
	timeout    time.Duration
}

// NewAgentManager creates a new agent manager
func NewAgentManager(coordinator *Coordinator) *AgentManager {
	return &AgentManager{
		coordinator: coordinator,
		timeout:    30 * time.Minute, // Default timeout
	}
}

// SpawnOptions contains options for spawning an agent
type SpawnOptions struct {
	Name         string
	Prompt       string
	Tools        []string
	Model        string
	ParentID     string
	Timeout      time.Duration
}

// Spawn spawns a new agent with options
func (m *AgentManager) Spawn(ctx context.Context, opts *SpawnOptions) (*Agent, error) {
	// Apply defaults
	if opts.Name == "" {
		opts.Name = "worker"
	}
	if opts.Timeout == 0 {
		opts.Timeout = m.timeout
	}
	if len(opts.Tools) == 0 {
		opts.Tools = getDefaultWorkerTools()
	}

	// Spawn via coordinator
	agent, err := m.coordinator.SpawnAgent(ctx, opts.Name, opts.Prompt, opts.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	// If there's a timeout, set up a cancel context
	if opts.Timeout > 0 {
		go m.monitorAgentTimeout(agent.ID, opts.Timeout)
	}

	return agent, nil
}

// monitorAgentTimeout monitors an agent and stops it after timeout
func (m *AgentManager) monitorAgentTimeout(agentID string, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		m.coordinator.StopAgent(agentID)
	}
}

// SendMessageTo sends a message to an agent
func (m *AgentManager) SendMessageTo(ctx context.Context, agentID string, message string) error {
	return m.coordinator.SendMessage(ctx, agentID, message)
}

// Stop stops an agent
func (m *AgentManager) Stop(agentID string) error {
	return m.coordinator.StopAgent(agentID)
}

// Get returns an agent by ID
func (m *AgentManager) Get(agentID string) (*Agent, bool) {
	return m.coordinator.GetAgent(agentID)
}

// List returns all agents
func (m *AgentManager) List() []*Agent {
	return m.coordinator.ListAgents()
}

// Running returns all running agents
func (m *AgentManager) Running() []*Agent {
	return m.coordinator.RunningAgents()
}

// Complete returns all complete agents
func (m *AgentManager) Complete() []*Agent {
	return m.coordinator.CompleteAgents()
}

// WaitForCompletion waits for an agent to complete
func (m *AgentManager) WaitForCompletion(ctx context.Context, agentID string) (*Agent, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			agent, ok := m.coordinator.GetAgent(agentID)
			if !ok {
				return nil, fmt.Errorf("agent not found: %s", agentID)
			}
			if agent.State == AgentStateComplete || agent.State == AgentStateFailed || agent.State == AgentStateStopped {
				return agent, nil
			}
		}
	}
}

// WaitForAll waits for all agents to complete
func (m *AgentManager) WaitForAll(ctx context.Context) ([]*Agent, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			running := m.coordinator.RunningAgents()
			if len(running) == 0 {
				return m.coordinator.ListAgents(), nil
			}
		}
	}
}

// ParseTaskNotification parses a task notification from XML-like format
func ParseTaskNotification(content string) (*TaskNotification, error) {
	notification := &TaskNotification{}

	// Extract task-id
	if idx := strings.Index(content, "<task-id>"); idx >= 0 {
		start := idx + len("<task-id>")
		end := strings.Index(content, "</task-id>")
		if end > start {
			notification.TaskID = content[start:end]
		}
	}

	// Extract status
	if idx := strings.Index(content, "<status>"); idx >= 0 {
		start := idx + len("<status>")
		end := strings.Index(content, "</status>")
		if end > start {
			notification.Status = content[start:end]
		}
	}

	// Extract output
	if idx := strings.Index(content, "<output>"); idx >= 0 {
		start := idx + len("<output>")
		end := strings.Index(content, "</output>")
		if end > start {
			notification.Output = content[start:end]
		}
	}

	// Extract error
	if idx := strings.Index(content, "<error>"); idx >= 0 {
		start := idx + len("<error>")
		end := strings.Index(content, "</error>")
		if end > start {
			notification.Error = content[start:end]
		}
	}

	if notification.TaskID == "" {
		return nil, fmt.Errorf("invalid task notification: missing task-id")
	}

	return notification, nil
}

// getDefaultWorkerTools returns the default tools for workers
func getDefaultWorkerTools() []string {
	return []string{
		"Bash",
		"Read",
		"Edit",
		"Write",
		"Grep",
		"Glob",
		"WebSearch",
	}
}

// GenerateAgentID generates a new unique agent ID
func GenerateAgentID() string {
	return uuid.New().String()
}
