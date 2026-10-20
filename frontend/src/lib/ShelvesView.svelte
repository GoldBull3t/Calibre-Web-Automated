<script>
  import { onMount, createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let customShelves = [];
  let magicShelves = [];
  let loading = true;
  let showCreateModal = false;
  let newShelfName = '';
  let newShelfPublic = true;
  let creating = false;

  const defaultMagicShelves = [
    { id: 1, name: 'Recently Added', description: 'Livros adicionados recentemente à coleção', icon: '🕒' },
    { id: 2, name: 'Highly Rated', description: 'Obras com avaliação de 4 e 5 estrelas', icon: '⭐' },
    { id: 3, name: 'Currently Reading', description: 'Livros em curso com progresso de leitura guardado', icon: '📖' },
    { id: 4, name: 'Without Covers', description: 'Obras que necessitam de extração ou atribuição de capa', icon: '🖼️' }
  ];

  onMount(loadShelves);

  async function loadShelves() {
    loading = true;
    try {
      const res = await fetch('/api/shelves');
      const data = await res.json();
      customShelves = data.custom || [];
      magicShelves = data.magic && data.magic.length > 0 ? data.magic : defaultMagicShelves;
    } catch (e) {
      magicShelves = defaultMagicShelves;
    } finally {
      loading = false;
    }
  }

  function selectShelf(id, name, isMagic) {
    dispatch('shelfSelect', { id, name, isMagic });
  }

  async function createShelf() {
    if (!newShelfName.trim()) return;
    creating = true;
    try {
      const res = await fetch('/api/shelves', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newShelfName.trim(), is_public: newShelfPublic })
      });
      if (res.ok) {
        newShelfName = '';
        showCreateModal = false;
        await loadShelves();
      }
    } catch (e) {
      console.error(e);
    } finally {
      creating = false;
    }
  }

  async function deleteShelf(id, e) {
    e.stopPropagation();
    if (!confirm('Tens a certeza que desejas eliminar esta estante?')) return;
    try {
      const res = await fetch(`/api/shelves/${id}`, { method: 'DELETE' });
      if (res.ok) {
        await loadShelves();
      }
    } catch (err) {
      console.error(err);
    }
  }
  function getShelfMeta(ms) {
    const iconStr = (ms.icon || '').toLowerCase();
    const nameStr = (ms.name || '').toLowerCase();

    if (iconStr.includes('time') || iconStr.includes('clock') || nameStr.includes('recent') || ms.id === 1) {
      return {
        type: 'time',
        title: ms.name === 'Recently Added' ? 'Adicionados Recentemente' : ms.name,
        desc: ms.description === 'Books added in the last 30 days' ? 'Livros adicionados à biblioteca nos últimos 30 dias' : (ms.description || 'Adicionados recentemente'),
        colorClass: 'icon-time'
      };
    }
    if (iconStr.includes('star') || iconStr.includes('rate') || nameStr.includes('rate') || nameStr.includes('avaliad') || ms.id === 2) {
      return {
        type: 'star',
        title: ms.name === 'Highly Rated' ? 'Mais Bem Avaliados' : ms.name,
        desc: ms.description === 'Books with 4 or 5 stars rating' ? 'Obras com avaliação de 4 ou 5 estrelas' : (ms.description || 'Melhores avaliações'),
        colorClass: 'icon-star'
      };
    }
    if (iconStr.includes('bookmark') || iconStr.includes('read') || nameStr.includes('reading') || nameStr.includes('curso') || ms.id === 3) {
      return {
        type: 'reading',
        title: ms.name === 'Currently Reading' ? 'A Ler Atualmente' : ms.name,
        desc: ms.description === 'Books currently being read via KOReader or web' ? 'Leituras em curso sincronizadas via KOReader ou leitor web' : (ms.description || 'Leituras ativas'),
        colorClass: 'icon-reading'
      };
    }
    if (iconStr.includes('picture') || iconStr.includes('cover') || nameStr.includes('cover') || nameStr.includes('capa') || ms.id === 4) {
      return {
        type: 'nocover',
        title: (ms.name === 'Books Without Covers' || ms.name === 'Without Covers') ? 'Livros Sem Capa' : ms.name,
        desc: ms.description === 'Books missing cover image' ? 'Obras que necessitam de atribuição ou extração de capa' : (ms.description || 'Sem imagem de capa'),
        colorClass: 'icon-nocover'
      };
    }
    return {
      type: 'default',
      title: ms.name,
      desc: ms.description || 'Estante inteligente',
      colorClass: 'icon-default'
    };
  }
</script>

