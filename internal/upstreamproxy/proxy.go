package upstreamproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

// Proxy represents an upstream proxy
type Proxy struct {
	URL        *url.URL
	Auth       *ProxyAuth
	Transport  http.RoundTripper
	mu         sync.RWMutex
}

// ProxyAuth represents proxy authentication
type ProxyAuth struct {
	Username string
	Password string
}

// NewProxy creates a new proxy
func NewProxy(proxyURL string) (*Proxy, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return &Proxy{
		URL:       u,
		Transport: http.DefaultTransport,
	}, nil
}

// SetAuth sets proxy authentication
func (p *Proxy) SetAuth(username, password string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Auth = &ProxyAuth{
		Username: username,
		Password: password,
	}
}

// RoundTrip implements http.RoundTripper
func (p *Proxy) RoundTrip(req *http.Request) (*http.Response, error) {
	p.mu.RLock()
	transport := p.Transport
	auth := p.Auth
	p.mu.RUnlock()

	// Clone the request
	newReq := *req
	newReq.URL = req.URL // Keep original URL, only proxy the host

	// Set the proxy URL
	newReq.URL.Host = p.URL.Host
	newReq.URL.Scheme = p.URL.Scheme
	newReq.Host = p.URL.Host

	// Add authentication if set
	if auth != nil {
		newReq.Header.Set("Proxy-Authorization", fmt.Sprintf("%s:%s", auth.Username, auth.Password))
	}

	return transport.RoundTrip(&newReq)
}

// Manager manages proxy configuration
type Manager struct {
	mu       sync.RWMutex
	httpProxy *Proxy
	httpsProxy *Proxy
	noProxy  []string
}

// NewManager creates a new proxy manager
func NewManager() *Manager {
	return &Manager{
		noProxy: []string{"localhost", "127.0.0.1"},
	}
}

// SetHTTPProxy sets the HTTP proxy
func (m *Manager) SetHTTPProxy(proxy *Proxy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpProxy = proxy
}

// SetHTTPSProxy sets the HTTPS proxy
func (m *Manager) SetHTTPSProxy(proxy *Proxy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpsProxy = proxy
}

// SetNoProxy sets the no-proxy list
func (m *Manager) SetNoProxy(hosts []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.noProxy = hosts
}

// GetProxy returns the proxy for a request
func (m *Manager) GetProxy(req *http.Request) *Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if req.URL.Scheme == "https" {
		return m.httpsProxy
	}
	return m.httpProxy
}

// ShouldProxy returns true if the host should be proxied
func (m *Manager) ShouldProxy(host string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, noProxy := range m.noProxy {
		if host == noProxy {
			return false
		}
	}
	return true
}

// Transport returns an http.RoundTripper that uses the proxy
func (m *Manager) Transport() http.RoundTripper {
	return &ProxyTransport{manager: m}
}

// ProxyTransport is an http.RoundTripper that routes through proxies
type ProxyTransport struct {
	manager *Manager
}

func (pt *ProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	proxy := pt.manager.GetProxy(req)
	if proxy == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return proxy.RoundTrip(req)
}

// DefaultManager is the global proxy manager
var DefaultManager = NewManager()

// SetHTTPProxy is a convenience function using the default manager
func SetHTTPProxy(proxyURL string) error {
	proxy, err := NewProxy(proxyURL)
	if err != nil {
		return err
	}
	DefaultManager.SetHTTPProxy(proxy)
	return nil
}

// SetHTTPSProxy is a convenience function using the default manager
func SetHTTPSProxy(proxyURL string) error {
	proxy, err := NewProxy(proxyURL)
	if err != nil {
		return err
	}
	DefaultManager.SetHTTPSProxy(proxy)
	return nil
}
