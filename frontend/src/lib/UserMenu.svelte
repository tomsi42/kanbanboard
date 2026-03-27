<script>
  let { userName, isAdmin = false, onProfile, onAdmin, onLogout } = $props();

  let open = $state(false);

  function toggle() {
    open = !open;
  }

  function handleProfile() {
    open = false;
    onProfile();
  }

  function handleLogout() {
    open = false;
    onLogout();
  }

  function handleClickOutside(e) {
    if (!e.target.closest('.user-menu')) {
      open = false;
    }
  }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="user-menu">
  <button class="trigger" onclick={toggle}>
    <span class="avatar">{userName.charAt(0).toUpperCase()}</span>
    <span class="name">{userName}</span>
    <span class="arrow">{open ? '▲' : '▼'}</span>
  </button>

  {#if open}
    <div class="menu">
      <button class="menu-item" onclick={handleProfile}>My Profile</button>
      {#if isAdmin}
        <button class="menu-item" onclick={() => { open = false; onAdmin(); }}>Admin</button>
      {/if}
      <hr />
      <button class="menu-item logout" onclick={handleLogout}>Sign Out</button>
    </div>
  {/if}
</div>

<style>
  .user-menu {
    position: relative;
  }

  .trigger {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 10px;
    background: none;
    border: 1px solid #ccc;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    color: #333;
  }

  .trigger:hover {
    background: #f5f5f5;
  }

  .avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: #4a90d9;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .name {
    color: #555;
  }

  .arrow {
    font-size: 0.6rem;
    color: #888;
  }

  .menu {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 4px;
    min-width: 160px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 100;
  }

  .menu-item {
    display: block;
    width: 100%;
    padding: 8px 12px;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    font-size: 0.875rem;
    color: #333;
  }

  .menu-item:hover {
    background: #f0f4ff;
  }

  .logout {
    color: #c00;
  }

  hr {
    margin: 4px 0;
    border: none;
    border-top: 1px solid #eee;
  }
</style>
