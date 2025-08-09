package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"DEBUG", DEBUG},
		{"Debug", DEBUG},
		{"info", INFO},
		{"INFO", INFO},
		{"Info", INFO},
		{"warn", WARN},
		{"WARN", WARN},
		{"warning", WARN},
		{"WARNING", WARN},
		{"error", ERROR},
		{"ERROR", ERROR},
		{"Error", ERROR},
		{"invalid", INFO}, // Should default to INFO
		{"", INFO},        // Should default to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLogLevel(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"}, // Invalid level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel(%d).String() = %q, expected %q", tt.level, result, tt.expected)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Test setting different levels
	levels := []LogLevel{DEBUG, INFO, WARN, ERROR}
	for _, level := range levels {
		SetLevel(level)
		if GetLevel() != level {
			t.Errorf("After SetLevel(%v), GetLevel() = %v", level, GetLevel())
		}
	}
}

func TestSetLevelFromString(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Test setting levels from strings
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DEBUG},
		{"info", INFO},
		{"warn", WARN},
		{"error", ERROR},
		{"invalid", INFO}, // Should default to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			SetLevelFromString(tt.input)
			if GetLevel() != tt.expected {
				t.Errorf("After SetLevelFromString(%q), GetLevel() = %v, expected %v",
					tt.input, GetLevel(), tt.expected)
			}
		})
	}
}

func TestLoggerShouldLog(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	tests := []struct {
		setLevel  LogLevel
		testLevel LogLevel
		shouldLog bool
	}{
		{DEBUG, DEBUG, true},
		{DEBUG, INFO, true},
		{DEBUG, WARN, true},
		{DEBUG, ERROR, true},
		{INFO, DEBUG, false},
		{INFO, INFO, true},
		{INFO, WARN, true},
		{INFO, ERROR, true},
		{WARN, DEBUG, false},
		{WARN, INFO, false},
		{WARN, WARN, true},
		{WARN, ERROR, true},
		{ERROR, DEBUG, false},
		{ERROR, INFO, false},
		{ERROR, WARN, false},
		{ERROR, ERROR, true},
	}

	for _, tt := range tests {
		t.Run(tt.setLevel.String()+"_"+tt.testLevel.String(), func(t *testing.T) {
			SetLevel(tt.setLevel)
			result := globalLogger.shouldLog(tt.testLevel)
			if result != tt.shouldLog {
				t.Errorf("With level %v, shouldLog(%v) = %v, expected %v",
					tt.setLevel, tt.testLevel, result, tt.shouldLog)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Capture logger output
	var buf bytes.Buffer
	originalOutput := globalLogger.logger.Writer()
	globalLogger.logger = log.New(&buf, "", 0)
	defer func() {
		globalLogger.logger = log.New(originalOutput, "", 0)
	}()

	// Test that messages are logged at appropriate levels
	SetLevel(INFO)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines (no debug message)
	if len(lines) != 3 {
		t.Errorf("Expected 3 log lines, got %d: %v", len(lines), lines)
	}

	// Check that each line has the correct prefix
	expectedPrefixes := []string{"[INFO]", "[WARN]", "[ERROR]"}
	for i, expectedPrefix := range expectedPrefixes {
		if i < len(lines) && !strings.HasPrefix(lines[i], expectedPrefix) {
			t.Errorf("Line %d should start with %q, got %q", i, expectedPrefix, lines[i])
		}
	}

	// Check message content
	if len(lines) >= 1 && !strings.Contains(lines[0], "info message") {
		t.Error("INFO line should contain 'info message'")
	}
	if len(lines) >= 2 && !strings.Contains(lines[1], "warn message") {
		t.Error("WARN line should contain 'warn message'")
	}
	if len(lines) >= 3 && !strings.Contains(lines[2], "error message") {
		t.Error("ERROR line should contain 'error message'")
	}
}

func TestLoggerOutputWithDebug(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Capture logger output
	var buf bytes.Buffer
	originalOutput := globalLogger.logger.Writer()
	globalLogger.logger = log.New(&buf, "", 0)
	defer func() {
		globalLogger.logger = log.New(originalOutput, "", 0)
	}()

	// Test that debug messages are logged when level is DEBUG
	SetLevel(DEBUG)

	Debug("debug message")
	Info("info message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 2 lines (including debug message)
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d: %v", len(lines), lines)
	}

	// Check that debug line is present
	if len(lines) >= 1 && !strings.HasPrefix(lines[0], "[DEBUG]") {
		t.Errorf("First line should start with [DEBUG], got %q", lines[0])
	}
	if len(lines) >= 1 && !strings.Contains(lines[0], "debug message") {
		t.Error("DEBUG line should contain 'debug message'")
	}
}

func TestIsDebugEnabled(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Test with different levels
	SetLevel(DEBUG)
	if !IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return true when level is DEBUG")
	}

	SetLevel(INFO)
	if IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return false when level is INFO")
	}

	SetLevel(ERROR)
	if IsDebugEnabled() {
		t.Error("IsDebugEnabled() should return false when level is ERROR")
	}
}

func TestIsLevelEnabled(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Test with INFO level set
	SetLevel(INFO)

	if IsLevelEnabled(DEBUG) {
		t.Error("DEBUG should not be enabled when level is INFO")
	}
	if !IsLevelEnabled(INFO) {
		t.Error("INFO should be enabled when level is INFO")
	}
	if !IsLevelEnabled(WARN) {
		t.Error("WARN should be enabled when level is INFO")
	}
	if !IsLevelEnabled(ERROR) {
		t.Error("ERROR should be enabled when level is INFO")
	}
}

func TestPrintDebug(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Capture logger output
	var buf bytes.Buffer
	originalOutput := globalLogger.logger.Writer()
	globalLogger.logger = log.New(&buf, "", 0)
	defer func() {
		globalLogger.logger = log.New(originalOutput, "", 0)
	}()

	// Test PrintDebug with DEBUG level
	SetLevel(DEBUG)
	PrintDebug("test debug message")

	output := buf.String()
	if !strings.Contains(output, "test debug message") {
		t.Error("PrintDebug should output message when debug is enabled")
	}

	// Clear buffer and test with INFO level
	buf.Reset()
	SetLevel(INFO)
	PrintDebug("test debug message 2")

	output = buf.String()
	if strings.Contains(output, "test debug message 2") {
		t.Error("PrintDebug should not output message when debug is disabled")
	}
}

func TestPrintInfo(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	// Capture logger output
	var buf bytes.Buffer
	originalOutput := globalLogger.logger.Writer()
	globalLogger.logger = log.New(&buf, "", 0)
	defer func() {
		globalLogger.logger = log.New(originalOutput, "", 0)
	}()

	// Test PrintInfo with INFO level
	SetLevel(INFO)
	PrintInfo("test info message")

	output := buf.String()
	if !strings.Contains(output, "test info message") {
		t.Error("PrintInfo should output message when info is enabled")
	}

	// Clear buffer and test with ERROR level
	buf.Reset()
	SetLevel(ERROR)
	PrintInfo("test info message 2")

	output = buf.String()
	if strings.Contains(output, "test info message 2") {
		t.Error("PrintInfo should not output message when info is disabled")
	}
}

// Test that the default level is INFO
func TestDefaultLevel(t *testing.T) {
	// Create a new logger to test default initialization
	// We can't test the global logger's initial state easily since it might be modified
	// by other tests, but we can test that ParseLogLevel defaults to INFO
	defaultLevel := ParseLogLevel("")
	if defaultLevel != INFO {
		t.Errorf("Default log level should be INFO, got %v", defaultLevel)
	}
}
