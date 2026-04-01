package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Note: WebSocket support requires github.com/gorilla/websocket
// For now, this is a stub implementation
// To enable: go get github.com/gorilla/websocket and uncomment the code

// WebSocketTransport implements WebSocket transport for MCP
type WebSocketTransport struct {
	config     *ServerConfig
	url        string
	headers    map[string]string
	readCh     chan *JSONRPCMessage
	writeCh    chan *JSONRPCMessage
	closed     bool
	mu         sync.Mutex
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(config *ServerConfig) *WebSocketTransport {
	return &WebSocketTransport{
		config:  config,
		url:     config.URL,
		headers: config.Headers,
		readCh:  make(chan *JSONRPCMessage, 100),
		writeCh: make(chan *JSONRPCMessage, 100),
	}
}

// Name returns the transport name
func (t *WebSocketTransport) Name() string {
	return "websocket"
}

// Connect establishes the WebSocket connection
func (t *WebSocketTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Parse URL
	if !strings.HasPrefix(t.url, "ws://") && !strings.HasPrefix(t.url, "wss://") {
		t.url = "ws://" + t.url
	}

	// TODO: Implement WebSocket connection using gorilla/websocket
	// Example:
	// dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	// conn, _, err := dialer.DialContext(ctx, t.url, nil)
	// if err != nil {
	//     return fmt.Errorf("failed to connect: %w", err)
	// }
	// t.conn = conn

	return fmt.Errorf("WebSocket transport not yet implemented - requires github.com/gorilla/websocket")
}

// Send sends a JSON-RPC message
func (t *WebSocketTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	select {
	case t.writeCh <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("WebSocket not connected")
	}
}

// Receive receives a JSON-RPC message
func (t *WebSocketTransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-t.readCh:
		return msg, nil
	default:
		return nil, fmt.Errorf("WebSocket not connected")
	}
}

// Close closes the connection
func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	close(t.writeCh)
	close(t.readCh)

	return nil
}

// WebSocketIDETransport implements WebSocket transport for IDE extensions
type WebSocketIDETransport struct {
	*WebSocketTransport
	ideName string
}

// NewWebSocketIDETransport creates a new WebSocket IDE transport
func NewWebSocketIDETransport(config *ServerConfig) *WebSocketIDETransport {
	return &WebSocketIDETransport{
		WebSocketTransport: NewWebSocketTransport(config),
		ideName:            config.URL,
	}
}

// AddAuthToken adds an auth token to the connection
func (t *WebSocketTransport) AddAuthToken(token string) {
	if t.headers == nil {
		t.headers = make(map[string]string)
	}
	t.headers["Authorization"] = "Bearer " + token
}

func init() {
	// Register WebSocket transports
	RegisterTransport(TransportWS, func(config *ServerConfig) Transport {
		return NewWebSocketTransport(config)
	})
}
