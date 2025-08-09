package logger

import (
	"strings"
	"testing"
)

func TestNewTestHelper(t *testing.T) {
	th := NewTestHelper()
	if th == nil {
		t.Fatal("NewTestHelper should not return nil")
	}

	// Should start with empty output
	output := th.GetOutput()
	if output != "" {
		t.Errorf("New test helper should have empty output, got: %q", output)
	}

	// Should have captured original state
	if th.originalLevel < 0 {
		t.Error("Original level should be captured")
	}

	if th.originalLogger == nil {
		t.Error("Original logger should be captured")
	}

	if th.buffer == nil {
		t.Error("Buffer should be initialized")
	}

	// Clean up
	th.Cleanup()
}

func TestTestHelper_SetLevel(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	// Test setting different levels
	levels := []LogLevel{DEBUG, INFO, WARN, ERROR}
	for _, level := range levels {
		th.SetLevel(level)
		if GetLevel() != level {
			t.Errorf("After th.SetLevel(%v), GetLevel() = %v", level, GetLevel())
		}
	}
}

func TestTestHelper_SetQuiet(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetQuiet()
	if GetLevel() != ERROR {
		t.Errorf("SetQuiet should set level to ERROR, got %v", GetLevel())
	}
}

func TestTestHelper_SetVerbose(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetVerbose()
	if GetLevel() != DEBUG {
		t.Errorf("SetVerbose should set level to DEBUG, got %v", GetLevel())
	}
}

func TestTestHelper_GetOutput(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(INFO)

	// Initially empty
	output := th.GetOutput()
	if output != "" {
		t.Errorf("Initial output should be empty, got: %q", output)
	}

	// Log something and check output
	Info("test message")
	output = th.GetOutput()
	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain 'test message', got: %q", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Output should contain '[INFO]', got: %q", output)
	}
}

func TestTestHelper_GetOutputLines(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(INFO)

	// Initially empty
	lines := th.GetOutputLines()
	if len(lines) != 0 {
		t.Errorf("Initial lines should be empty, got: %v", lines)
	}

	// Log multiple messages
	Info("first message")
	Warn("second message")

	lines = th.GetOutputLines()
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d: %v", len(lines), lines)
	}

	if !strings.Contains(lines[0], "first message") {
		t.Errorf("First line should contain 'first message', got: %q", lines[0])
	}

	if !strings.Contains(lines[1], "second message") {
		t.Errorf("Second line should contain 'second message', got: %q", lines[1])
	}
}

func TestTestHelper_ClearOutput(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(INFO)

	// Log something
	Info("test message")
	output := th.GetOutput()
	if output == "" {
		t.Error("Should have output before clearing")
	}

	// Clear and check
	th.ClearOutput()
	output = th.GetOutput()
	if output != "" {
		t.Errorf("Output should be empty after clearing, got: %q", output)
	}
}

func TestTestHelper_Cleanup(t *testing.T) {
	originalLevel := GetLevel()

	th := NewTestHelper()

	// Change level
	th.SetLevel(DEBUG)
	if GetLevel() == originalLevel {
		t.Error("Level should be changed before cleanup")
	}

	// Cleanup should restore original level
	th.Cleanup()
	if GetLevel() != originalLevel {
		t.Errorf("After cleanup, level should be %v, got %v", originalLevel, GetLevel())
	}
}

func TestQuietTests(t *testing.T) {
	originalLevel := GetLevel()

	cleanup := QuietTests()
	if GetLevel() != ERROR {
		t.Errorf("QuietTests should set level to ERROR, got %v", GetLevel())
	}

	cleanup()
	if GetLevel() != originalLevel {
		t.Errorf("After cleanup, level should be %v, got %v", originalLevel, GetLevel())
	}
}

func TestCaptureLogsFor(t *testing.T) {
	tests := []struct {
		level            LogLevel
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			level:         DEBUG,
			shouldContain: []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"},
		},
		{
			level:            INFO,
			shouldContain:    []string{"[INFO]", "[WARN]", "[ERROR]"},
			shouldNotContain: []string{"[DEBUG]"},
		},
		{
			level:            WARN,
			shouldContain:    []string{"[WARN]", "[ERROR]"},
			shouldNotContain: []string{"[DEBUG]", "[INFO]"},
		},
		{
			level:            ERROR,
			shouldContain:    []string{"[ERROR]"},
			shouldNotContain: []string{"[DEBUG]", "[INFO]", "[WARN]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			output := CaptureLogsFor(tt.level, func() {
				Debug("debug message")
				Info("info message")
				Warn("warn message")
				Error("error message")
			})

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got: %s", expected, output)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf("Output should NOT contain %q, got: %s", notExpected, output)
				}
			}
		})
	}
}

func TestTestHelper_AssertContainsLog(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(INFO)

	// Log a message
	Info("test message")

	// Create a mock testing.T
	mockT := &mockTestingT{}

	// Should pass - message exists
	th.AssertContainsLog(mockT, INFO, "test message")
	if mockT.errorCalled {
		t.Error("AssertContainsLog should not call Errorf when message exists")
	}

	// Should fail - level doesn't exist
	mockT.reset()
	th.AssertContainsLog(mockT, DEBUG, "test message")
	if !mockT.errorCalled {
		t.Error("AssertContainsLog should call Errorf when level doesn't exist")
	}

	// Should fail - message doesn't exist
	mockT.reset()
	th.AssertContainsLog(mockT, INFO, "nonexistent message")
	if !mockT.errorCalled {
		t.Error("AssertContainsLog should call Errorf when message doesn't exist")
	}
}

func TestTestHelper_AssertDoesNotContainLog(t *testing.T) {
	th := NewTestHelper()
	defer th.Cleanup()

	th.SetLevel(INFO)

	// Log a message
	Info("test message")

	// Create a mock testing.T
	mockT := &mockTestingT{}

	// Should pass - DEBUG level doesn't exist
	th.AssertDoesNotContainLog(mockT, DEBUG, "test message")
	if mockT.errorCalled {
		t.Error("AssertDoesNotContainLog should not call Errorf when level doesn't exist")
	}

	// Should pass - message doesn't exist
	mockT.reset()
	th.AssertDoesNotContainLog(mockT, INFO, "nonexistent message")
	if mockT.errorCalled {
		t.Error("AssertDoesNotContainLog should not call Errorf when message doesn't exist")
	}

	// Should fail - both level and message exist
	mockT.reset()
	th.AssertDoesNotContainLog(mockT, INFO, "test message")
	if !mockT.errorCalled {
		t.Error("AssertDoesNotContainLog should call Errorf when both level and message exist")
	}
}

func TestTestHelper_Integration(t *testing.T) {
	// Test a complete workflow
	th := NewTestHelper()
	defer th.Cleanup()

	// Start with verbose logging
	th.SetVerbose()

	Debug("debug message")
	Info("info message")

	lines := th.GetOutputLines()
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines with verbose logging, got %d", len(lines))
	}

	// Switch to quiet and clear
	th.SetQuiet()
	th.ClearOutput()

	Debug("should not appear")
	Info("should not appear")
	Warn("should not appear")
	Error("should appear")

	lines = th.GetOutputLines()
	if len(lines) != 1 {
		t.Errorf("Expected 1 line with quiet logging, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "should appear") {
		t.Errorf("Error message should appear, got: %v", lines)
	}
}

// mockTestingT implements the interface required by AssertContainsLog and AssertDoesNotContainLog
type mockTestingT struct {
	errorCalled bool
	lastError   string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
	m.lastError = format
}

func (m *mockTestingT) reset() {
	m.errorCalled = false
	m.lastError = ""
}
