package helpers

import (
	"strings"
	"testing"
)

func TestGPUStatus_IsEnabled(t *testing.T) {
	tests := []struct {
		name           string
		hasGPUDevice   bool
		privilegedMode bool
		expectedResult bool
	}{
		{
			name:           "both GPU device and privileged mode enabled",
			hasGPUDevice:   true,
			privilegedMode: true,
			expectedResult: true,
		},
		{
			name:           "GPU device present but privileged mode disabled",
			hasGPUDevice:   true,
			privilegedMode: false,
			expectedResult: false,
		},
		{
			name:           "privileged mode enabled but no GPU device",
			hasGPUDevice:   false,
			privilegedMode: true,
			expectedResult: false,
		},
		{
			name:           "neither GPU device nor privileged mode enabled",
			hasGPUDevice:   false,
			privilegedMode: false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &GPUStatus{
				HasGPUDevice:   tt.hasGPUDevice,
				PrivilegedMode: tt.privilegedMode,
			}

			result := status.IsEnabled()
			if result != tt.expectedResult {
				t.Errorf("expected IsEnabled() to return %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestParseGPUStatus(t *testing.T) {
	tests := []struct {
		name               string
		yamlOutput         string
		expectedGPUDevice  bool
		expectedPrivileged bool
		expectedError      string
	}{
		{
			name: "GPU enabled configuration",
			yamlOutput: `
architecture: x86_64
config:
  security.privileged: "true"
  security.nesting: "true"
devices:
  gpu:
    type: gpu
  root:
    path: /
    type: disk
`,
			expectedGPUDevice:  true,
			expectedPrivileged: true,
		},
		{
			name: "GPU disabled configuration",
			yamlOutput: `
architecture: x86_64
config:
  security.nesting: "true"
devices:
  root:
    path: /
    type: disk
`,
			expectedGPUDevice:  false,
			expectedPrivileged: false,
		},
		{
			name: "GPU device present but not privileged",
			yamlOutput: `
architecture: x86_64
config:
  security.nesting: "true"
devices:
  gpu:
    type: gpu
  root:
    path: /
    type: disk
`,
			expectedGPUDevice:  true,
			expectedPrivileged: false,
		},
		{
			name: "privileged mode but no GPU device",
			yamlOutput: `
architecture: x86_64
config:
  security.privileged: "true"
  security.nesting: "true"
devices:
  root:
    path: /
    type: disk
`,
			expectedGPUDevice:  false,
			expectedPrivileged: true,
		},
		{
			name: "GPU device with wrong type",
			yamlOutput: `
architecture: x86_64
config:
  security.privileged: "true"
devices:
  gpu:
    type: disk
  root:
    path: /
    type: disk
`,
			expectedGPUDevice:  false,
			expectedPrivileged: true,
		},
		{
			name: "privileged mode set to false",
			yamlOutput: `
architecture: x86_64
config:
  security.privileged: "false"
devices:
  gpu:
    type: gpu
`,
			expectedGPUDevice:  true,
			expectedPrivileged: false,
		},
		{
			name:          "invalid YAML",
			yamlOutput:    `invalid: yaml: content`,
			expectedError: "failed to parse container config YAML",
		},
		{
			name: "empty config",
			yamlOutput: `
architecture: x86_64
`,
			expectedGPUDevice:  false,
			expectedPrivileged: false,
		},
		{
			name: "default LXC output format (no YAML)",
			yamlOutput: `architecture: x86_64
config:
  security.privileged: "true"
  security.nesting: "true"
devices:
  gpu:
    type: gpu
`,
			expectedGPUDevice:  true,
			expectedPrivileged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := parseGPUStatus(tt.yamlOutput)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if status.HasGPUDevice != tt.expectedGPUDevice {
				t.Errorf("expected HasGPUDevice to be %v, got %v", tt.expectedGPUDevice, status.HasGPUDevice)
			}

			if status.PrivilegedMode != tt.expectedPrivileged {
				t.Errorf("expected PrivilegedMode to be %v, got %v", tt.expectedPrivileged, status.PrivilegedMode)
			}
		})
	}
}

func TestFormatGPUStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         *GPUStatus
		expectedOutput []string // strings that should be present in output
	}{
		{
			name: "GPU fully enabled",
			status: &GPUStatus{
				HasGPUDevice:   true,
				PrivilegedMode: true,
			},
			expectedOutput: []string{
				"GPU Configuration:",
				"GPU Device: present",
				"Privileged Mode: enabled",
				"GPU Status: enabled",
			},
		},
		{
			name: "GPU fully disabled",
			status: &GPUStatus{
				HasGPUDevice:   false,
				PrivilegedMode: false,
			},
			expectedOutput: []string{
				"GPU Configuration:",
				"GPU Device: absent",
				"Privileged Mode: disabled",
				"GPU Status: disabled",
			},
		},
		{
			name: "GPU device present but not privileged",
			status: &GPUStatus{
				HasGPUDevice:   true,
				PrivilegedMode: false,
			},
			expectedOutput: []string{
				"GPU Configuration:",
				"GPU Device: present",
				"Privileged Mode: disabled",
				"GPU Status: disabled",
			},
		},
		{
			name: "privileged mode but no GPU device",
			status: &GPUStatus{
				HasGPUDevice:   false,
				PrivilegedMode: true,
			},
			expectedOutput: []string{
				"GPU Configuration:",
				"GPU Device: absent",
				"Privileged Mode: enabled",
				"GPU Status: disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatGPUStatus(tt.status)

			// Check that all expected strings are present
			for _, expected := range tt.expectedOutput {
				if !contains(output, expected) {
					t.Errorf("expected output to contain '%s', got:\n%s", expected, output)
				}
			}

			// Check output structure (should end with newline)
			if !strings.HasSuffix(output, "\n") {
				t.Error("output should end with a newline")
			}
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) != 4 {
				t.Errorf("expected 4 lines in output, got %d", len(lines))
			}
		})
	}
}

// Helper function for string contains check (similar to cmd tests)
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
