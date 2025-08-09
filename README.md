# LXC Go CLI

A command-line tool for creating and managing LXC containers optimized for Docker development.

## Features

- **Container Creation**: Create LXC containers with Docker and Docker Compose pre-installed
- **Btrfs Storage**: Automatic Btrfs storage pool management for optimal performance
- **Interactive Shell**: Execute commands in containers as the `app` user
- **Port Forwarding**: Configure TCP/UDP port forwarding between host and container
- **Security Ready**: Pre-configured security settings for Docker-in-LXC

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/lxc-go-cli.git
cd lxc-go-cli

# Build the binary
make build
# or
go build -o lxc-go-cli .
```

## Quick Start

```bash
# Create a new container
./lxc-go-cli create --name mycontainer --image ubuntu:24.04

# Execute shell in container
./lxc-go-cli exec mycontainer

# Configure port forwarding
./lxc-go-cli port mycontainer 8080 80 tcp
```

## Commands

| Command | Description |
|---------|-------------|
| `create` | Create LXC container with Docker support |
| `exec` | Execute interactive shell as app user |
| `port` | Configure port forwarding (TCP/UDP/both) |

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
# HTTP traffic
lxc-go-cli port web-server 8080 80

# Database with both protocols
lxc-go-cli port db-server 5432 5432 both
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
