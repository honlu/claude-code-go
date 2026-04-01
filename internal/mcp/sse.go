package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SSETransport implements Server-Sent Events transport for MCP
type SSETransport struct {
	config     *ServerConfig
	url        string
	headers    map[string]string
	httpClient *http.Client
	conn       *http.Client
	resp       *http.Response
	readCh     chan *JSONRPCMessage
	closed     bool
	mu         sync.Mutex
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(config *ServerConfig) *SSETransport {
	return &SSETransport{
		config: config,
		url:    config.URL,
		headers: config.Headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		readCh: make(chan *JSONRPCMessage, 100),
	}
}

// Name returns the transport name
func (t *SSETransport) Name() string {
	return "sse"
}

// Connect establishes the SSE connection
func (t *SSETransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Create GET request for SSE
	req, err := http.NewRequestWithContext(ctx, "GET", t.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	t.resp = resp
	return nil
}

// Send sends a JSON-RPC message via POST
func (t *SSETransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// POST messages to the same URL
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.url, strings.NewReader(string(data)))
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
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

// Receive receives a JSON-RPC message from SSE stream
func (t *SSETransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-t.readCh:
		return msg, nil
	}
}

// StartReading starts reading from the SSE stream in background
func (t *SSETransport) StartReading(ctx context.Context) {
	go func() {
		defer close(t.readCh)

		if t.resp == nil || t.resp.Body == nil {
			return
		}

		reader := bufio.NewReader(t.resp.Body)
		var eventData string

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			lineStr := strings.TrimSpace(string(line))

			// Parse SSE format
			if strings.HasPrefix(lineStr, "data: ") {
				eventData = lineStr[6:]

				// Check for end of event (empty line)
				emptyLine, _ := reader.ReadBytes('\n')
				if strings.TrimSpace(string(emptyLine)) == "" {
					// Parse the accumulated data
					if eventData != "" && eventData != "[DONE]" {
						var msg JSONRPCMessage
						if err := json.Unmarshal([]byte(eventData), &msg); err == nil {
							select {
							case t.readCh <- &msg:
							case <-ctx.Done():
								return
							}
						}
					}
					eventData = ""
				}
			}
		}
	}()
}

// Close closes the connection
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	if t.resp != nil && t.resp.Body != nil {
		t.resp.Body.Close()
	}

	return nil
}

// SSEIDETransport implements SSE transport for IDE extensions
type SSEIDETransport struct {
	*SSETransport
	ideName string
}

// NewSSEIDETransport creates a new SSE IDE transport
func NewSSEIDETransport(config *ServerConfig) *SSEIDETransport {
	return &SSEIDETransport{
		SSETransport: NewSSETransport(config),
		ideName:      config.URL,
	}
}

func init() {
	// Register SSE transports
	RegisterTransport(TransportSSE, func(config *ServerConfig) Transport {
		return NewSSETransport(config)
	})
	RegisterTransport(TransportSSEIDE, func(config *ServerConfig) Transport {
		return NewSSEIDETransport(config)
	})
}
