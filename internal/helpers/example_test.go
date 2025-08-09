//go:build integration

package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ExampleMockUsage demonstrates how to use the MockLXC for testing
func ExampleMockUsage(t *testing.T) {
	// Create a mock LXC instance
	mock := NewMockLXC()

	// Configure the mock for your test scenario
	mock.SetBtrfsAvailable(true)
	mock.AddPool("test-pool")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test your code using the mock
	pools := mock.GetBtrfsStoragePools(ctx)
	if len(pools) == 0 {
		t.Error("Expected at least one pool")
	}

	// Test container operations
	containerName := "test-container"
	err := mock.CreateContainer(ctx, containerName, "ubuntu", "24.04", "amd64", "test-pool")
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	// Verify the container exists
	if !mock.ContainerExists(ctx, containerName) {
		t.Error("Container should exist after creation")
	}

	// Test error injection
	mock.SetError("startcontainer", fmt.Errorf("simulated start failure"))
	err = mock.StartContainer(ctx, containerName)
	if err == nil {
		t.Error("Expected start to fail due to injected error")
	}

	// Check call tracking
	if mock.GetCallCount("CreateContainer") != 1 {
		t.Error("Expected CreateContainer to be called once")
	}

	// Reset the mock for the next test
	mock.Reset()
}

// ExampleRealLXCUsage demonstrates how to conditionally use real LXC
func ExampleRealLXCUsage(t *testing.T) {
	// You can decide at runtime whether to use mock or real LXC
	var lxc LXCInterface

	if testing.Short() {
		// Use mock for fast tests
		lxc = NewMockLXC()
		t.Log("Using mock LXC for short test")
	} else {
		// Use real LXC for thorough integration testing
		lxc = NewRealLXC()
		t.Log("Using real LXC for integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Your test code works the same regardless of implementation
	available := lxc.IsBtrfsAvailable(ctx)
	t.Logf("Btrfs available: %v", available)

	pools := lxc.GetBtrfsStoragePools(ctx)
	t.Logf("Found %d pools", len(pools))
}

// ExampleContextUsage demonstrates proper context usage
func ExampleContextUsage(t *testing.T) {
	mock := NewMockLXC()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately to test cancellation handling
	cancel()

	// Operations should respect context cancellation
	// (In this mock implementation, context is not fully utilized,
	// but in real implementations it would be)
	pools := mock.GetBtrfsStoragePools(ctx)
	t.Logf("Pools returned even with cancelled context: %v", pools)

	// Test with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Very short timeout - in real implementation this might fail
	time.Sleep(2 * time.Millisecond) // Ensure timeout passes

	available := mock.IsBtrfsAvailable(ctx)
	t.Logf("Btrfs available with expired context: %v", available)
}
