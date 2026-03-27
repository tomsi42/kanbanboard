# Domain Model

## Entities (7)

| Entity | Fields |
|---|---|
| **User** | name, email, credentials, roles (admin, team manager) |
| **Team** | name, owner (user with team manager role), members (users) |
| **Project** | name, owner (user or team), visibility (public/private) |
| **Column** | name, position (within project) |
| **Task** | title, description, column, label (single), assignee (user), creator (user), parent task (optional), target version, priority, due date |
| **Label** | name, color (within project) |
| **Comment** | text, author (user), timestamp (on task) |

## Defaults on project creation

- **Columns:** Inbox, Todo, In Progress, Blocked, Done
- **Labels:** bug, feature, chore (each with distinct color)

## Key design decisions

- All work items are Tasks - no separate Bug/Feature/Subtask classes
- Subtasks are Tasks with a parent reference, move independently in columns
- Single label per task (not multiple)
- Labels are project-scoped - same text in different projects are independent
- Priority is a field on Task, not a label
- Columns must be defined before tasks are added
- Task assignee defaults to owner for personal projects, unassigned for team projects
- Creator and assignee are separate fields

## Napkin diagram

```
User ──belongs to──▶ Team
 │                    │
 owns                 owns
 ▼                    ▼
Project ──has──▶ Column ──has──▶ Task ──parent──▶ Task
 │                                │
 has                              has
 ▼                                ▼
Label ◀──tagged on──────────── Comment
```
