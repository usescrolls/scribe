# Dockerfile for running Go unit tests
#
# Build: docker build -f test.Dockerfile -t scribe-test .
# Run:   docker run --rm scribe-test
# Run with coverage: docker run --rm -v $(pwd)/coverage:/coverage scribe-test \
#                    sh -c "go test -coverprofile=/coverage/coverage.out ./internal/..."

FROM golang:1.24-alpine

# Install build dependencies (needed for CGO and git operations)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    git

WORKDIR /app

# Allow Go to auto-download the required toolchain version
ENV GOTOOLCHAIN=auto

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Set environment for tests
ENV CGO_ENABLED=1
ENV HOME=/tmp/test-home

# Create test home directory
RUN mkdir -p /tmp/test-home

# Default command: run tests with verbose output
CMD ["go", "test", "-v", "-count=1", "./internal/..."]
