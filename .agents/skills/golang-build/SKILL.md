# Golang Build Skill

This skill provides comprehensive guidance for building and packaging Go applications, specifically tailored for the E-Commerce Order Orchestration project's Lambda-based microservices architecture.

## Overview

The project uses Go to build Lambda function handlers that must be compiled for Linux (`GOOS=linux GOARCH=amd64`) and packaged as ZIP files for AWS Lambda deployment. This skill covers the full build pipeline from dependency management through final artifact packaging.

## Prerequisites

- **Go 1.21 or later** installed on your system
- **AWS Lambda Go SDK** available in the project's dependencies
- Understanding of the project structure: `cmd/api/`, `cmd/verifications/`, `cmd/wms-integration/`

## Build Process Steps

### Step 1: Initialize Go Module (First Time Only)

If the project doesn't have a `go.mod` file yet:

```bash
# From the project root
go mod init ecommerce-order-aws-go
```

This creates the module definition file that tracks all dependencies.

### Step 2: Install Dependencies

Install the official AWS Lambda Go SDK and any other required packages:

```bash
# Install AWS Lambda Go SDK
go get github.com/aws/aws-lambda-go/events
go get github.com/aws/aws-lambda-go/lambda

# Install other common dependencies (as needed)
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/service/stepfunctions

# Download all dependencies for the project
go mod download
go mod tidy
```

**Note:** `go mod tidy` removes unused dependencies and ensures consistency in `go.mod` and `go.sum`.

### Step 3: Build Lambda Function Binaries

Each Lambda function must be compiled independently for the Linux/AMD64 platform. The handler entry point is named `bootstrap` by convention in AWS Lambda.

#### Single Lambda Build Example

```bash
# Build the Order Ingestion API handler
cd cmd/api
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
cd ../../
```

**Build Flags Explained:**
- `GOOS=linux`: Target operating system (Linux)
- `GOARCH=amd64`: Target CPU architecture (x86-64)
- `CGO_ENABLED=0`: Disable C/C++ cgo bindings (required for static linking and Lambda compatibility)
- `-o bootstrap`: Output file name (AWS Lambda looks for this name by default)

#### Build All Lambda Functions

```bash
# From project root, build all handlers sequentially
for handler in api verifications/address verifications/credit verifications/payment wms-integration; do
  echo "Building $handler..."
  cd "cmd/$handler"
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
  cd ../../
done
```

Or use a **Makefile** for automation (recommended):

```makefile
.PHONY: build build-api build-verifications build-wms clean

build: build-api build-verifications build-wms

build-api:
	cd cmd/api && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	cd ../../

build-verifications:
	cd cmd/verifications/address && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	cd ../../../
	cd cmd/verifications/credit && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	cd ../../../
	cd cmd/verifications/payment && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	cd ../../../

build-wms:
	cd cmd/wms-integration && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	cd ../../

clean:
	find cmd -name bootstrap -delete
	find cmd -name function.zip -delete
```

Then run: `make build` or `make build-api` for individual handlers.

### Step 4: Package Binaries as ZIP Files

AWS Lambda requires function code to be packaged as ZIP files. Each `bootstrap` binary must be zipped independently.

```bash
# Package the compiled binary
cd cmd/api
zip function.zip bootstrap
cd ../../

# Verify the archive
unzip -l function.zip
```

**Output:**
```
Archive:  function.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
  8234567   2026-10-15 14:23   bootstrap
---------                     -------
  8234567                     1 file
```

### Step 5: Verify Build Artifacts

Ensure all build artifacts are correctly generated:

```bash
# Check for all bootstrap binaries
find cmd -name bootstrap -type f

# Verify ZIP file contents
for zip_file in $(find cmd -name function.zip); do
  echo "Checking $zip_file..."
  unzip -l "$zip_file"
done
```

### Step 6: Automated Testing

Before deployment, run tests:

```bash
# Run all tests in the project
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/models
```

## Build Automation via Makefile

Create a `Makefile` in the project root:

```makefile
.PHONY: build clean test coverage lint format help all

LAMBDA_HANDLERS := api verifications/address verifications/credit verifications/payment wms-integration

all: format lint test build

build:
	@echo "Building all Lambda handlers..."
	$(foreach handler,$(LAMBDA_HANDLERS),$(call build_handler,$(handler)))

define build_handler
	cd cmd/$(1) && \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go && \
	zip function.zip bootstrap && \
	cd ../../ && \
	echo "✓ Built $(1)"
endef

test:
	@echo "Running tests..."
	go test -v ./...

coverage:
	@echo "Generating coverage report..."
	go test -cover ./...

lint:
	@echo "Linting Go code..."
	golint ./...

format:
	@echo "Formatting Go code..."
	go fmt ./...

clean:
	@echo "Cleaning build artifacts..."
	find cmd -name bootstrap -delete
	find cmd -name function.zip -delete

help:
	@echo "Available targets:"
	@echo "  make build       - Build all Lambda handlers"
	@echo "  make test        - Run all tests"
	@echo "  make coverage    - Generate coverage report"
	@echo "  make lint        - Lint Go code"
	@echo "  make format      - Format Go code"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make all         - Format, lint, test, and build"
```

## Troubleshooting

### Build Error: "command not found: go"
- Ensure Go is installed: `go version`
- If not installed, download from [golang.org](https://golang.org/dl/)

### Error: "cannot find package"
- Run `go mod download` to fetch all dependencies
- Check `go.mod` for correct module path: `module ecommerce-order-aws-go`

### Lambda Runtime Error: "failed to start function"
- Verify the binary name is `bootstrap` (case-sensitive)
- Ensure binary was compiled with `GOOS=linux GOARCH=amd64`
- Check that the ZIP file contains the `bootstrap` file at the root level

### Large Binary Size
- Use build flags to reduce size:
  ```bash
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap main.go
  ```
  - `-s`: Omit symbol table
  - `-w`: Omit debugging information

## Integration with CI/CD

The build process integrates with GitHub Actions (see `.github/workflows/build-golang.yml`):

```yaml
# Typical GitHub Actions job for building Go code
- name: Build Lambda Functions
  run: |
    make clean
    make build
```

## Summary

The Golang build process for this project:
1. **Initialize** dependencies with `go mod init` and `go get`
2. **Compile** each Lambda handler binary for Linux/AMD64
3. **Package** each binary as a ZIP file for Lambda deployment
4. **Test** the code before deployment
5. **Automate** everything with a Makefile or GitHub Actions

For the full development workflow, refer to the project README.md and the Terraform Infrastructure as Code skill.
