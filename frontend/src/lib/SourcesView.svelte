<script>
  // @ts-nocheck
  import { onMount } from 'svelte';

  let sources = [];
  let isAdmin = false;
  let loadingSources = true;

  // Navegação interna da view: 'search' | 'providers'
  let currentSubTab = 'search';

  // Pesquisa
  let query = '';
  let selectedSource = '';
  let selectedFormat = 'all'; // 'all', 'epub', 'pdf', 'mobi'
  let viewMode = 'grid'; // 'grid' | 'table'
  let results = [];
  let errors = {};
  let searching = false;
  let searched = false;

  // Estado de downloads
  let downloadingMap = {}; // key -> 'downloading' | 'done' | 'error'
  let downloadErrorMap = {};

  // Estado de configuração / edição de provedores
  let editingId = null;
  let configForms = {}; // id -> { values: {}, enabled: true }
  let testingId = null;
  let testResults = {}; // id -> { ok: bool, msg: string }
  let savingId = null;
  let saveFeedback = {}; // id -> { ok: bool, msg: string }
  let showSecrets = {}; // id -> bool

  // Notificações globais
  let toastMsg = '';
  let toastKind = 'ok';
  let toastTimeout = null;

  function showToast(msg, kind = 'ok') {
    toastMsg = msg;
    toastKind = kind;
    if (toastTimeout) clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => { toastMsg = ''; }, 4500);
  }

  onMount(async () => {
    await fetchSources();
  });

  async function fetchSources() {
    loadingSources = true;
    try {
      const res = await fetch('/api/sources');
      if (res.ok) {
        const data = await res.json();
        sources = data.sources || [];
        isAdmin = !!data.is_admin;
        // Inicializa formulários
        const forms = {};
        for (const s of sources) {
          forms[s.id] = {
            enabled: s.enabled,
            values: { ...(s.values || {}) }
          };
        }
        configForms = forms;
      }
    } catch (err) {
      console.error('Falha ao obter fontes:', err);
      showToast('Falha ao comunicar com o servidor para obter fontes.', 'err');
    } finally {
      loadingSources = false;
    }
  }

  // Deteção inteligente de MD5 (Anna's Archive)
  $: isMd5Input = /^[0-9a-fA-F]{32}$/.test(query.trim()) || /annas-archive[^\/]*\/md5\/([0-9a-fA-F]{32})/i.test(query.trim());

  async function handleSearch() {
    const q = query.trim();
    if (q.length < 2) {
      showToast('Por favor, digite pelo menos 2 caracteres para pesquisar.', 'err');
      return;
    }

    searching = true;
    searched = true;
    errors = {};

    try {
      const params = new URLSearchParams({ q });
      if (selectedSource) {
        params.set('source', selectedSource);
      }
      const res = await fetch('/api/sources/search?' + params.toString());
      const data = await res.json();

      if (!res.ok) {
        showToast(data.error || 'Erro na pesquisa federada.', 'err');
        results = [];
        return;
      }

      results = data.results || [];
      errors = data.errors || {};

      if (results.length === 0 && Object.keys(errors).length === 0) {
        showToast('Nenhum livro encontrado para este termo.', 'info');
      }
    } catch (err) {
      console.error('Search error:', err);
      showToast('Erro de rede ao consultar fontes.', 'err');
    } finally {
      searching = false;
    }
  }

  function onKeyDown(e) {
    if (e.key === 'Enter') handleSearch();
    if (e.key === 'Escape') query = '';
  }

  async function triggerDownload(item) {
    const key = `${item.source}:${item.id}`;
    downloadingMap = { ...downloadingMap, [key]: 'downloading' };
    downloadErrorMap = { ...downloadErrorMap, [key]: null };

    try {
      const res = await fetch('/api/sources/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source: item.source, id: item.id })
      });
      const data = await res.json();

      if (!res.ok) {
        downloadingMap = { ...downloadingMap, [key]: 'error' };
        downloadErrorMap = { ...downloadErrorMap, [key]: data.error || 'Falha no download' };
        showToast(data.error || 'Erro ao descarregar livro.', 'err');
        return;
      }

      downloadingMap = { ...downloadingMap, [key]: 'done' };
      showToast(`✓ "${item.title}" transferido para a pasta de ingestão contínua!`, 'ok');
    } catch (err) {
      downloadingMap = { ...downloadingMap, [key]: 'error' };
      downloadErrorMap = { ...downloadErrorMap, [key]: 'Erro de ligação ao servidor' };
      showToast('Erro de rede ao iniciar download.', 'err');
    }
  }

  async function testConnection(sourceId) {
    testingId = sourceId;
    testResults = { ...testResults, [sourceId]: null };

    const form = configForms[sourceId] || { values: {} };

    try {
      const res = await fetch(`/api/sources/${sourceId}/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values })
      });
      const data = await res.json();

      if (res.ok) {
        testResults = { ...testResults, [sourceId]: { ok: true, msg: data.message || 'Ligação bem-sucedida!' } };
      } else {
        testResults = { ...testResults, [sourceId]: { ok: false, msg: data.error || 'Falha no teste de conexão.' } };
      }
    } catch (err) {
      testResults = { ...testResults, [sourceId]: { ok: false, msg: 'Erro de rede ao contactar servidor.' } };
    } finally {
      testingId = null;
    }
  }

  async function saveProviderConfig(sourceId) {
    savingId = sourceId;
    saveFeedback = { ...saveFeedback, [sourceId]: null };

    const form = configForms[sourceId] || { values: {}, enabled: true };

    try {
      const res = await fetch(`/api/sources/${sourceId}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values, enabled: form.enabled })
      });
      const data = await res.json();

      if (res.ok) {
        saveFeedback = { ...saveFeedback, [sourceId]: { ok: true, msg: 'Definições guardadas com sucesso!' } };
        sources = data.sources || sources;
        showToast(`Configurações de ${sourceName(sourceId)} gravadas!`, 'ok');
      } else {
        saveFeedback = { ...saveFeedback, [sourceId]: { ok: false, msg: data.error || 'Erro ao guardar.' } };
        showToast(data.error || 'Erro ao guardar definições.', 'err');
      }
    } catch (err) {
      saveFeedback = { ...saveFeedback, [sourceId]: { ok: false, msg: 'Erro de comunicação.' } };
      showToast('Erro de rede ao guardar definições.', 'err');
    } finally {
      savingId = null;
    }
  }

  function setMirror(sourceId, url) {
    if (configForms[sourceId]) {
      configForms[sourceId].values.base_url = url;
      configForms = { ...configForms };
    }
  }

  // Filtragem local de formato
  $: filteredResults = results.filter(r => {
    if (selectedFormat === 'all') return true;
    const fmt = (r.format || '').toLowerCase();
    if (selectedFormat === 'epub') return fmt.includes('epub');
    if (selectedFormat === 'pdf') return fmt.includes('pdf');
    if (selectedFormat === 'mobi') return fmt.includes('mobi') || fmt.includes('azw');
    if (selectedFormat === 'comic') return fmt.includes('cbr') || fmt.includes('cbz');
    return true;
  });

  function sourceMeta(id) {
    switch (id) {
      case 'annas':
        return {
          badgeClass: 'badge-annas',
          color: '#d946ef',
          label: "Anna's Archive",
          tag: 'Fast API',
          icon: '⚡'
        };
      case 'zlibrary':
        return {
          badgeClass: 'badge-zlib',
          color: '#06b6d4',
          label: 'Z-Library',
          tag: 'eAPI v2',
          icon: '🌐'
        };
      case 'openlibrary':
        return {
          badgeClass: 'badge-openlib',
          color: '#10b981',
          label: 'Open Library / IA',
          tag: 'Acesso Público',
          icon: '🏛️'
        };
      case 'gutenberg':
        return {
          badgeClass: 'badge-gutenberg',
          color: '#f59e0b',
          label: 'Project Gutenberg',
          tag: 'Domínio Público',
          icon: '📜'
        };
      default:
        return {
          badgeClass: 'badge-default',
          color: '#8b949e',
          label: id,
          tag: 'Provedor',
          icon: '📖'
        };
    }
  }

  function sourceName(id) {
    const s = sources.find(item => item.id === id);
    return s ? s.name : id;
  }

  $: configuredCount = sources.filter(s => s.configured && s.enabled).length;
