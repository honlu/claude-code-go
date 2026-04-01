package query

import (
	"testing"

	"github.com/ai/claude-code/internal/api"
)

func TestQueryEngineCreation(t *testing.T) {
	q := NewQueryEngine()
	if q == nil {
		t.Fatal("NewQueryEngine returned nil")
	}
	if q.model != "claude-sonnet-4-6" {
		t.Errorf("default model = %v, want %v", q.model, "claude-sonnet-4-6")
	}
	if q.maxIterations != 20 {
		t.Errorf("maxIterations = %v, want %v", q.maxIterations, 20)
	}
}

func TestQueryEngineWithOptions(t *testing.T) {
	q := NewQueryEngine(
		WithModel("claude-opus-4-6"),
		WithMaxRetries(5),
		WithMaxIterations(10),
	)

	if q.model != "claude-opus-4-6" {
		t.Errorf("model = %v, want %v", q.model, "claude-opus-4-6")
	}
	if q.maxRetries != 5 {
		t.Errorf("maxRetries = %v, want %v", q.maxRetries, 5)
	}
	if q.maxIterations != 10 {
		t.Errorf("maxIterations = %v, want %v", q.maxIterations, 10)
	}
}

func TestQueryEngineSetGetModel(t *testing.T) {
	q := NewQueryEngine()

	q.SetModel("claude-haiku-4-5")
	if q.GetModel() != "claude-haiku-4-5" {
		t.Errorf("GetModel() = %v, want %v", q.GetModel(), "claude-haiku-4-5")
	}
}

func TestQueryEngineClearMessages(t *testing.T) {
	q := NewQueryEngine()

	q.mu.Lock()
	q.messages = []api.Message{
		{Role: "user", Content: "hello"},
	}
	q.mu.Unlock()

	q.ClearMessages()

	if len(q.messages) != 0 {
		t.Errorf("len(messages) after ClearMessages = %d, want 0", len(q.messages))
	}
}

func TestQueryEngineBuildToolsInput(t *testing.T) {
	q := NewQueryEngine()

	// Test with nil tool list - should return empty slice, not nil
	result := q.buildToolsInput(nil)
	if len(result) != 0 {
		t.Errorf("buildToolsInput(nil) len = %d, want 0", len(result))
	}
}

func TestQueryEngineConvertInputSchema(t *testing.T) {
	q := NewQueryEngine()

	// Test with nil schema
	result := q.convertInputSchema(nil)
	if result != nil {
		t.Errorf("convertInputSchema(nil) = %v, want nil", result)
	}

	// Test with empty map
	result = q.convertInputSchema(map[string]any{})
	if result == nil {
		t.Error("convertInputSchema({}) returned nil, want empty schema")
	}
}

func TestQueryEngineProcessResponseNil(t *testing.T) {
	q := NewQueryEngine()

	_, err := q.processResponse(nil)
	if err == nil {
		t.Error("processResponse(nil) error = nil, want error")
	}
}

func TestQueryEngineGetWorkingDirectory(t *testing.T) {
	// Note: GetWorkingDirectory depends on bootstrap.Init() being called first
	// In integration tests, bootstrap.Init() is called
	// This test just verifies the function doesn't panic
	dir := GetWorkingDirectory()
	// Directory may be empty in unit test context without bootstrap.Init()
	_ = dir
}

func TestQueryEngineToolExecution(t *testing.T) {
	q := NewQueryEngine()

	// Test executeToolCalls with empty content
	err := q.executeToolCalls(nil)
	if err != nil {
		t.Errorf("executeToolCalls(nil) error = %v, want nil", err)
	}

	err = q.executeToolCalls([]api.ContentBlock{})
	if err != nil {
		t.Errorf("executeToolCalls([]) error = %v, want nil", err)
	}
}
