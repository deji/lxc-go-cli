FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    pkgconfig \
    lxc-dev \
    lxc

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o lxc-go-cli .

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    lxc \
    lxc-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary
COPY --from=builder /app/lxc-go-cli /usr/local/bin/

# Set entrypoint
ENTRYPOINT ["lxc-go-cli"] 