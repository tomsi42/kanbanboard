# Implementation Plan

## Phase 1: Foundation (v0.1.x)

| Sub-phase | Description | Version |
|---|---|---|
| 1.1 | Project setup - Go backend, Svelte frontend, PostgreSQL, Docker Compose, project structure | v0.1.1 |
| 1.2 | Database schema and migrations (all entities, hand-rolled runner) | v0.1.2 |
| 1.3 | Onboarding - first-time setup screen (admin account, application title) | v0.1.3 |
| 1.4 | Authentication - login, logout, sessions | v0.1.4 |

## Phase 2: Core Board (v0.2.x)

| Sub-phase | Description | Version |
|---|---|---|
| 2.1 | Project CRUD - create project with default columns and labels | v0.2.1 |
| 2.2 | Board view - render columns and task cards | v0.2.2 |
| 2.3 | Task CRUD - create, edit, move tasks. Side panel for detail. | v0.2.3 |
| 2.4 | Drag and drop between columns | v0.2.4 |
| 2.5 | Labels - assign to task, filter board by label | v0.2.5 |

## Phase 3: Customization (v0.3.x)

| Sub-phase | Description | Version |
|---|---|---|
| 3.1 | Profile editing | v0.3.1 |
| 3.2 | Project settings - edit columns, labels, visibility | v0.3.2 |

## Phase 4: Subtasks and Comments (v0.4.x)

| Sub-phase | Description | Version |
|---|---|---|
| 4.1 | Subtasks - create, show independently in columns with indicator | v0.4.1 |
| 4.2 | Subtask progress on parent card, warning on parent to Done | v0.4.2 |
| 4.3 | Comments - add, edit, delete own | v0.4.3 |

## Phase 5: Teams and Collaboration (v0.5.x)

| Sub-phase | Description | Version |
|---|---|---|
| 5.1 | Admin area - user management (create, edit, deactivate, assign roles) | v0.5.1 |
| 5.2 | Team management - create teams, add/remove members | v0.5.2 |
| 5.3 | Team project ownership - all members can edit | v0.5.3 |
| 5.4 | Project visibility - public/private, read-only for non-owners | v0.5.4 |

## Phase 6: Documentation and Release (v0.6.x)

| Sub-phase | Description | Version |
|---|---|---|
| 6.1 | User guide | v0.6.1 |
| 6.2 | README and LICENSE | v0.6.2 |
| 6.3 | Final review | v0.6.3 |
| - | Release | v1.0.0 |

## v1.0.1: Backend Tests and Skills

| Description | Version |
|---|---|
| Unit tests (handler authorization), integration tests (store layer), changelog, debrief skill | v1.0.1 |

## v1.1: Task Numbering and Visual Improvements

| Phase | Description | Version |
|---|---|---|
| 1 | Task numbering — backend, migration (add tag/number fields, backfill), API, tests | v1.1-snapshot-1 |
| 2 | Card visual improvements — label tinting, task number display, parent/subtask indicators, assignee initials | v1.1-snapshot-2 |
| 3 | Board context colours — column backgrounds by project type (blue/green/amber) | v1.1-snapshot-3 |
| 4 | API documentation + release | v1.1.0 |

## v1.1.2: Code Health Fixes

| Phase | Description | Version |
|---|---|---|
| 1 | Security & data integrity — authorization on column/label handlers, priority validation, transactional task creation | v1.1.2-snapshot-1 |
| 2 | Error handling & code quality — writeJSON helper, applyTaskUpdates, duplicate email 409, ListActiveUsersBasic, requireTeamOwner, session constant | v1.1.2-snapshot-2 |
| 3 | Tests & docs — HTTP-level handler tests, doc fixes, code-health made non-optional | v1.1.2-snapshot-3 |

## v1.2: Delete Projects, User Deletion, Task Search

| Phase | Description | Version |
|---|---|---|
| 1 | Delete projects — backend endpoint, frontend confirmation dialog, cascade tests | v1.2-snapshot-1 |
| 2 | User deletion — soft delete, cascade projects, transfer teams, unassign tasks, admin UI with impact confirmation | v1.2-snapshot-2 |
| 3 | Cross-project task search — search endpoint, visibility-respecting query, search UI with results pane | v1.2-snapshot-3 |
| 4 | Pre-release gate + release | v1.2.0 |

## v1.3: Passwords, Dependencies, Sub-subtasks, Agent Access, MCP

### Scope

