package bootstrap

import (
	"sync"
	"testing"
	"time"
)

func TestGetSetSessionId(t *testing.T) {
	// Reset state for test
	globalState.mu.Lock()
	globalState.sessionId = ""
	globalState.mu.Unlock()

	// Test SetSessionId and GetSessionId
	SetSessionId("test-session-123")
	if got := GetSessionId(); got != "test-session-123" {
		t.Errorf("GetSessionId() = %v, want %v", got, "test-session-123")
	}
}

func TestGetSetCwd(t *testing.T) {
	SetCwd("/tmp/test")
	if got := GetCwd(); got != "/tmp/test" {
		t.Errorf("GetCwd() = %v, want %v", got, "/tmp/test")
	}
}

func TestIsFeatureEnabled(t *testing.T) {
	// Set a feature
	SetFeature("test_feature", true)

	if !IsFeatureEnabled("test_feature") {
		t.Error("IsFeatureEnabled() = false, want true")
	}

	SetFeature("test_feature", false)

	if IsFeatureEnabled("test_feature") {
		t.Error("IsFeatureEnabled() = true, want false")
	}
}

func TestModelUsage(t *testing.T) {
	// Record usage
	RecordModelUsage("claude-sonnet-4-6", 1000, 500, 100*time.Millisecond)

	usage := GetModelUsage("claude-sonnet-4-6")
	if usage == nil {
		t.Fatal("GetModelUsage() returned nil")
	}

	// Fields are unexported, so we verify through state
	state := GetState()
	if state.modelUsage["claude-sonnet-4-6"] == nil {
		t.Error("modelUsage entry not found")
	}
}

func TestUpdateState(t *testing.T) {
	initialCost := GetTotalCostUSD()

	UpdateState(func(s *State) {
		s.totalCostUSD += 10.5
	})

	if cost := GetTotalCostUSD(); cost != initialCost+10.5 {
		t.Errorf("TotalCostUSD = %v, want %v", cost, initialCost+10.5)
	}
}

func TestConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup

	// Concurrent reads and writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessionId := GetSessionId()
			if sessionId == "" {
				t.Errorf("GetSessionId() returned empty string")
			}
			SetSessionId("concurrent-session")
		}(i)
	}

	wg.Wait()
}

func TestGetState(t *testing.T) {
	SetCwd("/test/path")
	SetModel("claude-sonnet-4-6")

	state := GetState()
	if state.cwd != "/test/path" {
		t.Errorf("GetState().cwd = %v, want %v", state.cwd, "/test/path")
	}
	if state.model != "claude-sonnet-4-6" {
		t.Errorf("GetState().model = %v, want %v", state.model, "claude-sonnet-4-6")
	}
}
