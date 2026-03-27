# User Guide

## Getting Started

### First-Time Setup

When the application is launched for the first time, you'll see the onboarding screen. Enter:

- **Application title** — the name shown in the header (e.g. "My Kanban Board")
- **Your name, email, and password** — this creates the administrator account

After completing setup, you're logged in automatically.

### Logging In

Go to the application URL and enter your email and password. If you don't have an account, ask your administrator to create one for you.

---

## Using the Board

### Creating a Project

1. Click the project dropdown in the top-left of the header
2. Select **+ New Project** at the bottom of the list
3. Enter a project name
4. Choose the owner: **Personal** (just you) or a **team** you manage
5. Click **Create Project**

The project is created with five default columns (Inbox, Todo, In Progress, Blocked, Done) and four default labels (task, bug, feature, chore).

### Adding Tasks

1. Click **+ Add Task** in the header
2. Type a task title and press **Enter** or click **Add**
3. The task appears in the first column (Inbox) with the "task" label

### Moving Tasks

**Drag and drop:** Grab a task card and drag it to another column. The task moves to the end of the target column.

**Via task detail:** Click a task to open the detail panel. Change the **Column** dropdown to move it.

### Task Details

Click any task card to open the detail panel on the right side. You can edit:

- **Title** — click to edit, saves when you click away
- **Description** — free text, saves when you click away
- **Column** — move the task to a different column
- **Label** — categorize the task (bug, feature, chore, task, or custom labels)
- **Assignee** — who's working on it (shown for team projects with multiple members)
- **Priority** — none, low, medium, or high
- **Due date** — when the task is due
- **Target version** — which release this is planned for

All changes save automatically.

### Subtasks

Subtasks let you break a task into smaller pieces.

1. Open a task's detail panel
2. Scroll to the **Subtasks** section
3. Click **+ Add subtask** and enter a title

Subtasks appear as independent cards on the board with a "↳ subtask" indicator. They move independently between columns.

**Progress tracking:** The parent task card shows subtask progress (e.g. "2/5"). When all subtasks reach the last column, the badge turns green.

**Done warning:** If you try to move a parent task to the last column (Done) while subtasks aren't all done, you'll see a confirmation dialog.

**Navigation:** In a subtask's detail panel, click the parent task name to navigate back. In a parent's detail panel, click a subtask to open it.

### Comments

1. Open a task's detail panel
2. Scroll to the **Comments** section
3. Type your comment and press **Enter** or click **Comment**

You can **edit** or **delete** your own comments using the links next to each comment.

### Filtering by Label

Use the filter dropdown in the header (next to the Add Task button) to show only tasks with a specific label. Select **All** to clear the filter. Column counts update to reflect the filtered view.

### Switching Projects

Click the project dropdown in the header to see all your projects. Click a project name to switch to it.

---

## For Team Managers

### Creating a Team

1. Click your name in the top-right to open the user menu
2. Select **My Teams**
3. Click **+ Create Team** and enter a team name

### Managing Team Members

1. Go to **My Teams**
2. Click a team name to expand it
3. **Add a member:** Select a user from the dropdown and click **Add**
4. **Remove a member:** Click the **✕** next to a member's name

### Creating a Team Project

1. Click the project dropdown → **+ New Project**
2. In the **Owner** dropdown, select your team instead of "Personal"
3. Click **Create Project**

All team members can view and edit tasks on team projects. Only the team owner (you) can access project settings.

### Assigning Tasks

On team projects, the task detail panel shows an **Assignee** dropdown with all team members. Select who should work on the task.

---

## For Administrators

### Creating Users

1. Click your name → **Admin**
2. Click **+ Create User**
3. Enter name, email, and password
4. Optionally check **Admin** or **Team Manager** roles
5. Click **Create**

Password requirements: minimum 8 characters, at least one letter and one number.

### Managing Users

In the Admin page, the user table shows all accounts. For each user you can:

- **Edit** — change name, email, toggle roles (Admin, Team Manager), activate/deactivate
- **Reset PW** — set a new password for the user (no current password needed)

### Deactivating Users

Edit a user and uncheck the **Active** checkbox. Deactivated users cannot log in.

### Roles

- **Admin** — can access the Admin area to manage all user accounts
- **Team Manager** — can create teams and manage team membership

A user can have both roles. Users without any role can create personal projects and work on team projects they're members of.

---

## Project Settings

Only the project owner can access settings. Click the **⚙** gear icon next to the project dropdown.

### Project Name and Visibility

- **Name** — edit and click away to save
- **Visibility:**
  - **Public** — everyone can view the board (read-only for non-owners)
  - **Private** — only the owner (or team members for team projects) can see it

### Managing Columns

- **Add:** Type a name and click **Add** at the bottom of the column list
- **Rename:** Edit the column name and click away
- **Reorder:** Use the **▲** and **▼** buttons
- **Delete:** Click **✕** (only works if the column has no tasks)

### Managing Labels

- **Add:** Type a name, pick a color, and click **Add**
- **Edit:** Change the name or color inline
- **Delete:** Click **✕** (only works if no tasks use the label)

---

## Profile

Click your name → **My Profile** to edit your account:

- **Name and email** — edit and click **Save Profile**
- **Change password** — enter your current password, then the new password twice. Click **Change Password**
