package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

// MockContainerManager for testing
type MockContainerManager struct {
	GetOrCreateBtrfsPoolFunc       func() (string, error)
	ContainerExistsFunc            func(name string) bool
	CreateContainerFunc            func(name, distro, release, arch, storagePool string) error
	ConfigureContainerSecurityFunc func(containerName string) error
	RunInContainerFunc             func(containerName string, args ...string) error
	RestartContainerFunc           func(name string) error
}

func (m *MockContainerManager) GetOrCreateBtrfsPool() (string, error) {
	if m.GetOrCreateBtrfsPoolFunc != nil {
		return m.GetOrCreateBtrfsPoolFunc()
	}
	return "", fmt.Errorf("GetOrCreateBtrfsPool not mocked")
}

func (m *MockContainerManager) ContainerExists(name string) bool {
	if m.ContainerExistsFunc != nil {
		return m.ContainerExistsFunc(name)
	}
	return false
}

func (m *MockContainerManager) CreateContainer(name, distro, release, arch, storagePool string) error {
	if m.CreateContainerFunc != nil {
		return m.CreateContainerFunc(name, distro, release, arch, storagePool)
	}
	return fmt.Errorf("CreateContainer not mocked")
}

func (m *MockContainerManager) ConfigureContainerSecurity(containerName string) error {
	if m.ConfigureContainerSecurityFunc != nil {
		return m.ConfigureContainerSecurityFunc(containerName)
	}
	return fmt.Errorf("ConfigureContainerSecurity not mocked")
}

func (m *MockContainerManager) RunInContainer(containerName string, args ...string) error {
	if m.RunInContainerFunc != nil {
		return m.RunInContainerFunc(containerName, args...)
	}
	return fmt.Errorf("RunInContainer not mocked")
}

func (m *MockContainerManager) RestartContainer(name string) error {
	if m.RestartContainerFunc != nil {
		return m.RestartContainerFunc(name)
	}
	return fmt.Errorf("RestartContainer not mocked")
}

func TestCreateCommand(t *testing.T) {
	// Test create command creation
	if createCmd == nil {
		t.Fatal("createCmd should not be nil")
	}

	// Test create command properties
	if createCmd.Use != "create" {
		t.Errorf("expected Use to be 'create', got '%s'", createCmd.Use)
	}

	if createCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if createCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestCreateCommandHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	// Set up command
	createCmd.SetArgs([]string{"--help"})

	// Execute command
	err := createCmd.Execute()

	// Close pipe and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check that help output contains expected content
	if !contains(output, "create") {
		t.Error("help output should contain 'create'")
	}

	if !contains(output, "Flags:") {
		t.Error("help output should contain 'Flags:'")
	}
}

func TestCreateContainerValidation(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		expectedError string
	}{
		{
			name:          "empty container name",
			containerName: "",
			expectedError: "container name is required",
		},
		{
			name:          "valid container name",
			containerName: "test-container",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &MockContainerManager{
				GetOrCreateBtrfsPoolFunc: func() (string, error) {
					return "test-pool", nil
				},
				ContainerExistsFunc: func(name string) bool {
					return false
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
			err := createContainer(manager, tt.containerName, "ubuntu:24.04", "10G")

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

func TestCreateContainerDefaultValues(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
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

	// Test with empty image and size (should use defaults)
	err := createContainer(manager, "test-container", "", "")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateContainerBtrfsPoolError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "", fmt.Errorf("btrfs not available")
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to get or create Btrfs storage pool") {
		t.Errorf("expected error about Btrfs storage pool, got '%s'", err.Error())
	}
}

func TestCreateContainerAlreadyExists(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return true
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("expected error about container already existing, got '%s'", err.Error())
	}
}

func TestCreateContainerCreationError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return fmt.Errorf("container creation failed")
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to create container") {
		t.Errorf("expected error about container creation, got '%s'", err.Error())
	}
}

func TestCreateContainerSecurityError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return fmt.Errorf("security configuration failed")
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to configure container security") {
		t.Errorf("expected error about security configuration, got '%s'", err.Error())
	}
}