</script>

<div class="sources-hub">
  <!-- Toast Notification Bar -->
  {#if toastMsg}
    <div class="hub-toast glass" class:err={toastKind === 'err'} class:info={toastKind === 'info'}>
      <span class="toast-dot"></span>
      <span class="toast-text">{toastMsg}</span>
      <button class="toast-close" on:click={() => { toastMsg = ''; }}>✕</button>
    </div>
  {/if}

  <!-- Header / Command Deck -->
  <header class="hub-header glass">
    <div class="header-main">
      <div class="header-title-group">
        <div class="hub-badge-pill">
          <span class="pulse-point"></span>
          <span class="hub-badge-text">CENTRAL DE AQUISIÇÃO FEDERADA</span>
        </div>
        <h1>Fontes &amp; Descoberta Global</h1>
        <p class="hub-subtitle">
          Pesquisa e obtenção simultânea em índices mundiais com 1-Click Ingest para a biblioteca.
        </p>
      </div>

      <!-- Quick Live Provider Health Pills -->
      <div class="provider-health-strip">
        {#each sources as s}
          {@const meta = sourceMeta(s.id)}
          <button 
            class="health-pill" 
            class:active-pill={s.enabled && s.configured} 
            class:unconfigured={!s.configured}
            on:click={() => { currentSubTab = 'providers'; editingId = s.id; }}
            title={s.enabled && s.configured ? `${s.name}: Operacional` : `${s.name}: Clique para configurar`}>
            <span class="health-dot" style="background: {s.enabled && s.configured ? meta.color : '#ef4444'};"></span>
            <span class="health-label">{meta.label}</span>
            <span class="health-state">{s.enabled && s.configured ? 'Pronto' : 'Requer Chave'}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- Main Sub-tabs Switcher -->
    <div class="tab-switcher-row">
      <div class="switcher-pills">
        <button 
          class="switch-tab-btn" 
          class:active={currentSubTab === 'search'} 
          on:click={() => { currentSubTab = 'search'; }}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <span>Pesquisa Federada</span>
          {#if results.length > 0}
            <span class="counter-badge">{results.length}</span>
          {/if}
        </button>

        <button 
          class="switch-tab-btn" 
          class:active={currentSubTab === 'providers'} 
          on:click={() => { currentSubTab = 'providers'; }}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"></circle>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
          </svg>
          <span>Gestor de Provedores &amp; API Keys</span>
          <span class="provider-count-pill">{configuredCount}/{sources.length} Ativos</span>
        </button>
      </div>
    </div>
  </header>

  <!-- ===================================================================== -->
  <!-- TAB 1: PESQUISA FEDERADA                                               -->
  <!-- ===================================================================== -->
  {#if currentSubTab === 'search'}
    <div class="search-tab-content">
      <!-- Search Omnibox & Filter Toolbar -->
      <section class="omnibox-card glass">
        <div class="omnibox-main-row">
          <div class="input-container">
            <svg class="search-glass-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input 
              type="text" 
              placeholder="Pesquisar por título, autor, assunto ou colar link / hash MD5 do Anna's Archive..." 
              bind:value={query} 
              on:keydown={onKeyDown}
              class="omnibox-input"
            />
            {#if query}
              <button class="input-clear-btn" on:click={() => { query = ''; }} title="Limpar">✕</button>
            {/if}
          </div>

          <button class="btn-execute-search" on:click={handleSearch} disabled={searching}>
            {#if searching}
              <div class="micro-spinner"></div>
              <span>A Pesquisar...</span>
            {:else}
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              <span>Explorar Catálogos</span>
            {/if}
          </button>
        </div>

        <!-- MD5 Direct Download Smart Tip -->
        {#if isMd5Input}
          <div class="md5-smart-detect">
            <span class="md5-icon">⚡</span>
            <div class="md5-text">
              <strong>Hash MD5 / URL Direto Identificado!</strong>
              <span>Esta consulta fará resolução de download direta via Anna's Archive Fast API sem passar por busca de texto.</span>
            </div>
          </div>
        {/if}

        <!-- Filter Chips Row -->
        <div class="filters-row">
          <div class="filter-group">
            <span class="filter-group-label">Fonte:</span>
            <div class="chips-list">
              <button 
                class="chip-btn" 
                class:active={selectedSource === ''} 
                on:click={() => { selectedSource = ''; }}>
                Todas as Fontes
              </button>
              {#each sources as s}
                {@const meta = sourceMeta(s.id)}
                <button 
                  class="chip-btn" 
                  class:active={selectedSource === s.id} 
                  on:click={() => { selectedSource = s.id; }}>
                  <span class="chip-color-dot" style="background: {meta.color};"></span>
                  {meta.label}
                </button>
              {/each}
            </div>
          </div>

          <div class="filter-group">
            <span class="filter-group-label">Formato:</span>
            <div class="chips-list">
              {#each ['all', 'epub', 'pdf', 'mobi', 'comic'] as fmt}
                <button 
                  class="chip-btn" 
                  class:active={selectedFormat === fmt} 
                  on:click={() => { selectedFormat = fmt; }}>
                  {fmt === 'all' ? 'Todos' : fmt === 'comic' ? 'CBR/CBZ' : fmt.toUpperCase()}
                </button>
              {/each}
            </div>
          </div>

          <!-- View Mode Toggle -->
          <div class="view-toggle-wrap">
            <button 
              class="view-toggle-btn" 
              class:active={viewMode === 'grid'} 
              on:click={() => { viewMode = 'grid'; }} 
              title="Grelha de Capas">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="7" height="7"></rect>
                <rect x="14" y="3" width="7" height="7"></rect>
                <rect x="14" y="14" width="7" height="7"></rect>
                <rect x="3" y="14" width="7" height="7"></rect>
              </svg>
            </button>
            <button 
              class="view-toggle-btn" 
              class:active={viewMode === 'table'} 
              on:click={() => { viewMode = 'table'; }} 
              title="Tabela Detalhada">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="8" y1="6" x2="21" y2="6"></line>
                <line x1="8" y1="12" x2="21" y2="12"></line>
                <line x1="8" y1="18" x2="21" y2="18"></line>
                <line x1="3" y1="6" x2="3.01" y2="6"></line>
                <line x1="3" y1="12" x2="3.01" y2="12"></line>
                <line x1="3" y1="18" x2="3.01" y2="18"></line>
              </svg>
            </button>
          </div>
        </div>
      </section>

      <!-- Provider Diagnostics / Errors Banners (if any source failed) -->
      {#if Object.keys(errors).length > 0}
        <div class="diagnostics-deck">
          {#each Object.entries(errors) as [srcId, errMsg]}
            {@const meta = sourceMeta(srcId)}
            <div class="diag-banner glass">
              <span class="diag-icon">⚠️</span>
              <div class="diag-body">
                <strong>{meta.label}</strong>
                <span>{errMsg}</span>
              </div>
              <button 
                class="diag-action-btn" 
                on:click={() => { currentSubTab = 'providers'; editingId = srcId; }}>
                Configurar Fonte
              </button>
            </div>
          {/each}
        </div>
      {/if}

      <!-- Results Showcase -->
      {#if searching}
        <div class="loading-stage glass">
          <div class="radar-spinner"></div>
          <h3>A consultar os índices globais...</h3>
          <p>A comunicar com APIs mundiais e a filtrar metadados e ficheiros válidos.</p>
        </div>
      {:else if filteredResults.length > 0}
        {#if viewMode === 'grid'}
          <!-- Visual Bookshelf Grid -->
          <div class="bookshelf-grid">
            {#each filteredResults as r (r.source + ':' + r.id)}
              {@const meta = sourceMeta(r.source)}
              {@const itemKey = `${r.source}:${r.id}`}
              {@const dlState = downloadingMap[itemKey]}

              <article class="book-card-stage glass">
                <!-- Cover Showcase -->
                <div class="book-cover-wrap">
                  {#if r.cover}
                    <img 
                      src={r.cover} 
                      alt={r.title} 
                      class="book-cover-img" 
                      loading="lazy" 
                      referrerpolicy="no-referrer"
                      on:error={(e) => { e.currentTarget.style.display = 'none'; e.currentTarget.nextElementSibling.style.display = 'flex'; }}
                    />
                    <div class="fallback-cover" style="display: none;">
                      <div class="fallback-graphic">📖</div>
                      <div class="fallback-title">{r.title}</div>
                    </div>
                  {:else}
                    <div class="fallback-cover">
                      <div class="fallback-graphic">{meta.icon}</div>
                      <div class="fallback-title">{r.title}</div>
                      <div class="fallback-author">{r.author || 'Autor Desconhecido'}</div>
                    </div>
                  {/if}

                  <!-- Format Badge Floating -->
                  <div class="cover-format-badge">
                    {(r.format || 'LIVRO').toUpperCase()}
                  </div>

                  <!-- Source Stamp -->
                  <div class="cover-source-stamp" style="background: {meta.color}22; border-color: {meta.color}55; color: {meta.color};">
                    <span>{meta.icon}</span>
                    <span>{meta.label}</span>
                  </div>
                </div>

                <!-- Book Information -->
                <div class="book-card-body">
                  <h4 class="card-title" title={r.title}>{r.title}</h4>
                  <div class="card-author" title={r.author || 'Autor Desconhecido'}>
                    {r.author || 'Autor Desconhecido'}
                  </div>

                  <div class="card-meta-row">
                    {#if r.year}
                      <span class="meta-tag year-tag">{r.year}</span>
                    {/if}
                    {#if r.size}
                      <span class="meta-tag size-tag">{r.size}</span>
                    {/if}
                  </div>

                  <!-- Ingest Action Button -->
                  <div class="card-action-wrap">
                    <button 
                      class="btn-ingest" 
                      class:state-downloading={dlState === 'downloading'}
                      class:state-done={dlState === 'done'}
                      class:state-error={dlState === 'error'}
                      on:click={() => triggerDownload(r)}
                      disabled={dlState === 'downloading' || dlState === 'done'}>
                      {#if dlState === 'downloading'}
                        <div class="micro-spinner"></div>
                        <span>A Transferir...</span>
                      {:else if dlState === 'done'}
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                          <polyline points="20 6 9 17 4 12"></polyline>
                        </svg>
                        <span>✓ Na Pasta Ingest</span>
                      {:else if dlState === 'error'}
                        <span>Tentar Novamente</span>
                      {:else}
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                          <polyline points="7 10 12 15 17 10"></polyline>
                          <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                        <span>1-Click Ingest</span>
                      {/if}
                    </button>
                    {#if downloadErrorMap[itemKey]}
                      <div class="card-error-text">{downloadErrorMap[itemKey]}</div>
                    {/if}
                  </div>
                </div>
              </article>
            {/each}
          </div>
        {:else}
          <!-- Technical Inspector Table -->
          <div class="technical-table-wrap glass">
            <table class="tech-table">
              <thead>
                <tr>
                  <th style="width: 50px;">Capa</th>
                  <th>Título &amp; Autor</th>
                  <th>Fonte</th>
                  <th>Ano</th>
                  <th>Formato</th>
                  <th>Tamanho</th>
                  <th style="text-align: right;">Ação</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredResults as r (r.source + ':' + r.id)}
                  {@const meta = sourceMeta(r.source)}
                  {@const itemKey = `${r.source}:${r.id}`}
                  {@const dlState = downloadingMap[itemKey]}
                  <tr>
                    <td class="table-cover-cell">
                      {#if r.cover}
                        <img src={r.cover} alt="" class="table-cover-thumb" loading="lazy" referrerpolicy="no-referrer" />
                      {:else}
                        <div class="table-cover-placeholder">{meta.icon}</div>
                      {/if}
                    </td>
                    <td>
                      <div class="table-book-title">{r.title}</div>
                      <div class="table-book-author">{r.author || '—'}</div>
                    </td>
                    <td>
                      <span class="source-tag" style="background: {meta.color}18; color: {meta.color}; border: 1px solid {meta.color}40;">
                        {meta.label}
                      </span>
                    </td>
                    <td>
                      <span class="mono-text">{r.year || '—'}</span>
                    </td>
                    <td>
                      <span class="format-pill">{(r.format || 'EPUB').toUpperCase()}</span>
                    </td>
                    <td>
                      <span class="mono-text">{r.size || '—'}</span>
                    </td>
                    <td style="text-align: right;">
                      <button 
                        class="btn-table-dl" 
                        class:btn-done={dlState === 'done'}
                        on:click={() => triggerDownload(r)}
                        disabled={dlState === 'downloading' || dlState === 'done'}>
                        {dlState === 'downloading' ? 'A transferir...' : dlState === 'done' ? '✓ Enviado' : 'Descarregar'}
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      {:else if searched}
        <div class="empty-results-stage glass">
          <div class="empty-icon">🔍</div>
          <h3>Nenhum resultado encontrado</h3>
          <p>Tente simplificar os termos de busca, alterar o formato selecionado ou verificar se os provedores estão com chaves ativas.</p>
        </div>
      {:else}
        <!-- Initial Welcome Discovery Banner -->
        <div class="welcome-discovery-deck glass">
          <div class="welcome-col">
            <span class="welcome-kicker">BIBLIOTECA UNIVERSAL</span>
            <h2>Mais de 100 Milhões de Livros ao Seu Alcance</h2>
            <p>
              Consulte simultaneamente acervos abertos de domínio público e redes de preservação digital.
              Qualquer livro selecionado é descarregado automaticamente com metadados e adicionado à biblioteca pelo Calibre-Web Automated.
            </p>
            <div class="supported-sources-row">
              <div class="source-feature-badge">⚡ Anna's Archive (Fast API)</div>
              <div class="source-feature-badge">🌐 Z-Library (eAPI v2)</div>
              <div class="source-feature-badge">🏛️ Internet Archive / OpenLibrary</div>
              <div class="source-feature-badge">📜 Project Gutenberg (+70.000 livros)</div>
            </div>
          </div>
        </div>
      {/if}
    </div>
  {/if}

  <!-- ===================================================================== -->
  <!-- TAB 2: GESTOR DE PROVEDORES & CREDENCIAIS (API STUDIO)                 -->
  <!-- ===================================================================== -->
  {#if currentSubTab === 'providers'}
    <div class="providers-tab-content">
      <div class="studio-intro glass">
        <div class="studio-intro-text">
          <h2>Gestor de Provedores &amp; Chaves de Acesso</h2>
          <p>
            Configure as credenciais e endereços de cada rede. As credenciais são armazenadas de forma segura no servidor
            em ficheiro encriptado com permissões de administrador.
          </p>
        </div>
        {#if !isAdmin}
          <div class="admin-notice-pill">
            <span class="notice-icon">🔒</span>
            <span>Apenas administradores podem modificar as credenciais das fontes.</span>
          </div>
        {/if}
      </div>

      <div class="providers-grid">
        {#each sources as s}
          {@const meta = sourceMeta(s.id)}
          {@const form = configForms[s.id] || { values: {}, enabled: true }}
          {@const test = testResults[s.id]}
          {@const save = saveFeedback[s.id]}
          {@const isTesting = testingId === s.id}
          {@const isSaving = savingId === s.id}

          <div class="provider-studio-card glass" class:card-disabled={!form.enabled}>
            <!-- Card Header -->
            <div class="card-head-row">
              <div class="provider-identity">
                <div class="provider-icon-box" style="background: {meta.color}20; color: {meta.color}; border: 1px solid {meta.color}40;">
                  {meta.icon}
                </div>
                <div>
                  <h3 class="provider-title">{s.name}</h3>
                  <div class="provider-tag-line">
                    <span class="source-type-tag" style="color: {meta.color};">{meta.tag}</span>
                    <span class="status-indicator-dot" class:dot-active={s.enabled && s.configured}></span>
                    <span class="status-text">{s.enabled && s.configured ? 'Operacional' : s.configured ? 'Desativado' : 'Requer Configuração'}</span>
                  </div>
                </div>
              </div>

              <!-- Master Toggle Switch -->
              {#if isAdmin}
                <label class="switch-control" title={form.enabled ? 'Desativar este provedor' : 'Ativar este provedor'}>
                  <input type="checkbox" bind:checked={form.enabled} on:change={() => saveProviderConfig(s.id)} />
                  <span class="switch-slider"></span>
                </label>
              {/if}
            </div>

            <!-- Description -->
            <p class="provider-desc">{s.description}</p>

            <!-- Configuration Fields Form -->
            {#if s.fields && s.fields.length > 0}
              <div class="provider-form-body">
                {#each s.fields as f}
                  <div class="field-row">
                    <div class="field-label-row">
                      <label for="{s.id}_{f.key}" class="field-label">{f.label}</label>
                      {#if f.secret}
                        <button 
                          type="button" 
                          class="toggle-secret-btn" 
                          on:click={() => { showSecrets[s.id + '_' + f.key] = !showSecrets[s.id + '_' + f.key]; }}>
                          {showSecrets[s.id + '_' + f.key] ? 'Ocultar' : 'Revelar'}
                        </button>
                      {/if}
                    </div>

                    <div class="field-input-wrap">
                      <input 
                        id="{s.id}_{f.key}"
                        type={f.secret && !showSecrets[s.id + '_' + f.key] ? 'password' : 'text'}
                        bind:value={form.values[f.key]}
                        placeholder={f.default || 'Introduza o valor...'}
                        disabled={!isAdmin || isSaving}
                        class="field-text-input"
                      />
                    </div>
                  </div>
                {/each}

                <!-- Recommended Mirrors Quick Selection (For Anna's Archive) -->
                {#if s.id === 'annas'}
                  <div class="mirror-helpers-box">
                    <span class="helpers-title">Mirrors Públicos Recomendados:</span>
                    <div class="mirrors-pills-row">
                      {#each ['https://annas-archive.li', 'https://annas-archive.se', 'https://annas-archive.org', 'https://annas-archive.pm'] as mirror}
                        <button 
                          type="button" 
                          class="mirror-pill-btn" 
                          on:click={() => setMirror(s.id, mirror)}>
                          {mirror.replace('https://', '')}
                        </button>
                      {/each}
                    </div>
                    <span class="helpers-hint">💡 O Anna's Archive requer a chave secreta ("Secret Key") gerada na página de Definições de Membro do site.</span>
                  </div>
                {/if}

                <!-- Z-Library Credentials Helper -->
                {#if s.id === 'zlibrary'}
                  <div class="mirror-helpers-box">
                    <span class="helpers-hint">
                      💡 Como obter as credenciais: Faça login no seu domínio pessoal do Z-Library. Abra as Ferramentas de Programador (F12) &gt; Armazenamento &gt; Cookies, e copie os valores de <code>remix_userid</code> e <code>remix_userkey</code>.
                    </span>
                  </div>
                {/if}

                <!-- Live Test Connection & Save Action Toolbar -->
                {#if isAdmin}
                  <div class="provider-action-toolbar">
                    <button 
                      class="btn-test-connection" 
                      on:click={() => testConnection(s.id)} 
                      disabled={isTesting || isSaving}>
                      {#if isTesting}
                        <div class="micro-spinner"></div>
                        <span>A Testar...</span>
                      {:else}
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
                          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
                        </svg>
                        <span>Testar Conexão</span>
                      {/if}
                    </button>

                    <button 
                      class="btn-save-provider" 
                      on:click={() => saveProviderConfig(s.id)} 
                      disabled={isSaving}>
                      {#if isSaving}
                        <div class="micro-spinner"></div>
                        <span>A Guardar...</span>
                      {:else}
                        <span>Guardar Alterações</span>
                      {/if}
                    </button>
                  </div>

                  <!-- Test Connection Feedback Message -->
                  {#if test}
                    <div class="feedback-banner" class:feedback-ok={test.ok} class:feedback-err={!test.ok}>
                      <span class="fb-icon">{test.ok ? '✓' : '✕'}</span>
                      <span class="fb-msg">{test.msg}</span>
                    </div>
                  {/if}

                  <!-- Save Feedback Message -->
                  {#if save}
                    <div class="feedback-banner" class:feedback-ok={save.ok} class:feedback-err={!save.ok}>
                      <span class="fb-icon">{save.ok ? '✓' : '✕'}</span>
                      <span class="fb-msg">{save.msg}</span>
                    </div>
                  {/if}
                {/if}
              </div>
            {:else}
              <!-- Zero configuration required notice -->
              <div class="zero-config-box">
                <span class="zero-config-check">✓</span>
                <span class="zero-config-text">Este provedor está 100% ativo e funcional sem necessidade de chave ou autenticação de API.</span>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  /* ========================================================================= */
  /* THEME & BASE LAYOUT: OBSIDIAN EDITORIAL PRECISION                          */
  /* ========================================================================= */
  .sources-hub {
    max-width: 1750px;
    margin: 0 auto;
    padding: 24px 28px 80px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    color: var(--text-primary, #ffffff);
  }

  /* Glassmorphism Surface Base */
  .glass {
    background: rgba(18, 24, 38, 0.75);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.35);
  }

  /* Hub Header */
  .hub-header {
    padding: 28px 32px 18px;
    border-radius: 20px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    border: 1px solid rgba(56, 189, 248, 0.15);
    background: linear-gradient(135deg, rgba(15, 23, 42, 0.9) 0%, rgba(30, 41, 59, 0.7) 100%);
  }

  .header-main {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 20px;
  }

  .header-title-group h1 {
    font-size: 26px;
    font-weight: 800;
    letter-spacing: -0.03em;
    margin: 8px 0 6px;
    color: #ffffff;
  }

  .hub-subtitle {
    font-size: 13.5px;
    color: #94a3b8;
    max-width: 650px;
    line-height: 1.5;
  }

  .hub-badge-pill {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    background: rgba(56, 189, 248, 0.12);
    border: 1px solid rgba(56, 189, 248, 0.3);
    padding: 4px 10px;
    border-radius: 9999px;
  }

  .pulse-point {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #38bdf8;
    box-shadow: 0 0 8px #38bdf8;
    animation: pulseGlow 2s infinite ease-in-out;
  }

  @keyframes pulseGlow {
    0%, 100% { transform: scale(1); opacity: 1; }
    50% { transform: scale(1.4); opacity: 0.6; }
  }

  .hub-badge-text {
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    color: #38bdf8;
  }

  /* Provider Health Pills in Header */
  .provider-health-strip {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    align-items: center;
  }

  .health-pill {
    display: flex;
    align-items: center;
    gap: 7px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    padding: 6px 12px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .health-pill:hover {
    background: rgba(255, 255, 255, 0.06);
    border-color: rgba(255, 255, 255, 0.18);
    transform: translateY(-1px);
  }

  .health-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
  }

  .health-label {
    font-size: 11.5px;
    font-weight: 600;
    color: #e2e8f0;
  }

  .health-state {
    font-size: 10px;
    color: #64748b;
    padding-left: 3px;
  }

  /* Sub-tabs Navigation Switcher */
  .tab-switcher-row {
    border-top: 1px solid rgba(255, 255, 255, 0.07);
    padding-top: 16px;
  }

  .switcher-pills {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .switch-tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 18px;
    border-radius: 10px;
    background: transparent;
    border: 1px solid transparent;
    color: #94a3b8;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.18s ease;
  }

  .switch-tab-btn:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.04);
  }

  .switch-tab-btn.active {
    background: #0284c7;
    color: #ffffff;
    border-color: #38bdf8;
    box-shadow: 0 4px 16px rgba(2, 132, 199, 0.35);
  }

  .counter-badge {
    background: rgba(255, 255, 255, 0.2);
    font-size: 10.5px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 9999px;
  }

  .provider-count-pill {
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.1);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 9999px;
    color: #cbd5e1;
  }

  /* ========================================================================= */
  /* TAB 1: SEARCH & OMNIBOX DECK                                             */
  /* ========================================================================= */
  .search-tab-content {
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .omnibox-card {
    padding: 22px 26px;
    border-radius: 18px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .omnibox-main-row {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .input-container {
    flex: 1;
    min-width: 280px;
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-glass-icon {
    position: absolute;
    left: 16px;
    color: #0ea5e9;
    pointer-events: none;
  }

  .omnibox-input {
    width: 100%;
    height: 50px;
    background: rgba(10, 15, 28, 0.85);
    border: 1.5px solid rgba(56, 189, 248, 0.2);
    border-radius: 12px;
    padding: 0 44px 0 48px;
    color: #ffffff;
    font-size: 14.5px;
    outline: none;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .omnibox-input:focus {
    border-color: #38bdf8;
    background: rgba(10, 15, 28, 0.98);
    box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.18), 0 8px 24px rgba(0, 0, 0, 0.4);
  }

  .omnibox-input::placeholder {
    color: #64748b;
    font-size: 13.5px;
  }

  .input-clear-btn {
    position: absolute;
    right: 14px;
    background: transparent;
    border: none;
    color: #64748b;
    font-size: 13px;
    cursor: pointer;
    padding: 6px;
  }

  .input-clear-btn:hover {
    color: #ffffff;
  }

  .btn-execute-search {
    height: 50px;
    padding: 0 24px;
    background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
    border: 1px solid #38bdf8;
    border-radius: 12px;
    color: #ffffff;
    font-size: 13.5px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 9px;
    box-shadow: 0 4px 16px rgba(2, 132, 199, 0.35);
    transition: all 0.18s ease;
    flex-shrink: 0;
  }

  .btn-execute-search:hover:not(:disabled) {
    background: linear-gradient(135deg, #0369a1 0%, #0284c7 100%);
    box-shadow: 0 6px 20px rgba(2, 132, 199, 0.5);
    transform: translateY(-1px);
  }

  .btn-execute-search:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* MD5 Smart Detect Box */
  .md5-smart-detect {
    background: rgba(217, 70, 239, 0.12);
    border: 1px solid rgba(217, 70, 239, 0.35);
    border-radius: 10px;
    padding: 12px 18px;
    display: flex;
    align-items: center;
    gap: 12px;
    animation: fadeIn 0.2s ease;
  }

  .md5-icon {
    font-size: 22px;
  }

  .md5-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 12.5px;
    color: #f0abfc;
  }

  .md5-text strong {
    color: #ffffff;
    font-weight: 700;
  }

  /* Filter Chips Row */
  .filters-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
    padding-top: 6px;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .filter-group-label {
    font-size: 11.5px;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .chips-list {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .chip-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 12px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .chip-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
  }

  .chip-btn.active {
    background: rgba(56, 189, 248, 0.18);
    border-color: #38bdf8;
    color: #38bdf8;
    font-weight: 600;
  }

  .chip-color-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .view-toggle-wrap {
    display: flex;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    overflow: hidden;
  }

  .view-toggle-btn {
    padding: 6px 10px;
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .view-toggle-btn:hover {
    color: #ffffff;
  }

  .view-toggle-btn.active {
    background: rgba(255, 255, 255, 0.1);
    color: #38bdf8;
  }

  /* Diagnostics Banners */
  .diagnostics-deck {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .diag-banner {
    padding: 10px 18px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    background: rgba(248, 113, 113, 0.1);
    border: 1px solid rgba(248, 113, 113, 0.25);
    color: #fca5a5;
    font-size: 12.5px;
  }

  .diag-body {
    flex: 1;
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .diag-body strong {
    color: #ffffff;
  }

  .diag-action-btn {
    padding: 4px 12px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #ffffff;
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
  }

  /* ========================================================================= */
  /* VISUAL BOOKSHELF GRID                                                    */
  /* ========================================================================= */
  .bookshelf-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 22px;
  }

  .book-card-stage {
    border-radius: 16px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s ease;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .book-card-stage:hover {
    transform: translateY(-4px);
    box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(56, 189, 248, 0.3);
  }

  .book-cover-wrap {
    width: 100%;
    aspect-ratio: 2 / 2.9;
    position: relative;
    background: #090d16;
    overflow: hidden;
  }

  .book-cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.3s ease;
  }

  .book-card-stage:hover .book-cover-img {
    transform: scale(1.03);
  }

  /* Fallback artistic cover when no image available */
  .fallback-cover {
    width: 100%;
    height: 100%;
    background: linear-gradient(145deg, #1e293b 0%, #0f172a 100%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 20px;
    text-align: center;
    gap: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .fallback-graphic {
    font-size: 32px;
  }

  .fallback-title {
    font-size: 13px;
    font-weight: 700;
    color: #e2e8f0;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .fallback-author {
    font-size: 11px;
    color: #94a3b8;
  }

  .cover-format-badge {
    position: absolute;
    top: 10px;
    right: 10px;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #38bdf8;
    font-family: var(--font-mono, monospace);
    font-size: 10px;
    font-weight: 700;
    padding: 3px 7px;
    border-radius: 6px;
    letter-spacing: 0.04em;
  }

  .cover-source-stamp {
    position: absolute;
    bottom: 8px;
    left: 8px;
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 3px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 700;
    backdrop-filter: blur(8px);
  }

  /* Card Body */
  .book-card-body {
    padding: 14px 14px 16px;
    display: flex;
    flex-direction: column;
    flex: 1;
    gap: 6px;
  }

  .card-title {
    font-size: 13px;
    font-weight: 700;
    color: #ffffff;
    line-height: 1.35;
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 35px;
  }

  .card-author {
    font-size: 11.5px;
    color: #94a3b8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .card-meta-row {
    display: flex;
    gap: 6px;
    margin: 4px 0 8px;
    align-items: center;
  }

  .meta-tag {
    font-size: 10.5px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    padding: 2px 6px;
    border-radius: 4px;
    color: #cbd5e1;
    font-family: var(--font-mono, monospace);
  }

  .card-action-wrap {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .btn-ingest {
    width: 100%;
    height: 36px;
    border-radius: 9px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: all 0.18s ease;
  }

  .btn-ingest:hover:not(:disabled) {
    background: #0284c7;
    border-color: #38bdf8;
    box-shadow: 0 4px 14px rgba(2, 132, 199, 0.35);
  }

  .btn-ingest.state-downloading {
    background: rgba(56, 189, 248, 0.15);
    border-color: #38bdf8;
    color: #38bdf8;
  }

  .btn-ingest.state-done {
    background: rgba(34, 197, 94, 0.18);
    border-color: #22c55e;
    color: #4ade80;
    cursor: default;
  }

  .btn-ingest.state-error {
    background: rgba(239, 68, 68, 0.15);
    border-color: #ef4444;
    color: #f87171;
  }

  .card-error-text {
    font-size: 10.5px;
    color: #f87171;
    text-align: center;
    line-height: 1.2;
  }

  /* ========================================================================= */
  /* TECHNICAL TABLE                                                          */
  /* ========================================================================= */
  .technical-table-wrap {
    border-radius: 16px;
    overflow-x: auto;
  }

  .tech-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
    text-align: left;
  }

  .tech-table th {
    background: rgba(0, 0, 0, 0.25);
    padding: 12px 16px;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #94a3b8;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .tech-table td {
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    color: #cbd5e1;
    vertical-align: middle;
  }

  .table-cover-cell {
    padding: 8px 16px;
  }

  .table-cover-thumb {
    width: 32px;
    height: 44px;
    object-fit: cover;
    border-radius: 4px;
  }

  .table-cover-placeholder {
    width: 32px;
    height: 44px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
  }

  .table-book-title {
    font-weight: 600;
    color: #ffffff;
    margin-bottom: 2px;
  }

  .table-book-author {
    font-size: 11.5px;
    color: #94a3b8;
  }

  .source-tag {
    font-size: 11px;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 6px;
  }

  .format-pill {
    background: rgba(255, 255, 255, 0.08);
    font-family: var(--font-mono, monospace);
    font-size: 11px;
    font-weight: 700;
    padding: 3px 7px;
    border-radius: 4px;
    color: #38bdf8;
  }

  .mono-text {
    font-family: var(--font-mono, monospace);
    font-size: 11.5px;
  }

  .btn-table-dl {
    padding: 6px 14px;
    border-radius: 7px;
    background: #0284c7;
    border: 1px solid #38bdf8;
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-table-dl:hover:not(:disabled) {
    background: #0369a1;
  }

  .btn-table-dl.btn-done {
    background: rgba(34, 197, 94, 0.2);
    border-color: #22c55e;
    color: #4ade80;
  }

  /* ========================================================================= */
  /* TAB 2: PROVIDERS & CREDENTIALS STUDIO                                    */
  /* ========================================================================= */
  .providers-tab-content {
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .studio-intro {
    padding: 24px 28px;
    border-radius: 18px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
  }

  .studio-intro-text h2 {
    font-size: 18px;
    font-weight: 700;
    margin: 0 0 4px;
  }

  .studio-intro-text p {
    font-size: 13px;
    color: #94a3b8;
    margin: 0;
  }

  .admin-notice-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.35);
    color: #f59e0b;
    font-size: 12px;
    font-weight: 600;
    padding: 6px 14px;
    border-radius: 9999px;
  }

  .providers-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
    gap: 22px;
  }

  .provider-studio-card {
    border-radius: 18px;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 18px;
    transition: all 0.2s ease;
  }

  .provider-studio-card.card-disabled {
    opacity: 0.65;
  }

  .card-head-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
  }

  .provider-identity {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .provider-icon-box {
    width: 44px;
    height: 44px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    flex-shrink: 0;
  }

  .provider-title {
    font-size: 16px;
    font-weight: 700;
    margin: 0 0 2px;
    color: #ffffff;
  }

  .provider-tag-line {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .source-type-tag {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .status-indicator-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #ef4444;
  }

  .status-indicator-dot.dot-active {
    background: #22c55e;
    box-shadow: 0 0 6px #22c55e;
  }

  .status-text {
    font-size: 11px;
    color: #94a3b8;
  }

  .provider-desc {
    font-size: 12.5px;
    color: #94a3b8;
    line-height: 1.5;
    margin: 0;
  }

  /* Form Body in Card */
  .provider-form-body {
    display: flex;
    flex-direction: column;
    gap: 14px;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 16px;
  }

  .field-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .field-label {
    font-size: 11.5px;
    font-weight: 600;
    color: #cbd5e1;
  }

  .toggle-secret-btn {
    background: transparent;
    border: none;
    color: #38bdf8;
    font-size: 11px;
    cursor: pointer;
  }

  .toggle-secret-btn:hover {
    text-decoration: underline;
  }

  .field-text-input {
    width: 100%;
    height: 38px;
    background: rgba(15, 23, 42, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 8px;
    padding: 0 12px;
    color: #ffffff;
    font-size: 13px;
    font-family: var(--font-mono, monospace);
    outline: none;
    transition: border-color 0.15s ease;
  }

  .field-text-input:focus {
    border-color: #38bdf8;
  }

  .mirror-helpers-box {
    background: rgba(56, 189, 248, 0.06);
    border: 1px dashed rgba(56, 189, 248, 0.25);
    border-radius: 8px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .helpers-title {
    font-size: 11px;
    font-weight: 600;
    color: #38bdf8;
  }

  .mirrors-pills-row {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .mirror-pill-btn {
    font-size: 11px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.14);
    padding: 3px 8px;
    border-radius: 6px;
    color: #e2e8f0;
    cursor: pointer;
    font-family: var(--font-mono, monospace);
    transition: all 0.15s ease;
  }

  .mirror-pill-btn:hover {
    background: rgba(56, 189, 248, 0.25);
    border-color: #38bdf8;
    color: #ffffff;
  }

  .helpers-hint {
    font-size: 11px;
    color: #94a3b8;
    line-height: 1.4;
  }

  .helpers-hint code {
    background: rgba(0, 0, 0, 0.4);
    padding: 1px 4px;
    border-radius: 4px;
    color: #38bdf8;
  }

  /* Provider Card Actions */
  .provider-action-toolbar {
    display: flex;
    gap: 10px;
    margin-top: 4px;
  }

  .btn-test-connection {
    flex: 1;
    height: 36px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #e2e8f0;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: all 0.15s ease;
  }

  .btn-test-connection:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.14);
    color: #ffffff;
  }

  .btn-save-provider {
    flex: 1;
    height: 36px;
    border-radius: 8px;
    background: #0284c7;
    border: 1px solid #38bdf8;
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-save-provider:hover:not(:disabled) {
    background: #0369a1;
  }

  .feedback-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 500;
    animation: fadeIn 0.2s ease;
  }

  .feedback-banner.feedback-ok {
    background: rgba(34, 197, 94, 0.15);
    border: 1px solid rgba(34, 197, 94, 0.3);
    color: #4ade80;
  }

  .feedback-banner.feedback-err {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
  }

  .zero-config-box {
    background: rgba(34, 197, 94, 0.08);
    border: 1px solid rgba(34, 197, 94, 0.2);
    border-radius: 10px;
    padding: 14px 18px;
    display: flex;
    align-items: center;
    gap: 10px;
    color: #86efac;
    font-size: 12.5px;
  }

  .zero-config-check {
    font-size: 16px;
    font-weight: 800;
    color: #22c55e;
  }

  /* Master Switch Toggle */
  .switch-control {
    position: relative;
    display: inline-block;
    width: 42px;
    height: 24px;
    cursor: pointer;
  }

  .switch-control input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .switch-slider {
    position: absolute;
    inset: 0;
    background: rgba(255, 255, 255, 0.15);
    border-radius: 9999px;
    transition: 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .switch-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background: #ffffff;
    border-radius: 50%;
    transition: 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .switch-control input:checked + .switch-slider {
    background: #0284c7;
  }

  .switch-control input:checked + .switch-slider:before {
    transform: translateX(18px);
  }

  /* Global Hub Toast */
  .hub-toast {
    position: fixed;
    bottom: 30px;
    right: 30px;
    padding: 12px 20px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    gap: 10px;
    z-index: 9999;
    border: 1px solid rgba(34, 197, 94, 0.4);
    background: rgba(15, 23, 42, 0.95);
    color: #ffffff;
    font-size: 13px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.6);
    animation: slideUp 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .hub-toast.err {
    border-color: rgba(239, 68, 68, 0.4);
  }

  .hub-toast.err .toast-dot {
    background: #ef4444;
  }

  .toast-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #22c55e;
  }

  .toast-close {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    margin-left: 8px;
  }

  /* Loading Stages & Animations */
  .loading-stage, .empty-results-stage {
    padding: 80px 20px;
    border-radius: 18px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .radar-spinner {
    width: 48px;
    height: 48px;
    border: 3px solid rgba(56, 189, 248, 0.15);
    border-top-color: #38bdf8;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .micro-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-top-color: #ffffff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @keyframes slideUp {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
  }

  /* Welcome Banner */
  .welcome-discovery-deck {
    padding: 40px 48px;
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .welcome-kicker {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.1em;
    color: #38bdf8;
  }

  .welcome-col h2 {
    font-size: 24px;
    font-weight: 800;
    margin: 8px 0 12px;
  }

  .welcome-col p {
    font-size: 14px;
    color: #94a3b8;
    max-width: 720px;
    line-height: 1.6;
    margin-bottom: 24px;
  }

  .supported-sources-row {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .source-feature-badge {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 8px 16px;
    border-radius: 10px;
    font-size: 12px;
    font-weight: 600;
    color: #e2e8f0;
  }

  @media (max-width: 900px) {
    .sources-hub {
      padding: 16px 16px 90px;
    }
    .providers-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
