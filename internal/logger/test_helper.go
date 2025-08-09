package logger

import (
	"bytes"
	"log"
)

// TestHelper provides utilities for testing with controlled logging
type TestHelper struct {
	originalLevel  LogLevel
	originalLogger *log.Logger
	buffer         *bytes.Buffer
}

// NewTestHelper creates a new test helper that captures log output
func NewTestHelper() *TestHelper {
	th := &TestHelper{
		originalLevel:  GetLevel(),
		originalLogger: globalLogger.logger,
		buffer:         &bytes.Buffer{},
	}

	// Set logger to capture output in buffer
	globalLogger.logger = log.New(th.buffer, "", 0)

	return th
}

// SetLevel sets the logging level for the test
func (th *TestHelper) SetLevel(level LogLevel) {
	SetLevel(level)
}

// SetQuiet sets the logging level to ERROR to minimize test output
func (th *TestHelper) SetQuiet() {
	SetLevel(ERROR)
}

// SetVerbose sets the logging level to DEBUG for verbose test output
func (th *TestHelper) SetVerbose() {
	SetLevel(DEBUG)
}

// GetOutput returns all captured log output
func (th *TestHelper) GetOutput() string {
	return th.buffer.String()
}

// GetOutputLines returns captured log output split by lines
func (th *TestHelper) GetOutputLines() []string {
	output := th.GetOutput()
	if output == "" {
		return []string{}
	}
	lines := []string{}
	for _, line := range bytes.Split([]byte(output), []byte("\n")) {
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	return lines
}

// ClearOutput clears the captured log output
func (th *TestHelper) ClearOutput() {
	th.buffer.Reset()
}

// Cleanup restores the original logger and level
// Should be called with defer in test functions
func (th *TestHelper) Cleanup() {
	SetLevel(th.originalLevel)
	globalLogger.logger = th.originalLogger
}

// QuietTests sets up quiet logging for tests that don't need log output
// Returns a cleanup function that should be called with defer
func QuietTests() func() {
	originalLevel := GetLevel()
	SetLevel(ERROR)
	return func() {
		SetLevel(originalLevel)
	}
}

// CaptureLogsFor runs a function while capturing all log output
// Returns the captured output
func CaptureLogsFor(level LogLevel, fn func()) string {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(level)
	fn()

	return th.GetOutput()
}

// AssertContainsLog checks if the captured output contains the expected log level and message
func (th *TestHelper) AssertContainsLog(t interface {
	Errorf(format string, args ...interface{})
}, level LogLevel, message string) {
	output := th.GetOutput()
	expectedPrefix := "[" + level.String() + "]"

	if !bytes.Contains([]byte(output), []byte(expectedPrefix)) {
		t.Errorf("Expected log output to contain level %s, got: %s", level.String(), output)
		return
	}

	if !bytes.Contains([]byte(output), []byte(message)) {
		t.Errorf("Expected log output to contain message %q, got: %s", message, output)
	}
}

// AssertDoesNotContainLog checks if the captured output does NOT contain the expected log level and message
func (th *TestHelper) AssertDoesNotContainLog(t interface {
	Errorf(format string, args ...interface{})
}, level LogLevel, message string) {
	output := th.GetOutput()
	expectedPrefix := "[" + level.String() + "]"

	if bytes.Contains([]byte(output), []byte(expectedPrefix)) && bytes.Contains([]byte(output), []byte(message)) {
		t.Errorf("Expected log output to NOT contain level %s with message %q, but it did: %s",
			level.String(), message, output)
	}
}
