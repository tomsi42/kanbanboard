<script>
  let { task, labels = [], allTasks = [], doneColumnId = '' } = $props();

  let label = $derived(labels.find(l => l.id === task.labelId));

  // Subtask progress for parent tasks
  let subtasks = $derived(
    !task.parentTaskId ? allTasks.filter(t => t.parentTaskId === task.id) : []
  );
  let doneCount = $derived(
    subtasks.filter(t => t.columnId === doneColumnId).length
  );
  let hasSubtasks = $derived(subtasks.length > 0);

  // Parent task name for subtask indicator
  let parentName = $derived(
    task.parentTaskId ? allTasks.find(t => t.id === task.parentTaskId)?.title : null
  );
</script>

<div class="card">
  <div class="card-top">
    {#if label}
      <span class="label" style="background: {label.color}">{label.name}</span>
    {/if}
    {#if hasSubtasks}
      <span class="progress" class:all-done={doneCount === subtasks.length}>
        {doneCount}/{subtasks.length}
      </span>
    {/if}
  </div>
  <span class="title">{task.title}</span>
  {#if parentName}
    <span class="subtask-indicator">↳ {parentName}</span>
  {/if}
</div>

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
    cursor: pointer;
    text-align: left;
    width: 100%;
    box-sizing: border-box;
    font-size: 0.875rem;
  }

  .card:hover {
    border-color: #4a90d9;
    box-shadow: 0 1px 4px rgba(74, 144, 217, 0.15);
  }

  .card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 4px;
  }

  .title {
    color: #333;
    line-height: 1.3;
  }

  .label {
    display: inline-block;
    padding: 1px 6px;
    border-radius: 3px;
    font-size: 0.7rem;
    color: white;
    font-weight: 500;
  }

  .progress {
    font-size: 0.7rem;
    color: #888;
    background: #eee;
    padding: 1px 6px;
    border-radius: 3px;
    font-weight: 500;
  }

  .progress.all-done {
    color: #0a0;
    background: #e6ffe6;
  }

  .subtask-indicator {
    font-size: 0.7rem;
    color: #888;
  }
</style>
