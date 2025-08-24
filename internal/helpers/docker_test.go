package helpers

import (
	"fmt"
	"strings"
	"testing"
)

// MockDockerInstaller implements DockerInstaller for testing
type MockDockerInstaller struct {
	RunInContainerFunc func(containerName string, args ...string) error
	CallLog            [][]string // Track all calls for verification
}

func (m *MockDockerInstaller) RunInContainer(containerName string, args ...string) error {
	// Log the call
	callArgs := append([]string{containerName}, args...)
	m.CallLog = append(m.CallLog, callArgs)

	if m.RunInContainerFunc != nil {
		return m.RunInContainerFunc(containerName, args...)
	}
	return nil // Default success
}

func TestInstallDockerInContainer_Success(t *testing.T) {
	installer := &MockDockerInstaller{}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify all expected calls were made
	// Prerequisites(1) + Create dir(1) + Download GPG(1) + Permissions(1) + Repository(1) + Update(1) + Install(1) + Enable(1) + Start(1) + Docker version(1) + Compose version(1) = 11 calls
	expectedCalls := 11
	if len(installer.CallLog) != expectedCalls {
		t.Errorf("expected %d calls, got %d", expectedCalls, len(installer.CallLog))
		for i, call := range installer.CallLog {
			t.Logf("Call %d: %v", i+1, call)
		}
	}

	// Verify the first call is prerequisites installation
	if len(installer.CallLog) > 0 {
		prereqCall := installer.CallLog[0]
		if len(prereqCall) < 5 || prereqCall[1] != "apt-get" || prereqCall[2] != "install" {
			t.Errorf("first call should be apt-get install prerequisites, got %v", prereqCall)
		}
		// Check that ca-certificates and curl are in the prerequisites
		hasCACerts := false
		hasCurl := false
		for _, arg := range prereqCall {
			if arg == "ca-certificates" {
				hasCACerts = true
			}
			if arg == "curl" {
				hasCurl = true
			}
		}
		if !hasCACerts || !hasCurl {
			t.Errorf("prerequisites call missing expected packages: %v", prereqCall)
		}
	}

	// Verify the Docker package installation call
	var dockerInstallCall []string
	for _, call := range installer.CallLog {
		if len(call) > 6 && call[1] == "apt-get" && call[2] == "install" && strings.Contains(strings.Join(call, " "), "docker-ce") {
			dockerInstallCall = call
			break
		}
	}

	if len(dockerInstallCall) == 0 {
		t.Error("Docker package installation call not found")
	} else {
		// Check that official Docker packages are being installed
		hasSudo := false
		hasDockerCE := false
		hasDockerCECli := false
		hasContainerd := false
		hasCompose := false
		for _, arg := range dockerInstallCall {
			if arg == "sudo" {
				hasSudo = true
			}
			if arg == "docker-ce" {
				hasDockerCE = true
			}
			if arg == "docker-ce-cli" {
				hasDockerCECli = true
			}
			if arg == "containerd.io" {
				hasContainerd = true
			}
			if arg == "docker-compose-plugin" {
				hasCompose = true
			}
		}
		if !hasSudo || !hasDockerCE || !hasDockerCECli || !hasContainerd || !hasCompose {
			t.Errorf("Docker install command missing expected packages: %v", dockerInstallCall)
		}
	}
}

func TestInstallDockerInContainer_PrerequisitesFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 2 && args[0] == "apt-get" && args[1] == "install" {
				// Check if this is the prerequisites installation (ca-certificates, curl)
				for _, arg := range args {
					if arg == "ca-certificates" || arg == "curl" {
						return fmt.Errorf("network error")
					}
				}
			}
			return nil
		},
	}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to install prerequisites") {
		t.Errorf("expected error about prerequisites failure, got '%s'", err.Error())
	}
}

func TestInstallDockerInContainer_GPGKeyFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 3 && args[0] == "curl" && args[1] == "-fsSL" && args[2] == "https://download.docker.com/linux/ubuntu/gpg" {
				return fmt.Errorf("curl failed")
			}
			return nil
		},
	}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to download Docker GPG key") {
		t.Errorf("expected error about GPG key failure, got '%s'", err.Error())
	}
}

func TestInstallDockerInContainer_RepositoryFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 2 && args[0] == "sh" && args[1] == "-c" && strings.Contains(args[2], "echo \"deb") {
				return fmt.Errorf("repository add failed")
			}
			return nil
		},
	}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to add Docker repository") {
		t.Errorf("expected error about repository failure, got '%s'", err.Error())
	}
}

func TestInstallDockerInContainer_PackageInstallFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 2 && args[0] == "apt-get" && args[1] == "install" {
				// Check if this is the Docker packages installation (docker-ce, etc.)
				for _, arg := range args {
					if arg == "docker-ce" {
						return fmt.Errorf("package installation failed")
					}
				}
			}
			return nil
		},
	}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to install Docker packages") {
		t.Errorf("expected error about package installation failure, got '%s'", err.Error())
	}
}

func TestInstallDockerInContainer_ServiceEnableFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 1 && args[0] == "systemctl" && args[1] == "enable" {
				return fmt.Errorf("systemctl enable failed")
			}
			return nil
		},
	}
	containerName := "test-container"

	err := InstallDockerInContainer(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to enable Docker service") {
		t.Errorf("expected error about service enable failure, got '%s'", err.Error())
	}
}

func TestVerifyDockerInstallation_Success(t *testing.T) {
	installer := &MockDockerInstaller{}
	containerName := "test-container"

	err := VerifyDockerInstallation(installer, containerName)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should have 2 calls: docker --version and docker compose version
	if len(installer.CallLog) != 2 {
		t.Errorf("expected 2 calls, got %d", len(installer.CallLog))
	}

	// Verify the calls are correct
	if len(installer.CallLog) >= 2 {
		dockerCall := installer.CallLog[0]
		composeCall := installer.CallLog[1]

		if len(dockerCall) < 2 || dockerCall[1] != "docker" || dockerCall[2] != "--version" {
			t.Errorf("first call should be 'docker --version', got %v", dockerCall)
		}

		if len(composeCall) < 4 || composeCall[1] != "docker" || composeCall[2] != "compose" || composeCall[3] != "version" {
			t.Errorf("second call should be 'docker compose version', got %v", composeCall)
		}
	}
}

func TestVerifyDockerInstallation_DockerVersionFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 1 && args[0] == "docker" && args[1] == "--version" {
				return fmt.Errorf("docker not found")
			}
			return nil
		},
	}
	containerName := "test-container"

	err := VerifyDockerInstallation(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Docker verification failed") {
		t.Errorf("expected Docker verification error, got '%s'", err.Error())
	}
}

func TestVerifyDockerInstallation_ComposeVersionFailure(t *testing.T) {
	installer := &MockDockerInstaller{
		RunInContainerFunc: func(containerName string, args ...string) error {
			if len(args) > 2 && args[0] == "docker" && args[1] == "compose" && args[2] == "version" {
				return fmt.Errorf("compose not found")
			}
			return nil
		},
	}
	containerName := "test-container"

	err := VerifyDockerInstallation(installer, containerName)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Docker Compose V2 verification failed") {
		t.Errorf("expected Docker Compose V2 verification error, got '%s'", err.Error())
	}
}
