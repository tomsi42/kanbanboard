<script>
  import { updateTask, deleteTask, createTask } from './api.js';

  let { task, project, onUpdate, onDelete, onClose, onTaskSelect } = $props();

  let title = $state(task.title);
  let description = $state(task.description || '');
  let columnId = $state(task.columnId);
  let labelId = $state(task.labelId || '');
  let priority = $state(task.priority || 'none');
  let targetVersion = $state(task.targetVersion || '');
  let dueDate = $state(task.dueDate ? task.dueDate.split('T')[0] : '');
  let saving = $state(false);
  let newSubtaskTitle = $state('');
  let addingSubtask = $state(false);

  // Reset form when task changes
  $effect(() => {
    title = task.title;
    description = task.description || '';
    columnId = task.columnId;
    labelId = task.labelId || '';
    priority = task.priority || 'none';
    targetVersion = task.targetVersion || '';
    dueDate = task.dueDate ? task.dueDate.split('T')[0] : '';
    newSubtaskTitle = '';
    addingSubtask = false;
  });

  // Derive subtasks and parent from project tasks
  let subtasks = $derived(
    (project.tasks || []).filter(t => t.parentTaskId === task.id)
  );

  let parentTask = $derived(
    task.parentTaskId ? (project.tasks || []).find(t => t.id === task.parentTaskId) : null
  );

  // Get column name for display
  function columnName(colId) {
    const col = project.columns.find(c => c.id === colId);
    return col ? col.name : '';
  }

  async function save(field, value) {
    if (saving) return;
    saving = true;
    try {
      await updateTask(project.id, task.id, { [field]: value });
      onUpdate();
    } catch (err) {
      // Revert on error could be added here
    } finally {
      saving = false;
    }
  }

  function handleTitleBlur() {
    if (title !== task.title && title.trim()) {
      save('title', title.trim());
    }
  }

  function handleDescriptionBlur() {
    if (description !== (task.description || '')) {
      save('description', description);
    }
  }

  function handleColumnChange() {
    if (columnId !== task.columnId) {
      save('columnId', columnId);
    }
  }

  function handleLabelChange() {
    save('labelId', labelId);
  }

  function handlePriorityChange() {
    if (priority !== task.priority) {
      save('priority', priority);
    }
  }

  function handleTargetVersionBlur() {
    if (targetVersion !== (task.targetVersion || '')) {
      save('targetVersion', targetVersion);
    }
  }

  function handleDueDateChange() {
    save('dueDate', dueDate);
  }

  async function handleAddSubtask() {
    if (!newSubtaskTitle.trim()) return;
    try {
      await createTask(project.id, {
        title: newSubtaskTitle.trim(),
        columnId: task.columnId,
        parentTaskId: task.id,
      });
      newSubtaskTitle = '';
      addingSubtask = false;
      onUpdate();
    } catch (err) {
      // Handle error
    }
  }

  function handleSubtaskKeydown(e) {
    if (e.key === 'Enter') handleAddSubtask();
    if (e.key === 'Escape') { addingSubtask = false; newSubtaskTitle = ''; }
  }

  async function handleDelete() {
    if (!confirm('Delete this task?')) return;
    try {
      await deleteTask(project.id, task.id);
      onDelete();
    } catch (err) {
      // Handle error
    }
  }
</script>

