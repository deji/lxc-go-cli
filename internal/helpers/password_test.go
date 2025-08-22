package helpers

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateSecurePassword(t *testing.T) {
	// Test password generation multiple times to ensure randomness
	passwords := make(map[string]bool)

	for i := 0; i < 10; i++ {
		password := GenerateSecurePassword()

		// Test length
		if len(password) != 16 {
			t.Errorf("expected password length 16, got %d", len(password))
		}

		// Test character set (letters and numbers only)
		validChars := regexp.MustCompile(`^[A-Za-z0-9]+$`)
		if !validChars.MatchString(password) {
			t.Errorf("password contains invalid characters: %s", password)
		}

		// Test uniqueness (very unlikely to get duplicates with secure random)
		if passwords[password] {
			t.Errorf("generated duplicate password: %s", password)
		}
		passwords[password] = true

		// Test that it contains both letters and numbers
		hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

		if !hasLetter {
			t.Errorf("password should contain at least one letter: %s", password)
		}
		if !hasNumber {
			t.Errorf("password should contain at least one number: %s", password)
		}
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
	}{
		{
			name:    "short string with numbers",
			length:  5,
			charset: "12345",
		},
		{
			name:    "medium string with letters",
			length:  10,
			charset: "abcde",
		},
		{
			name:    "single character",
			length:  1,
			charset: "X",
		},
		{
			name:    "empty charset fallback",
			length:  8,
			charset: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.charset == "" {
				// Test fallback behavior - should return default pattern
				result := generateRandomString(tt.length, tt.charset)
				if len(result) < tt.length {
					t.Errorf("expected at least %d characters, got %d", tt.length, len(result))
				}
				return
			}

			result := generateRandomString(tt.length, tt.charset)

			// Test length
			if len(result) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(result))
			}

			// Test character set
			for _, char := range result {
				if !strings.ContainsRune(tt.charset, char) {
					t.Errorf("result contains character not in charset: %c", char)
				}
			}
		})
	}
}

func TestContainerHasPassword(t *testing.T) {
	// Note: This test will fail in test environment without LXC
	// but tests the logic flow

	tests := []struct {
		name          string
		containerName string
		expected      bool
	}{
		{
			name:          "nonexistent container",
			containerName: "nonexistent-container",
			expected:      false,
		},
		{
			name:          "empty container name",
			containerName: "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainerHasPassword(tt.containerName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormatPasswordDisplay(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		password      string
		expected      []string // strings that should be present
	}{
		{
			name:          "normal case",
			containerName: "test-container",
			password:      "MyPassword123",
			expected: []string{
				"Password for 'app' user in 'test-container':",
				"MyPassword123",
			},
		},
		{
			name:          "container with special chars",
			containerName: "test-container-2",
			password:      "SecurePass456",
			expected: []string{
				"Password for 'app' user in 'test-container-2':",
				"SecurePass456",
			},
		},
		{
			name:          "empty password",
			containerName: "test-container",
			password:      "",
			expected: []string{
				"Password for 'app' user in 'test-container':",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPasswordDisplay(tt.containerName, tt.password)

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("expected output to contain '%s', got:\n%s", expected, result)
				}
			}

			// Check that output ends with newline
			if !strings.HasSuffix(result, "\n") {
				t.Error("output should end with newline")
			}
		})
	}
}

// Test password generation distribution
func TestPasswordGenerationDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping distribution test in short mode")
	}

	const numPasswords = 100
	charCounts := make(map[rune]int)

	for i := 0; i < numPasswords; i++ {
		password := GenerateSecurePassword()
		for _, char := range password {
			charCounts[char]++
		}
	}

	// Check that we're using a reasonable variety of characters
	if len(charCounts) < 20 {
		t.Errorf("expected at least 20 different characters used, got %d", len(charCounts))
	}

	// Check that both uppercase, lowercase, and numbers are used
	hasUpper := false
	hasLower := false
	hasDigit := false

	for char := range charCounts {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= 'a' && char <= 'z' {
			hasLower = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	if !hasUpper {
		t.Error("no uppercase letters found in generated passwords")
	}
	if !hasLower {
		t.Error("no lowercase letters found in generated passwords")
	}
	if !hasDigit {
		t.Error("no digits found in generated passwords")
	}
}

// Test for consistent behavior
func TestPasswordGenerationConsistency(t *testing.T) {
	// Generate many passwords and ensure they all meet requirements
	for i := 0; i < 50; i++ {
		password := GenerateSecurePassword()

		if len(password) != 16 {
			t.Fatalf("password %d has wrong length: %d", i, len(password))
		}

		// Should contain only valid characters
		for _, char := range password {
			if !((char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z') ||
				(char >= '0' && char <= '9')) {
				t.Fatalf("password %d contains invalid character: %c", i, char)
			}
		}
	}
}

func TestStoreContainerPassword(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		password      string
		expectedError string
	}{
		{
			name:          "empty container name",
			containerName: "",
			password:      "testpass123",
			expectedError: "container name is required",
		},
		{
			name:          "empty password",
			containerName: "test-container",
			password:      "",
			expectedError: "password is required",
		},
		{
			name:          "valid inputs",
			containerName: "test-container",
			password:      "testpass123",
			expectedError: "exec: \"lxc\": executable file not found in $PATH", // Expected in test environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StoreContainerPassword(tt.containerName, tt.password)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestGetContainerPasswordAdditional(t *testing.T) {
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
			expectedError: "exec: \"lxc\": executable file not found in $PATH", // Expected in test environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetContainerPassword(tt.containerName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestSetUserPassword(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		username      string
		password      string
		expectedError string
	}{
		{
			name:          "empty container name",
			containerName: "",
			username:      "app",
			password:      "testpass123",
			expectedError: "container name is required",
		},
		{
			name:          "empty username",
			containerName: "test-container",
			username:      "",
			password:      "testpass123",
			expectedError: "username is required",
		},
		{
			name:          "empty password",
			containerName: "test-container",
			username:      "app",
			password:      "",
			expectedError: "password is required",
		},
		{
			name:          "valid inputs",
			containerName: "test-container",
			username:      "app",
			password:      "testpass123",
			expectedError: "exec: \"lxc\": executable file not found in $PATH", // Expected in test environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetUserPassword(tt.containerName, tt.username, tt.password)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
