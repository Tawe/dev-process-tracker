package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/devports/devpt/pkg/models"
)

// Registry manages stored service definitions
type Registry struct {
	filePath string
	data     *models.Registry
	mu       sync.RWMutex
}

// NewRegistry creates a new registry instance
func NewRegistry(filePath string) *Registry {
	return &Registry{
		filePath: filePath,
		data: &models.Registry{
			Services: make(map[string]*models.ManagedService),
			Version:  "1.0",
		},
	}
}

// Load reads the registry from disk
// FilePath returns the registry file path.
func (r *Registry) FilePath() string {
	return r.filePath
}

func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if file exists
	_, err := os.Stat(r.filePath)
	if os.IsNotExist(err) {
		// File doesn't exist yet, initialize with empty registry
		r.data.Services = make(map[string]*models.ManagedService)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat registry file: %w", err)
	}

	// Read file
	content, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	// Parse JSON
	data := &models.Registry{}
	if err := json.Unmarshal(content, data); err != nil {
		return fmt.Errorf("failed to parse registry: %w", err)
	}

	r.data = data
	return nil
}

// AddService registers a new managed service
func (r *Registry) AddService(service *models.ManagedService) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data.Services[service.Name]; exists {
		return fmt.Errorf("service %q already exists", service.Name)
	}

	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now
	r.data.Services[service.Name] = service

	return r.save()
}

// UpdateService updates an existing managed service
func (r *Registry) UpdateService(service *models.ManagedService) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data.Services[service.Name]; !exists {
		return fmt.Errorf("service %q not found", service.Name)
	}

	service.UpdatedAt = time.Now()
	r.data.Services[service.Name] = service

	return r.save()
}

// RenameService renames a managed service by changing its registry key.
// All runtime/identity fields are preserved so a running service keeps
// resolving to its renamed entry. Done as one locked save() so there is no
// add+delete window for an orphan or duplicate.
func (r *Registry) RenameService(oldName, newName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if newName == oldName {
		return fmt.Errorf("new name %q is the same as the current name", newName)
	}
	svc, exists := r.data.Services[oldName]
	if !exists {
		return fmt.Errorf("service %q not found", oldName)
	}
	if _, exists := r.data.Services[newName]; exists {
		return fmt.Errorf("service %q already exists", newName)
	}

	svc.Name = newName
	svc.UpdatedAt = time.Now()
	r.data.Services[newName] = svc
	delete(r.data.Services, oldName)
	return r.save()
}

// UpsertService creates a service, or updates an existing one while
// preserving its runtime/identity fields. Used by `devpt add --force`.
func (r *Registry) UpsertService(in *models.ManagedService) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if existing, ok := r.data.Services[in.Name]; ok {
		existing.CWD = in.CWD
		existing.Command = in.Command
		existing.Ports = in.Ports
		existing.UpdatedAt = now
	} else {
		in.CreatedAt = now
		in.UpdatedAt = now
		r.data.Services[in.Name] = in
	}
	return r.save()
}

// GetService retrieves a service by name
func (r *Registry) GetService(name string) *models.ManagedService {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.data.Services[name]
}

// ListServices returns all managed services
func (r *Registry) ListServices() []*models.ManagedService {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*models.ManagedService, 0, len(r.data.Services))
	for _, svc := range r.data.Services {
		services = append(services, svc)
	}
	return services
}

// RemoveService removes a service from the registry
func (r *Registry) RemoveService(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data.Services[name]; !exists {
		return fmt.Errorf("service %q not found", name)
	}

	delete(r.data.Services, name)
	return r.save()
}

// UpdateServicePID updates the last PID for a service
func (r *Registry) UpdateServicePID(name string, pid int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.data.Services[name]
	if !exists {
		return fmt.Errorf("service %q not found", name)
	}

	svc.LastPID = &pid
	svc.LastProcessStartTime = nil
	now := time.Now()
	svc.LastStart = &now
	svc.LastStop = nil
	svc.UpdatedAt = now

	return r.save()
}

// UpdateServiceProcessIdentity updates the last confirmed process identity for a service.
func (r *Registry) UpdateServiceProcessIdentity(name string, pid int, processStartTime time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.data.Services[name]
	if !exists {
		return fmt.Errorf("service %q not found", name)
	}

	now := time.Now()
	svc.LastPID = &pid
	svc.LastProcessStartTime = &processStartTime
	svc.LastStart = &now
	svc.LastStop = nil
	svc.UpdatedAt = now

	return r.save()
}

// ClearServicePID marks a managed service as not running.
func (r *Registry) ClearServicePID(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.data.Services[name]
	if !exists {
		return fmt.Errorf("service %q not found", name)
	}

	now := time.Now()
	svc.LastPID = nil
	svc.LastProcessStartTime = nil
	svc.LastStop = &now
	svc.UpdatedAt = now
	return r.save()
}

// UpdateServiceResolvedCommand records the OS-resolved command for a service.
// This is the actual command visible via ps after the process starts, which may differ
// from the declared command (e.g. "bunx vite" -> "node .../vite").
func (r *Registry) UpdateServiceResolvedCommand(name, resolvedCommand string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.data.Services[name]
	if !exists {
		return fmt.Errorf("service %q not found", name)
	}

	svc.ResolvedCommand = resolvedCommand
	svc.UpdatedAt = time.Now()
	return r.save()
}

// save (internal) writes the registry without taking locks
func (r *Registry) save() error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	content, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(r.filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	return nil
}