<div class="panel" role="dialog">
    <div class="panel-header">
      <span class="saving-indicator">{saving ? 'Saving...' : ''}</span>
      <button class="close-btn" onclick={onClose}>✕</button>
    </div>

    <div class="panel-body">
      {#if parentTask}
        <div class="parent-link">
          <span class="parent-label">↳ Subtask of</span>
          <button class="parent-btn" onclick={() => onTaskSelect?.(parentTask)}>
            {parentTask.title}
          </button>
        </div>
      {/if}

      <div class="field">
        <input
          class="title-input"
          type="text"
          bind:value={title}
          onblur={handleTitleBlur}
          placeholder="Task title"
        />
      </div>

      <div class="field">
        <label for="description">Description</label>
        <textarea
          id="description"
          bind:value={description}
          onblur={handleDescriptionBlur}
          placeholder="Add a description..."
          rows="4"
        ></textarea>
      </div>

      <div class="field-row">
        <div class="field half">
          <label for="column">Column</label>
          <select id="column" bind:value={columnId} onchange={handleColumnChange}>
            {#each project.columns as col}
              <option value={col.id}>{col.name}</option>
            {/each}
          </select>
        </div>

        <div class="field half">
          <label for="label">Label</label>
          <select id="label" bind:value={labelId} onchange={handleLabelChange}>
            <option value="">None</option>
            {#each project.labels as lbl}
              <option value={lbl.id}>{lbl.name}</option>
            {/each}
          </select>
        </div>
      </div>

      <div class="field-row">
        <div class="field half">
          <label for="priority">Priority</label>
          <select id="priority" bind:value={priority} onchange={handlePriorityChange}>
            <option value="none">None</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
        </div>

        <div class="field half">
          <label for="dueDate">Due Date</label>
          <input id="dueDate" type="date" bind:value={dueDate} onchange={handleDueDateChange} />
        </div>
      </div>

      <div class="field">
        <label for="targetVersion">Target Version</label>
        <input
          id="targetVersion"
          type="text"
          bind:value={targetVersion}
          onblur={handleTargetVersionBlur}
          placeholder="e.g. v1.0"
        />
      </div>

      <!-- Subtasks section (only for non-subtasks) -->
      {#if !task.parentTaskId}
        <div class="subtasks-section">
          <label>Subtasks</label>
          {#if subtasks.length > 0}
            <div class="subtask-list">
              {#each subtasks as sub (sub.id)}
                <button class="subtask-item" onclick={() => onTaskSelect?.(sub)}>
                  <span class="subtask-title">{sub.title}</span>
                  <span class="subtask-column">{columnName(sub.columnId)}</span>
                </button>
              {/each}
            </div>
          {/if}
          {#if addingSubtask}
            <div class="add-subtask-input">
              <input
                type="text"
                placeholder="Subtask title..."
                bind:value={newSubtaskTitle}
                onkeydown={handleSubtaskKeydown}
              />
              <button class="add-confirm" onclick={handleAddSubtask}>Add</button>
              <button class="add-cancel" onclick={() => { addingSubtask = false; newSubtaskTitle = ''; }}>✕</button>
            </div>
          {:else}
            <button class="add-subtask-btn" onclick={() => addingSubtask = true}>
              + Add subtask
            </button>
          {/if}
        </div>
      {/if}

      <div class="meta">
        <span>Created: {new Date(task.createdAt).toLocaleDateString()}</span>
      </div>

      <div class="danger-zone">
        <button class="delete-btn" onclick={handleDelete}>Delete Task</button>
      </div>
    </div>
  </div>


<style>
  .panel {
    position: fixed;
    top: 0;
    right: 0;
    width: 400px;
    max-width: 90vw;
    background: white;
    box-shadow: -4px 0 16px rgba(0, 0, 0, 0.15);
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow-y: auto;
    z-index: 200;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;
  }

  .saving-indicator {
    font-size: 0.75rem;
    color: #888;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.1rem;
    cursor: pointer;
    color: #888;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .close-btn:hover {
    background: #f0f0f0;
    color: #333;
  }

  .panel-body {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .parent-link {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    background: #f0f4ff;
    border-radius: 4px;
  }

  .parent-label {
    font-size: 0.75rem;
    color: #666;
  }

  .parent-btn {
    background: none;
    border: none;
    color: #4a90d9;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0;
  }

  .parent-btn:hover {
    text-decoration: underline;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field-row {
    display: flex;
    gap: 12px;
  }

  .half {
    flex: 1;
  }

  label {
    font-size: 0.75rem;
    font-weight: 600;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .title-input {
    font-size: 1.1rem;
    font-weight: 600;
    border: 1px solid transparent;
    padding: 6px 8px;
    border-radius: 4px;
    color: #333;
    width: 100%;
    box-sizing: border-box;
  }

  .title-input:hover {
    border-color: #ddd;
  }

  .title-input:focus {
    outline: none;
    border-color: #4a90d9;
  }

  textarea {
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 8px;
    font-size: 0.875rem;
    font-family: inherit;
    resize: vertical;
    color: #333;
  }

  textarea:focus {
    outline: none;
    border-color: #4a90d9;
  }

  select, input[type="text"], input[type="date"] {
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 6px 8px;
    font-size: 0.875rem;
    color: #333;
    width: 100%;
    box-sizing: border-box;
  }

  select:focus, input:focus {
    outline: none;
    border-color: #4a90d9;
  }

  .subtasks-section {
    border-top: 1px solid #eee;
    padding-top: 12px;
  }

  .subtask-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin: 8px 0;
  }

  .subtask-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 8px;
    background: #f8f8f8;
    border: 1px solid #e8e8e8;
    border-radius: 4px;
    cursor: pointer;
    text-align: left;
  }

  .subtask-item:hover {
    border-color: #4a90d9;
    background: #f0f4ff;
  }

  .subtask-title {
    font-size: 0.85rem;
    color: #333;
  }

  .subtask-column {
    font-size: 0.7rem;
    color: #888;
    background: #e8e8e8;
    padding: 1px 6px;
    border-radius: 3px;
  }

  .add-subtask-btn {
    background: none;
    border: none;
    color: #4a90d9;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 4px 0;
    margin-top: 4px;
  }

  .add-subtask-btn:hover {
    text-decoration: underline;
  }

  .add-subtask-input {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-top: 6px;
  }

  .add-subtask-input input {
    flex: 1;
    padding: 5px 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .add-confirm {
    padding: 4px 10px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .add-cancel {
    padding: 4px 8px;
    background: none;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
    color: #888;
  }

  .meta {
    font-size: 0.75rem;
    color: #999;
    padding-top: 8px;
    border-top: 1px solid #eee;
  }

  .danger-zone {
    padding-top: 16px;
    border-top: 1px solid #eee;
  }

  .delete-btn {
    padding: 6px 12px;
    background: none;
    border: 1px solid #e53e3e;
    color: #e53e3e;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
  }

  .delete-btn:hover {
    background: #fff5f5;
  }
</style>