<div class="shelves-container">
  <!-- Header -->
  <div class="shelves-header glass">
    <div class="header-left">
      <h2>Estantes Mágicas & Personalizadas</h2>
      <p>Organiza as tuas obras por listas dinâmicas baseadas em regras inteligentes ou cria coleções temáticas próprias.</p>
    </div>

    <button class="btn btn-primary" on:click={() => { showCreateModal = true; }}>
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      <span>Criar Nova Estante</span>
    </button>
  </div>

  <!-- Magic Shelves Section -->
  <div class="section-title">
    <h3>Estantes Mágicas do Sistema</h3>
    <span class="sub-tag">Filtros dinâmicos automatizados</span>
  </div>

  <div class="shelves-grid">
    {#each magicShelves as ms}
      {@const meta = getShelfMeta(ms)}
      <div 
        class="magic-card glass glass-hover" 
        on:click={() => selectShelf(ms.id, meta.title, true)}
        on:keydown={(e) => { if (e.key === 'Enter') selectShelf(ms.id, meta.title, true); }}
        role="button" 
        tabindex="0">
        <div class="magic-icon {meta.colorClass}">
          {#if meta.type === 'time'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
          {:else if meta.type === 'star'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
          {:else if meta.type === 'reading'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
          {:else if meta.type === 'nocover'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>
          {:else}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
          {/if}
        </div>
        <div class="magic-content">
          <h4>{meta.title}</h4>
          <p>{meta.desc}</p>
        </div>
        <div class="arrow-btn">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"></polyline></svg>
        </div>
      </div>
    {/each}
  </div>

  <!-- Custom Shelves Section -->
  <div class="section-title" style="margin-top: 16px;">
    <h3>As Minhas Estantes Personalizadas</h3>
    <span class="sub-tag">{customShelves.length} {customShelves.length === 1 ? 'estante criada' : 'estantes criadas'}</span>
  </div>

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>A carregar estantes...</p>
    </div>
  {:else if customShelves.length === 0}
    <div class="empty-box glass">
      <p>Ainda não criaste estantes manuais. Clica em <strong>"Criar Nova Estante"</strong> para agrupar livros por tema!</p>
    </div>
  {:else}
    <div class="shelves-grid">
      {#each customShelves as cs}
        <div 
          class="custom-card glass glass-hover" 
          on:click={() => selectShelf(cs.id, cs.name, false)}
          on:keydown={(e) => { if (e.key === 'Enter') selectShelf(cs.id, cs.name, false); }}
          role="button" 
          tabindex="0">
          <div class="custom-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
          </div>
          <div class="custom-content">
            <h4>{cs.name}</h4>
            <span class="book-cnt">{cs.book_count || 0} {cs.book_count === 1 ? 'livro' : 'livros'}</span>
          </div>
          <button class="delete-btn" on:click={(e) => deleteShelf(cs.id, e)} title="Eliminar estante">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Create Shelf Modal -->
  {#if showCreateModal}
    <div class="modal-overlay" on:click={() => { showCreateModal = false; }} on:keydown={(e) => { if (e.key === 'Escape') showCreateModal = false; }} role="presentation">
      <div class="modal-content create-modal" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
        <h3>Criar Nova Estante</h3>
        <p>Define um nome para a tua estante personalizada.</p>

        <div class="form-group">
          <input 
            type="text" 
            placeholder="Ex: Favoritos de Ficção Científica, Leituras 2026..." 
            bind:value={newShelfName} 
          />
        </div>

        <label class="check-row">
          <input type="checkbox" bind:checked={newShelfPublic} />
          <span>Visível publicamente para outros utilizadores</span>
        </label>

        <div class="modal-actions">
          <button class="btn btn-secondary" on:click={() => { showCreateModal = false; }}>Cancelar</button>
          <button class="btn btn-primary" on:click={createShelf} disabled={!newShelfName.trim() || creating}>
            {creating ? 'A criar...' : 'Criar Estante'}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .shelves-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .shelves-header {
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

  .section-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .section-title h3 {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .sub-tag {
    font-size: 11px;
    color: var(--text-muted);
  }

  .shelves-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .magic-card {
    padding: 16px 18px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 16px;
    cursor: pointer;
    min-height: 84px;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }
  .magic-card:hover {
    transform: translateY(-2px);
    border-color: var(--accent);
  }

  .magic-icon {
    width: 44px;
    height: 44px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: transform 0.2s ease;
  }
  .magic-card:hover .magic-icon {
    transform: scale(1.05);
  }

  .icon-time {
    background: rgba(56, 189, 248, 0.12);
    color: #38bdf8;
    border: 1px solid rgba(56, 189, 248, 0.25);
  }
  .icon-star {
    background: rgba(245, 158, 11, 0.12);
    color: #f59e0b;
    border: 1px solid rgba(245, 158, 11, 0.25);
  }
  .icon-reading {
    background: rgba(16, 185, 129, 0.12);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.25);
  }
  .icon-nocover {
    background: rgba(244, 63, 94, 0.12);
    color: #f43f5e;
    border: 1px solid rgba(244, 63, 94, 0.25);
  }
  .icon-default {
    background: rgba(139, 92, 246, 0.12);
    color: #8b5cf6;
    border: 1px solid rgba(139, 92, 246, 0.25);
  }

  .magic-content {
    flex: 1;
    min-width: 0;
  }
  .magic-content h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 3px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .magic-content p {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.35;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .arrow-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    color: var(--text-muted);
    transition: all 0.2s ease;
    flex-shrink: 0;
  }
  .magic-card:hover .arrow-btn {
    color: var(--accent);
    transform: translateX(3px);
  }

  .custom-card {
    padding: 16px 18px;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .custom-card:hover {
    transform: translateY(-2px);
  }

  .custom-icon {
    width: 40px;
    height: 40px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--accent);
    flex-shrink: 0;
  }
  .custom-content {
    flex: 1;
    min-width: 0;
  }
  .custom-content h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .book-cnt {
    font-size: 11px;
    color: var(--text-muted);
  }

  .delete-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 6px;
    color: var(--text-muted);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }
  .delete-btn:hover {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.1);
  }

  .empty-box {
    padding: 32px;
    text-align: center;
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    font-size: 13px;
  }

  .create-modal {
    max-width: 440px;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .create-modal h3 {
    font-size: 17px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .create-modal p {
    font-size: 12px;
    color: var(--text-secondary);
  }
  .create-modal input[type="text"] {
    width: 100%;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 10px 14px;
    color: var(--text-primary);
    font-size: 13px;
    outline: none;
  }
  .create-modal input[type="text"]:focus {
    border-color: var(--accent);
  }
  .check-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }
  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 6px;
  }

  .loading-state {
    text-align: center;
    padding: 40px;
    color: var(--text-secondary);
  }
  .spinner {
    width: 28px;
    height: 28px;
    border: 3px solid rgba(56, 189, 248, 0.15);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin: 0 auto 12px;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
