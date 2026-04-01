package keybindings

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Keystroke represents a key combination
type Keystroke struct {
	Keys     []string
	Chord    bool // true if this is a multi-key chord
	Alt      bool
	Control  bool
	Shift    bool
	Meta     bool
}

// ParseKeystroke parses a keystroke string
func ParseKeystroke(s string) (*Keystroke, error) {
	keys := strings.Split(s, "-")
	ks := &Keystroke{
		Keys: make([]string, 0),
	}

	for i, key := range keys {
		key = strings.ToLower(key)
		switch key {
		case "alt":
			ks.Alt = true
		case "ctrl", "control":
			ks.Control = true
		case "shift":
			ks.Shift = true
		case "meta", "cmd", "command":
			ks.Meta = true
		default:
			if i == 0 && len(keys) > 1 {
				ks.Chord = true
			}
			ks.Keys = append(ks.Keys, key)
		}
	}

	if len(ks.Keys) == 0 {
		return nil, fmt.Errorf("empty keystroke")
	}

	return ks, nil
}

// String returns the string representation
func (ks *Keystroke) String() string {
	var parts []string
	if ks.Control {
		parts = append(parts, "ctrl")
	}
	if ks.Alt {
		parts = append(parts, "alt")
	}
	if ks.Shift {
		parts = append(parts, "shift")
	}
	if ks.Meta {
		parts = append(parts, "meta")
	}
	parts = append(parts, ks.Keys...)
	return strings.Join(parts, "-")
}

// Matches checks if this keystroke matches another
func (ks *Keystroke) Matches(other *Keystroke) bool {
	if ks.Chord != other.Chord {
		return false
	}
	if len(ks.Keys) != len(other.Keys) {
		return false
	}
	for i := range ks.Keys {
		if ks.Keys[i] != other.Keys[i] {
			return false
		}
	}
	return ks.Alt == other.Alt && ks.Control == other.Control &&
		ks.Shift == other.Shift && ks.Meta == other.Meta
}

// Action represents a keybinding action
type Action struct {
	Name        string
	Description string
	Keystrokes  []*Keystroke
	Handler     func() error
	Context     string // context name for conditional bindings
}

// Binding represents a keybinding
type Binding struct {
	Keystroke *Keystroke
	Action    *Action
	Priority  int // higher priority wins
}

// Resolver resolves keybindings to actions
type Resolver struct {
	mu       sync.RWMutex
	bindings map[string][]*Binding // keyed by context name
	defaultCtx string
}

// NewResolver creates a new keybinding resolver
func NewResolver() *Resolver {
	return &Resolver{
		bindings:  make(map[string][]*Binding),
		defaultCtx: "global",
	}
}

// SetDefaultContext sets the default context
func (r *Resolver) SetDefaultContext(ctx string) {
	r.defaultCtx = ctx
}

// Register registers an action
func (r *Resolver) Register(action *Action, keystrokes ...*Keystroke) {
	r.mu.Lock()
	defer r.mu.Unlock()

	action.Keystrokes = keystrokes
	for _, ks := range keystrokes {
		binding := &Binding{
			Keystroke: ks,
			Action:    action,
			Priority:  0,
		}
		r.bindings[r.defaultCtx] = append(r.bindings[r.defaultCtx], binding)
	}
}

// RegisterContext registers a context-specific action
func (r *Resolver) RegisterContext(ctx string, action *Action, keystrokes ...*Keystroke) {
	r.mu.Lock()
	defer r.mu.Unlock()

	action.Keystrokes = keystrokes
	for _, ks := range keystrokes {
		binding := &Binding{
			Keystroke: ks,
			Action:    action,
			Priority:  1, // context bindings have higher priority
		}
		r.bindings[ctx] = append(r.bindings[ctx], binding)
	}
}

// Resolve resolves a keystroke to an action
func (r *Resolver) Resolve(ks *Keystroke) (*Action, bool) {
	return r.ResolveContext(r.defaultCtx, ks)
}

// ResolveContext resolves a keystroke in a specific context
func (r *Resolver) ResolveContext(ctx string, ks *Keystroke) (*Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check context-specific bindings first
	bindings, ok := r.bindings[ctx]
	if ok {
		for _, binding := range bindings {
			if binding.Keystroke.Matches(ks) {
				return binding.Action, true
			}
		}
	}

	// Check global bindings
	global, ok := r.bindings["global"]
	if ok && ctx != "global" {
		for _, binding := range global {
			if binding.Keystroke.Matches(ks) {
				return binding.Action, true
			}
		}
	}

	return nil, false
}

// List returns all bindings for a context
func (r *Resolver) List(ctx string) []*Binding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bindings := r.bindings[ctx]
	result := make([]*Binding, len(bindings))
	copy(result, bindings)
	return result
}

// Clear clears all bindings
func (r *Resolver) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bindings = make(map[string][]*Binding)
}

// DefaultResolver is the global keybinding resolver
var DefaultResolver = NewResolver()

// Register registers an action with the default resolver
func Register(action *Action, keystrokes ...*Keystroke) {
	DefaultResolver.Register(action, keystrokes...)
}

// Resolve resolves a keystroke
func Resolve(ks *Keystroke) (*Action, bool) {
	return DefaultResolver.Resolve(ks)
}

// Parse is a convenience function to parse and resolve
func Parse(s string) (*Action, bool) {
	ks, err := ParseKeystroke(s)
	if err != nil {
		return nil, false
	}
	return Resolve(ks)
}

// MatchBinding represents a parsed binding
type MatchBinding struct {
	Pattern *regexp.Regexp
	Action  string
}

// Parser parses keybinding strings
type Parser struct {
	bindings []*MatchBinding
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// AddBinding adds a binding pattern
func (p *Parser) AddBinding(pattern, action string) error {
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return err
	}
	p.bindings = append(p.bindings, &MatchBinding{
		Pattern: re,
		Action:  action,
	})
	return nil
}

// Parse parses a keystroke string against registered patterns
func (p *Parser) Parse(s string) (string, bool) {
	for _, b := range p.bindings {
		if b.Pattern.MatchString(s) {
			return b.Action, true
		}
	}
	return "", false
}
