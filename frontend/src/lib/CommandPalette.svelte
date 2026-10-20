<script>
  // @ts-nocheck
  import { onMount, createEventDispatcher } from 'svelte';
  export let isOpen = false;

  const dispatch = createEventDispatcher();
  let searchInput;
  let query = '';
  let selectedIndex = 0;
  let bookResults = [];
  let isSearchingBooks = false;
  let searchTimeout = null;

  const defaultNavigation = [
    { id: 'nav-lib', category: 'Navegação', title: 'Biblioteca Principal', desc: 'Ver todos os livros, capas e ordenação', icon: 'book', action: () => dispatch('navigate', { tab: 'library' }) },
    { id: 'nav-browse', category: 'Navegação', title: 'Explorar Catálogo', desc: 'Navegar por autores, séries, categorias e formatos', icon: 'compass', action: () => dispatch('navigate', { tab: 'browse' }) },
    { id: 'nav-shelves', category: 'Navegação', title: 'Estantes & Magic Shelves', desc: 'Gerir coleções e estantes inteligentes', icon: 'bookmark', action: () => dispatch('navigate', { tab: 'shelves' }) },
    { id: 'nav-fts', category: 'Pesquisa', title: 'Full-Text Search (FTS5)', desc: 'Pesquisa profunda no interior dos ficheiros EPUB/PDF', icon: 'search', action: () => dispatch('navigate', { tab: 'fts' }) },
    { id: 'nav-mam', category: 'Ferramentas', title: 'Cheeseshop Tracker (MAM)', desc: 'Pesquisa e download direto da comunidade 🐭🧀', icon: 'zap', action: () => dispatch('navigate', { tab: 'mam' }) },
    { id: 'nav-dups', category: 'Ferramentas', title: 'Detector de Duplicados', desc: 'Identificar e resolver livros duplicados', icon: 'copy', action: () => dispatch('navigate', { tab: 'duplicates' }) },
    { id: 'nav-plugins', category: 'Sistema', title: 'Gestor de Plugins & DeDRM', desc: 'Gerir extensões ativas do Calibre', icon: 'cpu', action: () => dispatch('navigate', { tab: 'plugins' }) },
    { id: 'nav-stats', category: 'Sistema', title: 'Estatísticas & Ocupação', desc: 'Métricas de leitura, autores e memória Go', icon: 'bar-chart', action: () => dispatch('navigate', { tab: 'stats' }) },
    { id: 'nav-settings', category: 'Sistema', title: 'Definições do Servidor CWA', desc: 'Configurar conversões, backups e ingestão', icon: 'settings', action: () => dispatch('navigate', { tab: 'cwa_settings' }) },
  ];

  const quickActions = [
    { id: 'act-upload', category: 'Ações Rápidas', title: 'Adicionar / Enviar Livro', desc: 'Fazer upload de EPUB, PDF, CBR, CBZ ou MOBI', icon: 'upload', action: () => dispatch('openUpload') },
    { id: 'act-theme', category: 'Ações Rápidas', title: 'Alternar Tema Claro / Escuro', desc: 'Mudar modo de cor da interface', icon: 'sun-moon', action: () => dispatch('toggleTheme') },
    { id: 'act-refresh', category: 'Ações Rápidas', title: 'Recarregar Biblioteca', desc: 'Atualizar listagem e sincronizar metadados', icon: 'refresh', action: () => dispatch('refresh') }
  ];

  $: {
    if (isOpen) {
      setTimeout(() => {
        if (searchInput) searchInput.focus();
      }, 50);
    } else {
      query = '';
      bookResults = [];
      selectedIndex = 0;
    }
  }

  $: filteredActions = query.trim() ? [
    ...defaultNavigation.filter(item => 
      item.title.toLowerCase().includes(query.toLowerCase()) || 
      item.desc.toLowerCase().includes(query.toLowerCase())
    ),
    ...quickActions.filter(item => 
      item.title.toLowerCase().includes(query.toLowerCase()) || 
      item.desc.toLowerCase().includes(query.toLowerCase())
    )
  ] : [...quickActions, ...defaultNavigation];

  $: allVisibleItems = [
    ...bookResults.map(b => ({
      id: `book-${b.id}`,
      category: 'Livros Encontrados',
      title: b.title,
      desc: b.authors ? b.authors.join(', ') : 'Autor desconhecido',
      icon: 'book-open',
      coverId: b.id,
      action: () => dispatch('selectBook', { book: b })
    })),
    ...filteredActions
  ];

  $: if (selectedIndex >= allVisibleItems.length && allVisibleItems.length > 0) {
    selectedIndex = allVisibleItems.length - 1;
  }

  function handleQueryChange() {
    selectedIndex = 0;
    if (searchTimeout) clearTimeout(searchTimeout);
    const q = query.trim();
    if (!q) {
      bookResults = [];
      return;
    }

    searchTimeout = setTimeout(async () => {
      isSearchingBooks = true;
      try {
        const res = await fetch(`/api/books?limit=5&query=${encodeURIComponent(q)}`);
        if (res.ok) {
          const data = await res.json();
          bookResults = data.items || [];
        }
      } catch (err) {
        console.error('CommandPalette search error:', err);
      } finally {
        isSearchingBooks = false;
      }
    }, 180);
  }

  function handleKeyDown(e) {
    if (!isOpen) return;

    if (e.key === 'Escape') {
      e.preventDefault();
      close();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (allVisibleItems.length > 0) {
        selectedIndex = (selectedIndex + 1) % allVisibleItems.length;
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (allVisibleItems.length > 0) {
        selectedIndex = (selectedIndex - 1 + allVisibleItems.length) % allVisibleItems.length;
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (allVisibleItems[selectedIndex]) {
        executeItem(allVisibleItems[selectedIndex]);
      }
    }
  }

  function executeItem(item) {
    if (item && item.action) {
      item.action();
      close();
    }
  }

  function close() {
    dispatch('close');
  }

  onMount(() => {
    function onGlobalKeyDown(e) {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        if (isOpen) {
          close();
        } else {
          dispatch('open');
        }
      }
    }
    window.addEventListener('keydown', onGlobalKeyDown);
    return () => window.removeEventListener('keydown', onGlobalKeyDown);
  });
</script>

<svelte:window on:keydown={handleKeyDown} />

{#if isOpen}
  <div class="palette-backdrop" on:click={close} role="presentation">
    <div class="palette-modal glass" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
      <div class="palette-search-wrap">
        <svg class="search-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <input 
          bind:this={searchInput}
          bind:value={query} 
          on:input={handleQueryChange}
          type="text" 
          placeholder="O que procuras? Digita para pesquisar livros, vistas ou atalhos..." 
          class="palette-input"
        />
        {#if query}
          <button class="clear-btn" on:click={() => { query = ''; handleQueryChange(); }}>✕</button>
        {/if}
        <button class="palette-esc" on:click={close} aria-label="Fechar">ESC</button>
      </div>

      <div class="palette-results">
        {#if allVisibleItems.length === 0}
          <div class="palette-empty">
            <p>Nenhum resultado para "<strong>{query}</strong>"</p>
            <span>Tenta pesquisar por outro título, autor ou comando.</span>
          </div>
        {:else}
          {#each allVisibleItems as item, idx (item.id)}
            <div 
              class="palette-item" 
              class:selected={idx === selectedIndex}
              on:mouseenter={() => { selectedIndex = idx; }}
              on:click={() => executeItem(item)}
              on:keydown={(e) => { if (e.key === 'Enter') executeItem(item); }}
              role="button"
              tabindex="0"
            >
              <div class="item-icon-wrap">
                {#if item.coverId}
                  <img src="/cover/{item.coverId}" alt="" class="item-cover-thumb" on:error={(e) => { e.target.style.display = 'none'; }} />
                {:else if item.icon === 'book'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
                {:else if item.icon === 'compass'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
                {:else if item.icon === 'bookmark'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
                {:else if item.icon === 'search'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                {:else if item.icon === 'zap'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                {:else if item.icon === 'copy'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                {:else if item.icon === 'cpu'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><line x1="9" y1="9" x2="15" y2="9"></line><line x1="9" y1="15" x2="15" y2="15"></line></svg>
                {:else if item.icon === 'bar-chart'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
                {:else if item.icon === 'settings'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
                {:else if item.icon === 'upload'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
                {:else if item.icon === 'sun-moon'}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
                {:else}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                {/if}
              </div>

              <div class="item-text">
                <div class="item-title-row">
                  <span class="item-title">{item.title}</span>
                  <span class="item-category-tag">{item.category}</span>
                </div>
                <span class="item-desc">{item.desc}</span>
              </div>

              <div class="item-enter-hint">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 10 4 15 9 20"></polyline><path d="M20 4v7a4 4 0 0 1-4 4H4"></path></svg>
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <div class="palette-footer">
        <div class="footer-hint">
          <kbd>↑</kbd><kbd>↓</kbd> navegar
        </div>
        <div class="footer-hint">
          <kbd>↵</kbd> selecionar
        </div>
        <div class="footer-hint">
          <kbd>esc</kbd> fechar
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .palette-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    z-index: 99998;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 12vh;
    animation: fade-in 0.18s ease-out;
  }

  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .palette-modal {
    width: 620px;
    max-width: calc(100vw - 32px);
    max-height: 72vh;
    background: var(--bg-card, #161b22);
    border: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
    border-radius: 16px;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.55), 0 0 0 1px rgba(255, 255, 255, 0.08);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: scale-up 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes scale-up {
    from { transform: scale(0.96) translateY(-8px); opacity: 0; }
    to { transform: scale(1) translateY(0); opacity: 1; }
  }

  .palette-search-wrap {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  }

  .search-icon {
    color: var(--accent, #6366f1);
    flex-shrink: 0;
  }

  .palette-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text-primary, #ffffff);
    font-size: 1.05rem;
    font-weight: 500;
  }

  .palette-input::placeholder {
    color: var(--text-secondary, #8b949e);
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary, #8b949e);
    cursor: pointer;
    font-size: 0.9rem;
    padding: 4px;
    border-radius: 4px;
  }

  .clear-btn:hover {
    color: var(--text-primary, #ffffff);
  }

  .palette-esc {
    font-size: 0.75rem;
    padding: 3px 6px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: var(--text-secondary, #8b949e);
    cursor: pointer;
    user-select: none;
  }

  .palette-results {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .palette-empty {
    padding: 36px 20px;
    text-align: center;
    color: var(--text-secondary, #8b949e);
  }

  .palette-empty p {
    margin: 0 0 6px 0;
    color: var(--text-primary, #ffffff);
    font-size: 1rem;
  }

  .palette-empty span {
    font-size: 0.85rem;
  }

  .palette-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.12s ease;
    user-select: none;
  }

  .palette-item.selected {
    background: var(--accent-subtle, rgba(99, 102, 241, 0.15));
  }

  .item-icon-wrap {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.06);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary, #8b949e);
    flex-shrink: 0;
    overflow: hidden;
  }

  .palette-item.selected .item-icon-wrap {
    background: var(--accent, #6366f1);
    color: #ffffff;
  }

  .item-cover-thumb {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .item-text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .item-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .item-title {
    font-size: 0.95rem;
    font-weight: 500;
    color: var(--text-primary, #ffffff);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-category-tag {
    font-size: 0.72rem;
    padding: 2px 6px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-secondary, #8b949e);
    text-transform: uppercase;
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  .item-desc {
    font-size: 0.8rem;
    color: var(--text-secondary, #8b949e);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-enter-hint {
    opacity: 0;
    color: var(--accent, #6366f1);
    transition: opacity 0.15s ease;
  }

  .palette-item.selected .item-enter-hint {
    opacity: 1;
  }

  .palette-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 16px;
    padding: 10px 18px;
    border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
    background: rgba(0, 0, 0, 0.15);
  }

  .footer-hint {
    font-size: 0.75rem;
    color: var(--text-secondary, #8b949e);
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .footer-hint kbd {
    font-size: 0.7rem;
    padding: 2px 5px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.12);
  }
</style>
