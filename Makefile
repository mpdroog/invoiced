.PHONY: build build-linux build-linux-arm64 build-all test lint lint-fix tygo frontend install-hooks clean

# Build the invoiced binary
build:
	go build -o invoiced .

# Cross-compile for Linux (amd64)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o invoiced-linux-amd64 .

# Cross-compile for Linux (arm64)
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o invoiced-linux-arm64 .

# Build for all platforms
build-all: build build-linux build-linux-arm64

# Run all tests
test:
	go test ./...

# Run linter
lint:
	~/go/bin/golangci-lint run ./...

# Run linter and fix issues where possible
lint-fix:
	~/go/bin/golangci-lint run --fix ./...

# Generate TypeScript types from Go structs
tygo:
	~/go/bin/tygo generate

# Build frontend (runs tygo first to ensure types are up-to-date)
frontend: tygo
	cd static-src && yarn build

# Install git hooks
install-hooks:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed successfully"

# Clean build artifacts
clean:
	rm -f invoiced invoiced-linux-amd64 invoiced-linux-arm64
	rm -f contrib/desktop/desktop
	rm -f contrib/gen/gen
	rm -f contrib/reindex/reindex

# Run all checks (useful before committing)
check: lint test build
	@echo "All checks passed!"
