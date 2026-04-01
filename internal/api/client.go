package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client is the Anthropic API client
type Client struct {
	provider     Provider
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	maxRetries   int
	authHeaders  map[string]string
}

// ClientOption is a function that configures the client
type ClientOption func(*Client)

// NewClient creates a new API client
func NewClient(provider Provider, opts ...ClientOption) (*Client, error) {
	c := &Client{
		provider:   provider,
		httpClient: &http.Client{Timeout: 600 * time.Second},
		maxRetries: 5,
	}

	// Set defaults based on provider first
	switch provider {
	case ProviderDirect:
		c.baseURL = getDirectBaseURL()
		c.apiKey = getAPIKey()
	case ProviderBedrock:
		c.baseURL = getBedrockBaseURL()
	case ProviderFoundry:
		c.baseURL = getFoundryBaseURL()
	case ProviderVertex:
		c.baseURL = getVertexBaseURL()
	}

	// Apply options (overrides defaults)
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithAPIKey sets the API key
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets the HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithMaxRetries sets the max retries
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// getAPIKey gets the API key from environment
func getAPIKey() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	return ""
}

// getDirectBaseURL gets the base URL for direct API
func getDirectBaseURL() string {
	if url := os.Getenv("ANTHROPIC_API_URL"); url != "" {
		return url
	}
	return "https://api.anthropic.com"
}

// getBedrockBaseURL gets the base URL for AWS Bedrock
func getBedrockBaseURL() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	return fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
}

// getFoundryBaseURL gets the base URL for Azure Foundry
func getFoundryBaseURL() string {
	if url := os.Getenv("ANTHROPIC_FOUNDRY_BASE_URL"); url != "" {
		return url
	}
	resource := os.Getenv("ANTHROPIC_FOUNDRY_RESOURCE")
	if resource != "" {
		return fmt.Sprintf("https://%s.services.ai.azure.com", resource)
	}
	return ""
}

// getVertexBaseURL gets the base URL for Google Vertex AI
func getVertexBaseURL() string {
	project := os.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
	region := os.Getenv("CLOUD_ML_REGION")
	if region == "" {
		region = "us-east5"
	}
	if project != "" {
		return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s", region, project, region)
	}
	return ""
}

// CreateMessage creates a message
func (c *Client) CreateMessage(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

// CreateMessageStream creates a streaming message
func (c *Client) CreateMessageStream(ctx context.Context, req *MessageRequest) (<-chan StreamEvent, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	events := make(chan StreamEvent, 100)
	go c.readStream(resp.Body, events)
	return events, nil
}

// setHeaders sets the common headers
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-app", "cli")
	req.Header.Set("anthropic-version", "2023-06-01")

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	// Add custom headers from environment
	if customHeaders := os.Getenv("ANTHROPIC_CUSTOM_HEADERS"); customHeaders != "" {
		headers := parseCustomHeaders(customHeaders)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
}

// parseCustomHeaders parses custom headers from environment
func parseCustomHeaders(headers string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(headers, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// handleResponse handles the HTTP response
func (c *Client) handleResponse(resp *http.Response) (*MessageResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var msgResp MessageResponse
		if err := json.Unmarshal(body, &msgResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &msgResp, nil
	}

	// Handle error
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &APIError{
				Type:    errResp.Type,
				Message: errResp.Error.Message,
				Code:    errResp.Error.Code,
			}
		}
	}

	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

// readStream reads from the response body and sends events
func (c *Client) readStream(body io.Reader, events chan<- StreamEvent) {
	defer close(events)

	reader := io.Reader(body)
	for {
		data, err := readSSEEvent(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			events <- StreamEvent{Type: "error", Data: err.Error()}
			return
		}

		if data == "" {
			continue
		}

		// Parse the event
		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		events <- event
	}
}

// readSSEEvent reads a single SSE event
func readSSEEvent(reader io.Reader) (string, error) {
	var data string
	var prefix []byte
	buf := make([]byte, 1)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			return data, err
		}
		if n == 0 {
			continue
		}

		prefix = append(prefix, buf[0])

		// Check for SSE event format
		if len(prefix) >= 6 && string(prefix[len(prefix)-6:]) == "data: " {
			data = string(prefix[:len(prefix)-6])
			prefix = prefix[:0]
		}

		// End of event
		if buf[0] == '\n' {
			if len(prefix) > 0 {
				data += string(prefix)
				prefix = prefix[:0]
			}
			// Check for [DONE]
			if data == "[DONE]" {
				return "", io.EOF
			}
			if strings.HasPrefix(data, "data: ") {
				return strings.TrimPrefix(data, "data: "), nil
			}
			data = ""
		}
	}
}

// ProviderFromEnv determines the provider from environment variables
func ProviderFromEnv() Provider {
	switch {
	case os.Getenv("CLAUDE_CODE_USE_BEDROCK") != "":
		return ProviderBedrock
	case os.Getenv("CLAUDE_CODE_USE_FOUNDRY") != "":
		return ProviderFoundry
	case os.Getenv("CLAUDE_CODE_USE_VERTEX") != "":
		return ProviderVertex
	default:
		return ProviderDirect
	}
}
