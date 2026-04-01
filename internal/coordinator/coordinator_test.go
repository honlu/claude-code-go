package coordinator

import (
	"testing"
)

func TestCoordinatorTypes(t *testing.T) {
	if ModeNormal != "normal" {
		t.Errorf("ModeNormal = %v, want %v", ModeNormal, "normal")
	}
	if ModeCoordinator != "coordinator" {
		t.Errorf("ModeCoordinator = %v, want %v", ModeCoordinator, "coordinator")
	}

	if AgentStatePending != "pending" {
		t.Errorf("AgentStatePending = %v, want %v", AgentStatePending, "pending")
	}
	if AgentStateRunning != "running" {
		t.Errorf("AgentStateRunning = %v, want %v", AgentStateRunning, "running")
	}
	if AgentStateComplete != "complete" {
		t.Errorf("AgentStateComplete = %v, want %v", AgentStateComplete, "complete")
	}
}

func TestNewCoordinator(t *testing.T) {
	config := &CoordinatorConfig{Enabled: true}
	c := NewCoordinator(config)

	if c == nil {
		t.Fatal("NewCoordinator() returned nil")
	}
	if len(c.agents) != 0 {
		t.Errorf("len(c.agents) = %d, want 0", len(c.agents))
	}
}

func TestCoordinatorSpawnAgent(t *testing.T) {
	c := NewCoordinator(&CoordinatorConfig{Enabled: true})

	// Spawn without tool call (mock mode)
	agent, err := c.SpawnAgent(nil, "test-agent", "test prompt", []string{"Bash"})
	if err != nil {
		t.Errorf("SpawnAgent() error = %v", err)
	}
	if agent == nil {
		t.Fatal("SpawnAgent() returned nil agent")
	}
	if agent.Name != "test-agent" {
		t.Errorf("agent.Name = %v, want %v", agent.Name, "test-agent")
	}
	if agent.State != AgentStatePending {
		t.Errorf("agent.State = %v, want %v", agent.State, AgentStatePending)
	}
	if agent.ID == "" {
		t.Error("agent.ID is empty")
	}
}

func TestCoordinatorGetAgent(t *testing.T) {
	c := NewCoordinator(&CoordinatorConfig{Enabled: true})

	agent, _ := c.SpawnAgent(nil, "test-agent", "test prompt", nil)

	got, ok := c.GetAgent(agent.ID)
	if !ok {
		t.Error("GetAgent() returned false, want true")
	}
	if got.ID != agent.ID {
		t.Errorf("got.ID = %v, want %v", got.ID, agent.ID)
	}

	_, ok = c.GetAgent("non-existent")
	if ok {
		t.Error("GetAgent() for non-existent returned true, want false")
	}
}

func TestCoordinatorListAgents(t *testing.T) {
	c := NewCoordinator(&CoordinatorConfig{Enabled: true})

	c.SpawnAgent(nil, "agent1", "", nil)
	c.SpawnAgent(nil, "agent2", "", nil)

	agents := c.ListAgents()
	if len(agents) != 2 {
		t.Errorf("len(ListAgents()) = %d, want 2", len(agents))
	}
}

func TestCoordinatorRunningAgents(t *testing.T) {
	c := NewCoordinator(&CoordinatorConfig{Enabled: true})

	agent1, _ := c.SpawnAgent(nil, "agent1", "", nil)
	c.SpawnAgent(nil, "agent2", "", nil)

	// Without tool call, agents stay in pending state
	running := c.RunningAgents()
	if len(running) != 0 {
		t.Errorf("len(RunningAgents()) = %d, want 0 (pending state)", len(running))
	}

	_ = agent1 // use variable
}

func TestCoordinatorRemoveAgent(t *testing.T) {
	c := NewCoordinator(&CoordinatorConfig{Enabled: true})

	agent, _ := c.SpawnAgent(nil, "test-agent", "", nil)
	c.RemoveAgent(agent.ID)

	if len(c.agents) != 0 {
		t.Errorf("len(c.agents) = %d, want 0 after RemoveAgent", len(c.agents))
	}
}

func TestParseTaskNotification(t *testing.T) {
	content := `<task-id>agent-123</task-id><status>completed</status><output>Done!</output>`
	notification, err := ParseTaskNotification(content)
	if err != nil {
		t.Errorf("ParseTaskNotification() error = %v", err)
	}
	if notification.TaskID != "agent-123" {
		t.Errorf("TaskID = %v, want %v", notification.TaskID, "agent-123")
	}
	if notification.Status != "completed" {
		t.Errorf("Status = %v, want %v", notification.Status, "completed")
	}
	if notification.Output != "Done!" {
		t.Errorf("Output = %v, want %v", notification.Output, "Done!")
	}
}

func TestGetSystemPrompt(t *testing.T) {
	prompt := GetSystemPrompt([]string{"Bash", "Read"}, []string{"mcp-server"})
	if prompt == "" {
		t.Fatal("GetSystemPrompt() returned empty string")
	}

	// Check that tools are mentioned
	if !contains(prompt, "Bash") || !contains(prompt, "Read") {
		t.Error("System prompt doesn't contain expected tools")
	}

	// Check that MCP servers are mentioned
	if !contains(prompt, "mcp-server") {
		t.Error("System prompt doesn't contain MCP server")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
