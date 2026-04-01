package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPTransport implements HTTP transport for MCP
type HTTPTransport struct {
	config     *ServerConfig
	url        string
	headers    map[string]string
	httpClient *http.Client
	closed     bool
	mu         sync.Mutex
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(config *ServerConfig) *HTTPTransport {
	return &HTTPTransport{
		config: config,
		url:    config.URL,
		headers: config.Headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the transport name
func (t *HTTPTransport) Name() string {
	return "http"
}

// Connect establishes the HTTP connection
func (t *HTTPTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// HTTP transport doesn't need a persistent connection
	// Just verify the endpoint is reachable
	req, err := http.NewRequestWithContext(ctx, "GET", t.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// Send sends a JSON-RPC message
func (t *HTTPTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Receive receives a JSON-RPC message
func (t *HTTPTransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	// HTTP is request/response based, so we read the response body
	// This is typically called after Send() for the same message ID
	return nil, fmt.Errorf("Receive not supported for HTTP transport - use SendRequest instead")
}

// SendRequest sends a request and returns the response
func (t *HTTPTransport) SendRequest(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	if msg.ID == "" {
		return nil, fmt.Errorf("message ID is required for HTTP transport")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	var rpcResp JSONRPCMessage
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &rpcResp, nil
}

// Close closes the connection
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	return nil
}

// StreamResponse represents a streaming HTTP response
type StreamResponse struct {
	reader *strings.Reader
}

// Read implements io.Reader
func (r *StreamResponse) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func init() {
	// Register HTTP transport
	RegisterTransport(TransportHTTP, func(config *ServerConfig) Transport {
		return NewHTTPTransport(config)
	})
}
