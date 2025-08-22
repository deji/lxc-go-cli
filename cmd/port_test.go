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
	ContainerExistsFunc    func(ctx context.Context, name string) bool
	RunLXCCommandFunc      func(ctx context.Context, args ...string) error
	GetContainerConfigFunc func(ctx context.Context, containerName string) ([]byte, error)
	ExistingContainers     map[string]bool
	RunCommandError        error
	GetConfigError         error
	ContainerConfigs       map[string][]byte
	Calls                  map[string]int
	LastCommand            []string
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

func (m *MockContainerPortManager) GetContainerConfig(ctx context.Context, containerName string) ([]byte, error) {
	m.trackCall("GetContainerConfig")
	if m.GetContainerConfigFunc != nil {
		return m.GetContainerConfigFunc(ctx, containerName)
	}
	if m.GetConfigError != nil {
		return nil, m.GetConfigError
	}
	if m.ContainerConfigs != nil {
		if config, exists := m.ContainerConfigs[containerName]; exists {
			return config, nil
		}
	}
	// Return empty config by default
	return []byte("devices: {}"), nil
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
	if portCmd.Use != "port <add|list>" {
		t.Errorf("expected 'port <add|list>', got '%s'", portCmd.Use)
	}

	if portCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if portCmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Test subcommands exist
	if portAddCmd == nil {
		t.Fatal("portAddCmd should not be nil")
	}

	if portListCmd == nil {
		t.Fatal("portListCmd should not be nil")
	}

	// Test port add command properties
	if portAddCmd.Use != "add <container-name> <host-port> <container-port> [tcp|udp|both]" {
		t.Errorf("expected specific Use format for add, got '%s'", portAddCmd.Use)
	}

	// Test port list command properties
	if portListCmd.Use != "list <container-name>" {
		t.Errorf("expected specific Use format for list, got '%s'", portListCmd.Use)
	}
}

