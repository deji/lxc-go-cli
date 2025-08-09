/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/lxc-go-cli/internal/helpers"
	"github.com/yourusername/lxc-go-cli/internal/logger"
)

var (
	containerName string
	imageName     string
	storageSize   string
)

// ContainerManager interface for dependency injection
type ContainerManager interface {
	GetOrCreateBtrfsPool() (string, error)
	ContainerExists(name string) bool
	CreateContainer(name, distro, release, arch, storagePool string) error
	ConfigureContainerSecurity(containerName string) error
	RunInContainer(containerName string, args ...string) error
	RestartContainer(name string) error
}

// DefaultContainerManager implements ContainerManager using helpers
type DefaultContainerManager struct{}

func (d *DefaultContainerManager) GetOrCreateBtrfsPool() (string, error) {
	return helpers.GetOrCreateBtrfsPool()
}

func (d *DefaultContainerManager) ContainerExists(name string) bool {
	return helpers.ContainerExists(name)
}

func (d *DefaultContainerManager) CreateContainer(name, distro, release, arch, storagePool string) error {
	return helpers.CreateContainer(name, distro, release, arch, storagePool)
}

func (d *DefaultContainerManager) ConfigureContainerSecurity(containerName string) error {
	return helpers.ConfigureContainerSecurity(containerName)
}

func (d *DefaultContainerManager) RunInContainer(containerName string, args ...string) error {
	return helpers.RunInContainer(containerName, args...)
}

func (d *DefaultContainerManager) RestartContainer(name string) error {
	return helpers.RestartContainer(name)
}

// createContainer creates a container with the given parameters
func createContainer(manager ContainerManager, name, image, size string) error {
	if name == "" {
		return fmt.Errorf("container name is required (use --name)")
	}
	if image == "" {
		image = "ubuntu:24.04"
	}
	if size == "" {
		size = "10G"
	}

	logger.Info("Creating container '%s' with image '%s' and storage size '%s'...", name, image, size)

	// Get or create a Btrfs storage pool without changing system default
	logger.Info("Checking for Btrfs storage pool...")
	storagePool, err := manager.GetOrCreateBtrfsPool()
	if err != nil {
		return fmt.Errorf("failed to get or create Btrfs storage pool: %w", err)
	}
	logger.Info("Using Btrfs storage pool: '%s'", storagePool)

	// Check if container already exists
	if manager.ContainerExists(name) {
		return fmt.Errorf("container '%s' already exists", name)
	}

	// Parse image string
	distro, release, arch := helpers.ParseImageString(image)

	// Create the container using LXC CLI
	logger.Info("Creating container with image %s:%s:%s using storage pool '%s'...", distro, release, arch, storagePool)
	if err := manager.CreateContainer(name, distro, release, arch, storagePool); err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Configure security settings for Docker
	logger.Info("Configuring container security settings for Docker...")
	if err := manager.ConfigureContainerSecurity(name); err != nil {
		return fmt.Errorf("failed to configure container security: %w", err)
	}

	logger.Info("Container created and started. Setting up Docker, Docker Compose, and app user...")

	// Update package index
	logger.Debug("Updating package index...")
	if err := manager.RunInContainer(name, "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}

	// Install Docker and Docker Compose
	logger.Debug("Installing Docker and Docker Compose...")
	if err := manager.RunInContainer(name, "apt-get", "install", "-y", "docker.io", "docker-compose"); err != nil {
		return fmt.Errorf("failed to install Docker and Docker Compose: %w", err)
	}

	// Create 'app' user and add to docker and sudo groups
	logger.Debug("Creating 'app' user...")
	if err := manager.RunInContainer(name, "useradd", "-m", "-s", "/bin/bash", "app"); err != nil {
		return fmt.Errorf("failed to create 'app' user: %w", err)
	}
	logger.Debug("Adding 'app' user to docker and sudo groups...")
	if err := manager.RunInContainer(name, "usermod", "-aG", "docker,sudo", "app"); err != nil {
		return fmt.Errorf("failed to add 'app' user to docker and sudo groups: %w", err)
	}

	// Restart container to ensure all settings take effect
	logger.Info("Restarting container to apply all settings...")
	if err := manager.RestartContainer(name); err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	logger.Info("Container setup complete!")
	return nil
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an LXC container ready for Docker use",
	Long: `Creates an LXC container, installs Docker and Docker Compose, and sets up a non-root 'app' user with docker and sudo access.

Example:
  lxc-go-cli create --name mycontainer --image ubuntu:24.04 --size 10G`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := &DefaultContainerManager{}
		return createContainer(manager, containerName, imageName, storageSize)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// createCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lxc-go-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	createCmd.Flags().StringVarP(&containerName, "name", "n", "", "Container name (required)")
	createCmd.Flags().StringVarP(&imageName, "image", "i", "ubuntu:24.04", "Container image (default: ubuntu:24.04)")
	createCmd.Flags().StringVarP(&storageSize, "size", "s", "10G", "Storage size (default: 10G)")
	createCmd.MarkFlagRequired("name")
}
