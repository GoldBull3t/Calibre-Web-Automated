<script>
  // @ts-nocheck
  import { onMount } from 'svelte';

  // Navegação por abas: 'search' | 'settings'
  let activeTab = 'search';

  // Dados das fontes
  let sources = [];
  let isAdmin = false;
  let loadingSources = true;

  // Estado de pesquisa
  let query = '';
  let selectedSource = '';
  let selectedFormat = 'all';
  let results = [];
  let errors = {};
  let isSearching = false;
  let hasSearched = false;

  // Estado de transferências: id -> 'idle' | 'downloading' | 'success' | 'error'
  let downloadState = {};
  let downloadError = {};

  // Estado de configuração: sourceId -> { values: {}, enabled: bool }
  let configForms = {};
  let isSaving = {};
  let saveMessage = {};
  let isTesting = {};
  let testMessage = {};
  let showPassword = {};

  // Mensagem de feedback de topo
  let notification = null;
  let notificationTimer = null;

  function notify(text, type = 'info') {
    notification = { text, type };
    if (notificationTimer) clearTimeout(notificationTimer);
    notificationTimer = setTimeout(() => { notification = null; }, 4000);
  }

  onMount(async () => {
    await loadSources();
  });

  async function loadSources() {
    loadingSources = true;
    try {
      const res = await fetch('/api/sources');
      if (res.ok) {
        const data = await res.json();
        sources = data.sources || [];
        isAdmin = Boolean(data.is_admin);

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
      console.error('Erro ao carregar fontes:', err);
      notify('Não foi possível carregar a lista de fontes.', 'error');
    } finally {
      loadingSources = false;
    }
  }

  // Deteção de MD5 ou URL direto do Anna's Archive
  $: isDirectMd5 = /^[0-9a-fA-F]{32}$/.test(query.trim()) || /annas-archive[^\/]*\/md5\/([0-9a-fA-F]{32})/i.test(query.trim());

  async function handleSearch() {
    const q = query.trim();
    if (q.length < 2) {
      notify('Introduza pelo menos 2 caracteres para pesquisar.', 'error');
      return;
    }

    isSearching = true;
    hasSearched = true;
    errors = {};

    try {
      const params = new URLSearchParams({ q });
      if (selectedSource) {
        params.set('source', selectedSource);
      }
      const res = await fetch('/api/sources/search?' + params.toString());
      const data = await res.json();

      if (!res.ok) {
        notify(data.error || 'Erro ao efetuar pesquisa.', 'error');
        results = [];
        return;
      }

      results = data.results || [];
      errors = data.errors || {};
    } catch (err) {
      console.error('Erro na pesquisa:', err);
      notify('Erro de comunicação ao contactar o servidor.', 'error');
    } finally {
      isSearching = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter') handleSearch();
  }

  async function downloadBook(item) {
    const key = `${item.source}:${item.id}`;
    downloadState = { ...downloadState, [key]: 'downloading' };
    downloadError = { ...downloadError, [key]: null };

    try {
      const res = await fetch('/api/sources/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source: item.source, id: item.id })
      });
      const data = await res.json();

      if (!res.ok) {
        downloadState = { ...downloadState, [key]: 'error' };
        downloadError = { ...downloadError, [key]: data.error || 'Falha no download' };
        notify(data.error || 'Falha ao descarregar livro.', 'error');
        return;
      }

      downloadState = { ...downloadState, [key]: 'success' };
      notify(`"${item.title}" transferido para a pasta de ingestão.`, 'success');
    } catch (err) {
      downloadState = { ...downloadState, [key]: 'error' };
      downloadError = { ...downloadError, [key]: 'Erro de rede' };
      notify('Erro de rede ao iniciar download.', 'error');
    }
  }

  async function testConnection(sourceId) {
    isTesting = { ...isTesting, [sourceId]: true };
    testMessage = { ...testMessage, [sourceId]: null };

    const form = configForms[sourceId] || { values: {} };

    try {
      const res = await fetch(`/api/sources/${sourceId}/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values })
      });
      const data = await res.json();

      if (res.ok) {
        testMessage = { ...testMessage, [sourceId]: { ok: true, text: data.message || 'Ligação estabelecida com sucesso.' } };
      } else {
        testMessage = { ...testMessage, [sourceId]: { ok: false, text: data.error || 'Falha no teste de ligação.' } };
      }
    } catch (err) {
      testMessage = { ...testMessage, [sourceId]: { ok: false, text: 'Erro de comunicação ao testar ligação.' } };
    } finally {
      isTesting = { ...isTesting, [sourceId]: false };
    }
  }

  async function saveConfig(sourceId) {
    isSaving = { ...isSaving, [sourceId]: true };
    saveMessage = { ...saveMessage, [sourceId]: null };

    const form = configForms[sourceId] || { values: {}, enabled: true };

    try {
      const res = await fetch(`/api/sources/${sourceId}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values, enabled: form.enabled })
      });
      const data = await res.json();

      if (res.ok) {
        saveMessage = { ...saveMessage, [sourceId]: { ok: true, text: 'Configuração guardada com sucesso.' } };
        sources = data.sources || sources;
        notify(`Configuração de ${sourceName(sourceId)} guardada.`, 'success');
      } else {
        saveMessage = { ...saveMessage, [sourceId]: { ok: false, text: data.error || 'Erro ao guardar configuração.' } };
        notify(data.error || 'Erro ao guardar configuração.', 'error');
      }
    } catch (err) {
      saveMessage = { ...saveMessage, [sourceId]: { ok: false, text: 'Erro de comunicação ao guardar.' } };
      notify('Erro de rede ao guardar configuração.', 'error');
    } finally {
      isSaving = { ...isSaving, [sourceId]: false };
    }
  }

  function sourceName(id) {
    const s = sources.find(item => item.id === id);
    return s ? s.name : id;
  }

  // Filtragem local por formato
  $: filteredResults = results.filter(r => {
    if (selectedFormat === 'all') return true;
    const fmt = (r.format || '').toLowerCase();
    if (selectedFormat === 'epub') return fmt.includes('epub');
    if (selectedFormat === 'pdf') return fmt.includes('pdf');
    if (selectedFormat === 'mobi') return fmt.includes('mobi') || fmt.includes('azw');
    return true;
  });

  $: configuredCount = sources.filter(s => s.configured && s.enabled).length;
