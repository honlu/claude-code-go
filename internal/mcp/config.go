package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents MCP configuration
type Config struct {
	Servers map[string]*ServerConfig `json:"mcp_servers,omitempty"`
}

// LoadConfig loads MCP configuration from files
func LoadConfig() (*Config, error) {
	config := &Config{
		Servers: make(map[string]*ServerConfig),
	}

	// Load from multiple scopes: enterprise > local > project > user > plugin
	scopes := []struct {
		scope ConfigScope
		path  string
	}{
		{ConfigScopeEnterprise, getEnterpriseConfigPath()},
		{ConfigScopeLocal, getLocalConfigPath()},
		{ConfigScopeProject, getProjectConfigPath()},
		{ConfigScopeUser, getUserConfigPath()},
	}

	for _, s := range scopes {
		if path := s.path; path != "" {
			if err := loadConfigFile(path, config); err == nil {
				break // Stop at first valid config
			}
		}
	}

	return config, nil
}

// getEnterpriseConfigPath returns the enterprise config path
func getEnterpriseConfigPath() string {
	return os.Getenv("CLAUDE_CODE_ENTERPRISE_CONFIG")
}

// getLocalConfigPath returns the local config path
func getLocalConfigPath() string {
	if path := os.Getenv("CLAUDE_CODE_MCP_CONFIG"); path != "" {
		return path
	}
	return ""
}

// getProjectConfigPath returns the project config path
func getProjectConfigPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, ".claude", "mcp.json")
}

// getUserConfigPath returns the user config path
func getUserConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "mcp.json")
}

// loadConfigFile loads configuration from a file
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	// Merge with existing config
	for name, server := range fileConfig.Servers {
		if config.Servers == nil {
			config.Servers = make(map[string]*ServerConfig)
		}
		server.Scope = ConfigScopeLocal // Default scope
		config.Servers[name] = server
	}

	return nil
}

// GetServer returns a server configuration by name
func (c *Config) GetServer(name string) (*ServerConfig, bool) {
	server, ok := c.Servers[name]
	return server, ok
}

// AddServer adds a server configuration
func (c *Config) AddServer(name string, config *ServerConfig) {
	if c.Servers == nil {
		c.Servers = make(map[string]*ServerConfig)
	}
	c.Servers[name] = config
}

// RemoveServer removes a server configuration
func (c *Config) RemoveServer(name string) {
	delete(c.Servers, name)
}

// Save saves the configuration to the user config path
func (c *Config) Save() error {
	path := getUserConfigPath()
	if path == "" {
		return fmt.Errorf("cannot determine user config path")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ValidateConfig validates server configuration
func ValidateConfig(config *ServerConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.Name == "" {
		return fmt.Errorf("server name is required")
	}

	switch config.Transport {
	case TransportStdio:
		if config.Command == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
	case TransportSSE, TransportHTTP:
		if config.URL == "" {
			return fmt.Errorf("url is required for %s transport", config.Transport)
		}
	case TransportWS:
		if config.URL == "" {
			return fmt.Errorf("url is required for websocket transport")
		}
	default:
		return fmt.Errorf("unknown transport type: %s", config.Transport)
	}

	return nil
}

// FilterByScope filters servers by configuration scope
func FilterByScope(servers map[string]*ServerConfig, scope ConfigScope) map[string]*ServerConfig {
	result := make(map[string]*ServerConfig)
	for name, server := range servers {
		if server.Scope == scope {
			result[name] = server
		}
	}
	return result
}

// ApplyEnterprisePolicies applies enterprise policies to configuration
func ApplyEnterprisePolicies(config *Config, policies *EnterprisePolicies) *Config {
	if policies == nil {
		return config
	}

	filtered := make(map[string]*ServerConfig)
	for name, server := range config.Servers {
		// Check allowlist
		if len(policies.AllowedServers) > 0 {
			if !contains(policies.AllowedServers, name) {
				continue
			}
		}

		// Check denylist
		if len(policies.DeniedServers) > 0 {
			if contains(policies.DeniedServers, name) {
				continue
			}
		}

		filtered[name] = server
	}

	config.Servers = filtered
	return config
}

// EnterprisePolicies represents enterprise MCP policies
type EnterprisePolicies struct {
	AllowedServers []string `json:"allowed_mcp_servers,omitempty"`
	DeniedServers  []string `json:"denied_mcp_servers,omitempty"`
}

// contains checks if a slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ParseServerConfig parses a server configuration from various formats
func ParseServerConfig(name string, data any) (*ServerConfig, error) {
	config := &ServerConfig{Name: name}

	switch v := data.(type) {
	case map[string]any:
		// Parse command-based config (stdio)
		if cmd, ok := v["command"].(string); ok {
			config.Transport = TransportStdio
			config.Command = cmd
			if args, ok := v["args"].([]any); ok {
				for _, a := range args {
					if s, ok := a.(string); ok {
						config.Args = append(config.Args, s)
					}
				}
			}
			if env, ok := v["env"].(map[string]any); ok {
				config.Env = make(map[string]string)
				for k, v := range env {
					if s, ok := v.(string); ok {
						config.Env[k] = s
					}
				}
			}
			return config, nil
		}

		// Parse URL-based config (sse, http, ws)
		if url, ok := v["url"].(string); ok {
			if strings.HasPrefix(url, "http") {
				if _, ok := v["type"]; ok {
					t := v["type"].(string)
					switch t {
					case "sse":
						config.Transport = TransportSSE
					case "http":
						config.Transport = TransportHTTP
					default:
						config.Transport = TransportSSE
					}
				} else {
					config.Transport = TransportSSE
				}
			} else {
				config.Transport = TransportWS
			}
			config.URL = url

			if headers, ok := v["headers"].(map[string]any); ok {
				config.Headers = make(map[string]string)
				for k, v := range headers {
					if s, ok := v.(string); ok {
						config.Headers[k] = s
					}
				}
			}
			return config, nil
		}

	default:
		return nil, fmt.Errorf("unsupported config format")
	}

	return nil, fmt.Errorf("invalid server config")
}
