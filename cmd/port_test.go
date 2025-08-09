package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// MockContainerPortManager for testing port command
type MockContainerPortManager struct {
	ContainerExistsFunc func(ctx context.Context, name string) bool
	RunLXCCommandFunc   func(ctx context.Context, args ...string) error
	ExistingContainers  map[string]bool
	RunCommandError     error
	Calls               map[string]int
	LastCommand         []string
}

func (m *MockContainerPortManager) ContainerExists(ctx context.Context, name string) bool {
	m.trackCall("ContainerExists")
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(ctx, name)
	}
	if m.ExistingContainers != nil {
		return m.ExistingContainers[name]
	}
	return false
}

func (m *MockContainerPortManager) RunLXCCommand(ctx context.Context, args ...string) error {
	m.trackCall("RunLXCCommand")
	m.LastCommand = args
	if m.RunLXCCommandFunc != nil {
		return m.RunLXCCommandFunc(ctx, args...)
	}
	if m.RunCommandError != nil {
		return m.RunCommandError
	}
	return nil
}

func (m *MockContainerPortManager) trackCall(method string) {
	if m.Calls == nil {
		m.Calls = make(map[string]int)
	}
	m.Calls[method]++
}

func (m *MockContainerPortManager) GetCallCount(method string) int {
	if m.Calls == nil {
		return 0
	}
	return m.Calls[method]
}

