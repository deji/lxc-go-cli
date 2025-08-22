/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/deji/lxc-go-cli/internal/helpers"
	"github.com/deji/lxc-go-cli/internal/logger"
	"github.com/spf13/cobra"
)

var (
	execTimeout time.Duration
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec <container-name>",
	Short: "Execute an interactive shell in an LXC container as app user",
	Long: `Execute an interactive shell in an LXC container as the 'app' user.
This command runs 'lxc exec <container-name> -- su - app' to provide
an interactive shell session in the specified container as the app user with proper environment and group memberships.

Example:
  lxc-go-cli exec mycontainer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
		defer cancel()

		manager := &DefaultContainerExecManager{}
		return execContainer(ctx, manager, containerName)
	},
}

// ContainerExecManager interface for dependency injection
type ContainerExecManager interface {
	ContainerExists(ctx context.Context, name string) bool
	ExecInteractiveShell(ctx context.Context, containerName string) error
}

// DefaultContainerExecManager implements ContainerExecManager using helpers
type DefaultContainerExecManager struct{}

func (d *DefaultContainerExecManager) ContainerExists(ctx context.Context, name string) bool {
	return helpers.ContainerExists(name)
}

func (d *DefaultContainerExecManager) ExecInteractiveShell(ctx context.Context, containerName string) error {
	// Use lxc exec with su to properly load user environment and groups
	cmd := exec.Command("lxc", "exec", containerName, "--", "su", "-", "app")

	// Connect stdin, stdout, stderr for interactive session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug("Executing: lxc exec %s -- su - app", containerName)

	// Run the interactive command
	return cmd.Run()
}

// execContainer executes a shell in the container as app user
func execContainer(ctx context.Context, manager ContainerExecManager, containerName string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	// Check if container exists
	if !manager.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	logger.Info("Executing interactive shell in container '%s' as app user...", containerName)

	// Use the manager to execute the interactive shell
	err := manager.ExecInteractiveShell(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to execute interactive shell in container '%s': %w", containerName, err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Add timeout flag
	execCmd.Flags().DurationVarP(&execTimeout, "timeout", "t", 30*time.Second, "Timeout for the exec operation")
}
