package helpers

import (
	"fmt"

	"github.com/deji/lxc-go-cli/internal/logger"
)

// DockerInstaller interface for dependency injection
type DockerInstaller interface {
	RunInContainer(containerName string, args ...string) error
}

// InstallDockerInContainer installs Docker, Docker Compose V2, and sudo using Docker's official repository
func InstallDockerInContainer(installer DockerInstaller, containerName string) error {
	// Step 1: Install prerequisites for Docker repository (matching Docker docs)
	logger.Debug("Installing prerequisites for Docker repository...")
	if err := installer.RunInContainer(containerName, "apt-get", "install", "-y", "ca-certificates", "curl"); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}

	// Step 2: Add Docker's official GPG key (following Docker docs exactly)
	logger.Debug("Creating keyrings directory...")
	if err := installer.RunInContainer(containerName, "install", "-m", "0755", "-d", "/etc/apt/keyrings"); err != nil {
		return fmt.Errorf("failed to create keyrings directory: %w", err)
	}

	logger.Debug("Downloading Docker's official GPG key...")
	if err := installer.RunInContainer(containerName, "curl", "-fsSL", "https://download.docker.com/linux/ubuntu/gpg", "-o", "/etc/apt/keyrings/docker.asc"); err != nil {
		return fmt.Errorf("failed to download Docker GPG key: %w", err)
	}

	logger.Debug("Setting GPG key permissions...")
	if err := installer.RunInContainer(containerName, "chmod", "a+r", "/etc/apt/keyrings/docker.asc"); err != nil {
		return fmt.Errorf("failed to set GPG key permissions: %w", err)
	}

	// Step 3: Add Docker repository to apt sources (exact command from Docker docs)
	logger.Debug("Adding Docker repository...")
	repoCmd := `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null`
	if err := installer.RunInContainer(containerName, "sh", "-c", repoCmd); err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}

	// Step 4: Update package index with new repository
	logger.Debug("Updating package index with Docker repository...")
	if err := installer.RunInContainer(containerName, "apt-get", "update"); err != nil {
		return fmt.Errorf("failed to update package index after adding Docker repository: %w", err)
	}

	// Step 5: Install Docker packages (matching official docs exactly)
	logger.Debug("Installing sudo and Docker packages from official repository...")
	if err := installer.RunInContainer(containerName, "apt-get", "install", "-y", "sudo", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"); err != nil {
		return fmt.Errorf("failed to install Docker packages: %w", err)
	}

	// Step 6: Enable and start Docker service
	logger.Debug("Enabling and starting Docker service...")
	if err := installer.RunInContainer(containerName, "systemctl", "enable", "docker"); err != nil {
		return fmt.Errorf("failed to enable Docker service: %w", err)
	}

	if err := installer.RunInContainer(containerName, "systemctl", "start", "docker"); err != nil {
		return fmt.Errorf("failed to start Docker service: %w", err)
	}

	// Step 7: Verify Docker installation
	return VerifyDockerInstallation(installer, containerName)
}

// VerifyDockerInstallation verifies that Docker and Docker Compose V2 are working
func VerifyDockerInstallation(installer DockerInstaller, containerName string) error {
	logger.Debug("Verifying Docker installation...")

	// Verify Docker Engine
	if err := installer.RunInContainer(containerName, "docker", "--version"); err != nil {
		return fmt.Errorf("Docker verification failed: %w", err)
	}

	// Verify Docker Compose V2 (note: "docker compose" not "docker-compose")
	if err := installer.RunInContainer(containerName, "docker", "compose", "version"); err != nil {
		return fmt.Errorf("Docker Compose V2 verification failed: %w", err)
	}

	logger.Info("Docker and Docker Compose V2 installation verified successfully")
	return nil
}
