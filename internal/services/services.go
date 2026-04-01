package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Service represents a background service
type Service interface {
	Name() string
	Start() error
	Stop() error
	IsRunning() bool
}

// BaseService provides common service functionality
type BaseService struct {
	name    string
	mu      sync.RWMutex
	running bool
}

// NewBaseService creates a new base service
func NewBaseService(name string) *BaseService {
	return &BaseService{name: name}
}

func (s *BaseService) Name() string {
	return s.name
}

func (s *BaseService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *BaseService) setRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = running
}

// HTTPService represents an HTTP service
type HTTPService struct {
	*BaseService
	addr    string
	server  *http.Server
	handler http.Handler
}

// NewHTTPService creates a new HTTP service
func NewHTTPService(name, addr string, handler http.Handler) *HTTPService {
	return &HTTPService{
		BaseService: NewBaseService(name),
		addr:    addr,
		handler: handler,
	}
}

// Start starts the HTTP service
func (s *HTTPService) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("service %s is already running", s.Name())
	}

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		s.setRunning(true)
		s.server.ListenAndServe()
	}()

	return nil
}

// Stop stops the HTTP service
func (s *HTTPService) Stop() error {
	if !s.IsRunning() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.setRunning(false)
	return s.server.Shutdown(ctx)
}

// HealthService represents a health check service
type HealthService struct {
	*BaseService
	checks []HealthCheck
}

// HealthCheck is a function that checks health
type HealthCheck func() error

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{
		BaseService: NewBaseService("health"),
		checks:      make([]HealthCheck, 0),
	}
}

// AddCheck adds a health check
func (s *HealthService) AddCheck(check HealthCheck) {
	s.checks = append(s.checks, check)
}

// Check runs all health checks
func (s *HealthService) Check() error {
	for _, check := range s.checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

// MetricsService represents a metrics collection service
type MetricsService struct {
	*BaseService
	metrics map[string]int64
	mu      sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	return &MetricsService{
		BaseService: NewBaseService("metrics"),
		metrics:     make(map[string]int64),
	}
}

// Incr increments a counter
func (s *MetricsService) Incr(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name]++
}

// Get returns a metric value
func (s *MetricsService) Get(name string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics[name]
}

// GetAll returns all metrics
func (s *MetricsService) GetAll() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]int64, len(s.metrics))
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}

// ServiceManager manages all services
type ServiceManager struct {
	mu       sync.RWMutex
	services map[string]Service
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]Service),
	}
}

// Register registers a service
func (sm *ServiceManager) Register(service Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.services[service.Name()]; ok {
		return fmt.Errorf("service %s is already registered", service.Name())
	}
	sm.services[service.Name()] = service
	return nil
}

// Unregister unregisters a service
func (sm *ServiceManager) Unregister(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	service, ok := sm.services[name]
	if !ok {
		return fmt.Errorf("service %s is not registered", name)
	}

	if service.IsRunning() {
		service.Stop()
	}
	delete(sm.services, name)
	return nil
}

// Get returns a service by name
func (sm *ServiceManager) Get(name string) (Service, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.services[name]
	return s, ok
}

// Start starts all registered services
func (sm *ServiceManager) Start() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, service := range sm.services {
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service.Name(), err)
		}
	}
	return nil
}

// Stop stops all registered services
func (sm *ServiceManager) Stop() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, service := range sm.services {
		if service.IsRunning() {
			if err := service.Stop(); err != nil {
				return fmt.Errorf("failed to stop service %s: %w", service.Name(), err)
			}
		}
	}
	return nil
}

// List returns all registered services
func (sm *ServiceManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	names := make([]string, 0, len(sm.services))
	for name := range sm.services {
		names = append(names, name)
	}
	return names
}

// DefaultServiceManager is the global service manager
var DefaultServiceManager = NewServiceManager()

// Register registers a service with the default manager
func Register(service Service) error {
	return DefaultServiceManager.Register(service)
}
