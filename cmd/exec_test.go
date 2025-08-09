package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockContainerExecManager for testing exec command
type MockContainerExecManager struct {
	ContainerExistsFunc      func(ctx context.Context, name string) bool
	ExecInteractiveShellFunc func(ctx context.Context, containerName string) error
	ExistingContainers       map[string]bool
	ExecShellError           error
	Calls                    map[string]int
}

func (m *MockContainerExecManager) ContainerExists(ctx context.Context, name string) bool {
	m.trackCall("ContainerExists")
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(ctx, name)
	}
	if m.ExistingContainers != nil {
		return m.ExistingContainers[name]
	}
	return false
}

func (m *MockContainerExecManager) ExecInteractiveShell(ctx context.Context, containerName string) error {
	m.trackCall("ExecInteractiveShell")
	if m.ExecInteractiveShellFunc != nil {
		return m.ExecInteractiveShellFunc(ctx, containerName)
	}
	if m.ExecShellError != nil {
		return m.ExecShellError
	}
	return nil
}

func (m *MockContainerExecManager) trackCall(method string) {
	if m.Calls == nil {
		m.Calls = make(map[string]int)
	}
	m.Calls[method]++
}

func (m *MockContainerExecManager) GetCallCount(method string) int {
	if m.Calls == nil {
		return 0
	}
	return m.Calls[method]
}

func TestExecCommand(t *testing.T) {
	// Test exec command creation
	if execCmd == nil {
		t.Fatal("execCmd should not be nil")
	}

	// Test exec command properties
	if execCmd.Use != "exec <container-name>" {
		t.Errorf("expected Use to be 'exec <container-name>', got '%s'", execCmd.Use)
	}

	if execCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if execCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestExecCommandArgs(t *testing.T) {
	// Test that the command expects exactly 1 argument
	if execCmd.Args == nil {
		t.Error("execCmd should have Args validation")
	}

	// Test with no args (should fail)
	err := execCmd.Args(execCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	// Test with one arg (should pass)
	err = execCmd.Args(execCmd, []string{"container-name"})
	if err != nil {
		t.Errorf("should pass with one argument: %v", err)
	}

	// Test with too many args (should fail)
	err = execCmd.Args(execCmd, []string{"container-name", "extra-arg"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}
}

func TestExecContainer(t *testing.T) {
	tests := []struct {
		name            string
		containerName   string
		containerExists bool
		runCommandError error
		expectedError   string
	}{
		{
			name:          "empty container name",
			containerName: "",
			expectedError: "container name is required",
		},
		{
			name:            "container does not exist",
			containerName:   "nonexistent-container",
			containerExists: false,
			expectedError:   "container 'nonexistent-container' does not exist",
		},
		{
			name:            "successful exec",
			containerName:   "test-container",
			containerExists: true,
			expectedError:   "",
		},
		{
			name:            "exec command fails",
			containerName:   "test-container",
			containerExists: true,
			runCommandError: fmt.Errorf("exec failed"),
			expectedError:   "failed to execute interactive shell in container 'test-container'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := &MockContainerExecManager{
				ExistingContainers: map[string]bool{
					"test-container": tt.containerExists,
				},
				ExecShellError: tt.runCommandError,
			}

			err := execContainer(ctx, manager, tt.containerName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify that ContainerExists was called
			if tt.containerName != "" {
				if manager.GetCallCount("ContainerExists") != 1 {
					t.Error("ContainerExists should be called once")
				}
				// For successful cases, ExecInteractiveShell should be called
				if tt.expectedError == "" && tt.containerName != "" && tt.containerExists {
					if manager.GetCallCount("ExecInteractiveShell") != 1 {
						t.Error("ExecInteractiveShell should be called once for successful exec")
					}
				}
			}
		})
	}
}

func TestExecContainerWithContext(t *testing.T) {
	manager := &MockContainerExecManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test with background context
	ctx := context.Background()
	err := execContainer(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should succeed with background context: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should still work since our mock doesn't respect context cancellation
	// In a real implementation, this would check context.Done()
	err = execContainer(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should work with cancelled context in mock: %v", err)
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	err = execContainer(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should work with expired timeout in mock: %v", err)
	}
}

func TestExecContainerCommandValidation(t *testing.T) {
	ctx := context.Background()
	manager := &MockContainerExecManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Now we can properly mock the ExecInteractiveShell functionality
	manager.ExecInteractiveShellFunc = func(ctx context.Context, containerName string) error {
		// Mock successful execution
		return nil
	}

	err := execContainer(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should execute successfully with mock: %v", err)
	}
}

func TestDefaultContainerExecManager(t *testing.T) {
	// Test that DefaultContainerExecManager implements ContainerExecManager interface
	var manager ContainerExecManager = &DefaultContainerExecManager{}
	_ = manager // Use the variable to avoid unused variable warning

	// Test the wrapper methods (they will likely fail without LXC but should not panic)
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultContainerExecManager methods should not panic: %v", r)
		}
	}()

	// Test ContainerExists
	exists := manager.ContainerExists(ctx, "test-container")
	t.Logf("ContainerExists returned: %v", exists)

	// Test ExecInteractiveShell (will likely fail but shouldn't panic)
	err := manager.ExecInteractiveShell(ctx, "test-container")
	t.Logf("ExecInteractiveShell returned error: %v", err)
}

func TestExecCommandFlags(t *testing.T) {
	// Test timeout flag
	timeoutFlag := execCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag should exist")
	}

	if timeoutFlag.Shorthand != "t" {
		t.Errorf("expected timeout flag shorthand to be 't', got '%s'", timeoutFlag.Shorthand)
	}

	if timeoutFlag.DefValue != "30s" {
		t.Errorf("expected timeout flag default to be '30s', got '%s'", timeoutFlag.DefValue)
	}
}

func TestExecContainerError(t *testing.T) {
	ctx := context.Background()

	// Test with manager that has existing container
	manager := &MockContainerExecManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
		ExecShellError: fmt.Errorf("command execution failed"),
	}

	err := execContainer(ctx, manager, "test-container")
	if err == nil {
		t.Error("should return error when lxc command fails")
	}

	if !contains(err.Error(), "failed to execute interactive shell in container") {
		t.Errorf("error should contain context about interactive shell execution: %v", err)
	}
}

func TestExecContainerWithMockValidation(t *testing.T) {
	ctx := context.Background()
	callHistory := make([]string, 0)

	manager := &MockContainerExecManager{
		ContainerExistsFunc: func(ctx context.Context, name string) bool {
			callHistory = append(callHistory, fmt.Sprintf("ContainerExists(%s)", name))
			return true
		},
	}

	// Add a mock function for ExecInteractiveShell
	manager.ExecInteractiveShellFunc = func(ctx context.Context, containerName string) error {
		callHistory = append(callHistory, fmt.Sprintf("ExecInteractiveShell(%s)", containerName))
		return nil
	}

	err := execContainer(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("execContainer should succeed with mock: %v", err)
	}

	// Verify call order
	expectedCalls := []string{
		"ContainerExists(test-container)",
		"ExecInteractiveShell(test-container)",
	}

	if len(callHistory) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d", len(expectedCalls), len(callHistory))
	}

	for i, expected := range expectedCalls {
		if i < len(callHistory) && callHistory[i] != expected {
			t.Errorf("expected call[%d] to be '%s', got '%s'", i, expected, callHistory[i])
		}
	}
}
