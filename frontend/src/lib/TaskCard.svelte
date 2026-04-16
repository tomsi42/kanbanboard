<script>
  let { task, labels = [], allTasks = [], doneColumnId = '', project = null, members = [] } = $props();

  let label = $derived(labels.find(l => l.id === task.labelId));

  // Subtask detection for parent tasks
  let hasSubtasks = $derived(
    !task.parentTaskId ? allTasks.some(t => t.parentTaskId === task.id) : false
  );

  // Parent task name for subtask indicator
  let parentName = $derived(
    task.parentTaskId ? allTasks.find(t => t.id === task.parentTaskId)?.title : null
  );

  // Task reference (e.g. "KB-7")
  let taskRef = $derived(
    project?.tag ? `${project.tag}-${task.taskNumber}` : null
  );

  // Assignee initials
  let assigneeInitials = $derived(() => {
    if (!task.assigneeId) return null;
    const member = members.find(m => m.id === task.assigneeId);
    if (!member) return null;
    return member.name.split(/\s+/).map(w => w[0]).join('').toUpperCase().substring(0, 2);
  });

  // Actively blocked: has at least one blocker not yet in the Done column
  let isBlocked = $derived(
    task.blockedBy?.length > 0 &&
    task.blockedBy.some(id => {
      const blocker = allTasks.find(t => t.id === id);
      return blocker && blocker.columnId !== doneColumnId;
    })
  );

  // Light tint of label colour for card background
  function lightTint(hex, opacity = 0.12) {
    if (!hex) return 'white';
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${opacity})`;
  }

  let cardBg = $derived(label ? lightTint(label.color) : 'white');
</script>

<div class="card" style="background: {cardBg}">
  <div class="card-top">
    <div class="card-top-left">
      {#if label}
        <span class="label" style="background: {label.color}">{label.name}</span>
      {/if}
      {#if hasSubtasks}
        <span class="parent-icon" title="Has subtasks">▤</span>
      {/if}
      {#if isBlocked}
        <span class="blocked-icon" title="Blocked">●</span>
      {/if}
    </div>
    <div class="card-top-right">
      {#if taskRef}
        <span class="task-ref">{taskRef}</span>
      {/if}
      {#if assigneeInitials()}
        <span class="avatar" title={members.find(m => m.id === task.assigneeId)?.name}>{assigneeInitials()}</span>
      {/if}
    </div>
  </div>
  {#if parentName}
    <span class="parent-name">{parentName}</span>
  {/if}
  <span class="title">{#if task.parentTaskId}▶ {/if}{task.title}</span>
</div>

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px 10px;
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

  .card-top-left {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .card-top-right {
    display: flex;
    align-items: center;
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

  .parent-icon {
    font-size: 0.75rem;
    color: #888;
  }

  .blocked-icon {
    font-size: 0.6rem;
    color: #c00;
    line-height: 1;
  }

  .task-ref {
    font-size: 0.7rem;
    color: #888;
    font-weight: 500;
  }

  .avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #4a90d9;
    color: white;
    font-size: 0.6rem;
    font-weight: 600;
    flex-shrink: 0;
  }

  .parent-name {
    font-size: 0.8rem;
    color: #666;
    line-height: 1.2;
    margin: 2px 0;
  }

</style>
