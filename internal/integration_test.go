package integration

import (
	"context"
	"os"
	"testing"

	"github.com/ai/claude-code/internal/api"
	"github.com/ai/claude-code/internal/bootstrap"
	"github.com/ai/claude-code/internal/coordinator"
	"github.com/ai/claude-code/internal/mcp"
	"github.com/ai/claude-code/internal/tools"
	"github.com/ai/claude-code/internal/tools/bash"
	"github.com/ai/claude-code/internal/tools/edit"
	"github.com/ai/claude-code/internal/tools/glob"
	"github.com/ai/claude-code/internal/tools/grep"
	"github.com/ai/claude-code/internal/tools/read"
	"github.com/ai/claude-code/internal/tools/write"
	_ "github.com/ai/claude-code/internal/tools" // to trigger init
)

// TestIntegrationBootstrapAndTools tests the integration between bootstrap and tools
func TestIntegrationBootstrapAndTools(t *testing.T) {
	// Initialize bootstrap
	bootstrap.Init()

	// Verify tools are registered via their subpackages
	// (importing them triggers their init() which registers the tools)
	registry := tools.GetRegistry()
	if registry == nil {
		t.Fatal("GetRegistry() returned nil")
	}

	// Check that built-in tools are registered
	// Note: tools are registered via init() in subpackages
	expectedTools := []string{"Bash", "Read", "Edit", "Write", "Grep", "Glob"}
	_ = bash.NewTool()   // ensure package is used
	_ = read.NewTool()
	_ = edit.NewTool()
	_ = write.NewTool()
	_ = grep.NewTool()
	_ = glob.NewTool()

	for _, name := range expectedTools {
		tool, ok := registry.Get(name)
		if !ok {
			t.Errorf("Tool %s not registered", name)
			continue
		}
		if tool.Name() != name {
			t.Errorf("tool.Name() = %v, want %v", tool.Name(), name)
		}
	}
}

// TestIntegrationCoordinatorWithConfig tests the coordinator with config
func TestIntegrationCoordinatorWithConfig(t *testing.T) {
	config := &coordinator.CoordinatorConfig{
		Enabled:    true,
		WorkerTools: []string{"Bash", "Read", "Edit"},
		MCPServers: []string{},
		SimpleMode: false,
	}

	c := coordinator.NewCoordinator(config)
	if c == nil {
		t.Fatal("NewCoordinator() returned nil")
	}

	// Spawn an agent
	agent, err := c.SpawnAgent(context.Background(), "test-worker", "Test prompt", config.WorkerTools)
	if err != nil {
		t.Errorf("SpawnAgent() error = %v", err)
	}
	if agent == nil {
		t.Fatal("SpawnAgent() returned nil agent")
	}

	// Verify agent is in the coordinator
	got, ok := c.GetAgent(agent.ID)
	if !ok {
		t.Error("GetAgent() returned false for spawned agent")
	}
	if got.ID != agent.ID {
		t.Errorf("got.ID = %v, want %v", got.ID, agent.ID)
	}

	// List all agents
	agents := c.ListAgents()
	if len(agents) != 1 {
		t.Errorf("len(ListAgents()) = %d, want 1", len(agents))
	}

	// Remove agent
	c.RemoveAgent(agent.ID)
	agents = c.ListAgents()
	if len(agents) != 0 {
		t.Errorf("len(ListAgents()) after remove = %d, want 0", len(agents))
	}
}

// TestIntegrationAPIClient tests the API client initialization
func TestIntegrationAPIClient(t *testing.T) {
	// Test with direct provider
	client, err := api.NewClient(api.ProviderDirect)
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	// Test with options - verify no error occurs
	_, err = api.NewClient(api.ProviderDirect,
		api.WithAPIKey("test-key"),
		api.WithBaseURL("https://custom.api.com"),
		api.WithMaxRetries(10),
	)
	if err != nil {
		t.Errorf("NewClient() with options error = %v", err)
	}
}

// TestIntegrationMCPConfig tests MCP configuration loading
func TestIntegrationMCPConfig(t *testing.T) {
	// Test creating config and adding servers
	config := &mcp.Config{}

	// Add a test server
	config.AddServer("test-server", &mcp.ServerConfig{
		Name:      "test-server",
		Transport: mcp.TransportStdio,
		Command:   "/usr/bin/test-server",
	})

	server, ok := config.GetServer("test-server")
	if !ok {
		t.Error("GetServer() returned false for added server")
	}
	if server.Name != "test-server" {
		t.Errorf("server.Name = %v, want %v", server.Name, "test-server")
	}

	// Remove server
	config.RemoveServer("test-server")
	_, ok = config.GetServer("test-server")
	if ok {
		t.Error("GetServer() returned true after RemoveServer()")
	}
}

// TestIntegrationProviderDetection tests provider detection from environment
func TestIntegrationProviderDetection(t *testing.T) {
	// Reset environment
	os.Unsetenv("CLAUDE_CODE_USE_BEDROCK")
	os.Unsetenv("CLAUDE_CODE_USE_FOUNDRY")
	os.Unsetenv("CLAUDE_CODE_USE_VERTEX")

	// Test direct (default)
	provider := api.ProviderFromEnv()
	if provider != api.ProviderDirect {
		t.Errorf("ProviderFromEnv() = %v, want %v", provider, api.ProviderDirect)
	}
}

// TestIntegrationConcurrentAgents tests concurrent agent operations
func TestIntegrationConcurrentAgents(t *testing.T) {
	c := coordinator.NewCoordinator(&coordinator.CoordinatorConfig{Enabled: true})

	// Spawn multiple agents concurrently
	agentCount := 10
	results := make(chan *coordinator.Agent, agentCount)

	for i := 0; i < agentCount; i++ {
		go func(id int) {
			agent, _ := c.SpawnAgent(context.Background(), "worker", "Test", []string{})
			results <- agent
		}(i)
	}

	// Collect results
	for i := 0; i < agentCount; i++ {
		<-results
	}

	// Verify all agents are registered
	agents := c.ListAgents()
	if len(agents) != agentCount {
		t.Errorf("len(ListAgents()) = %d, want %d", len(agents), agentCount)
	}
}
