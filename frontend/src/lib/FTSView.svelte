<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let query = '';
  let results = [];
  let count = 0;
  let loading = false;
  let reindexing = false;
  let reindexMessage = '';

  async function executeSearch() {
    if (!query.trim()) return;
    loading = true;
    try {
      const res = await fetch(`/api/search/fts?q=${encodeURIComponent(query.trim())}`);
      const data = await res.json();
      results = data.results || [];
      count = data.count || 0;
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function onKeyDown(e) {
    if (e.key === 'Enter') {
      executeSearch();
    }
  }

  async function triggerReindex() {
    reindexing = true;
    reindexMessage = 'A reindexar ficheiros EPUB na base FTS5...';
    try {
      const res = await fetch('/api/search/fts/reindex', { method: 'POST' });
      const data = await res.json();
      reindexMessage = data.message || 'Reindexação iniciada!';
    } catch (e) {
      reindexMessage = 'Erro ao iniciar reindexação.';
    } finally {
      setTimeout(() => { reindexing = false; }, 3000);
    }
  }

  function openBook(bookId) {
    dispatch('selectBook', { book: { id: bookId } });
  }

  function readBook(bookId) {
    window.location.href = `/read/${bookId}`;
  }
</script>

<div class="fts-container">
  <!-- Search Hero Header -->
  <div class="fts-hero glass">
    <div class="fts-badge">
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
      <span>SQLite FTS5 Porter Stemmer</span>
    </div>

    <h2>Pesquisa no Texto Integral dos Livros</h2>
    <p>Encontra citações, diálogos e qualquer termo dentro do conteúdo de todos os livros EPUB da biblioteca.</p>

    <!-- Search Input Bar -->
    <div class="fts-bar">
      <input 
        type="text" 
        placeholder="Ex: 'Sherlock Holmes', 'inteligência artificial', 'filosofia'..." 
        bind:value={query}
        on:keydown={onKeyDown}
      />
      <button class="btn btn-primary" on:click={executeSearch} disabled={loading}>
        {#if loading}
          <span>A pesquisar...</span>
        {:else}
          <span>Pesquisar FTS</span>
        {/if}
      </button>
    </div>

    <!-- Reindex Action -->
    <div class="reindex-row">
      <button class="btn btn-secondary btn-sm" on:click={triggerReindex} disabled={reindexing}>
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
        <span>Reindexar Biblioteca</span>
      </button>
      {#if reindexMessage}
        <span class="reindex-msg">{reindexMessage}</span>
      {/if}
    </div>
  </div>

  <!-- Results Section -->
  <div class="results-section">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>A varrer índice FTS5...</p>
      </div>
    {:else if results.length > 0}
      <div class="results-header">
        <span>Encontrados <strong>{count}</strong> resultados para "<em>{query}</em>"</span>
      </div>

      <div class="results-list">
        {#each results as r}
          <div class="result-card glass glass-hover">
            <div class="thumb-box" on:click={() => openBook(r.book_id)} on:keydown={(e) => { if (e.key === 'Enter') openBook(r.book_id); }} role="button" tabindex="0">
              <img src="/cover/{r.book_id}" alt={r.title} class="thumb-img" />
            </div>

            <div class="result-info">
              <div class="result-title-row">
                <button class="result-title-btn" on:click={() => openBook(r.book_id)}>{r.title}</button>
                <button class="btn btn-primary btn-sm" on:click={() => readBook(r.book_id)}>
                  Ler neste ponto
                </button>
              </div>

              <div class="result-author">por {r.author}</div>

              <!-- Context Snippet with <mark> tags -->
              <div class="snippet-box">
                <span class="quote-icon">“</span>
                <span class="snippet-text">{@html r.snippet}</span>
                <span class="quote-icon">”</span>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {:else if query && !loading}
      <div class="empty-results glass">
        <p>Nenhuma ocorrência encontrada para o termo "<strong>{query}</strong>".</p>
        <p class="hint">Dica: Se adicionaste novos livros recentemente, clica em "Reindexar Biblioteca".</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .fts-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .fts-hero {
    padding: 32px;
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 12px;
  }

  .fts-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 600;
    color: var(--accent);
    background: var(--accent-glow);
    padding: 3px 10px;
    border-radius: var(--radius-full);
  }

  .fts-hero h2 {
    font-size: 26px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .fts-hero p {
    color: var(--text-secondary);
    font-size: 14px;
    max-width: 600px;
  }

  .fts-bar {
    display: flex;
    gap: 10px;
    width: 100%;
    max-width: 680px;
    margin-top: 10px;
  }

  .fts-bar input {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-secondary);
    border-radius: var(--radius-md);
    padding: 12px 18px;
    color: var(--text-primary);
    font-size: 14px;
    outline: none;
    transition: all 0.2s;
  }

  .fts-bar input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }

  .reindex-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 8px;
  }

  .reindex-msg {
    font-size: 12px;
    color: var(--accent);
  }

  .results-header {
    font-size: 14px;
    color: var(--text-secondary);
    margin-bottom: 16px;
  }
  .results-header strong {
    color: var(--text-primary);
  }

  .results-list {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .result-card {
    display: flex;
    gap: 18px;
    padding: 16px;
    border-radius: var(--radius-md);
    align-items: flex-start;
  }

  .thumb-box {
    width: 60px;
    aspect-ratio: 2 / 3;
    border-radius: var(--radius-sm);
    overflow: hidden;
    cursor: pointer;
    flex-shrink: 0;
    box-shadow: var(--shadow-sm);
  }

  .thumb-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .result-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .result-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .result-title-btn {
    background: transparent;
    border: none;
    padding: 0;
    text-align: left;
    font-family: inherit;
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    cursor: pointer;
  }
  .result-title-btn:hover {
    color: var(--accent);
  }

  .result-author {
    font-size: 12px;
    color: var(--text-muted);
  }

  .snippet-box {
    margin-top: 8px;
    padding: 10px 14px;
    background: rgba(0, 0, 0, 0.25);
    border-left: 3px solid var(--accent);
    border-radius: 4px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--text-secondary);
  }

  .quote-icon {
    color: var(--accent);
    font-weight: bold;
    font-size: 16px;
  }

  .empty-results {
    padding: 40px;
    text-align: center;
    border-radius: var(--radius-md);
    color: var(--text-secondary);
  }
  .hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 6px;
  }

  .loading-state {
    text-align: center;
    padding: 60px;
    color: var(--text-secondary);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
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
