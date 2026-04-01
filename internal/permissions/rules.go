package permissions

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

// ToolNameRule matches tools by name pattern
type ToolNameRule struct {
	name       string
	pattern    *regexp.Regexp
	allow      bool
	reason     string
}

// NewToolNameRule creates a new tool name matching rule
func NewToolNameRule(pattern string, allow bool, reason string) *ToolNameRule {
	return &ToolNameRule{
		name:   "ToolNameRule",
		pattern: regexp.MustCompile(pattern),
		allow:  allow,
		reason: reason,
	}
}

// Name returns the rule name
func (r *ToolNameRule) Name() string {
	return r.name
}

// Check checks if the tool name matches the pattern
func (r *ToolNameRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	if r.pattern.MatchString(toolName) {
		return r.allow, r.reason
	}
	return true, ""
}

// PathRule matches file paths by pattern
type PathRule struct {
	name        string
	pathPattern *regexp.Regexp
	allow       bool
	reason      string
}

// NewPathRule creates a new path matching rule
func NewPathRule(pattern string, allow bool, reason string) *PathRule {
	return &PathRule{
		name:        "PathRule",
		pathPattern: regexp.MustCompile(pattern),
		allow:       allow,
		reason:      reason,
	}
}

// Name returns the rule name
func (r *PathRule) Name() string {
	return r.name
}

// Check checks if the path matches the pattern
func (r *PathRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	path, ok := args["path"].(string)
	if !ok {
		return true, ""
	}

	if r.pathPattern.MatchString(path) {
		return r.allow, r.reason
	}
	return true, ""
}

// ReadOnlyRule denies write operations
type ReadOnlyRule struct {
	name string
}

// NewReadOnlyRule creates a new read-only rule
func NewReadOnlyRule() *ReadOnlyRule {
	return &ReadOnlyRule{name: "ReadOnlyRule"}
}

// Name returns the rule name
func (r *ReadOnlyRule) Name() string {
	return r.name
}

// Check checks if the tool is a write tool
func (r *ReadOnlyRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	writeTools := []string{"Write", "Edit", "Bash"}
	for _, wt := range writeTools {
		if toolName == wt {
			return false, "Read-only mode: write operations are denied"
		}
	}
	return true, ""
}

// SafePathsRule allows only safe paths
type SafePathsRule struct {
	name         string
	allowedPaths []string
	deniedPaths  []string
}

// NewSafePathsRule creates a rule for safe path checking
func NewSafePathsRule(allowed, denied []string) *SafePathsRule {
	return &SafePathsRule{
		name:         "SafePathsRule",
		allowedPaths: allowed,
		deniedPaths:  denied,
	}
}

// Name returns the rule name
func (r *SafePathsRule) Name() string {
	return r.name
}

// Check checks if the path is safe
func (r *SafePathsRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	path, ok := args["path"].(string)
	if !ok {
		return true, ""
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, "Could not resolve path"
	}

	// Check denied patterns first
	for _, denied := range r.deniedPaths {
		if strings.HasPrefix(absPath, denied) {
			return false, "Path matches denied pattern: " + denied
		}
	}

	// If allowed patterns are specified, check them
	if len(r.allowedPaths) > 0 {
		for _, allowed := range r.allowedPaths {
			if strings.HasPrefix(absPath, allowed) {
				return true, "Path matches allowed pattern: " + allowed
			}
		}
		return false, "Path not in allowed directories"
	}

	return true, ""
}

// BashCommandRule checks bash commands
type BashCommandRule struct {
	name           string
	allowedCmds    []*regexp.Regexp
	deniedCmds     []*regexp.Regexp
	deniedPatterns []*regexp.Regexp
}

// NewBashCommandRule creates a rule for bash command checking
func NewBashCommandRule() *BashCommandRule {
	return &BashCommandRule{
		name:           "BashCommandRule",
		allowedCmds:    []*regexp.Regexp{},
		deniedCmds:     []*regexp.Regexp{},
		deniedPatterns: []*regexp.Regexp{},
	}
}

// AddDeniedCommand adds a denied command pattern
func (r *BashCommandRule) AddDeniedCommand(pattern string) {
	r.deniedCmds = append(r.deniedCmds, regexp.MustCompile(pattern))
}

// AddDeniedPattern adds a denied pattern
func (r *BashCommandRule) AddDeniedPattern(pattern string) {
	r.deniedPatterns = append(r.deniedPatterns, regexp.MustCompile(pattern))
}

// Name returns the rule name
func (r *BashCommandRule) Name() string {
	return r.name
}

// Check checks if the bash command is allowed
func (r *BashCommandRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	if toolName != "Bash" {
		return true, ""
	}

	cmd, ok := args["command"].(string)
	if !ok {
		return true, ""
	}

	// Check denied commands
	for _, denied := range r.deniedCmds {
		if denied.MatchString(cmd) {
			return false, "Command matches denied pattern"
		}
	}

	// Check denied patterns
	for _, denied := range r.deniedPatterns {
		if denied.MatchString(cmd) {
			return false, "Command contains denied pattern"
		}
	}

	return true, ""
}

// CompositeRule combines multiple rules
type CompositeRule struct {
	name  string
	rules []PermissionRule
	mode  string // "all" or "any"
}

// NewCompositeRule creates a rule that combines multiple rules
func NewCompositeRule(name string, mode string, rules ...PermissionRule) *CompositeRule {
	return &CompositeRule{
		name:  name,
		rules: rules,
		mode:  mode,
	}
}

// Name returns the rule name
func (r *CompositeRule) Name() string {
	return r.name
}

// Check evaluates all rules
func (r *CompositeRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
	for _, rule := range r.rules {
		allowed, msg := rule.Check(ctx, toolName, args)
		if r.mode == "all" {
			if !allowed {
				return false, msg
			}
		} else {
			if allowed {
				return true, msg
			}
		}
	}

	if r.mode == "all" {
		return true, ""
	}
	return false, "No rules matched"
}
