.PHONY: build build-frontend dev run clean install deps test test-verbose coverage coverage-html wails-generate

BINARY_NAME=scribe
VERSION=1.0.0
BUILD_DIR=build

# Build for current platform (frontend + Go)
build: build-frontend
	mkdir -p $(BUILD_DIR)/bin
	go build -ldflags="-s -w" -o $(BUILD_DIR)/bin/$(BINARY_NAME) .

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
