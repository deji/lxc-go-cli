package helpers

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestIsPortAvailable(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		protocol string
		expected bool
	}{
		{
			name:     "invalid port too low",
			port:     0,
			protocol: "tcp",
			expected: false,
		},
		{
			name:     "invalid port too high",
			port:     65536,
			protocol: "tcp",
			expected: false,
		},
		{
			name:     "unknown protocol",
			port:     8080,
			protocol: "unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPortAvailable(tt.port, tt.protocol)
			if result != tt.expected {
				t.Errorf("IsPortAvailable(%d, %s) = %v, expected %v", tt.port, tt.protocol, result, tt.expected)
			}
		})
	}
}

func TestIsPortAvailable_TCPPortInUse(t *testing.T) {
	// Start a TCP listener on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Get the assigned port
	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Test that the port shows as unavailable
	available := IsPortAvailable(port, "tcp")
	if available {
		t.Errorf("Port %d should be unavailable (TCP listener active)", port)
	}

	// Test the same port with UDP should be available
	available = IsPortAvailable(port, "udp")
	if !available {
		t.Errorf("Port %d should be available for UDP (only TCP listener active)", port)
	}
}

func TestIsPortAvailable_UDPPortInUse(t *testing.T) {
	// Start a UDP listener on a random port
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Fatalf("Failed to create UDP listener: %v", err)
	}
	defer conn.Close()

	// Get the assigned port
	addr := conn.LocalAddr().(*net.UDPAddr)
	port := addr.Port

	// Test that the port shows as unavailable
	available := IsPortAvailable(port, "udp")
	if available {
		t.Errorf("Port %d should be unavailable (UDP listener active)", port)
	}

	// Test the same port with TCP should be available
	available = IsPortAvailable(port, "tcp")
	if !available {
		t.Errorf("Port %d should be available for TCP (only UDP listener active)", port)
	}
}

func TestIsPortAvailable_AvailablePorts(t *testing.T) {
	// Port 0 is a special case - our function rejects it as invalid
	// Test with a high port that should be available
	testPort := 65432 // Very high port, unlikely to be in use
	
	// Test TCP on a presumably available port
	available := IsPortAvailable(testPort, "tcp")
	if !available {
		t.Logf("Port %d not available for TCP (may be in use in test environment)", testPort)
	}

	// Test UDP on a presumably available port
	available = IsPortAvailable(testPort, "udp")
	if !available {
		t.Logf("Port %d not available for UDP (may be in use in test environment)", testPort)
	}
}

func TestValidatePortMapping(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		protocol    string
		timeout     time.Duration
		setupServer bool
		expectError bool
	}{
		{
			name:        "invalid protocol",
			port:        8080,
			protocol:    "unknown",
			timeout:     1 * time.Second,
			setupServer: false,
			expectError: true,
		},
		{
			name:        "zero timeout uses default",
			port:        8080,
			protocol:    "tcp",
			timeout:     0,
			setupServer: false,
			expectError: true, // No server listening
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortMapping(tt.port, tt.protocol, tt.timeout)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidatePortMapping_WithServer(t *testing.T) {
	// Start a TCP server
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Accept connections in the background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed
			}
			conn.Close()
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Test validation against the server
	err = ValidatePortMapping(port, "tcp", 2*time.Second)
	if err != nil {
		t.Errorf("Validation should succeed with running server: %v", err)
	}
}

func TestFormatPortConflictError(t *testing.T) {
	err := FormatPortConflictError("8080", "tcp")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	errStr := err.Error()
	expectedParts := []string{
		"host port 8080 (tcp) is already in use",
		"Suggestions:",
		"Use a different host port:",
		"Check what's using the port:",
		"Force creation anyway:",
		"--force",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errStr, part) {
			t.Errorf("Error message should contain '%s', got: %s", part, errStr)
		}
	}
}

func TestGetPortUsageInfo(t *testing.T) {
	info := GetPortUsageInfo(8080)
	
	expectedParts := []string{
		"Port 8080",
		"may be in use",
		"ss -tuln",
		"netstat -tuln",
	}

	for _, part := range expectedParts {
		if !strings.Contains(info, part) {
			t.Errorf("Port usage info should contain '%s', got: %s", part, info)
		}
	}
}

func TestPortAvailability_Integration(t *testing.T) {
	// Test the full workflow of checking availability and then using a port
	
	// 1. Check that a high port is available
	testPort := 34567 // Unlikely to be in use
	if !IsPortAvailable(testPort, "tcp") {
		t.Skip("Test port is not available, skipping integration test")
	}

	// 2. Start a server on that port
	listener, err := net.Listen("tcp", ":34567")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	// 3. Check that the port now shows as unavailable
	if IsPortAvailable(testPort, "tcp") {
		t.Error("Port should now be unavailable after starting server")
	}

	// 4. Test that validation works
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	err = ValidatePortMapping(testPort, "tcp", 1*time.Second)
	if err != nil {
		t.Errorf("Validation should work with running server: %v", err)
	}
}

func TestProtocolCaseInsensitive(t *testing.T) {
	tests := []struct {
		protocol string
		port     int
	}{
		{"TCP", 65433},
		{"tcp", 65434},
		{"UDP", 65435},
		{"udp", 65436},
		{"Tcp", 65437},
		{"Udp", 65438},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			// Should not panic or fail for case variations
			result := IsPortAvailable(tt.port, tt.protocol)
			// High ports should generally be available, but we don't fail if they're not
			// since it depends on the test environment
			t.Logf("Port %d protocol %s availability: %v", tt.port, tt.protocol, result)
		})
	}
}
