package permissions

import (
	"context"
	"fmt"
	"sync"
)

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Behavior      string            // "allow", "deny", "prompt"
	Message       string            // Human-readable message
	UpdatedInput  map[string]any    // Modified input after permission check
	SkipApproved  bool              // Skip further approval if true
}

// PermissionRule defines a rule for permission checking
type PermissionRule interface {
	// Name returns the rule name
	Name() string
	// Check evaluates the rule against the given context
	Check(ctx context.Context, toolName string, args map[string]any) (bool, string)
}

// ToolContext provides context for permission checks
type ToolContext interface {
	GetWorkingDirectory() string
	GetCurrentUser() string
	IsInteractive() bool
}

// PermissionEngine handles permission checking
type PermissionEngine struct {
	alwaysAllow  []PermissionRule
	alwaysDeny   []PermissionRule
	alwaysAsk    []PermissionRule
	rules        []PermissionRule
	mode         PermissionMode
	mu           sync.RWMutex
}

// PermissionMode defines how to handle permission checks
type PermissionMode string

const (
	ModeAccept   PermissionMode = "accept"   // Auto-accept all
	ModeDeny     PermissionMode = "deny"     // Auto-deny all
	ModePrompt   PermissionMode = "prompt"   // Always prompt
	ModeApprove  PermissionMode = "approve" // Approve once
)

// NewPermissionEngine creates a new permission engine
func NewPermissionEngine() *PermissionEngine {
	return &PermissionEngine{
		alwaysAllow: []PermissionRule{},
		alwaysDeny:  []PermissionRule{},
		alwaysAsk:   []PermissionRule{},
		rules:       []PermissionRule{},
		mode:        ModePrompt,
	}
}

// SetMode sets the permission mode
func (e *PermissionEngine) SetMode(mode PermissionMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mode = mode
}

// GetMode gets the current permission mode
func (e *PermissionEngine) GetMode() PermissionMode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.mode
}

// AddRule adds a rule to the engine
func (e *PermissionEngine) AddRule(rule PermissionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// AddAlwaysAllow adds a rule to the always-allow list
func (e *PermissionEngine) AddAlwaysAllow(rule PermissionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alwaysAllow = append(e.alwaysAllow, rule)
}

// AddAlwaysDeny adds a rule to the always-deny list
func (e *PermissionEngine) AddAlwaysDeny(rule PermissionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alwaysDeny = append(e.alwaysDeny, rule)
}

// AddAlwaysAsk adds a rule to the always-ask list
func (e *PermissionEngine) AddAlwaysAsk(rule PermissionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alwaysAsk = append(e.alwaysAsk, rule)
}

// Check evaluates permission for a tool call
func (e *PermissionEngine) Check(ctx context.Context, toolName string, args map[string]any, tc ToolContext) (PermissionResult, error) {
	e.mu.RLock()
	mode := e.mode
	e.mu.RUnlock()

	// Check always-deny first
	for _, rule := range e.alwaysDeny {
		if allowed, _ := rule.Check(ctx, toolName, args); !allowed {
			return PermissionResult{
				Behavior: "deny",
				Message:  fmt.Sprintf("Denied by rule: %s", rule.Name()),
			}, nil
		}
	}

	// Check always-allow
	for _, rule := range e.alwaysAllow {
		if allowed, msg := rule.Check(ctx, toolName, args); allowed {
			return PermissionResult{
				Behavior: "allow",
				Message:  msg,
			}, nil
		}
	}

	// Auto-accept if mode is accept
	if mode == ModeAccept {
		return PermissionResult{
			Behavior: "allow",
			Message:  "Auto-accepted (accept mode)",
		}, nil
	}

	// Auto-deny if mode is deny
	if mode == ModeDeny {
		return PermissionResult{
			Behavior: "deny",
			Message:  "Auto-denied (deny mode)",
		}, nil
	}

	// Check always-ask rules
	for _, rule := range e.alwaysAsk {
		if allowed, msg := rule.Check(ctx, toolName, args); !allowed {
			return PermissionResult{
				Behavior: "prompt",
				Message:  msg,
			}, nil
		}
	}

	// Check custom rules
	for _, rule := range e.rules {
		if allowed, msg := rule.Check(ctx, toolName, args); !allowed {
			return PermissionResult{
				Behavior: "prompt",
				Message:  fmt.Sprintf("Rule '%s' requires approval: %s", rule.Name(), msg),
			}, nil
		}
	}

	// Default: prompt for approval
	return PermissionResult{
		Behavior: "prompt",
		Message:  fmt.Sprintf("Tool '%s' requires approval", toolName),
	}, nil
}

// LoadRules loads rules from a rules file (future extension)
func (e *PermissionEngine) LoadRules(rules []PermissionRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rules...)
}

// DefaultEngine is the global default permission engine
var DefaultEngine = NewPermissionEngine()

// Check is a convenience function using the default engine
func Check(ctx context.Context, toolName string, args map[string]any, tc ToolContext) (PermissionResult, error) {
	return DefaultEngine.Check(ctx, toolName, args, tc)
}
