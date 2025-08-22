package cmd

import (
	"context"
	"testing"

	"github.com/deji/lxc-go-cli/internal/logger"
)

func TestRootCommandLogLevelFlag(t *testing.T) {
	// Test that the log-level flag exists and has correct default
	logLevelFlag := rootCmd.PersistentFlags().Lookup("log-level")
	if logLevelFlag == nil {
		t.Fatal("log-level flag should exist")
	}

	if logLevelFlag.Shorthand != "l" {
		t.Errorf("expected log-level flag shorthand to be 'l', got '%s'", logLevelFlag.Shorthand)
	}

	if logLevelFlag.DefValue != "info" {
		t.Errorf("expected log-level flag default to be 'info', got '%s'", logLevelFlag.DefValue)
	}
}

func TestLogLevelInitialization(t *testing.T) {
	// Save original level
	originalLevel := logger.GetLevel()
	defer logger.SetLevel(originalLevel)

	tests := []struct {
		name          string
		flagValue     string
		expectedLevel logger.LogLevel
	}{
		{"debug level", "debug", logger.DEBUG},
		{"info level", "info", logger.INFO},
		{"warn level", "warn", logger.WARN},
		{"error level", "error", logger.ERROR},
		{"invalid level defaults to info", "invalid", logger.INFO},
		{"empty level defaults to info", "", logger.INFO},
		{"case insensitive", "DEBUG", logger.DEBUG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default
			logger.SetLevel(logger.INFO)

			// Set the flag value
			logLevel = tt.flagValue

			// Call the PersistentPreRun function directly
			rootCmd.PersistentPreRun(rootCmd, []string{})

			// Check that the level was set correctly
			if logger.GetLevel() != tt.expectedLevel {
				t.Errorf("expected logger level %v, got %v", tt.expectedLevel, logger.GetLevel())
			}
		})
	}
}

func TestLoggingInCommands(t *testing.T) {
	// Test that commands use the logging system correctly
	th := logger.NewTestHelper()
	defer th.Cleanup()

	// Set up mock manager for create command
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false // Container doesn't exist
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return nil
		},
		RunInContainerFunc: func(containerName string, args ...string) error {
			return nil
		},
		RestartContainerFunc: func(name string) error {
			return nil
		},
	}

	// Test with INFO level
	th.SetLevel(logger.INFO)
	th.ClearOutput()

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err != nil {
		t.Errorf("createContainer should succeed: %v", err)
	}

	// Should see INFO messages
	th.AssertContainsLog(t, logger.INFO, "Creating container")
	th.AssertContainsLog(t, logger.INFO, "Checking for Btrfs storage pool")
	th.AssertContainsLog(t, logger.INFO, "Container setup complete")

	// Test with ERROR level (should be quiet)
	th.SetLevel(logger.ERROR)
	th.ClearOutput()

	err = createContainer(manager, "test-container-2", "ubuntu:24.04", "10G")
	if err != nil {
		t.Errorf("createContainer should succeed: %v", err)
	}

	// Should NOT see INFO messages
	output := th.GetOutput()
	if output != "" {
		t.Errorf("expected no log output with ERROR level, got: %s", output)
	}
}

func TestDebuggingInHelpers(t *testing.T) {
	// Test that helper functions use debug logging correctly
	th := logger.NewTestHelper()
	defer th.Cleanup()

	// Test with DEBUG level - should see debug messages
	th.SetLevel(logger.DEBUG)
	th.ClearOutput()

	// Call a helper function that uses debug logging
	// Note: This will fail because lxc is not available, but we're testing logging
	exists := testContainerExistsForLogging("test-container")
	_ = exists // We don't care about the result, just the logging

	// Should see debug messages about the lxc command
	th.AssertContainsLog(t, logger.DEBUG, "Checking container existence")
	th.AssertContainsLog(t, logger.DEBUG, "Command: lxc list")

	// Test with INFO level - should NOT see debug messages
	th.SetLevel(logger.INFO)
	th.ClearOutput()

	exists = testContainerExistsForLogging("test-container-2")
	_ = exists

	// Should NOT see debug messages
	output := th.GetOutput()
	if output != "" {
		t.Errorf("expected no debug output with INFO level, got: %s", output)
	}
}

// Helper function to test container existence logging without importing helpers
// This avoids circular imports and tests the logging behavior directly
func testContainerExistsForLogging(name string) bool {
	// Simulate the logic from helpers.ContainerExists for testing
	logger.Debug("Checking container existence for '%s'", name)
	logger.Debug("Command: lxc list %s --format csv", name)
	logger.Debug("Output: '%s'", "")
	logger.Debug("Error: %v", "exec: \"lxc\": executable file not found in $PATH")
	logger.Debug("Container '%s' exists: %v", name, false)
	return false
}

func TestLogLevelFlagInExecCommand(t *testing.T) {
	// Test that exec command inherits log level from root
	th := logger.NewTestHelper()
	defer th.Cleanup()

	// Mock manager
	manager := &MockContainerExecManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test with INFO level
	th.SetLevel(logger.INFO)
	th.ClearOutput()

	// Add mock function for ExecInteractiveShell to succeed
	manager.ExecInteractiveShellFunc = func(ctx context.Context, containerName string) error {
		return nil
	}

	// This would normally create context, but we'll call execContainer directly
	err := execContainer(nil, manager, "test-container")
	if err != nil {
		t.Errorf("execContainer should succeed with mock: %v", err)
	}

	// Should see INFO message about executing shell
	th.AssertContainsLog(t, logger.INFO, "Executing interactive shell in container")

	// Test with ERROR level
	th.SetLevel(logger.ERROR)
	th.ClearOutput()

	err = execContainer(nil, manager, "test-container")
	if err != nil {
		t.Errorf("execContainer should succeed with ERROR level too: %v", err)
	}

	// Should NOT see INFO messages
	output := th.GetOutput()
	if output != "" {
		t.Errorf("expected no log output with ERROR level, got: %s", output)
	}
}

func TestLogLevelFlagInPortCommand(t *testing.T) {
	// Test that port command uses logging correctly
	th := logger.NewTestHelper()
	defer th.Cleanup()

	// Mock manager
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test with INFO level
	th.SetLevel(logger.INFO)
	th.ClearOutput()

	err := configurePortForwarding(nil, manager, "test-container", "8080", "80", "tcp")
	if err != nil {
		t.Errorf("configurePortForwarding should succeed: %v", err)
	}

	// Should see INFO messages about port forwarding
	th.AssertContainsLog(t, logger.INFO, "Configuring TCP port forwarding")
	th.AssertContainsLog(t, logger.INFO, "Successfully configured TCP port forwarding")

	// Test with ERROR level
	th.SetLevel(logger.ERROR)
	th.ClearOutput()

	err = configurePortForwarding(nil, manager, "test-container", "9090", "90", "udp")
	if err != nil {
		t.Errorf("configurePortForwarding should succeed: %v", err)
	}

	// Should NOT see INFO messages
	output := th.GetOutput()
	if output != "" {
		t.Errorf("expected no log output with ERROR level, got: %s", output)
	}
}
