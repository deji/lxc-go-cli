#!/bin/bash

set -e

echo "Building LXC Go CLI..."

# Build with optimizations
echo "Building optimized binary..."
go build -ldflags="-s -w" -o lxc-go-cli .

# Build for different platforms
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o lxc-go-cli .

echo "Build complete!"
echo "Available binaries:"
ls -la lxc-go-cli* 