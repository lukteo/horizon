# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Horizon is a monorepo with a React/TypeScript frontend (`client/`) and a Go backend (`server/`), using OpenAPI as the contract between them. Development is primarily driven via `make` commands.

## Common Commands

### Initial Setup
```bash
make init          # Full setup: DB + server deps + client deps
make dup           # Start Docker services (postgres, redis, client)
```

### Development
```bash
make rserver       # Run Go server locally (go run ./cmd/server/main.go)
make rclient       # Run client dev server (bun run dev)
```

### Lint & Format
```bash
make lint          # Lint entire codebase (biome + golangci-lint)
make format        # Format entire codebase
make format-client # Format client only (biome)
make format-server # Format server only (goimports + golines + gofmt)
```

### Code Generation (run after modifying openapi/openapi.yml)
```bash
make gen-openapi   # Regenerates client RTK Query code and server handler interface
```

### Database Migrations
```bash
make mcreate       # Create a new migration (prompts for name)
make mup           # Apply pending migrations + regenerate Jet models
make mdown         # Rollback 1 migration
make mstatus       # Show migration status
make mreset        # Drop and recreate database (destructive)
```

### Docker
```bash
make ddown         # Stop all containers
make drestart      # Restart all containers
make dlogs         # Tail logs
make dshell-db     # Open psql in the postgres container
```

## Architecture

### Contract-First API Development

`openapi/openapi.yml` is the single source of truth for the API contract. Changes here drive code generation for both sides:
- **Server**: generates a strict Go interface in `server/generated/oapi/` — handlers must satisfy this interface
- **Client**: generates RTK Query hooks in `client/src/` via `@rtk-query/codegen-openapi`

Always run `make gen-openapi` after modifying the OpenAPI spec.

### Backend (`server/`)

Entry point: `cmd/server/main.go` → `internal/boot/` → Chi router → OpenAPI strict handler

- **`internal/boot/`** — server initialization, router setup, middleware registration
- **`internal/config/`** — lazy-loaded service providers (DB, Clerk, Logger) via `config/provider/`
- **`internal/middleware/`** — Clerk JWT auth middleware (injects user into context)
- **`internal/web/`** — HTTP handlers; `handler.go` embeds the base struct, each file adds endpoint implementations
- **`generated/oapi/`** — do not edit; regenerated from OpenAPI spec
- **`generated/horizon/`** — do not edit; Jet-generated type-safe SQL models from DB schema
- **`migrations/`** — Goose SQL migrations (sequential numbered files)

Authentication uses Clerk — JWT is validated in middleware, user data is synced to the `users` table (keyed on `clerk_id`).

Database queries use **Go-Jet** (type-safe SQL builder). Models in `generated/horizon/public/model/` and query builders in `generated/horizon/public/table/`.

### Frontend (`client/`)

Entry: `src/main.tsx` → `src/App.tsx` → `src/routes/` → pages/components

- **`src/routes/`** — React Router route definitions
- **`src/pages/`** — page-level components
- **`src/components/`** — reusable components
- Path alias `@` maps to `src/`

Stack: React 19, Vite, Ant Design, Redux Toolkit, React Router, React Hook Form + Zod, i18next.

The React Compiler (Babel plugin) is enabled — avoid manual `useMemo`/`useCallback` where the compiler handles it.

### Tooling

- **Biome** handles both linting and formatting for the client (tabs, double quotes)
- **golangci-lint** with 60+ linters for the server (config in `server/.golangci.yaml`)
- **Bun** is the package manager for the client (`bun.lock` — do not use npm/yarn)
