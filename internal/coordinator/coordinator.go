package coordinator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Coordinator manages multi-agent coordination
type Coordinator struct {
	mu       sync.RWMutex
	agents   map[string]*Agent
	config   *CoordinatorConfig
	toolCall func(ctx context.Context, toolName string, args map[string]any) (any, error)
}

// NewCoordinator creates a new coordinator
func NewCoordinator(config *CoordinatorConfig) *Coordinator {
	return &Coordinator{
		agents: make(map[string]*Agent),
		config: config,
	}
}

// SetToolCall sets the tool call function for spawning agents
func (c *Coordinator) SetToolCall(fn func(ctx context.Context, toolName string, args map[string]any) (any, error)) {
	c.toolCall = fn
}

// SpawnAgent spawns a new worker agent
func (c *Coordinator) SpawnAgent(ctx context.Context, name string, prompt string, tools []string) (*Agent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate agent ID
	agentID := uuid.New().String()

	agent := &Agent{
		ID:        agentID,
		Name:      name,
		State:     AgentStatePending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tools:     tools,
	}

	// Store agent
	c.agents[agentID] = agent

	// If we have a tool call function, use it to spawn the agent
	if c.toolCall != nil {
		// Build the agent spawn arguments
		args := map[string]any{
			"name":    name,
			"prompt": prompt,
		}
		if len(tools) > 0 {
			args["tools"] = tools
		}

		// Call the AGENT_TOOL to spawn
		_, err := c.toolCall(ctx, "Agent", args)
		if err != nil {
			agent.State = AgentStateFailed
			agent.Error = err.Error()
			return agent, fmt.Errorf("failed to spawn agent: %w", err)
		}

		agent.State = AgentStateRunning
	}

	return agent, nil
}

// SendMessage sends a message to an agent
func (c *Coordinator) SendMessage(ctx context.Context, agentID string, message string) error {
	c.mu.RLock()
	_, ok := c.agents[agentID]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if c.toolCall == nil {
		return fmt.Errorf("tool call not configured")
	}

	args := map[string]any{
		"to":     agentID,
		"message": message,
	}

	_, err := c.toolCall(ctx, "SendMessage", args)
	return err
}

// StopAgent stops a running agent
func (c *Coordinator) StopAgent(agentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, ok := c.agents[agentID]
	if !ok {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if c.toolCall != nil {
		args := map[string]any{
			"task_id": agentID,
		}
		_, err := c.toolCall(context.Background(), "TaskStop", args)
		if err != nil {
			return fmt.Errorf("failed to stop agent: %w", err)
		}
	}

	agent.State = AgentStateStopped
	agent.UpdatedAt = time.Now()
	return nil
}

// HandleNotification handles a task notification from an agent
func (c *Coordinator) HandleNotification(notification *TaskNotification) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, ok := c.agents[notification.TaskID]
	if !ok {
		// Unknown agent, might be a stale notification
		return nil
	}

	switch notification.Status {
	case "completed":
		agent.State = AgentStateComplete
		if notification.Output != "" {
			agent.Result = &AgentResult{
				Output: notification.Output,
			}
		}
	case "failed":
		agent.State = AgentStateFailed
		agent.Error = notification.Error
	case "stopped":
		agent.State = AgentStateStopped
	default:
		return fmt.Errorf("unknown status: %s", notification.Status)
	}

	agent.UpdatedAt = time.Now()
	return nil
}

// GetAgent returns an agent by ID
func (c *Coordinator) GetAgent(id string) (*Agent, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	agent, ok := c.agents[id]
	return agent, ok
}

// ListAgents returns all agents
func (c *Coordinator) ListAgents() []*Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	agents := make([]*Agent, 0, len(c.agents))
	for _, agent := range c.agents {
		agents = append(agents, agent)
	}
	return agents
}

// RunningAgents returns all running agents
func (c *Coordinator) RunningAgents() []*Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	agents := make([]*Agent, 0)
	for _, agent := range c.agents {
		if agent.State == AgentStateRunning {
			agents = append(agents, agent)
		}
	}
	return agents
}

// CompleteAgents returns all complete agents
func (c *Coordinator) CompleteAgents() []*Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	agents := make([]*Agent, 0)
	for _, agent := range c.agents {
		if agent.State == AgentStateComplete {
			agents = append(agents, agent)
		}
	}
	return agents
}

// RemoveAgent removes an agent
func (c *Coordinator) RemoveAgent(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.agents, id)
}

// GetSystemPrompt returns the coordinator system prompt
func GetSystemPrompt(workerTools []string, mcpServers []string) string {
	var toolsList string
	if len(workerTools) > 0 {
		toolsList = strings.Join(workerTools, ", ")
	} else {
		toolsList = "Bash, Read, Edit, Write, Grep, Glob, and MCP tools"
	}

	var mcpInfo string
	if len(mcpServers) > 0 {
		mcpInfo = fmt.Sprintf("\n\nWorkers also have access to MCP tools from connected MCP servers: %s", strings.Join(mcpServers, ", "))
	}

	return fmt.Sprintf(`You are Claude Code, an AI assistant that orchestrates software engineering tasks across multiple workers.

## 1. Your Role

You are a **coordinator**. Your job is to:
- Help the user achieve their goal
- Direct workers to research, implement and verify code changes
- Synthesize results and communicate with the user
- Answer questions directly when possible — don't delegate work that you can handle without tools

Every message you send is to the user. Worker results and system notifications are internal signals, not conversation partners — never thank or acknowledge them. Summarize new information for the user as it arrives.

## 2. Your Tools

You have access to these special coordination tools:
- **Agent** - Spawn a new worker
- **SendMessage** - Continue an existing worker
- **TaskStop** - Stop a running worker

Workers spawned via the Agent tool have access to these tools: %s.%s

## 3. How to Coordinate

When calling Agent:
- Do not use one worker to check on another. Workers will notify you when they are done.
- Do not use workers to trivially report file contents or run commands. Give them higher-level tasks.
- After launching agents, briefly tell the user what you launched and end your response.
- Do not fabricate or predict agent results — results arrive as separate messages.

## 4. Task Notifications

Worker results arrive as task notifications. They contain:
- Task ID
- Status (completed, failed, stopped)
- Output (if completed)

Synthesize the information and present it to the user.`, toolsList, mcpInfo)
}
