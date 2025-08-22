/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deji/lxc-go-cli/internal/helpers"
	"github.com/deji/lxc-go-cli/internal/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	portTimeout time.Duration
	forcePort   bool
)

// portCmd represents the port command
var portCmd = &cobra.Command{
	Use:   "port <add|list>",
	Short: "Manage port forwarding for LXC containers",
	Long: `Manage port forwarding between host and container using LXC proxy devices.

Available subcommands:
  add   - Add port forwarding rule
  list  - List existing port forwarding rules

Examples:
  lxc-go-cli port add mycontainer 8080 80        # Add TCP port forwarding
  lxc-go-cli port add mycontainer 5432 5432 udp  # Add UDP port forwarding
  lxc-go-cli port list mycontainer               # List all port mappings`,
}

// portAddCmd represents the port add subcommand
var portAddCmd = &cobra.Command{
	Use:   "add <container-name> <host-port> <container-port> [tcp|udp|both]",
	Short: "Add port forwarding rule for an LXC container",
	Long: `Add port forwarding from host to container using LXC proxy devices.
This command creates a proxy device that forwards traffic from the host port
to the container port using the specified protocol.

The protocol parameter is optional and defaults to 'tcp'.
When 'both' is specified, both TCP and UDP forwarding rules are created.

Examples:
  lxc-go-cli port add mycontainer 8080 80        # defaults to tcp
  lxc-go-cli port add mycontainer 8080 80 tcp    # explicit tcp
  lxc-go-cli port add mycontainer 5432 5432 udp  # udp only
  lxc-go-cli port add mycontainer 3000 3000 both # both tcp and udp`,
	Args: cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]
		hostPort := args[1]
		containerPort := args[2]

		// Protocol is optional, defaults to "tcp"
		protocol := "tcp"
		if len(args) > 3 {
			protocol = args[3]
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), portTimeout)
		defer cancel()

		manager := &DefaultContainerPortManager{}
		return configurePortForwarding(ctx, manager, containerName, hostPort, containerPort, protocol, forcePort)
	},
}

