package store

import "testing"

func TestAddBlocker_basic(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	// A blocks B
	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker: %v", err)
	}

	blockers, err := GetBlockersForTask(db, taskB.ID)
	if err != nil {
		t.Fatalf("GetBlockersForTask: %v", err)
	}
	if len(blockers) != 1 || blockers[0] != taskA.ID {
		t.Errorf("blockers = %v, want [%s]", blockers, taskA.ID)
	}
}

func TestAddBlocker_selfReference(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")

	err := AddBlocker(db, project.ID, taskA.ID, taskA.ID)
	if err != ErrSelfDependency {
		t.Errorf("err = %v, want ErrSelfDependency", err)
	}
}

func TestAddBlocker_crossProject(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	p1 := seedProject(t, db, "Project 1", &user.ID, nil)
	p2 := seedProject(t, db, "Project 2", &user.ID, nil)
	cols1, _ := GetColumnsForProject(db, p1.ID)
	cols2, _ := GetColumnsForProject(db, p2.ID)

	taskA := seedTask(t, db, p1.ID, cols1[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, p2.ID, cols2[0].ID, user.ID, "Task B")

	err := AddBlocker(db, p1.ID, taskA.ID, taskB.ID)
	if err != ErrCrossProjectDependency {
		t.Errorf("err = %v, want ErrCrossProjectDependency", err)
	}
}

func TestAddBlocker_cycleDetection(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")
	taskC := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task C")

	// A blocks B, B blocks C
	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker A→B: %v", err)
	}
	if err := AddBlocker(db, project.ID, taskB.ID, taskC.ID); err != nil {
		t.Fatalf("AddBlocker B→C: %v", err)
	}

	// Attempting C blocks A would create a cycle: A→B→C→A
	err := AddBlocker(db, project.ID, taskC.ID, taskA.ID)
	if err != ErrDependencyCycle {
		t.Errorf("err = %v, want ErrDependencyCycle", err)
	}
}

func TestAddBlocker_directCycle(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	// A blocks B
	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker A→B: %v", err)
	}

	// B blocks A would create a direct cycle
	err := AddBlocker(db, project.ID, taskB.ID, taskA.ID)
	if err != ErrDependencyCycle {
		t.Errorf("err = %v, want ErrDependencyCycle", err)
	}
}

func TestRemoveBlocker(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker: %v", err)
	}

	if err := RemoveBlocker(db, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("RemoveBlocker: %v", err)
	}

	blockers, err := GetBlockersForTask(db, taskB.ID)
	if err != nil {
		t.Fatalf("GetBlockersForTask: %v", err)
	}
	if len(blockers) != 0 {
		t.Errorf("blockers = %v, want empty after removal", blockers)
	}
}

func TestRemoveBlocker_noOp(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	// Removing a non-existent dependency should not error
	if err := RemoveBlocker(db, taskA.ID, taskB.ID); err != nil {
		t.Errorf("RemoveBlocker on non-existent dependency: %v", err)
	}
}

func TestListBlockersForProject(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")
	taskC := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task C")

	// A blocks C, B blocks C
	if err := AddBlocker(db, project.ID, taskA.ID, taskC.ID); err != nil {
		t.Fatalf("AddBlocker A→C: %v", err)
	}
	if err := AddBlocker(db, project.ID, taskB.ID, taskC.ID); err != nil {
		t.Fatalf("AddBlocker B→C: %v", err)
	}

	m, err := ListBlockersForProject(db, project.ID)
	if err != nil {
		t.Fatalf("ListBlockersForProject: %v", err)
	}
	if len(m[taskC.ID]) != 2 {
		t.Errorf("task C has %d blockers, want 2", len(m[taskC.ID]))
	}
	// A and B should not be blocked
	if len(m[taskA.ID]) != 0 || len(m[taskB.ID]) != 0 {
		t.Error("tasks A and B should not be blocked")
	}
}

func TestListTasksForProject_hydratesBlockedBy(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker: %v", err)
	}

	tasks, err := ListTasksForProject(db, project.ID)
	if err != nil {
		t.Fatalf("ListTasksForProject: %v", err)
	}

	for _, task := range tasks {
		if task.ID == taskB.ID {
			if len(task.BlockedBy) != 1 || task.BlockedBy[0] != taskA.ID {
				t.Errorf("task B BlockedBy = %v, want [%s]", task.BlockedBy, taskA.ID)
			}
		}
		if task.ID == taskA.ID && len(task.BlockedBy) != 0 {
			t.Errorf("task A should not be blocked, got %v", task.BlockedBy)
		}
	}
}

func TestAddBlocker_deleteCascades(t *testing.T) {
	db := testDB(t)
	cleanTables(t, db)
	user := seedUser(t, db, "Alice", "alice@test.com")
	project := seedProject(t, db, "Board", &user.ID, nil)
	columns, _ := GetColumnsForProject(db, project.ID)

	taskA := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task A")
	taskB := seedTask(t, db, project.ID, columns[0].ID, user.ID, "Task B")

	if err := AddBlocker(db, project.ID, taskA.ID, taskB.ID); err != nil {
		t.Fatalf("AddBlocker: %v", err)
	}

	// Deleting the blocker task should remove the dependency
	if err := DeleteTask(db, taskA.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	blockers, err := GetBlockersForTask(db, taskB.ID)
	if err != nil {
		t.Fatalf("GetBlockersForTask: %v", err)
	}
	if len(blockers) != 0 {
		t.Errorf("blockers = %v, want empty after blocker task deleted", blockers)
	}
}
