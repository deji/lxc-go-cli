package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockPasswordManager for testing password command
type MockPasswordManager struct {
	ContainerExistsFunc      func(ctx context.Context, name string) bool
	GetContainerPasswordFunc func(ctx context.Context, containerName string) (string, error)

	ExistingContainers map[string]bool
	StoredPasswords    map[string]string
	Calls              map[string]int
	GetPasswordError   error
}

func NewMockPasswordManager() *MockPasswordManager {
	return &MockPasswordManager{
		ExistingContainers: make(map[string]bool),
		StoredPasswords:    make(map[string]string),
		Calls:              make(map[string]int),
	}
}

func (m *MockPasswordManager) ContainerExists(ctx context.Context, name string) bool {
	m.trackCall("ContainerExists")
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(ctx, name)
	}
	return m.ExistingContainers[name]
}

func (m *MockPasswordManager) GetContainerPassword(ctx context.Context, containerName string) (string, error) {
	m.trackCall("GetContainerPassword")
	if m.GetContainerPasswordFunc != nil {
		return m.GetContainerPasswordFunc(ctx, containerName)
	}
	if m.GetPasswordError != nil {
		return "", m.GetPasswordError
	}
	if password, exists := m.StoredPasswords[containerName]; exists {
		return password, nil
	}
	return "", fmt.Errorf("no password found for container '%s'", containerName)
}

func (m *MockPasswordManager) trackCall(method string) {
	if m.Calls == nil {
		m.Calls = make(map[string]int)
	}
	m.Calls[method]++
}

func (m *MockPasswordManager) GetCallCount(method string) int {
	return m.Calls[method]
}

func TestPasswordCommand(t *testing.T) {
	// Test password command creation
	if passwordCmd == nil {
		t.Fatal("passwordCmd should not be nil")
	}

	// Test password command properties
	if passwordCmd.Use != "password <container-name>" {
		t.Errorf("expected Use to be 'password <container-name>', got '%s'", passwordCmd.Use)
	}

	if passwordCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if passwordCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestPasswordCommandArgs(t *testing.T) {
	// Test that the command expects exactly 1 argument
	if passwordCmd.Args == nil {
		t.Error("passwordCmd should have Args validation")
	}

	// Test with wrong number of args
	err := passwordCmd.Args(passwordCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	err = passwordCmd.Args(passwordCmd, []string{"container", "extra"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}

	// Test with correct number of args (should pass)
	err = passwordCmd.Args(passwordCmd, []string{"container"})
	if err != nil {
		t.Errorf("should pass with one argument: %v", err)
	}
}

func TestRetrievePassword(t *testing.T) {
	tests := []struct {
		name            string
		containerName   string
		containerExists bool
		storedPassword  string
		passwordError   error
		expectedError   string
		expectedCalls   map[string]int
	}{
		{
			name:          "empty container name",
			containerName: "",
			expectedError: "container name is required",
			expectedCalls: map[string]int{},
		},
		{
			name:            "container does not exist",
			containerName:   "nonexistent",
			containerExists: false,
			expectedError:   "container 'nonexistent' does not exist",
			expectedCalls:   map[string]int{"ContainerExists": 1},
		},
		{
			name:            "successful password retrieval",
			containerName:   "test-container",
			containerExists: true,
			storedPassword:  "MyPassword123",
			expectedError:   "",
			expectedCalls:   map[string]int{"ContainerExists": 1, "GetContainerPassword": 1},
		},
		{
			name:            "password not found",
			containerName:   "test-container",
			containerExists: true,
			passwordError:   fmt.Errorf("no password found"),
			expectedError:   "failed to retrieve password",
			expectedCalls:   map[string]int{"ContainerExists": 1, "GetContainerPassword": 1},
		},
		{
			name:            "password retrieval error",
			containerName:   "test-container",
			containerExists: true,
			passwordError:   fmt.Errorf("metadata access failed"),
			expectedError:   "failed to retrieve password",
			expectedCalls:   map[string]int{"ContainerExists": 1, "GetContainerPassword": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockPasswordManager()
			manager.ExistingContainers[tt.containerName] = tt.containerExists
			if tt.storedPassword != "" {
				manager.StoredPasswords[tt.containerName] = tt.storedPassword
			}
			manager.GetPasswordError = tt.passwordError

			err := retrievePassword(ctx, manager, tt.containerName)

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

func TestDefaultPasswordManager(t *testing.T) {
	// Test that DefaultPasswordManager implements PasswordManager interface
	var manager PasswordManager = &DefaultPasswordManager{}
	_ = manager // Use the variable to avoid unused variable warning

	// Test the wrapper methods (they will likely fail without LXC but should not panic)
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultPasswordManager methods should not panic: %v", r)
		}
	}()

	// Test ContainerExists
	exists := manager.ContainerExists(ctx, "test-container")
	t.Logf("ContainerExists returned: %v", exists)

	// Test GetContainerPassword (will likely fail but shouldn't panic)
	password, err := manager.GetContainerPassword(ctx, "test-container")
	t.Logf("GetContainerPassword returned: password=%s, err=%v", password, err)
}

func TestPasswordCommandFlags(t *testing.T) {
	// Test timeout flag
	timeoutFlag := passwordCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag should exist")
	}

	if timeoutFlag.Shorthand != "t" {
		t.Errorf("expected timeout flag shorthand to be 't', got '%s'", timeoutFlag.Shorthand)
	}

	if timeoutFlag.DefValue != "10s" {
		t.Errorf("expected timeout flag default to be '10s', got '%s'", timeoutFlag.DefValue)
	}
}

func TestPasswordWithContext(t *testing.T) {
	manager := NewMockPasswordManager()
	manager.ExistingContainers["test-container"] = true
	manager.StoredPasswords["test-container"] = "TestPassword123"

	// Test with background context
	ctx := context.Background()
	err := retrievePassword(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should succeed with background context: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should still work since our mock doesn't respect context cancellation
	err = retrievePassword(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should work with cancelled context in mock: %v", err)
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	err = retrievePassword(ctx, manager, "test-container")
	if err != nil {
		t.Errorf("should work with expired timeout in mock: %v", err)
	}
}

func TestPasswordManagerEdgeCases(t *testing.T) {
	manager := NewMockPasswordManager()

	// Test with special container names
	specialNames := []string{
		"container-with-dashes",
		"container_with_underscores",
		"container123",
		"ContainerWithCaps",
	}

	for _, name := range specialNames {
		t.Run(fmt.Sprintf("container_%s", name), func(t *testing.T) {
			ctx := context.Background()
			manager.ExistingContainers[name] = true
			manager.StoredPasswords[name] = "TestPassword"

			err := retrievePassword(ctx, manager, name)
			if err != nil {
				t.Errorf("should handle container name '%s': %v", name, err)
			}
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		expectedError string
	}{
		{
			name:          "empty string",
			containerName: "",
			expectedError: "container name is required",
		},
		{
			name:          "whitespace only",
			containerName: "   ",
			expectedError: "", // This will pass validation but fail on container existence
		},
		{
			name:          "valid name",
			containerName: "valid-container",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := NewMockPasswordManager()

			// For valid names, set up container to exist
			if tt.expectedError == "" && tt.containerName != "   " {
				manager.ExistingContainers[tt.containerName] = true
				manager.StoredPasswords[tt.containerName] = "TestPassword"
			}

			err := retrievePassword(ctx, manager, tt.containerName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil && tt.containerName != "   " {
				t.Errorf("expected no error for valid input, got %v", err)
			}
		})
	}
}