func TestPortAddCommandArgs(t *testing.T) {
	// Test that the port add command expects 3-4 arguments
	if portAddCmd.Args == nil {
		t.Error("portAddCmd should have Args validation")
	}

	// Test with wrong number of args
	err := portAddCmd.Args(portAddCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	err = portAddCmd.Args(portAddCmd, []string{"container"})
	if err == nil {
		t.Error("should fail with too few arguments")
	}

	err = portAddCmd.Args(portAddCmd, []string{"container", "host"})
	if err == nil {
		t.Error("should fail with too few arguments")
	}

	err = portAddCmd.Args(portAddCmd, []string{"container", "host", "container", "protocol", "extra"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}

	// Test with correct number of args (should pass)
	err = portAddCmd.Args(portAddCmd, []string{"container", "8080", "80"})
	if err != nil {
		t.Errorf("should pass with three arguments (protocol optional): %v", err)
	}

	err = portAddCmd.Args(portAddCmd, []string{"container", "8080", "80", "tcp"})
	if err != nil {
		t.Errorf("should pass with four arguments: %v", err)
	}
}

func TestPortListCommandArgs(t *testing.T) {
	// Test that the port list command expects exactly 1 argument
	if portListCmd.Args == nil {
		t.Error("portListCmd should have Args validation")
	}

	// Test with wrong number of args
	err := portListCmd.Args(portListCmd, []string{})
	if err == nil {
		t.Error("should fail with no arguments")
	}

	err = portListCmd.Args(portListCmd, []string{"container", "extra"})
	if err == nil {
		t.Error("should fail with too many arguments")
	}

	// Test with correct number of args (should pass)
	err = portListCmd.Args(portListCmd, []string{"container"})
	if err != nil {
		t.Errorf("should pass with one argument: %v", err)
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

			err := configurePortForwarding(ctx, manager, tt.containerName, tt.hostPort, tt.containerPort, tt.protocol, false)

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

	err := configurePortForwardingForProtocol(ctx, manager, "test-container", "8080", "80", "tcp", false)
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
		"connect=tcp:0.0.0.0:80", "listen=tcp:0.0.0.0:8080",
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

	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "both", false)
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
	if !contains(strings.Join(tcpCmd, " "), "connect=tcp:0.0.0.0:80") {
		t.Error("TCP command should have correct connect parameter")
	}

	// Check UDP command
	udpCmd := commandHistory[1]
	if !contains(strings.Join(udpCmd, " "), "test-container-8080-80-udp") {
		t.Error("second command should be for UDP")
	}
	if !contains(strings.Join(udpCmd, " "), "connect=udp:0.0.0.0:80") {
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
	// Test timeout flag on port add command
	timeoutFlag := portAddCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag should exist on portAddCmd")
	}

	if timeoutFlag.Shorthand != "t" {
		t.Errorf("expected timeout flag shorthand to be 't', got '%s'", timeoutFlag.Shorthand)
	}

	if timeoutFlag.DefValue != "30s" {
		t.Errorf("expected timeout flag default to be '30s', got '%s'", timeoutFlag.DefValue)
	}

	// Test timeout flag on port list command
	timeoutFlag = portListCmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("timeout flag should exist on portListCmd")
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
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp", false)
	if err != nil {
		t.Errorf("should succeed with background context: %v", err)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should still work since our mock doesn't respect context cancellation
	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp", false)
	if err != nil {
		t.Errorf("should work with cancelled context in mock: %v", err)
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(2 * time.Millisecond)

	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp", false)
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
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "TCP", false)
	if err != nil {
		t.Errorf("should handle uppercase protocol: %v", err)
	}

	// Test mixed case protocol
	err = configurePortForwarding(ctx, manager, "test-container", "8080", "80", "BoTh", false)
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
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "both", false)
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
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "", false)
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

	// Test GetContainerConfig with empty name
	_, err = manager.GetContainerConfig(ctx, "")
	if err == nil {
		t.Error("should fail with empty container name")
	}
	if !contains(err.Error(), "container name is required") {
		t.Errorf("expected 'container name is required' error, got: %v", err)
	}
}

func TestPortForwardingWithForceFlag(t *testing.T) {
	ctx := context.Background()
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
	}

	// Test with force flag - should bypass port availability check
	err := configurePortForwarding(ctx, manager, "test-container", "8080", "80", "tcp", true)
	if err != nil {
		t.Errorf("should succeed with force flag: %v", err)
	}

	// Test without force flag on a commonly used port (likely to be taken)
	// This might fail in test environment due to port checking
	err = configurePortForwarding(ctx, manager, "test-container", "80", "80", "tcp", false)
	// We can't guarantee the result since it depends on the test environment
	t.Logf("Port 80 availability check result: %v", err)
}

func TestPortAddCommandFlags(t *testing.T) {
	// Test that the force flag exists
	forceFlag := portAddCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("force flag should exist on portAddCmd")
	}

	if forceFlag.Shorthand != "f" {
		t.Errorf("expected force flag shorthand to be 'f', got '%s'", forceFlag.Shorthand)
	}

	if forceFlag.DefValue != "false" {
		t.Errorf("expected force flag default to be 'false', got '%s'", forceFlag.DefValue)
	}
}

func TestPortAvailabilityIntegration(t *testing.T) {
	// Test that port availability checking is working in the command flow
	ctx := context.Background()

	// Create a mock that simulates port availability checking
	manager := &MockContainerPortManager{
		ExistingContainers: map[string]bool{
			"test-container": true,
		},
		RunLXCCommandFunc: func(ctx context.Context, args ...string) error {
			// Simulate successful LXC command
			return nil
		},
	}

	// Test with a very high port number that should be available
	err := configurePortForwarding(ctx, manager, "test-container", "65000", "80", "tcp", false)
	if err != nil {
		t.Errorf("should succeed with high port number: %v", err)
	}

	// Test force flag bypasses the check completely
	err = configurePortForwarding(ctx, manager, "test-container", "80", "80", "tcp", true)
	if err != nil {
		t.Errorf("should succeed with force flag even on low port: %v", err)
	}
}

