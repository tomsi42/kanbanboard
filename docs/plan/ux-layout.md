# UX Layout

## Landing page

User lands on the board view of their current/last project. No dashboard in between.

## Screen inventory

### 1. Login
- Email + password
- Shown when session is missing/expired

### 2. Onboarding (first-time only)
- Shown when no users exist in the database
- Create admin account (name, email, password)
- Set application title
- Redirects to login after setup

### 3. Board View (main screen)
- Top bar: project dropdown, add task button, filter, user menu
- Columns fill the main area with task cards
- Drag and drop between columns

```
[Project Name ▼]  [+ Add Task]  [Filter ▼]     [User Menu ▼]
 |- Project A                                    |- My Profile
 |- Project B                                    |- My Teams
 |- Team Project C                               |- Admin (if admin)
 +- + New Project                                +- Log out
+---------+---------+---------+---------+---------+
| Inbox   | Todo    | In Prog | Blocked | Done    |
|         |         |         |         |         |
| [Card]  | [Card]  | [Card]  |         | [Card]  |
| [Card]  |         | [~sub]  |         |         |
+---------+---------+---------+---------+---------+
```

### 4. Task Detail (side panel)
- Slides in from the right, board dimmed but visible
- Fields: title, description, column, label, assignee, priority, due date, target version
- Subtasks section: list with add button
- Comments section: list with add, edit own, delete own

```
+------------------------------+------------------+
| Board (dimmed)               | Task Detail      |
|                              |                  |
|                              | Title            |
|                              | Description      |
|                              | Column: [v]      |
|                              | Label: [v]       |
|                              | Assignee: [v]    |
|                              | Priority: [v]    |
|                              | Due date: [...]  |
|                              | Target ver: [...]|
|                              |                  |
|                              | -- Subtasks --   |
|                              | > Subtask 1      |
|                              | > Subtask 2      |
|                              | [+ Add subtask]  |
|                              |                  |
|                              | -- Comments --   |
|                              | Alice: fixed it  |
|                              | [Add comment]    |
+------------------------------+------------------+
```

### 5. Project Settings
- Edit columns: add, remove, rename, reorder
- Edit labels: add, remove, edit name/color
- Visibility: public/private toggle

### 6. Admin Area
- User management: create, edit, deactivate users
- Assign/remove roles

### 7. My Teams
- List teams you own
- Create new team
- Add/remove team members

## Task card on the board

Shows:
- Title
- Label (colored tag)
- Assignee (initials)
- Subtask progress if has subtasks (e.g. "3/5")
- Subtask indicator if it is a subtask

## Key UX decisions

- Subtasks are visually identical to tasks with a small indicator
- Parent card shows subtask progress (e.g. "3/5 done")
- Warning when moving parent to Done if subtasks aren't all done
- New task defaults to leftmost column
- Team task assignee defaults to unassigned
- Personal project task assignee defaults to owner
- Comments: add, edit own, delete own
