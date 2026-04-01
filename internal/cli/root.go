package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ai/claude-code/internal/bootstrap"
	"github.com/ai/claude-code/internal/tools"
	_ "github.com/ai/claude-code/internal/tools/bash"
	_ "github.com/ai/claude-code/internal/tools/edit"
	_ "github.com/ai/claude-code/internal/tools/glob"
	_ "github.com/ai/claude-code/internal/tools/grep"
	_ "github.com/ai/claude-code/internal/tools/read"
	_ "github.com/ai/claude-code/internal/tools/write"
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "claude",
		Short: "Claude Code - AI-powered coding assistant",
		Long: `Claude Code is an agentic coding tool that lives in your terminal,
understands your codebase, and helps you code faster by executing
routine tasks, explaining complex code, and handling git workflows.

Learn more at https://claude.com/product/claude-code`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			bootstrap.Init()
			return nil
		},
	}

	// Add global flags
	root.PersistentFlags().StringP("model", "m", "", "Model to use")
	root.PersistentFlags().Bool("print", false, "Print response only, no tool execution")
	root.PersistentFlags().String("permission-mode", "", "Permission mode (accept, deny, etc.)")
	root.PersistentFlags().Bool("dangerously-skip-permissions", false, "Skip permission prompts")

	// Add subcommands
	root.AddCommand(
		NewAgentCommand(),
		NewVersionCommand(),
		NewToolsCommand(),
	)

	return root
}

// NewAgentCommand returns the main agent interaction command
func NewAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [message]",
		Short: "Start an interactive session with Claude",
		Long: `Start an interactive session with Claude Code.

If a message is provided as an argument, it will be sent to Claude
without entering interactive mode.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			model, _ := cmd.Flags().GetString("model")
			printMode, _ := cmd.Flags().GetBool("print")

			if len(args) > 0 {
				return runSingleMessage(ctx, args[0], model, printMode)
			}
			return runAgent(ctx, model, printMode)
		},
	}

	cmd.Flags().Bool("print", false, "Print response only")
	cmd.Flags().BoolP("continue", "c", false, "Continue the previous session")
	cmd.Flags().Bool("resume", false, "Resume a previous session")
	cmd.Flags().String("model", "", "Model to use")

	return cmd
}

// runAgent is the main agent loop
func runAgent(ctx context.Context, model string, printMode bool) error {
	repl := NewREPL(getModelOrDefault(model), printMode)
	return repl.Run(ctx)
}

// runSingleMessage processes a single message and exits
func runSingleMessage(ctx context.Context, message, model string, printMode bool) error {
	return RunSingleMessage(ctx, message, getModelOrDefault(model), printMode)
}

// NewToolsCommand returns the tools list command
func NewToolsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "List available tools",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			listTools()
		},
	}
}

// listTools prints all available tools
func listTools() {
	registry := tools.GetRegistry()
	toolList := registry.List()

	fmt.Printf("Available tools (%d):\n\n", len(toolList))

	for _, tool := range toolList {
		fmt.Printf("  %s - %s\n", tool.Name(), tool.Description(nil))
	}
}

// NewVersionCommand returns the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Claude Code Go")
			fmt.Println("Version: 0.1.0-dev")
			fmt.Printf("Go Version: %s\n", getGoVersion())
			fmt.Printf("Working Dir: %s\n", bootstrap.GetCwd())
		},
	}
}

// getModelOrDefault returns the model or default
func getModelOrDefault(model string) string {
	if model != "" {
		return model
	}
	return "claude-sonnet-4-6" // default
}

// getGoVersion returns the Go version
func getGoVersion() string {
	return "unknown"
}

func init() {
	// Set working directory early
	bootstrap.Init()
}
