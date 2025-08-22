package helpers

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/deji/lxc-go-cli/internal/logger"
	"gopkg.in/yaml.v2"
)

// GPUStatus represents the GPU configuration status of a container
type GPUStatus struct {
	HasGPUDevice   bool
	PrivilegedMode bool
}

// IsEnabled returns true if GPU is fully enabled (both device and privileged mode)
func (s *GPUStatus) IsEnabled() bool {
	return s.HasGPUDevice && s.PrivilegedMode
}

// ContainerConfig represents the relevant parts of LXC container configuration
type ContainerConfig struct {
	Config  map[string]string            `yaml:"config"`
	Devices map[string]map[string]string `yaml:"devices"`
}

// GetContainerGPUStatus retrieves the GPU configuration status for a container
func GetContainerGPUStatus(containerName string) (*GPUStatus, error) {
	// Use standard lxc config show command (outputs YAML by default)
	cmd := exec.Command("lxc", "config", "show", containerName)
	logger.Debug("Getting GPU status for container: lxc config show %s", containerName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Command failed with output: %s", string(output))
		return nil, fmt.Errorf("failed to get container config: %w (output: %s)", err, string(output))
	}

	logger.Debug("Command succeeded with output length: %d bytes", len(output))

	return parseGPUStatus(string(output))
}

// parseGPUStatus parses the YAML output from lxc config show
func parseGPUStatus(yamlOutput string) (*GPUStatus, error) {
	var config ContainerConfig

	err := yaml.Unmarshal([]byte(yamlOutput), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container config YAML: %w", err)
	}

	status := &GPUStatus{}

	// Check for privileged mode
	if privileged, exists := config.Config["security.privileged"]; exists && privileged == "true" {
		status.PrivilegedMode = true
		logger.Debug("Privileged mode is enabled")
	} else {
		logger.Debug("Privileged mode is disabled or not set")
	}

	// Check for GPU device
	if gpuDevice, exists := config.Devices["gpu"]; exists {
		if deviceType, typeExists := gpuDevice["type"]; typeExists && deviceType == "gpu" {
			status.HasGPUDevice = true
			logger.Debug("GPU device is present")
		} else {
			logger.Debug("GPU device exists but type is not 'gpu': %v", gpuDevice)
		}
	} else {
		logger.Debug("GPU device is not present")
	}

	logger.Debug("GPU status: device=%v, privileged=%v, enabled=%v",
		status.HasGPUDevice, status.PrivilegedMode, status.IsEnabled())

	return status, nil
}

// EnableContainerGPU enables GPU access for a container (idempotent)
func EnableContainerGPU(containerName string) error {
	logger.Info("Enabling GPU for container '%s'...", containerName)

	// Check current status
	status, err := GetContainerGPUStatus(containerName)
	if err != nil {
		return fmt.Errorf("failed to check current GPU status: %w", err)
	}

	// If already fully enabled, return success
	if status.IsEnabled() {
		logger.Info("GPU is already enabled for container '%s'", containerName)
		return nil
	}

	// Add GPU device if not present
	if !status.HasGPUDevice {
		logger.Debug("Adding GPU device to container '%s'", containerName)
		cmd := exec.Command("lxc", "config", "device", "add", containerName, "gpu", "gpu")
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Debug("Failed to add GPU device: %s", string(output))
			return fmt.Errorf("failed to add GPU device: %w (output: %s)", err, string(output))
		}
		logger.Debug("GPU device added successfully")
	}

	// Set privileged mode if not enabled
	if !status.PrivilegedMode {
		logger.Debug("Setting privileged mode for container '%s'", containerName)
		cmd := exec.Command("lxc", "config", "set", containerName, "security.privileged", "true")
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Debug("Failed to set privileged mode: %s", string(output))
			return fmt.Errorf("failed to set privileged mode: %w (output: %s)", err, string(output))
		}
		logger.Debug("Privileged mode set successfully")
	}

	logger.Info("GPU enabled successfully for container '%s'", containerName)
	return nil
}

// DisableContainerGPU disables GPU access for a container (idempotent)
func DisableContainerGPU(containerName string) error {
	logger.Info("Disabling GPU for container '%s'...", containerName)

	// Check current status
	status, err := GetContainerGPUStatus(containerName)
	if err != nil {
		return fmt.Errorf("failed to check current GPU status: %w", err)
	}

	// If already fully disabled, return success
	if !status.HasGPUDevice && !status.PrivilegedMode {
		logger.Info("GPU is already disabled for container '%s'", containerName)
		return nil
	}

	// Remove GPU device if present
	if status.HasGPUDevice {
		logger.Debug("Removing GPU device from container '%s'", containerName)
		cmd := exec.Command("lxc", "config", "device", "remove", containerName, "gpu")
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Debug("Failed to remove GPU device: %s", string(output))
			return fmt.Errorf("failed to remove GPU device: %w (output: %s)", err, string(output))
		}
		logger.Debug("GPU device removed successfully")
	}

	// Disable privileged mode if enabled
	if status.PrivilegedMode {
		logger.Debug("Disabling privileged mode for container '%s'", containerName)
		cmd := exec.Command("lxc", "config", "set", containerName, "security.privileged", "false")
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Debug("Failed to disable privileged mode: %s", string(output))
			return fmt.Errorf("failed to disable privileged mode: %w (output: %s)", err, string(output))
		}
		logger.Debug("Privileged mode disabled successfully")
	}

	logger.Info("GPU disabled successfully for container '%s'", containerName)
	return nil
}

// FormatGPUStatus returns a formatted string representation of GPU status
func FormatGPUStatus(status *GPUStatus) string {
	var result strings.Builder

	result.WriteString("GPU Configuration:\n")

	if status.HasGPUDevice {
		result.WriteString("  GPU Device: present\n")
	} else {
		result.WriteString("  GPU Device: absent\n")
	}

	if status.PrivilegedMode {
		result.WriteString("  Privileged Mode: enabled\n")
	} else {
		result.WriteString("  Privileged Mode: disabled\n")
	}

	if status.IsEnabled() {
		result.WriteString("  GPU Status: enabled\n")
	} else {
		result.WriteString("  GPU Status: disabled\n")
	}

	return result.String()
}
