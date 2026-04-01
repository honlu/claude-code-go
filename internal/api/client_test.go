package api

import (
	"os"
	"testing"
)

func TestProviderFromEnv(t *testing.T) {
	// Reset environment
	os.Unsetenv("CLAUDE_CODE_USE_BEDROCK")
	os.Unsetenv("CLAUDE_CODE_USE_FOUNDRY")
	os.Unsetenv("CLAUDE_CODE_USE_VERTEX")

	// Test direct provider (default)
	if got := ProviderFromEnv(); got != ProviderDirect {
		t.Errorf("ProviderFromEnv() = %v, want %v", got, ProviderDirect)
	}

	// Test Bedrock
	os.Setenv("CLAUDE_CODE_USE_BEDROCK", "true")
	if got := ProviderFromEnv(); got != ProviderBedrock {
		t.Errorf("ProviderFromEnv() = %v, want %v", got, ProviderBedrock)
	}
	os.Unsetenv("CLAUDE_CODE_USE_BEDROCK")

	// Test Foundry
	os.Setenv("CLAUDE_CODE_USE_FOUNDRY", "true")
	if got := ProviderFromEnv(); got != ProviderFoundry {
		t.Errorf("ProviderFromEnv() = %v, want %v", got, ProviderFoundry)
	}
	os.Unsetenv("CLAUDE_CODE_USE_FOUNDRY")

	// Test Vertex
	os.Setenv("CLAUDE_CODE_USE_VERTEX", "true")
	if got := ProviderFromEnv(); got != ProviderVertex {
		t.Errorf("ProviderFromEnv() = %v, want %v", got, ProviderVertex)
	}
	os.Unsetenv("CLAUDE_CODE_USE_VERTEX")
}

func TestParseCustomHeaders(t *testing.T) {
	headers := parseCustomHeaders("Authorization:Bearer token123,X-Custom:header")
	if headers["Authorization"] != "Bearer token123" {
		t.Errorf("Authorization = %v, want %v", headers["Authorization"], "Bearer token123")
	}
	if headers["X-Custom"] != "header" {
		t.Errorf("X-Custom = %v, want %v", headers["X-Custom"], "header")
	}
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(ProviderDirect)
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
	if c.provider != ProviderDirect {
		t.Errorf("c.provider = %v, want %v", c.provider, ProviderDirect)
	}
	if c.maxRetries != 5 {
		t.Errorf("c.maxRetries = %v, want %v", c.maxRetries, 5)
	}
}

func TestClientOptions(t *testing.T) {
	c, _ := NewClient(ProviderDirect,
		WithAPIKey("test-key"),
		WithBaseURL("https://custom.api.com"),
		WithMaxRetries(3),
	)

	if c.apiKey != "test-key" {
		t.Errorf("apiKey = %v, want %v", c.apiKey, "test-key")
	}
	if c.baseURL != "https://custom.api.com" {
		t.Errorf("baseURL = %v, want %v", c.baseURL, "https://custom.api.com")
	}
	if c.maxRetries != 3 {
		t.Errorf("maxRetries = %v, want %v", c.maxRetries, 3)
	}
}
