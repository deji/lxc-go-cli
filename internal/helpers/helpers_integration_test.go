//go:build integration

package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests using mocks with context
// Run with: go test -tags=integration ./...

// createMockLXC creates a mock LXC instance for testing
func createMockLXC() *MockLXC {
	return NewMockLXC()
}

// createRealLXC creates a real LXC instance for testing when LXC_REAL=1
func createRealLXC() *RealLXC {
	return NewRealLXC()
}

// getLXCInterface returns either mock or real LXC based on environment
func getLXCInterface(t *testing.T) LXCInterface {
	if os.Getenv("LXC_REAL") == "1" {
		t.Log("Using real LXC interface")
		return createRealLXC()
	}
	t.Log("Using mock LXC interface")
	return createMockLXC()
}

func TestGetBtrfsStoragePools_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	pools := lxc.GetBtrfsStoragePools(ctx)

	t.Logf("GetBtrfsStoragePools returned: %v", pools)

	// With mock, we should have some default pools and pools should not be nil
	if mock, ok := lxc.(*MockLXC); ok {
		if pools == nil {
			t.Error("GetBtrfsStoragePools should not return nil with mock")
		}

		if len(pools) == 0 {
			t.Error("Mock should return some default pools")
		}

		// Verify call was tracked
		if mock.GetCallCount("GetBtrfsStoragePools") != 1 {
			t.Error("Expected GetBtrfsStoragePools to be called once")
		}
	} else {
		// With real LXC, it might return nil or empty slice depending on system state
		// This is acceptable as LXC may not be available or configured
		t.Logf("Real LXC returned %d pools", len(pools))
	}
}

func TestIsBtrfsAvailable_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	available := lxc.IsBtrfsAvailable(ctx)

	t.Logf("IsBtrfsAvailable returned: %v", available)

	// With mock, this should be configurable and default to true
	if mock, ok := lxc.(*MockLXC); ok {
		if !available {
			t.Error("Mock should default to Btrfs being available")
		}

		// Test setting it to false
		mock.SetBtrfsAvailable(false)
		if lxc.IsBtrfsAvailable(ctx) {
			t.Error("Should respect mock configuration")
		}

		// Verify call was tracked
		if mock.GetCallCount("IsBtrfsAvailable") < 2 {
			t.Error("Expected IsBtrfsAvailable to be called at least twice")
		}
	}
}

func TestGetDefaultStoragePoolType_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	poolType := lxc.GetDefaultStoragePoolType(ctx)

	t.Logf("GetDefaultStoragePoolType returned: %v", poolType)

	// With mock, should return a default type
	if mock, ok := lxc.(*MockLXC); ok {
		if poolType == "" {
			t.Error("Mock should return a default pool type")
		}

		if mock.GetCallCount("GetDefaultStoragePoolType") != 1 {
			t.Error("Expected GetDefaultStoragePoolType to be called once")
		}
	}
}

func TestContainerExists_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	testContainer := "nonexistent-container-test-12345"

	// This container should not exist initially
	exists := lxc.ContainerExists(ctx, testContainer)
	if exists {
		t.Errorf("Container '%s' should not exist initially", testContainer)
	}

	// With mock, we can add it and test again
	if mock, ok := lxc.(*MockLXC); ok {
		mock.AddContainer(testContainer)
		if !lxc.ContainerExists(ctx, testContainer) {
			t.Error("Container should exist after adding to mock")
		}

		mock.RemoveContainer(testContainer)
		if lxc.ContainerExists(ctx, testContainer) {
			t.Error("Container should not exist after removing from mock")
		}

		if mock.GetCallCount("ContainerExists") < 3 {
			t.Error("Expected ContainerExists to be called at least 3 times")
		}
	}
}

func TestCreateContainer_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	containerName := "test-container-temp"

	// Skip for real LXC to avoid creating actual containers
	if _, ok := lxc.(*RealLXC); ok {
		t.Skip("Skipping CreateContainer test with real LXC - would create actual containers")
	}

	// Test with mock
	if mock, ok := lxc.(*MockLXC); ok {
		// Should succeed with valid parameters
		err := lxc.CreateContainer(ctx, containerName, "ubuntu", "24.04", "amd64", "default-btrfs")
		if err != nil {
			t.Errorf("CreateContainer failed: %v", err)
		}

		// Container should exist now
		if !lxc.ContainerExists(ctx, containerName) {
			t.Error("Container should exist after creation")
		}

		// Try to create the same container again (should fail)
		err = lxc.CreateContainer(ctx, containerName, "ubuntu", "24.04", "amd64", "default-btrfs")
		if err == nil {
			t.Error("Creating duplicate container should fail")
		}

		// Test with non-existent storage pool
		err = lxc.CreateContainer(ctx, "another-container", "ubuntu", "24.04", "amd64", "nonexistent-pool")
		if err == nil {
			t.Error("Creating container with non-existent pool should fail")
		}

		if mock.GetCallCount("CreateContainer") < 3 {
			t.Error("Expected CreateContainer to be called at least 3 times")
		}
	}
}

