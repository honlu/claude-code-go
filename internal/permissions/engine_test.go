package permissions

import (
	"context"
	"testing"
)

type mockToolContext struct {
	workingDir string
	user       string
	interactive bool
}

func (m *mockToolContext) GetWorkingDirectory() string {
	return m.workingDir
}

func (m *mockToolContext) GetCurrentUser() string {
	return m.user
}

func (m *mockToolContext) IsInteractive() bool {
	return m.interactive
}

func TestNewPermissionEngine(t *testing.T) {
	e := NewPermissionEngine()
	if e == nil {
		t.Fatal("NewPermissionEngine returned nil")
	}
	if e.GetMode() != ModePrompt {
		t.Errorf("GetMode() = %v, want %v", e.GetMode(), ModePrompt)
	}
}

func TestPermissionEngineSetMode(t *testing.T) {
	e := NewPermissionEngine()

	e.SetMode(ModeAccept)
	if e.GetMode() != ModeAccept {
		t.Errorf("GetMode() = %v, want %v", e.GetMode(), ModeAccept)
	}

	e.SetMode(ModeDeny)
	if e.GetMode() != ModeDeny {
		t.Errorf("GetMode() = %v, want %v", e.GetMode(), ModeDeny)
	}
}

func TestPermissionEngineAddRule(t *testing.T) {
	e := NewPermissionEngine()

	rule := NewToolNameRule("Bash", false, "denied")
	e.AddRule(rule)

	e.SetMode(ModePrompt)

	ctx := context.Background()
	tc := &mockToolContext{workingDir: "/tmp", user: "test", interactive: true}

	result, err := e.Check(ctx, "Bash", map[string]any{}, tc)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "prompt" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "prompt")
	}
}

func TestPermissionEngineAlwaysAllow(t *testing.T) {
	e := NewPermissionEngine()

	// Add a rule that always allows Read
	rule := NewToolNameRule("Read", true, "allowed")
	e.AddAlwaysAllow(rule)

	ctx := context.Background()
	tc := &mockToolContext{workingDir: "/tmp", user: "test", interactive: true}

	result, err := e.Check(ctx, "Read", map[string]any{}, tc)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "allow" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "allow")
	}
}

func TestPermissionEngineAlwaysDeny(t *testing.T) {
	e := NewPermissionEngine()

	// Add a rule that always denies Bash
	rule := NewToolNameRule("Bash", false, "denied")
	e.AddAlwaysDeny(rule)

	ctx := context.Background()
	tc := &mockToolContext{workingDir: "/tmp", user: "test", interactive: true}

	result, err := e.Check(ctx, "Bash", map[string]any{}, tc)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "deny" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "deny")
	}
}

func TestPermissionEngineModeAccept(t *testing.T) {
	e := NewPermissionEngine()
	e.SetMode(ModeAccept)

	ctx := context.Background()
	tc := &mockToolContext{workingDir: "/tmp", user: "test", interactive: true}

	result, err := e.Check(ctx, "Bash", map[string]any{}, tc)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "allow" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "allow")
	}
}

func TestPermissionEngineModeDeny(t *testing.T) {
	e := NewPermissionEngine()
	e.SetMode(ModeDeny)

	ctx := context.Background()
	tc := &mockToolContext{workingDir: "/tmp", user: "test", interactive: true}

	result, err := e.Check(ctx, "Bash", map[string]any{}, tc)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "deny" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "deny")
	}
}

func TestDefaultEngine(t *testing.T) {
	// Test that DefaultEngine exists and works
	result, err := Check(context.Background(), "Test", map[string]any{}, nil)
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if result.Behavior != "prompt" {
		t.Errorf("Check() behavior = %v, want %v", result.Behavior, "prompt")
	}
}
