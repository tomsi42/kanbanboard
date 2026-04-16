# User Stories

## User types

One user type with roles:
- Every user can create projects and tasks
- **Team manager** role: can create teams and manage membership of teams you own
- **Administrator** role: can create/manage user accounts, system-level settings
- Roles can overlap (a user can be both admin and team manager)
- No self-registration - admin creates accounts

## Personas

- **Tom, Developer & Admin** — built the app because Jira is too complicated for small teams. Uses it for personal and team projects. Creates accounts and assigns roles. Now deploying internally and setting up AI agents.
- **Arne, Scrum Master / Team Owner** — manages a small dev team (Tom, Kåre, Siri). Wants at-a-glance status. Creates team projects. Assigns tasks to agents alongside human team members.
- **Kåre, Developer** — on Arne's team, also has private projects. Uses private visibility.
- **Siri, Developer** — on Arne's team. Day-to-day user — picks up tasks, subtasks, comments.
- **Claude, AI Agent** — assigned tasks by Arne or Tom. Needs to discover tasks assigned to it, read task details and blockers, and update task status when work is complete. Accesses the board programmatically via MCP tools.

## Must have (v1.0)

1. **As a user, I want to create a project with customizable columns** so that I can organize work the way I prefer. *(Default columns: Inbox, Todo, In Progress, Blocked, Done)*
2. **As a user, I want to create tasks and move them between columns** so that I can track progress.
3. **As a user, I want to add subtasks to a task** so that I can break work into smaller pieces. *(Subtasks appear and move independently in columns.)*
4. **As a user, I want to label tasks** (single label per task, from project-scoped labels) so that I can categorize and filter my work. *(Default labels: bug, feature, chore)*
5. **As a user, I want to log in and manage my profile** so that my work is secure and personal.
6. **As a team manager, I want to create teams and manage members** so that my team can collaborate on shared projects.
7. **As a user, I want to control project visibility** - public (everyone can view, only owner edits) or private (only owner views and edits). Owner is a user or team.
8. **As an administrator, I want to create and manage user accounts** so that I control who has access.

## Should have (v1.0)

9. **As an administrator, I want to assign roles** (team manager, administrator) so that I can delegate responsibilities.

## v1.1 Stories

10. **As a user, I want tasks to have a sequential number with a project tag** (e.g. KB-7) so that I can reference tasks in branch names and commits. *(Tag: 2-4 uppercase letters, unique, immutable after first task. Numbers never reused.)*
11. **As Arne, I want to see at a glance who is assigned to each task** so that I know what everyone is working on at standup. *(Assignee initials displayed on task cards.)*
12. **As a user, I want task cards to be visually tinted by their label colour** so that I can quickly scan the board and see the mix of work types.
13. **As a user, I want subtask cards to show their parent task name** so that I can see the relationship when subtasks are in different columns.
14. **As Kåre, I want to see at a glance whether I'm on a team, personal, or private board** so that I don't accidentally create tasks in the wrong project. *(Column backgrounds tinted by board type.)*

## v1.2 Stories

15. **As Tom, I want to delete projects** so that I can clean up dummy and unused projects. *(Owner only. Confirmation dialog showing impact — number of tasks that will be deleted.)*
16. **As Tom, I want to search for tasks across all projects** so that I can quickly find a task by number or title without scanning every board. *(Search by task number and/or title. Results in right-side pane. Respects visibility — only shows tasks in projects the user can see. Clicking a result switches to that project and opens the task.)*
17. **As Tom (admin), I want to delete user accounts** so that I can clean up when someone leaves. *(Soft delete — user record preserved with name for history. Owned projects deleted with confirmation. Teams transferred to another member or admin. Tasks unassigned. Three user states: active, inactive, deleted.)*

## Acceptance criteria (v1.0)

| # | Done when |
|---|-----------|
| 1 | Create project, add/remove/reorder columns. Default columns added on creation. Owner can edit column names. |
| 2 | Tasks on board with title and description. Move between columns (drag and drop). |
| 3 | Subtask linked to parent. Appears in column independently. Moves independently. |
| 4 | Task has single label from project's label set. Can filter board by label. Default labels on project creation. |
| 5 | Login, logout. Sessions persist. User can edit own profile. Password policy: min 8 chars, at least one letter and one number. |
| 6 | Create team, add/remove members. Only team manager who owns the team can manage it. |
| 7 | Default public. Toggle to private. Non-owners see public projects read-only. |
| 8 | Admin creates users with name/email/password. Admin can deactivate users. Password policy enforced on user creation. |
| 9 | Admin assigns/removes roles. Users can have multiple roles. |

## Acceptance criteria (v1.1)

