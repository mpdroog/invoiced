.PHONY: build test lint lint-fix tygo frontend install-hooks clean

# Build the invoiced binary
build:
	go build -o invoiced .

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
	rm -f invoiced
	rm -f contrib/desktop/desktop
	rm -f contrib/gen/gen
	rm -f contrib/reindex/reindex

# Run all checks (useful before committing)
check: lint test build
	@echo "All checks passed!"
