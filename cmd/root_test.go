package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test root command creation
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	// Test root command properties
	if rootCmd.Use != "lxc-go-cli" {
		t.Errorf("expected Use to be 'lxc-go-cli', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if rootCmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestRootCommandHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	// Set up command
	rootCmd.SetArgs([]string{"--help"})

	// Execute command
	err := rootCmd.Execute()

	// Close pipe and read output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check that help output contains expected content
	if !contains(output, "lxc-go-cli") {
		t.Error("help output should contain 'lxc-go-cli'")
	}

	if !contains(output, "create") {
		t.Error("help output should contain 'create' command")
	}
}

func TestRootCommandToggleFlag(t *testing.T) {
	// Test toggle flag
	rootCmd.SetArgs([]string{"--toggle"})

	// This should not error since toggle flag is defined
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("expected no error with toggle flag, got %v", err)
	}
}

func TestRootCommandInvalidFlag(t *testing.T) {
	// Test invalid flag
	rootCmd.SetArgs([]string{"--invalid-flag"})

	// This should error since invalid flag is not defined
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error with invalid flag")
	}
}

func TestRootCommandNoArgs(t *testing.T) {
	// Test command with no args
	rootCmd.SetArgs([]string{})

	// This should not error since no Run function is defined
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("expected no error with no args, got %v", err)
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test the Execute function
	// We'll test it by calling it and ensuring it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute function panicked: %v", r)
		}
	}()

	// Set up a simple test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Replace root command temporarily for testing
	originalRoot := rootCmd
	rootCmd = testCmd
	defer func() { rootCmd = originalRoot }()

	// Test Execute function
	Execute()
}

func TestExecuteFunctionWithError(t *testing.T) {
	// Test the Execute function with an error
	// Since Execute calls os.Exit(1) on error, we need to test it differently
	// We'll test that the error handling path works by checking the command execution directly

	// Set up a test command that returns an error
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return &testError{message: "test error"}
		},
	}

	// Test that the command returns an error when executed directly
	err := testCmd.Execute()
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Check that the error message is correct
	if err.Error() != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", err.Error())
	}
}

// Test error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
