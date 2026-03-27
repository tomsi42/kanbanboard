# Kanban Board

A lightweight kanban board for individuals and small teams. Built as an alternative to heavyweight tools like Jira and basic tools like Kanboard.

## Tech Stack

- **Backend:** Go 1.26, standard library HTTP
- **Frontend:** Svelte 5
- **Database:** PostgreSQL 18
- **Deployment:** Docker Compose

## Development

### Prerequisites

- Go 1.26+
- Node.js 22+
- PostgreSQL 18 (or use Docker Compose)

### Run with Docker Compose

```bash
docker compose up --build
```

The app will be available at http://localhost:8080.

### Run in development mode

Start the database:

```bash
docker compose up db
```

Start the backend:

```bash
cd backend
go run ./cmd/server
```

Start the frontend (in a separate terminal):

```bash
cd frontend
npm run dev
```

The frontend dev server runs at http://localhost:5173 and proxies API calls to the Go backend at http://localhost:8080.

## Project Structure

```
kanbanboard/
├── backend/           # Go backend
│   ├── cmd/server/    # Entry point
│   ├── internal/      # Application packages
│   └── migrations/    # SQL migration files
├── frontend/          # Svelte frontend
│   └── src/           # Source code
├── docs/
│   ├── plan/          # Planning documents
│   └── skills/        # Claude Code skills
└── docker-compose.yml
```
