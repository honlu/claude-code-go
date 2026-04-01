package mcp

import "time"

// TransportType represents the transport type
type TransportType string

const (
	TransportStdio  TransportType = "stdio"
	TransportSSE    TransportType = "sse"
	TransportSSEIDE TransportType = "sse-ide"
	TransportHTTP   TransportType = "http"
	TransportWS     TransportType = "ws"
	TransportSDK    TransportType = "sdk"
)

// ConfigScope represents the configuration scope
type ConfigScope string

const (
	ConfigScopeLocal     ConfigScope = "local"
	ConfigScopeUser      ConfigScope = "user"
	ConfigScopeProject   ConfigScope = "project"
	ConfigScopeDynamic   ConfigScope = "dynamic"
	ConfigScopeEnterprise ConfigScope = "enterprise"
	ConfigScopeClaudeAI  ConfigScope = "claudeai"
	ConfigScopeManaged   ConfigScope = "managed"
)

// ServerConfig represents the base MCP server configuration
type ServerConfig struct {
	Name        string            `json:"name"`
	Transport   TransportType     `json:"transport"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	OAuthConfig *OAuthConfig      `json:"oauth,omitempty"`
	Scope       ConfigScope       `json:"scope,omitempty"`
}

// StdioServerConfig represents stdio transport configuration
type StdioServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// SSEServerConfig represents SSE transport configuration
type SSEServerConfig struct {
	URL         string       `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	OAuthConfig *OAuthConfig `json:"oauth,omitempty"`
}

// HTTPServerConfig represents HTTP transport configuration
type HTTPServerConfig struct {
	URL         string       `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	OAuthConfig *OAuthConfig `json:"oauth,omitempty"`
}

// WebSocketServerConfig represents WebSocket transport configuration
type WebSocketServerConfig struct {
	URL        string `json:"url"`
	AuthToken  string `json:"authToken,omitempty"`
}

// OAuthConfig represents OAuth configuration for MCP server
type OAuthConfig struct {
	ClientID             string `json:"clientId,omitempty"`
	CallbackPort         int    `json:"callbackPort,omitempty"`
	AuthServerMetadataURL string `json:"authServerMetadataUrl,omitempty"`
	XAA                  bool   `json:"xaa,omitempty"`
}

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
	ID      string `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSONRPCErrorCode definitions
const (
	JSONRPCParseError     = -32700
	JSONRPCInvalidRequest = -32600
	JSONRPCMethodNotFound = -32601
	JSONRPCInvalidParams  = -32602
	JSONRPCInternalError  = -32603
)

// ConnectionState represents the state of an MCP server connection
type ConnectionState string

const (
	StatePending   ConnectionState = "pending"
	StateConnecting ConnectionState = "connecting"
	StateConnected  ConnectionState = "connected"
	StateFailed     ConnectionState = "failed"
	StateNeedsAuth  ConnectionState = "needs_auth"
	StateDisabled   ConnectionState = "disabled"
)

// MCPServerConnection represents a connection to an MCP server
type MCPServerConnection struct {
	Name       string          `json:"name"`
	State      ConnectionState `json:"state"`
	Config     *ServerConfig   `json:"config,omitempty"`
	Error      string          `json:"error,omitempty"`
	LastConnected time.Time    `json:"lastConnected,omitempty"`
	Tools      []MCPTool       `json:"tools,omitempty"`
	Resources  []MCPResource   `json:"resources,omitempty"`
}

// MCPTool represents an MCP tool
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema *InputSchema   `json:"input_schema,omitempty"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// InputSchema defines the input schema for an MCP tool
type InputSchema struct {
	Type       string            `json:"type,omitempty"`
	Properties map[string]any    `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
}

// ToolAnnotations provides annotations for tool behavior
type ToolAnnotations struct {
	ReadOnlyHint      bool `json:"readOnlyHint,omitempty"`
	DestructiveHint   bool `json:"destructiveHint,omitempty"`
	OpenWorldHint     bool `json:"openWorldHint,omitempty"`
	IdempotentHint    bool `json:"idempotentHint,omitempty"`
}

// MCPResource represents an MCP resource
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// CallToolResult represents the result of calling an MCP tool
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool          `json:"is_error,omitempty"`
}

// ContentBlock represents a content block in MCP
type ContentBlock struct {
	Type     string `json:"type"` // "text", "image", "resource"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	URI      string `json:"uri,omitempty"`
}

// InitializeResult represents the result of initialization
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Tools    *ToolsCapability    `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability  `json:"prompts,omitempty"`
	Logging    *LoggingCapability  `json:"logging,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"list_changed,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	ListChanged bool `json:"list_changed,omitempty"`
	Subscribe   bool `json:"subscribe,omitempty"`
}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"list_changed,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

// NotifyMethod represents a JSON-RPC notification method
type NotifyMethod string

const (
	NotifyInitialized       NotifyMethod = "initialized"
	NotifyToolsListChanged  NotifyMethod = "notifications/tools/list_changed"
	NotifyResourcesUpdated  NotifyMethod = "notifications/resources/updated"
	NotifyResourcesListChanged NotifyMethod = "notifications/resources/list_changed"
	NotifyPromptsListChanged NotifyMethod = "notifications/prompts/list_changed"
	NotifyLoggingMessage    NotifyMethod = "notifications/message"
)

// RequestMethod represents a JSON-RPC request method
type RequestMethod string

const (
	RequestInitialize     RequestMethod = "initialize"
	RequestToolsList      RequestMethod = "tools/list"
	RequestToolsCall      RequestMethod = "tools/call"
	RequestResourcesList  RequestMethod = "resources/list"
	RequestResourcesRead  RequestMethod = "resources/read"
	RequestResourcesSubscribe RequestMethod = "resources/subscribe"
	RequestPromptsList    RequestMethod = "prompts/list"
	RequestPromptsGet     RequestMethod = "prompts/get"
	RequestLoggingSetLevel RequestMethod = "logging/setLevel"
)
