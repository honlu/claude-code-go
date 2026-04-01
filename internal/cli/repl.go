package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ai/claude-code/internal/api"
	"github.com/ai/claude-code/internal/bootstrap"
	"github.com/ai/claude-code/internal/query"
)

// REPL is the interactive Read-Eval-Print-Loop
type REPL struct {
	engine   *query.QueryEngine
	reader   *bufio.Reader
	history  []string
	model    string
	printMode bool
}

// NewREPL creates a new REPL instance
func NewREPL(model string, printMode bool) *REPL {
	return &REPL{
		reader:    bufio.NewReader(os.Stdin),
		history:   []string{},
		model:     model,
		printMode: printMode,
	}
}

// Run starts the REPL
func (r *REPL) Run(ctx context.Context) error {
	// Initialize query engine
	r.engine = query.NewQueryEngine(
		query.WithModel(r.model),
	)

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Claude Code Go - Interactive Mode")
	fmt.Printf("Model: %s\n", r.model)
	fmt.Printf("Working directory: %s\n", bootstrap.GetCwd())
	fmt.Println()
	fmt.Println("Type your message and press Enter to send.")
	fmt.Println("Use /exit or Ctrl+C to quit.")
	fmt.Println("Use /clear to clear conversation history.")
	fmt.Println("Use /model <name> to change the model.")
	fmt.Println()

	// Main loop
	for {
		select {
		case <-sigCh:
			fmt.Println("\nExiting...")
			return nil
		default:
			input, err := r.readLine()
			if err != nil {
				return err
			}

			if input == "" {
				continue
			}

			// Add to history
			r.history = append(r.history, input)

			// Handle special commands
			if strings.HasPrefix(input, "/") {
				if err := r.handleCommand(input); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				continue
			}

			// Process query
			if err := r.sendToClaude(ctx, input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

// readLine reads a line from stdin
func (r *REPL) readLine() (string, error) {
	fmt.Print("> ")
	input, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// handleCommand handles special commands
func (r *REPL) handleCommand(input string) error {
	parts := strings.SplitN(input, " ", 2)
	cmd := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	switch cmd {
	case "/exit", "/quit":
		fmt.Println("Exiting...")
		os.Exit(0)
	case "/clear":
		r.engine.ClearMessages()
		fmt.Println("Conversation cleared.")
	case "/model":
		if args == "" {
			fmt.Printf("Current model: %s\n", r.engine.GetModel())
		} else {
			r.engine.SetModel(args)
			r.model = args
			fmt.Printf("Model changed to: %s\n", args)
		}
	case "/history":
		for i, h := range r.history {
			fmt.Printf("%d: %s\n", i+1, h)
		}
	case "/help":
		r.printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		r.printHelp()
	}
	return nil
}

// printHelp prints help information
func (r *REPL) printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  /exit, /quit - Exit the REPL")
	fmt.Println("  /clear - Clear conversation history")
	fmt.Println("  /model [name] - Show or change model")
	fmt.Println("  /history - Show command history")
	fmt.Println("  /help - Show this help")
}

// sendToClaude sends input to Claude and displays the response
func (r *REPL) sendToClaude(ctx context.Context, input string) error {
	if r.printMode {
		// Just print what would be sent
		fmt.Printf("[Would send to Claude]: %s\n", input)
		return nil
	}

	fmt.Println("[Claude]: ", startThinkingAnimation(ctx))

	// Stream the response
	var response strings.Builder
	err := r.engine.QueryStream(ctx, input, func(chunk string) error {
		response.WriteString(chunk)
		// Print in real-time
		fmt.Print(chunk)
		return nil
	})

	fmt.Println() // New line after response

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return nil
}

// startThinkingAnimation shows thinking indicator
func startThinkingAnimation(ctx context.Context) string {
	return "thinking..."
}

// RunSingleMessage runs a single message and returns
func RunSingleMessage(ctx context.Context, message, model string, printMode bool) error {
	engine := query.NewQueryEngine(
		query.WithModel(model),
	)

	if printMode {
		fmt.Printf("[Would send to Claude]: %s\n", message)
		return nil
	}

	fmt.Println("[Claude]: thinking...")

	var response strings.Builder
	err := engine.QueryStream(ctx, message, func(chunk string) error {
		response.WriteString(chunk)
		fmt.Print(chunk)
		return nil
	})

	fmt.Println()

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return nil
}

// InitAPI initializes the API client with the given options
func InitAPI(provider api.Provider, opts ...api.ClientOption) (*api.Client, error) {
	return api.NewClient(provider, opts...)
}
