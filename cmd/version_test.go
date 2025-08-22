package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	// Test that the version command exists
	if versionCmd == nil {
		t.Fatal("version command should exist")
	}

	if versionCmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short == "" {
		t.Error("version command should have a short description")
	}

	if versionCmd.Long == "" {
		t.Error("version command should have a long description")
	}
}

func TestVersionCommandArgs(t *testing.T) {
	// Version command should accept no arguments
	// Since versionCmd doesn't set Args explicitly, it accepts any args by default
	// This is fine for the version command - extra args are just ignored

	// Test that the command can be called with no args
	if versionCmd.Args != nil {
		err := versionCmd.Args(versionCmd, []string{})
		if err != nil {
			t.Errorf("Version command should accept no arguments: %v", err)
		}
	}

	// Test that the command can be called with extra args (should be ignored)
	if versionCmd.Args != nil {
		err := versionCmd.Args(versionCmd, []string{"extra", "args"})
		if err != nil {
			t.Logf("Version command with extra args: %v", err)
		}
	}
}

func TestVersionCommandFlags(t *testing.T) {
	// Test that the detailed flag exists
	detailedFlag := versionCmd.Flags().Lookup("detailed")
	if detailedFlag == nil {
		t.Error("detailed flag should exist on versionCmd")
	}

	if detailedFlag.Shorthand != "d" {
		t.Errorf("expected detailed flag shorthand to be 'd', got '%s'", detailedFlag.Shorthand)
	}

	if detailedFlag.DefValue != "false" {
		t.Errorf("expected detailed flag default to be 'false', got '%s'", detailedFlag.DefValue)
	}
}

func TestSetVersionInfo(t *testing.T) {
	// Save original values
	originalVersion := version
	originalGitCommit := gitCommit
	originalBuildTime := buildTime

	// Test setting version info
	testVersion := "1.abcd.20250102123456"
	testCommit := "1234567890abcdef"
	testTime := "20250102123456"

	SetVersionInfo(testVersion, testCommit, testTime)

	if version != testVersion {
		t.Errorf("expected version to be '%s', got '%s'", testVersion, version)
	}

	if gitCommit != testCommit {
		t.Errorf("expected gitCommit to be '%s', got '%s'", testCommit, gitCommit)
	}

	if buildTime != testTime {
		t.Errorf("expected buildTime to be '%s', got '%s'", testTime, buildTime)
	}

	// Restore original values
	SetVersionInfo(originalVersion, originalGitCommit, originalBuildTime)
}

func TestShowVersion(t *testing.T) {
	// Set test version info
	testVersion := "1.abcd.20250102123456"
	testCommit := "1234567890abcdef"
	testTime := "20250102123456"

	SetVersionInfo(testVersion, testCommit, testTime)

	tests := []struct {
		name     string
		detailed bool
		contains []string
	}{
		{
			name:     "simple version",
			detailed: false,
			contains: []string{
				"lxc-go-cli 1.abcd.20250102123456",
			},
		},
		{
			name:     "detailed version",
			detailed: true,
			contains: []string{
				"lxc-go-cli version information:",
				"Version:    1.abcd.20250102123456",
				"Git Commit: 1234567890abcdef",
				"Build Time: 20250102123456",
				"Go Version:",
				"Platform:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock command with the detailed flag
			cmd := &cobra.Command{}
			cmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")

			if tt.detailed {
				cmd.Flags().Set("detailed", "true")
			}

			// Capture output by setting cmd's output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Run the show version function
			showVersion(cmd)

			output := buf.String()

			// Check that all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("output should contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

func TestVersionCommandExecution(t *testing.T) {
	// Test actual command execution by calling showVersion directly
	testVersion := "1.test.20250102123456"
	SetVersionInfo(testVersion, "test-commit", "test-time")

	// Create a command with proper output capturing
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")

	// Capture output using a buffer
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Call showVersion directly
	showVersion(cmd)

	output := buf.String()

	// Should contain the version
	if !strings.Contains(output, testVersion) {
		t.Errorf("output should contain version '%s', got: '%s'", testVersion, output)
	}
}

func TestVersionIsRegistered(t *testing.T) {
	// Check that version command is registered with root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command should be registered with root command")
	}
}