func TestPortCommand(t *testing.T) {
	// Test port command creation
	if portCmd == nil {
		t.Fatal("portCmd should not be nil")
	}

	// Test port command properties
	if portCmd.Use != "port <container-name> <host-port> <container-port> [tcp|udp|both]" {
		t.Errorf("expected specific Use format, got '%s'", portCmd.Use)
	}

	if portCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if portCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestPortCommandArgs(t *testing.T) {
	// Test that the command expects 3-4 arguments
	if portCmd.Args == nil {
		t.Error("portCmd should have Args validation")
	}

	// Test with wrong number of args
	err := portCmd.Args(portCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	err = portCmd.Args(portCmd, []string{"container"})
	if err == nil {
		t.Error("should fail with too few arguments")
	}

	err = portCmd.Args(portCmd, []string{"container", "host"})
	if err == nil {
		t.Error("should fail with too few arguments")
	}

	err = portCmd.Args(portCmd, []string{"container", "host", "container", "protocol", "extra"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}

	// Test with correct number of args (should pass)
	err = portCmd.Args(portCmd, []string{"container", "8080", "80"})
	if err != nil {
		t.Errorf("should pass with three arguments (protocol optional): %v", err)
	}

	err = portCmd.Args(portCmd, []string{"container", "8080", "80", "tcp"})
	if err != nil {
		t.Errorf("should pass with four arguments: %v", err)
	}
}

func TestValidatePortForwardingArgs(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		hostPort      string
		containerPort string
		protocol      string
		expectedError string
	}{
		{
			name:          "empty container name",
			containerName: "",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "container name is required",
		},
		{
			name:          "empty host port",
			containerName: "test-container",
			hostPort:      "",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "host port is required",
		},
		{
			name:          "invalid host port - not a number",
			containerName: "test-container",
			hostPort:      "abc",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "invalid host port 'abc': must be a number",
		},
		{
			name:          "invalid host port - too low",
			containerName: "test-container",
			hostPort:      "0",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "invalid host port '0': must be between 1 and 65535",
		},
		{
			name:          "invalid host port - too high",
			containerName: "test-container",
			hostPort:      "65536",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "invalid host port '65536': must be between 1 and 65535",
		},
		{
			name:          "empty container port",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "",
			protocol:      "tcp",
			expectedError: "container port is required",
		},
		{
			name:          "invalid container port - not a number",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "xyz",
			protocol:      "tcp",
			expectedError: "invalid container port 'xyz': must be a number",
		},
		{
			name:          "invalid container port - too low",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "0",
			protocol:      "tcp",
			expectedError: "invalid container port '0': must be between 1 and 65535",
		},
		{
			name:          "invalid container port - too high",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "65536",
			protocol:      "tcp",
			expectedError: "invalid container port '65536': must be between 1 and 65535",
		},
		{
			name:          "invalid protocol",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "invalid",
			expectedError: "invalid protocol 'invalid': must be 'tcp', 'udp', or 'both'",
		},
		{
			name:          "valid tcp",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "",
		},
		{
			name:          "valid udp",
			containerName: "test-container",
			hostPort:      "5353",
			containerPort: "53",
			protocol:      "udp",
			expectedError: "",
		},
		{
			name:          "valid both",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "both",
			expectedError: "",
		},
		{
			name:          "valid tcp uppercase",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "TCP",
			expectedError: "",
		},
		{
			name:          "valid port edge cases",
			containerName: "test-container",
			hostPort:      "1",
			containerPort: "65535",
			protocol:      "tcp",
			expectedError: "",
		},
		{
			name:          "empty protocol defaults to tcp",
			containerName: "test-container",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortForwardingArgs(tt.containerName, tt.hostPort, tt.containerPort, tt.protocol)

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

func TestConfigurePortForwarding(t *testing.T) {
	tests := []struct {
		name            string
		containerName   string
		hostPort        string
		containerPort   string
		protocol        string
		containerExists bool
		runCommandError error
		expectedError   string
		expectedCalls   int
	}{
		{
			name:          "invalid arguments",
			containerName: "",
			hostPort:      "8080",
			containerPort: "80",
			protocol:      "tcp",
			expectedError: "container name is required",
			expectedCalls: 0,
		},
		{
			name:            "container does not exist",
			containerName:   "nonexistent",
			hostPort:        "8080",
			containerPort:   "80",
			protocol:        "tcp",
			containerExists: false,
			expectedError:   "container 'nonexistent' does not exist",
			expectedCalls:   0,
		},
		{
			name:            "successful tcp forwarding",
			containerName:   "test-container",
			hostPort:        "8080",
			containerPort:   "80",
			protocol:        "tcp",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   1,
		},
		{
			name:            "successful udp forwarding",
			containerName:   "test-container",
			hostPort:        "5353",
			containerPort:   "53",
			protocol:        "udp",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   1,
		},
		{
			name:            "successful both forwarding",
			containerName:   "test-container",
			hostPort:        "8080",
			containerPort:   "80",
			protocol:        "both",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   2, // Both TCP and UDP
		},
		{
			name:            "command execution fails",
			containerName:   "test-container",
			hostPort:        "8080",
			containerPort:   "80",
			protocol:        "tcp",
			containerExists: true,
			runCommandError: fmt.Errorf("lxc command failed"),
			expectedError:   "failed to configure tcp port forwarding",
			expectedCalls:   1,
		},
		{
			name:            "successful default protocol (empty string)",
			containerName:   "test-container",
			hostPort:        "8080",
			containerPort:   "80",
			protocol:        "",
			containerExists: true,
			expectedError:   "",
			expectedCalls:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := &MockContainerPortManager{
				ExistingContainers: map[string]bool{
					"test-container": tt.containerExists,
				},
				RunCommandError: tt.runCommandError,
			}

			err := configurePortForwarding(ctx, manager, tt.containerName, tt.hostPort, tt.containerPort, tt.protocol)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			// Verify expected number of RunLXCCommand calls
			if manager.GetCallCount("RunLXCCommand") != tt.expectedCalls {
				t.Errorf("expected %d RunLXCCommand calls, got %d", tt.expectedCalls, manager.GetCallCount("RunLXCCommand"))
			}
		})
	}
}

func TestConfigurePortForwardingForProtocol(t *testing.T) {
	ctx := context.Background()
	commandHistory := make([][]string, 0)

	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
		RunLXCCommandFunc: func(ctx context.Context, args ...string) error {
			commandHistory = append(commandHistory, args)
			return nil
		},
	}

	err := configurePortForwardingForProtocol(ctx, manager, "test-container", "8080", "80", "tcp")
	if err != nil {
		t.Errorf("should succeed: %v", err)
	}

	// Verify the command was constructed correctly
	if len(commandHistory) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commandHistory))
	}

	cmd := commandHistory[0]
	expectedCmd := []string{
		"lxc", "config", "device", "add", "test-container",
		"test-container-8080-80-tcp", "proxy",
		"connect=tcp:0.0.0.0:8080", "listen=tcp:0.0.0.0:80",
	}

	if len(cmd) != len(expectedCmd) {
		t.Errorf("expected command with %d args, got %d", len(expectedCmd), len(cmd))
	}

	for i, expected := range expectedCmd {
		if i < len(cmd) && cmd[i] != expected {
			t.Errorf("expected cmd[%d] to be '%s', got '%s'", i, expected, cmd[i])
		}
	}
}

