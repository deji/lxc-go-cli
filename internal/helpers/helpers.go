package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/deji/lxc-go-cli/internal/logger"
)

// ParseImageString parses an image string in format "distro:release:arch"
// Returns default values if parts are missing
func ParseImageString(image string) (distro, release, arch string) {
	parts := strings.Split(image, ":")
	distro = "ubuntu"
	release = "24.04"
	arch = "amd64"
	if len(parts) > 0 && parts[0] != "" {
		distro = parts[0]
	}
	if len(parts) > 1 && parts[1] != "" {
		release = parts[1]
	}
	if len(parts) > 2 && parts[2] != "" {
		arch = parts[2]
	}
	return
}

// IsBtrfsAvailable checks if Btrfs is available as a storage backend
func IsBtrfsAvailable() bool {
	cmd := exec.Command("lxc", "storage", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "btrfs")
}

// GetDefaultStoragePoolType returns the type of the default storage pool
func GetDefaultStoragePoolType() string {
	cmd := exec.Command("lxc", "storage", "show", "default")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`(?m)^\s*driver:\s*(\w+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// StoragePool represents a storage pool from LXC
type StoragePool struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
}

// GetBtrfsStoragePools returns a list of existing Btrfs storage pools
func GetBtrfsStoragePools() []string {
	// Use JSON format for reliable parsing
	cmd := exec.Command("lxc", "storage", "list", "-f", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to table format if JSON fails
		return getBtrfsPoolsFromTable()
	}

	return parseBtrfsPoolsFromJSON(string(out))
}

// parseBtrfsPoolsFromJSON parses Btrfs pools from JSON output
func parseBtrfsPoolsFromJSON(jsonOutput string) []string {
	var pools []StoragePool

	// Parse the JSON array
	err := json.Unmarshal([]byte(jsonOutput), &pools)
	if err != nil {
		logger.Debug("JSON parsing failed: %v", err)
		return getBtrfsPoolsFromTable()
	}

	var btrfsPools []string
	for _, pool := range pools {
		if pool.Driver == "btrfs" {
			btrfsPools = append(btrfsPools, pool.Name)
		}
	}

	// Debug output
	logger.Debug("Found Btrfs pools from JSON: %v", btrfsPools)

	return btrfsPools
}

// getBtrfsPoolsFromTable is a fallback method using table format
func getBtrfsPoolsFromTable() []string {
	cmd := exec.Command("lxc", "storage", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	return parseBtrfsPoolsFromTable(string(out))
}

// parseBtrfsPoolsFromTable parses Btrfs pools from table format output
func parseBtrfsPoolsFromTable(tableOutput string) []string {
	var pools []string

	lines := strings.Split(tableOutput, "\n")
	for _, line := range lines {
		// Skip header lines and empty lines
		if strings.Contains(line, "NAME") || strings.Contains(line, "---+") || strings.TrimSpace(line) == "" {
			continue
		}

		// Check if line contains btrfs driver
		if strings.Contains(line, "btrfs") {
			// Split by pipe and extract fields
			fields := strings.Split(line, "|")
			if len(fields) >= 3 {
				// First field is pool name, second field is driver
				poolName := strings.TrimSpace(fields[1])
				driver := strings.TrimSpace(fields[2])

				if poolName != "" && driver == "btrfs" {
					pools = append(pools, poolName)
				}
			}
		}
	}

	// Debug output
	logger.Debug("Found Btrfs pools from table: %v", pools)

	return pools
}

// CreateBtrfsStoragePool creates a new Btrfs storage pool
func CreateBtrfsStoragePool(name string) error {
	cmd := exec.Command("lxc", "storage", "create", name, "btrfs")
	return cmd.Run()
}

// GetOrCreateBtrfsPool returns an existing Btrfs pool or creates a new one
func GetOrCreateBtrfsPool() (string, error) {
	// Check if Btrfs is available
	if !IsBtrfsAvailable() {
		return "", fmt.Errorf("btrfs is not available on this system")
	}

	// Look for existing Btrfs pools
	btrfsPools := GetBtrfsStoragePools()
	if len(btrfsPools) > 0 {
		// Use the first existing Btrfs pool
		return btrfsPools[0], nil
	}

	// Create a new Btrfs pool with a unique name
	poolName := "btrfs-pool"
	err := CreateBtrfsStoragePool(poolName)
	if err != nil {
		return "", fmt.Errorf("failed to create Btrfs storage pool: %w", err)
	}

	return poolName, nil
}

// ContainerExists checks if a container exists
func ContainerExists(name string) bool {
	cmd := exec.Command("lxc", "list", name, "--format", "csv")

	// For debugging, capture output
	output, err := cmd.CombinedOutput()

	// Debug output using structured logging
	logger.Debug("Checking container existence for '%s'", name)
	logger.Debug("Command: lxc list %s --format csv", name)
	logger.Debug("Output: '%s'", string(output))
	logger.Debug("Error: %v", err)

	// Container exists if command succeeds AND output is not empty
	// Empty output means no container found
	exists := err == nil && len(strings.TrimSpace(string(output))) > 0
	logger.Debug("Container '%s' exists: %v", name, exists)

	return exists
}

// CreateContainer creates a new LXC container
func CreateContainer(name, distro, release, arch, storagePool string) error {
	// Create container with specific storage pool
	// LXC expects format: lxc launch remote:image container_name
	// For ubuntu:24.04:amd64, we need to use: ubuntu:24.04
	imageName := fmt.Sprintf("%s:%s", distro, release)

	args := []string{"launch", imageName, name, "--storage", storagePool}
	cmd := exec.Command("lxc", args...)

	// Debug output
	logger.Debug("Executing: lxc %v", args)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Command failed with output: %s", string(output))
		return fmt.Errorf("lxc launch failed: %w", err)
	}

	logger.Debug("Command succeeded with output: %s", string(output))
	return nil
}

// StartContainer starts an existing container
func StartContainer(name string) error {
	cmd := exec.Command("lxc", "start", name)

	// Debug output
	logger.Debug("Starting container: lxc start %s", name)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Start failed with output: %s", string(output))
		return fmt.Errorf("lxc start failed: %w", err)
	}

	logger.Debug("Start succeeded with output: %s", string(output))
	return nil
}

// RestartContainer restarts an existing container
func RestartContainer(name string) error {
	cmd := exec.Command("lxc", "restart", name)

	// Debug output
	logger.Debug("Restarting container: lxc restart %s", name)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Restart failed with output: %s", string(output))
		return fmt.Errorf("lxc restart failed: %w", err)
	}

	logger.Debug("Restart succeeded with output: %s", string(output))
	return nil
}

// RunInContainer executes a command inside a container
func RunInContainer(containerName string, args ...string) error {
	cmdArgs := append([]string{"exec", containerName, "--"}, args...)
	cmd := exec.Command("lxc", cmdArgs...)

	logger.Debug("Executing in container '%s': lxc exec %s -- %v", containerName, containerName, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Command failed with output: %s", string(output))
		return fmt.Errorf("command failed: %w (output: %s)", err, string(output))
	}

	logger.Debug("Command succeeded with output: %s", string(output))
	return nil
}

// EnsureBtrfsStoragePool ensures a Btrfs storage pool exists and is set as default
// This is kept for backward compatibility but is not the preferred approach
func EnsureBtrfsStoragePool() error {
	// Check if Btrfs is available
	if !IsBtrfsAvailable() {
		return fmt.Errorf("btrfs is not available on this system")
	}

	// Check current default pool
	defaultType := GetDefaultStoragePoolType()
	if defaultType == "btrfs" {
		return nil // Already using Btrfs as default
	}

	// Look for existing Btrfs pools
	btrfsPools := GetBtrfsStoragePools()
	if len(btrfsPools) > 0 {
		// Use the first Btrfs pool as default
		return SetDefaultStoragePool(btrfsPools[0])
	}

	// Create a new Btrfs pool
	poolName := "btrfs-pool"
	err := CreateBtrfsStoragePool(poolName)
	if err != nil {
		return fmt.Errorf("failed to create Btrfs storage pool: %w", err)
	}

	// Set it as default
	err = SetDefaultStoragePool(poolName)
	if err != nil {
		return fmt.Errorf("failed to set Btrfs storage pool as default: %w", err)
	}

	return nil
}

// SetDefaultStoragePool sets the specified pool as the default
func SetDefaultStoragePool(name string) error {
	cmd := exec.Command("lxc", "storage", "set-default", name)
	return cmd.Run()
}

// RunHostCommand executes a command directly on the host with context support
func RunHostCommand(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Create command with context for timeout/cancellation support
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	logger.Debug("Executing host command: %v", args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Host command failed with output: %s", string(output))
		return fmt.Errorf("command failed: %w (output: %s)", err, string(output))
	}

	logger.Debug("Host command succeeded with output: %s", string(output))
	return nil
}

// ConfigureContainerSecurity sets up security settings needed for Docker
func ConfigureContainerSecurity(containerName string) error {
	// Security settings needed for Docker to work in LXC containers
	settings := map[string]string{
		"security.nesting":                     "true",
		"security.syscalls.intercept.mknod":    "true",
		"security.syscalls.intercept.setxattr": "true",
	}

	for key, value := range settings {
		cmd := exec.Command("lxc", "config", "set", containerName, key, value)

		// Debug output
		logger.Debug("Setting %s=%s for container %s", key, value, containerName)

		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Debug("Failed to set %s: %s", key, string(output))
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}
