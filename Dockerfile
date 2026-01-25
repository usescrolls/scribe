# Lightweight Alpine-based Dockerfile for testing Linux URL scheme IPC
#
# Build: docker build -t agenthub-test .
# Test:  docker run --rm agenthub-test /app/test-ipc.sh

# Build stage - Alpine with Go
FROM golang:1.21-alpine AS builder

# Install build dependencies for systray (GTK3)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    gtk+3.0-dev \
    libayatana-appindicator-dev

WORKDIR /build

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/

# Build the binary
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o agenthub-middleware ./cmd/agenthub-middleware

# Runtime stage - minimal Alpine
FROM alpine:3.19

# Install only runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    gtk+3.0 \
    libayatana-appindicator \
    bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/agenthub-middleware /app/agenthub-middleware

RUN chmod +x /app/agenthub-middleware

# Create test script
RUN echo '#!/bin/bash' > /app/test-ipc.sh && \
    echo 'set -e' >> /app/test-ipc.sh && \
    echo 'echo "Starting middleware in background..."' >> /app/test-ipc.sh && \
    echo '/app/agenthub-middleware -debug -no-gui &' >> /app/test-ipc.sh && \
    echo 'PID=$!' >> /app/test-ipc.sh && \
    echo 'sleep 2' >> /app/test-ipc.sh && \
    echo 'echo ""' >> /app/test-ipc.sh && \
    echo 'echo "Middleware started with PID $PID"' >> /app/test-ipc.sh && \
    echo 'echo ""' >> /app/test-ipc.sh && \
    echo 'echo "Sending test URL to running instance..."' >> /app/test-ipc.sh && \
    echo '/app/agenthub-middleware "agenthub://install?name=test-plugin&source=github&repo=test/repo"' >> /app/test-ipc.sh && \
    echo 'echo ""' >> /app/test-ipc.sh && \
    echo 'sleep 1' >> /app/test-ipc.sh && \
    echo 'echo "Checking registry:"' >> /app/test-ipc.sh && \
    echo 'cat ~/.agenthub-middleware/data/registry.json 2>/dev/null || echo "(not created)"' >> /app/test-ipc.sh && \
    echo 'echo ""' >> /app/test-ipc.sh && \
    echo 'echo "Stopping middleware..."' >> /app/test-ipc.sh && \
    echo 'kill $PID 2>/dev/null || true' >> /app/test-ipc.sh && \
    chmod +x /app/test-ipc.sh

CMD ["/bin/bash"]
