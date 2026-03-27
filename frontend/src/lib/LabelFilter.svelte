<script>
  let { labels = [], filterLabelId = '', onChange } = $props();
</script>

<div class="filter">
  <select value={filterLabelId} onchange={(e) => onChange(e.target.value)}>
    <option value="">All</option>
    {#each labels as label}
      <option value={label.id}>{label.name}</option>
    {/each}
  </select>
  {#if filterLabelId}
    {@const activeLabel = labels.find(l => l.id === filterLabelId)}
    {#if activeLabel}
      <span class="active-dot" style="background: {activeLabel.color}"></span>
    {/if}
  {/if}
</div>

<style>
  .filter {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  select {
    padding: 5px 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.875rem;
    color: #333;
    background: white;
    cursor: pointer;
  }

  select:focus {
    outline: none;
    border-color: #4a90d9;
  }

  .active-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
  }
</style>