func TestConfigurePortForwardingBothProtocols(t *testing.T) {
	ctx := context.Background()
	commandHistory := make([][]string, 0)

	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
		RunLXCCommandFunc: func(ctx context.Context, args ...string) error {
			commandHistory = append(commandHistory, args)
			return nil
		},
	}

	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "both")
	if err != nil {
		t.Errorf("should succeed: %v", err)
	}

	// Should have 2 commands: one for TCP, one for UDP
	if len(commandHistory) != 2 {
		t.Fatalf("expected 2 commands for 'both' protocol, got %d", len(commandHistory))
	}

	// Check TCP command
	tcpCmd := commandHistory[0]
	if !contains(strings.Join(tcpCmd, " "), "test-container-8080-80-tcp") {
		t.Error("first command should be for TCP")
	}
	if !contains(strings.Join(tcpCmd, " "), "connect=tcp:0.0.0.0:8080") {
		t.Error("TCP command should have correct connect parameter")
	}

	// Check UDP command
	udpCmd := commandHistory[1]
	if !contains(strings.Join(udpCmd, " "), "test-container-8080-80-udp") {
		t.Error("second command should be for UDP")
	}
	if !contains(strings.Join(udpCmd, " "), "connect=udp:0.0.0.0:8080") {
		t.Error("UDP command should have correct connect parameter")
	}
}

func TestDefaultContainerPortManager(t *testing.T) {
	// Test that DefaultContainerPortManager implements ContainerPortManager interface
	var manager ContainerPortManager = &DefaultContainerPortManager{}
	_ = manager // Use the variable to avoid unused variable warning

	// Test the wrapper methods (they will likely fail without LXC but should not panic)
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultContainerPortManager methods should not panic: %v", r)
		}
	}()

	// Test ContainerExists
	exists := manager.ContainerExists(ctx, "test-container")
	t.Logf("ContainerExists returned: %v", exists)

	// Test RunLXCCommand (will likely fail but shouldn't panic)
	err := manager.RunLXCCommand(ctx, "echo", "test")
	t.Logf("RunLXCCommand returned error: %v", err)
}

func TestPortCommandFlags(t *testing.T) {
	// Test timeout flag
	timeoutFlag := portCmd.Flags().Lookup("timeout")
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

func TestPortForwardingWithContext(t *testing.T) {
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test with background context
	ctx := context.Background()
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp")
	if err != nil {
		t.Errorf("should succeed with background context: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should still work since our mock doesn't respect context cancellation
	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp")
	if err != nil {
		t.Errorf("should work with cancelled context in mock: %v", err)
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp")
	if err != nil {
		t.Errorf("should work with expired timeout in mock: %v", err)
	}
}

func TestPortForwardingEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Test protocol case insensitivity
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test uppercase protocol
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "TCP")
	if err != nil {
		t.Errorf("should handle uppercase protocol: %v", err)
	}

	// Test mixed case protocol
	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "BoTh")
	if err != nil {
		t.Errorf("should handle mixed case protocol: %v", err)
	}
}

func TestPortForwardingPartialFailure(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
		RunLXCCommandFunc: func(ctx context.Context, args ...string) error {
			callCount++
			// Fail on the second call (UDP) when protocol is "both"
			if callCount == 2 {
				return fmt.Errorf("second command failed")
			}
			return nil
		},
	}

	// Test that if UDP fails when protocol is "both", the whole operation fails
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "both")
	if err == nil {
		t.Error("should fail when second command fails")
	}

	if !contains(err.Error(), "failed to configure udp port forwarding") {
		t.Errorf("error should indicate UDP failure: %v", err)
	}
}

func TestPortCommandWithDefaultProtocol(t *testing.T) {
	ctx := context.Background()
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test configuring port forwarding with empty protocol (should default to tcp)
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "")
	if err != nil {
		t.Errorf("should succeed with empty protocol: %v", err)
	}

	// Should have called RunLXCCommand once for TCP
	if manager.GetCallCount("RunLXCCommand") != 1 {
		t.Errorf("expected 1 RunLXCCommand call, got %d", manager.GetCallCount("RunLXCCommand"))
	}
}

func TestValidatePortForwardingArgsWithDefaults(t *testing.T) {
	// Test empty protocol is handled correctly
	err := validatePortForwardingArgs("test-container", "8080", "80", "")
	if err != nil {
		t.Errorf("should succeed with empty protocol (defaults to tcp): %v", err)
	}

	// Test that validation still works after default is applied
	err = validatePortForwardingArgs("test-container", "8080", "80", "BOTH")
	if err != nil {
		t.Errorf("should succeed with uppercase BOTH protocol: %v", err)
	}
}

func TestDefaultContainerPortManagerEdgeCases(t *testing.T) {
	manager := &DefaultContainerPortManager{}
	ctx := context.Background()

	// Test RunLXCCommand with no arguments
	err := manager.RunLXCCommand(ctx)
	if err == nil {
		t.Error("should fail with no arguments")
	}
	if !contains(err.Error(), "no command provided") {
		t.Errorf("expected 'no command provided' error, got: %v", err)
	}

	// Test RunLXCCommand with arguments (will fail but should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RunLXCCommand should not panic: %v", r)
		}
	}()

	err = manager.RunLXCCommand(ctx, "echo", "test")
	t.Logf("RunLXCCommand with arguments returned: %v", err)
}
