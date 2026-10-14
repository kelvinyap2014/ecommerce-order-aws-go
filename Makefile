.PHONY: help build test coverage vet format clean all

# Color output
BLUE := \033[0;34m
GREEN := \033[0;32m
NC := \033[0m # No Color

# Default target
help:
	@echo "$(BLUE)E-Commerce Order Orchestration - Available targets:$(NC)"
	@echo ""
	@echo "$(GREEN)Build:$(NC)"
	@echo "  make build           - Build all Lambda function binaries"
	@echo "  make clean           - Remove build artifacts (bootstrap, function.zip)"
	@echo ""
	@echo "$(GREEN)Testing & Quality:$(NC)"
	@echo "  make test            - Run all unit tests"
	@echo "  make coverage        - Generate test coverage report"
	@echo "  make vet             - Run go vet static analyzer"
	@echo "  make format          - Format Go code (gofmt)"
	@echo ""
	@echo "$(GREEN)Dependencies:$(NC)"
	@echo "  make deps            - Download and tidy dependencies"
	@echo ""
	@echo "$(GREEN)CI/CD:$(NC)"
	@echo "  make all             - Format, vet, test, and build (complete pipeline)"
	@echo ""

# Download and manage dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies up to date$(NC)"

# Build all Lambda functions
build: clean
	@echo "Building Lambda functions..."
	@cd cmd/api && \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
		zip -q function.zip bootstrap && \
		echo "$(GREEN)✓ Built Order Ingestion API$(NC)" && \
		cd ../../
	@echo ""
	@echo "$(GREEN)All Lambda functions built successfully!$(NC)"

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests passed$(NC)"

# Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

# Run go vet static analyzer
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "$(GREEN)✓ Vet passed$(NC)"

# Format code
format:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

# Check format without modifying
format-check:
	@echo "Checking code format..."
	@if [ -n "$$(go fmt ./...)" ]; then \
		echo "$(RED)Code is not formatted. Run 'make format'$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Format check passed$(NC)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@find cmd -name bootstrap -delete
	@find cmd -name function.zip -delete
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

# Complete pipeline: format → vet → test → build
all: format-check vet test build
	@echo ""
	@echo "$(GREEN)✅ Complete pipeline passed!$(NC)"
