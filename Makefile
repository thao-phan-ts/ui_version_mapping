# UI Version Mapping Tool Makefile

.PHONY: build test clean lint fmt vet install-deps run help

# Variables
BINARY_NAME=ui-version-check
BUILD_DIR=bin
CMD_DIR=cmd/ui-version-check
SCRIPTS_DIR=scripts

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@cd $(SCRIPTS_DIR) && go test -v

# Run specific test
test-complete:
	@echo "Running complete search test..."
	@cd $(SCRIPTS_DIR) && go test -v -run TestCompleteSearch

test-ab:
	@echo "Running A/B testing analysis test..."
	@cd $(SCRIPTS_DIR) && go test -v -run TestABTestingAnalysis

test-journey:
	@echo "Running journey export test..."
	@cd $(SCRIPTS_DIR) && go test -v -run TestIndividualJourneyExport

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@cd $(SCRIPTS_DIR) && go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...
	@cd $(SCRIPTS_DIR) && go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run --timeout=5m
	@cd $(SCRIPTS_DIR) && golangci-lint run --timeout=5m

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@cd $(SCRIPTS_DIR) && go mod tidy

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@./scripts/setup.sh auto

setup-local:
	@echo "Setting up local development..."
	@./scripts/setup.sh local

setup-remote:
	@echo "Setting up remote development..."
	@./scripts/setup.sh remote

setup-clean:
	@echo "Cleaning up old setup..."
	@./scripts/setup.sh clean

# Run the tool with default parameters
run: build
	@echo "Running $(BINARY_NAME) with default parameters..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -help

# Run example analysis
run-example: build
	@echo "Running example analysis..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -config 9054 -lead-source organic

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(SCRIPTS_DIR)/go.mod $(SCRIPTS_DIR)/go.sum
	@echo "Clean complete"

# Check code quality
check: fmt vet lint test
	@echo "All checks passed!"

# Show help
help:
	@echo "UI Version Mapping Tool - Makefile Commands"
	@echo ""
	@echo "Setup Commands:"
	@echo "  setup         Auto-setup development environment"
	@echo "  setup-local   Setup with local git submodules"
	@echo "  setup-remote  Setup for remote GitHub API"
	@echo "  setup-clean   Clean old setup"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         Build the binary"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Test Commands:"
	@echo "  test          Run all tests"
	@echo "  test-complete Run complete search test"
	@echo "  test-ab       Run A/B testing analysis test"
	@echo "  test-journey  Run journey export test"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           Format code"
	@echo "  vet           Vet code for issues"
	@echo "  lint          Run golangci-lint"
	@echo "  check         Run all quality checks"
	@echo ""
	@echo "Run Commands:"
	@echo "  run           Show help for the built tool"
	@echo "  run-example   Run example analysis"
	@echo ""
	@echo "Utility:"
	@echo "  install-deps  Install/update dependencies"
	@echo "  help          Show this help message" 