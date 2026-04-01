package bootstrap

import (
	"os"
	"path/filepath"
)

// Init initializes the bootstrap state
func Init() {
	// Set current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}
	SetCwd(cwd)

	// Find project root (contains .git, go.mod, package.json, etc.)
	projectRoot := findProjectRoot(cwd)
	SetProjectRoot(projectRoot)
}

// findProjectRoot looks for common project markers to determine the project root
func findProjectRoot(cwd string) string {
	// Common markers that indicate a project root
	markers := []string{
		".git",
		"go.mod",
		"package.json",
		"Cargo.toml",
		"pyproject.toml",
		"requirements.txt",
		"Pipfile",
		"Makefile",
		"CLAUDE.md",
	}

	dir := cwd
	for {
		for _, marker := range markers {
			path := filepath.Join(dir, marker)
			if _, err := os.Stat(path); err == nil {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return cwd
		}
		dir = parent
	}
}
