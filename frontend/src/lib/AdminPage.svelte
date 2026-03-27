<script>
  import { adminListUsers, adminCreateUser, adminUpdateUser, adminResetPassword } from './api.js';
  import { validatePassword } from './validate.js';

  let { onBack } = $props();

  let users = $state([]);
  let loading = $state(true);
  let error = $state('');
  let message = $state('');

  // Create user form
  let showCreateForm = $state(false);
  let newName = $state('');
  let newEmail = $state('');
  let newPassword = $state('');
  let newIsAdmin = $state(false);
  let newIsTeamManager = $state(false);
  let createError = $state('');

  // Edit user
  let editingUser = $state(null);
  let editName = $state('');
  let editEmail = $state('');
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

    if (!newName.trim() || !newEmail.trim() || !newPassword) {
      createError = 'All fields are required.';
      return;
    }

    const pwError = validatePassword(newPassword);
    if (pwError) { createError = pwError; return; }

    try {
      await adminCreateUser({
        name: newName.trim(),
        email: newEmail.trim(),
        password: newPassword,
        isAdmin: newIsAdmin,
        isTeamManager: newIsTeamManager,
      });
      showCreateForm = false;
      newName = ''; newEmail = ''; newPassword = '';
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
    editEmail = user.email;
    editIsActive = user.isActive;
    editIsAdmin = user.isAdmin;
    editIsTeamManager = user.isTeamManager;
    editError = '';
  }

  async function handleSaveEdit(e) {
    e.preventDefault();
    editError = '';

    if (!editName.trim() || !editEmail.trim()) {
      editError = 'Name and email are required.';
      return;
    }

    try {
      await adminUpdateUser(editingUser.id, {
        name: editName.trim(),
        email: editEmail.trim(),
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
              <label>Email</label>
              <input type="email" bind:value={newEmail} required />
            </div>
          </div>
          <div class="field">
            <label>Password</label>
            <input type="password" bind:value={newPassword} required />
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
              <input type="email" bind:value={editEmail} required />
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
            <th>Email</th>
            <th>Roles</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each users as user (user.id)}
            <tr class:inactive={!user.isActive}>
              <td>{user.name}</td>
              <td>{user.email}</td>
              <td>
                {#if user.isAdmin}<span class="badge admin">Admin</span>{/if}
                {#if user.isTeamManager}<span class="badge tm">Team Mgr</span>{/if}
              </td>
              <td>
                <span class="status" class:active={user.isActive}>
                  {user.isActive ? 'Active' : 'Inactive'}
                </span>
              </td>
              <td class="actions">
                <button onclick={() => startEdit(user)}>Edit</button>
                <button onclick={() => { resetUserId = user.id; resetPassword = ''; resetError = ''; message = ''; }}>Reset PW</button>
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

  h1 { font-size: 1.5rem; color: #333; margin: 0; }
  h2 { font-size: 1.1rem; color: #333; margin: 0 0 12px; }

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

  .error { color: #c00; font-size: 0.85rem; margin: 4px 0; }
  .success { color: #0a0; font-size: 0.85rem; margin: 0 0 12px; }
</style>
