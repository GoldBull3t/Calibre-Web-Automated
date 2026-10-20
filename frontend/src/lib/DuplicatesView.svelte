<script>
  // @ts-nocheck
  import { onMount } from 'svelte';

  let groups = [];
  let loading = true;
  let scanning = false;
  let resolvingHash = null;
  let message = '';

  onMount(loadDuplicates);

  async function loadDuplicates() {
    loading = true;
    try {
      const res = await fetch('/api/duplicates');
      const data = await res.json();
      groups = data.groups || [];
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  async function triggerScan() {
    scanning = true;
    message = 'A varrer a biblioteca em busca de títulos repetidos...';
    try {
      const res = await fetch('/api/duplicates/scan', { method: 'POST' });
      const data = await res.json();
      message = `Verificação concluída: ${data.duplicates_count || 0} grupo(s) de duplicados encontrados.`;
      await loadDuplicates();
    } catch (e) {
      message = 'Erro ao executar verificação.';
    } finally {
      scanning = false;
    }
  }

  async function resolveGroup(g) {
    if (!g.books || g.books.length < 2) return;
    resolvingHash = g.hash;
    const keptId = g.books[0].id;
    const sourceIds = g.books.slice(1).map(b => b.id);

    try {
      const res = await fetch('/api/duplicates/resolve', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ kept_id: keptId, source_ids: sourceIds })
      });
      const data = await res.json();
      if (data.success) {
        message = `✓ Formatos unificados com sucesso no livro #${keptId}!`;
        await loadDuplicates();
      }
    } catch (e) {
      message = 'Erro ao unificar duplicados.';
    } finally {
      resolvingHash = null;
    }
  }

  async function dismissGroup(hash) {
    try {
      await fetch(`/api/duplicates/dismiss/${hash}`, { method: 'POST' });
      groups = groups.filter(g => g.hash !== hash);
    } catch (e) {
      console.error(e);
    }
  }
</script>

<div class="duplicates-container">
  <!-- Header -->
  <div class="dup-header glass">
    <div class="header-left">
      <h2>Gestor Inteligente de Duplicados 🔍</h2>
      <p>Identifica títulos repetidos com formatos distintos (ex: EPUB e PDF separados) e permite unificá-los num registo limpo com 1 clique.</p>
    </div>

    <button class="btn btn-primary" on:click={triggerScan} disabled={scanning}>
      {#if scanning}
        <span>A analisar...</span>
      {:else}
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <span>Verificar Duplicados</span>
      {/if}
    </button>
  </div>

  {#if message}
    <div class="msg-banner glass">
      <span>{message}</span>
    </div>
  {/if}

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>A procurar duplicados...</p>
    </div>
  {:else if groups.length === 0}
    <div class="empty-box glass">
      <div class="check-icon">
        <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
      </div>
      <h3>Biblioteca Perfeita!</h3>
      <p>Não foram detetados livros duplicados na base de dados.</p>
    </div>
  {:else}
    <div class="groups-list">
      {#each groups as g (g.hash || g.title)}
        <div class="group-card glass">
          <div class="group-header">
            <div>
              <h4 class="group-title">{g.title}</h4>
              <span class="group-author">por {g.author || 'Autor'}</span>
            </div>

            <div class="group-actions">
              <button 
                class="btn btn-secondary btn-sm" 
                on:click={() => dismissGroup(g.hash)}>
                Descartar Aviso
              </button>
              <button 
                class="btn btn-primary btn-sm" 
                on:click={() => resolveGroup(g)}
                disabled={resolvingHash === g.hash}>
                {resolvingHash === g.hash ? 'A unificar...' : 'Unificar Formatos'}
              </button>
            </div>
          </div>

          <!-- Books inside group -->
          <div class="books-in-group">
            {#if g.books}
              {#each g.books as b}
                <div class="sub-book glass">
                  <div class="sub-cover">
                    <img src="/cover/{b.id}" alt={b.title} />
                  </div>
                  <div class="sub-info">
                    <span class="sub-id">ID: #{b.id}</span>
                    <span class="sub-t">{b.title}</span>
                    <div class="sub-fmts">
                      {#if b.formats}
                        {#each b.formats as f}
                          <span class="fmt-pill">{f.format}</span>
                        {/each}
                      {/if}
                    </div>
                  </div>
                </div>
              {/each}
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .duplicates-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .dup-header {
    padding: 24px;
    border-radius: var(--radius-lg);
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
  }

  .header-left h2 {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .header-left p {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .msg-banner {
    padding: 12px 18px;
    border-radius: var(--radius-md);
    background: rgba(56, 189, 248, 0.1);
    border-color: rgba(56, 189, 248, 0.25);
    color: var(--accent);
    font-size: 13px;
  }

  .groups-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .group-card {
    padding: 20px;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .group-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px;
  }

  .group-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .group-author {
    font-size: 12px;
    color: var(--text-muted);
  }

  .group-actions {
    display: flex;
    gap: 8px;
  }

  .books-in-group {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
  }

  .sub-book {
    padding: 10px;
    border-radius: var(--radius-sm);
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .sub-cover {
    width: 44px;
    aspect-ratio: 2 / 3;
    border-radius: 4px;
    overflow: hidden;
    background: rgba(0,0,0,0.2);
    flex-shrink: 0;
  }
  .sub-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .sub-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow: hidden;
  }
  .sub-id {
    font-size: 10px;
    color: var(--text-muted);
    font-family: var(--font-mono);
  }
  .sub-t {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sub-fmts {
    display: flex;
    gap: 4px;
    margin-top: 2px;
  }
  .fmt-pill {
    background: rgba(255, 255, 255, 0.08);
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 3px;
    color: var(--accent);
    font-weight: 600;
  }

  .empty-box {
    text-align: center;
    padding: 60px 20px;
    border-radius: var(--radius-lg);
  }
  .check-icon {
    font-size: 40px;
    margin-bottom: 8px;
  }
  .empty-box h3 {
    font-size: 17px;
    color: var(--text-primary);
  }
  .empty-box p {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .loading-state {
    text-align: center;
    padding: 60px;
    color: var(--text-secondary);
  }
  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(56, 189, 248, 0.15);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin: 0 auto 12px;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>

