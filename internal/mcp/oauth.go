package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// OAuthHandler handles OAuth 2.0 + PKCE for MCP servers
type OAuthHandler struct {
	config    *OAuthConfig
	serverURL string
	clientID  string
	authURL   string
	tokenURL  string
	scopes    []string

	// Token storage
	tokenFile string
	mu        sync.RWMutex
	tokens    *TokenResponse
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
}

// AuthCodeConfig holds OAuth authorization code flow configuration
type AuthCodeConfig struct {
	ClientID    string
	AuthURL     string
	TokenURL    string
	RedirectURI string
	Scopes      []string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(config *OAuthConfig, serverURL string) *OAuthHandler {
	return &OAuthHandler{
		config:    config,
		serverURL: serverURL,
		clientID:  config.ClientID,
		scopes:    []string{"mcpp"},
	}
}

// OAuthTokenFile returns the path to the token file
func OAuthTokenFile(serverName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "mcp-oauth-"+sanitizeFilename(serverName)+".json")
}

// sanitizeFilename makes a string safe for use in filenames
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	return name
}

// StartAuthCodeFlow starts the OAuth authorization code flow
func (h *OAuthHandler) StartAuthCodeFlow(ctx context.Context) (string, error) {
	// Generate PKCE verifier and challenge
	verifier, err := generatePKCEVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	challenge, err := generatePKCEChallenge(verifier)
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE challenge: %w", err)
	}

	// Generate state
	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL
	_ = h.buildAuthURL(state, verifier, challenge) // authURL would be used to open browser

	// Start local server for callback
	callbackPort := h.config.CallbackPort
	if callbackPort == 0 {
		callbackPort = 8080
	}

	code, err := h.waitForCallback(ctx, callbackPort, state)
	if err != nil {
		return "", fmt.Errorf("failed to receive callback: %w", err)
	}

	// Exchange code for tokens
	tokens, err := h.exchangeCode(code, verifier)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Store tokens
	h.tokens = tokens
	if err := h.saveTokens(); err != nil {
		return "", fmt.Errorf("failed to save tokens: %w", err)
	}

	return tokens.AccessToken, nil
}

// buildAuthURL builds the OAuth authorization URL
func (h *OAuthHandler) buildAuthURL(state, verifier, challenge string) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {h.clientID},
		"redirect_uri":          {fmt.Sprintf("http://localhost:%d/callback", h.config.CallbackPort)},
		"scope":                 {strings.Join(h.scopes, " ")},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}

	if h.authURL != "" {
		return h.authURL + "?" + params.Encode()
	}

	// Discover auth URL from server metadata
	return fmt.Sprintf("%s/oauth/authorize?%s", h.serverURL, params.Encode())
}

// waitForCallback starts a local server and waits for the OAuth callback
func (h *OAuthHandler) waitForCallback(ctx context.Context, port int, expectedState string) (string, error) {
	// This would typically start an HTTP server
	// For now, return an error indicating it's not implemented
	return "", fmt.Errorf("OAuth callback server not implemented")
}

