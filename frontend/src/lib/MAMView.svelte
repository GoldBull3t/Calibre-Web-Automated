<script>
  import { onMount } from 'svelte';
  export let mamStatus = null;

  let query = '';
  let category = 'all';
  let results = [];
  let loading = false;
  let mamCookieInput = '';
  let savingConfig = false;
  let configMessage = '';
  let downloadingId = null;
  let downloadMessage = '';

  onMount(() => {
    loadStatus();
  });

  async function loadStatus() {
    try {
      const res = await fetch('/api/mam/status');
      mamStatus = await res.json();
    } catch (e) {
      console.error(e);
    }
  }

  async function saveMamConfig() {
    if (!mamCookieInput.trim()) return;
    savingConfig = true;
    try {
      const res = await fetch('/api/mam/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mam_id: mamCookieInput.trim() })
      });
      const data = await res.json();
      if (data.status === 'success') {
        configMessage = 'Configuração gravada com sucesso!';
        await loadStatus();
      } else {
        configMessage = 'Erro: ' + (data.error || 'Falha ao validar cookie');
      }
    } catch (e) {
      configMessage = 'Erro de comunicação com o servidor.';
    } finally {
      savingConfig = false;
    }
  }

  async function searchMAM() {
    if (!query.trim()) return;
    loading = true;
    downloadMessage = '';
    try {
      const res = await fetch(`/api/mam/search?q=${encodeURIComponent(query.trim())}`);
      const data = await res.json();
      results = data.results || [];
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function onKeyDown(e) {
    if (e.key === 'Enter') searchMAM();
  }

  async function downloadTorrent(id, title) {
    downloadingId = id;
    try {
      const res = await fetch(`/api/mam/download/${id}`, { method: 'POST' });
      const data = await res.json();
      if (data.status === 'success') {
        downloadMessage = `✓ "${title}" enviado para a pasta de ingestão contínua!`;
      } else {
        downloadMessage = 'Erro ao transferir torrent: ' + (data.error || 'Falha');
      }
    } catch (e) {
      downloadMessage = 'Erro de rede ao solicitar download.';
    } finally {
      downloadingId = null;
    }
  }
</script>

<div class="mam-container">
  <!-- Hero Card -->
  <div class="mam-hero glass">
    <div class="mam-header-row">
      <div class="mam-title-box">
        <div class="cheese-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
        </div>
        <div>
          <h2>The Friendliest & Warmest Place on the Internet</h2>
          <p>Catálogo integrado de eBooks, Audiobooks e Banda Desenhada com 1-Click Ingest.</p>
        </div>
      </div>

      {#if mamStatus && mamStatus.configured && mamStatus.account}
        <!-- Account Status Pills -->
        <div class="stats-pills">
          <div class="pill">
            <span class="pill-k">Rácio</span>
            <span class="pill-v" style="color: #22c55e;">{mamStatus.account.ratio}</span>
          </div>
          <div class="pill">
            <span class="pill-k">VIP</span>
            <span class="pill-v">{mamStatus.account.vip ? 'Sim ★' : 'Não'}</span>
          </div>
          <div class="pill">
            <span class="pill-k">Bonus Points</span>
            <span class="pill-v">{mamStatus.account.bonus_points}</span>
          </div>
          <div class="pill">
            <span class="pill-k">Seeding</span>
            <span class="pill-v">{mamStatus.account.seeding}</span>
          </div>
        </div>
      {/if}
    </div>

    <!-- MAM Configuration if not configured -->
    {#if !mamStatus || !mamStatus.configured}
      <div class="config-box">
        <h4>Configuração do Cookie MAM</h4>
        <p>Insere o teu valor da cookie <code>mam_id</code> da tua conta MyAnonamouse para ativar a pesquisa e downloads automáticos:</p>
        <div class="config-inputs">
          <input 
            type="password" 
            placeholder="mam_id cookie..." 
            bind:value={mamCookieInput} 
          />
          <button class="btn btn-primary" on:click={saveMamConfig} disabled={savingConfig}>
            {savingConfig ? 'A guardar...' : 'Guardar & Conectar'}
          </button>
        </div>
        {#if configMessage}
          <div class="config-status">{configMessage}</div>
        {/if}
      </div>
    {:else}
      <!-- Search Input -->
      <div class="search-row">
        <input 
          type="text" 
          placeholder="Pesquisar por título, autor no catálogo MAM..." 
          bind:value={query}
          on:keydown={onKeyDown}
        />
        <button class="btn btn-primary" on:click={searchMAM} disabled={loading}>
          {loading ? 'A pesquisar...' : 'Pesquisar MAM'}
        </button>
      </div>
    {/if}
  </div>

  {#if downloadMessage}
    <div class="alert-box glass">
      <span>{downloadMessage}</span>
    </div>
  {/if}

  <!-- Results Section -->
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>A consultar API MAM...</p>
    </div>
  {:else if results.length > 0}
    <div class="results-table-wrap glass">
      <table class="mam-table">
        <thead>
          <tr>
            <th>Título & Autor</th>
            <th>Tamanho</th>
            <th>S / L</th>
            <th>VIP</th>
            <th>Ação</th>
          </tr>
        </thead>
        <tbody>
          {#each results as r}
            <tr>
              <td>
                <div class="torrent-title">{r.title}</div>
                <div class="torrent-author">por {r.author || 'MAM Upload'}</div>
              </td>
              <td><span class="size-tag">{r.size}</span></td>
              <td>
                <span class="seeders">▲ {r.seeders}</span> / <span class="leechers">▼ {r.leechers}</span>
              </td>
              <td>
                {#if r.vip}
                  <span class="badge badge-warning">VIP</span>
                {:else}
                  <span class="badge badge-accent">FREE</span>
                {/if}
              </td>
              <td>
                <button 
                  class="btn btn-primary btn-sm" 
                  on:click={() => downloadTorrent(r.id, r.title)}
                  disabled={downloadingId === r.id}>
                  {#if downloadingId === r.id}
                    <span>A enviar...</span>
                  {:else}
                    <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                    <span>1-Click Ingest</span>
                  {/if}
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .mam-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .mam-hero {
    padding: 28px;
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .mam-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
  }

  .mam-title-box {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .cheese-icon {
    font-size: 36px;
  }

  .mam-title-box h2 {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .mam-title-box p {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .stats-pills {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .pill {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 6px 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .pill-k {
    font-size: 10px;
    text-transform: uppercase;
    color: var(--text-muted);
    font-weight: 600;
  }

  .pill-v {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .config-box {
    padding: 20px;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid var(--border-secondary);
    border-radius: var(--radius-md);
  }

  .config-box h4 {
    margin-bottom: 6px;
    font-size: 14px;
    color: var(--accent);
  }

  .config-box p {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 12px;
  }

  .config-inputs {
    display: flex;
    gap: 10px;
    max-width: 500px;
  }

  .config-inputs input {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 8px 12px;
    color: var(--text-primary);
    outline: none;
  }

  .config-status {
    margin-top: 8px;
    font-size: 12px;
    color: var(--accent);
  }

  .search-row {
    display: flex;
    gap: 10px;
    max-width: 700px;
  }

  .search-row input {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 10px 14px;
    color: var(--text-primary);
    font-size: 13px;
    outline: none;
  }

  .search-row input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }

  .alert-box {
    padding: 12px 18px;
    border-radius: var(--radius-md);
    background: rgba(34, 197, 94, 0.12);
    border: 1px solid rgba(34, 197, 94, 0.3);
    color: #22c55e;
    font-weight: 500;
    font-size: 13px;
  }

  .results-table-wrap {
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .mam-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
    text-align: left;
  }

  .mam-table th {
    padding: 12px 16px;
    background: rgba(255, 255, 255, 0.03);
    color: var(--text-muted);
    font-size: 11px;
    text-transform: uppercase;
    font-weight: 600;
    border-bottom: 1px solid var(--border-primary);
  }

  .mam-table td {
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-primary);
    color: var(--text-secondary);
  }

  .torrent-title {
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 2px;
  }

  .torrent-author {
    font-size: 11px;
    color: var(--text-muted);
  }

  .size-tag {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .seeders {
    color: #22c55e;
    font-weight: 600;
  }

  .leechers {
    color: #f87171;
  }

  .loading-state {
    text-align: center;
    padding: 60px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    color: var(--text-secondary);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(56, 189, 248, 0.15);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>

