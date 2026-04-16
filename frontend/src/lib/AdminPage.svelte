<script>
  import { adminListUsers, adminCreateUser, adminUpdateUser, adminResetPassword, adminGetDeleteImpact, adminDeleteUser, adminGetSettings, adminUpdateSettings } from './api.js';
  import { validatePassword } from './validate.js';

  let { onBack, currentUserId = '', onUserDeleted } = $props();

  let users = $state([]);
  let loading = $state(true);
  let error = $state('');
  let message = $state('');

  // Settings
  let registrationEnabled = $state(false);
  let settingsLoading = $state(true);

  async function loadSettings() {
    try {
      const s = await adminGetSettings();
      registrationEnabled = s.registrationEnabled;
    } catch {
      // non-fatal
    } finally {
      settingsLoading = false;
    }
  }

  async function toggleRegistration() {
    try {
      const s = await adminUpdateSettings({ registrationEnabled: !registrationEnabled });
      registrationEnabled = s.registrationEnabled;
      message = `Registration ${registrationEnabled ? 'enabled' : 'disabled'}.`;
    } catch (err) {
      error = err.message;
    }
  }

  $effect(() => { loadSettings(); });

  // Create user form
  let showCreateForm = $state(false);
  let newName = $state('');
  let newEmail = $state('');
  let newUsername = $state('');
  let newPassword = $state('');
  let newIsAdmin = $state(false);
  let newIsTeamManager = $state(false);
  let createError = $state('');

  // Edit user
  let editingUser = $state(null);
  let editName = $state('');
  let editEmail = $state('');
  let editUsername = $state('');
  let editIsActive = $state(true);
  let editIsAdmin = $state(false);
  let editIsTeamManager = $state(false);
  let editError = $state('');

  // Reset password
  let resetUserId = $state(null);
  let resetPassword = $state('');
  let resetError = $state('');

  async function loadUsers() {
    loading = true;
    try {
      users = await adminListUsers();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  $effect(() => { loadUsers(); });

  async function handleCreateUser(e) {
    e.preventDefault();
    createError = '';

    if (!newName.trim() || !newPassword) {
      createError = 'Name and password are required.';
      return;
    }
    if (!newEmail.trim() && !newUsername.trim()) {
      createError = 'At least one of email or username is required.';
      return;
    }

    const pwError = validatePassword(newPassword);
    if (pwError) { createError = pwError; return; }

    try {
      await adminCreateUser({
        name: newName.trim(),
        email: newEmail.trim(),
        username: newUsername.trim(),
        password: newPassword,
        isAdmin: newIsAdmin,
        isTeamManager: newIsTeamManager,
      });
      showCreateForm = false;
      newName = ''; newEmail = ''; newUsername = ''; newPassword = '';
      newIsAdmin = false; newIsTeamManager = false;
      message = 'User created.';
      await loadUsers();
    } catch (err) {
      createError = err.message;
    }
  }

  function startEdit(user) {
    editingUser = user;
    editName = user.name;
    editEmail = user.email || '';
    editUsername = user.username || '';
    editIsActive = user.isActive;
    editIsAdmin = user.isAdmin;
    editIsTeamManager = user.isTeamManager;
    editError = '';
  }

  async function handleSaveEdit(e) {
    e.preventDefault();
    editError = '';

    if (!editName.trim()) {
      editError = 'Name is required.';
      return;
    }
    if (!editEmail.trim() && !editUsername.trim()) {
      editError = 'At least one of email or username is required.';
      return;
    }

    try {
      await adminUpdateUser(editingUser.id, {
        name: editName.trim(),
        email: editEmail.trim(),
        username: editUsername.trim(),
        isActive: editIsActive,
        isAdmin: editIsAdmin,
        isTeamManager: editIsTeamManager,
      });
      editingUser = null;
      message = 'User updated.';
      await loadUsers();
    } catch (err) {
      editError = err.message;
    }
  }

  async function handleResetPassword(e) {
    e.preventDefault();
    resetError = '';

    if (!resetPassword) { resetError = 'Password is required.'; return; }
    const pwError = validatePassword(resetPassword);
    if (pwError) { resetError = pwError; return; }

    try {
      await adminResetPassword(resetUserId, resetPassword);
      resetUserId = null;
      resetPassword = '';
      message = 'Password reset.';
    } catch (err) {
      resetError = err.message;
    }
  }

  async function handleDeleteUser(user) {
    try {
      const impact = await adminGetDeleteImpact(user.id);

      let msg = `Delete user "${user.name}"?\n\n`;
      if (impact.projectCount > 0) {
        msg += `${impact.projectCount} project(s) with ${impact.taskCount} task(s) will be permanently deleted.\n`;
      }
      if (impact.teamTransfers && impact.teamTransfers.length > 0) {
        msg += `\nTeam ownership transfers:\n`;
        for (const t of impact.teamTransfers) {
          msg += `  ${t.teamName} → ${t.newOwner}\n`;
        }
      }
      msg += `\nThis cannot be undone.`;

      if (!confirm(msg)) return;

      await adminDeleteUser(user.id);
      message = `User "${user.name}" has been deleted.`;
      await loadUsers();
      onUserDeleted?.();
    } catch (err) {
      error = err.message;
    }
  }
</script>

<div class="admin-page">
  <div class="header">
    <button class="back-btn" onclick={onBack}>← Back to Board</button>
    <h1>Admin — User Management</h1>
  </div>

  {#if message}
    <p class="success">{message}</p>
  {/if}
  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="content">
    <section class="settings-section">
      <h2>Settings</h2>
      {#if !settingsLoading}
        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-label">Open Registration</span>
            <span class="setting-desc">Allow anyone with the link to create an account</span>
          </div>
          <button
            class="toggle-btn"
            class:on={registrationEnabled}
            onclick={toggleRegistration}
          >
            {registrationEnabled ? 'On' : 'Off'}
          </button>
        </div>
      {/if}
    </section>

    <div class="toolbar">
      <button class="create-btn" onclick={() => { showCreateForm = true; message = ''; }}>
        + Create User
      </button>
    </div>

    {#if showCreateForm}
      <section class="form-section">
        <h2>New User</h2>
        <form onsubmit={handleCreateUser}>
          <div class="form-row">
            <div class="field">
              <label>Name</label>
              <input type="text" bind:value={newName} required />
            </div>
            <div class="field">
              <label>Email <span class="optional">(or username)</span></label>
              <input type="email" bind:value={newEmail} />
            </div>
          </div>
          <div class="form-row">
            <div class="field">
              <label>Username <span class="optional">(or email)</span></label>
              <input type="text" bind:value={newUsername} placeholder="e.g. agent_coder" />
            </div>
            <div class="field">
              <label>Password</label>
              <input type="password" bind:value={newPassword} required />
            </div>
          </div>
          <div class="checkbox-row">
            <label><input type="checkbox" bind:checked={newIsAdmin} /> Admin</label>
            <label><input type="checkbox" bind:checked={newIsTeamManager} /> Team Manager</label>
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

    {#if editingUser}
      <section class="form-section">
        <h2>Edit: {editingUser.name}</h2>
        <form onsubmit={handleSaveEdit}>
          <div class="form-row">
            <div class="field">
              <label>Name</label>
              <input type="text" bind:value={editName} required />
            </div>
            <div class="field">
              <label>Email</label>
              <input type="email" bind:value={editEmail} />
            </div>
          </div>
          <div class="form-row">
            <div class="field">
              <label>Username</label>
              <input type="text" bind:value={editUsername} placeholder="e.g. agent_coder" />
            </div>
          </div>
          <div class="checkbox-row">
            <label><input type="checkbox" bind:checked={editIsActive} /> Active</label>
            <label><input type="checkbox" bind:checked={editIsAdmin} /> Admin</label>
            <label><input type="checkbox" bind:checked={editIsTeamManager} /> Team Manager</label>
          </div>
          {#if editError}
            <p class="error">{editError}</p>
          {/if}
          <div class="form-actions">
            <button type="submit" class="save-btn">Save</button>
            <button type="button" class="cancel-btn" onclick={() => editingUser = null}>Cancel</button>
          </div>
        </form>
      </section>
    {/if}

    {#if resetUserId}
      <section class="form-section">
        <h2>Reset Password</h2>
        <form onsubmit={handleResetPassword}>
          <div class="field">
            <label>New Password</label>
            <input type="password" bind:value={resetPassword} required />
          </div>
          {#if resetError}
            <p class="error">{resetError}</p>
          {/if}
          <div class="form-actions">
            <button type="submit" class="save-btn">Reset</button>
            <button type="button" class="cancel-btn" onclick={() => { resetUserId = null; resetPassword = ''; }}>Cancel</button>
          </div>
        </form>
      </section>
    {/if}

    {#if loading}
      <p>Loading users...</p>
    {:else}
      <table class="user-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email / Username</th>
            <th>Roles</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each users as user (user.id)}
            <tr class:inactive={!user.isActive} class:deleted={user.deletedAt}>
              <td>
                {user.name}
                {#if user.deletedAt}<span class="badge deleted">Deleted</span>{/if}
              </td>
              <td>
                {#if user.email}<div>{user.email}</div>{/if}
                {#if user.username}<div class="username-cell">@{user.username}</div>{/if}
              </td>
              <td>
                {#if user.isAdmin}<span class="badge admin">Admin</span>{/if}
                {#if user.isTeamManager}<span class="badge tm">Team Mgr</span>{/if}
              </td>
              <td>
                {#if user.deletedAt}
                  <span class="status deleted-status">Deleted</span>
                {:else}
                  <span class="status" class:active={user.isActive}>
                    {user.isActive ? 'Active' : 'Inactive'}
                  </span>
                {/if}
              </td>
              <td class="actions">
                {#if !user.deletedAt}
                  <button onclick={() => startEdit(user)}>Edit</button>
                  <button onclick={() => { resetUserId = user.id; resetPassword = ''; resetError = ''; message = ''; }}>Reset PW</button>
                  {#if user.id !== currentUserId}
                    <button class="delete-user-btn" onclick={() => handleDeleteUser(user)}>Delete</button>
                  {/if}
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .admin-page {
    max-width: 800px;
    margin: 0 auto;
    padding: 24px;
  }

  .header { margin-bottom: 24px; }

  .back-btn {
    background: none; border: none; color: #4a90d9;
    cursor: pointer; font-size: 0.875rem; padding: 0; margin-bottom: 8px;
  }
  .back-btn:hover { text-decoration: underline; }

  .optional { font-weight: 400; color: #999; font-size: 0.75rem; }

  .username-cell { font-size: 0.8rem; color: #666; font-family: monospace; }

  h1 { font-size: 1.5rem; color: #333; margin: 0; }
  h2 { font-size: 1.1rem; color: #333; margin: 0 0 12px; }

  .settings-section {
    background: white; border: 1px solid #e0e0e0; border-radius: 6px;
    padding: 16px 20px; margin-bottom: 20px;
  }

  .setting-row {
    display: flex; align-items: center; justify-content: space-between; gap: 16px;
  }

  .setting-info {
    display: flex; flex-direction: column; gap: 2px;
  }

  .setting-label {
    font-size: 0.875rem; font-weight: 500; color: #333;
  }

  .setting-desc {
    font-size: 0.8rem; color: #888;
  }

  .toggle-btn {
    padding: 5px 16px; border-radius: 4px; font-size: 0.875rem;
    font-weight: 500; cursor: pointer; border: 1px solid #ccc;
    background: #f0f0f0; color: #666; min-width: 52px;
  }
  .toggle-btn.on {
    background: #4a90d9; color: white; border-color: #4a90d9;
  }
  .toggle-btn:hover { opacity: 0.85; }

  .toolbar { margin-bottom: 16px; }

  .create-btn {
    padding: 8px 16px; background: #4a90d9; color: white;
    border: none; border-radius: 4px; font-size: 0.875rem; cursor: pointer;
  }
  .create-btn:hover { background: #357abd; }

  .form-section {
    background: white; border: 1px solid #e0e0e0; border-radius: 6px;
    padding: 20px; margin-bottom: 16px;
  }

  .form-row { display: flex; gap: 12px; }
  .form-row .field { flex: 1; }
  .field { margin-bottom: 12px; }
  .field label { display: block; font-size: 0.8rem; font-weight: 500; color: #555; margin-bottom: 4px; }

  input[type="text"], input[type="email"], input[type="password"] {
    width: 100%; padding: 6px 10px; border: 1px solid #ccc;
    border-radius: 4px; font-size: 0.875rem; box-sizing: border-box;
  }
  input:focus { outline: none; border-color: #4a90d9; }

  .checkbox-row {
    display: flex; gap: 16px; margin-bottom: 12px;
  }
  .checkbox-row label {
    display: flex; align-items: center; gap: 4px;
    font-size: 0.875rem; color: #555; cursor: pointer;
  }

  .form-actions { display: flex; gap: 8px; }

  .save-btn {
    padding: 6px 16px; background: #4a90d9; color: white;
    border: none; border-radius: 4px; font-size: 0.875rem; cursor: pointer;
  }
  .save-btn:hover { background: #357abd; }

  .cancel-btn {
    padding: 6px 16px; background: none; border: 1px solid #ccc;
    border-radius: 4px; font-size: 0.875rem; cursor: pointer; color: #555;
  }

  .user-table {
    width: 100%; border-collapse: collapse;
    background: white; border: 1px solid #e0e0e0; border-radius: 6px;
  }

  th, td {
    padding: 10px 12px; text-align: left; font-size: 0.875rem;
    border-bottom: 1px solid #eee;
  }
  th { font-weight: 600; color: #555; background: #f8f8f8; }

  tr.inactive { opacity: 0.5; }
  tr.deleted { opacity: 0.4; background: #f8f8f8; }

  .badge {
    display: inline-block; padding: 1px 6px; border-radius: 3px;
    font-size: 0.7rem; font-weight: 500; margin-right: 4px;
  }
  .badge.admin { background: #e8f0fe; color: #1a73e8; }
  .badge.tm { background: #e6ffe6; color: #0a0; }

  .status { font-size: 0.8rem; }
  .status.active { color: #0a0; }

  .actions { display: flex; gap: 6px; }
  .actions button {
    padding: 3px 8px; background: none; border: 1px solid #ddd;
    border-radius: 3px; font-size: 0.75rem; cursor: pointer; color: #555;
  }
  .actions button:hover { background: #f0f0f0; }
  .delete-user-btn { color: #c00 !important; border-color: #e0c0c0 !important; }
  .delete-user-btn:hover { background: #fff5f5 !important; }

  .badge.deleted { background: #f0e0e0; color: #c00; }
  .deleted-status { color: #c00; }

  .error { color: #c00; font-size: 0.85rem; margin: 4px 0; }
  .success { color: #0a0; font-size: 0.85rem; margin: 0 0 12px; }
</style>
