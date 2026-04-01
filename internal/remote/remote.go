package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client represents a remote client
type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
	mu         sync.RWMutex
}

// NewClient creates a new remote client
func NewClient(baseURL, authToken string) (*Client, error) {
	_, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		baseURL: baseURL,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Request represents a remote request
type Request struct {
	Method  string
	Path    string
	Body    any
	Headers map[string]string
}

// Response represents a remote response
type Response struct {
	StatusCode int
	Body       any
	Headers    map[string]string
}

// Do executes a remote request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	c.mu.RLock()
	baseURL := c.baseURL
	c.mu.RUnlock()

	// Build URL
	u, err := url.Parse(baseURL + req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.mu.RLock()
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	c.mu.RUnlock()

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Execute
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    map[string]string{},
	}, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, &Request{Method: "GET", Path: path})
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	return c.Do(ctx, &Request{Method: "POST", Path: path, Body: body})
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body any) (*Response, error) {
	return c.Do(ctx, &Request{Method: "PUT", Path: path, Body: body})
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, &Request{Method: "DELETE", Path: path})
}

// Server represents a remote server
type Server struct {
	addr    string
	handler http.Handler
	server  *http.Server
}

// NewServer creates a new remote server
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

// Start starts the server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