func TestStartContainer_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	testContainer := "nonexistent-container-test-12345"

	// Should return an error since container doesn't exist
	err := lxc.StartContainer(ctx, testContainer)
	if err == nil {
		t.Error("StartContainer should return error for non-existent container")
	}

	// With mock, test successful start
	if mock, ok := lxc.(*MockLXC); ok {
		mock.AddContainer(testContainer)
		err = lxc.StartContainer(ctx, testContainer)
		if err != nil {
			t.Errorf("StartContainer should succeed for existing container: %v", err)
		}

		if mock.GetCallCount("StartContainer") < 2 {
			t.Error("Expected StartContainer to be called at least twice")
		}
	}
}

func TestRestartContainer_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	testContainer := "nonexistent-container-test-12345"

	// Should return an error since container doesn't exist
	err := lxc.RestartContainer(ctx, testContainer)
	if err == nil {
		t.Error("RestartContainer should return error for non-existent container")
	}

	// With mock, test successful restart
	if mock, ok := lxc.(*MockLXC); ok {
		mock.AddContainer(testContainer)
		err = lxc.RestartContainer(ctx, testContainer)
		if err != nil {
			t.Errorf("RestartContainer should succeed for existing container: %v", err)
		}

		if mock.GetCallCount("RestartContainer") < 2 {
			t.Error("Expected RestartContainer to be called at least twice")
		}
	}
}

func TestRunInContainer_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	testContainer := "nonexistent-container-test-12345"

	// Should return an error since container doesn't exist
	err := lxc.RunInContainer(ctx, testContainer, "echo", "test")
	if err == nil {
		t.Error("RunInContainer should return error for non-existent container")
	}

	// With mock, test successful command execution
	if mock, ok := lxc.(*MockLXC); ok {
		mock.AddContainer(testContainer)

		// Test various commands
		commands := [][]string{
			{"echo", "test"},
			{"apt-get", "update"},
			{"apt-get", "install", "-y", "ca-certificates", "curl"},
			{"install", "-m", "0755", "-d", "/etc/apt/keyrings"},
			{"curl", "-fsSL", "https://download.docker.com/linux/ubuntu/gpg", "-o", "/etc/apt/keyrings/docker.asc"},
			{"chmod", "a+r", "/etc/apt/keyrings/docker.asc"},
			{"apt-get", "update"},
			{"apt-get", "install", "-y", "sudo", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
			{"systemctl", "enable", "docker"},
			{"systemctl", "start", "docker"},
			{"docker", "--version"},
			{"docker", "compose", "version"},
			{"useradd", "-m", "testuser"},
			{"usermod", "-aG", "docker", "testuser"},
		}

		for _, cmd := range commands {
			err = lxc.RunInContainer(ctx, testContainer, cmd...)
			if err != nil {
				t.Errorf("RunInContainer should succeed for command %v: %v", cmd, err)
			}
		}

		if mock.GetCallCount("RunInContainer") < len(commands)+1 {
			t.Errorf("Expected RunInContainer to be called at least %d times", len(commands)+1)
		}
	}
}

func TestConfigureContainerSecurity_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	testContainer := "nonexistent-container-test-12345"

	// Should return an error since container doesn't exist
	err := lxc.ConfigureContainerSecurity(ctx, testContainer)
	if err == nil {
		t.Error("ConfigureContainerSecurity should return error for non-existent container")
	}

	// With mock, test successful configuration
	if mock, ok := lxc.(*MockLXC); ok {
		mock.AddContainer(testContainer)
		err = lxc.ConfigureContainerSecurity(ctx, testContainer)
		if err != nil {
			t.Errorf("ConfigureContainerSecurity should succeed for existing container: %v", err)
		}

		if mock.GetCallCount("ConfigureContainerSecurity") < 2 {
			t.Error("Expected ConfigureContainerSecurity to be called at least twice")
		}
	}
}

