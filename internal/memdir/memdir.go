package memdir

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
)

// MemoryDirectory represents an in-memory directory structure
type MemoryDirectory struct {
	mu     sync.RWMutex
	nodes  map[string]*Node
	root   *Node
}

// Node represents a node in the memory directory
type Node struct {
	Name     string
	IsDir    bool
	Content  []byte
	Children []*Node
	Parent   *Node
	Meta     map[string]any
}

// New creates a new memory directory
func New() *MemoryDirectory {
	root := &Node{
		Name:     "/",
		IsDir:    true,
		Children: make([]*Node, 0),
		Meta:     make(map[string]any),
	}

	return &MemoryDirectory{
		nodes: make(map[string]*Node),
		root:  root,
	}
}

// Get returns a node by path
func (md *MemoryDirectory) Get(path string) (*Node, bool) {
	md.mu.RLock()
	defer md.mu.RUnlock()
	node, ok := md.nodes[path]
	return node, ok
}

// Set sets a node at a path
func (md *MemoryDirectory) Set(path string, node *Node) {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.nodes[path] = node
}

// Mkdir creates a directory
func (md *MemoryDirectory) Mkdir(path string) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	if _, ok := md.nodes[path]; ok {
		return fmt.Errorf("path already exists: %s", path)
	}

	// Create parent if needed
	parentPath := filepath.Dir(path)
	parent, ok := md.nodes[parentPath]
	if !ok {
		return fmt.Errorf("parent directory does not exist: %s", parentPath)
	}

	node := &Node{
		Name:     filepath.Base(path),
		IsDir:    true,
		Children: make([]*Node, 0),
		Parent:   parent,
		Meta:     make(map[string]any),
	}

	md.nodes[path] = node
	parent.Children = append(parent.Children, node)
	return nil
}

// WriteFile writes a file
func (md *MemoryDirectory) WriteFile(path string, content []byte) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	parentPath := filepath.Dir(path)
	parent, ok := md.nodes[parentPath]
	if !ok {
		return fmt.Errorf("parent directory does not exist: %s", parentPath)
	}

	node := &Node{
		Name:    filepath.Base(path),
		IsDir:   false,
		Content: content,
		Parent:  parent,
		Meta:    make(map[string]any),
	}

	md.nodes[path] = node
	parent.Children = append(parent.Children, node)
	return nil
}

// ReadFile reads a file
func (md *MemoryDirectory) ReadFile(path string) ([]byte, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	node, ok := md.nodes[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if node.IsDir {
		return nil, fmt.Errorf("path is a directory: %s", path)
	}
	return node.Content, nil
}

// Remove removes a path
func (md *MemoryDirectory) Remove(path string) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	node, ok := md.nodes[path]
	if !ok {
		return fmt.Errorf("path not found: %s", path)
	}

	// Remove from parent's children
	if node.Parent != nil {
		children := node.Parent.Children
		for i, child := range children {
			if child == node {
				node.Parent.Children = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	delete(md.nodes, path)
	return nil
}

// List lists a directory
func (md *MemoryDirectory) List(path string) ([]*Node, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	node, ok := md.nodes[path]
	if !ok {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if !node.IsDir {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	result := make([]*Node, len(node.Children))
	copy(result, node.Children)
	return result, nil
}

// Exists checks if a path exists
func (md *MemoryDirectory) Exists(path string) bool {
	md.mu.RLock()
	defer md.mu.RUnlock()
	_, ok := md.nodes[path]
	return ok
}

// SetMeta sets metadata on a node
func (md *MemoryDirectory) SetMeta(path string, key string, value any) error {
	md.mu.Lock()
	defer md.mu.Unlock()

	node, ok := md.nodes[path]
	if !ok {
		return fmt.Errorf("path not found: %s", path)
	}
	node.Meta[key] = value
	return nil
}

// GetMeta gets metadata from a node
func (md *MemoryDirectory) GetMeta(path, key string) (any, bool) {
	md.mu.RLock()
	defer md.mu.RUnlock()

	node, ok := md.nodes[path]
	if !ok {
		return nil, false
	}
	v, ok := node.Meta[key]
	return v, ok
}

// MarshalJSON serializes the memory directory to JSON
func (md *MemoryDirectory) MarshalJSON() ([]byte, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()
	return json.Marshal(md.nodes)
}

// DefaultMemoryDirectory is the global memory directory
var DefaultMemoryDirectory = New()

// Mkdir is a convenience function using the default directory
func Mkdir(path string) error {
	return DefaultMemoryDirectory.Mkdir(path)
}

// WriteFile is a convenience function using the default directory
func WriteFile(path string, content []byte) error {
	return DefaultMemoryDirectory.WriteFile(path, content)
}

// ReadFile is a convenience function using the default directory
func ReadFile(path string) ([]byte, error) {
	return DefaultMemoryDirectory.ReadFile(path)
}