// exchangeCode exchanges an authorization code for tokens
func (h *OAuthHandler) exchangeCode(code, verifier string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {h.clientID},
		"code":          {code},
		"redirect_uri":  {fmt.Sprintf("http://localhost:%d/callback", h.config.CallbackPort)},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequest("POST", h.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	tokens.IssuedAt = time.Now()
	return &tokens, nil
}

// RefreshTokens refreshes the access token
func (h *OAuthHandler) RefreshTokens() error {
	if h.tokens == nil || h.tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {h.clientID},
		"refresh_token": {h.tokens.RefreshToken},
	}

	req, err := http.NewRequest("POST", h.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return err
	}

	tokens.IssuedAt = time.Now()
	h.tokens = &tokens

	return h.saveTokens()
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (h *OAuthHandler) GetAccessToken() (string, error) {
	h.mu.RLock()
	if h.tokens != nil && h.isTokenValid() {
		token := h.tokens.AccessToken
		h.mu.RUnlock()
		return token, nil
	}
	h.mu.RUnlock()

	// Try to load from file
	if err := h.loadTokens(); err == nil && h.tokens != nil && h.isTokenValid() {
		return h.tokens.AccessToken, nil
	}

	// Try to refresh
	if h.tokens != nil && h.tokens.RefreshToken != "" {
		if err := h.RefreshTokens(); err == nil {
			return h.tokens.AccessToken, nil
		}
	}

	// Start new auth flow
	return h.StartAuthCodeFlow(context.Background())
}

// isTokenValid checks if the current token is valid
func (h *OAuthHandler) isTokenValid() bool {
	if h.tokens == nil || h.tokens.AccessToken == "" {
		return false
	}
	// Token is valid for 5 minutes
	return time.Since(h.tokens.IssuedAt) < time.Duration(h.tokens.ExpiresIn-300)*time.Second
}

// saveTokens saves tokens to disk
func (h *OAuthHandler) saveTokens() error {
	if h.tokenFile == "" {
		return fmt.Errorf("token file not set")
	}

	dir := filepath.Dir(h.tokenFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(h.tokens, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.tokenFile, data, 0600)
}

// loadTokens loads tokens from disk
func (h *OAuthHandler) loadTokens() error {
	if h.tokenFile == "" {
		return fmt.Errorf("token file not set")
	}

	data, err := os.ReadFile(h.tokenFile)
	if err != nil {
		return err
	}

	var tokens TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		return err
	}

	h.tokens = &tokens
	return nil
}

// RevokeTokens revokes the current tokens
func (h *OAuthHandler) RevokeTokens() error {
	if h.tokens == nil || h.tokens.AccessToken == "" {
		return nil
	}

	// Send revocation request
	data := url.Values{"token": {h.tokens.AccessToken}}
	req, err := http.NewRequest("POST", h.tokenURL+"/revoke", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Clear tokens
	h.tokens = nil
	os.Remove(h.tokenFile)

	return nil
}

// PKCE helper functions

func generatePKCEVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generatePKCEChallenge(verifier string) (string, error) {
	h := sha256.New()
	h.Write([]byte(verifier))
	d := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(d), nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// XAAHandler handles Cross-App Access (SEP-990)
type XAAHandler struct {
	OAuthHandler
	issuer        string
	idpClientID   string
	idpIDToken    string
}

// NewXAAHandler creates a new XAA handler
func NewXAAHandler(config *OAuthConfig, serverURL string) *XAAHandler {
	return &XAAHandler{
		OAuthHandler: *NewOAuthHandler(config, serverURL),
	}
}

// PerformTokenExchange performs RFC 8693 Token Exchange for XAA
func (h *XAAHandler) PerformTokenExchange() (*TokenResponse, error) {
	// Exchange ID token for ID-JAG
	idJagToken, err := h.exchangeIDTokenForIDJAG()
	if err != nil {
		return nil, fmt.Errorf("failed to exchange ID token: %w", err)
	}

	// Exchange ID-JAG for access token (RFC 7523)
	return h.exchangeIDJAGForAccessToken(idJagToken)
}

// exchangeIDTokenForIDJAG performs RFC 8693 Token Exchange
func (h *XAAHandler) exchangeIDTokenForIDJAG() (string, error) {
	// This is a simplified implementation
	// In production, this would call the token exchange endpoint
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token": {h.idpIDToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:id_token"},
		"requested_token_type": {"urn:ietf:params:oauth:token-type:jwt"},
	}

	req, err := http.NewRequest("POST", h.issuer+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if token, ok := result["access_token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("no access token in response")
}

// exchangeIDJAGForAccessToken performs RFC 7523 JWT Bearer Grant
func (h *XAAHandler) exchangeIDJAGForAccessToken(idJagToken string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {idJagToken},
		"scope":      {strings.Join(h.scopes, " ")},
	}

	req, err := http.NewRequest("POST", h.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JWT bearer grant failed: %s", string(body))
	}

	var tokens TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	tokens.IssuedAt = time.Now()
	return &tokens, nil
}
