package cmd

import (
	"github.com/yourusername/lxc-go-cli/internal/logger"
)

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// setupQuietTesting sets up quiet logging for tests that don't need log output
// Returns a cleanup function that should be called with defer
func setupQuietTesting() func() {
	return logger.QuietTests()
}

// setupVerboseTesting sets up debug-level logging for tests that need detailed output
// Returns a cleanup function that should be called with defer
func setupVerboseTesting() func() {
	originalLevel := logger.GetLevel()
	logger.SetLevel(logger.DEBUG)
	return func() {
		logger.SetLevel(originalLevel)
	}
}