| Feature | Stories |
|---|---|
| Stricter password policy | #18 |
| Critical priority | #25 |
| Task blocked-by dependencies | #19, #20 |
| Sub-subtasks (max depth 2) | #21 |
| Agent usernames + login | #22 |
| API tokens for programmatic access | #23 |
| MCP server + bug report workflow | #24, #26 |

### Minor vs major

v1.3 is a **minor version**. All features are additive: new tables, new fields, new endpoints, new binary. No existing behavior changes. Existing users are unaffected unless they change their password (at which point the stricter policy applies). API is backwards-compatible throughout.

### Impact analysis

**Password policy (#18) + Critical priority (#25)**
- `backend/internal/validate/validate.go` — update `Password()` and `Priority()` functions
- `backend/internal/validate/validate_test.go` — update test cases
- DB: tasks table `priority` check constraint extended to include `'critical'`; new migration
- Frontend: priority dropdowns updated (Critical option, distinct colour — e.g. deep red)
- `docs/plan/domain-model.md` — already updated

**Task dependencies (#19, #20)**
- New migration: `task_dependencies (blocker_task_id UUID, blocked_task_id UUID, PRIMARY KEY (blocker_task_id, blocked_task_id), FOREIGN KEY both → tasks ON DELETE CASCADE)`
- Index: `idx_task_dependencies_blocked` on `blocked_task_id`
- `model.go`: no new type (dependency is a join table). `Task` gets `BlockedBy []string` (populated at query time, omitempty).
- New store functions: `AddBlocker`, `RemoveBlocker`, `ListBlockersForProject` (all dependencies for a project in one query)
- `ListTasksForProject` extended to hydrate `BlockedBy` per task via a single extra query
- New handlers: `HandleAddBlocker` (`POST /api/v1/projects/{id}/tasks/{id}/blockers`), `HandleRemoveBlocker` (`DELETE /api/v1/projects/{id}/tasks/{id}/blockers/{blockerId}`)
- Validation in `AddBlocker`: reject self-reference, reject if either task is not in the same project, reject if adding would create a cycle (simple DFS)
- Frontend `TaskCard.svelte`: derive `isBlocked` from `task.blockedBy` and the done column ID. Show red ● in `card-top-left` when actively blocked.
- Frontend `TaskDetailPanel.svelte`: "Blocked by" section with task search (within project), list of blockers with remove button, each blocker navigable via `onTaskSelect`.
- Regression: moving a task to Done should not auto-remove dependencies (that's intentional — the dependency exists even after completion).

**Sub-subtasks (#21)**
- No DB migration needed (self-referential FK already in place).
- `HandleCreateTask`: fetch parent task depth before inserting. Depth = parent has no parent (depth 1) or parent has a parent (depth 2). If depth would exceed 2, return 400.
- `store.MoveTask`: current logic cascades subtasks in the same column. Must extend to cascade all descendants recursively (using a recursive CTE or explicit two-level update).
- Frontend `TaskDetailPanel.svelte`: remove `{#if !task.parentTaskId}` guard on subtasks section. Replace with computed `depth` (0, 1, or 2 derived from `project.tasks`). Show subtasks section when `depth < 2`. Pass `depth` context to "add subtask" button logic.
- `TaskCard.svelte`: no change needed — existing parent name display already works for sub-subtasks.
- New store test: `CreateTask` at depth 3 returns error. `MoveTask` with two-level subtask tree cascades correctly.

**Agent usernames (#22)**
- New migration: add `username TEXT UNIQUE` column to `users` (nullable — existing users have none).
- Index: `idx_users_username` on `username`.
- `store.GetUserByUsername(db, username)` — new function, excludes deleted users.
- `HandleLogin`: detect identifier type (contains `@` → email lookup, else → username lookup). Return same error as before to avoid enumeration.
- `store.CreateUser` / `UpdateUserAdmin`: accept username field.
- Admin UI: username field in create/edit user forms (optional for human users).
- `validate.Username()`: 3–20 chars, lowercase letters/numbers/hyphens only.
- Registration form: no username field (self-registered users use email only).
- Note: agent users may have no email — relax the NOT NULL constraint on `email` (migration needed). Uniqueness constraint on email: change to `UNIQUE WHERE email IS NOT NULL` (partial index).

**API tokens (#23)**
- New migration: `api_tokens (id UUID PK, user_id UUID FK, name TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE, last_used_at TIMESTAMPTZ, created_at TIMESTAMPTZ, expires_at TIMESTAMPTZ)`
- Index: `idx_api_tokens_user_id`.
- New store functions: `CreateApiToken`, `GetApiTokenByHash`, `ListApiTokensForUser`, `DeleteApiToken`
- Token generation: `crypto/rand` 32-byte token, base64url-encoded. Store SHA-256 hash only.
- `middleware.RequireAuth`: check `Authorization: Bearer <token>` header first, then fall back to session cookie. On token match, update `last_used_at`.
- Admin UI: "API Tokens" sub-section on user detail page. Create token (name input → shows value once in a modal). List tokens (name, created, last used). Revoke button.
- New handlers: `HandleCreateApiToken`, `HandleListApiTokens`, `HandleDeleteApiToken` under `/api/v1/admin/users/{userId}/tokens`.

**MCP server + bug report workflow (#24, #26)**

Agent pipeline model: agents poll for tasks assigned to them → pick one → work → hand off to the next agent (move column + reassign) → repeat. Planning and bug-report task creation happens in Tom's Claude Code session via the REST API directly (no MCP needed for creation by humans).

- New binary: `backend/cmd/mcp/main.go` — standalone MCP server.
- Configuration: `KANBAN_API_URL` and `KANBAN_API_TOKEN` environment variables.
- Transport: MCP over stdio (standard for local agent use; HTTP+SSE can be added later).
- Tools exposed:
  - `list_my_tasks` — tasks assigned to the token's user across visible projects. Returns: task ref (e.g. KB-3), title, description, column, priority, due date, active blocker refs. Enough context to decide what to work on without a follow-up call.
  - `get_task` — full task detail by project tag + task number. Includes subtasks (ref + title + column), blockers (ref + title + column — agent can see if blockers are done), comments.
  - `create_task` — create a task in a project. Accepts: project tag, title, description, priority, column name (default: first column), parent task ref (optional, for subtasks). Returns the new task ref. Used by agents doing bug-triage or decomposition work.
  - `handoff_task` — the core pipeline operation. Atomically: move task to a named column + reassign to a named user (username). E.g. coder agent hands off to tester: `handoff_task(ref="KB-3", column="Testing", assignTo="claude-test")`. Returns updated task.
  - `add_comment` — add a comment to a task (progress, findings, blockers encountered).
- Uses `github.com/mark3labs/mcp-go` or equivalent MCP SDK for Go.
- New REST endpoints:
  - `GET /api/v1/tasks/mine` — tasks assigned to the authenticated user, across all visible projects.
  - `POST /api/v1/projects/{id}/tasks/{id}/handoff` — move + reassign in one transaction (also usable directly from the REST API).
- Dockerfile: MCP binary built as a separate stage in the existing Dockerfile.
- `docker-compose.deploy.yml`: MCP service entry (optional; can also be run as a CLI tool pointed at the deployed API).

### Regression risks

- Login flow: adding username lookup must not break email login for existing users.
- Subtask move cascade: existing single-level subtask behavior must be preserved.
- Task list API: adding `BlockedBy` field to task responses is additive (omitempty), but frontend must handle absent field gracefully.
- Password validation: existing users with old passwords are not affected until they change passwords.

### Implementation phases

| Phase | Description | Version |
|---|---|---|
| 1 | Password policy + Critical priority — stricter `validate.Password()`, extend `validate.Priority()` with critical, DB migration for check constraint, frontend priority dropdowns, tests | v1.3-snapshot-1 |
| 2 | Task dependencies backend — migration, store (add/remove/list, hydrate in task list), handlers, cycle detection, tests | v1.3-snapshot-2 |
| 3 | Task dependencies frontend — blocked indicator (●) on card, blocked-by section in detail panel | v1.3-snapshot-3 |
| 4 | Sub-subtasks backend — depth enforcement in CreateTask, recursive MoveTask cascade, tests | v1.3-snapshot-4 |
| 5 | Sub-subtasks frontend — depth-aware subtask section in detail panel | v1.3-snapshot-5 |
| 6 | Agent usernames — migration (username + nullable email), validate.Username(), login with username or email, admin UI | v1.3-snapshot-6 |
| 7 | API tokens — migration, store, Bearer auth middleware, admin UI (create/list/revoke) | v1.3-snapshot-7 |
| 8 | MCP server — binary, `list_my_tasks` + `handoff_task` REST endpoints, MCP tools (list, get, create, handoff, comment), Docker wiring | v1.3-snapshot-8 |
| 9 | Pre-release code health gate + release | v1.3.0 |
