<script>
  import { listTeams, createTeam, updateTeam, deleteTeam, listTeamMembers, addTeamMember, removeTeamMember, listUsersBasic } from './api.js';

  let { onBack } = $props();

  let teams = $state([]);
  let allUsers = $state([]);
  let loading = $state(true);
  let message = $state('');
  let error = $state('');

  // Create team
  let showCreateForm = $state(false);
  let newTeamName = $state('');
  let createError = $state('');

  // Expanded team (showing members)
  let expandedTeamId = $state(null);
  let members = $state([]);
  let addUserId = $state('');

  // Rename
  let renamingTeamId = $state(null);
  let renameValue = $state('');

  async function loadTeams() {
    loading = true;
    try {
      [teams, allUsers] = await Promise.all([listTeams(), listUsersBasic()]);
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  $effect(() => { loadTeams(); });

  async function handleCreateTeam(e) {
    e.preventDefault();
    createError = '';
    if (!newTeamName.trim()) { createError = 'Team name is required.'; return; }
    try {
      await createTeam(newTeamName.trim());
      showCreateForm = false;
      newTeamName = '';
      message = 'Team created.';
      await loadTeams();
    } catch (err) {
      createError = err.message;
    }
  }

  async function toggleExpand(teamId) {
    if (expandedTeamId === teamId) {
      expandedTeamId = null;
      members = [];
      return;
    }
    expandedTeamId = teamId;
    try {
      members = await listTeamMembers(teamId);
    } catch {
      members = [];
    }
  }

  async function handleAddMember() {
    if (!addUserId || !expandedTeamId) return;
    try {
      await addTeamMember(expandedTeamId, addUserId);
      addUserId = '';
      members = await listTeamMembers(expandedTeamId);
      message = 'Member added.';
    } catch (err) {
      error = err.message;
    }
  }

  async function handleRemoveMember(userId) {
    if (!confirm('Remove this member from the team?')) return;
    try {
      await removeTeamMember(expandedTeamId, userId);
      members = await listTeamMembers(expandedTeamId);
      message = 'Member removed.';
    } catch (err) {
      error = err.message;
    }
  }

  function startRename(team) {
    renamingTeamId = team.id;
    renameValue = team.name;
  }

  async function handleRename(e) {
    e.preventDefault();
    if (!renameValue.trim()) return;
    try {
      await updateTeam(renamingTeamId, renameValue.trim());
      renamingTeamId = null;
      message = 'Team renamed.';
      await loadTeams();
    } catch (err) {
      error = err.message;
    }
  }

  async function handleDeleteTeam(teamId) {
    if (!confirm('Delete this team?')) return;
    error = '';
    try {
      await deleteTeam(teamId);
      if (expandedTeamId === teamId) expandedTeamId = null;
      message = 'Team deleted.';
      await loadTeams();
    } catch (err) {
      error = err.message;
    }
  }

  function availableUsers(currentMembers) {
    const memberIds = new Set(currentMembers.map(m => m.id));
    return allUsers.filter(u => !memberIds.has(u.id));
  }
</script>

<div class="teams-page">
  <div class="header">
    <button class="back-btn" onclick={onBack}>← Back to Board</button>
    <h1>My Teams</h1>
  </div>

  {#if message}
    <p class="success">{message}</p>
  {/if}
  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="content">
    <div class="toolbar">
      <button class="create-btn" onclick={() => { showCreateForm = true; message = ''; }}>
        + Create Team
      </button>
    </div>

    {#if showCreateForm}
      <section class="form-section">
        <h2>New Team</h2>
        <form onsubmit={handleCreateTeam}>
          <div class="field">
            <label>Team Name</label>
            <input type="text" bind:value={newTeamName} required />
          </div>
          {#if createError}
            <p class="error">{createError}</p>
          {/if}
          <div class="form-actions">
            <button type="submit" class="save-btn">Create</button>
            <button type="button" class="cancel-btn" onclick={() => showCreateForm = false}>Cancel</button>
          </div>
        </form>
      </section>
    {/if}

    {#if loading}
      <p>Loading teams...</p>
    {:else if teams.length === 0 && !showCreateForm}
      <p class="empty">No teams yet. Create one to get started.</p>
    {:else}
      <div class="team-list">
        {#each teams as team (team.id)}
          <div class="team-card">
            <div class="team-header">
              {#if renamingTeamId === team.id}
                <form class="rename-form" onsubmit={handleRename}>
                  <input type="text" bind:value={renameValue} />
                  <button type="submit" class="save-btn small">Save</button>
                  <button type="button" class="cancel-btn small" onclick={() => renamingTeamId = null}>✕</button>
                </form>
              {:else}
                <button class="team-name" onclick={() => toggleExpand(team.id)}>
                  <span>{expandedTeamId === team.id ? '▼' : '▶'}</span>
                  {team.name}
                </button>
                <div class="team-actions">
                  <button onclick={() => startRename(team)}>Rename</button>
                  <button class="delete" onclick={() => handleDeleteTeam(team.id)}>Delete</button>
                </div>
              {/if}
            </div>

            {#if expandedTeamId === team.id}
              <div class="team-body">
                <h3>Members ({members.length})</h3>
                {#if members.length > 0}
                  <div class="member-list">
                    {#each members as member (member.id)}
                      <div class="member-row">
                        <span class="member-name">{member.name}</span>
                        <span class="member-email">{member.email}</span>
                        <button class="remove-btn" onclick={() => handleRemoveMember(member.id)}>✕</button>
                      </div>
                    {/each}
                  </div>
                {:else}
                  <p class="empty-members">No members yet.</p>
                {/if}

                {#if availableUsers(members).length > 0}
                  <div class="add-member">
                    <select bind:value={addUserId}>
                      <option value="">Select user...</option>
                      {#each availableUsers(members) as u}
                        <option value={u.id}>{u.name} ({u.email})</option>
                      {/each}
                    </select>
                    <button class="save-btn small" onclick={handleAddMember} disabled={!addUserId}>Add</button>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .teams-page { max-width: 600px; margin: 0 auto; padding: 24px; }
  .header { margin-bottom: 24px; }
  .back-btn { background: none; border: none; color: #4a90d9; cursor: pointer; font-size: 0.875rem; padding: 0; margin-bottom: 8px; }
  .back-btn:hover { text-decoration: underline; }
  h1 { font-size: 1.5rem; color: #333; margin: 0; }
  h2 { font-size: 1.1rem; color: #333; margin: 0 0 12px; }
  h3 { font-size: 0.9rem; color: #555; margin: 0 0 8px; }

  .toolbar { margin-bottom: 16px; }
  .create-btn { padding: 8px 16px; background: #4a90d9; color: white; border: none; border-radius: 4px; font-size: 0.875rem; cursor: pointer; }
  .create-btn:hover { background: #357abd; }

  .form-section { background: white; border: 1px solid #e0e0e0; border-radius: 6px; padding: 20px; margin-bottom: 16px; }
  .field { margin-bottom: 12px; }
  .field label { display: block; font-size: 0.8rem; font-weight: 500; color: #555; margin-bottom: 4px; }
  input[type="text"] { width: 100%; padding: 6px 10px; border: 1px solid #ccc; border-radius: 4px; font-size: 0.875rem; box-sizing: border-box; }
  input:focus { outline: none; border-color: #4a90d9; }

  .form-actions { display: flex; gap: 8px; }
  .save-btn { padding: 6px 16px; background: #4a90d9; color: white; border: none; border-radius: 4px; font-size: 0.875rem; cursor: pointer; }
  .save-btn:hover { background: #357abd; }
  .save-btn:disabled { opacity: 0.5; cursor: default; }
  .save-btn.small { padding: 4px 10px; font-size: 0.8rem; }
  .cancel-btn { padding: 6px 16px; background: none; border: 1px solid #ccc; border-radius: 4px; font-size: 0.875rem; cursor: pointer; color: #555; }
  .cancel-btn.small { padding: 4px 8px; font-size: 0.8rem; }

  .team-list { display: flex; flex-direction: column; gap: 8px; }
  .team-card { background: white; border: 1px solid #e0e0e0; border-radius: 6px; }

  .team-header { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px; }
  .team-name { background: none; border: none; font-size: 0.95rem; font-weight: 500; color: #333; cursor: pointer; display: flex; align-items: center; gap: 8px; padding: 0; }
  .team-name:hover { color: #4a90d9; }

  .team-actions { display: flex; gap: 6px; }
  .team-actions button { padding: 3px 8px; background: none; border: 1px solid #ddd; border-radius: 3px; font-size: 0.75rem; cursor: pointer; color: #555; }
  .team-actions button:hover { background: #f0f0f0; }
  .team-actions button.delete { color: #c00; border-color: #e0c0c0; }
  .team-actions button.delete:hover { background: #fff5f5; }

  .rename-form { display: flex; gap: 6px; align-items: center; flex: 1; }
  .rename-form input { flex: 1; }

  .team-body { padding: 0 16px 16px; border-top: 1px solid #eee; }
  .team-body h3 { margin-top: 12px; }

  .member-list { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; }
  .member-row { display: flex; align-items: center; gap: 8px; padding: 6px 8px; background: #f8f8f8; border-radius: 4px; }
  .member-name { font-size: 0.85rem; font-weight: 500; color: #333; }
  .member-email { font-size: 0.8rem; color: #888; flex: 1; }
  .remove-btn { padding: 2px 6px; background: none; border: 1px solid #e0e0e0; border-radius: 3px; cursor: pointer; color: #c00; font-size: 0.75rem; }
  .remove-btn:hover { background: #fff5f5; }

  .empty-members { font-size: 0.85rem; color: #888; margin: 4px 0 12px; }

  .add-member { display: flex; gap: 6px; align-items: center; }
  .add-member select { flex: 1; padding: 5px 8px; border: 1px solid #ccc; border-radius: 4px; font-size: 0.85rem; }

  .error { color: #c00; font-size: 0.85rem; margin: 4px 0; }
  .success { color: #0a0; font-size: 0.85rem; margin: 0 0 12px; }
  .empty { color: #888; font-size: 0.9rem; }
</style>
