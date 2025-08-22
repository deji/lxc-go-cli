package helpers

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MockLXC implements LXCInterface with configurable mock behavior
type MockLXC struct {
	mu sync.RWMutex

	// Configuration
	BtrfsAvailable     bool
	DefaultPoolType    string
	ExistingPools      []string
	ExistingContainers map[string]bool
	GPUStates          map[string]*GPUStatus
	Passwords          map[string]string

	// Error injection
	CreatePoolError       error
	CreateContainerError  error
	StartContainerError   error
	RestartContainerError error
	RunCommandError       error
	SecurityConfigError   error
	SetDefaultPoolError   error
	GPUStatusError        error
	EnableGPUError        error
	DisableGPUError       error
	StorePasswordError    error
	GetPasswordError      error
	SetPasswordError      error

	// Call tracking
	Calls map[string]int
}

// NewMockLXC creates a new mock LXC implementation with sensible defaults
func NewMockLXC() *MockLXC {
	return &MockLXC{
		BtrfsAvailable:     true,
		DefaultPoolType:    "btrfs",
		ExistingPools:      []string{"default-btrfs", "test-pool"},
		ExistingContainers: make(map[string]bool),
		GPUStates:          make(map[string]*GPUStatus),
		Passwords:          make(map[string]string),
		Calls:              make(map[string]int),
	}
}

// IsBtrfsAvailable checks if Btrfs is available as a storage backend
func (m *MockLXC) IsBtrfsAvailable(ctx context.Context) bool {
	m.trackCall("IsBtrfsAvailable")
	return m.BtrfsAvailable
}

// GetDefaultStoragePoolType returns the type of the default storage pool
func (m *MockLXC) GetDefaultStoragePoolType(ctx context.Context) string {
	m.trackCall("GetDefaultStoragePoolType")
	return m.DefaultPoolType
}

// GetBtrfsStoragePools returns a list of existing Btrfs storage pools
func (m *MockLXC) GetBtrfsStoragePools(ctx context.Context) []string {
	m.trackCall("GetBtrfsStoragePools")
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy of pools to prevent modification
	pools := make([]string, len(m.ExistingPools))
	copy(pools, m.ExistingPools)
	return pools
}

