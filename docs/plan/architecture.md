# Architecture

## Tech stack

| Layer | Choice | Rationale |
|---|---|---|
| **Backend** | Go, standard library HTTP + lightweight router | Simple, single binary, natural Docker fit, strong backend language for the developer |
| **Database** | PostgreSQL, standard `database/sql` + pgx driver | Relational data, team concurrency, no ORM - plain SQL |
| **Frontend** | Svelte (plain, not SvelteKit) | Drag-and-drop is the core interaction; HTMX struggles with this. Svelte is lightweight, compiler-based, minimal boilerplate |
| **Auth** | Session-based (cookie), bcrypt passwords | Simple, fits server-rendered approach |
| **Deployment** | Docker Compose (Go app + PostgreSQL) | Single server target, simple ops |

## Tool versions

| Tool | Version |
|---|---|
| Go | 1.26 |
| Svelte | 5 (latest 5.x) |
| PostgreSQL | 18 |
| Vite | latest stable |

## API conventions

- **Style:** REST
- **URL structure:** `/api/v1/{resource}` (e.g. `/api/v1/projects`, `/api/v1/tasks`)
- **Response format:** JSON
- **Error format:** `{ "error": "message" }` with appropriate HTTP status codes
- **Authentication:** Session cookie (browser) or `Authorization: Bearer <token>` (agents/programmatic). Middleware checks Bearer first.
- **JSON naming:** camelCase (matches JavaScript/Svelte convention)

## Key decisions

- **Svelte over HTMX:** Kanban drag-and-drop is fundamentally a client-side interaction. HTMX would require bolting on Sortable.js and Alpine.js. Svelte handles this natively with better ecosystem support (svelte-dnd-action).
- **No ORM:** Standard `database/sql` with pgx driver. 9 entities, manageable query count. No magic, full control.
- **Go backend serves built Svelte frontend** as static files. In development, Svelte dev server proxies API calls to Go.
- **Hand-rolled migration runner:** Simple Go code that tracks applied migrations in a `migrations` table and applies pending `.sql` files in order on startup.
- **Onboarding:** First-time setup screen when no users exist. Creates admin account and sets application title.
- **Dual authentication:** Session cookie (browser) and Bearer API token (programmatic/agent access). Middleware checks Bearer header first, then falls back to cookie. Same authorization logic applies to both.
- **MCP server as a separate binary:** `backend/cmd/mcp/` — a standalone process that wraps the REST API over stdio transport. Configured with an API token and base URL. Keeps the MCP concern isolated from the main server and allows it to be run as a CLI tool or sidecar.

## Project structure

```
kanbanboard/
├── backend/
│   ├── cmd/server/          # main.go - REST API + frontend serving entry point
│   ├── cmd/mcp/             # main.go - MCP server binary (v1.3)
│   ├── internal/
│   │   ├── model/           # domain entities
│   │   ├── store/           # database access (PostgreSQL)
│   │   ├── handler/         # REST API handlers + authorization logic
│   │   ├── middleware/      # auth middleware (RequireAuth, RequireAdmin)
│   │   └── validate/        # input validation (password, tag, priority, username)
│   ├── migrations/          # SQL migration files
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── lib/             # Svelte components and API client
│   │   └── App.svelte       # root component
│   ├── package.json
│   └── vite.config.js
├── docker-compose.yml
├── README.md
├── LICENSE
└── docs/
    └── plan/
```
