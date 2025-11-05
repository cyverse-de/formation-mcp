# Formation MCP Justfile
# Common development tasks for the Formation MCP server

# Default recipe - show available commands
default:
    @just --list

# Build the formation-mcp binary
build:
    CGO_ENABLED=0 go build -o formation-mcp ./cmd/formation-mcp

# Build with version info
build-versioned VERSION:
    CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o formation-mcp ./cmd/formation-mcp

# Run all tests
test:
    go test ./... -v

# Run tests with coverage
test-coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
    go test ./... -race -v

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...
    gofmt -s -w .

# Tidy dependencies
tidy:
    go mod tidy

# Clean build artifacts
clean:
    rm -f formation-mcp
    rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install:
    go install ./cmd/formation-mcp

# Run the formation-mcp server in development mode (with debug logging)
dev:
    @echo "Starting formation-mcp in development mode..."
    LOG_LEVEL=debug ./formation-mcp

# Build for multiple platforms (Linux, macOS, Windows)
build-all VERSION:
    @echo "Building for Linux (amd64)..."
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o dist/formation-mcp-linux-amd64 ./cmd/formation-mcp
    @echo "Building for Linux (arm64)..."
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o dist/formation-mcp-linux-arm64 ./cmd/formation-mcp
    @echo "Building for macOS (amd64)..."
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o dist/formation-mcp-darwin-amd64 ./cmd/formation-mcp
    @echo "Building for macOS (arm64)..."
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o dist/formation-mcp-darwin-arm64 ./cmd/formation-mcp
    @echo "Building for Windows (amd64)..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X main.version={{VERSION}}" -o dist/formation-mcp-windows-amd64.exe ./cmd/formation-mcp
    @echo "All builds complete! Check the dist/ directory."

# Run all checks (format, lint, test)
check: fmt lint test
    @echo "All checks passed!"

# Quick build and test cycle
quick: build test
    @echo "Quick build and test complete!"

# Show binary information
info:
    @echo "Binary info:"
    @ls -lh formation-mcp 2>/dev/null || echo "Binary not found. Run 'just build' first."
    @echo ""
    @echo "Version:"
    @./formation-mcp --version 2>/dev/null || echo "Binary not found or not executable."

# Run benchmarks
bench:
    go test ./... -bench=. -benchmem

# Generate mocks (if using mockgen)
mocks:
    @echo "Generating mocks..."
    go generate ./...

# Verify dependencies are secure (requires govulncheck)
security:
    govulncheck ./...

# Update dependencies to latest versions
update-deps:
    go get -u ./...
    go mod tidy

# Create distribution archive
dist VERSION: clean
    @mkdir -p dist
    @just build-all {{VERSION}}
    @echo "Creating distribution archives..."
    @cd dist && tar -czf formation-mcp-{{VERSION}}-linux-amd64.tar.gz formation-mcp-linux-amd64
    @cd dist && tar -czf formation-mcp-{{VERSION}}-linux-arm64.tar.gz formation-mcp-linux-arm64
    @cd dist && tar -czf formation-mcp-{{VERSION}}-darwin-amd64.tar.gz formation-mcp-darwin-amd64
    @cd dist && tar -czf formation-mcp-{{VERSION}}-darwin-arm64.tar.gz formation-mcp-darwin-arm64
    @cd dist && zip formation-mcp-{{VERSION}}-windows-amd64.zip formation-mcp-windows-amd64.exe
    @echo "Distribution archives created in dist/"
    @ls -lh dist/*.tar.gz dist/*.zip

# Run specific test package
test-pkg PACKAGE:
    go test -v ./{{PACKAGE}}/...

# Show test coverage for specific package
cover-pkg PACKAGE:
    go test ./{{PACKAGE}}/... -coverprofile=coverage-{{PACKAGE}}.out
    go tool cover -func=coverage-{{PACKAGE}}.out

# Run go vet
vet:
    go vet ./...

# Check for code smells (requires staticcheck)
staticcheck:
    staticcheck ./...
