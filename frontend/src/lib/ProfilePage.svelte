<script>
  import { updateProfile, changePassword } from './api.js';
  import { validatePassword } from './validate.js';

  let { user, onBack, onProfileUpdated } = $props();

  // Profile fields
  let name = $state(user.name);
  let email = $state(user.email);
  let profileMessage = $state('');
  let profileError = $state('');
  let savingProfile = $state(false);

  // Password fields
  let currentPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let passwordMessage = $state('');
  let passwordError = $state('');
  let savingPassword = $state(false);

  async function handleSaveProfile(e) {
    e.preventDefault();
    profileMessage = '';
    profileError = '';

    if (!name.trim() || !email.trim()) {
      profileError = 'Name and email are required.';
      return;
    }

    savingProfile = true;
    try {
      const updated = await updateProfile({ name: name.trim(), email: email.trim() });
      profileMessage = 'Profile updated.';
      onProfileUpdated(updated);
    } catch (err) {
      profileError = err.message;
    } finally {
      savingProfile = false;
    }
  }

  async function handleChangePassword(e) {
    e.preventDefault();
    passwordMessage = '';
    passwordError = '';

    if (!currentPassword || !newPassword) {
      passwordError = 'All password fields are required.';
      return;
    }

    if (newPassword !== confirmPassword) {
      passwordError = 'New passwords do not match.';
      return;
    }

    const validationMsg = validatePassword(newPassword);
    if (validationMsg) {
      passwordError = validationMsg;
      return;
    }

    savingPassword = true;
    try {
      await changePassword({ currentPassword, newPassword });
      passwordMessage = 'Password changed successfully.';
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
    } catch (err) {
      passwordError = err.message;
    } finally {
      savingPassword = false;
    }
  }
</script>

<div class="profile-page">
  <div class="header">
    <button class="back-btn" onclick={onBack}>← Back to Board</button>
    <h1>My Profile</h1>
  </div>

  <div class="content">
    <section>
      <h2>Profile Information</h2>
      <form onsubmit={handleSaveProfile}>
        <div class="field">
          <label for="name">Name</label>
          <input id="name" type="text" bind:value={name} required />
        </div>
        <div class="field">
          <label for="email">Email</label>
          <input id="email" type="email" bind:value={email} required />
        </div>
        {#if profileError}
          <p class="error">{profileError}</p>
        {/if}
        {#if profileMessage}
          <p class="success">{profileMessage}</p>
        {/if}
        <button type="submit" class="save-btn" disabled={savingProfile}>
          {savingProfile ? 'Saving...' : 'Save Profile'}
        </button>
      </form>
    </section>

    <section>
      <h2>Change Password</h2>
      <form onsubmit={handleChangePassword}>
        <div class="field">
          <label for="currentPassword">Current Password</label>
          <input id="currentPassword" type="password" bind:value={currentPassword} required />
        </div>
        <div class="field">
          <label for="newPassword">New Password</label>
          <input id="newPassword" type="password" bind:value={newPassword} required />
        </div>
        <div class="field">
          <label for="confirmPassword">Confirm New Password</label>
          <input id="confirmPassword" type="password" bind:value={confirmPassword} required />
        </div>
        {#if passwordError}
          <p class="error">{passwordError}</p>
        {/if}
        {#if passwordMessage}
          <p class="success">{passwordMessage}</p>
        {/if}
        <button type="submit" class="save-btn" disabled={savingPassword}>
          {savingPassword ? 'Changing...' : 'Change Password'}
        </button>
      </form>
    </section>
  </div>
</div>

<style>
  .profile-page {
    max-width: 500px;
    margin: 0 auto;
    padding: 24px;
  }

  .header {
    margin-bottom: 32px;
  }

  .back-btn {
    background: none;
    border: none;
    color: #4a90d9;
    cursor: pointer;
    font-size: 0.875rem;
    padding: 0;
    margin-bottom: 8px;
  }

  .back-btn:hover {
    text-decoration: underline;
  }

  h1 {
    font-size: 1.5rem;
    color: #333;
    margin: 0;
  }

  h2 {
    font-size: 1.1rem;
    color: #333;
    margin: 0 0 16px;
  }

  section {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 24px;
  }

  .field {
    margin-bottom: 12px;
  }

  label {
    display: block;
    font-size: 0.8rem;
    font-weight: 500;
    color: #555;
    margin-bottom: 4px;
  }

  input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.9rem;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #4a90d9;
    box-shadow: 0 0 0 2px rgba(74, 144, 217, 0.2);
  }

  .error {
    color: #c00;
    font-size: 0.85rem;
    margin: 8px 0;
  }

  .success {
    color: #0a0;
    font-size: 0.85rem;
    margin: 8px 0;
  }

  .save-btn {
    padding: 8px 16px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    margin-top: 4px;
  }

  .save-btn:hover:not(:disabled) {
    background: #357abd;
  }

  .save-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
