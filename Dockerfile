# Build stage
FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git build-base

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the MCP server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION}" \
    -o gitopia-mcp-server ./cmd/server

# Runtime stage
FROM ubuntu:24.04

# Set non-interactive frontend to avoid prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    wget \
    bash \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Install git-remote-gitopia using the official script
RUN wget -O - https://get.gitopia.com | bash

# Create non-root user for security
RUN groupadd --system mcp && useradd --system --gid mcp mcp
# Ensure home directory exists and is owned by mcp
RUN mkdir -p /home/mcp && chown mcp:mcp /home/mcp
# Set HOME environment variable for the mcp user
ENV HOME=/home/mcp

# Copy the built binary
COPY --from=builder /app/gitopia-mcp-server /usr/local/bin/gitopia-mcp-server

# Set up MCP directories structure
RUN mkdir -p /home/mcp/.mcp/gitopia/workspace \
    && mkdir -p /home/mcp/.mcp/gitopia/config \
    && chown -R mcp:mcp /home/mcp/.mcp

# Switch to non-root user
USER mcp
WORKDIR /home/mcp/.mcp/gitopia/workspace

# Set environment variables for configuration
ENV MCP_WORKSPACE_PATH=/home/mcp/.mcp/gitopia/workspace
ENV GIT_USER_NAME="Gitopia MCP Server"
ENV GIT_USER_EMAIL="mcp@gitopia.com"

# Health check to ensure the server starts properly
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/gitopia-mcp-server", "--version"] || exit 1

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/gitopia-mcp-server"]

# Default command (can be overridden)
CMD ["stdio"]
