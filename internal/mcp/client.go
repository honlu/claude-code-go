package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Client represents an MCP client
type Client struct {
	mu       sync.RWMutex
	servers  map[string]*ServerConnection
	config   *Config
	transport Transport
}

// ServerConnection represents a connection to an MCP server
type ServerConnection struct {
	Name       string
	Config    *ServerConfig
	Transport  Transport
	State      ConnectionState
	Capabilities *ServerCapabilities
	Tools      []MCPTool
	Resources  []MCPResource
	Error      error
	LastSeen   time.Time
}

// NewClient creates a new MCP client
func NewClient() *Client {
	return &Client{
		servers: make(map[string]*ServerConnection),
	}
}

// LoadConfig loads configuration for the client
func (c *Client) LoadConfig(config *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = config

	// Create server connections
	for name, serverConfig := range config.Servers {
		transport, err := CreateTransport(serverConfig)
		if err != nil {
			continue // Skip invalid transports
		}

		c.servers[name] = &ServerConnection{
			Name:      name,
			Config:    serverConfig,
			Transport: transport,
			State:     StatePending,
		}
	}

	return nil
}

// Connect connects to a server by name
func (c *Client) Connect(ctx context.Context, name string) error {
	c.mu.RLock()
	server, ok := c.servers[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	// Connect transport
	if err := server.Transport.Connect(ctx); err != nil {
		server.State = StateFailed
		server.Error = err
		return fmt.Errorf("failed to connect to %s: %w", name, err)
	}

	// Initialize
	if err := c.initializeServer(ctx, server); err != nil {
		server.State = StateFailed
		server.Error = err
		return fmt.Errorf("failed to initialize %s: %w", name, err)
	}

	server.State = StateConnected
	server.LastSeen = time.Now()
	return nil
}

// initializeServer initializes an MCP server
func (c *Client) initializeServer(ctx context.Context, server *ServerConnection) error {
	// Send initialize request
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":   "claude-code-go",
			"version": "0.1.0",
		},
	}

	msg := CreateRequest("initialize", params, generateID())
	if err := server.Transport.Send(ctx, msg); err != nil {
		return err
	}

	// Receive response
	resp, err := server.Transport.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive initialize response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	// Parse capabilities
	if result, ok := resp.Result.(map[string]any); ok {
		if caps, ok := result["capabilities"].(map[string]any); ok {
			server.Capabilities = parseCapabilities(caps)
		}
	}

	// Send initialized notification
	notify := CreateNotification("initialized", nil)
	if err := server.Transport.Send(ctx, notify); err != nil {
		return err
	}

	// List tools
	if err := c.listTools(ctx, server); err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	return nil
}

// listTools lists tools from a server
func (c *Client) listTools(ctx context.Context, server *ServerConnection) error {
	msg := CreateRequest("tools/list", nil, generateID())
	if err := server.Transport.Send(ctx, msg); err != nil {
		return err
	}

	resp, err := server.Transport.Receive(ctx)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	if result, ok := resp.Result.(map[string]any); ok {
		if tools, ok := result["tools"].([]any); ok {
			server.Tools = parseTools(tools)
		}
	}

	return nil
}

// CallTool calls a tool on a server
func (c *Client) CallTool(ctx context.Context, serverName, toolName string, args map[string]any) (*CallToolResult, error) {
	c.mu.RLock()
	server, ok := c.servers[serverName]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	params := map[string]any{
		"name": toolName,
		"arguments": args,
	}

	msg := CreateRequest("tools/call", params, generateID())
	if err := server.Transport.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to send call request: %w", err)
	}

	resp, err := server.Transport.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive call response: %w", err)
	}

	if resp.Error != nil {
		return &CallToolResult{
			Content: []ContentBlock{{Type: "text", Text: resp.Error.Message}},
			IsError: true,
		}, nil
	}

	return parseCallToolResult(resp)
}

// Disconnect disconnects from a server
func (c *Client) Disconnect(serverName string) error {
	c.mu.RLock()
	server, ok := c.servers[serverName]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("server not found: %s", serverName)
	}

	server.State = StatePending
	return server.Transport.Close()
}

// GetServer returns a server connection
func (c *Client) GetServer(name string) (*ServerConnection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	server, ok := c.servers[name]
	return server, ok
}

// ListServers returns all server names
func (c *Client) ListServers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.servers))
	for name := range c.servers {
		names = append(names, name)
	}
	return names
}

// Helper functions

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func parseCapabilities(data map[string]any) *ServerCapabilities {
	caps := &ServerCapabilities{}

	if tools, ok := data["tools"].(map[string]any); ok {
		caps.Tools = &ToolsCapability{}
		if listChanged, ok := tools["listChanged"].(bool); ok {
			caps.Tools.ListChanged = listChanged
		}
	}

	if resources, ok := data["resources"].(map[string]any); ok {
		caps.Resources = &ResourcesCapability{}
		if listChanged, ok := resources["listChanged"].(bool); ok {
			caps.Resources.ListChanged = listChanged
		}
		if subscribe, ok := resources["subscribe"].(bool); ok {
			caps.Resources.Subscribe = subscribe
		}
	}

	return caps
}

func parseTools(data []any) []MCPTool {
	tools := make([]MCPTool, 0, len(data))
	for _, v := range data {
		if m, ok := v.(map[string]any); ok {
			tool := MCPTool{
				Name:        getString(m, "name"),
				Description: getString(m, "description"),
			}
			if inputSchema, ok := m["input_schema"].(map[string]any); ok {
				tool.InputSchema = &InputSchema{
					Type:       getString(inputSchema, "type"),
					Properties: getMap(inputSchema, "properties"),
					Required:   getStringArray(m, "required"),
				}
			}
			if annotations, ok := m["annotations"].(map[string]any); ok {
				tool.Annotations = &ToolAnnotations{
					ReadOnlyHint:      getBool(annotations, "readOnlyHint"),
					DestructiveHint:   getBool(annotations, "destructiveHint"),
					OpenWorldHint:     getBool(annotations, "openWorldHint"),
					IdempotentHint:    getBool(annotations, "idempotentHint"),
				}
			}
			tools = append(tools, tool)
		}
	}
	return tools
}

func parseCallToolResult(resp *JSONRPCMessage) (*CallToolResult, error) {
	result := &CallToolResult{
		Content: make([]ContentBlock, 0),
	}

	if resp.Result == nil {
		return result, nil
	}

	if m, ok := resp.Result.(map[string]any); ok {
		if content, ok := m["content"].([]any); ok {
			for _, v := range content {
				if cb, ok := v.(map[string]any); ok {
					block := ContentBlock{
						Type: getString(cb, "type"),
						Text: getString(cb, "text"),
					}
					if block.Type == "image" {
						block.Data = getString(cb, "data")
						block.MimeType = getString(cb, "mimeType")
					}
					result.Content = append(result.Content, block)
				}
			}
		}
		if isError, ok := m["isError"].(bool); ok {
			result.IsError = isError
		}
	}

	return result, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return nil
}

func getStringArray(m map[string]any, key string) []string {
	if arr, ok := m[key].([]any); ok {
		result := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
