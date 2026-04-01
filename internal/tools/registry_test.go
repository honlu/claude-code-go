package tools

import (
	"context"
	"testing"
)

type mockTool struct {
	name        string
	description string
}

func (m *mockTool) Name() string                    { return m.name }
func (m *mockTool) Description(_ map[string]any) string { return m.description }
func (m *mockTool) InputSchema() any                { return nil }
func (m *mockTool) OutputSchema() any               { return nil }
func (m *mockTool) Call(ctx context.Context, args map[string]any, tc ToolContext) (*ToolResult, error) {
	return &ToolResult{Data: "mock result"}, nil
}
func (m *mockTool) IsConcurrencySafe(_ map[string]any) bool { return false }
func (m *mockTool) IsReadOnly(_ map[string]any) bool   { return false }
func (m *mockTool) IsDestructive(_ map[string]any) bool { return false }
func (m *mockTool) IsOpenWorld(_ map[string]any) bool   { return false }
func (m *mockTool) CheckPermissions(_ context.Context, _ map[string]any, _ ToolContext) (PermissionResult, error) {
	return PermissionResult{Behavior: "allow"}, nil
}
func (m *mockTool) UserFacingName() string            { return m.name }
func (m *mockTool) RenderToolUseMessage(_ map[string]any) string { return "" }
func (m *mockTool) RenderToolResultMessage(_ *ToolResult) string { return "" }
func (m *mockTool) IsMCP() bool                       { return false }
func (m *mockTool) MCPInfo() (string, string)         { return "", "" }

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "MockTool", description: "desc"}

	r.Register(tool)

	got, ok := r.Get("MockTool")
	if !ok {
		t.Error("Get() returned false, want true")
	}
	if got.Name() != "MockTool" {
		t.Errorf("got.Name() = %v, want %v", got.Name(), "MockTool")
	}
}

func TestRegistryGetOrError(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "TestTool", description: "desc"}
	r.Register(tool)

	got, err := r.GetOrError("TestTool")
	if err != nil {
		t.Errorf("GetOrError() error = %v", err)
	}
	if got.Name() != "TestTool" {
		t.Errorf("got.Name() = %v, want %v", got.Name(), "TestTool")
	}

	_, err = r.GetOrError("NonExistent")
	if err == nil {
		t.Error("GetOrError() error = nil, want error")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "Tool1", description: ""})
	r.Register(&mockTool{name: "Tool2", description: ""})

	tools := r.List()
	if len(tools) != 2 {
		t.Errorf("len(List()) = %d, want %d", len(tools), 2)
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Register a tool with the global registry
	tool := &mockTool{name: "GlobalTestTool", description: "desc"}
	RegisterTool(tool)

	got, ok := GetTool("GlobalTestTool")
	if !ok {
		t.Error("GetTool() returned false, want true")
	}
	if got.Name() != "GlobalTestTool" {
		t.Errorf("got.Name() = %v, want %v", got.Name(), "GlobalTestTool")
	}
}
