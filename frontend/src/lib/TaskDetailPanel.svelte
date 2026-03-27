<script>
  import { updateTask, deleteTask } from './api.js';

  let { task, project, onUpdate, onDelete, onClose } = $props();

  let title = $state(task.title);
  let description = $state(task.description || '');
  let columnId = $state(task.columnId);
  let labelId = $state(task.labelId || '');
  let priority = $state(task.priority || 'none');
  let targetVersion = $state(task.targetVersion || '');
  let dueDate = $state(task.dueDate ? task.dueDate.split('T')[0] : '');
  let saving = $state(false);

  // Reset form when task changes
  $effect(() => {
    title = task.title;
    description = task.description || '';
    columnId = task.columnId;
    labelId = task.labelId || '';
    priority = task.priority || 'none';
    targetVersion = task.targetVersion || '';
    dueDate = task.dueDate ? task.dueDate.split('T')[0] : '';
  });

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

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="overlay" onclick={onClose} role="presentation">
  <!-- svelte-ignore a11y_interactive_supports_focus a11y_click_events_have_key_events -->
  <div class="panel" onclick={(e) => e.stopPropagation()} role="dialog">
    <div class="panel-header">
      <span class="saving-indicator">{saving ? 'Saving...' : ''}</span>
      <button class="close-btn" onclick={onClose}>✕</button>
    </div>

    <div class="panel-body">
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

      <div class="meta">
        <span>Created: {new Date(task.createdAt).toLocaleDateString()}</span>
      </div>

      <div class="danger-zone">
        <button class="delete-btn" onclick={handleDelete}>Delete Task</button>
      </div>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.2);
    z-index: 200;
    display: flex;
    justify-content: flex-end;
  }

  .panel {
    width: 400px;
    max-width: 90vw;
    background: white;
    box-shadow: -4px 0 16px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow-y: auto;
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
