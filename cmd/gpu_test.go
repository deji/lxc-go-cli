package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/deji/lxc-go-cli/internal/helpers"
)

// MockGPUManager for testing GPU command
type MockGPUManager struct {
	ContainerExistsFunc  func(ctx context.Context, name string) bool
	GetGPUStatusFunc     func(ctx context.Context, containerName string) (*helpers.GPUStatus, error)
	EnableGPUFunc        func(ctx context.Context, containerName string) error
	DisableGPUFunc       func(ctx context.Context, containerName string) error
	RestartContainerFunc func(ctx context.Context, name string) error

	ExistingContainers map[string]bool
	GPUStates          map[string]*helpers.GPUStatus
	Calls              map[string]int
	EnableError        error
	DisableError       error
	StatusError        error
	RestartError       error
}

func NewMockGPUManager() *MockGPUManager {
	return &MockGPUManager{
		ExistingContainers: make(map[string]bool),
		GPUStates:          make(map[string]*helpers.GPUStatus),
		Calls:              make(map[string]int),
	}
}

func (m *MockGPUManager) ContainerExists(ctx context.Context, name string) bool {
	m.trackCall("ContainerExists")
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(ctx, name)
	}
	return m.ExistingContainers[name]
}

func (m *MockGPUManager) GetGPUStatus(ctx context.Context, containerName string) (*helpers.GPUStatus, error) {
	m.trackCall("GetGPUStatus")
	if m.GetGPUStatusFunc != nil {
		return m.GetGPUStatusFunc(ctx, containerName)
	}
	if m.StatusError != nil {
		return nil, m.StatusError
	}
	if status, exists := m.GPUStates[containerName]; exists {
		return status, nil
	}
	return &helpers.GPUStatus{HasGPUDevice: false, PrivilegedMode: false}, nil
}

func (m *MockGPUManager) EnableGPU(ctx context.Context, containerName string) error {
	m.trackCall("EnableGPU")
	if m.EnableGPUFunc != nil {
		return m.EnableGPUFunc(ctx, containerName)
	}
	if m.EnableError != nil {
		return m.EnableError
	}
	m.GPUStates[containerName] = &helpers.GPUStatus{HasGPUDevice: true, PrivilegedMode: true}
	return nil
}

func (m *MockGPUManager) DisableGPU(ctx context.Context, containerName string) error {
	m.trackCall("DisableGPU")
	if m.DisableGPUFunc != nil {
		return m.DisableGPUFunc(ctx, containerName)
	}
	if m.DisableError != nil {
		return m.DisableError
	}
	m.GPUStates[containerName] = &helpers.GPUStatus{HasGPUDevice: false, PrivilegedMode: false}
	return nil
}

func (m *MockGPUManager) RestartContainer(ctx context.Context, name string) error {
	m.trackCall("RestartContainer")
	if m.RestartContainerFunc != nil {
		return m.RestartContainerFunc(ctx, name)
	}
	return m.RestartError
}

func (m *MockGPUManager) trackCall(method string) {
	if m.Calls == nil {
		m.Calls = make(map[string]int)
	}
	m.Calls[method]++
}

func (m *MockGPUManager) GetCallCount(method string) int {
	return m.Calls[method]
}

