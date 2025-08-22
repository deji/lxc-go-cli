/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/deji/lxc-go-cli/internal/helpers"
	"github.com/deji/lxc-go-cli/internal/logger"
	"github.com/spf13/cobra"
)

var (
	gpuTimeout time.Duration
)

// gpuCmd represents the gpu command
var gpuCmd = &cobra.Command{
	Use:   "gpu <container-name> <enable|disable|status>",
	Short: "Configure GPU access for an LXC container",
	Long: `Configure GPU access for an LXC container by managing GPU device assignment and privileged mode.

This command enables or disables GPU access by:
- Adding/removing a GPU device to/from the container
- Setting/unsetting privileged mode (required for GPU access)
- Restarting the container to apply changes

Actions:
  enable  - Enable GPU access (adds GPU device and sets privileged mode)
  disable - Disable GPU access (removes GPU device and unsets privileged mode)  
  status  - Show current GPU configuration

Examples:
  lxc-go-cli gpu mycontainer enable   # Enable GPU access
  lxc-go-cli gpu mycontainer disable  # Disable GPU access
  lxc-go-cli gpu mycontainer status   # Show GPU status`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]
		action := strings.ToLower(args[1])

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), gpuTimeout)
		defer cancel()

		manager := &DefaultGPUManager{}
		return handleGPUAction(ctx, manager, containerName, action)
	},
}

// GPUManager interface for dependency injection
type GPUManager interface {
	ContainerExists(ctx context.Context, name string) bool
	GetGPUStatus(ctx context.Context, containerName string) (*helpers.GPUStatus, error)
	EnableGPU(ctx context.Context, containerName string) error
	DisableGPU(ctx context.Context, containerName string) error
	RestartContainer(ctx context.Context, name string) error
}

// DefaultGPUManager implements GPUManager using helpers
type DefaultGPUManager struct{}

func (d *DefaultGPUManager) ContainerExists(ctx context.Context, name string) bool {
	return helpers.ContainerExists(name)
}

func (d *DefaultGPUManager) GetGPUStatus(ctx context.Context, containerName string) (*helpers.GPUStatus, error) {
	return helpers.GetContainerGPUStatus(containerName)
}

func (d *DefaultGPUManager) EnableGPU(ctx context.Context, containerName string) error {
	return helpers.EnableContainerGPU(containerName)
}

func (d *DefaultGPUManager) DisableGPU(ctx context.Context, containerName string) error {
	return helpers.DisableContainerGPU(containerName)
}

func (d *DefaultGPUManager) RestartContainer(ctx context.Context, name string) error {
	return helpers.RestartContainer(name)
}

// validateGPUArgs validates the arguments for GPU operations
func validateGPUArgs(containerName, action string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	if action == "" {
		return fmt.Errorf("action is required")
	}

	validActions := []string{"enable", "disable", "status"}
	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}

	return fmt.Errorf("invalid action '%s': must be 'enable', 'disable', or 'status'", action)
}

// handleGPUAction handles the GPU action for a container
func handleGPUAction(ctx context.Context, manager GPUManager, containerName, action string) error {
	// Validate arguments
	if err := validateGPUArgs(containerName, action); err != nil {
		return err
	}

	// Check if container exists
	if !manager.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	switch action {
	case "enable":
		return handleGPUEnable(ctx, manager, containerName)
	case "disable":
		return handleGPUDisable(ctx, manager, containerName)
	case "status":
		return handleGPUStatus(ctx, manager, containerName)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

// handleGPUEnable enables GPU access for a container
func handleGPUEnable(ctx context.Context, manager GPUManager, containerName string) error {
	logger.Info("Enabling GPU access for container '%s'...", containerName)

	// Enable GPU
	if err := manager.EnableGPU(ctx, containerName); err != nil {
		return fmt.Errorf("failed to enable GPU: %w", err)
	}

	// Restart container to apply changes
	logger.Info("Restarting container '%s' to apply GPU changes...", containerName)
	if err := manager.RestartContainer(ctx, containerName); err != nil {
		return fmt.Errorf("failed to restart container after enabling GPU: %w", err)
	}

	logger.Info("GPU access enabled successfully for container '%s'", containerName)
	return nil
}

// handleGPUDisable disables GPU access for a container
func handleGPUDisable(ctx context.Context, manager GPUManager, containerName string) error {
	logger.Info("Disabling GPU access for container '%s'...", containerName)

	// Disable GPU
	if err := manager.DisableGPU(ctx, containerName); err != nil {
		return fmt.Errorf("failed to disable GPU: %w", err)
	}

	// Restart container to apply changes
	logger.Info("Restarting container '%s' to apply GPU changes...", containerName)
	if err := manager.RestartContainer(ctx, containerName); err != nil {
		return fmt.Errorf("failed to restart container after disabling GPU: %w", err)
	}

	logger.Info("GPU access disabled successfully for container '%s'", containerName)
	return nil
}

// handleGPUStatus shows GPU status for a container
func handleGPUStatus(ctx context.Context, manager GPUManager, containerName string) error {
	logger.Debug("Getting GPU status for container '%s'", containerName)

	status, err := manager.GetGPUStatus(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to get GPU status: %w", err)
	}

	// Format and display status
	fmt.Print(helpers.FormatGPUStatus(status))
	return nil
}

func init() {
	rootCmd.AddCommand(gpuCmd)

	// Add timeout flag
	gpuCmd.Flags().DurationVarP(&gpuTimeout, "timeout", "t", 60*time.Second, "Timeout for GPU operations")
}

