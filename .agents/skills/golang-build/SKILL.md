# Golang Build Skill

Quick workflow for building, testing, and packaging Go Lambda functions.

## Standard Workflow

After making code changes:

```bash
# Clean previous builds
make clean

# Run full pipeline: format → lint → test → build
make all
```

This runs all 4 stages automatically:
1. **Format** — `go fmt ./...`
2. **Vet** — `go vet ./...`
3. **Tests** — `go test -v -race ./...`
4. **Build** — Compiles all Lambda functions to `cmd/<handler>/bootstrap` and zips them

**If errors occur:** Read the error message, fix the source file, and run `make all` again.

## Individual Commands

| Command | Purpose |
|---|---|
| `make fmt` | Format code only |
| `make vet` | Run linter only |
| `make test` | Run unit tests only |
| `make build` | Build binaries only |
| `make clean` | Remove build artifacts |

## Build Output

After successful `make all`:
- `cmd/api/function.zip` — Order ingestion handler
- `cmd/verifications/check-address/function.zip` — Address check handler
- `cmd/verifications/check-credit/function.zip` — Credit check handler
- `cmd/verifications/check-payment/function.zip` — Payment check handler
- `cmd/verifications/stream-processor/function.zip` — DynamoDB Stream processor
- `cmd/verifications/update-status/function.zip` — Order status update handler

ZIP files are ready for Terraform deployment.

## Dependency Management

See README "Go (Golang) development" section for:
- Required Go version
- AWS SDK setup
- How to add new dependencies

## Notes

- All Lambda functions compile for **Linux AMD64** (`GOOS=linux GOARCH=amd64`)
- `CGO_ENABLED=0` ensures static linking (required for Lambda)
- Handler entry point is always `bootstrap`
- Use struct tags for JSON/DynamoDB mapping: `dynamodbav:"OrderId"`

See README for comprehensive setup guide.
