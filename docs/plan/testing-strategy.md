# Testing Strategy

## Approach

Backend tested automatically with Go's built-in testing. Frontend and user workflows tested manually by the user with a checklist after each sub-phase.

## Test levels

| What | How | When |
|---|---|---|
| Business logic (roles, visibility, subtask rules) | `go test` unit tests | Written during development |
| Database queries | `go test` integration tests against PostgreSQL | Written during development |
| API endpoints | `go test` with HTTP test helpers | Written during development |
| Frontend components | Manual testing by user | After each sub-phase |
| User workflows | Manual testing by user with checklist | After each sub-phase |

## Risk assessment

| Layer | Risk | Rationale |
|---|---|---|
| Business logic (role checks, visibility, subtask progress) | **High** | Bugs here break authorization and core behavior |
| Database queries | **Medium** | Wrong queries = wrong data |
| API endpoints | **Medium** | Contract between frontend and backend |
| Svelte components | **Low** | Visual, caught during manual review |
| Full user workflows | **Medium** | Caught during manual review |

## Conventions

- **Location:** `backend/internal/{package}/*_test.go` (Go convention - tests next to source)
- **Naming:** `TestFunctionName_scenario` (e.g. `TestCreateProject_defaultColumns`)
- **Run:** `go test ./...` from the backend directory
- **Gate:** All tests must pass before a sub-phase is presented for acceptance

## Manual test checklists

After each sub-phase, a manual test checklist is provided to the user covering:
- Frontend behavior and appearance
- User workflows relevant to the sub-phase
- Edge cases to verify

## What is explicitly NOT tested automatically

- Svelte components (manual review is sufficient for the project scale)
- Drag-and-drop interactions (complex to automate, easy to verify manually)
- Visual styling and layout (caught during manual review)
