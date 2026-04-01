package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Getenv gets environment variable with default value
func Getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ExpandEnv expands environment variables in a string
func ExpandEnv(s string) string {
	return os.Expand(s, func(key string) string {
		return os.Getenv(key)
	})
}

// AbsPath returns the absolute path of a file
func AbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return abs, nil
}

// EnsureDir ensures a directory exists
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// ReadFile reads a file and returns its contents
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// WriteFile writes content to a file
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GenerateID generates a random ID
func GenerateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return prefix + hex.EncodeToString(bytes)
}

// Retry retries a function until it succeeds or maxAttempts is reached
func Retry(maxAttempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}

// CommandExists checks if a command exists
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// TrimSpace trims whitespace from a string
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// SplitLines splits a string into lines
func SplitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}

// Contains checks if a slice contains a value
func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Unique returns unique values from a slice
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// MatchPattern checks if a string matches a pattern (simple glob-like matching)
func MatchPattern(pattern, text string) bool {
	regex := regexp.MustCompile("^" + regexp.QuoteMeta(pattern) + "$")
	return regex.MatchString(text)
}

// SanitizeFilename sanitizes a filename
func SanitizeFilename(name string) string {
	// Replace invalid characters
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	return reg.ReplaceAllString(name, "_")
}

// Truncate truncates a string to max length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Timestamp returns the current timestamp as a string
func Timestamp() string {
	return time.Now().Format(time.RFC3339)
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Duration formats a duration as human-readable string
func Duration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
