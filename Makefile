.PHONY: build build-frontend dev run clean install deps test test-verbose coverage coverage-html wails-generate wails-ensure-bindings \
        app app-run lint lint-fix install-hooks

BINARY_NAME=scribe
BUILD_DIR=build
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")-dev
LDFLAGS=-s -w -X github.com/usescrolls/scribe/internal.Version=$(VERSION)

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
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/bin/$(BINARY_NAME) .
else
	CGO_ENABLED=1 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/bin/$(BINARY_NAME) .
endif

# Build frontend only (generates bindings if missing)
build-frontend: wails-ensure-bindings
	cd frontend && pnpm run build

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
	rm -rf frontend/bindings
	rm -rf frontend/coverage
	rm -rf coverage
	rm -f coverage*.out

# Install to ~/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/bin/$(BINARY_NAME) $(HOME)/.local/bin/
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"
	@echo "Make sure ~/.local/bin is in your PATH"

# Download dependencies and install tools
deps:
	go install github.com/wailsapp/wails/v3/cmd/wails3@latest
	go mod download
	go mod tidy
	cd frontend && pnpm install

# Generate Wails v3 bindings (always regenerates)
wails-generate:
	wails3 generate bindings

# Generate Wails v3 bindings only if they don't exist
wails-ensure-bindings:
	@if [ ! -d frontend/bindings ]; then \
		echo "Generating Wails v3 bindings..."; \
		wails3 generate bindings; \
	fi

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

# Run linter
lint:
	golangci-lint run ./...

# Run linter and fix issues
lint-fix:
	golangci-lint run --fix ./...

# Install git hooks
install-hooks:
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	cp scripts/hooks/commit-msg .git/hooks/commit-msg
	chmod +x .git/hooks/commit-msg
	@echo "Git hooks installed"

# ============================================================================
# macOS App Bundle
# ============================================================================

# Create macOS .app bundle (requires build first)
app: build
	VERSION=$(VERSION) ./packaging/macos/create-app.sh
	@echo ""
	@echo "To run the app with the proper icon:"
	@echo "  make app-run"
	@echo "  or: open build/Scribe.app"

# Run the macOS .app bundle
app-run:
	open $(BUILD_DIR)/Scribe.app