func TestSetDefaultStoragePool_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)

	// Skip for real LXC to avoid modifying system configuration
	if _, ok := lxc.(*RealLXC); ok {
		t.Skip("Skipping SetDefaultStoragePool test with real LXC - would modify system configuration")
	}

	// Test with mock
	if mock, ok := lxc.(*MockLXC); ok {
		// Should fail for non-existent pool
		err := lxc.SetDefaultStoragePool(ctx, "nonexistent-pool-test-12345")
		if err == nil {
			t.Error("SetDefaultStoragePool should return error for non-existent pool")
		}

		// Should succeed for existing pool
		err = lxc.SetDefaultStoragePool(ctx, "default-btrfs")
		if err != nil {
			t.Errorf("SetDefaultStoragePool should succeed for existing pool: %v", err)
		}

		if mock.GetCallCount("SetDefaultStoragePool") < 2 {
			t.Error("Expected SetDefaultStoragePool to be called at least twice")
		}
	}
}

func TestCreateBtrfsStoragePool_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)

	// Skip for real LXC to avoid creating actual storage pools
	if _, ok := lxc.(*RealLXC); ok {
		t.Skip("Skipping CreateBtrfsStoragePool test with real LXC - would create actual storage pools")
	}

	// Test with mock
	if mock, ok := lxc.(*MockLXC); ok {
		poolName := "test-pool-temp"

		// Should succeed for new pool
		err := lxc.CreateBtrfsStoragePool(ctx, poolName)
		if err != nil {
			t.Errorf("CreateBtrfsStoragePool should succeed: %v", err)
		}

		// Should fail for duplicate pool
		err = lxc.CreateBtrfsStoragePool(ctx, poolName)
		if err == nil {
			t.Error("CreateBtrfsStoragePool should fail for duplicate pool")
		}

		if mock.GetCallCount("CreateBtrfsStoragePool") < 2 {
			t.Error("Expected CreateBtrfsStoragePool to be called at least twice")
		}
	}
}

func TestGetOrCreateBtrfsPool_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	pool, err := lxc.GetOrCreateBtrfsPool(ctx)

	t.Logf("GetOrCreateBtrfsPool returned: pool=%s, err=%v", pool, err)

	// With mock, should succeed and return a pool name
	if mock, ok := lxc.(*MockLXC); ok {
		if err != nil {
			t.Errorf("GetOrCreateBtrfsPool should succeed with mock: %v", err)
		}

		if pool == "" {
			t.Error("GetOrCreateBtrfsPool should return a non-empty pool name")
		}

		// Test with Btrfs unavailable
		mock.SetBtrfsAvailable(false)
		_, err = lxc.GetOrCreateBtrfsPool(ctx)
		if err == nil {
			t.Error("GetOrCreateBtrfsPool should fail when Btrfs is unavailable")
		}

		if mock.GetCallCount("GetOrCreateBtrfsPool") < 2 {
			t.Error("Expected GetOrCreateBtrfsPool to be called at least twice")
		}
	}
}

func TestEnsureBtrfsStoragePool_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)
	err := lxc.EnsureBtrfsStoragePool(ctx)

	t.Logf("EnsureBtrfsStoragePool returned: %v", err)

	// With mock, should succeed when properly configured
	if mock, ok := lxc.(*MockLXC); ok {
		if err != nil {
			t.Errorf("EnsureBtrfsStoragePool should succeed with mock: %v", err)
		}

		// Test with Btrfs unavailable
		mock.SetBtrfsAvailable(false)
		err = lxc.EnsureBtrfsStoragePool(ctx)
		if err == nil {
			t.Error("EnsureBtrfsStoragePool should fail when Btrfs is unavailable")
		}

		if mock.GetCallCount("EnsureBtrfsStoragePool") < 2 {
			t.Error("Expected EnsureBtrfsStoragePool to be called at least twice")
		}
	}
}

// TestErrorInjection tests the mock's error injection capabilities
func TestErrorInjection_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lxc := getLXCInterface(t)

	// Only test with mock
	mock, ok := lxc.(*MockLXC)
	if !ok {
		t.Skip("Error injection test only works with mock")
	}

	// Test various error injections
	testContainer := "test-container"

	// Test CreateContainer error
	mock.SetError("createcontainer", fmt.Errorf("mock create error"))
	err := lxc.CreateContainer(ctx, testContainer, "ubuntu", "24.04", "amd64", "default-btrfs")
	if err == nil || err.Error() != "mock create error" {
		t.Error("Should return injected create error")
	}

	// Reset and test StartContainer error
	mock.Reset()
	mock.AddContainer(testContainer)
	mock.SetError("startcontainer", fmt.Errorf("mock start error"))
	err = lxc.StartContainer(ctx, testContainer)
	if err == nil || err.Error() != "mock start error" {
		t.Error("Should return injected start error")
	}

	// Test RunCommand error
	mock.Reset()
	mock.AddContainer(testContainer)
	mock.SetError("runcommand", fmt.Errorf("mock run error"))
	err = lxc.RunInContainer(ctx, testContainer, "echo", "test")
	if err == nil || err.Error() != "mock run error" {
		t.Error("Should return injected run error")
	}
}
