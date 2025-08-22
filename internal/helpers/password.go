package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"github.com/deji/lxc-go-cli/internal/logger"
)

// GenerateSecurePassword creates a 16-character password with guaranteed character distribution
func GenerateSecurePassword() string {
	// Guarantee at least 2 of each character type for better security
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	numbers := "0123456789"
	allChars := uppercase + lowercase + numbers

	var result []byte

	// Add 2 guaranteed uppercase letters
	result = append(result, getRandomChar(uppercase), getRandomChar(uppercase))
	// Add 2 guaranteed lowercase letters
	result = append(result, getRandomChar(lowercase), getRandomChar(lowercase))
	// Add 2 guaranteed numbers
	result = append(result, getRandomChar(numbers), getRandomChar(numbers))

	// Fill remaining 10 positions randomly from all characters
	for i := 0; i < 10; i++ {
		result = append(result, getRandomChar(allChars))
	}

	// Shuffle the entire result to avoid predictable patterns
	shuffleBytes(result)

	return string(result)
}

// generateRandomString creates a random string of specified length using given character set
func generateRandomString(length int, charset string) string {
	// Handle edge cases
	if length <= 0 {
		return ""
	}
	if len(charset) == 0 {
		// Fallback to default pattern if charset is empty
		logger.Debug("Empty charset provided, using fallback")
		return "DefaultPassword123"[:length]
	}

	result := make([]byte, length)
	charsetLen := len(charset)

	// Use crypto/rand for secure random generation
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to a default pattern if crypto/rand fails (shouldn't happen)
		logger.Debug("Failed to generate secure random bytes: %v", err)
		fallback := "DefaultPassword123"
		if length <= len(fallback) {
			return fallback[:length]
		}
		return fallback
	}

	for i := 0; i < length; i++ {
		result[i] = charset[int(randomBytes[i])%charsetLen]
	}

	return string(result)
}

// StoreContainerPassword stores password in LXC metadata with base64 encoding
func StoreContainerPassword(containerName, password string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	logger.Debug("Storing password for container '%s'", containerName)

	// Encode password with base64 for basic obfuscation
	encoded := base64.StdEncoding.EncodeToString([]byte(password))

	// Store in LXC metadata using user.app-password key
	cmd := exec.Command("lxc", "config", "set", containerName, "user.app-password", encoded)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Failed to store password: %s", string(output))
		return fmt.Errorf("failed to store password in container metadata: %w (output: %s)", err, string(output))
	}

	logger.Debug("Password stored successfully in container metadata")
	return nil
}

// GetContainerPassword retrieves password from LXC metadata
func GetContainerPassword(containerName string) (string, error) {
	if containerName == "" {
		return "", fmt.Errorf("container name is required")
	}

	logger.Debug("Retrieving password for container '%s'", containerName)

	// Get password from LXC metadata
	cmd := exec.Command("lxc", "config", "get", containerName, "user.app-password")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Failed to retrieve password: %s", string(output))
		return "", fmt.Errorf("failed to retrieve password from container metadata: %w (output: %s)", err, string(output))
	}

	encoded := strings.TrimSpace(string(output))
	if encoded == "" {
		return "", fmt.Errorf("no password found for container '%s' (container may not have been created with this tool)", containerName)
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		logger.Debug("Failed to decode password: %v", err)
		return "", fmt.Errorf("failed to decode stored password: %w", err)
	}

	logger.Debug("Password retrieved successfully")
	return string(decoded), nil
}

// SetUserPassword sets the password for a user inside a container using chpasswd
func SetUserPassword(containerName, username, password string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	logger.Debug("Setting password for user '%s' in container '%s'", username, containerName)

	// Use chpasswd to set the password securely
	// Format: "username:password" | chpasswd
	passwordInput := fmt.Sprintf("%s:%s", username, password)
	cmd := exec.Command("lxc", "exec", containerName, "--", "bash", "-c", fmt.Sprintf("echo '%s' | chpasswd", passwordInput))

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Failed to set user password: %s", string(output))
		return fmt.Errorf("failed to set password for user '%s': %w (output: %s)", username, err, string(output))
	}

	logger.Debug("Password set successfully for user '%s'", username)
	return nil
}

// getRandomChar securely selects a random character from the given charset
func getRandomChar(charset string) byte {
	if len(charset) == 0 {
		return 'A' // Fallback
	}

	randomBytes := make([]byte, 1)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to first character if crypto/rand fails
		return charset[0]
	}

	return charset[int(randomBytes[0])%len(charset)]
}

// shuffleBytes performs Fisher-Yates shuffle on byte slice using crypto/rand
func shuffleBytes(slice []byte) {
	for i := len(slice) - 1; i > 0; i-- {
		randomBytes := make([]byte, 1)
		_, err := rand.Read(randomBytes)
		if err != nil {
			// Skip shuffle if crypto/rand fails (not critical)
			return
		}

		j := int(randomBytes[0]) % (i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// ContainerHasPassword checks if a container has a stored password
func ContainerHasPassword(containerName string) bool {
	_, err := GetContainerPassword(containerName)
	return err == nil
}

// FormatPasswordDisplay formats password information for display to user
func FormatPasswordDisplay(containerName, password string) string {
	return fmt.Sprintf("Password for 'app' user in '%s': %s\n", containerName, password)
}