// CreateBtrfsStoragePool creates a new Btrfs storage pool
func (m *MockLXC) CreateBtrfsStoragePool(ctx context.Context, name string) error {
	m.trackCall("CreateBtrfsStoragePool")

	if m.CreatePoolError != nil {
		return m.CreatePoolError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if pool already exists
	for _, pool := range m.ExistingPools {
		if pool == name {
			return fmt.Errorf("storage pool '%s' already exists", name)
		}
	}

	// Add to existing pools
	m.ExistingPools = append(m.ExistingPools, name)
	return nil
}

// GetOrCreateBtrfsPool returns an existing Btrfs pool or creates a new one
func (m *MockLXC) GetOrCreateBtrfsPool(ctx context.Context) (string, error) {
	m.trackCall("GetOrCreateBtrfsPool")

	if !m.IsBtrfsAvailable(ctx) {
		return "", fmt.Errorf("btrfs is not available on this system")
	}

	pools := m.GetBtrfsStoragePools(ctx)
	if len(pools) > 0 {
		return pools[0], nil
	}

	// Create a new pool
	poolName := "mock-btrfs-pool"
	err := m.CreateBtrfsStoragePool(ctx, poolName)
	if err != nil {
		return "", fmt.Errorf("failed to create Btrfs storage pool: %w", err)
	}

	return poolName, nil
}

// SetDefaultStoragePool sets the specified pool as the default
func (m *MockLXC) SetDefaultStoragePool(ctx context.Context, name string) error {
	m.trackCall("SetDefaultStoragePool")

	if m.SetDefaultPoolError != nil {
		return m.SetDefaultPoolError
	}

	// Check if pool exists
	pools := m.GetBtrfsStoragePools(ctx)
	for _, pool := range pools {
		if pool == name {
			m.DefaultPoolType = "btrfs"
			return nil
		}
	}

	return fmt.Errorf("storage pool '%s' not found", name)
}

// EnsureBtrfsStoragePool ensures a Btrfs storage pool exists and is set as default
func (m *MockLXC) EnsureBtrfsStoragePool(ctx context.Context) error {
	m.trackCall("EnsureBtrfsStoragePool")

	if !m.IsBtrfsAvailable(ctx) {
		return fmt.Errorf("btrfs is not available on this system")
	}

	if m.GetDefaultStoragePoolType(ctx) == "btrfs" {
		return nil
	}

	pools := m.GetBtrfsStoragePools(ctx)
	if len(pools) > 0 {
		return m.SetDefaultStoragePool(ctx, pools[0])
	}

	// Create a new pool
	poolName := "mock-btrfs-pool"
	err := m.CreateBtrfsStoragePool(ctx, poolName)
	if err != nil {
		return fmt.Errorf("failed to create Btrfs storage pool: %w", err)
	}

	return m.SetDefaultStoragePool(ctx, poolName)
}

// ContainerExists checks if a container exists
func (m *MockLXC) ContainerExists(ctx context.Context, name string) bool {
	m.trackCall("ContainerExists")
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ExistingContainers[name]
}

// CreateContainer creates a new LXC container
func (m *MockLXC) CreateContainer(ctx context.Context, name, distro, release, arch, storagePool string) error {
	m.trackCall("CreateContainer")

	if m.CreateContainerError != nil {
		return m.CreateContainerError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if container already exists
	if m.ExistingContainers[name] {
		return fmt.Errorf("container '%s' already exists", name)
	}

	// Validate storage pool exists
	poolExists := false
	for _, pool := range m.ExistingPools {
		if pool == storagePool {
			poolExists = true
			break
		}
	}
	if !poolExists {
		return fmt.Errorf("storage pool '%s' does not exist", storagePool)
	}

	// Add container
	m.ExistingContainers[name] = true
	return nil
}

// StartContainer starts an existing container
func (m *MockLXC) StartContainer(ctx context.Context, name string) error {
	m.trackCall("StartContainer")

	if m.StartContainerError != nil {
		return m.StartContainerError
	}

	if !m.ContainerExists(ctx, name) {
		return fmt.Errorf("container '%s' does not exist", name)
	}

	return nil
}

// RestartContainer restarts an existing container
func (m *MockLXC) RestartContainer(ctx context.Context, name string) error {
	m.trackCall("RestartContainer")

	if m.RestartContainerError != nil {
		return m.RestartContainerError
	}

	if !m.ContainerExists(ctx, name) {
		return fmt.Errorf("container '%s' does not exist", name)
	}

	return nil
}

// RunInContainer executes a command inside a container
func (m *MockLXC) RunInContainer(ctx context.Context, containerName string, args ...string) error {
	m.trackCall("RunInContainer")

	if m.RunCommandError != nil {
		return m.RunCommandError
	}

	if !m.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	// Simulate command execution
	if len(args) > 0 {
		// Check for specific failure scenarios
		if args[0] == "apt-get" && len(args) > 1 && args[1] == "update" {
			// Simulate package update - always succeeds in mock
		} else if args[0] == "apt-get" && len(args) > 1 && args[1] == "install" {
			// Simulate package installation - always succeeds in mock
		} else if args[0] == "useradd" {
			// Simulate user creation - always succeeds in mock
		} else if args[0] == "usermod" {
			// Simulate user modification - always succeeds in mock
		}
	}

	return nil
}

// ConfigureContainerSecurity sets up security settings needed for Docker
func (m *MockLXC) ConfigureContainerSecurity(ctx context.Context, containerName string) error {
	m.trackCall("ConfigureContainerSecurity")

	if m.SecurityConfigError != nil {
		return m.SecurityConfigError
	}

	if !m.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	return nil
}

// Helper methods for testing

// trackCall increments the call counter for a method
func (m *MockLXC) trackCall(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls[method]++
}

// GetCallCount returns the number of times a method was called
func (m *MockLXC) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Calls[method]
}

// Reset clears all call counters and resets state
func (m *MockLXC) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = make(map[string]int)
	m.ExistingContainers = make(map[string]bool)

	// Reset errors
	m.CreatePoolError = nil
	m.CreateContainerError = nil
	m.StartContainerError = nil
	m.RestartContainerError = nil
	m.RunCommandError = nil
	m.SecurityConfigError = nil
	m.SetDefaultPoolError = nil
}

// AddContainer adds a container to the mock state
func (m *MockLXC) AddContainer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExistingContainers[name] = true
}

// RemoveContainer removes a container from the mock state
func (m *MockLXC) RemoveContainer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.ExistingContainers, name)
}

// AddPool adds a storage pool to the mock state
func (m *MockLXC) AddPool(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExistingPools = append(m.ExistingPools, name)
}

