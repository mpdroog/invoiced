# Invoiced Architecture Overview

This document provides a high-level overview of the invoiced codebase for quick onboarding.

## Technology Stack

- **Backend**: Go 1.22+
- **Frontend**: React 18, TypeScript (strict mode), Vite
- **Data Storage**: TOML files on filesystem with Git versioning
- **Search Index**: SQLite with WAL mode
- **Build Tools**: Make, golangci-lint v2, tygo (Go-to-TypeScript types)

## Data Model

All data is stored as TOML files, organized by entity and year:

```
{dbPath}/
  {entity}/
    debtors.toml           # Customer definitions
    projects.toml          # Project configurations
    counters.toml          # Invoice ID counters
    logo.png               # Entity logo
    {year}/
      concepts/
        sales-invoices/    # Draft invoices
        hours/             # Draft hour registrations
      Q1/ Q2/ Q3/ Q4/      # Quarterly buckets
        sales-invoices-unpaid/
        sales-invoices-paid/
        hours/
        purchase-invoices-unpaid/
        purchase-invoices-paid/
  entities.toml            # User/company configuration
  index.db                 # SQLite search index
```

## Key Packages

### Backend (Go)

| Package | Purpose |
|---------|---------|
| `db/` | TOML file operations with Git auto-commit |
| `entities/` | Debtors, projects, company management |
| `invoice/` | Invoice lifecycle (create, finalize, pay, PDF/XML) |
| `hour/` | Hour registration and billing |
| `purchase/` | Purchase invoice handling |
| `taxes/` | Tax calculations and quarterly summaries |
| `idx/` | SQLite search index (sync on commit, full rebuild) |
| `middleware/` | Authentication (AES-encrypted session cookies) |
| `writer/` | JSON/msgpack response encoding |

### Frontend (TypeScript)

| Directory | Purpose |
|-----------|---------|
| `app/cmp/dashboard/` | Revenue charts, quick stats |
| `app/cmp/invoices/` | Invoice list and editor |
| `app/cmp/hours/` | Hour registration list and editor |
| `app/cmp/projects/` | Debtors and projects CRUD |
| `app/cmp/purchases/` | Purchase invoice management |
| `app/cmp/taxes/` | Tax summary views |
| `app/cmp/git/` | Git sync status |
| `app/shared/` | Navbar, modals, common components |
| `app/types/` | Generated TypeScript types (via tygo) |

## API Structure

REST API on `/api/v1/`:

| Endpoint Pattern | Package | Description |
|------------------|---------|-------------|
| `/entities` | entities | List/manage business entities |
| `/debtors/:entity` | entities | CRUD for customers |
| `/projects/:entity` | entities | CRUD for projects |
| `/invoices/:entity/:year` | invoice | Invoice CRUD + lifecycle |
| `/hours/:entity/:year` | hour | Hour registration CRUD |
| `/purchases/:entity/:year` | purchase | Purchase invoice CRUD |
| `/taxes/:entity/:year/:quarter` | taxes | Tax calculations |
| `/git/:entity/*` | git | Git sync operations |

## Database Transaction Pattern

```go
// Read-only (no commit)
db.View(func(t *db.Txn) error {
    return t.Open(path, &data)
})

// Write (auto-commits to Git)
db.Update(db.Commit{Name, Email, Message}, func(t *db.Txn) error {
    return t.Save(path, isNew, data)
})
```

## SQLite Index

The `idx/` package maintains a SQLite cache for fast queries:

- **Tables**: invoices, hours, purchase_invoices, debtors, projects
- **Sync**: `db.OnCommit` callback triggers `idx.SyncPath()` after every Git commit
- **Rebuild**: `idx.Rebuild()` called on startup if index is empty, or after Git pull/reset

## Frontend Routing

Hash-based routing in `app/main.tsx`:

```
#                                    -> EntitiesApp (company list)
#entity/year/                        -> DashboardApp
#entity/year/hours                   -> HoursList
#entity/year/hours/edit/bucket/id    -> HoursEdit
#entity/year/invoices                -> InvoicesList
#entity/year/invoices/edit/bucket/id -> InvoicesEdit
#entity/year/projects                -> ProjectsList
#entity/year/projects/debtor/edit/id -> DebtorEdit
#entity/year/projects/project/edit/id-> ProjectEdit
```

## Type Generation

Go structs are converted to TypeScript via tygo (configured in `tygo.yaml`):

```bash
make tygo  # Generates static-src/app/types/*.ts
```

## Linting

- **Go**: `make lint` (golangci-lint v2 with strict rules)
- **TypeScript**: `yarn lint` (ESLint with strict-boolean-expressions)

See `CLAUDE.md` for detailed linting rules and common fixes.

## Quick Commands

```bash
make build         # Build Go binary
make frontend      # Build frontend (runs tygo first)
make check         # Run lint + test + build
make tygo          # Generate TypeScript types

cd static-src
yarn build         # Build frontend only
yarn lint:fix      # Auto-fix ESLint issues
```
