<script>
  import { onMount, createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let activeSubTab = 'authors'; // 'authors', 'series', 'tags', 'formats'
  let filterText = '';
  let authors = [];
  let series = [];
  let tags = [];
  let formats = [];
  let loading = true;

  onMount(loadData);

  async function loadData() {
    loading = true;
    try {
      const [resA, resS, resT, resF] = await Promise.all([
        fetch('/api/authors'),
        fetch('/api/series'),
        fetch('/api/tags'),
        fetch('/api/formats')
      ]);
      authors = await resA.json();
      series = await resS.json();
      tags = await resT.json();
      formats = await resF.json();
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function selectItem(type, item) {
    dispatch('filterSelect', { type, item });
  }

  $: filteredAuthors = authors.filter(a => a.name.toLowerCase().includes(filterText.toLowerCase()));
  $: filteredSeries = series.filter(s => s.name.toLowerCase().includes(filterText.toLowerCase()));
  $: filteredTags = tags.filter(t => t.name.toLowerCase().includes(filterText.toLowerCase()));
</script>

<div class="browse-container">
  <!-- Header Bar -->
  <div class="browse-header glass">
    <div class="header-left">
      <h2>Explorar Biblioteca Universal</h2>
      <p>Navega pelos autores, sagas, categorias temáticas e formatos dos livros da tua coleção.</p>
    </div>

    <!-- Sub-tabs -->
    <div class="sub-nav">
      <button 
        class="sub-btn" 
        class:active={activeSubTab === 'authors'} 
        on:click={() => { activeSubTab = 'authors'; filterText = ''; }}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
        <span>Autores</span>
        <span class="count-tag">{authors.length}</span>
      </button>

      <button 
        class="sub-btn" 
        class:active={activeSubTab === 'series'} 
        on:click={() => { activeSubTab = 'series'; filterText = ''; }}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
        <span>Séries</span>
        <span class="count-tag">{series.length}</span>
      </button>

      <button 
        class="sub-btn" 
        class:active={activeSubTab === 'tags'} 
        on:click={() => { activeSubTab = 'tags'; filterText = ''; }}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line></svg>
        <span>Categorias</span>
        <span class="count-tag">{tags.length}</span>
      </button>

      <button 
        class="sub-btn" 
        class:active={activeSubTab === 'formats'} 
        on:click={() => { activeSubTab = 'formats'; filterText = ''; }}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
        <span>Formatos</span>
        <span class="count-tag">{formats.length}</span>
      </button>
    </div>
  </div>

  <!-- Filter Search Input -->
  {#if activeSubTab !== 'formats'}
    <div class="filter-box">
      <input 
        type="text" 
        placeholder="Filtrar por nome..." 
        bind:value={filterText} 
      />
    </div>
  {/if}

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>A carregar catálogo...</p>
    </div>
  {:else}
    <!-- Content Grids -->
    <div class="items-grid">
      {#if activeSubTab === 'authors'}
        {#each filteredAuthors as a}
          <div 
            class="entity-card glass glass-hover" 
            on:click={() => selectItem('author', a)} 
            on:keydown={(e) => { if (e.key === 'Enter') selectItem('author', a); }}
            role="button" 
            tabindex="0">
            <div class="entity-avatar">{a.name.slice(0, 2).toUpperCase()}</div>
            <div class="entity-info">
              <span class="entity-name">{a.name}</span>
              <span class="entity-sub">{a.count !== undefined ? a.count : (a.book_count || 0)} {a.count === 1 ? 'livro' : 'livros'}</span>
            </div>
          </div>
        {/each}
      {:else if activeSubTab === 'series'}
        {#each filteredSeries as s}
          <div 
            class="entity-card glass glass-hover" 
            on:click={() => selectItem('series', s)} 
            on:keydown={(e) => { if (e.key === 'Enter') selectItem('series', s); }}
            role="button" 
            tabindex="0">
            <div class="entity-avatar series">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
            </div>
            <div class="entity-info">
              <span class="entity-name">{s.name}</span>
              <span class="entity-sub">{s.count !== undefined ? s.count : (s.book_count || 0)} {s.count === 1 ? 'volume' : 'volumes'}</span>
            </div>
          </div>
        {/each}
      {:else if activeSubTab === 'tags'}
        {#each filteredTags as t}
          <div 
            class="entity-card glass glass-hover" 
            on:click={() => selectItem('tag', t)} 
            on:keydown={(e) => { if (e.key === 'Enter') selectItem('tag', t); }}
            role="button" 
            tabindex="0">
            <div class="entity-avatar tag">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line></svg>
            </div>
            <div class="entity-info">
              <span class="entity-name">{t.name}</span>
              <span class="entity-sub">{t.count !== undefined ? t.count : (t.book_count || 0)} {t.count === 1 ? 'obra' : 'obras'}</span>
            </div>
          </div>
        {/each}
      {:else if activeSubTab === 'formats'}
        {#each formats as f}
          <div 
            class="entity-card glass glass-hover" 
            on:click={() => selectItem('format', f)} 
            on:keydown={(e) => { if (e.key === 'Enter') selectItem('format', f); }}
            role="button" 
            tabindex="0">
            <div class="entity-avatar format">{(f.name || f.format || 'DOC').toUpperCase()}</div>
            <div class="entity-info">
              <span class="entity-name">{f.name || f.format}</span>
              <span class="entity-sub">{f.count !== undefined ? f.count : (f.book_count || 0)} ficheiros</span>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style>
  .browse-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .browse-header {
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

  .sub-nav {
    display: flex;
    background: rgba(255, 255, 255, 0.04);
    padding: 4px;
    border-radius: var(--radius-md);
    gap: 4px;
  }

  .sub-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    padding: 6px 14px;
    font-size: 13px;
    font-weight: 500;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .sub-btn:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .sub-btn.active {
    background: var(--accent);
    color: white;
    font-weight: 600;
  }

  .count-tag {
    font-size: 11px;
    background: rgba(0, 0, 0, 0.2);
    padding: 1px 6px;
    border-radius: 10px;
  }

  .filter-box {
    max-width: 400px;
  }
  .filter-box input {
    width: 100%;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 8px 14px;
    color: var(--text-primary);
    outline: none;
  }
  .filter-box input:focus {
    border-color: var(--accent);
  }

  .items-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 14px;
  }

  .entity-card {
    padding: 14px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .entity-card:hover {
    transform: translateY(-2px);
  }

  .entity-avatar {
    width: 42px;
    height: 42px;
    background: rgba(56, 189, 248, 0.12);
    border: 1px solid rgba(56, 189, 248, 0.3);
    border-radius: var(--radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    color: var(--accent);
    font-size: 13px;
    flex-shrink: 0;
  }
  .entity-avatar.series {
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.3);
    font-size: 18px;
  }
  .entity-avatar.tag {
    background: rgba(168, 85, 247, 0.12);
    border-color: rgba(168, 85, 247, 0.3);
    font-size: 18px;
  }
  .entity-avatar.format {
    background: rgba(34, 197, 94, 0.12);
    border-color: rgba(34, 197, 94, 0.3);
    color: #22c55e;
    font-size: 11px;
  }

  .entity-info {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .entity-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .entity-sub {
    font-size: 11px;
    color: var(--text-muted);
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

