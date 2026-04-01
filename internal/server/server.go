package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Handler is an HTTP handler function
type Handler func(context.Context, *Request) (*Response, error)

// Request represents an HTTP request
type Request struct {
	Method      string
	Path        string
	Query       map[string]string
	Headers     map[string]string
	Body        json.RawMessage
	Context     context.Context
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       any
}

// ServeMux is an HTTP request multiplexer
type ServeMux struct {
	mu      sync.RWMutex
	routes  map[string]map[string]Handler // method -> path -> handler
	notFound Handler
}

// NewServeMux creates a new serve mux
func NewServeMux() *ServeMux {
	return &ServeMux{
		routes: make(map[string]map[string]Handler),
	}
}

// Handle registers a handler for a method and path
func (mux *ServeMux) Handle(method, path string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if mux.routes[method] == nil {
		mux.routes[method] = make(map[string]Handler)
	}
	mux.routes[method][path] = handler
}

// HandleFunc registers a handler function for a method and path
func (mux *ServeMux) HandleFunc(method, path string, fn func(context.Context, *Request) (*Response, error)) {
	mux.Handle(method, path, Handler(fn))
}

// Get registers a GET handler
func (mux *ServeMux) Get(path string, handler Handler) {
	mux.Handle("GET", path, handler)
}

// Post registers a POST handler
func (mux *ServeMux) Post(path string, handler Handler) {
	mux.Handle("POST", path, handler)
}

// Put registers a PUT handler
func (mux *ServeMux) Put(path string, handler Handler) {
	mux.Handle("PUT", path, handler)
}

// Delete registers a DELETE handler
func (mux *ServeMux) Delete(path string, handler Handler) {
	mux.Handle("DELETE", path, handler)
}

// Patch registers a PATCH handler
func (mux *ServeMux) Patch(path string, handler Handler) {
	mux.Handle("PATCH", path, handler)
}

// ServeHTTP makes mux implement http.Handler
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	method := r.Method
	path := r.URL.Path

	if handlers, ok := mux.routes[method]; ok {
		if handler, ok := handlers[path]; ok {
			// Convert query parameters
			query := make(map[string]string)
			for k, v := range r.URL.Query() {
				if len(v) > 0 {
					query[k] = v[0]
				}
			}

			req := &Request{
				Method:  method,
				Path:    path,
				Query:   query,
				Headers: map[string]string{},
				Context: r.Context(),
			}

			// Copy headers
			for k, v := range r.Header {
				if len(v) > 0 {
					req.Headers[k] = v[0]
				}
			}

			// Read body
			if r.Body != nil {
				defer r.Body.Close()
				json.NewDecoder(r.Body).Decode(&req.Body)
			}

			resp, err := handler(r.Context(), req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Write response
			w.WriteHeader(resp.StatusCode)
			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			if resp.Body != nil {
				json.NewEncoder(w).Encode(resp.Body)
			}
			return
		}
	}

	if mux.notFound != nil {
		mux.notFound(r.Context(), &Request{Context: r.Context()})
		return
	}

	http.NotFound(w, r)
}

// SetNotFound sets the not found handler
func (mux *ServeMux) SetNotFound(handler Handler) {
	mux.notFound = handler
}

// Middleware is an HTTP middleware function
type Middleware func(http.Handler) http.Handler

// Use applies middleware to the mux
func (mux *ServeMux) Use(mw Middleware) {
	// This would wrap the ServeHTTP method
	_ = mw
}

// Server represents an HTTP server
type Server struct {
	addr    string
	mux     *ServeMux
	server  *http.Server
}

// NewServer creates a new server
func NewServer(addr string) *Server {
	mux := NewServeMux()
	return &Server{
		addr: addr,
		mux:  mux,
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

// Mux returns the server's serve mux
func (s *Server) Mux() *ServeMux {
	return s.mux
}

// Start starts the server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Helper function to create JSON response
func JSON(statusCode int, body any) *Response {
	return &Response{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
}

// Helper function to create error response
func Error(statusCode int, message string) *Response {
	return JSON(statusCode, map[string]string{"error": message})
}
