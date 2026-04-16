# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [1.3-snapshot-1] - 2026-04-16

### Added
- Critical priority level for tasks (critical > high > medium > low > none)
- Priority dropdown highlights critical tasks in red

### Changed
- Password policy strengthened: now requires uppercase letter, lowercase letter, number, and special character (minimum 8 characters)

## [1.2.2] - 2026-04-08

### Added
- Self-serve user registration — admin can enable/disable open registration from Admin settings
- Registration toggle in Admin page (Settings section)
- Create Account form on the login page, shown only when registration is open
- `POST /api/v1/auth/register` endpoint (public, gated by registration setting, auto-logs in)
- `GET /api/v1/admin/settings` and `PUT /api/v1/admin/settings` endpoints
- Deploy packaging scripts — `scripts/package.sh` builds and bundles the app image for offline deployment; `deploy.sh` and `docker-compose.deploy.yml` for the target machine

## [1.2.1] - 2026-03-31

### Added
- App title and version displayed in footer on all authenticated pages
- Version injected at build time via Go ldflags and Docker build arg

### Changed
- Board layout uses viewport-constrained flex layout (footer always visible)

## [1.2.0] - 2026-03-30

### Added
- Delete projects from project settings with cascade confirmation dialog
- User deletion (soft delete) — admin can delete users with impact preview, owned projects cascade-deleted, teams transferred, tasks unassigned, name preserved for history
- Cross-project task search by title or task number (e.g. KB-7) with visibility-respecting results panel
- Search button in header, results panel with debounced search-as-you-type
- Delete user impact preview endpoint
- API documentation for all v1.2 endpoints

### Changed
- User deletion runs in a single database transaction (all-or-nothing cascade)
- Shared team transfer logic extracted into `ResolveNewTeamOwner` helper
- Admin user list sorts deleted users to bottom
- Deleted users shown greyed out with "Deleted" badge in admin
- Email cleared on user deletion to allow reuse
- Project list refreshes after user deletion

## [1.1.2] - 2026-03-30

### Fixed
- Authorization checks added to all column and label handlers (previously any authenticated user could modify any project)
- Priority validation enforced on task updates (must be none/low/medium/high)
- Task creation wrapped in database transaction (prevents race conditions)
- Duplicate email now returns 409 instead of 500 (admin create, setup, profile update)
- Cookie name corrected in API docs (session_token, not session)
- Default labels corrected in domain model docs (4 labels, not 3)
- Architecture docs aligned with actual project structure

### Added
- `writeJSON` helper centralizing response encoding with error logging
- `applyTaskUpdates` function extracted from HandleUpdateTask for clarity
- `requireTeamOwner` helper replacing 5 duplicated ownership checks
- `ListActiveUsersBasic` store function (no longer fetches password hashes for user listings)
- `IsUniqueViolation` store helper for PostgreSQL constraint error detection
- HTTP-level handler integration tests (authorization and validation)
- Code-health skill made non-optional in version-planning and debrief

### Changed
- Session duration in setup uses constant from auth (no longer hardcoded)

## [1.1.1] - 2026-03-30

### Changed
- Subtask icon changed from ↳ to ▶ on task cards
- Parent task name on subtask cards: larger font, more spacing
- Removed ↳ icon from task detail panel

## [1.1.0] - 2026-03-30

### Added
- Task numbering: projects have a unique tag (2-4 uppercase letters), tasks are auto-numbered sequentially (e.g. KB-7). Numbers are never reused.
- Label-tinted task cards: card backgrounds use a light tint of the label colour
- Task number displayed on cards and in the task detail panel
- Parent task indicator (▤) on cards with subtasks
- Subtask cards show parent name above title with ↳ prefix
- Assignee initials displayed on task cards when assigned
- Board context icons in header: 👤 personal, 👥 team, 🔒 private
- Subtle board background tinting by project type
- Tag input with auto-suggest on project creation
- Tag display in project settings (locked after tasks exist)
- API documentation (`docs/api.md`)
- Version-planning skill (`docs/skills/version-planning/`)
- Code-health skill (`docs/skills/code-health/`)
- Personas added to user stories document

### Changed
- Default "task" label colour changed from blue (#4a90d9) to cyan (#0891b2)
- Debrief skill updated with code-health references
- Dev-workflow skill updated with snapshot versioning for post-release development
- Planning documents updated with v1.1 stories and implementation plan

## [1.0.1] - 2026-03-27

### Added
- Unit tests for handler authorization logic (visibility, edit permission, ownership)
- Integration tests for store layer against PostgreSQL (users, sessions, teams, projects, tasks, comments)
- Test infrastructure: test database helpers, seed functions, `-short` flag support
- `backend/TESTING.md` with setup and run instructions
- Debrief skill (`docs/skills/debrief/`)
- This changelog

### Changed
- Extracted authorization decisions into pure functions (`authz.go`) for testability
- Handler functions delegate to pure authorization functions via `resolveTeamContext`
- Updated dev-workflow and planning-checklist skills with changelog and debrief references
- Added validation rules to domain-encoding exit criteria

## [1.0.0] - 2026-03-27

Initial release. Kanban board for individuals and small teams with projects, tasks, subtasks, comments, teams, labels, drag-and-drop, and role-based access control.
