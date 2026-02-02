.PHONY: build build-frontend dev run clean install deps test test-verbose coverage coverage-html wails-generate \
        docker-test docker-test-coverage docker-test-race docker-test-build docker-test-clean

BINARY_NAME=scribe
VERSION=1.0.0
BUILD_DIR=build

# macOS deployment target (set to current OS version to avoid linker warnings)
MACOS_VERSION := $(shell sw_vers -productVersion 2>/dev/null || echo "")

# Build for current platform (frontend + Go)
build: build-frontend
	mkdir -p $(BUILD_DIR)/bin
ifeq ($(shell uname),Darwin)
	CGO_ENABLED=1 \
	MACOSX_DEPLOYMENT_TARGET=$(MACOS_VERSION) \
	CGO_CFLAGS="-mmacosx-version-min=$(MACOS_VERSION)" \
	CGO_LDFLAGS="-mmacosx-version-min=$(MACOS_VERSION)" \
	go build -ldflags="-s -w" -o $(BUILD_DIR)/bin/$(BINARY_NAME) .
else
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BUILD_DIR)/bin/$(BINARY_NAME) .
endif

# Build frontend only
build-frontend:
	cd frontend && npm run build

# Development mode with hot reload (Wails v3)
dev:
	wails3 dev

# Run CLI command directly
run:
	go run . list

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Install to ~/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/bin/$(BINARY_NAME) $(HOME)/.local/bin/
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"
	@echo "Make sure ~/.local/bin is in your PATH"

# Download dependencies
deps:
	go mod download
	go mod tidy
	cd frontend && npm install

# Generate Wails v3 bindings
wails-generate:
	wails3 generate bindings

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./internal/... ./cli/...

# Run tests with coverage
coverage:
	go test -cover ./internal/... ./cli/...

# Generate HTML coverage report
coverage-html:
	mkdir -p $(BUILD_DIR)
	go test -coverprofile=$(BUILD_DIR)/coverage.out ./internal/... ./cli/...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# ============================================================================
# Docker Test Targets
# ============================================================================

# Build the test Docker image
docker-test-build:
	docker build -f test.Dockerfile -t scribe-test .

# Run tests in Docker
docker-test: docker-test-build
	docker run --rm scribe-test

# Run tests with coverage in Docker
docker-test-coverage: docker-test-build
	mkdir -p coverage
	docker run --rm -v $(PWD)/coverage:/coverage scribe-test \
		sh -c "go test -v -count=1 -coverprofile=/coverage/coverage.out ./internal/... && \
		       go tool cover -func=/coverage/coverage.out"
	@echo "Coverage report saved to coverage/coverage.out"

# Run tests with race detector in Docker
docker-test-race: docker-test-build
	docker run --rm scribe-test go test -v -race -count=1 ./internal/...

# Run filtered tests in Docker (usage: make docker-test-filter TEST_PATTERN=TestSkill)
docker-test-filter: docker-test-build
	docker run --rm scribe-test go test -v -count=1 -run "$(TEST_PATTERN)" ./internal/...

# Clean Docker test artifacts
docker-test-clean:
	docker rmi scribe-test 2>/dev/null || true
	rm -rf coverage/
