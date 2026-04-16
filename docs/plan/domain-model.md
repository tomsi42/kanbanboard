# Domain Model

## Entities (9)

| Entity | Fields |
|---|---|
| **User** | name, email (optional for agent users), username (optional, unique, 3–20 chars), credentials, roles (admin, team manager), state (active/inactive/deleted) |
| **Team** | name, owner (user with team manager role), members (users) |
| **Project** | name, owner (user or team), visibility (public/private), tag (unique, 2-4 uppercase letters), next task number (counter) |
| **Column** | name, position (within project) |
| **Task** | title, description, column, label (single), assignee (user), creator (user), parent task (optional, max depth 2), target version, priority (none/low/medium/high/critical), due date, task number (sequential within project), blockedBy (derived — list of task IDs from TaskDependency) |
| **Label** | name, color (within project) |
| **Comment** | text, author (user), timestamp (on task) |
| **TaskDependency** | blockerTaskId → blockedTaskId (many-to-many, within same project) |
| **ApiToken** | name, user, tokenHash, lastUsedAt, createdAt, expiresAt (optional) |

## Defaults on project creation

- **Columns:** Inbox, Todo, In Progress, Blocked, Done
- **Labels:** task (cyan), bug (red), feature (green), chore (grey)

## Validation rules

### Password policy
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character (e.g. !@#$%^&*()-_=+[]{}|;':",.<>?/`~)

### Username (for agent users)
- 3–20 characters
- Lowercase letters, numbers, and hyphens only (no spaces)
- Unique across all users
- Optional for human users; used in place of email for agent accounts
- Login accepts either email or username as the identifier

### Project tag
- 2-4 uppercase letters only (A-Z)
- Unique across all projects
- Required at project creation
- Immutable once the project has tasks

### Other input validation
- Email: valid email format, unique per user
- User name: required, non-empty
- Project name: required, non-empty
- Project tag: required, 2-4 uppercase letters, unique
- Task title: required, non-empty
- Column name: required, non-empty
- Label name: required, non-empty
- Priority: one of 'none', 'low', 'medium', 'high', 'critical'
- Visibility: one of 'public', 'private'

## Key design decisions

- All work items are Tasks - no separate Bug/Feature/Subtask classes
- Tasks support up to 2 levels of nesting: task → subtask → sub-subtask. Sub-subtasks cannot have children.
- Subtasks and sub-subtasks move independently in columns. Moving a parent cascades to all descendants.
- Task dependencies (blocked-by) are within the same project only. Cycles are rejected.
- A task is considered "actively blocked" when it has ≥1 blocker that is not in the Done column.
- Single label per task (not multiple)
- Labels are project-scoped - same text in different projects are independent
- Priority is a field on Task, not a label
- Columns must be defined before tasks are added
- Task assignee defaults to owner for personal projects, unassigned for team projects
- Creator and assignee are separate fields
- Task numbers are sequential per project, assigned atomically, never reused
- Project tag is editable only while the project has zero tasks
- User deletion is soft — record preserved with name for historical references
- Three user states: active (can log in), inactive (reversible, cannot log in), deleted (permanent, cannot log in)
- Deleting a user cascades: owned projects deleted, teams transferred, tasks unassigned
- Agent users authenticate via API tokens (Bearer header) — no cookie session required
- API tokens are long-lived, named, revocable. Token value shown once on creation; only hash stored.

## Napkin diagram

```
User ──belongs to──▶ Team
 │                    │
 owns                 owns
 ▼                    ▼
Project ──has──▶ Column ──has──▶ Task ──parent──▶ Task ──parent──▶ Task (max depth 2)
 │                                │
 has                          blocks/blocked-by (TaskDependency, same project)
 ▼                                │
Label ◀──tagged on────────────────┤
                                  has
                                  ▼
                               Comment

User ──has──▶ ApiToken
```
