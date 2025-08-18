# Makefile for {{.project_name}}

.PHONY: build test clean run install help

# Variables
BINARY_NAME={{.project_name}}
MODULE_NAME={{.module_name}}
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%d %H:%M:%S UTC")

# Build flags
LDFLAGS=-X '$(MODULE_NAME)/cmd.Version=$(VERSION)' \
        -X '$(MODULE_NAME)/cmd.Commit=$(COMMIT)' \
        -X '$(MODULE_NAME)/cmd.Date=$(DATE)'

# Default target
all: build

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "✅ Build complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

# Run the application
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

# Install the application
install:
	@echo "📦 Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" .

# Development build with debugging
dev-build:
	@echo "🔧 Building for development..."
	@go build -gcflags="all=-N -l" -o $(BINARY_NAME) .

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run

# Tidy dependencies
tidy:
	@echo "📋 Tidying dependencies..."
	@go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  run        - Build and run the application"
	@echo "  install    - Install the application"
	@echo "  dev-build  - Build with debugging symbols"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  tidy       - Tidy dependencies"
	@echo "  help       - Show this help"