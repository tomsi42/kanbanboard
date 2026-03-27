# Kanban Board

A lightweight kanban board for individuals and small teams. Built as an alternative to heavyweight tools like Jira and overly basic tools like Kanboard.

## Features

- **Kanban board** with customizable columns and drag-and-drop
- **Task management** with labels, priority, due dates, and target versions
- **Subtasks** that move independently on the board with progress tracking
- **Comments** on tasks for team discussion
- **Teams** with shared project ownership and member management
- **Project visibility** — public (read-only for others) or private
- **User management** with roles (admin, team manager)
- **Label filtering** to focus on specific task types
- **Self-hosted** via Docker Compose

## Quick Start

```bash
docker compose up --build
```

Open http://localhost:8080. On first launch, you'll set up the admin account and application title.

## Development

### Prerequisites

- Go 1.26+
- Node.js 22+
- Docker and Docker Compose

### Run in development mode

Start the database:

```bash
docker compose up db
```

Start the backend (in `backend/`):

```bash
cd backend
go run ./cmd/server
```

Start the frontend (in `frontend/`):

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server runs at http://localhost:5173 and proxies API calls to the Go backend at http://localhost:8080.

### Run tests

```bash
cd backend
go test ./...
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `kanban` | PostgreSQL user |
| `DB_PASSWORD` | `kanban` | PostgreSQL password |
| `DB_NAME` | `kanbanboard` | PostgreSQL database name |
| `STATIC_DIR` | `../../frontend/dist` | Path to built frontend files |
| `MIGRATIONS_DIR` | `migrations` | Path to SQL migration files |

## Project Structure

```
kanbanboard/
├── backend/
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── handler/       # REST API handlers
│   │   ├── middleware/     # Auth middleware
│   │   ├── model/         # Domain entities
│   │   ├── store/         # Database access (plain SQL)
│   │   └── validate/      # Input validation
│   └── migrations/        # SQL migration files
├── frontend/
│   └── src/
│       └── lib/           # Svelte components
├── docs/
│   ├── user-guide.md      # User guide
│   ├── plan/              # Planning documents
│   └── skills/            # Claude Code skills
├── docker-compose.yml
├── LICENSE
└── README.md
```

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, standard library HTTP |
| Database | PostgreSQL 18, plain SQL (no ORM) |
| Frontend | Svelte 5, Vite |
| Drag & Drop | @thisux/sveltednd |
| Auth | Session-based (cookie + bcrypt) |
| Deployment | Docker Compose |

## Documentation

- [User Guide](docs/user-guide.md) — how to use the application

## License

[BSD-3-Clause](LICENSE)
