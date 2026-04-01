package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Aliases     []string
	Usage       string
	Examples    []string
	Execute     func(args []string) error
}

// Registry manages available commands
type Registry struct {
	mu        sync.RWMutex
	commands   map[string]*Command
	byAlias    map[string]*Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
		byAlias:  make(map[string]*Command),
	}
}

// Register registers a command
func (r *Registry) Register(cmd *Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		r.byAlias[alias] = cmd
	}
}

// Get returns a command by name or alias
func (r *Registry) Get(name string) (*Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cmd, ok := r.commands[name]; ok {
		return cmd, true
	}
	if cmd, ok := r.byAlias[name]; ok {
		return cmd, true
	}
	return nil, false
}

// List returns all registered commands
func (r *Registry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmds := make([]*Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// Execute executes a command by name
func (r *Registry) Execute(name string, args []string) error {
	cmd, ok := r.Get(name)
	if !ok {
		return fmt.Errorf("command not found: %s", name)
	}
	return cmd.Execute(args)
}

// DefaultRegistry is the global command registry
var DefaultRegistry = NewRegistry()

// Register is a convenience function using the default registry
func Register(cmd *Command) {
	DefaultRegistry.Register(cmd)
}

// Get is a convenience function using the default registry
func Get(name string) (*Command, bool) {
	return DefaultRegistry.Get(name)
}

// ListAll returns all registered commands
func ListAll() []*Command {
	return DefaultRegistry.List()
}

// Execute runs a command
func Execute(name string, args []string) error {
	return DefaultRegistry.Execute(name, args)
}

// ShellCommand executes a shell command
func ShellCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Bash executes a bash command
func Bash(command string) error {
	return ShellCommand("bash", "-c", command)
}

// Output executes a command and returns output
func Output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// OutputSlice executes a command and returns output lines
func OutputSlice(name string, args ...string) ([]string, error) {
	output, err := Output(name, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}

// HelpCommand creates a help command
func HelpCommand() *Command {
	return &Command{
		Name:        "help",
		Description: "Show help for a command",
		Aliases:     []string{"?"},
		Usage:       "help [command]",
		Examples:    []string{"help", "help agent"},
		Execute: func(args []string) error {
			if len(args) == 0 {
				fmt.Println("Available commands:")
				for _, cmd := range ListAll() {
					fmt.Printf("  %s - %s\n", cmd.Name, cmd.Description)
				}
				return nil
			}

			cmd, ok := Get(args[0])
			if !ok {
				return fmt.Errorf("command not found: %s", args[0])
			}

			fmt.Printf("Command: %s\n", cmd.Name)
			if len(cmd.Aliases) > 0 {
				fmt.Printf("Aliases: %s\n", strings.Join(cmd.Aliases, ", "))
			}
			fmt.Printf("Usage: %s\n", cmd.Usage)
			if len(cmd.Examples) > 0 {
				fmt.Println("Examples:")
				for _, ex := range cmd.Examples {
					fmt.Printf("  %s %s\n", cmd.Name, ex)
				}
			}
			return nil
		},
	}
}

func init() {
	Register(HelpCommand())
}
