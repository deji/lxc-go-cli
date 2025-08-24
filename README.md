# LXC Go CLI

A command-line tool for creating and managing LXC containers optimized for Docker development.

## Features

- **Container Creation**: Create LXC containers with Docker and Docker Compose V2 from Docker's official repository
- **Btrfs Storage**: Automatic Btrfs storage pool management for optimal performance
- **Interactive Shell**: Execute commands in containers as the `app` user
- **Port Forwarding**: Configure TCP/UDP port forwarding between host and container
- **Security Ready**: Pre-configured security settings for Docker-in-LXC

## Installation

```bash
# Clone the repository
git clone https://github.com/deji/lxc-go-cli.git
cd lxc-go-cli

# Build the binary
make build
```

## Quick Start

```bash
# Create a new container based on Ubuntu 24.04 LTS
./lxc-go-cli create --name mycontainer

# Execute shell in container
./lxc-go-cli exec mycontainer

# Add port forwarding
./lxc-go-cli port add mycontainer 8080 80 tcp
```

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create LXC container with Docker and Compose V2 support |
| `exec` | Execute interactive shell as app user |
| `port add` | Add port forwarding rules for containers |
| `port list` | List existing port forwarding rules |
| `gpu` | Configure GPU access for containers (enable/disable/status) |
| `password` | Retrieve stored 'app' user password for container |
| `version` | Display version information |
| `completion` | Generate shell autocompletion scripts |

## Usage Examples

### Create Container
```bash
# Basic container
lxc-go-cli create --name dev-container

# Custom image and storage
lxc-go-cli create --name web-server --image ubuntu:22.04 --size 20G
```

### Port Forwarding
```bash
# Add TCP port forwarding (default protocol)
lxc-go-cli port add web-server 8080 80

# Add UDP port forwarding
lxc-go-cli port add db-server 5432 5432 udp

# Add both TCP and UDP protocols
lxc-go-cli port add app-server 3000 3000 both

# List existing port mappings
lxc-go-cli port list web-server

# Force port mapping (even if port appears in use)
lxc-go-cli port add web-server 8080 80 --force
```

### GPU Access
```bash
# Enable GPU access for container
lxc-go-cli gpu dev-container enable

# Check GPU status
lxc-go-cli gpu dev-container status

# Disable GPU access
lxc-go-cli gpu dev-container disable
```

### Password Management
```bash
# Retrieve app user password for container
lxc-go-cli password mycontainer
```

### Version Information
```bash
# Show version
lxc-go-cli version

# Show detailed version information
lxc-go-cli version --detailed
```

### Debugging
```bash
# Enable detailed logging
lxc-go-cli --log-level debug create --name test-container
```

## Development

### Testing
```bash
# Unit tests (no LXC required)
make test

# Integration tests with mocks
make test-integration

# With real LXC (requires LXC installation)
LXC_REAL=1 make test-integration

# Coverage report
make coverage-detailed
```

### Build
```bash
# Standard build
make build

# Optimized build
./build.sh
```

## Requirements

- **Runtime**: LXC, Btrfs support
- **Development**: Go 1.23+, Make

## Architecture

- **Commands** (`cmd/`): CLI interface using Cobra
- **Helpers** (`internal/helpers/`): LXC operations and business logic
- **Logger** (`internal/logger/`): Structured logging with configurable levels
- **Testing**: Comprehensive mocks for CI/CD without LXC dependencies

## License

Licensed under the terms specified in the LICENSE file.