func TestCreateContainerPackageUpdateError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return nil
		},
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 0 && args[0] == "apt-get" && len(args) > 1 && args[1] == "update" {
				return fmt.Errorf("package update failed")
			}
			return nil
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to update package index") {
		t.Errorf("expected error about package update, got '%s'", err.Error())
	}
}

func TestCreateContainerDockerInstallError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return nil
		},
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 0 && args[0] == "apt-get" && len(args) > 1 && args[1] == "install" {
				return fmt.Errorf("docker installation failed")
			}
			return nil
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to install Docker and Docker Compose") {
		t.Errorf("expected error about docker installation, got '%s'", err.Error())
	}
}

func TestCreateContainerUserCreationError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return nil
		},
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 0 && args[0] == "useradd" {
				return fmt.Errorf("user creation failed")
			}
			return nil
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to create 'app' user") {
		t.Errorf("expected error about user creation, got '%s'", err.Error())
	}
}

func TestCreateContainerUserGroupError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
		},
		CreateContainerFunc: func(name, distro, release, arch, storagePool string) error {
			return nil
		},
		ConfigureContainerSecurityFunc: func(containerName string) error {
			return nil
		},
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 0 && args[0] == "usermod" {
				return fmt.Errorf("user group modification failed")
			}
			return nil
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to add 'app' user to docker and sudo groups") {
		t.Errorf("expected error about user group modification, got '%s'", err.Error())
	}
}

func TestCreateContainerRestartError(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
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
			return fmt.Errorf("restart failed")
		},
	}

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !contains(err.Error(), "failed to restart container") {
		t.Errorf("expected error about container restart, got '%s'", err.Error())
	}
}

func TestCreateContainerSuccess(t *testing.T) {
	manager := &MockContainerManager{
		GetOrCreateBtrfsPoolFunc: func() (string, error) {
			return "test-pool", nil
		},
		ContainerExistsFunc: func(name string) bool {
			return false
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

	err := createContainer(manager, "test-container", "ubuntu:24.04", "10G")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateCommandFlags(t *testing.T) {
	// Test flag existence
	nameFlag := createCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("name flag should exist")
	}

	imageFlag := createCmd.Flags().Lookup("image")
	if imageFlag == nil {
		t.Error("image flag should exist")
	}
	if imageFlag.DefValue != "ubuntu:24.04" {
		t.Errorf("expected image flag default to be 'ubuntu:24.04', got '%s'", imageFlag.DefValue)
	}

	sizeFlag := createCmd.Flags().Lookup("size")
	if sizeFlag == nil {
		t.Error("size flag should exist")
	}
	if sizeFlag.DefValue != "10G" {
		t.Errorf("expected size flag default to be '10G', got '%s'", sizeFlag.DefValue)
	}
}

func TestDefaultContainerManager(t *testing.T) {
	// Test that DefaultContainerManager implements ContainerManager interface
	var manager ContainerManager = &DefaultContainerManager{}
	_ = manager // Use the variable to avoid unused variable warning
}

func TestDefaultContainerManagerMethods(t *testing.T) {
	// Test that all methods in the interface don't panic when called
	manager := &DefaultContainerManager{}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultContainerManager methods should not panic: %v", r)
		}
	}()

	// Test all interface methods (they will likely fail without LXC but should not panic)
	pool, err := manager.GetOrCreateBtrfsPool()
	t.Logf("GetOrCreateBtrfsPool returned: pool=%s, err=%v", pool, err)

	exists := manager.ContainerExists("test-container")
	t.Logf("ContainerExists returned: %v", exists)

	err = manager.CreateContainer("test", "ubuntu", "24.04", "amd64", "default")
	t.Logf("CreateContainer returned: %v", err)

	err = manager.ConfigureContainerSecurity("test")
	t.Logf("ConfigureContainerSecurity returned: %v", err)

	err = manager.RunInContainer("test", "echo", "hello")
	t.Logf("RunInContainer returned: %v", err)

	err = manager.RestartContainer("test")
	t.Logf("RestartContainer returned: %v", err)
}
