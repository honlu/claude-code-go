package bootstrap

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// State holds all global mutable state for the CLI
type State struct {
	mu sync.RWMutex

	// Session identifiers
	sessionId         string
	parentSessionId   string
	resumeSessionId   string
	switchSessionFlag bool

	// Working directory
	cwd       string
	projectRoot string

	// Usage tracking
	totalCostUSD     float64
	totalAPIDuration int64 // milliseconds
	totalToolDuration int64

	modelUsage map[string]*ModelUsage

	// Session info
	isInteractive bool
	isHeadless   bool
	bareMode     bool

	// API configuration
	apiKeySource       string
	authTokenSource    string

	// Model configuration
	model             string
	fastModel         string
	smallFastModel    string

	// Feature flags
	featureFlags map[string]bool

	// Telemetry
	meter any // OpenTelemetry meter - interface{} to avoid import cycle

	// Session plugins
	inlinePlugins []string

	// Agent color state
	agentColorMap  map[string]string
	agentColorIndex int

	// Session-created teams
	sessionCreatedTeams map[string]bool

	// Container info
	containerId string
}

// ModelUsage tracks API usage per model
type ModelUsage struct {
	inputTokens  int
	outputTokens int
	apiCalls     int
	duration     int64 // milliseconds
}

// Global state instance
var globalState = &State{
	sessionId:       uuid.New().String(),
	modelUsage:      make(map[string]*ModelUsage),
	featureFlags:    make(map[string]bool),
	agentColorMap:   make(map[string]string),
	sessionCreatedTeams: make(map[string]bool),
}

var stateMu = &globalState.mu

// Accessors

func GetSessionId() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.sessionId
}

func SetSessionId(id string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.sessionId = id
}

func GetParentSessionId() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.parentSessionId
}

func SetParentSessionId(id string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.parentSessionId = id
}

func GetCwd() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.cwd
}

func SetCwd(cwd string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.cwd = cwd
}

func GetProjectRoot() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.projectRoot
}

func SetProjectRoot(root string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.projectRoot = root
}

func IsInteractive() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.isInteractive
}

func SetInteractive(v bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.isInteractive = v
}

func IsHeadless() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.isHeadless
}

func SetHeadless(v bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.isHeadless = v
}

func IsBareMode() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.bareMode
}

func SetBareMode(v bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.bareMode = v
}

func GetModel() string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.model
}

func SetModel(model string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.model = model
}

func GetTotalCostUSD() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.totalCostUSD
}

func AddCostUSD(cost float64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.totalCostUSD += cost
}

func GetModelUsage(model string) *ModelUsage {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.modelUsage[model]
}

func RecordModelUsage(model string, inputTokens, outputTokens int, duration time.Duration) {
	stateMu.Lock()
	defer stateMu.Unlock()
	usage, ok := globalState.modelUsage[model]
	if !ok {
		usage = &ModelUsage{}
		globalState.modelUsage[model] = usage
	}
	usage.inputTokens += inputTokens
	usage.outputTokens += outputTokens
	usage.apiCalls++
	usage.duration += duration.Milliseconds()
}

func GetTotalAPIDuration() int64 {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.totalAPIDuration
}

func GetTotalToolDuration() int64 {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.totalToolDuration
}

func IsFeatureEnabled(name string) bool {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return globalState.featureFlags[name]
}

func SetFeature(name string, enabled bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	globalState.featureFlags[name] = enabled
}

// UpdateState applies a function to modify the state (similar to React setState)
func UpdateState(f func(*State)) {
	stateMu.Lock()
	defer stateMu.Unlock()
	f(globalState)
}

// GetState returns a copy of the current state
func GetState() State {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return *globalState
}