// RemovePool removes a storage pool from the mock state
func (m *MockLXC) RemovePool(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, pool := range m.ExistingPools {
		if pool == name {
			m.ExistingPools = append(m.ExistingPools[:i], m.ExistingPools[i+1:]...)
			break
		}
	}
}

// SetBtrfsAvailable sets whether Btrfs is available
func (m *MockLXC) SetBtrfsAvailable(available bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BtrfsAvailable = available
}

// GetContainerGPUStatus retrieves the GPU configuration status for a container
func (m *MockLXC) GetContainerGPUStatus(ctx context.Context, containerName string) (*GPUStatus, error) {
	m.trackCall("GetContainerGPUStatus")

	if m.GPUStatusError != nil {
		return nil, m.GPUStatusError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return the configured state or default (disabled)
	if status, exists := m.GPUStates[containerName]; exists {
		// Return a copy to prevent modification
		return &GPUStatus{
			HasGPUDevice:   status.HasGPUDevice,
			PrivilegedMode: status.PrivilegedMode,
		}, nil
	}

	// Default state - GPU disabled
	return &GPUStatus{
		HasGPUDevice:   false,
		PrivilegedMode: false,
	}, nil
}

// EnableContainerGPU enables GPU access for a container
func (m *MockLXC) EnableContainerGPU(ctx context.Context, containerName string) error {
	m.trackCall("EnableContainerGPU")

	if m.EnableGPUError != nil {
		return m.EnableGPUError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Set GPU enabled state
	m.GPUStates[containerName] = &GPUStatus{
		HasGPUDevice:   true,
		PrivilegedMode: true,
	}

	return nil
}

// DisableContainerGPU disables GPU access for a container
func (m *MockLXC) DisableContainerGPU(ctx context.Context, containerName string) error {
	m.trackCall("DisableContainerGPU")

	if m.DisableGPUError != nil {
		return m.DisableGPUError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Set GPU disabled state
	m.GPUStates[containerName] = &GPUStatus{
		HasGPUDevice:   false,
		PrivilegedMode: false,
	}

	return nil
}

// SetGPUState sets the GPU state for a container (for testing)
func (m *MockLXC) SetGPUState(containerName string, hasGPU, privileged bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GPUStates[containerName] = &GPUStatus{
		HasGPUDevice:   hasGPU,
		PrivilegedMode: privileged,
	}
}

// StoreContainerPassword stores password in mock storage
func (m *MockLXC) StoreContainerPassword(ctx context.Context, containerName, password string) error {
	m.trackCall("StoreContainerPassword")

	if m.StorePasswordError != nil {
		return m.StorePasswordError
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.Passwords[containerName] = password
	return nil
}

// GetContainerPassword retrieves password from mock storage
func (m *MockLXC) GetContainerPassword(ctx context.Context, containerName string) (string, error) {
	m.trackCall("GetContainerPassword")

	if m.GetPasswordError != nil {
		return "", m.GetPasswordError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	password, exists := m.Passwords[containerName]
	if !exists {
		return "", fmt.Errorf("no password found for container '%s'", containerName)
	}

	return password, nil
}

// SetUserPassword sets password for a user in mock container
func (m *MockLXC) SetUserPassword(ctx context.Context, containerName, username, password string) error {
	m.trackCall("SetUserPassword")

	if m.SetPasswordError != nil {
		return m.SetPasswordError
	}

	// In mock, we just simulate success
	return nil
}

// SetPassword sets a password for a container (for testing)
func (m *MockLXC) SetPassword(containerName, password string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Passwords[containerName] = password
}

// SetError sets an error for a specific operation
func (m *MockLXC) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch strings.ToLower(operation) {
	case "createpool":
		m.CreatePoolError = err
	case "createcontainer":
		m.CreateContainerError = err
	case "startcontainer":
		m.StartContainerError = err
	case "restartcontainer":
		m.RestartContainerError = err
	case "runcommand":
		m.RunCommandError = err
	case "securityconfig":
		m.SecurityConfigError = err
	case "setdefaultpool":
		m.SetDefaultPoolError = err
	case "gpustatus":
		m.GPUStatusError = err
	case "enablegpu":
		m.EnableGPUError = err
	case "disablegpu":
		m.DisableGPUError = err
	case "storepassword":
		m.StorePasswordError = err
	case "getpassword":
		m.GetPasswordError = err
	case "setpassword":
		m.SetPasswordError = err
	}
}
