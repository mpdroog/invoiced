# InvoiceD

A self-hosted invoicing and hour registration system for freelancers and small businesses.

## Why InvoiceD?

**Own your data** - All data is stored as human-readable TOML files on your filesystem. No database lock-in, no proprietary formats. If something breaks, you can read and edit your invoices with any text editor.

**Built-in audit trail** - Every change is automatically committed to a local Git repository. Full version history, easy rollback, and optional sync to remote Git for backups.

**Works offline** - Runs entirely on your machine. No internet required, no cloud dependency, no subscription fees.

**Fast** - Pages load instantly. No waiting for cloud servers.

## Features

- Hour registration with conversion to invoices
- PDF invoice generation with UBL XML export
- Multi-entity support (manage multiple companies)
- CAMT053 bank statement import for payment matching
- VAT handling for NL, EU (reverse charge), and international exports
- Quarterly tax summaries

## Quick Start

### Build

```bash
# Frontend
cd static-src && yarn install && npm run build

# Backend
go build
```

### Setup

```bash
# Generate auth credentials
cd contrib/gen && go build && ./gen

# Initialize database
mkdir ~/billingdb
git clone https://github.com/mpdroog/acct-example ~/billingdb
# Edit ~/billingdb/entities.toml with your company details and generated credentials

# Run
./invoiced -v -d ~/billingdb
# Open http://localhost:9999
```

See [CLAUDE.md](CLAUDE.md) for detailed build commands and architecture overview.

## Tech Stack

- **Backend**: Go with Git integration
- **Frontend**: React, TypeScript, Vite
- **Storage**: TOML files + Git

## License

Open source - contributions welcome.

<a href="https://www.buymeacoffee.com/mpdroog">
    <img alt="Buy me a coffee" src="https://img.shields.io/static/v1.svg?label=%20&message=Buy%20me%20a%20coffee&color=579fbf&logo=buy%20me%20a%20coffee&logoColor=white"/>
</a>
