# User Stories

## User types

One user type with roles:
- Every user can create projects and tasks
- **Team manager** role: can create teams and manage membership of teams you own
- **Administrator** role: can create/manage user accounts, system-level settings
- Roles can overlap (a user can be both admin and team manager)
- No self-registration - admin creates accounts

## Must have

1. **As a user, I want to create a project with customizable columns** so that I can organize work the way I prefer. *(Default columns: Inbox, Todo, In Progress, Blocked, Done)*
2. **As a user, I want to create tasks and move them between columns** so that I can track progress.
3. **As a user, I want to add subtasks to a task** so that I can break work into smaller pieces. *(Subtasks appear and move independently in columns.)*
4. **As a user, I want to label tasks** (single label per task, from project-scoped labels) so that I can categorize and filter my work. *(Default labels: bug, feature, chore)*
5. **As a user, I want to log in and manage my profile** so that my work is secure and personal.
6. **As a team manager, I want to create teams and manage members** so that my team can collaborate on shared projects.
7. **As a user, I want to control project visibility** - public (everyone can view, only owner edits) or private (only owner views and edits). Owner is a user or team.
8. **As an administrator, I want to create and manage user accounts** so that I control who has access.

## Should have

9. **As an administrator, I want to assign roles** (team manager, administrator) so that I can delegate responsibilities.

## Acceptance criteria

| # | Done when |
|---|-----------|
| 1 | Create project, add/remove/reorder columns. Default columns added on creation. Owner can edit column names. |
| 2 | Tasks on board with title and description. Move between columns (drag and drop). |
| 3 | Subtask linked to parent. Appears in column independently. Moves independently. |
| 4 | Task has single label from project's label set. Can filter board by label. Default labels on project creation. |
| 5 | Login, logout. Sessions persist. User can edit own profile. |
| 6 | Create team, add/remove members. Only team manager who owns the team can manage it. |
| 7 | Default public. Toggle to private. Non-owners see public projects read-only. |
| 8 | Admin creates users with name/email/password. Admin can deactivate users. |
| 9 | Admin assigns/removes roles. Users can have multiple roles. |
