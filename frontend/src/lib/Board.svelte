<script>
  import TaskCard from './TaskCard.svelte';

  let { project, onTaskClick } = $props();

  function tasksForColumn(columnId) {
    return (project.tasks || []).filter(t => t.columnId === columnId);
  }
</script>

<div class="board">
  {#each project.columns as column}
    <div class="column">
      <div class="column-header">
        <span class="column-name">{column.name}</span>
        <span class="column-count">{tasksForColumn(column.id).length}</span>
      </div>
      <div class="column-body">
        {#each tasksForColumn(column.id) as task (task.id)}
          <TaskCard {task} labels={project.labels} onclick={onTaskClick} />
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
  }
</style>
