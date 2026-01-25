.PHONY: build build-all run clean install install-linux install-windows app dmg docker-build docker-test

BINARY_NAME=scribe
VERSION=1.0.0
BUILD_DIR=build
PACKAGING_DIR=packaging/macos

# Build for current platform
build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/scribe

# Build for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/scribe
	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/scribe
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/scribe
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/scribe
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/scribe

# Run locally
run:
	go run ./cmd/scribe

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Install to ~/.local/bin
install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"
	@echo "Make sure ~/.local/bin is in your PATH"

# Download dependencies
deps:
	go mod download
	go mod tidy

# Create macOS .app bundle (requires binary to exist in build/)
app:
	@echo "Creating macOS app bundle..."
	@./$(PACKAGING_DIR)/create-app.sh

# Create macOS DMG installer
dmg: app
	@echo "Creating macOS DMG installer..."
	@./$(PACKAGING_DIR)/create-dmg.sh

# Build everything including DMG (for releases)
release: build-all dmg
	@echo "Release build complete!"
	@echo "Binaries: $(BUILD_DIR)/"
	@echo "DMG: $(BUILD_DIR)/Scribe-Installer.dmg"

# Install on Linux (run from Linux system)
install-linux: build
	@./packaging/linux/install.sh

# Build Windows binary and show install instructions
# Note: Actual installation must be done on Windows
install-windows:
	@echo "Building Windows binary..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/scribe
	@echo ""
	@echo "Windows binary built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"
	@echo ""
	@echo "To install on Windows:"
	@echo "  1. Copy $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe to the Windows machine"
	@echo "  2. Copy packaging/windows/install.ps1 to the same location"
	@echo "  3. Run in PowerShell (as admin for system-wide, or with -UserInstall):"
	@echo "       .\\install.ps1"
	@echo "       .\\install.ps1 -UserInstall  # for user-only install"
	@echo ""

# Build Docker image for Linux testing (builds inside container)
docker-build:
	@echo "Building Docker image (compiles inside container)..."
	docker build -t scribe-test .

# Test IPC in Docker container
docker-test: docker-build
	@echo "Running IPC test in Docker..."
	@echo ""
	docker run --rm scribe-test /app/test-ipc.sh
