package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ai/claude-code/internal/bootstrap"
	"github.com/ai/claude-code/internal/cli"
)

func main() {
	// Initialize bootstrap state
	bootstrap.Init()

	// Create root command
	rootCmd := cli.NewRootCommand()

	// Execute
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
