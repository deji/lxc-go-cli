/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/deji/lxc-go-cli/internal/helpers"
	"github.com/deji/lxc-go-cli/internal/logger"
	"github.com/spf13/cobra"
)

var (
	passwordTimeout time.Duration
)

// passwordCmd represents the password command
var passwordCmd = &cobra.Command{
	Use:   "password <container-name>",
	Short: "Retrieve the stored password for the 'app' user in a container",
	Long: `Retrieve the stored password for the 'app' user in a container.

This command retrieves the password that was generated during container creation
and stored in the container's metadata. The password is needed for sudo access
within the container.

Example:
  lxc-go-cli password mycontainer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), passwordTimeout)
		defer cancel()

		manager := &DefaultPasswordManager{}
		return retrievePassword(ctx, manager, containerName)
	},
}

// PasswordManager interface for dependency injection
type PasswordManager interface {
	ContainerExists(ctx context.Context, name string) bool
	GetContainerPassword(ctx context.Context, containerName string) (string, error)
}

// DefaultPasswordManager implements PasswordManager using helpers
type DefaultPasswordManager struct{}

func (d *DefaultPasswordManager) ContainerExists(ctx context.Context, name string) bool {
	return helpers.ContainerExists(name)
}

func (d *DefaultPasswordManager) GetContainerPassword(ctx context.Context, containerName string) (string, error) {
	return helpers.GetContainerPassword(containerName)
}

// retrievePassword retrieves and displays the stored password for a container
func retrievePassword(ctx context.Context, manager PasswordManager, containerName string) error {
	// Validate container name
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	// Check if container exists
	if !manager.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	logger.Debug("Retrieving password for container '%s'", containerName)

	// Get stored password
	password, err := manager.GetContainerPassword(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to retrieve password: %w", err)
	}

	// Display password using the helper formatter
	fmt.Print(helpers.FormatPasswordDisplay(containerName, password))
	return nil
}

func init() {
	rootCmd.AddCommand(passwordCmd)

	// Add timeout flag
	passwordCmd.Flags().DurationVarP(&passwordTimeout, "timeout", "t", 10*time.Second, "Timeout for password retrieval operation")
}

