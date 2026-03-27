# Kanban Board - Claude Code Navigation

## Project Overview

Lightweight kanban board for individuals and small teams. Go backend + Svelte frontend + PostgreSQL.

## Build & Run

```bash
# Docker (full stack)
docker compose up --build

# Dev mode - backend
cd backend && go run ./cmd/server

# Dev mode - frontend
cd frontend && npm run dev

# Run backend tests
cd backend && go test ./...
```

## Project Structure

- `backend/cmd/server/` - Go entry point
- `backend/internal/auth/` - Authentication and sessions
- `backend/internal/model/` - Domain entities
- `backend/internal/store/` - Database access (PostgreSQL, plain SQL)
- `backend/internal/handler/` - REST API handlers
- `backend/internal/middleware/` - HTTP middleware
- `backend/migrations/` - SQL migration files
- `frontend/src/` - Svelte application

## Key Decisions

- No ORM - standard `database/sql` with pgx driver
- REST API at `/api/v1/`
- JSON responses, camelCase naming
- Session-based auth with cookies
- Hand-rolled migration runner

## Planning & Skills

- `docs/plan/` - Planning documents (user stories, domain model, architecture, UX, dev workflow, testing, implementation plan)
- `docs/skills/` - Planning skills (reusable across projects)

## Versioning

`v0.{phase}.{subphase}` during development. `v1.0.0` at completion.
