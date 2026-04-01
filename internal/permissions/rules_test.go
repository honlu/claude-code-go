package permissions

import (
	"context"
	"testing"
)

func TestToolNameRule(t *testing.T) {
	rule := NewToolNameRule("Bash", false, "denied")

	ctx := context.Background()

	// Test matching tool
	allowed, msg := rule.Check(ctx, "Bash", map[string]any{})
	if allowed {
		t.Error("Check() allowed = true, want false for matching tool")
	}
	if msg != "denied" {
		t.Errorf("Check() msg = %v, want %v", msg, "denied")
	}

	// Test non-matching tool
	allowed, msg = rule.Check(ctx, "Read", map[string]any{})
	if !allowed {
		t.Error("Check() allowed = false, want true for non-matching tool")
	}
}

func TestPathRule(t *testing.T) {
	rule := NewPathRule("/etc/.*", false, "denied")

	ctx := context.Background()

	// Test matching path
	allowed, _ := rule.Check(ctx, "Read", map[string]any{"path": "/etc/passwd"})
	if allowed {
		t.Error("Check() allowed = true, want false for matching path")
	}

	// Test non-matching path
	allowed, _ = rule.Check(ctx, "Read", map[string]any{"path": "/home/user/file.txt"})
	if !allowed {
		t.Error("Check() allowed = false, want true for non-matching path")
	}

	// Test missing path
	allowed, _ = rule.Check(ctx, "Read", map[string]any{})
	if !allowed {
		t.Error("Check() allowed = false, want true for missing path")
	}
}

func TestReadOnlyRule(t *testing.T) {
	rule := NewReadOnlyRule()

	ctx := context.Background()

	// Test write tools
	for _, tool := range []string{"Write", "Edit", "Bash"} {
		allowed, _ := rule.Check(ctx, tool, map[string]any{})
		if allowed {
			t.Errorf("Check() allowed = true for %s, want false", tool)
		}
	}

	// Test read tool
	allowed, _ := rule.Check(ctx, "Read", map[string]any{})
	if !allowed {
		t.Error("Check() allowed = false for Read, want true")
	}
}

func TestSafePathsRule(t *testing.T) {
	allowedPaths := []string{"/home/user/"}
	deniedPaths := []string{"/etc/", "/root/"}

	rule := NewSafePathsRule(allowedPaths, deniedPaths)

	ctx := context.Background()

	// Test denied path
	allowed, _ := rule.Check(ctx, "Read", map[string]any{"path": "/etc/passwd"})
	if allowed {
		t.Error("Check() allowed = true for /etc/passwd, want false")
	}

	// Test allowed path
	allowed, _ = rule.Check(ctx, "Read", map[string]any{"path": "/home/user/file.txt"})
	if !allowed {
		t.Error("Check() allowed = false for /home/user/file.txt, want true")
	}

	// Test unspecified path with allowed list
	allowed, _ = rule.Check(ctx, "Read", map[string]any{"path": "/var/log"})
	if allowed {
		t.Error("Check() allowed = true for /var/log (not in allowed list), want false")
	}
}

func TestBashCommandRule(t *testing.T) {
	rule := NewBashCommandRule()
	rule.AddDeniedCommand("^rm -rf")
	rule.AddDeniedPattern(";.*rm -rf")

	ctx := context.Background()

	// Test denied command
	allowed, _ := rule.Check(ctx, "Bash", map[string]any{"command": "rm -rf /"})
	if allowed {
		t.Error("Check() allowed = true for 'rm -rf /', want false")
	}

	// Test allowed command
	allowed, _ = rule.Check(ctx, "Bash", map[string]any{"command": "ls -la"})
	if !allowed {
		t.Error("Check() allowed = false for 'ls -la', want true")
	}

	// Test non-Bash tool
	allowed, _ = rule.Check(ctx, "Read", map[string]any{"command": "rm -rf"})
	if !allowed {
		t.Error("Check() allowed = false for Read tool, want true (should skip check)")
	}
}

func TestCompositeRule(t *testing.T) {
	// Create composite rule with "all" mode - all must match
	rule1 := NewToolNameRule("Bash", true, "")
	rule2 := NewToolNameRule(".*", true, "")

	composite := NewCompositeRule("composite", "all", rule1, rule2)

	ctx := context.Background()

	// Test with matching tool
	allowed, _ := composite.Check(ctx, "Bash", map[string]any{})
	if !allowed {
		t.Error("Composite 'all' mode: allowed = false, want true")
	}
}

func TestCompositeRuleAnyMode(t *testing.T) {
	// Create composite rule with "any" mode - any can match
	rule1 := NewToolNameRule("Bash", false, "denied")
	rule2 := NewToolNameRule("Read", true, "")

	composite := NewCompositeRule("composite", "any", rule1, rule2)

	ctx := context.Background()

	// Test - rule2 matches
	allowed, _ := composite.Check(ctx, "Read", map[string]any{})
	if !allowed {
		t.Error("Composite 'any' mode: allowed = false, want true")
	}
}
