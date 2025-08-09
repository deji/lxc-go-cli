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
