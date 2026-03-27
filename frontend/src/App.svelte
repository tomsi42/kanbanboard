<script>
  let apiStatus = $state('checking...');

  async function checkHealth() {
    try {
      const response = await fetch('/api/v1/health');
      const data = await response.json();
      apiStatus = data.status === 'ok' ? 'Connected' : 'Error';
    } catch {
      apiStatus = 'Disconnected';
    }
  }

  $effect(() => {
    checkHealth();
  });
</script>

<main>
  <h1>Kanban Board</h1>
  <p>API status: <span class="status" class:connected={apiStatus === 'Connected'}>{apiStatus}</span></p>
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    font-family: system-ui, -apple-system, sans-serif;
  }

  h1 {
    font-size: 2rem;
    color: #333;
  }

  .status {
    font-weight: bold;
    color: #c00;
  }

  .status.connected {
    color: #0a0;
  }
</style>
