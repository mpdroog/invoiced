# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

**Always use `make` commands when available.** This ensures consistent builds and proper tooling.

**IMPORTANT: Never run `golangci-lint` directly. Always use `make lint` or `make lint-fix`.**

**Git commits: Do not add "Co-Authored-By" lines.**

### Common Commands
```bash
make build                         # Build the invoiced binary
make frontend                      # Build frontend (runs tygo first)
make test                          # Run all Go tests
make lint                          # Run golangci-lint on all Go code
make lint-fix                      # Auto-fix lint issues where possible
make tygo                          # Generate TypeScript types from Go structs
make check                         # Run lint + test + build (pre-commit check)
make clean                         # Remove build artifacts
make install-hooks                 # Install git pre-commit hooks
```

### Running the Server
```bash
./invoiced -v -d ~/billingdb       # Run with verbose logging and custom db path
./invoiced -h localhost:9999       # Specify HTTP listen address (default)
./invoiced -c ./config.toml        # Specify config file path
```

### Frontend Development
**Note: Use yarn for package management (yarn.lock is committed).**
```bash
cd static-src
yarn install                       # Install dependencies (first time only)
yarn build                         # Typecheck + lint + build to ../static/assets/
yarn lint                          # Run ESLint only
yarn lint:fix                      # Auto-fix ESLint errors
```

### Contrib Utilities
```bash
cd contrib/gen && go build && ./gen   # Generate auth credentials (IV, Salt, Hash)
cd contrib/desktop && go build        # Build desktop wrapper with systray
```

### Rebuilding the Search Index
The SQLite search index is stored at `{dbPath}/index.db` (default: `acct/index.db`). To rebuild:
```bash
rm acct/index.db                      # Delete the index file
# Restart the server - it rebuilds automatically when empty
```

## Architecture Overview

### Data Storage Model
- All data stored as TOML files on the filesystem (human-readable, corruption-resistant)
- Git repository tracks all changes (automatic commits for audit trail)
- Database path specified via `-d` flag, defaults to `acct` directory
- Structure: `{entity}/{year}/{quarter}/` with buckets like `sales-invoices-paid`, `sales-invoices-unpaid`, `hours`

### Core Packages

**db/** - Filesystem abstraction with Git integration
- Uses go-git for version control
- `db.View()` for read-only transactions, `db.Update()` for write with auto-commit
- TOML encoding/decoding via BurntSushi/toml
- Path security via regex validation

**invoice/** - Invoice lifecycle management
- States: NEW -> CONCEPT -> FINAL (unpaid) -> paid
- PDF generation via gofpdf, UBL XML export
- CAMT053 bank statement parsing for payment matching
- VAT handling: NL (standard), EU0 (reverse charge), WORLD0 (export)

**hour/** - Hour registration tracking
- Converts to invoices via `Bill()` endpoint
- Links to invoice via `HourFile` field in invoice metadata

**taxes/** - Tax calculation and quarterly summaries
- Aggregates by quarter (Q1-Q4)
- Separates NL, EU (ICP), and Export invoices

**middleware/** - Authentication and HTTP utilities
- Session-based auth with AES-encrypted cookies
- Entity/company access control via `entities.toml`

**entities/** - Multi-tenant company management
- Debtors/customers with autocomplete search
- Project management per entity

### API Structure
RESTful API on `/api/v1/`:
- `/entities` - List/manage business entities
- `/invoices/:entity/:year` - Invoice CRUD
- `/hours/:entity/:year` - Hour registration
- `/summary/:entity/:year` - Tax summaries
- `/taxes/:entity/:year/:quarter` - Quarterly tax calculations

### Configuration
- `config.toml` - SMTP queues for email
- `entities.toml` (in db path) - Companies, users, auth credentials
- Generate credentials: `contrib/gen` creates IV, Salt, Hash for entities.toml

### Invoice ID Format
Generated via `utils.CreateInvoiceID()` combining current date and sequential counter per entity.

### Frontend
- Located in `static-src/`, built to `static/assets/`
- React 18, TypeScript (strict mode), Vite bundler
- Axios for API calls, Recharts for metrics visualization
- ESLint with TypeScript rules (no-floating-promises, no-misused-promises)
- TypeScript strict settings: strictNullChecks, noUncheckedIndexedAccess

## Linting

The project uses golangci-lint v2 configured in `.golangci.yml`. **Always run `make lint` before committing Go code.**

### Enabled Linters
- **govet, staticcheck, errcheck** - Core Go analysis
- **revive** - Style and best practices (replaces golint)
- **gosec** - Security vulnerability detection
- **errorlint** - Proper error wrapping with `%w` and `errors.Is()`
- **goconst** - Repeated strings that should be constants
- **noctx** - HTTP requests must include context
- **bodyclose** - HTTP response bodies must be closed
- **misspell** - Spelling errors in comments and strings

### Common Lint Fixes

**Error wrapping** (errorlint):
```go
// Bad
return fmt.Errorf("failed: %s", err)
// Good
return fmt.Errorf("failed: %w", err)
```

**Error comparison** (errorlint):
```go
// Bad
if err == io.EOF { ... }
// Good
if errors.Is(err, io.EOF) { ... }
```

**HTTP with context** (noctx):
```go
// Bad
http.Get(url)
// Good
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
http.DefaultClient.Do(req)
```

**Doc comments** (revive):
```go
// Bad
// helper function for X
func DoSomething() {}
// Good
// DoSomething performs X operation.
func DoSomething() {}
```

**Naming conventions** (revive var-naming):
```go
// Bad: ID not Id, URL not Url, HTTP not Http
InvoiceId, ApiUrl, HttpClient
// Good
InvoiceID, APIURL, HTTPClient
```

**Type stuttering** - When a type name repeats the package name:
```go
// Bad
package config
type ConfigQueue struct {} // config.ConfigQueue stutters

// Good - rename the type
type Queue struct {} // config.Queue is clear

// Or add nolint if renaming would break API
type ConfigQueue struct {} //nolint:revive // backwards compatibility
```

**Unused parameters**:
```go
// Bad
func handler(w http.ResponseWriter, r *http.Request) {
    // r is never used
}
// Good
func handler(w http.ResponseWriter, _ *http.Request) {
    // underscore indicates intentionally unused
}
```

### JSON API Conventions

**Always initialize slices for JSON responses.** A nil slice encodes as `null`, not `[]`. The `writer.Encode()` function will error if it detects a nil slice.

```go
// Bad - encodes as null
var items []Item
return writer.Encode(w, r, Response{Items: items})

// Good - encodes as []
items := []Item{}
return writer.Encode(w, r, Response{Items: items})
```

This applies to:
- Direct slices passed to `writer.Encode()`
- Slice fields in response structs
- Slices returned from functions that end up in API responses

### Suppressing Lint Warnings

Use `//nolint` directives sparingly and with justification:
```go
func example() { //nolint:gosec // G404: math/rand OK for non-crypto use
    rand.Intn(100)
}
```

### Excluded Paths
The following paths are excluded from linting (configured in `.golangci.yml`):
- `embed/` - Embedded static files
- `contrib/` - Utility tools
- `static-src/` - Frontend code (has its own ESLint)