// portListCmd represents the port list subcommand
var portListCmd = &cobra.Command{
	Use:   "list <container-name>",
	Short: "List port forwarding rules for an LXC container",
	Long: `List all existing port forwarding rules for the specified container.
This command shows all proxy devices configured for port forwarding,
displaying the protocol, host port, container port, and device name.

Examples:
  lxc-go-cli port list mycontainer  # List all port mappings for mycontainer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), portTimeout)
		defer cancel()

		manager := &DefaultContainerPortManager{}
		return listPortForwarding(ctx, manager, containerName)
	},
}

// ContainerPortManager interface for dependency injection
type ContainerPortManager interface {
	ContainerExists(ctx context.Context, name string) bool
	RunLXCCommand(ctx context.Context, args ...string) error
	GetContainerConfig(ctx context.Context, containerName string) ([]byte, error)
}

// DefaultContainerPortManager implements ContainerPortManager using helpers
type DefaultContainerPortManager struct{}

func (d *DefaultContainerPortManager) ContainerExists(ctx context.Context, name string) bool {
	return helpers.ContainerExists(name)
}

func (d *DefaultContainerPortManager) RunLXCCommand(ctx context.Context, args ...string) error {
	// For LXC config commands, we need to run them on the host, not in a container
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Use exec.Command to run LXC commands directly on the host
	return helpers.RunHostCommand(ctx, args...)
}

func (d *DefaultContainerPortManager) GetContainerConfig(ctx context.Context, containerName string) ([]byte, error) {
	if containerName == "" {
		return nil, fmt.Errorf("container name is required")
	}

	// Execute lxc config show command using exec.CommandContext
	cmd := exec.CommandContext(ctx, "lxc", "config", "show", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Failed to get container config: %s", string(output))
		return nil, fmt.Errorf("failed to get container config: %w (output: %s)", err, string(output))
	}

	return output, nil
}

// validatePortForwardingArgs validates the arguments for port forwarding
func validatePortForwardingArgs(containerName, hostPort, containerPort, protocol string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	// Validate host port
	if hostPort == "" {
		return fmt.Errorf("host port is required")
	}
	hostPortNum, err := strconv.Atoi(hostPort)
	if err != nil {
		return fmt.Errorf("invalid host port '%s': must be a number", hostPort)
	}
	if hostPortNum < 1 || hostPortNum > 65535 {
		return fmt.Errorf("invalid host port '%s': must be between 1 and 65535", hostPort)
	}

	// Validate container port
	if containerPort == "" {
		return fmt.Errorf("container port is required")
	}
	containerPortNum, err := strconv.Atoi(containerPort)
	if err != nil {
		return fmt.Errorf("invalid container port '%s': must be a number", containerPort)
	}
	if containerPortNum < 1 || containerPortNum > 65535 {
		return fmt.Errorf("invalid container port '%s': must be between 1 and 65535", containerPort)
	}

	// Validate protocol - empty defaults to tcp
	if protocol == "" {
		protocol = "tcp"
	}
	protocol = strings.ToLower(protocol)
	if protocol != "tcp" && protocol != "udp" && protocol != "both" {
		return fmt.Errorf("invalid protocol '%s': must be 'tcp', 'udp', or 'both'", protocol)
	}

	return nil
}

// configurePortForwarding configures port forwarding for a container
func configurePortForwarding(ctx context.Context, manager ContainerPortManager, containerName, hostPort, containerPort, protocol string, force bool) error {
	// Validate arguments
	if err := validatePortForwardingArgs(containerName, hostPort, containerPort, protocol); err != nil {
		return err
	}

	// Check if container exists
	if !manager.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	// Handle empty protocol (default to tcp)
	if protocol == "" {
		protocol = "tcp"
	}
	protocol = strings.ToLower(protocol)

	// Configure port forwarding based on protocol
	switch protocol {
	case "tcp":
		return configurePortForwardingForProtocol(ctx, manager, containerName, hostPort, containerPort, "tcp", force)
	case "udp":
		return configurePortForwardingForProtocol(ctx, manager, containerName, hostPort, containerPort, "udp", force)
	case "both":
		// Configure both TCP and UDP
		if err := configurePortForwardingForProtocol(ctx, manager, containerName, hostPort, containerPort, "tcp", force); err != nil {
			return err
		}
		return configurePortForwardingForProtocol(ctx, manager, containerName, hostPort, containerPort, "udp", force)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// configurePortForwardingForProtocol configures port forwarding for a specific protocol
func configurePortForwardingForProtocol(ctx context.Context, manager ContainerPortManager, containerName, hostPort, containerPort, protocol string, force bool) error {
	// Check port availability unless forced
	if !force {
		hostPortNum, err := strconv.Atoi(hostPort)
		if err != nil {
			return fmt.Errorf("invalid host port '%s': %w", hostPort, err)
		}

		if !helpers.IsPortAvailable(hostPortNum, protocol) {
			return helpers.FormatPortConflictError(hostPort, protocol)
		}
	}

	deviceName := fmt.Sprintf("%s-%s-%s-%s", containerName, hostPort, containerPort, protocol)
	connectAddr := fmt.Sprintf("%s:0.0.0.0:%s", protocol, hostPort)
	listenAddr := fmt.Sprintf("%s:0.0.0.0:%s", protocol, containerPort)

	logger.Info("Configuring %s port forwarding: %s:%s -> %s:%s",
		strings.ToUpper(protocol), "0.0.0.0", hostPort, containerName, containerPort)

	// Use lxc config device add to create the proxy device
	err := manager.RunLXCCommand(ctx, "lxc", "config", "device", "add", containerName, deviceName, "proxy",
		fmt.Sprintf("connect=%s", connectAddr), fmt.Sprintf("listen=%s", listenAddr))
	if err != nil {
		return fmt.Errorf("failed to configure %s port forwarding %s:%s -> %s:%s: %w",
			protocol, "0.0.0.0", hostPort, containerName, containerPort, err)
	}

	logger.Info("Successfully configured %s port forwarding %s:%s -> %s:%s",
		strings.ToUpper(protocol), "0.0.0.0", hostPort, containerName, containerPort)

	return nil
}

// ContainerConfig represents the structure of lxc config show output
type ContainerConfig struct {
	Devices map[string]Device `yaml:"devices"`
}

// Device represents a device configuration in LXC
type Device struct {
	Type    string `yaml:"type"`
	Connect string `yaml:"connect,omitempty"`
	Listen  string `yaml:"listen,omitempty"`
}

// PortMapping represents a port forwarding configuration
type PortMapping struct {
	DeviceName    string
	Protocol      string
	HostPort      string
	ContainerPort string
	HostIP        string
	ContainerIP   string
}

// listPortForwarding lists all port forwarding rules for a container
func listPortForwarding(ctx context.Context, manager ContainerPortManager, containerName string) error {
	if containerName == "" {
		return fmt.Errorf("container name is required")
	}

	// Check if container exists
	if !manager.ContainerExists(ctx, containerName) {
		return fmt.Errorf("container '%s' does not exist", containerName)
	}

	// Get container configuration
	configData, err := manager.GetContainerConfig(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to get container configuration: %w", err)
	}

	// Parse port mappings from configuration
	mappings, err := parsePortMappingsFromConfig(configData, containerName)
	if err != nil {
		return fmt.Errorf("failed to parse port mappings: %w", err)
	}

	// Display results
	if len(mappings) == 0 {
		fmt.Printf("No port forwarding rules found for container '%s'\n", containerName)
		return nil
	}

	fmt.Printf("Port mappings for container '%s':\n", containerName)
	fmt.Print(formatPortMappings(mappings))
	return nil
}

// parsePortMappingsFromConfig parses YAML config data to extract port mappings
func parsePortMappingsFromConfig(yamlData []byte, containerName string) ([]PortMapping, error) {
	var config ContainerConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse container configuration: %w", err)
	}

	var mappings []PortMapping
	for deviceName, device := range config.Devices {
		// Only process proxy devices that match our naming convention
		if device.Type == "proxy" && isPortDevice(deviceName, containerName) {
			mapping, err := parsePortMapping(deviceName, device)
			if err != nil {
				logger.Debug("Failed to parse port mapping for device '%s': %v", deviceName, err)
				continue
			}
			mappings = append(mappings, *mapping)
		}
	}

	return mappings, nil
}

// isPortDevice checks if a device name matches our port forwarding naming convention
func isPortDevice(deviceName, containerName string) bool {
	// Expected pattern: {containerName}-{hostPort}-{containerPort}-{protocol}
	pattern := fmt.Sprintf(`^%s-\d+-\d+-(tcp|udp)$`, regexp.QuoteMeta(containerName))
	matched, err := regexp.MatchString(pattern, deviceName)
	if err != nil {
		logger.Debug("Failed to match device name pattern: %v", err)
		return false
	}
	return matched
}

// parsePortMapping extracts port mapping information from device configuration
func parsePortMapping(deviceName string, device Device) (*PortMapping, error) {
	// Extract protocol, host port, container port from device name
	// Expected format: {containerName}-{hostPort}-{containerPort}-{protocol}
	parts := strings.Split(deviceName, "-")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid device name format: %s", deviceName)
	}

	protocol := parts[len(parts)-1]
	containerPort := parts[len(parts)-2]
	hostPort := parts[len(parts)-3]

	// Parse connect and listen addresses
	hostIP, containerIP := "0.0.0.0", "0.0.0.0"

	// Parse connect field (format: tcp:IP:PORT or udp:IP:PORT)
	if device.Connect != "" {
		connectParts := strings.Split(device.Connect, ":")
		if len(connectParts) == 3 {
			hostIP = connectParts[1]
		}
	}

	// Parse listen field (format: tcp:IP:PORT or udp:IP:PORT)
	if device.Listen != "" {
		listenParts := strings.Split(device.Listen, ":")
		if len(listenParts) == 3 {
			containerIP = listenParts[1]
		}
	}

	return &PortMapping{
		DeviceName:    deviceName,
		Protocol:      strings.ToUpper(protocol),
		HostPort:      hostPort,
		ContainerPort: containerPort,
		HostIP:        hostIP,
		ContainerIP:   containerIP,
	}, nil
}

// formatPortMappings formats port mappings for display
func formatPortMappings(mappings []PortMapping) string {
	if len(mappings) == 0 {
		return ""
	}

	var result strings.Builder

	// Table header
	result.WriteString("PROTOCOL  HOST PORT  CONTAINER PORT  HOST IP      CONTAINER IP  DEVICE NAME\n")
	result.WriteString("--------  ---------  --------------  -----------  ------------  -----------\n")

	// Table rows
	for _, mapping := range mappings {
		result.WriteString(fmt.Sprintf("%-8s  %-9s  %-14s  %-11s  %-12s  %s\n",
			mapping.Protocol,
			mapping.HostPort,
			mapping.ContainerPort,
			mapping.HostIP,
			mapping.ContainerIP,
			mapping.DeviceName,
		))
	}

	return result.String()
}

func init() {
	rootCmd.AddCommand(portCmd)

	// Add subcommands
	portCmd.AddCommand(portAddCmd)
	portCmd.AddCommand(portListCmd)

	// Add timeout flag to both subcommands
	portAddCmd.Flags().DurationVarP(&portTimeout, "timeout", "t", 30*time.Second, "Timeout for the port configuration operation")
	portListCmd.Flags().DurationVarP(&portTimeout, "timeout", "t", 30*time.Second, "Timeout for the port configuration operation")

	// Add force flag to port add command
	portAddCmd.Flags().BoolVarP(&forcePort, "force", "f", false, "Force port mapping creation even if port appears to be in use")
}