func TestGPUCommand(t *testing.T) {
	// Test GPU command creation
	if gpuCmd == nil {
		t.Fatal("gpuCmd should not be nil")
	}

	// Test GPU command properties
	if gpuCmd.Use != "gpu <container-name> <enable|disable|status>" {
		t.Errorf("expected specific Use format, got '%s'", gpuCmd.Use)
	}

	if gpuCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if gpuCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestGPUCommandArgs(t *testing.T) {
	// Test that the command expects exactly 2 arguments
	if gpuCmd.Args == nil {
		t.Error("gpuCmd should have Args validation")
	}

	// Test with wrong number of args
	err := gpuCmd.Args(gpuCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	err = gpuCmd.Args(gpuCmd, []string{"container"})
	if err == nil {
		t.Error("should fail with only one argument")
	}

	err = gpuCmd.Args(gpuCmd, []string{"container", "enable", "extra"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}

	// Test with correct number of args (should pass)
	err = gpuCmd.Args(gpuCmd, []string{"container", "enable"})
	if err != nil {
		t.Errorf("should pass with two arguments: %v", err)
	}
}

func TestValidateGPUArgs(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		action        string
		expectedError string
	}{
		{
			name:          "empty container name",
			containerName: "",
			action:        "enable",
			expectedError: "container name is required",
		},
		{
			name:          "empty action",
			containerName: "test-container",
			action:        "",
			expectedError: "action is required",
		},
		{
			name:          "invalid action",
			containerName: "test-container",
			action:        "invalid",
			expectedError: "invalid action 'invalid': must be 'enable', 'disable', or 'status'",
		},
		{
			name:          "valid enable action",
			containerName: "test-container",
			action:        "enable",
			expectedError: "",
		},
		{
			name:          "valid disable action",
			containerName: "test-container",
			action:        "disable",
			expectedError: "",
		},
		{
			name:          "valid status action",
			containerName: "test-container",
			action:        "status",
			expectedError: "",
		},
		{
			name:          "case insensitive action",
			containerName: "test-container",
			action:        "ENABLE",
			expectedError: "invalid action 'ENABLE': must be 'enable', 'disable', or 'status'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGPUArgs(tt.containerName, tt.action)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestHandleGPUAction(t *testing.T) {
	tests := []struct {
		name            string
		containerName   string
		action          string
		containerExists bool
		expectedError   string
		expectedCalls   map[string]int
	}{
		{
			name:          "invalid arguments",
			containerName: "",
			action:        "enable",
			expectedError: "container name is required",
			expectedCalls: map[string]int{},
		},
		{
			name:            "container does not exist",
			containerName:   "nonexistent",
			action:          "enable",
			containerExists: false,
			expectedError:   "container 'nonexistent' does not exist",
			expectedCalls:   map[string]int{"ContainerExists": 1},
		},
		{
			name:            "invalid action",
			containerName:   "test-container",
			action:          "invalid",
			containerExists: true,
			expectedError:   "invalid action 'invalid'",
			expectedCalls:   map[string]int{},
		},
		{
			name:            "successful enable",
			containerName:   "test-container",
			action:          "enable",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   map[string]int{"ContainerExists": 1, "EnableGPU": 1, "RestartContainer": 1},
		},
		{
			name:            "successful disable",
			containerName:   "test-container",
			action:          "disable",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   map[string]int{"ContainerExists": 1, "DisableGPU": 1, "RestartContainer": 1},
		},
		{
			name:            "successful status",
			containerName:   "test-container",
			action:          "status",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   map[string]int{"ContainerExists": 1, "GetGPUStatus": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockGPUManager()
			manager.ExistingContainers["test-container"] = tt.containerExists

			err := handleGPUAction(ctx, manager, tt.containerName, tt.action)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify expected calls
			for method, expectedCount := range tt.expectedCalls {
				actualCount := manager.GetCallCount(method)
				if actualCount != expectedCount {
					t.Errorf("expected %d calls to %s, got %d", expectedCount, method, actualCount)
				}
			}
		})
	}
}

func TestHandleGPUEnable(t *testing.T) {
	tests := []struct {
		name         string
		enableError  error
		restartError error
		expectedErr  string
	}{
		{
			name:        "successful enable",
			expectedErr: "",
		},
		{
			name:        "enable GPU fails",
			enableError: fmt.Errorf("enable failed"),
			expectedErr: "failed to enable GPU",
		},
		{
			name:         "restart fails",
			restartError: fmt.Errorf("restart failed"),
			expectedErr:  "failed to restart container after enabling GPU",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockGPUManager()
			manager.EnableError = tt.enableError
			manager.RestartError = tt.restartError

			err := handleGPUEnable(ctx, manager, "test-container")

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedErr)
				} else if !contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify calls
			if manager.GetCallCount("EnableGPU") != 1 {
				t.Errorf("expected 1 call to EnableGPU, got %d", manager.GetCallCount("EnableGPU"))
			}

			// Only restart if enable succeeded
			expectedRestartCalls := 0
			if tt.enableError == nil {
				expectedRestartCalls = 1
			}
			if manager.GetCallCount("RestartContainer") != expectedRestartCalls {
				t.Errorf("expected %d calls to RestartContainer, got %d", expectedRestartCalls, manager.GetCallCount("RestartContainer"))
			}
		})
	}
}

func TestHandleGPUDisable(t *testing.T) {
	tests := []struct {
		name         string
		disableError error
		restartError error
		expectedErr  string
	}{
		{
			name:        "successful disable",
			expectedErr: "",
		},
		{
			name:         "disable GPU fails",
			disableError: fmt.Errorf("disable failed"),
			expectedErr:  "failed to disable GPU",
		},
		{
			name:         "restart fails",
			restartError: fmt.Errorf("restart failed"),
			expectedErr:  "failed to restart container after disabling GPU",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockGPUManager()
			manager.DisableError = tt.disableError
			manager.RestartError = tt.restartError

			err := handleGPUDisable(ctx, manager, "test-container")

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedErr)
				} else if !contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify calls
			if manager.GetCallCount("DisableGPU") != 1 {
				t.Errorf("expected 1 call to DisableGPU, got %d", manager.GetCallCount("DisableGPU"))
			}

			// Only restart if disable succeeded
			expectedRestartCalls := 0
			if tt.disableError == nil {
				expectedRestartCalls = 1
			}
			if manager.GetCallCount("RestartContainer") != expectedRestartCalls {
				t.Errorf("expected %d calls to RestartContainer, got %d", expectedRestartCalls, manager.GetCallCount("RestartContainer"))
			}
		})
	}
}

