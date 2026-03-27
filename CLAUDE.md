# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

### Backend (Go)
```bash
go build                           # Build the invoiced binary
go run main.go                     # Run directly
./invoiced -v -d ~/billingdb       # Run with verbose logging and custom db path
./invoiced -h localhost:9999       # Specify HTTP listen address (default)
./invoiced -c ./config.toml        # Specify config file path
```

### Frontend (ReactJS/TypeScript/Vite)
```bash
cd static-src
yarn install                       # Install dependencies
npm run build                      # Build to ../static/assets/
npm run dev                        # Dev server on port 5173 (proxies /api to localhost:9999)
```

### Tests
```bash
go test ./...                      # Run all tests
go test ./rules/...                # Run tests in specific package
go test -v ./invoice/camt053/...   # Verbose test output
```

### Contrib Utilities
```bash
cd contrib/gen && go build && ./gen   # Generate auth credentials (IV, Salt, Hash)
cd contrib/desktop && go build        # Build desktop wrapper with systray
cd contrib/reindex && go build        # Rebuild search index
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
Generated via `utils.CreateInvoiceId()` combining current date and sequential counter per entity.

### Frontend
- Located in `static-src/`, built to `static/assets/`
- React 15, TypeScript, Vite bundler
- Axios for API calls, Chartist for metrics visualization
