<script>
  import { login, register } from './api.js';
  import { validatePassword } from './validate.js';

  let { appTitle = 'Kanban Board', registrationEnabled = false, onLogin } = $props();

  let mode = $state('login'); // 'login' | 'register'

  let name = $state('');
  let loginValue = $state('');
  let email = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let error = $state('');
  let submitting = $state(false);

  function switchMode(m) {
    mode = m;
    error = '';
    name = '';
    loginValue = '';
    email = '';
    password = '';
    confirmPassword = '';
  }

  async function handleLogin(e) {
    e.preventDefault();
    error = '';

    if (!loginValue || !password) {
      error = 'Username/email and password are required.';
      return;
    }

    submitting = true;
    try {
      const user = await login(loginValue, password);
      onLogin(user);
    } catch (err) {
      error = err.message;
    } finally {
      submitting = false;
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    error = '';

    if (!name || !email || !password) {
      error = 'All fields are required.';
      return;
    }

    if (password !== confirmPassword) {
      error = 'Passwords do not match.';
      return;
    }

    const pwError = validatePassword(password);
    if (pwError) { error = pwError; return; }

    submitting = true;
    try {
      const user = await register(name, email, password);
      onLogin(user);
    } catch (err) {
      error = err.message;
    } finally {
      submitting = false;
    }
  }
</script>

<div class="login">
  <h1>{appTitle}</h1>

  {#if mode === 'login'}
    <form onsubmit={handleLogin}>
      <div class="field">
        <label for="login">Username or email</label>
        <input id="login" type="text" bind:value={loginValue} required autocomplete="username" />
      </div>

      <div class="field">
        <label for="password">Password</label>
        <input id="password" type="password" bind:value={password} required />
      </div>

      {#if error}
        <p class="error">{error}</p>
      {/if}

      <button type="submit" disabled={submitting}>
        {submitting ? 'Signing in...' : 'Sign In'}
      </button>
    </form>

    {#if registrationEnabled}
      <p class="switch">
        No account? <button class="link-btn" onclick={() => switchMode('register')}>Create one</button>
      </p>
    {/if}

  {:else}
    <form onsubmit={handleRegister}>
      <div class="field">
        <label for="reg-name">Name</label>
        <input id="reg-name" type="text" bind:value={name} required />
      </div>

      <div class="field">
        <label for="reg-email">Email</label>
        <input id="reg-email" type="email" bind:value={email} required />
      </div>

      <div class="field">
        <label for="reg-password">Password</label>
        <input id="reg-password" type="password" bind:value={password} required />
      </div>

      <div class="field">
        <label for="reg-confirm">Confirm Password</label>
        <input id="reg-confirm" type="password" bind:value={confirmPassword} required />
      </div>

      {#if error}
        <p class="error">{error}</p>
      {/if}

      <button type="submit" disabled={submitting}>
        {submitting ? 'Creating account...' : 'Create Account'}
      </button>
    </form>

    <p class="switch">
      Already have an account? <button class="link-btn" onclick={() => switchMode('login')}>Sign in</button>
    </p>
  {/if}
</div>

<style>
  .login {
    max-width: 360px;
    margin: 0 auto;
    padding: 48px 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    min-height: 100vh;
    justify-content: center;
  }

  h1 {
    font-size: 1.75rem;
    color: #333;
    margin: 0 0 32px;
  }

  form {
    width: 100%;
  }

  .field {
    margin-bottom: 16px;
  }

  label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #555;
    margin-bottom: 4px;
  }

  input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 1rem;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #4a90d9;
    box-shadow: 0 0 0 2px rgba(74, 144, 217, 0.2);
  }

  .error {
    color: #c00;
    font-size: 0.875rem;
    margin: 12px 0;
  }

  button[type="submit"] {
    width: 100%;
    padding: 10px;
    background: #4a90d9;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    margin-top: 8px;
  }

  button[type="submit"]:hover:not(:disabled) {
    background: #357abd;
  }

  button[type="submit"]:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .switch {
    margin-top: 20px;
    font-size: 0.875rem;
    color: #666;
  }

  .link-btn {
    background: none;
    border: none;
    color: #4a90d9;
    cursor: pointer;
    font-size: 0.875rem;
    padding: 0;
    text-decoration: underline;
  }

  .link-btn:hover {
    color: #357abd;
  }
</style>