func TestHandleGPUStatus(t *testing.T) {
	tests := []struct {
		name        string
		statusError error
		expectedErr string
	}{
		{
			name:        "successful status",
			expectedErr: "",
		},
		{
			name:        "status fails",
			statusError: fmt.Errorf("status failed"),
			expectedErr: "failed to get GPU status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockGPUManager()
			manager.StatusError = tt.statusError

			err := handleGPUStatus(ctx, manager, "test-container")

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedErr)
				} else if !contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify calls
			if manager.GetCallCount("GetGPUStatus") != 1 {
				t.Errorf("expected 1 call to GetGPUStatus, got %d", manager.GetCallCount("GetGPUStatus"))
			}
		})
	}
}

func TestDefaultGPUManager(t *testing.T) {
	// Test that DefaultGPUManager implements GPUManager interface
	var manager GPUManager = &DefaultGPUManager{}
	_ = manager // Use the variable to avoid unused variable warning

	// Test the wrapper methods (they will likely fail without LXC but should not panic)
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultGPUManager methods should not panic: %v", r)
		}
	}()

	// Test all interface methods
	exists := manager.ContainerExists(ctx, "test-container")
	t.Logf("ContainerExists returned: %v", exists)

	status, err := manager.GetGPUStatus(ctx, "test-container")
	t.Logf("GetGPUStatus returned: status=%v, err=%v", status, err)

	err = manager.EnableGPU(ctx, "test-container")
	t.Logf("EnableGPU returned: %v", err)

	err = manager.DisableGPU(ctx, "test-container")
	t.Logf("DisableGPU returned: %v", err)

	err = manager.RestartContainer(ctx, "test-container")
	t.Logf("RestartContainer returned: %v", err)
}

func TestGPUCommandFlags(t *testing.T) {
	// Test timeout flag
	timeoutFlag := gpuCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag should exist")
	}

	if timeoutFlag.Shorthand != "t" {
		t.Errorf("expected timeout flag shorthand to be 't', got '%s'", timeoutFlag.Shorthand)
	}

	if timeoutFlag.DefValue != "1m0s" {
		t.Errorf("expected timeout flag default to be '1m0s', got '%s'", timeoutFlag.DefValue)
	}
}

func TestGPUWithContext(t *testing.T) {
	manager := NewMockGPUManager()
	manager.ExistingContainers["test-container"] = true

	// Test with background context
	ctx := context.Background()
	err := handleGPUAction(ctx, manager, "test-container", "status")
	if err != nil {
		t.Errorf("should succeed with background context: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should still work since our mock doesn't respect context cancellation
	err = handleGPUAction(ctx, manager, "test-container", "status")
	if err != nil {
		t.Errorf("should work with cancelled context in mock: %v", err)
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	err = handleGPUAction(ctx, manager, "test-container", "status")
	if err != nil {
		t.Errorf("should work with expired timeout in mock: %v", err)
	}
}

func TestGPUIdempotentOperations(t *testing.T) {
	ctx := context.Background()
	manager := NewMockGPUManager()
	manager.ExistingContainers["test-container"] = true

	// Test enabling GPU multiple times
	err := handleGPUEnable(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("first enable should succeed: %v", err)
	}

	// Reset call counts for second test
	manager.Calls = make(map[string]int)

	err = handleGPUEnable(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("second enable should succeed (idempotent): %v", err)
	}

	// Test disabling GPU multiple times
	err = handleGPUDisable(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("first disable should succeed: %v", err)
	}

	// Reset call counts for second test
	manager.Calls = make(map[string]int)

	err = handleGPUDisable(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("second disable should succeed (idempotent): %v", err)
	}
}

func TestGPUActionCaseHandling(t *testing.T) {
	ctx := context.Background()
	manager := NewMockGPUManager()
	manager.ExistingContainers["test-container"] = true

	// Test that action is case-sensitive in current implementation
	err := handleGPUAction(ctx, manager, "test-container", "ENABLE")
	if err == nil {
		t.Error("should fail with uppercase action (case sensitive)")
	}

	err = handleGPUAction(ctx, manager, "test-container", "Enable")
	if err == nil {
		t.Error("should fail with mixed case action (case sensitive)")
	}

	// But lowercase should work
	err = handleGPUAction(ctx, manager, "test-container", "enable")
	if err != nil {
		t.Errorf("should succeed with lowercase action: %v", err)
	}
}
