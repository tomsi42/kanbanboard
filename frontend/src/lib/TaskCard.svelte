<script>
  let { task, labels = [], onclick } = $props();

  let label = $derived(labels.find(l => l.id === task.labelId));
</script>

<button class="card" onclick={() => onclick?.(task)}>
  {#if label}
    <span class="label" style="background: {label.color}">{label.name}</span>
  {/if}
  <span class="title">{task.title}</span>
  {#if task.parentTaskId}
    <span class="subtask-indicator">↳ subtask</span>
  {/if}
</button>

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
    align-self: flex-start;
  }

  .subtask-indicator {
    font-size: 0.7rem;
    color: #888;
  }
</style>