| # | Done when |
|---|-----------|
| 10 | Project has a unique tag (2-4 uppercase letters). Tasks auto-numbered sequentially. Number displayed on cards, detail panel, and API. Tag editable only when project has zero tasks. Deleted task numbers not reused. |
| 11 | Assignee initials shown on task cards when assigned. No indicator when unassigned. |
| 12 | Task card background uses a light tint of the label colour. Task label default colour is cyan (#0891b2). |
| 13 | Subtask cards show parent name above title. ↳ prefix before subtask title. Parent cards show ▤ icon. |
| 14 | Board context: icons in header (👤/👥/🔒), subtle board background tinting by project type. |

## Acceptance criteria (v1.2)

| # | Done when |
|---|-----------|
| 15 | Owner can delete a project from settings. Confirmation shows task count. All columns, labels, tasks, and comments cascade-deleted. |
| 16 | Search icon in header. Search by task number or title (case-insensitive). Results pane shows task number, title, project name, label. Respects visibility. Clicking a result switches project and opens task detail. Mutually exclusive with task detail panel. |
| 17 | Admin can delete a user. Confirmation shows impact (projects to delete, teams to transfer). Owned projects deleted (cascade). Teams transferred to member or admin. Tasks unassigned. User marked as deleted (permanent). Name preserved on historical tasks and comments. |

## v1.3 Stories

18. **As a user, I want the system to enforce a stronger password policy** so that accounts are harder to compromise. *(Requires: uppercase letter, lowercase letter, number, special character, minimum 8 characters.)*
19. **As Siri, I want to see a blocked indicator on task cards** so that I immediately know a task cannot be started yet without opening it. *(Red circle on card when the task has one or more active blockers — i.e. blocking tasks not yet in Done.)*
20. **As Arne, I want to mark a task as blocked by another task** so the team knows which tasks must be completed before others can begin. *(Within the same project. Blockers listed by task number and title in the task detail panel. Clickable to navigate to blocker.)*
21. **As a user, I want to create subtasks of subtasks** so I can break complex work into finer-grained steps. *(Maximum depth 2: task → subtask → sub-subtask. Sub-subtasks cannot have children.)*
22. **As Tom (admin), I want to create agent user accounts with a short username** so that AI agents can be identified distinctly from human users and log in without a real email address. *(Username: 3–20 chars, lowercase letters/numbers/hyphens. Unique. Login accepts username or email.)*
23. **As Tom (admin), I want to issue API tokens for agent users** so that agents can authenticate programmatically without cookie-based sessions that expire and require re-login. *(Named tokens. Shown once on creation. Admin can revoke. Accepted as `Authorization: Bearer <token>`.)*
24. **As Claude (AI agent), I want to access kanban tasks via MCP tools** so I can discover tasks assigned to me, read their details and blockers, and hand off completed work to the next agent in the pipeline without manual coordination. *(Tools: list my tasks, get task, create task, handoff task, add comment.)*
25. **As Tom, I want task priority to include Critical** so that bugs and tasks requiring immediate attention can be distinguished from High priority work. *(Priority order: Critical > High > Medium > Low > None.)*
26. **As Tom, I want Claude to generate bug report tasks from a codebase review** so that found bugs are tracked on the board with proper priorities and fix subtasks. *(Claude creates tasks via the REST API in Tom's session. Each bug = one task with priority critical/high/medium/low. Each task gets subtasks: investigate, fix, test.)*

## Acceptance criteria (v1.3)

| # | Done when |
|---|-----------|
| 18 | Password requires: 8+ chars, at least one uppercase, one lowercase, one number, one special character (!@#$%^&\* etc.). Clear error messages per missing requirement. Enforced at setup, admin user creation, self-registration, and password change. Existing users not forced to change unless they change their password. |
| 19 | Red filled circle (●) shown on task card when task has ≥1 active blocker (blocking task is not in the Done column). No indicator when no blockers or all blockers are done. |
| 20 | Task detail panel has "Blocked by" section. Can search for a blocker task by number or title within the same project. Blockers listed by task number + title, each with a remove button. Clicking a blocker navigates to it. Adding a task as its own blocker is rejected. Adding a task that would create a cycle is rejected. |
| 21 | Subtask detail panel shows "+ Add subtask" and the subtasks section. Sub-subtask detail panel does not show the subtasks section. Creating a child of a sub-subtask is rejected with an error. Moving a task cascades to all descendants (not just direct children). Sub-subtask cards show the parent subtask name. |
| 22 | Admin can set/edit username when creating or editing a user. Username format: 3–20 chars, lowercase letters/numbers/hyphens, no spaces. Unique across all users. Login form accepts username or email in the identifier field. Agent users can have no email (username is their only identifier). |
| 23 | Admin can create named API tokens for any user from the admin user page. Token value shown exactly once in a copy-prompt on creation; only the hash stored. Admin can revoke tokens. Tokens have no expiry by default. All API endpoints accept `Authorization: Bearer <token>` as an alternative to the session cookie. |
| 24 | MCP server deployed as a separate binary alongside the app. Configured with an API token and the board's base URL. Exposes tools: `list_my_tasks` (tasks assigned to the token's user, includes enough context to prioritise work), `get_task` (by project tag + number, includes subtasks, blockers with their column status), `create_task` (with optional parent for subtasks), `handoff_task` (move to named column + reassign to named user atomically), `add_comment`. Tool responses use task number references (e.g. KB-3) throughout. |
| 25 | Priority field supports Critical, High, Medium, Low, None. Critical shown in UI with distinct colour. Validation enforces the expanded set at all entry points. |
| 26 | No specific UI story — this is a workflow: Tom opens Claude Code, asks it to review a codebase for bugs, Claude creates tasks (priority = critical/high/medium/low) via the REST API using Tom's session. Each bug task gets subtasks: Investigate, Fix, Test. Done when tasks appear correctly on the board. |