func TestListPortForwarding(t *testing.T) {
	tests := []struct {
		name            string
		containerName   string
		containerExists bool
		configData      string
		configError     error
		expectedError   string
		expectedOutput  string
	}{
		{
			name:          "empty container name",
			containerName: "",
			expectedError: "container name is required",
		},
		{
			name:            "container does not exist",
			containerName:   "nonexistent",
			containerExists: false,
			expectedError:   "container 'nonexistent' does not exist",
		},
		{
			name:            "config error",
			containerName:   "test-container",
			containerExists: true,
			configError:     fmt.Errorf("config command failed"),
			expectedError:   "failed to get container configuration",
		},
		{
			name:            "no port mappings",
			containerName:   "test-container",
			containerExists: true,
			configData:      "devices: {}",
			expectedOutput:  "No port forwarding rules found for container 'test-container'",
		},
		{
			name:            "valid port mappings",
			containerName:   "test-container",
			containerExists: true,
			configData: `devices:
  test-container-8080-80-tcp:
    type: proxy
    connect: tcp:0.0.0.0:8080
    listen: tcp:0.0.0.0:80
  test-container-5432-5432-udp:
    type: proxy
    connect: udp:127.0.0.1:5432
    listen: udp:0.0.0.0:5432
  non-port-device:
    type: disk
    path: /mnt`,
			expectedOutput: "Port mappings for container 'test-container':",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			manager := &MockContainerPortManager{
				ExistingContainers: map[string]bool{
					"test-container": tt.containerExists,
				},
				ContainerConfigs: map[string][]byte{
					"test-container": []byte(tt.configData),
				},
				GetConfigError: tt.configError,
			}

			err := listPortForwarding(ctx, manager, tt.containerName)

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

func TestParsePortMappingsFromConfig(t *testing.T) {
	tests := []struct {
		name          string
		yamlData      string
		containerName string
		expectedCount int
		expectedError string
	}{
		{
			name:          "invalid yaml",
			yamlData:      "invalid: yaml: content:",
			containerName: "test-container",
			expectedError: "failed to parse container configuration",
		},
		{
			name:          "empty devices",
			yamlData:      "devices: {}",
			containerName: "test-container",
			expectedCount: 0,
		},
		{
			name: "valid port mappings",
			yamlData: `devices:
  test-container-8080-80-tcp:
    type: proxy
    connect: tcp:0.0.0.0:8080
    listen: tcp:0.0.0.0:80
  test-container-5432-5432-udp:
    type: proxy
    connect: udp:127.0.0.1:5432
    listen: udp:0.0.0.0:5432
  non-port-device:
    type: disk
    path: /mnt
  other-container-8080-80-tcp:
    type: proxy
    connect: tcp:0.0.0.0:8080
    listen: tcp:0.0.0.0:80`,
			containerName: "test-container",
			expectedCount: 2, // Only devices matching container name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mappings, err := parsePortMappingsFromConfig([]byte(tt.yamlData), tt.containerName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if len(mappings) != tt.expectedCount {
				t.Errorf("expected %d mappings, got %d", tt.expectedCount, len(mappings))
			}
		})
	}
}

func TestIsPortDevice(t *testing.T) {
	tests := []struct {
		name          string
		deviceName    string
		containerName string
		expected      bool
	}{
		{
			name:          "valid tcp device",
			deviceName:    "test-container-8080-80-tcp",
			containerName: "test-container",
			expected:      true,
		},
		{
			name:          "valid udp device",
			deviceName:    "test-container-5432-5432-udp",
			containerName: "test-container",
			expected:      true,
		},
		{
			name:          "wrong container name",
			deviceName:    "other-container-8080-80-tcp",
			containerName: "test-container",
			expected:      false,
		},
		{
			name:          "invalid protocol",
			deviceName:    "test-container-8080-80-http",
			containerName: "test-container",
			expected:      false,
		},
		{
			name:          "invalid format - too few parts",
			deviceName:    "test-container-8080-tcp",
			containerName: "test-container",
			expected:      false,
		},
		{
			name:          "invalid format - non-numeric port",
			deviceName:    "test-container-abc-80-tcp",
			containerName: "test-container",
			expected:      false,
		},
		{
			name:          "container name with special chars",
			deviceName:    "test-container-2-8080-80-tcp",
			containerName: "test-container-2",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPortDevice(tt.deviceName, tt.containerName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		name          string
		deviceName    string
		device        Device
		expectedPort  *PortMapping
		expectedError string
	}{
		{
			name:       "valid tcp mapping",
			deviceName: "test-container-8080-80-tcp",
			device: Device{
				Type:    "proxy",
				Connect: "tcp:0.0.0.0:8080",
				Listen:  "tcp:0.0.0.0:80",
			},
			expectedPort: &PortMapping{
				DeviceName:    "test-container-8080-80-tcp",
				Protocol:      "TCP",
				HostPort:      "8080",
				ContainerPort: "80",
				HostIP:        "0.0.0.0",
				ContainerIP:   "0.0.0.0",
			},
		},
		{
			name:       "valid udp mapping with different IPs",
			deviceName: "test-container-5432-5432-udp",
			device: Device{
				Type:    "proxy",
				Connect: "udp:127.0.0.1:5432",
				Listen:  "udp:192.168.1.1:5432",
			},
			expectedPort: &PortMapping{
				DeviceName:    "test-container-5432-5432-udp",
				Protocol:      "UDP",
				HostPort:      "5432",
				ContainerPort: "5432",
				HostIP:        "127.0.0.1",
				ContainerIP:   "192.168.1.1",
			},
		},
		{
			name:          "invalid device name format",
			deviceName:    "invalid-format",
			device:        Device{Type: "proxy"},
			expectedError: "invalid device name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := parsePortMapping(tt.deviceName, tt.device)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else {
				if mapping.DeviceName != tt.expectedPort.DeviceName {
					t.Errorf("expected DeviceName '%s', got '%s'", tt.expectedPort.DeviceName, mapping.DeviceName)
				}
				if mapping.Protocol != tt.expectedPort.Protocol {
					t.Errorf("expected Protocol '%s', got '%s'", tt.expectedPort.Protocol, mapping.Protocol)
				}
				if mapping.HostPort != tt.expectedPort.HostPort {
					t.Errorf("expected HostPort '%s', got '%s'", tt.expectedPort.HostPort, mapping.HostPort)
				}
				if mapping.ContainerPort != tt.expectedPort.ContainerPort {
					t.Errorf("expected ContainerPort '%s', got '%s'", tt.expectedPort.ContainerPort, mapping.ContainerPort)
				}
				if mapping.HostIP != tt.expectedPort.HostIP {
					t.Errorf("expected HostIP '%s', got '%s'", tt.expectedPort.HostIP, mapping.HostIP)
				}
				if mapping.ContainerIP != tt.expectedPort.ContainerIP {
					t.Errorf("expected ContainerIP '%s', got '%s'", tt.expectedPort.ContainerIP, mapping.ContainerIP)
				}
			}
		})
	}
}

func TestFormatPortMappings(t *testing.T) {
	tests := []struct {
		name     string
		mappings []PortMapping
		expected string
	}{
		{
			name:     "empty mappings",
			mappings: []PortMapping{},
			expected: "",
		},
		{
			name: "single mapping",
			mappings: []PortMapping{
				{
					DeviceName:    "test-container-8080-80-tcp",
					Protocol:      "TCP",
					HostPort:      "8080",
					ContainerPort: "80",
					HostIP:        "0.0.0.0",
					ContainerIP:   "0.0.0.0",
				},
			},
			expected: "PROTOCOL  HOST PORT  CONTAINER PORT  HOST IP      CONTAINER IP  DEVICE NAME",
		},
		{
			name: "multiple mappings",
			mappings: []PortMapping{
				{
					DeviceName:    "test-container-8080-80-tcp",
					Protocol:      "TCP",
					HostPort:      "8080",
					ContainerPort: "80",
					HostIP:        "0.0.0.0",
					ContainerIP:   "0.0.0.0",
				},
				{
					DeviceName:    "test-container-5432-5432-udp",
					Protocol:      "UDP",
					HostPort:      "5432",
					ContainerPort: "5432",
					HostIP:        "127.0.0.1",
					ContainerIP:   "192.168.1.1",
				},
			},
			expected: "PROTOCOL  HOST PORT  CONTAINER PORT  HOST IP      CONTAINER IP  DEVICE NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPortMappings(tt.mappings)
			if tt.expected == "" {
				if result != "" {
					t.Errorf("expected empty string, got '%s'", result)
				}
			} else {
				if !contains(result, tt.expected) {
					t.Errorf("expected result to contain '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}
