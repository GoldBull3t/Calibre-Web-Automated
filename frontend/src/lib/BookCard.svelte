<script>
  // @ts-nocheck
  import { createEventDispatcher } from 'svelte';
  import { offlineBookIds } from './pwa.js';
  export let book;
  export let selectMode = false;
  export let isSelected = false;

  const dispatch = createEventDispatcher();

  function onCardClick() {
    if (selectMode) {
      dispatch('toggleSelect', { id: book.id });
    } else {
      dispatch('select', { book });
    }
  }

  function onCheckboxClick(e) {
    e.stopPropagation();
    dispatch('toggleSelect', { id: book.id });
  }

  function onRead(e) {
    e.stopPropagation();
    window.location.href = `/read/${book.id}`;
  }

  function onDownload(e, format) {
    e.stopPropagation();
    window.location.href = `/download/${book.id}/${format.toLowerCase()}`;
  }
</script>

<div 
  class="book-card glass glass-hover" 
  class:selected={isSelected}
  on:click={onCardClick} 
  on:keydown={(e) => { if (e.key === 'Enter') onCardClick(); }} 
  role="button" 
  tabindex="0">
  
  <!-- Cover Container -->
  <div class="cover-wrap">
    <img 
      src="/cover/{book.id}" 
      alt={book.title} 
      loading="lazy" 
      class="cover-img"
      on:error={(e) => { e.target.src = '/static/images/cover_placeholder.svg'; }}
    />

    <!-- Multi-select checkbox -->
    {#if selectMode}
      <div 
        class="checkbox-wrap" 
        on:click={onCheckboxClick} 
        on:keydown={(e) => { if (e.key === 'Enter') onCheckboxClick(e); }}
        role="button" 
        tabindex="0">
        <input type="checkbox" checked={isSelected} />
      </div>
    {/if}

    <!-- Read Status Badge -->
    {#if book.read_status && !selectMode}
      <div class="read-badge" title="Lido">
        <svg viewBox="0 0 24 24"><polyline points="20 6 9 17 4 12"></polyline></svg>
      </div>
    {/if}

    <!-- Offline Badge -->
    {#if $offlineBookIds && $offlineBookIds.has(book.id) && !selectMode}
      <div class="offline-badge" title="Guardado no dispositivo para leitura offline">
        <svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
        <span>Offline</span>
      </div>
    {/if}

    <!-- Hover Actions Overlay (only when not in select mode) -->
    {#if !selectMode}
      <div class="hover-overlay">
        <button class="btn btn-primary btn-sm action-btn" on:click={onRead}>
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
          <span>Ler</span>
        </button>

        {#if book.formats && book.formats.length > 0}
          <button class="btn btn-secondary btn-sm action-btn" on:click={(e) => onDownload(e, book.formats[0].format)}>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="15"></line></svg>
            <span>{book.formats[0].format}</span>
          </button>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Book Info -->
  <div class="info-wrap">
    <h3 class="title" title={book.title}>{book.title}</h3>
    
    <p class="author" title={book.authors ? book.authors.join(', ') : ''}>
      {book.authors && book.authors.length ? book.authors.join(', ') : 'Autor Desconhecido'}
    </p>

    <!-- Series Badge -->
    {#if book.series}
      <div class="series-pill" title="{book.series} #{book.series_index}">
        <span>{book.series}</span>
        <span class="series-index">#{book.series_index}</span>
      </div>
    {/if}

    <!-- Rating & Formats Row -->
    <div class="bottom-row">
      <div class="rating">
        {#if book.rating > 0}
          <span class="stars">{'★'.repeat(Math.round(book.rating / 2))}</span>
        {/if}
      </div>

      <div class="format-badges">
        {#if book.formats}
          {#each book.formats.slice(0, 2) as f}
            <span class="format-tag">{f.format}</span>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .book-card {
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-md);
    overflow: hidden;
    cursor: pointer;
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s ease, border-color 0.2s ease;
    user-select: none;
    background: var(--bg-card);
    position: relative;
  }

  .book-card:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
  }

  .book-card.selected {
    border-color: var(--accent);
    box-shadow: 0 0 12px var(--accent-glow);
  }

  .cover-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 2 / 3;
    overflow: hidden;
    background: rgba(0, 0, 0, 0.2);
  }

  .cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.3s ease;
  }

  .book-card:hover .cover-img {
    transform: scale(1.03);
  }

  .checkbox-wrap {
    position: absolute;
    top: 8px;
    left: 8px;
    z-index: 10;
    width: 22px;
    height: 22px;
    background: rgba(0, 0, 0, 0.6);
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(4px);
  }
  .checkbox-wrap input {
    cursor: pointer;
  }

  .read-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 24px;
    height: 24px;
    background: rgba(34, 197, 94, 0.9);
    border-radius: var(--radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 2px 6px rgba(0,0,0,0.3);
  }

  .read-badge svg {
    width: 14px;
    height: 14px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
  }

  .offline-badge {
    position: absolute;
    top: 8px;
    left: 8px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 7px;
    background: rgba(16, 185, 129, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    color: #ffffff;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    backdrop-filter: blur(4px);
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
    z-index: 5;
  }

  .hover-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, rgba(0,0,0,0.2) 60%, transparent 100%);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .book-card:hover .hover-overlay {
    opacity: 1;
  }

  .action-btn {
    padding: 6px 12px;
    font-size: 11px;
    border-radius: var(--radius-sm);
  }

  .info-wrap {
    padding: 12px 14px 14px;
    display: flex;
    flex-direction: column;
    flex: 1;
  }

  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 4px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.35;
  }

  .author {
    font-size: 12px;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-bottom: 8px;
  }

  .series-pill {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: var(--accent);
    background: var(--accent-glow);
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
    margin-bottom: 8px;
  }

  .series-index {
    font-weight: 700;
  }

  .bottom-row {
    margin-top: auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11px;
  }

  .stars {
    color: #f59e0b;
    font-size: 12px;
    letter-spacing: 1px;
  }

  .format-badges {
    display: flex;
    gap: 4px;
  }

  .format-tag {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--border-primary);
    padding: 1px 6px;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 600;
    color: var(--text-muted);
  }
</style>
