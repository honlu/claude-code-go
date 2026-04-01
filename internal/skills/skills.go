package skills

import (
	"context"
	"fmt"
	"sync"
)

// Skill represents a reusable skill
type Skill interface {
	ID() string
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]any) (any, error)
}

// BaseSkill provides common skill functionality
type BaseSkill struct {
	id          string
	name        string
	description string
}

func (s *BaseSkill) ID() string   { return s.id }
func (s *BaseSkill) Name() string { return s.name }

// SkillRegistry manages skills
type SkillRegistry struct {
	mu     sync.RWMutex
	skills map[string]Skill
}

// NewRegistry creates a new skill registry
func NewRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]Skill),
	}
}

// Register registers a skill
func (r *SkillRegistry) Register(skill Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.skills[skill.ID()]; ok {
		return fmt.Errorf("skill %s is already registered", skill.ID())
	}
	r.skills[skill.ID()] = skill
	return nil
}

// Unregister unregisters a skill
func (r *SkillRegistry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.skills, id)
}

// Get returns a skill by ID
func (r *SkillRegistry) Get(id string) (Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[id]
	return s, ok
}

// List returns all skills
func (r *SkillRegistry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skills := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills
}

// Execute executes a skill by ID
func (r *SkillRegistry) Execute(ctx context.Context, id string, args map[string]any) (any, error) {
	skill, ok := r.Get(id)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", id)
	}
	return skill.Execute(ctx, args)
}

// DefaultRegistry is the global skill registry
var DefaultRegistry = NewRegistry()

// Register registers a skill with the default registry
func Register(skill Skill) error {
	return DefaultRegistry.Register(skill)
}

// Get returns a skill by ID from the default registry
func Get(id string) (Skill, bool) {
	return DefaultRegistry.Get(id)
}

// List returns all skills from the default registry
func List() []Skill {
	return DefaultRegistry.List()
}

// Execute executes a skill by ID from the default registry
func Execute(ctx context.Context, id string, args map[string]any) (any, error) {
	return DefaultRegistry.Execute(ctx, id, args)
}

// FunctionSkill is a skill implemented by a function
type FunctionSkill struct {
	BaseSkill
	handler func(context.Context, map[string]any) (any, error)
}

func (f *FunctionSkill) Description() string { return f.description }

func (f *FunctionSkill) Execute(ctx context.Context, args map[string]any) (any, error) {
	return f.handler(ctx, args)
}

// RegisterFunc registers a function as a skill
func RegisterFunc(id, name, description string, handler func(context.Context, map[string]any) (any, error)) error {
	return Register(&FunctionSkill{
		BaseSkill: BaseSkill{
			id:          id,
			name:        name,
			description: description,
		},
		handler: handler,
	})
}