</script>

<div class="sources-view">
  <!-- Notification banner -->
  {#if notification}
    <div class="notification-banner" class:success={notification.type === 'success'} class:error={notification.type === 'error'}>
      <span class="notification-text">{notification.text}</span>
      <button class="notification-close" on:click={() => { notification = null; }} aria-label="Fechar">✕</button>
    </div>
  {/if}

  <!-- Page Header -->
  <header class="page-header">
    <div class="page-header-content">
      <h1 class="page-title">Fontes de Livros</h1>
      <p class="page-description">Pesquisa federada em catálogos externos com importação direta para a pasta de ingestão.</p>
    </div>

    <!-- Tab navigation -->
    <nav class="tabs-nav" aria-label="Abas de fontes">
      <button 
        class="tab-link" 
        class:active={activeTab === 'search'} 
        on:click={() => { activeTab = 'search'; }}>
        Pesquisar
        {#if results.length > 0}
          <span class="tab-badge">{results.length}</span>
        {/if}
      </button>

      <button 
        class="tab-link" 
        class:active={activeTab === 'settings'} 
        on:click={() => { activeTab = 'settings'; }}>
        Configurações
        <span class="tab-badge">{configuredCount}/{sources.length} ativas</span>
      </button>
    </nav>
  </header>

  <!-- ===================================================================== -->
  <!-- TAB: PESQUISAR                                                         -->
  <!-- ===================================================================== -->
  {#if activeTab === 'search'}
    <section class="search-panel">
      <!-- Search Form -->
      <div class="search-form-card">
        <div class="search-inputs-grid">
          <div class="form-field query-field">
            <label for="search-query" class="field-label">Termo de pesquisa ou identificador</label>
            <div class="input-with-button">
              <input 
                id="search-query"
                type="text" 
                class="text-input" 
                placeholder="Título, autor, ISBN ou hash MD5..." 
                bind:value={query}
                on:keydown={handleKeydown}
              />
              {#if query}
                <button type="button" class="clear-input-btn" on:click={() => { query = ''; }} aria-label="Limpar texto">✕</button>
              {/if}
            </div>
          </div>

          <div class="form-field source-field">
            <label for="search-source" class="field-label">Fonte</label>
            <select id="search-source" class="select-input" bind:value={selectedSource}>
              <option value="">Todas as fontes ativas</option>
              {#each sources as s}
                <option value={s.id}>{s.name} {!s.configured ? '(Requer chave)' : ''}</option>
              {/each}
            </select>
          </div>

          <div class="form-field format-field">
            <label for="search-format" class="field-label">Formato</label>
            <select id="search-format" class="select-input" bind:value={selectedFormat}>
              <option value="all">Todos os formatos</option>
              <option value="epub">EPUB</option>
              <option value="pdf">PDF</option>
              <option value="mobi">MOBI / AZW3</option>
            </select>
          </div>

          <div class="form-actions">
            <button type="button" class="btn btn-primary search-submit-btn" on:click={handleSearch} disabled={isSearching}>
              {#if isSearching}
                <span class="spinner-inline"></span>
                <span>A pesquisar...</span>
              {:else}
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
                <span>Pesquisar</span>
              {/if}
            </button>
          </div>
        </div>

        {#if isDirectMd5}
          <div class="direct-md5-note">
            <span class="note-label">Deteção direta:</span> O termo introduzido corresponde a um hash MD5. O download será efetuado diretamente através da API do Anna's Archive sem necessidade de pesquisa textual.
          </div>
        {/if}
      </div>

      <!-- Provider Errors Banner (if any) -->
      {#if Object.keys(errors).length > 0}
        <div class="provider-errors-list">
          {#each Object.entries(errors) as [srcId, errMsg]}
            <div class="error-item">
              <span class="error-source">{sourceName(srcId)}:</span>
              <span class="error-message">{errMsg}</span>
              {#if isAdmin}
                <button type="button" class="error-action-link" on:click={() => { activeTab = 'settings'; }}>
                  Configurar fonte
                </button>
              {/if}
            </div>
          {/each}
        </div>
      {/if}

      <!-- Results Area -->
      <div class="results-container">
        {#if isSearching}
          <!-- Loading State: Structured Skeletons -->
          <div class="table-card">
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 48px;"></th>
                  <th>Título e Autor</th>
                  <th>Fonte</th>
                  <th>Formato</th>
                  <th>Ano</th>
                  <th>Tamanho</th>
                  <th style="width: 140px; text-align: right;">Ação</th>
                </tr>
              </thead>
              <tbody>
                {#each [1, 2, 3, 4, 5] as _}
                  <tr class="skeleton-row">
                    <td><div class="skeleton skeleton-thumb"></div></td>
                    <td>
                      <div class="skeleton skeleton-text" style="width: 65%;"></div>
                      <div class="skeleton skeleton-text skeleton-sub" style="width: 40%;"></div>
                    </td>
                    <td><div class="skeleton skeleton-badge"></div></td>
                    <td><div class="skeleton skeleton-text" style="width: 40px;"></div></td>
                    <td><div class="skeleton skeleton-text" style="width: 35px;"></div></td>
                    <td><div class="skeleton skeleton-text" style="width: 50px;"></div></td>
                    <td style="text-align: right;"><div class="skeleton skeleton-button" style="margin-left: auto;"></div></td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else if filteredResults.length > 0}
          <!-- Success State: Structured Data Table -->
          <div class="table-card">
            <div class="table-header-info">
              <span class="results-count">{filteredResults.length} {filteredResults.length === 1 ? 'livro encontrado' : 'livros encontrados'}</span>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 48px;"></th>
                  <th>Título e Autor</th>
                  <th>Fonte</th>
                  <th>Formato</th>
                  <th>Ano</th>
                  <th>Tamanho</th>
                  <th style="width: 140px; text-align: right;">Ação</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredResults as r (r.source + ':' + r.id)}
                  {@const key = `${r.source}:${r.id}`}
                  {@const state = downloadState[key]}
                  <tr>
                    <td class="col-thumb">
                      {#if r.cover}
                        <img src={r.cover} alt="" class="book-thumb" loading="lazy" referrerpolicy="no-referrer" />
                      {:else}
                        <div class="book-thumb-empty">
                          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
                            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
                          </svg>
                        </div>
                      {/if}
                    </td>

                    <td class="col-title">
                      <div class="book-title" title={r.title}>{r.title}</div>
                      <div class="book-author">{r.author || 'Autor desconhecido'}</div>
                    </td>

                    <td class="col-source">
                      <span class="source-tag">{sourceName(r.source)}</span>
                    </td>

                    <td class="col-format">
                      <span class="format-tag">{(r.format || 'EPUB').toUpperCase()}</span>
                    </td>

                    <td class="col-year">
                      <span class="tabular-text">{r.year || '—'}</span>
                    </td>

                    <td class="col-size">
                      <span class="tabular-text">{r.size || '—'}</span>
                    </td>

                    <td class="col-action" style="text-align: right;">
                      <button 
                        type="button"
                        class="btn btn-secondary btn-download" 
                        class:btn-success={state === 'success'}
                        class:btn-error={state === 'error'}
                        on:click={() => downloadBook(r)}
                        disabled={state === 'downloading' || state === 'success'}>
                        {#if state === 'downloading'}
                          <span class="spinner-inline"></span>
                          <span>A transferir...</span>
                        {:else if state === 'success'}
                          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                            <polyline points="20 6 9 17 4 12"></polyline>
                          </svg>
                          <span>Importado</span>
                        {:else if state === 'error'}
                          <span>Tentar de novo</span>
                        {:else}
                          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                          </svg>
                          <span>Descarregar</span>
                        {/if}
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else if hasSearched}
          <!-- Empty State after search -->
          <div class="state-box">
            <h3 class="state-title">Nenhum livro encontrado</h3>
            <p class="state-desc">Não foram encontrados resultados para "{query}". Tente ajustar os termos de pesquisa ou remover os filtros de formato.</p>
          </div>
        {:else}
          <!-- Initial Empty State -->
          <div class="state-box">
            <h3 class="state-title">Pesquisa federada de livros</h3>
            <p class="state-desc">Escreva o título, autor ou assunto no campo acima para pesquisar simultaneamente nas fontes ativas. Os livros descarregados são colocados na pasta de ingestão contínua do Calibre-Web.</p>
          </div>
        {/if}
      </div>
    </section>
  {/if}

  <!-- ===================================================================== -->
  <!-- TAB: CONFIGURAÇÕES DE FONTES                                           -->
  <!-- ===================================================================== -->
  {#if activeTab === 'settings'}
    <section class="settings-panel">
      {#if !isAdmin}
        <div class="notice-callout">
          <span class="notice-icon">ℹ</span>
          <span>Apenas utilizadores com perfil de Administrador podem alterar as configurações e credenciais das fontes.</span>
        </div>
      {/if}

      <div class="providers-list">
        {#each sources as s}
          {@const form = configForms[s.id] || { values: {}, enabled: true }}
          {@const test = testMessage[s.id]}
          {@const save = saveMessage[s.id]}
          {@const testing = isTesting[s.id]}
          {@const saving = isSaving[s.id]}

          <article class="provider-card">
            <!-- Header do Provedor -->
            <div class="provider-card-header">
              <div class="provider-title-area">
                <h3 class="provider-name">{s.name}</h3>
                <span class="status-badge" class:status-active={s.enabled && s.configured} class:status-inactive={!s.enabled && s.configured} class:status-unconfigured={!s.configured}>
                  {s.enabled && s.configured ? 'Ativo' : s.configured ? 'Desativado' : 'Requer configuração'}
                </span>
              </div>

              {#if isAdmin}
                <div class="provider-toggle">
                  <label class="toggle-switch" for="toggle-{s.id}">
                    <input 
                      id="toggle-{s.id}"
                      type="checkbox" 
                      bind:checked={form.enabled} 
                      on:change={() => saveConfig(s.id)}
                    />
                    <span class="toggle-slider"></span>
                  </label>
                  <span class="toggle-label">{form.enabled ? 'Ativado' : 'Desativado'}</span>
                </div>
              {/if}
            </div>

            <p class="provider-description">{s.description}</p>

            <!-- Campos do formulário se existirem -->
            {#if s.fields && s.fields.length > 0}
              <div class="provider-fields-form">
                {#each s.fields as f}
                  <div class="form-group">
                    <div class="field-header">
                      <label for="field-{s.id}-{f.key}" class="field-label">
                        {f.label}
                        {#if !f.optional}<span class="required-indicator">*</span>{/if}
                      </label>
                      {#if f.secret}
                        <button 
                          type="button" 
                          class="link-button" 
                          on:click={() => { showPassword[s.id + '_' + f.key] = !showPassword[s.id + '_' + f.key]; }}>
                          {showPassword[s.id + '_' + f.key] ? 'Ocultar' : 'Mostrar'}
                        </button>
                      {/if}
                    </div>

                    <input 
                      id="field-{s.id}-{f.key}"
                      type={f.secret && !showPassword[s.id + '_' + f.key] ? 'password' : 'text'}
                      class="text-input" 
                      placeholder={f.default || ''}
                      bind:value={form.values[f.key]}
                      disabled={!isAdmin || saving}
                    />
                  </div>
                {/each}

                {#if s.id === 'annas'}
                  <div class="helper-text-block">
                    <span class="helper-label">Mirrors conhecidos:</span>
                    <div class="quick-links">
                      {#each ['https://annas-archive.li', 'https://annas-archive.se', 'https://annas-archive.org'] as mirror}
                        <button 
                          type="button" 
                          class="link-pill" 
                          on:click={() => { form.values.base_url = mirror; configForms = { ...configForms }; }}>
                          {mirror}
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}

                {#if s.id === 'zlibrary'}
                  <div class="helper-text-block">
                    Para utilizar a Z-Library, inicie sessão no seu domínio pessoal, abra as ferramentas de programador do navegador (F12) e consulte os cookies <code>remix_userid</code> e <code>remix_userkey</code>.
                  </div>
                {/if}

                <!-- Feedback Inline -->
                {#if test}
                  <div class="inline-feedback" class:feedback-success={test.ok} class:feedback-error={!test.ok}>
                    {test.text}
                  </div>
                {/if}

                {#if save}
                  <div class="inline-feedback" class:feedback-success={save.ok} class:feedback-error={!save.ok}>
                    {save.text}
                  </div>
                {/if}

                <!-- Ações do Provedor -->
                {#if isAdmin}
                  <div class="provider-actions">
                    <button 
                      type="button" 
                      class="btn btn-secondary" 
                      on:click={() => testConnection(s.id)} 
                      disabled={testing || saving}>
                      {#if testing}
                        <span class="spinner-inline"></span>
                        <span>A testar ligação...</span>
                      {:else}
                        <span>Testar ligação</span>
                      {/if}
                    </button>

                    <button 
                      type="button" 
                      class="btn btn-primary" 
                      on:click={() => saveConfig(s.id)} 
                      disabled={saving}>
                      {#if saving}
                        <span class="spinner-inline"></span>
                        <span>A guardar...</span>
                      {:else}
                        <span>Guardar</span>
                      {/if}
                    </button>
                  </div>
                {/if}
              </div>
            {:else}
              <div class="free-source-notice">
                Esta fonte está disponível publicamente e não requer chaves de autenticação ou configuração adicional.
              </div>
            {/if}
          </article>
        {/each}
      </div>
    </section>
  {/if}
</div>

<style>
  /* ========================================================================= */
  /* DESIGN SYSTEM: SHADCN / RADIX / ZINC NEUTRAL CLEAN                       */
  /* ========================================================================= */
  .sources-view {
    max-width: 1280px;
    margin: 0 auto;
    padding: 24px 20px 80px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* Page Header */
  .page-header {
    display: flex;
    flex-direction: column;
    gap: 16px;
    border-bottom: 1px solid var(--border-primary);
    padding-bottom: 16px;
  }

  .page-header-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .page-title {
    font-size: 24px;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--text-primary);
    margin: 0;
  }

  .page-description {
    font-size: 14px;
    color: var(--text-secondary);
    margin: 0;
    line-height: 1.5;
  }

  /* Tabs Navigation */
  .tabs-nav {
    display: flex;
    gap: 8px;
    border-bottom: 1px solid transparent;
  }

  .tab-link {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-secondary);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: color 0.15s ease, border-color 0.15s ease;
  }

  .tab-link:hover {
    color: var(--text-primary);
  }

  .tab-link.active {
    color: var(--text-primary);
    border-bottom-color: var(--accent);
    font-weight: 600;
  }

  .tab-badge {
    font-size: 12px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-secondary);
    padding: 1px 7px;
    border-radius: var(--radius-sm);
  }

  .tab-link.active .tab-badge {
    background: var(--accent-glow);
    color: var(--accent);
  }

  /* Notification Banner */
  .notification-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 14px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
  }

  .notification-banner.success {
    border-color: rgba(34, 197, 94, 0.35);
    background: rgba(34, 197, 94, 0.08);
    color: #4ade80;
  }

  .notification-banner.error {
    border-color: rgba(239, 68, 68, 0.35);
    background: rgba(239, 68, 68, 0.08);
    color: #f87171;
  }

  .notification-close {
    background: transparent;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 12px;
    padding: 2px 6px;
  }

  /* Search Panel */
  .search-panel {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .search-form-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-md);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .search-inputs-grid {
    display: grid;
    grid-template-columns: 1fr 200px 180px auto;
    gap: 12px;
    align-items: flex-end;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
  }

  .input-with-button {
    position: relative;
    display: flex;
    align-items: center;
  }

  .text-input, .select-input {
    width: 100%;
    height: 38px;
    background: var(--bg-primary);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 0 12px;
    color: var(--text-primary);
    font-size: 14px;
    outline: none;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .text-input:focus, .select-input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
  }

  .clear-input-btn {
    position: absolute;
    right: 8px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    font-size: 12px;
  }

  .clear-input-btn:hover {
    color: var(--text-primary);
  }

  .search-submit-btn {
    height: 38px;
    padding: 0 16px;
  }

  .direct-md5-note {
    font-size: 13px;
    color: var(--text-secondary);
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 8px 12px;
  }

  .note-label {
    font-weight: 600;
    color: var(--accent);
  }

  /* Provider Errors */
  .provider-errors-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .error-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    background: rgba(239, 68, 68, 0.08);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-sm);
    padding: 8px 12px;
    color: #fca5a5;
  }

  .error-source {
    font-weight: 600;
    color: #f87171;
  }

  .error-message {
    flex: 1;
  }

  .error-action-link {
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 12px;
    text-decoration: underline;
    cursor: pointer;
  }

  /* Data Table */
  .table-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .table-header-info {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-primary);
    font-size: 13px;
    color: var(--text-secondary);
  }

  .results-count {
    font-weight: 500;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
    text-align: left;
  }

  .data-table th {
    background: rgba(255, 255, 255, 0.02);
    padding: 10px 14px;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border-primary);
  }

  .data-table td {
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-primary);
    color: var(--text-primary);
    vertical-align: middle;
  }

  .data-table tbody tr:last-child td {
    border-bottom: none;
  }

  .data-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.02);
  }

  .col-thumb {
    width: 44px;
    padding-right: 0;
  }

  .book-thumb {
    width: 36px;
    height: 50px;
    object-fit: cover;
    border-radius: 3px;
    border: 1px solid var(--border-primary);
    display: block;
  }

  .book-thumb-empty {
    width: 36px;
    height: 50px;
    background: var(--bg-primary);
    border: 1px solid var(--border-primary);
    border-radius: 3px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .book-title {
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.4;
  }

  .book-author {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  .source-tag {
    font-size: 12px;
    color: var(--text-secondary);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    padding: 2px 8px;
    border-radius: var(--radius-sm);
  }

  .format-tag {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-secondary);
  }

  .tabular-text {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-secondary);
  }

  .btn-download {
    font-size: 12px;
    height: 32px;
    padding: 0 12px;
  }

  .btn-download.btn-success {
    background: rgba(34, 197, 94, 0.15);
    border-color: rgba(34, 197, 94, 0.3);
    color: #4ade80;
  }

  .btn-download.btn-error {
    border-color: rgba(239, 68, 68, 0.3);
    color: #f87171;
  }

  /* Empty & Loading States */
  .state-box {
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-md);
    padding: 48px 24px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .state-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .state-desc {
    font-size: 14px;
    color: var(--text-secondary);
    max-width: 520px;
    margin: 0;
    line-height: 1.5;
  }

  /* Structured Skeleton Loading */
  .skeleton {
    background: rgba(255, 255, 255, 0.06);
    border-radius: 4px;
  }

  .skeleton-thumb {
    width: 36px;
    height: 50px;
  }

  .skeleton-text {
    height: 14px;
    margin-bottom: 4px;
  }

  .skeleton-sub {
    height: 11px;
  }

  .skeleton-badge {
    width: 70px;
    height: 20px;
  }

  .skeleton-button {
    width: 90px;
    height: 30px;
  }

  .spinner-inline {
    width: 12px;
    height: 12px;
    border: 2px solid currentColor;
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
    display: inline-block;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* ========================================================================= */
  /* TAB: CONFIGURAÇÕES DE FONTES                                              */
  /* ========================================================================= */
  .settings-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .notice-callout {
    display: flex;
    align-items: center;
    gap: 10px;
    background: rgba(56, 189, 248, 0.08);
    border: 1px solid rgba(56, 189, 248, 0.2);
    border-radius: var(--radius-sm);
    padding: 10px 14px;
    font-size: 13px;
    color: #bae6fd;
  }

  .providers-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .provider-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-md);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .provider-card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }

  .provider-title-area {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .provider-name {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  .status-badge {
    font-size: 11px;
    font-weight: 500;
    padding: 2px 8px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-primary);
    color: var(--text-secondary);
  }

  .status-badge.status-active {
    background: rgba(34, 197, 94, 0.12);
    border-color: rgba(34, 197, 94, 0.25);
    color: #4ade80;
  }

  .status-badge.status-inactive {
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.25);
    color: #fcd34d;
  }

  .provider-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .toggle-switch {
    position: relative;
    display: inline-block;
    width: 36px;
    height: 20px;
    cursor: pointer;
  }

  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    inset: 0;
    background: rgba(255, 255, 255, 0.15);
    border-radius: 9999px;
    transition: 0.2s ease;
  }

  .toggle-slider::before {
    position: absolute;
    content: "";
    height: 14px;
    width: 14px;
    left: 3px;
    bottom: 3px;
    background: white;
    border-radius: 50%;
    transition: 0.2s ease;
  }

  .toggle-switch input:checked + .toggle-slider {
    background: var(--accent);
  }

  .toggle-switch input:checked + .toggle-slider::before {
    transform: translateX(16px);
  }

  .toggle-label {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .provider-description {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.5;
    margin: 0;
  }

  .provider-fields-form {
    display: flex;
    flex-direction: column;
    gap: 14px;
    border-top: 1px solid var(--border-primary);
    padding-top: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-width: 600px;
  }

  .field-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .required-indicator {
    color: var(--danger);
  }

  .link-button {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    text-decoration: underline;
  }

  .link-button:hover {
    color: var(--text-primary);
  }

  .helper-text-block {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .helper-text-block code {
    background: rgba(255, 255, 255, 0.08);
    padding: 1px 4px;
    border-radius: 3px;
    color: var(--text-primary);
  }

  .quick-links {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    margin-top: 4px;
  }

  .link-pill {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    padding: 2px 8px;
    font-size: 11px;
    cursor: pointer;
  }

  .link-pill:hover {
    color: var(--text-primary);
    border-color: var(--border-secondary);
  }

  .provider-actions {
    display: flex;
    gap: 10px;
    margin-top: 8px;
  }

  .inline-feedback {
    font-size: 13px;
    padding: 8px 12px;
    border-radius: var(--radius-sm);
  }

  .inline-feedback.feedback-success {
    background: rgba(34, 197, 94, 0.08);
    border: 1px solid rgba(34, 197, 94, 0.25);
    color: #4ade80;
  }

  .inline-feedback.feedback-error {
    background: rgba(239, 68, 68, 0.08);
    border: 1px solid rgba(239, 68, 68, 0.25);
    color: #f87171;
  }

  .free-source-notice {
    font-size: 13px;
    color: var(--text-secondary);
    border-top: 1px solid var(--border-primary);
    padding-top: 14px;
  }

  @media (max-width: 900px) {
    .search-inputs-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
