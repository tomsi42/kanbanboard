package store

import (
	"testing"

	"kanbanboard/internal/model"
)

func TestCreateTask_positionAutoIncrement(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col := columns[0]

	t1 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 1")
	t2 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 2")
	t3 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 3")

	if t1.Position != 0 {
		t.Errorf("task 1 position = %d, want 0", t1.Position)
	}
	if t2.Position != 1 {
		t.Errorf("task 2 position = %d, want 1", t2.Position)
	}
	if t3.Position != 2 {
		t.Errorf("task 3 position = %d, want 2", t3.Position)
	}
}

func TestCreateTask_numberAutoIncrement(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col := columns[0]

	t1 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 1")
	t2 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 2")
	t3 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 3")

	if t1.TaskNumber != 1 {
		t.Errorf("task 1 number = %d, want 1", t1.TaskNumber)
	}
	if t2.TaskNumber != 2 {
		t.Errorf("task 2 number = %d, want 2", t2.TaskNumber)
	}
	if t3.TaskNumber != 3 {
		t.Errorf("task 3 number = %d, want 3", t3.TaskNumber)
	}
}

func TestCreateTask_numberNeverReused(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col := columns[0]

	t1 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 1")
	t2 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 2")

	// Delete task 2
	if err := DeleteTask(db, t2.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	// Create another task — should get number 3, not reuse 2
	t3 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 3")

	if t1.TaskNumber != 1 {
		t.Errorf("task 1 number = %d, want 1", t1.TaskNumber)
	}
	if t3.TaskNumber != 3 {
		t.Errorf("task 3 number = %d, want 3 (should not reuse deleted number 2)", t3.TaskNumber)
	}
}

func TestGetTask_found(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	created := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Test Task")

	task, err := GetTask(db, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Test Task" {
		t.Errorf("title = %q, want %q", task.Title, "Test Task")
	}
}

func TestGetTask_notFound(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)

	_, err := GetTask(db, "00000000-0000-0000-0000-000000000000")
	if err != ErrTaskNotFound {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestListTasksForProject_orderedByColumnThenPosition(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	// Create tasks in col 1 and col 0 to test ordering
	seedTask(t, db, project.ID, columns[1].ID, user.ID, "Col1 Task1")
	seedTask(t, db, project.ID, columns[0].ID, user.ID, "Col0 Task1")
	seedTask(t, db, project.ID, columns[0].ID, user.ID, "Col0 Task2")

	tasks, err := ListTasksForProject(db, project.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	// Column 0 tasks should come first (ordered by column position)
	if tasks[0].Title != "Col0 Task1" {
		t.Errorf("first task = %q, want %q", tasks[0].Title, "Col0 Task1")
	}
	if tasks[1].Title != "Col0 Task2" {
		t.Errorf("second task = %q, want %q", tasks[1].Title, "Col0 Task2")
	}
	if tasks[2].Title != "Col1 Task1" {
		t.Errorf("third task = %q, want %q", tasks[2].Title, "Col1 Task1")
	}
}

func TestMoveTask_sameColumn(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col := columns[0]

	t1 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 1")
	t2 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 2")
	t3 := seedTask(t, db, project.ID, col.ID, user.ID, "Task 3")

	// Move Task 3 to position 0 (top)
	if err := MoveTask(db, t3.ID, col.ID, 0); err != nil {
		t.Fatalf("move task: %v", err)
	}

	// Verify new order: Task3, Task1, Task2
	tasks, _ := ListTasksForProject(db, project.ID)
	var colTasks []model.Task
	for _, tk := range tasks {
		if tk.ColumnID == col.ID {
			colTasks = append(colTasks, tk)
		}
	}

	if len(colTasks) != 3 {
		t.Fatalf("got %d tasks in column, want 3", len(colTasks))
	}
	if colTasks[0].ID != t3.ID {
		t.Errorf("first task should be Task 3")
	}
	if colTasks[1].ID != t1.ID {
		t.Errorf("second task should be Task 1")
	}
	if colTasks[2].ID != t2.ID {
		t.Errorf("third task should be Task 2")
	}
}

func TestMoveTask_crossColumn(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col0 := columns[0]
	col1 := columns[1]

	t1 := seedTask(t, db, project.ID, col0.ID, user.ID, "Task 1")
	t2 := seedTask(t, db, project.ID, col0.ID, user.ID, "Task 2")

	// Move Task 1 to col1
	if err := MoveTask(db, t1.ID, col1.ID, 0); err != nil {
		t.Fatalf("move task: %v", err)
	}

	// Verify Task 1 is now in col1
	moved, err := GetTask(db, t1.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if moved.ColumnID != col1.ID {
		t.Error("task should be in new column")
	}

	// Verify col0 has only Task 2 at position 0
	remaining, err := GetTask(db, t2.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if remaining.Position != 0 {
		t.Errorf("remaining task position = %d, want 0", remaining.Position)
	}
}

func TestGetTaskDepth_topLevel(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	task := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Top level")
	depth, err := GetTaskDepth(db, task.ID)
	if err != nil {
		t.Fatalf("GetTaskDepth: %v", err)
	}
	if depth != 0 {
		t.Errorf("depth = %d, want 0", depth)
	}
}

func TestGetTaskDepth_subtask(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	parent := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Parent")
	subtask, _ := CreateTask(db, model.Task{
		ProjectID: project.ID, ColumnID: columns[0].ID, CreatorID: user.ID,
		ParentTaskID: &parent.ID, Title: "Subtask", Priority: "none",
	})

	depth, err := GetTaskDepth(db, subtask.ID)
	if err != nil {
		t.Fatalf("GetTaskDepth: %v", err)
	}
	if depth != 1 {
		t.Errorf("depth = %d, want 1", depth)
	}
}

func TestGetTaskDepth_subSubtask(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	parent := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Parent")
	subtask, _ := CreateTask(db, model.Task{
		ProjectID: project.ID, ColumnID: columns[0].ID, CreatorID: user.ID,
		ParentTaskID: &parent.ID, Title: "Subtask", Priority: "none",
	})
	subSubtask, _ := CreateTask(db, model.Task{
		ProjectID: project.ID, ColumnID: columns[0].ID, CreatorID: user.ID,
		ParentTaskID: &subtask.ID, Title: "Sub-subtask", Priority: "none",
	})

	depth, err := GetTaskDepth(db, subSubtask.ID)
	if err != nil {
		t.Fatalf("GetTaskDepth: %v", err)
	}
	if depth != 2 {
		t.Errorf("depth = %d, want 2", depth)
	}
}

func TestMoveTask_subtasksFollow(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col0 := columns[0]
	col1 := columns[1]

	parent := seedTask(t, db, project.ID, col0.ID, user.ID, "Parent")

	subtask, err := CreateTask(db, model.Task{
		ProjectID:    project.ID,
		ColumnID:     col0.ID,
		CreatorID:    user.ID,
		ParentTaskID: &parent.ID,
		Title:        "Subtask",
		Priority:     "none",
	})
	if err != nil {
		t.Fatalf("create subtask: %v", err)
	}

	if err := MoveTask(db, parent.ID, col1.ID, 0); err != nil {
		t.Fatalf("move parent: %v", err)
	}

	moved, err := GetTask(db, subtask.ID)
	if err != nil {
		t.Fatalf("get subtask: %v", err)
	}
	if moved.ColumnID != col1.ID {
		t.Error("subtask should have followed parent to new column")
	}
}

func TestMoveTask_subSubtasksFollow(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)
	col0 := columns[0]
	col1 := columns[1]

	parent := seedTask(t, db, project.ID, col0.ID, user.ID, "Parent")

	subtask, err := CreateTask(db, model.Task{
		ProjectID:    project.ID,
		ColumnID:     col0.ID,
		CreatorID:    user.ID,
		ParentTaskID: &parent.ID,
		Title:        "Subtask",
		Priority:     "none",
	})
	if err != nil {
		t.Fatalf("create subtask: %v", err)
	}

	subSubtask, err := CreateTask(db, model.Task{
		ProjectID:    project.ID,
		ColumnID:     col0.ID,
		CreatorID:    user.ID,
		ParentTaskID: &subtask.ID,
		Title:        "Sub-subtask",
		Priority:     "none",
	})
	if err != nil {
		t.Fatalf("create sub-subtask: %v", err)
	}

	// Move parent — both subtask and sub-subtask should follow
	if err := MoveTask(db, parent.ID, col1.ID, 0); err != nil {
		t.Fatalf("move parent: %v", err)
	}

	movedSub, err := GetTask(db, subtask.ID)
	if err != nil {
		t.Fatalf("get subtask: %v", err)
	}
	if movedSub.ColumnID != col1.ID {
		t.Error("subtask should have followed parent to new column")
	}

	movedSubSub, err := GetTask(db, subSubtask.ID)
	if err != nil {
		t.Fatalf("get sub-subtask: %v", err)
	}
	if movedSubSub.ColumnID != col1.ID {
		t.Error("sub-subtask should have followed parent to new column")
	}
}

func TestUpdateTask(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	task := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Original")
	task.Title = "Updated"
	task.Description = "New description"
	task.Priority = "high"

	updated, err := UpdateTask(db, task)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("title = %q, want %q", updated.Title, "Updated")
	}

	fetched, _ := GetTask(db, task.ID)
	if fetched.Description != "New description" {
		t.Errorf("description = %q, want %q", fetched.Description, "New description")
	}
	if fetched.Priority != "high" {
		t.Errorf("priority = %q, want %q", fetched.Priority, "high")
	}
}

func TestDeleteTask(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	task := seedTask(t, db, project.ID, columns[0].ID, user.ID, "To Delete")

	if err := DeleteTask(db, task.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	_, err := GetTask(db, task.ID)
	if err != ErrTaskNotFound {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestSearchTasks_findsByTitle(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	seedTask(t, db, project.ID, columns[0].ID, user.ID, "Fix login bug")
	seedTask(t, db, project.ID, columns[0].ID, user.ID, "Add signup page")

	results, err := SearchTasks(db, user.ID, "login")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "Fix login bug" {
		t.Errorf("title = %q, want %q", results[0].Title, "Fix login bug")
	}
	if results[0].ProjectName != "Board" {
		t.Errorf("project name = %q, want %q", results[0].ProjectName, "Board")
	}
}

func TestSearchTasks_findsByTaskNumber(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")

	project, err := CreateProject(db, model.Project{
		Name:        "My Board",
		Tag:         "MB",
		Visibility:  "private",
		OwnerUserID: &user.ID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	_ = CreateDefaultColumns(db, project.ID)
	columns, _ := GetColumnsForProject(db, project.ID)

	seedTask(t, db, project.ID, columns[0].ID, user.ID, "First task")
	seedTask(t, db, project.ID, columns[0].ID, user.ID, "Second task")

	results, err := SearchTasks(db, user.ID, "MB-2")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Title != "Second task" {
		t.Errorf("title = %q, want %q", results[0].Title, "Second task")
	}
}

func TestSearchTasks_respectsVisibility(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	alice := seedUser(t, db, "Alice", "alice@test.com")
	bob := seedUser(t, db, "Bob", "bob@test.com")

	// Alice's private project
	project := seedProject(t, db, "Secret", &alice.ID, nil) // private by default
	columns, _ := GetColumnsForProject(db, project.ID)
	seedTask(t, db, project.ID, columns[0].ID, alice.ID, "Hidden task")

	// Bob should not find Alice's private task
	results, err := SearchTasks(db, bob.ID, "Hidden")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0 (private project should be hidden)", len(results))
	}

	// Alice should find her own task
	results, err = SearchTasks(db, alice.ID, "Hidden")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}
