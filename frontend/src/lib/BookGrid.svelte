<script>
  import { createEventDispatcher } from 'svelte';
  import BookCard from './BookCard.svelte';
  import SkeletonGrid from './SkeletonGrid.svelte';

  export let books = [];
  export let totalBooks = 0;
  export let currentPage = 1;
  export let pageSize = 30;
  export let currentSort = 'new';
  export let currentFilter = 'all'; // 'all', 'read', 'unread', 'hot', 'random'
  export let currentFilterTitle = '';
  export let loading = false;

  const dispatch = createEventDispatcher();

  let selectMode = false;
  let selectedIds = new Set();
  let batchLoading = false;

  $: totalPages = Math.ceil(totalBooks / pageSize) || 1;

  function onSortChange(e) {
    dispatch('sortChange', { sort: e.target.value });
  }

  function onFilterChange(filter) {
    currentFilter = filter;
    dispatch('filterChange', { filter });
  }

  function onPageChange(page) {
    if (page >= 1 && page <= totalPages) {
      dispatch('pageChange', { page });
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  }

  function handleSelectBook(event) {
    dispatch('selectBook', event.detail);
  }

  function toggleBookSelect(event) {
    const id = event.detail.id;
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
    selectedIds = new Set(selectedIds);
  }

  function toggleSelectMode() {
    selectMode = !selectMode;
    if (!selectMode) {
      selectedIds = new Set();
    }
  }

  function selectAllCurrent() {
    books.forEach(b => selectedIds.add(b.id));
    selectedIds = new Set(selectedIds);
  }

  function clearSelection() {
    selectedIds = new Set();
  }

  async function handleBatchAction(action) {
    if (selectedIds.size === 0) return;
    if (action === 'delete' && !confirm(`Eliminar ${selectedIds.size} livros selecionados? Esta ação não pode ser desfeita.`)) {
      return;
    }

    batchLoading = true;
    try {
      const res = await fetch('/api/books/batch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          action,
          book_ids: Array.from(selectedIds)
        })
      });
      if (res.ok) {
        clearSelection();
        selectMode = false;
        dispatch('refresh');
      }
    } catch (e) {
      console.error(e);
    } finally {
      batchLoading = false;
    }
  }
</script>

