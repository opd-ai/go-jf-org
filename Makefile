# Makefile for go-jf-org

.PHONY: all build test clean install lint fmt help coverage build-all release

# Binary name
BINARY_NAME=go-jf-org
VERSION=0.8.0-dev

# Build directory
BUILD_DIR=bin
DIST_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## coverage: Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@$(GOCMD) tool cover -func=coverage.out | grep total

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html
	@echo "Clean complete"

## install: Install the binary system-wide
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ -n "$(PREFIX)" ]; then \
		install -d $(PREFIX)/bin; \
		install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(PREFIX)/bin/$(BINARY_NAME); \
		echo "Installed to $(PREFIX)/bin/$(BINARY_NAME)"; \
	elif [ -n "$(GOPATH)" ]; then \
		install -d $(GOPATH)/bin; \
		install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME); \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	else \
		install -d /usr/local/bin; \
		install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	fi

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## mod: Tidy and verify dependencies
mod:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	$(GOMOD) verify

## run: Build and run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	@echo "Building Linux amd64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	@echo "Building Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	@echo "Building macOS amd64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "Building macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "Building Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Build complete for all platforms in $(DIST_DIR)/"

## release: Create release archives
release: build-all
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && \
		tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip -q $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && \
		sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release archives created in $(DIST_DIR)/"

## ci: Run CI checks (test, lint, build)
ci: test lint build
	@echo "✓ All CI checks passed"
