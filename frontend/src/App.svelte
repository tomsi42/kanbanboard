<script>
  import { getSetupStatus, getAppTitle, getMe, logout as apiLogout, listProjects, getProject, createTask } from './lib/api.js';
  import Onboarding from './lib/Onboarding.svelte';
  import Login from './lib/Login.svelte';
  import ProjectDropdown from './lib/ProjectDropdown.svelte';
  import CreateProjectModal from './lib/CreateProjectModal.svelte';
  import Board from './lib/Board.svelte';

  let loading = $state(true);
  let setupRequired = $state(false);
  let appTitle = $state('Kanban Board');
  let currentUser = $state(null);
  let projects = $state([]);
  let currentProject = $state(null);
  let showCreateProject = $state(false);
  let addingTask = $state(false);
  let newTaskTitle = $state('');

  async function checkStatus() {
    loading = true;
    try {
      const [status, titleData] = await Promise.all([
        getSetupStatus(),
        getAppTitle().catch(() => ({ title: 'Kanban Board' })),
      ]);
      setupRequired = status.setupRequired;
      appTitle = titleData.title;

      if (!setupRequired) {
        try {
          currentUser = await getMe();
          await loadProjects();
        } catch {
          currentUser = null;
        }
      }
    } catch {
      // API unreachable
    } finally {
      loading = false;
    }
  }

  async function loadProjects() {
    projects = await listProjects();
    if (projects.length > 0 && !currentProject) {
      await selectProject(projects[0]);
    }
  }

  async function selectProject(project) {
    currentProject = await getProject(project.id);
  }

  async function reloadCurrentProject() {
    if (currentProject) {
      currentProject = await getProject(currentProject.id);
    }
  }

  function handleSetupComplete() {
    checkStatus();
  }

  async function handleLogin(user) {
    currentUser = user;
    await loadProjects();
  }

  async function handleLogout() {
    await apiLogout();
    currentUser = null;
    projects = [];
    currentProject = null;
  }

  async function handleProjectCreated(project) {
    showCreateProject = false;
    projects = [...projects, project];
    currentProject = project;
  }

  function startAddTask() {
    addingTask = true;
    newTaskTitle = '';
  }

  async function submitAddTask() {
    if (!newTaskTitle.trim() || !currentProject) return;

    const firstColumn = currentProject.columns[0];
    if (!firstColumn) return;

    try {
      await createTask(currentProject.id, {
        title: newTaskTitle.trim(),
        columnId: firstColumn.id,
      });
      newTaskTitle = '';
      addingTask = false;
      await reloadCurrentProject();
    } catch (err) {
      // Keep the input open on error
    }
  }

  function cancelAddTask() {
    addingTask = false;
    newTaskTitle = '';
  }

  function handleAddTaskKeydown(e) {
    if (e.key === 'Enter') {
      submitAddTask();
    } else if (e.key === 'Escape') {
      cancelAddTask();
    }
  }

  function handleTaskClick(task) {
    // Task detail panel comes in Phase 2.3
  }

  $effect(() => {
    checkStatus();
  });
</script>

{#if loading}
  <div class="center">
    <p>Loading...</p>
  </div>
{:else if setupRequired}
  <Onboarding onComplete={handleSetupComplete} />
{:else if !currentUser}
  <Login {appTitle} onLogin={handleLogin} />
{:else}
  <div class="app">
    <header>
      <div class="header-left">
        <ProjectDropdown
          {projects}
          {currentProject}
          onSelect={selectProject}
          onCreateNew={() => showCreateProject = true}
        />
        {#if currentProject}
          {#if addingTask}
            <div class="add-task-inline">
              <input
                type="text"
                placeholder="Task title..."
                bind:value={newTaskTitle}
                onkeydown={handleAddTaskKeydown}
              />
              <button class="add-confirm" onclick={submitAddTask}>Add</button>
              <button class="add-cancel" onclick={cancelAddTask}>✕</button>
            </div>
          {:else}
            <button class="add-task-btn" onclick={startAddTask}>+ Add Task</button>
          {/if}
        {/if}
      </div>
      <div class="header-right">
        <span class="user-name">{currentUser.name}</span>
        <button class="sign-out" onclick={handleLogout}>Sign Out</button>
      </div>
    </header>

    <main>
      {#if projects.length === 0}
        <div class="center empty-state">
          <h2>Welcome to {appTitle}</h2>
          <p>Create your first project to get started.</p>
          <button class="create-btn" onclick={() => showCreateProject = true}>
            Create Project
          </button>
        </div>
      {:else if currentProject}
        <Board project={currentProject} onTaskClick={handleTaskClick} />
      {/if}
    </main>
  </div>

  {#if showCreateProject}
    <CreateProjectModal
      onCreated={handleProjectCreated}
      onCancel={() => showCreateProject = false}
    />
  {/if}
{/if}

<style>
  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
  }

  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: #f5f5f5;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 16px;
    border-bottom: 1px solid #e0e0e0;
    background: #fff;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .user-name {
    color: #555;
    font-size: 0.875rem;
  }

  .sign-out {
    padding: 6px 12px;
    background: none;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    color: #555;
  }

  .sign-out:hover {
    background: #f5f5f5;
  }

  main {
    flex: 1;
    overflow-x: auto;
  }

  .empty-state {
    min-height: calc(100vh - 50px);
  }

  .empty-state h2 {
    color: #333;
    margin: 0 0 8px;
  }

  .empty-state p {
    color: #666;
    margin: 0 0 24px;
  }

  .create-btn {
    padding: 10px 24px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
  }

  .create-btn:hover {
    background: #357abd;
  }

  .add-task-btn {
    padding: 6px 12px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
  }

  .add-task-btn:hover {
    background: #357abd;
  }

  .add-task-inline {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .add-task-inline input {
    padding: 5px 10px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.875rem;
    width: 200px;
  }

  .add-task-inline input:focus {
    outline: none;
    border-color: #4a90d9;
  }

  .add-confirm {
    padding: 5px 10px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
  }

  .add-cancel {
    padding: 5px 8px;
    background: none;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    color: #888;
  }

  p {
    color: #666;
  }
</style>
