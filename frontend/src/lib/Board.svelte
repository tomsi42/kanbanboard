<script>
  import { dndzone } from 'svelte-dnd-action';
  import TaskCard from './TaskCard.svelte';

  let { project, onTaskClick, onTaskMove, filterLabelId = '' } = $props();

  // Build column data with their tasks for dndzone, applying label filter
  let columnData = $derived(
    project.columns.map(col => ({
      ...col,
      tasks: (project.tasks || [])
        .filter(t => t.columnId === col.id)
        .filter(t => !filterLabelId || t.labelId === filterLabelId)
        .sort((a, b) => a.position - b.position),
    }))
  );

  // Local mutable copy for drag state
  let localColumns = $state([]);

  $effect(() => {
    localColumns = columnData.map(col => ({
      ...col,
      tasks: [...col.tasks],
    }));
  });

  function handleConsider(columnId, e) {
    const colIdx = localColumns.findIndex(c => c.id === columnId);
    if (colIdx >= 0) {
      localColumns[colIdx].tasks = e.detail.items;
    }
  }

  function handleFinalize(columnId, e) {
    const colIdx = localColumns.findIndex(c => c.id === columnId);
    if (colIdx >= 0) {
      localColumns[colIdx].tasks = e.detail.items;

      // Find the moved task and its new position
      const tasks = e.detail.items;
      for (let i = 0; i < tasks.length; i++) {
        const task = tasks[i];
        // Check if this task moved to this column or changed position
        if (task.columnId !== columnId || task.position !== i) {
          onTaskMove?.(task.id, columnId, i);
          return;
        }
      }
    }
  }
</script>

<div class="board">
  {#each localColumns as column (column.id)}
    <div class="column">
      <div class="column-header">
        <span class="column-name">{column.name}</span>
        <span class="column-count">{column.tasks.length}</span>
      </div>
      <div
        class="column-body"
        use:dndzone={{ items: column.tasks, flipDurationMs: 200, type: 'task' }}
        onconsider={(e) => handleConsider(column.id, e)}
        onfinalize={(e) => handleFinalize(column.id, e)}
      >
        {#each column.tasks as task (task.id)}
          <div class="card-wrapper">
            <TaskCard {task} labels={project.labels} onclick={onTaskClick} />
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .board {
    display: flex;
    gap: 12px;
    padding: 16px;
    min-height: calc(100vh - 50px);
    overflow-x: auto;
  }

  .column {
    flex: 0 0 260px;
    background: #e8e8e8;
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 82px);
  }

  .column-header {
    padding: 10px 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .column-name {
    font-weight: 600;
    font-size: 0.875rem;
    color: #333;
  }

  .column-count {
    font-size: 0.75rem;
    color: #888;
    background: #d0d0d0;
    padding: 1px 6px;
    border-radius: 10px;
  }

  .column-body {
    padding: 4px 8px 8px;
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 60px;
  }

  .card-wrapper {
    /* Needed for dndzone to properly track elements */
  }
</style>
