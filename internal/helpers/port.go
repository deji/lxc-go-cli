package helpers

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/deji/lxc-go-cli/internal/logger"
)

// IsPortAvailable checks if a port is available for use on the host
func IsPortAvailable(port int, protocol string) bool {
	if port < 1 || port > 65535 {
		return false
	}

	address := fmt.Sprintf(":%d", port)
	protocol = strings.ToLower(protocol)

	switch protocol {
	case "tcp":
		listener, err := net.Listen("tcp", address)
		if err != nil {
			logger.Debug("Port %d (TCP) appears to be in use: %v", port, err)
			return false
		}
		listener.Close()
		logger.Debug("Port %d (TCP) is available", port)
		return true
	case "udp":
		conn, err := net.ListenPacket("udp", address)
		if err != nil {
			logger.Debug("Port %d (UDP) appears to be in use: %v", port, err)
			return false
		}
		conn.Close()
		logger.Debug("Port %d (UDP) is available", port)
		return true
	default:
		logger.Debug("Unknown protocol '%s' for port availability check", protocol)
		return false
	}
}

// ValidatePortMapping attempts to validate that a port mapping is actually working
// This is a best-effort check and may have false positives/negatives
func ValidatePortMapping(hostPort int, protocol string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// Give LXC a moment to set up the proxy
	time.Sleep(100 * time.Millisecond)

	address := fmt.Sprintf("localhost:%d", hostPort)
	protocol = strings.ToLower(protocol)

	switch protocol {
	case "tcp":
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return fmt.Errorf("port %d (TCP) appears to be non-functional: %w", hostPort, err)
		}
		conn.Close()
		logger.Debug("Port %d (TCP) mapping validated successfully", hostPort)
		return nil
	case "udp":
		// UDP is connectionless, so this is a basic reachability test
		conn, err := net.Dial("udp", address)
		if err != nil {
			return fmt.Errorf("port %d (UDP) appears to be non-functional: %w", hostPort, err)
		}
		conn.Close()
		logger.Debug("Port %d (UDP) mapping validated (basic check)", hostPort)
		return nil
	default:
		return fmt.Errorf("cannot validate unknown protocol '%s'", protocol)
	}
}

// FormatPortConflictError creates a helpful error message when a port is in use
func FormatPortConflictError(hostPort, protocol string) error {
	return fmt.Errorf(`host port %s (%s) is already in use

Suggestions:
  • Use a different host port: lxc-go-cli port add <container> <other-port> <container-port> %s
  • Check what's using the port: ss -tuln | grep :%s
  • Force creation anyway: lxc-go-cli port add <container> %s <container-port> %s --force

Note: Forced creation may result in non-functional port mapping`,
		hostPort, protocol, protocol, hostPort, hostPort, protocol)
}

// GetPortUsageInfo returns information about what might be using a port
// This is a best-effort informational function
func GetPortUsageInfo(port int) string {
	// Try to get basic info about port usage
	// This is primarily for debugging/informational purposes
	return fmt.Sprintf("Port %d may be in use by another service. Use 'ss -tuln | grep :%d' or 'netstat -tuln | grep :%d' to investigate.",
		port, port, port)
}
