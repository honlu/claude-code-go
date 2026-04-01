package api

import "time"

// Provider represents the API provider type
type Provider int

const (
	ProviderDirect Provider = iota
	ProviderBedrock
	ProviderFoundry
	ProviderVertex
)

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []ContentBlock
}

// ContentBlock represents a content block
type ContentBlock struct {
	Type         string         `json:"type"`
	Text         string         `json:"text,omitempty"`
	ToolUse      *ToolUse       `json:"tool_use,omitempty"`
	Id           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Input        map[string]any `json:"input,omitempty"`
	ToolUseId    string         `json:"tool_use_id,omitempty"`
	Content      string         `json:"content,omitempty"`
	IsError      bool           `json:"is_error,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema *InputSchema   `json:"input_schema,omitempty"`
}

// InputSchema defines the input schema for a tool
type InputSchema struct {
	Type       string            `json:"type,omitempty"`
	Properties map[string]any    `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
}

// ToolUse represents a tool use block from the API
type ToolUse struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Input     map[string]any `json:"input"`
	Id        string `json:"id"`
}

// ToolResult represents a tool result block
type ToolResult struct {
	Type      string `json:"type"`
	ToolUseId string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// MessageRequest represents a request to create a message
type MessageRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	MaxTokens   int       `json:"max_tokens"`
	Tools       []Tool    `json:"tools,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	TopK        int       `json:"top_k,omitempty"`
}

// MessageResponse represents a response from createMessage
type MessageResponse struct {
	Id        string         `json:"id"`
	Type      string         `json:"type"`
	Role      string         `json:"role"`
	Content   []ContentBlock `json:"content"`
	Model     string         `json:"model"`
	StopReason string        `json:"stop_reason"`
	StopSequence interface{}  `json:"stop_sequence"`
	Usage     MessageUsage    `json:"usage"`
}

// MessageUsage represents token usage
type MessageUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamEvent represents a streaming event from the API
type StreamEvent struct {
	Type string      `json:"type"`
	Data any         `json:"data"`
}

// StreamMessageStartEvent is sent when the message starts
type StreamMessageStartEvent struct {
	Type    string         `json:"type"`
	Message MessageResponse `json:"message"`
}

// StreamContentBlockStartEvent is sent when a content block starts
type StreamContentBlockStartEvent struct {
	Type        string `json:"type"`
	Index       int    `json:"index"`
	ContentBlock ContentBlock `json:"content_block"`
}

// StreamContentBlockDeltaEvent is sent for content deltas
type StreamContentBlockDeltaEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"delta"`
}

// StreamMessageDeltaEvent is sent when the message is complete
type StreamMessageDeltaEvent struct {
	Type    string `json:"type"`
	Delta   struct {
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
	Usage MessageUsage `json:"usage"`
}

// StreamMessageStopEvent is sent when streaming is complete
type StreamMessageStopEvent struct {
	Type string `json:"type"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Type string `json:"type"`
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ModelInfo contains information about a model
type ModelInfo struct {
	Name         string
	DisplayName  string
	InputTokens  int
	OutputTokens int
	SupportsVision bool
	SupportsTools  bool
}

// PricingInfo contains pricing data
type PricingInfo struct {
	InputCostPerToken  float64
	OutputCostPerToken float64
	Currency          string
}

// DefaultModels contains the default models
var DefaultModels = map[string]ModelInfo{
	"claude-opus-4-5": {
		Name:         "claude-opus-4-5",
		DisplayName:  "Claude Opus 4.5",
		InputTokens:  200000,
		OutputTokens: 8192,
		SupportsVision: true,
		SupportsTools:  true,
	},
	"claude-sonnet-4-6": {
		Name:         "claude-sonnet-4-6",
		DisplayName:  "Claude Sonnet 4.6",
		InputTokens:  200000,
		OutputTokens: 8192,
		SupportsVision: true,
		SupportsTools:  true,
	},
	"claude-haiku-4-5": {
		Name:         "claude-haiku-4-5",
		DisplayName:  "Claude Haiku 4.5",
		InputTokens:  200000,
		OutputTokens: 8192,
		SupportsVision: true,
		SupportsTools:  true,
	},
}

// Default pricing (placeholder - actual pricing varies)
var DefaultPricing = map[string]PricingInfo{
	"claude-opus-4-5": {InputCostPerToken: 0.000015, OutputCostPerToken: 0.000075},
	"claude-sonnet-4-6": {InputCostPerToken: 0.000003, OutputCostPerToken: 0.000015},
	"claude-haiku-4-5": {InputCostPerToken: 0.0000008, OutputCostPerToken: 0.000004},
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	Type       string
	LimitType  string // "five_hour", "seven_day", etc.
	RetryAfter time.Duration
}

// APIError represents an API error
type APIError struct {
	Type    string
	Message string
	Code    string
}

func (e *APIError) Error() string {
	return e.Message
}

// IsRateLimitError returns true if this is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	return e.Type == "rate_limit_error"
}
