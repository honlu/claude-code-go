package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Transport represents an MCP transport
type Transport interface {
	// Connect establishes the connection
	Connect(ctx context.Context) error

	// Send sends a JSON-RPC message
	Send(ctx context.Context, msg *JSONRPCMessage) error

	// Receive receives a JSON-RPC message (blocking)
	Receive(ctx context.Context) (*JSONRPCMessage, error)

	// Close closes the connection
	Close() error

	// Name returns the transport name
	Name() string
}

// TransportFunc is a function that creates a transport
type TransportFunc func(config *ServerConfig) Transport

// TransportRegistry holds registered transport constructors
type TransportRegistry struct {
	mu       sync.RWMutex
	registry map[TransportType]TransportFunc
}

// NewTransportRegistry creates a new transport registry
func NewTransportRegistry() *TransportRegistry {
	return &TransportRegistry{
		registry: make(map[TransportType]TransportFunc),
	}
}

// Register registers a transport constructor
func (r *TransportRegistry) Register(t TransportType, fn TransportFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry[t] = fn
}

// Get gets a transport constructor
func (r *TransportRegistry) Get(t TransportType) (TransportFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.registry[t]
	return fn, ok
}

// Global transport registry
var globalTransportRegistry = NewTransportRegistry()

// RegisterTransport registers a transport globally
func RegisterTransport(t TransportType, fn TransportFunc) {
	globalTransportRegistry.Register(t, fn)
}

// CreateTransport creates a transport from a server config
func CreateTransport(config *ServerConfig) (Transport, error) {
	fn, ok := globalTransportRegistry.Get(config.Transport)
	if !ok {
		return nil, fmt.Errorf("unsupported transport: %s", config.Transport)
	}
	return fn(config), nil
}

// JSONRPCTransport provides common JSON-RPC transport functionality
type JSONRPCTransport struct {
	mu      sync.Mutex
	pending map[string]chan *JSONRPCMessage
}

// NewJSONRPCTransport creates a new JSON-RPC transport
func NewJSONRPCTransport() *JSONRPCTransport {
	return &JSONRPCTransport{
		pending: make(map[string]chan *JSONRPCMessage),
	}
}

// SendRequest sends a request and waits for response
func (t *JSONRPCTransport) SendRequest(ctx context.Context, msg *JSONRPCMessage, sendFunc func(*JSONRPCMessage) error) (*JSONRPCMessage, error) {
	if msg.ID == "" {
		return nil, fmt.Errorf("message ID is required for requests")
	}

	// Create response channel
	respCh := make(chan *JSONRPCMessage, 1)
	t.mu.Lock()
	t.pending[msg.ID] = respCh
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.pending, msg.ID)
		t.mu.Unlock()
	}()

	// Send the request
	if err := sendFunc(msg); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		return resp, nil
	}
}

// HandleResponse handles an incoming JSON-RPC response
func (t *JSONRPCTransport) HandleResponse(msg *JSONRPCMessage) {
	if msg.ID == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, ok := t.pending[msg.ID]; ok {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ParseJSONRPC parses a JSON-RPC message
func ParseJSONRPC(data []byte) (*JSONRPCMessage, error) {
	var msg JSONRPCMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC message: %w", err)
	}
	return &msg, nil
}

// MarshalJSONRPC marshals a JSON-RPC message
func MarshalJSONRPC(msg *JSONRPCMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// CreateRequest creates a new JSON-RPC request
func CreateRequest(method string, params any, id string) *JSONRPCMessage {
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// CreateResponse creates a new JSON-RPC response
func CreateResponse(id string, result any) *JSONRPCMessage {
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// CreateError creates a new JSON-RPC error response
func CreateError(id string, code int, message string) *JSONRPCMessage {
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}

// CreateNotification creates a new JSON-RPC notification (no ID)
func CreateNotification(method string, params any) *JSONRPCMessage {
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// IsRequest returns true if the message is a request
func IsRequest(msg *JSONRPCMessage) bool {
	return msg.Method != "" && msg.ID != ""
}

// IsNotification returns true if the message is a notification
func IsNotification(msg *JSONRPCMessage) bool {
	return msg.Method != "" && msg.ID == ""
}

// IsResponse returns true if the message is a response
func IsResponse(msg *JSONRPCMessage) bool {
	return msg.Result != nil || msg.Error != nil
}

// ReadJSON reads a JSON object from an io.Reader
func ReadJSON(r io.Reader) (any, error) {
	decoder := json.NewDecoder(r)
	var v any
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
