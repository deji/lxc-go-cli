package helpers

import (
	"context"
)

// LXCInterface defines the interface for LXC operations
type LXCInterface interface {
	// Storage operations
	IsBtrfsAvailable(ctx context.Context) bool
	GetDefaultStoragePoolType(ctx context.Context) string
	GetBtrfsStoragePools(ctx context.Context) []string
	CreateBtrfsStoragePool(ctx context.Context, name string) error
	GetOrCreateBtrfsPool(ctx context.Context) (string, error)
	SetDefaultStoragePool(ctx context.Context, name string) error
	EnsureBtrfsStoragePool(ctx context.Context) error

	// Container operations
	ContainerExists(ctx context.Context, name string) bool
	CreateContainer(ctx context.Context, name, distro, release, arch, storagePool string) error
	StartContainer(ctx context.Context, name string) error
	RestartContainer(ctx context.Context, name string) error
	RunInContainer(ctx context.Context, containerName string, args ...string) error
	ConfigureContainerSecurity(ctx context.Context, containerName string) error

	// GPU operations
	GetContainerGPUStatus(ctx context.Context, containerName string) (*GPUStatus, error)
	EnableContainerGPU(ctx context.Context, containerName string) error
	DisableContainerGPU(ctx context.Context, containerName string) error

	// Password operations
	StoreContainerPassword(ctx context.Context, containerName, password string) error
	GetContainerPassword(ctx context.Context, containerName string) (string, error)
	SetUserPassword(ctx context.Context, containerName, username, password string) error
}

// RealLXC implements LXCInterface using actual LXC commands
type RealLXC struct{}

// NewRealLXC creates a new real LXC implementation
func NewRealLXC() *RealLXC {
	return &RealLXC{}
}

// IsBtrfsAvailable checks if Btrfs is available as a storage backend
func (r *RealLXC) IsBtrfsAvailable(ctx context.Context) bool {
	return IsBtrfsAvailable()
}

// GetDefaultStoragePoolType returns the type of the default storage pool
func (r *RealLXC) GetDefaultStoragePoolType(ctx context.Context) string {
	return GetDefaultStoragePoolType()
}

// GetBtrfsStoragePools returns a list of existing Btrfs storage pools
func (r *RealLXC) GetBtrfsStoragePools(ctx context.Context) []string {
	return GetBtrfsStoragePools()
}

// CreateBtrfsStoragePool creates a new Btrfs storage pool
func (r *RealLXC) CreateBtrfsStoragePool(ctx context.Context, name string) error {
	return CreateBtrfsStoragePool(name)
}

// GetOrCreateBtrfsPool returns an existing Btrfs pool or creates a new one
func (r *RealLXC) GetOrCreateBtrfsPool(ctx context.Context) (string, error) {
	return GetOrCreateBtrfsPool()
}

// SetDefaultStoragePool sets the specified pool as the default
func (r *RealLXC) SetDefaultStoragePool(ctx context.Context, name string) error {
	return SetDefaultStoragePool(name)
}

// EnsureBtrfsStoragePool ensures a Btrfs storage pool exists and is set as default
func (r *RealLXC) EnsureBtrfsStoragePool(ctx context.Context) error {
	return EnsureBtrfsStoragePool()
}

// ContainerExists checks if a container exists
func (r *RealLXC) ContainerExists(ctx context.Context, name string) bool {
	return ContainerExists(name)
}

// CreateContainer creates a new LXC container
func (r *RealLXC) CreateContainer(ctx context.Context, name, distro, release, arch, storagePool string) error {
	return CreateContainer(name, distro, release, arch, storagePool)
}

// StartContainer starts an existing container
func (r *RealLXC) StartContainer(ctx context.Context, name string) error {
	return StartContainer(name)
}

// RestartContainer restarts an existing container
func (r *RealLXC) RestartContainer(ctx context.Context, name string) error {
	return RestartContainer(name)
}

// RunInContainer executes a command inside a container
func (r *RealLXC) RunInContainer(ctx context.Context, containerName string, args ...string) error {
	return RunInContainer(containerName, args...)
}

// ConfigureContainerSecurity sets up security settings needed for Docker
func (r *RealLXC) ConfigureContainerSecurity(ctx context.Context, containerName string) error {
	return ConfigureContainerSecurity(containerName)
}

// GetContainerGPUStatus retrieves the GPU configuration status for a container
func (r *RealLXC) GetContainerGPUStatus(ctx context.Context, containerName string) (*GPUStatus, error) {
	return GetContainerGPUStatus(containerName)
}

// EnableContainerGPU enables GPU access for a container
func (r *RealLXC) EnableContainerGPU(ctx context.Context, containerName string) error {
	return EnableContainerGPU(containerName)
}

// DisableContainerGPU disables GPU access for a container
func (r *RealLXC) DisableContainerGPU(ctx context.Context, containerName string) error {
	return DisableContainerGPU(containerName)
}

// StoreContainerPassword stores password in LXC metadata
func (r *RealLXC) StoreContainerPassword(ctx context.Context, containerName, password string) error {
	return StoreContainerPassword(containerName, password)
}

// GetContainerPassword retrieves password from LXC metadata
func (r *RealLXC) GetContainerPassword(ctx context.Context, containerName string) (string, error) {
	return GetContainerPassword(containerName)
}

// SetUserPassword sets the password for a user inside a container
func (r *RealLXC) SetUserPassword(ctx context.Context, containerName, username, password string) error {
	return SetUserPassword(containerName, username, password)
}