<div class="grid-view">
  <!-- Controls Bar -->
  <div class="controls-bar glass">
    <div class="controls-left">
      <!-- Quick Filter Pills -->
      <div class="filter-pills">
        <button 
          class="pill-btn" 
          class:active={currentFilter === 'all'} 
          on:click={() => onFilterChange('all')}>
          Todos
        </button>
        <button 
          class="pill-btn" 
          class:active={currentFilter === 'read'} 
          on:click={() => onFilterChange('read')}>
          Lidos
        </button>
        <button 
          class="pill-btn" 
          class:active={currentFilter === 'unread'} 
          on:click={() => onFilterChange('unread')}>
          Não Lidos
        </button>
        <button 
          class="pill-btn" 
          class:active={currentFilter === 'hot'} 
          on:click={() => onFilterChange('hot')}>
          Destaques
        </button>
        <button 
          class="pill-btn" 
          class:active={currentFilter === 'random'} 
          on:click={() => onFilterChange('random')}>
          Aleatório
        </button>
      </div>

      {#if currentFilterTitle}
        <div class="active-filter-badge">
          <span>{currentFilterTitle}</span>
          <button class="clear-filter" on:click={() => dispatch('clearCustomFilter')}>✕</button>
        </div>
      {/if}
    </div>

    <div class="controls-right">
      <!-- Multi-select toggle -->
      <button 
        class="btn btn-secondary btn-sm" 
        class:btn-primary={selectMode} 
        on:click={toggleSelectMode}>
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 11 12 14 22 4"></polyline><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path></svg>
        <span>{selectMode ? 'Concluir Seleção' : 'Seleção Múltipla'}</span>
      </button>

      <label class="sort-label">
        <span>Ordenar:</span>
        <select class="sort-select" value={currentSort} on:change={onSortChange}>
          <option value="new">Mais Recentes</option>
          <option value="old">Mais Antigos</option>
          <option value="abc">Título (A-Z)</option>
          <option value="zyx">Título (Z-A)</option>
          <option value="authaz">Autor (A-Z)</option>
          <option value="pubnew">Data de Publicação</option>
          <option value="random">Aleatório</option>
        </select>
      </label>
    </div>
  </div>

  <!-- Loading State (Skeleton Screen) -->
  {#if loading}
    <SkeletonGrid count={pageSize || 30} />
  {:else if books.length === 0}
    <!-- Empty State -->
    <div class="empty-state glass">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
      </div>
      <h3>Nenhum livro encontrado</h3>
      <p>Experimenta alterar os termos de pesquisa ou o filtro selecionado.</p>
    </div>
  {:else}
    <!-- Books Grid -->
    <div class="books-grid">
      {#each books as book (book.id)}
        <BookCard 
          {book} 
          {selectMode} 
          isSelected={selectedIds.has(book.id)}
          on:select={handleSelectBook} 
          on:toggleSelect={toggleBookSelect}
        />
      {/each}
    </div>

    <!-- Pagination -->
    {#if totalPages > 1}
      <div class="pagination-bar">
        <button 
          class="btn btn-secondary" 
          disabled={currentPage <= 1} 
          on:click={() => onPageChange(currentPage - 1)}>
          ← Anterior
        </button>

        <span class="page-indicator">
          Página <strong>{currentPage}</strong> de <strong>{totalPages}</strong>
          <span class="total-books-indicator">({totalBooks} livros)</span>
        </span>

        <button 
          class="btn btn-secondary" 
          disabled={currentPage >= totalPages} 
          on:click={() => onPageChange(currentPage + 1)}>
          Seguinte →
        </button>
      </div>
    {/if}
  {/if}

  <!-- Floating Batch Action Bar -->
  {#if selectMode && selectedIds.size > 0}
    <div class="batch-bar glass">
      <div class="batch-count">
        <strong>{selectedIds.size}</strong> {selectedIds.size === 1 ? 'livro selecionado' : 'livros selecionados'}
      </div>

      <div class="batch-actions">
        <button class="btn btn-secondary btn-sm" on:click={selectAllCurrent}>
          Selecionar Página
        </button>
        <button class="btn btn-secondary btn-sm" on:click={() => handleBatchAction('mark_read')} disabled={batchLoading}>
          Marcar Lido
        </button>
        <button class="btn btn-secondary btn-sm" on:click={() => handleBatchAction('mark_unread')} disabled={batchLoading}>
          Marcar Não Lido
        </button>
        <button class="btn btn-secondary btn-sm" on:click={() => handleBatchAction('archive')} disabled={batchLoading}>
          Arquivar
        </button>
        <button class="btn btn-danger btn-sm" on:click={() => handleBatchAction('delete')} disabled={batchLoading}>
          Eliminar
        </button>
        <button class="btn btn-secondary btn-sm" on:click={clearSelection}>
          ✕
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .grid-view {
    max-width: 1750px;
    margin: 0 auto;
    padding: 20px 24px;
    display: flex;
    flex-direction: column;
    gap: 18px;
    position: relative;
  }

  .controls-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 18px;
    border-radius: var(--radius-md);
    flex-wrap: wrap;
    gap: 12px;
  }

  .controls-left {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .filter-pills {
    display: flex;
    gap: 4px;
    background: rgba(255, 255, 255, 0.04);
    padding: 3px;
    border-radius: var(--radius-sm);
  }

  .pill-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 12px;
    padding: 4px 10px;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .pill-btn:hover {
    color: var(--text-primary);
  }
  .pill-btn.active {
    background: var(--accent);
    color: white;
    font-weight: 600;
  }

  .active-filter-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--accent-glow);
    border: 1px solid var(--accent);
    color: var(--accent);
    padding: 2px 10px;
    border-radius: var(--radius-full);
    font-size: 12px;
    font-weight: 600;
  }
  .clear-filter {
    background: transparent;
    border: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 11px;
  }

  .controls-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .sort-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .sort-select {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    padding: 5px 10px;
    font-size: 12px;
    outline: none;
    cursor: pointer;
  }

  .books-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 18px;
  }

  .empty-state {
    text-align: center;
    padding: 60px 20px;
    border-radius: var(--radius-lg);
  }
  .empty-icon {
    font-size: 48px;
    margin-bottom: 12px;
  }
  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 6px;
    color: var(--text-primary);
  }
  .empty-state p {
    color: var(--text-secondary);
    font-size: 13px;
  }


  .pagination-bar {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    margin-top: 16px;
  }
  .page-indicator {
    font-size: 13px;
    color: var(--text-secondary);
  }
  .page-indicator strong {
    color: var(--text-primary);
  }
  .total-books-indicator {
    color: var(--text-muted);
    font-size: 11px;
    margin-left: 4px;
  }

  /* Floating Batch Bar */
  .batch-bar {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 1000;
    padding: 12px 24px;
    border-radius: var(--radius-full);
    display: flex;
    align-items: center;
    gap: 20px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.6);
    border: 1px solid var(--border-secondary);
    animation: slideUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }
  @keyframes slideUp {
    from { transform: translate(-50%, 40px); opacity: 0; }
    to { transform: translate(-50%, 0); opacity: 1; }
  }

  .batch-count {
    font-size: 13px;
    color: var(--text-primary);
    white-space: nowrap;
  }
  .batch-count strong {
    color: var(--accent);
  }

  .batch-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  @media (max-width: 640px) {
    .books-grid {
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;
    }
    .batch-bar {
      flex-direction: column;
      border-radius: var(--radius-lg);
      width: 90%;
    }
  }
</style>
