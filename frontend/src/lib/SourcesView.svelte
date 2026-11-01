<script>
  // @ts-nocheck
  import { onMount } from 'svelte';

  // Sub-abas: 'search' | 'settings'
  let activeSubTab = 'search';

  let sources = [];
  let isAdmin = false;
  let loadingSources = true;

  // Pesquisa
  let query = '';
  let selectedSource = '';
  let selectedFormat = 'all';
  let results = [];
  let errors = {};
  let searching = false;
  let searched = false;

  // Estado de download
  let downloadingMap = {};
  let downloadErrorMap = {};

  // Configuração das fontes
  let configForms = {};
  let isSaving = {};
  let saveFeedback = {};
  let isTesting = {};
  let testFeedback = {};
  let showSecretMap = {};

  // Alerta global do CWA
  let alertMessage = '';
  let alertType = 'success';
  let alertTimer = null;

  function showAlert(msg, type = 'success') {
    alertMessage = msg;
    alertType = type;
    if (alertTimer) clearTimeout(alertTimer);
    alertTimer = setTimeout(() => { alertMessage = ''; }, 4500);
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
    } catch (e) {
      showAlert('Failed to connect to backend sources API.', 'danger');
    } finally {
      loadingSources = false;
    }
  }

  $: isDirectMd5 = /^[0-9a-fA-F]{32}$/.test(query.trim()) || /annas-archive[^\/]*\/md5\/([0-9a-fA-F]{32})/i.test(query.trim());

  async function executeSearch() {
    const q = query.trim();
    if (q.length < 2) {
      showAlert('Please enter at least 2 characters to search.', 'warning');
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
        showAlert(data.error || 'Search request failed.', 'danger');
        results = [];
        return;
      }

      results = data.results || [];
      errors = data.errors || {};
    } catch (e) {
      showAlert('Network error while querying book sources.', 'danger');
    } finally {
      searching = false;
    }
  }

  function onKeyDown(e) {
    if (e.key === 'Enter') executeSearch();
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
        downloadErrorMap = { ...downloadErrorMap, [key]: data.error || 'Download failed' };
        showAlert(data.error || 'Failed to download book.', 'danger');
        return;
      }

      downloadingMap = { ...downloadingMap, [key]: 'done' };
      showAlert(`✓ "${item.title}" successfully sent to continuous ingest directory!`, 'success');
    } catch (e) {
      downloadingMap = { ...downloadingMap, [key]: 'error' };
      downloadErrorMap = { ...downloadErrorMap, [key]: 'Network error' };
      showAlert('Network error while requesting download.', 'danger');
    }
  }

  async function testConnection(sourceId) {
    isTesting = { ...isTesting, [sourceId]: true };
    testFeedback = { ...testFeedback, [sourceId]: null };

    const form = configForms[sourceId] || { values: {} };

    try {
      const res = await fetch(`/api/sources/${sourceId}/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values })
      });
      const data = await res.json();

      if (res.ok) {
        testFeedback = { ...testFeedback, [sourceId]: { ok: true, msg: data.message || 'Connection successful!' } };
      } else {
        testFeedback = { ...testFeedback, [sourceId]: { ok: false, msg: data.error || 'Connection failed.' } };
      }
    } catch (e) {
      testFeedback = { ...testFeedback, [sourceId]: { ok: false, msg: 'Network error during connection test.' } };
    } finally {
      isTesting = { ...isTesting, [sourceId]: false };
    }
  }

  async function saveConfig(sourceId) {
    isSaving = { ...isSaving, [sourceId]: true };
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
        saveFeedback = { ...saveFeedback, [sourceId]: { ok: true, msg: 'Configuration saved successfully.' } };
        sources = data.sources || sources;
        showAlert(`Settings for ${sourceName(sourceId)} saved.`, 'success');
      } else {
        saveFeedback = { ...saveFeedback, [sourceId]: { ok: false, msg: data.error || 'Failed to save settings.' } };
        showAlert(data.error || 'Failed to save settings.', 'danger');
      }
    } catch (e) {
      saveFeedback = { ...saveFeedback, [sourceId]: { ok: false, msg: 'Network communication error.' } };
      showAlert('Network error while saving.', 'danger');
    } finally {
      isSaving = { ...isSaving, [sourceId]: false };
    }
  }

  function sourceName(id) {
    const s = sources.find(item => item.id === id);
    return s ? s.name : id;
  }

  $: filteredResults = results.filter(r => {
    if (selectedFormat === 'all') return true;
    const fmt = (r.format || '').toLowerCase();
    if (selectedFormat === 'epub') return fmt.includes('epub');
    if (selectedFormat === 'pdf') return fmt.includes('pdf');
    if (selectedFormat === 'mobi') return fmt.includes('mobi') || fmt.includes('azw');
    return true;
  });

  $: activeCount = sources.filter(s => s.configured && s.enabled).length;
</script>

<div class="cwa-container">
  <!-- CWA Page Header (Identical to /me & /cwa-settings) -->
  <div class="cwa-page-header">
    <div>
      <h1 class="cwa-page-title">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right: 8px; vertical-align: middle;">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="2" y1="12" x2="22" y2="12"></line>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
        </svg>
        Calibre-Web Automated Book Sources
      </h1>
      <p class="cwa-page-subtitle">
        Federated multi-source search and direct automated acquisition from Anna's Archive, Z-Library, Open Library &amp; Project Gutenberg.
      </p>
    </div>
    <div>
      <span class="cwa-badge cwa-badge-info">CWA Ingestion Engine</span>
    </div>
  </div>

  <!-- Global Alert Bar -->
  {#if alertMessage}
    <div class="cwa-alert cwa-alert-{alertType}">
      <span>{alertMessage}</span>
      <button class="cwa-alert-close" on:click={() => { alertMessage = ''; }}>✕</button>
    </div>
  {/if}

  <!-- CWA Metrics Grid (Matching /me pattern) -->
  <div class="cwa-metrics-grid">
    <div class="cwa-metric-box">
      <div class="cwa-metric-label">Active Sources</div>
      <div class="cwa-metric-value cyan">{activeCount} / {sources.length}</div>
    </div>
    <div class="cwa-metric-box">
      <div class="cwa-metric-label">Supported Formats</div>
      <div class="cwa-metric-value green">EPUB • PDF • MOBI</div>
    </div>
    <div class="cwa-metric-box">
      <div class="cwa-metric-label">Pipeline Target</div>
      <div class="cwa-metric-value amber">Auto-Ingest Watcher</div>
    </div>
  </div>

  <!-- Main CWA Panel -->
  <div class="cwa-panel">
    <!-- Sub-navigation Tabs matching CWA design -->
    <div class="cwa-subnav">
      <button 
        class="cwa-tab-btn" 
        class:active={activeSubTab === 'search'} 
        on:click={() => { activeSubTab = 'search'; }}>
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"></circle>
          <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
        </svg>
        <span>Search &amp; Download</span>
        {#if results.length > 0}
          <span class="cwa-badge cwa-badge-info" style="margin-left: 6px;">{results.length}</span>
        {/if}
      </button>

      <button 
        class="cwa-tab-btn" 
        class:active={activeSubTab === 'settings'} 
        on:click={() => { activeSubTab = 'settings'; }}>
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3"></circle>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
        </svg>
        <span>Provider Settings &amp; API Keys</span>
      </button>
    </div>

    <!-- =================================================================== -->
    <!-- SUB-TAB 1: SEARCH & ACQUISITION                                      -->
    <!-- =================================================================== -->
    {#if activeSubTab === 'search'}
      <div class="cwa-search-section">
        <!-- Search Controls -->
        <div class="cwa-search-bar-row">
          <div class="search-input-wrap">
            <input 
              type="text" 
              class="form-control cwa-input" 
              placeholder="Search by title, author, ISBN or paste Anna's Archive MD5 hash..." 
              bind:value={query} 
              on:keydown={onKeyDown}
            />
            {#if query}
              <button class="clear-btn" on:click={() => { query = ''; }}>✕</button>
            {/if}
          </div>

          <div class="search-select-wrap">
            <select class="form-control cwa-select" bind:value={selectedSource}>
              <option value="">All Active Sources</option>
              {#each sources as s}
                <option value={s.id}>{s.name} {!s.configured ? '(No API key)' : ''}</option>
              {/each}
            </select>
          </div>

          <div class="search-select-wrap" style="max-width: 170px;">
            <select class="form-control cwa-select" bind:value={selectedFormat}>
              <option value="all">All Formats</option>
              <option value="epub">EPUB</option>
              <option value="pdf">PDF</option>
              <option value="mobi">MOBI / AZW3</option>
            </select>
          </div>

          <button class="btn btn-primary" on:click={executeSearch} disabled={searching} style="height: 38px;">
            {#if searching}
              <span>Searching...</span>
            {:else}
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 4px; vertical-align: middle;">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              <span>Search</span>
            {/if}
          </button>
        </div>

        {#if isDirectMd5}
          <div class="cwa-hint-box" style="margin-top: 12px;">
            <strong>MD5 Direct Download Detected:</strong> The query matches an MD5 hash. The book will be downloaded directly via Anna's Archive Fast Download API.
          </div>
        {/if}

        <!-- Error messages if sources failed -->
        {#if Object.keys(errors).length > 0}
          <div style="margin-top: 14px;">
            {#each Object.entries(errors) as [srcId, errMsg]}
              <div class="cwa-alert cwa-alert-warning" style="margin-bottom: 8px;">
                <strong>{sourceName(srcId)}:</strong> {errMsg}
                {#if isAdmin}
                  <a href="javascript:void(0)" on:click={() => { activeSubTab = 'settings'; }} style="margin-left: 10px; text-decoration: underline; color: inherit;">Configure</a>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        <!-- Table Results matching CWA Table style -->
        <div style="margin-top: 24px;">
          {#if searching}
            <div style="text-align: center; padding: 40px; color: #94a3b8;">
              <p>Querying external book providers...</p>
            </div>
          {:else if filteredResults.length > 0}
            <h2 class="cwa-panel-title" style="margin-bottom: 12px; font-size: 16px;">
              Search Results ({filteredResults.length})
            </h2>
            <table class="table cwa-table">
              <thead>
                <tr>
                  <th style="width: 44px;">Cover</th>
                  <th>Title &amp; Author</th>
                  <th>Source</th>
                  <th>Format</th>
                  <th>Year</th>
                  <th>Size</th>
                  <th style="text-align: right; width: 140px;">Action</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredResults as r (r.source + ':' + r.id)}
                  {@const key = `${r.source}:${r.id}`}
                  {@const state = downloadingMap[key]}
                  <tr>
                    <td style="padding: 8px 12px;">
                      {#if r.cover}
                        <img src={r.cover} alt="" class="cwa-book-thumb" loading="lazy" referrerpolicy="no-referrer" />
                      {:else}
                        <div class="cwa-book-thumb-empty">📖</div>
                      {/if}
                    </td>
                    <td>
                      <div class="cwa-row-title">{r.title}</div>
                      <div class="cwa-row-author">{r.author || 'Unknown author'}</div>
                    </td>
                    <td>
                      <span class="cwa-badge cwa-badge-info">{sourceName(r.source)}</span>
                    </td>
                    <td>
                      <span class="cwa-tag-format">{(r.format || 'EPUB').toUpperCase()}</span>
                    </td>
                    <td>
                      <span class="cwa-mono-cell">{r.year || '—'}</span>
                    </td>
                    <td>
                      <span class="cwa-mono-cell">{r.size || '—'}</span>
                    </td>
                    <td style="text-align: right;">
                      <button 
                        class="btn btn-primary btn-sm"
                        class:btn-success={state === 'done'}
                        class:btn-danger={state === 'error'}
                        on:click={() => triggerDownload(r)}
                        disabled={state === 'downloading' || state === 'done'}>
                        {#if state === 'downloading'}
                          <span>Downloading...</span>
                        {:else if state === 'done'}
                          <span>✓ Ingested</span>
                        {:else if state === 'error'}
                          <span>Retry</span>
                        {:else}
                          <span>1-Click Ingest</span>
                        {/if}
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {:else if searched}
            <div style="text-align: center; padding: 48px 20px; color: #94a3b8;">
              <p style="font-size: 15px; color: #ffffff; margin-bottom: 6px;">No books found</p>
              <p>No results matched "{query}". Try adjusting keywords or check provider configurations.</p>
            </div>
          {:else}
            <!-- Empty initial state with CWA Operation Cards -->
            <div style="padding: 10px 0;">
              <p style="color: #94a3b8; font-size: 13.5px; margin-bottom: 16px;">
                Enter a title or author above to search across active digital libraries. Selected books are transferred directly into the CWA auto-ingestion folder.
              </p>
              
              <div class="cwa-quick-providers-overview">
                {#each sources as s}
                  <div class="cwa-op-card" style="margin-bottom: 10px; cursor: pointer;" on:click={() => { selectedSource = s.id; }}>
                    <div>
                      <div class="op-title">
                        <span class="glyphicon glyphicon-book" style="color: #38bdf8; margin-right: 8px;"></span>
                        {s.name}
                      </div>
                      <div class="op-desc">{s.description}</div>
                    </div>
                    <span class="cwa-badge {s.enabled && s.configured ? 'cwa-badge-info' : 'cwa-badge-warning'}">
                      {s.enabled && s.configured ? 'Active' : 'Setup Required'}
                    </span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- =================================================================== -->
    <!-- SUB-TAB 2: PROVIDER SETTINGS & API KEYS                             -->
    <!-- =================================================================== -->
    {#if activeSubTab === 'settings'}
      <div class="cwa-settings-section">
        <h2 class="cwa-panel-title">
          <span class="glyphicon glyphicon-cog"></span>
          Book Provider Credentials &amp; Services
        </h2>

        {#if !isAdmin}
          <div class="cwa-alert cwa-alert-warning">
            Administrator permissions are required to modify provider settings and keys.
          </div>
        {/if}

        <div class="cwa-providers-config-list">
          {#each sources as s}
            {@const form = configForms[s.id] || { values: {}, enabled: true }}
            {@const test = testFeedback[s.id]}
            {@const save = saveFeedback[s.id]}
            {@const testing = isTesting[s.id]}
            {@const saving = isSaving[s.id]}

            <div class="cwa-setting-row" style="flex-direction: column; align-items: stretch; gap: 14px; margin-bottom: 20px; padding-bottom: 20px; border-bottom: 1px solid rgba(255,255,255,0.06);">
              <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                <div class="cwa-setting-info">
                  <div class="cwa-setting-title">
                    <span class="glyphicon glyphicon-cloud" style="color: #38bdf8; margin-right: 6px;"></span>
                    {s.name}
                    <span class="cwa-badge {s.enabled && s.configured ? 'cwa-badge-info' : 'cwa-badge-warning'}" style="margin-left: 8px;">
                      {s.enabled && s.configured ? 'Active' : s.configured ? 'Disabled' : 'Requires API Key'}
                    </span>
                  </div>
                  <div class="cwa-setting-desc">{s.description}</div>
                </div>

                {#if isAdmin}
                  <div class="cwa-setting-control">
                    <label class="cwa-switch" for="switch-{s.id}">
                      <input 
                        id="switch-{s.id}" 
                        type="checkbox" 
                        bind:checked={form.enabled} 
                        on:change={() => saveConfig(s.id)} 
                      />
                      <span class="cwa-slider"></span>
                    </label>
                  </div>
                {/if}
              </div>

              {#if s.fields && s.fields.length > 0}
                <div class="cwa-fields-group">
                  {#each s.fields as f}
                    <div class="cwa-field-row">
                      <div style="display: flex; justify-content: space-between; margin-bottom: 4px;">
                        <label for="input-{s.id}-{f.key}" style="font-size: 12px; font-weight: 600; color: #cbd5e1;">
                          {f.label}
                        </label>
                        {#if f.secret}
                          <a href="javascript:void(0)" on:click={() => { showSecretMap[s.id + '_' + f.key] = !showSecretMap[s.id + '_' + f.key]; }} style="font-size: 11px; color: #38bdf8;">
                            {showSecretMap[s.id + '_' + f.key] ? 'Hide' : 'Show'}
                          </a>
                        {/if}
                      </div>

                      <input 
                        id="input-{s.id}-{f.key}"
                        type={f.secret && !showSecretMap[s.id + '_' + f.key] ? 'password' : 'text'}
                        class="form-control cwa-input"
                        placeholder={f.default || ''}
                        bind:value={form.values[f.key]}
                        disabled={!isAdmin || saving}
                      />
                    </div>
                  {/each}

                  {#if s.id === 'annas'}
                    <div class="cwa-hint-box">
                      <strong>Recommended Mirrors:</strong>
                      <div style="display: flex; gap: 6px; margin-top: 4px;">
                        {#each ['https://annas-archive.li', 'https://annas-archive.se', 'https://annas-archive.org'] as mirror}
                          <button class="btn btn-default btn-xs" type="button" on:click={() => { form.values.base_url = mirror; configForms = { ...configForms }; }}>
                            {mirror.replace('https://', '')}
                          </button>
                        {/each}
                      </div>
                    </div>
                  {/if}

                  {#if s.id === 'zlibrary'}
                    <div class="cwa-hint-box">
                      <strong>Z-Library Credentials:</strong> Log in to your personal domain, press F12 &gt; Application/Storage &gt; Cookies, and copy the <code>remix_userid</code> and <code>remix_userkey</code> values.
                    </div>
                  {/if}

                  {#if test}
                    <div class="cwa-alert {test.ok ? 'cwa-alert-success' : 'cwa-alert-warning'}" style="margin: 8px 0;">
                      {test.msg}
                    </div>
                  {/if}

                  {#if save}
                    <div class="cwa-alert {save.ok ? 'cwa-alert-success' : 'cwa-alert-warning'}" style="margin: 8px 0;">
                      {save.msg}
                    </div>
                  {/if}

                  {#if isAdmin}
                    <div style="display: flex; gap: 10px; margin-top: 10px;">
                      <button class="btn btn-default" type="button" on:click={() => testConnection(s.id)} disabled={testing || saving}>
                        {testing ? 'Testing connection...' : 'Test Connection'}
                      </button>
                      <button class="btn btn-primary" type="button" on:click={() => saveConfig(s.id)} disabled={saving}>
                        {saving ? 'Saving...' : 'Save Settings'}
                      </button>
                    </div>
                  {/if}
                </div>
              {:else}
                <div style="font-size: 13px; color: #94a3b8; padding-top: 6px;">
                  This is a public service. No API keys or credentials are required.
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  /* ========================================================================= */
  /* CALIBRE-WEB AUTOMATED (CWA) 2026 DESIGN SYSTEM                            */
  /* Fully unified with /me, /admin, /cwa-settings & /cwa-stats                */
  /* ========================================================================= */
  .cwa-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 24px 20px 80px;
  }

  /* Page Header matching /me */
  .cwa-page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    margin-bottom: 24px;
    flex-wrap: wrap;
    gap: 16px;
  }

  .cwa-page-title {
    font-size: 22px;
    font-weight: 700;
    letter-spacing: -0.3px;
    color: #ffffff;
    margin: 0 0 4px;
  }

  .cwa-page-subtitle {
    font-size: 13px;
    color: #94a3b8;
    margin: 0;
    max-width: 720px;
    line-height: 1.5;
  }

  /* Badges */
  .cwa-badge {
    font-size: 11px;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 6px;
    letter-spacing: 0.4px;
    text-transform: uppercase;
  }

  .cwa-badge-info {
    background: rgba(56, 189, 248, 0.15);
    color: #38bdf8;
    border: 1px solid rgba(56, 189, 248, 0.3);
  }

  .cwa-badge-warning {
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    border: 1px solid rgba(245, 158, 11, 0.3);
  }

  /* Metrics Grid (Identical to /me) */
  .cwa-metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
  }

  .cwa-metric-box {
    background: rgba(18, 25, 38, 0.65);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 12px;
    padding: 16px 20px;
  }

  .cwa-metric-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #94a3b8;
    margin-bottom: 6px;
    font-weight: 600;
  }

  .cwa-metric-value {
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.5px;
  }

  .cwa-metric-value.cyan { color: #38bdf8; }
  .cwa-metric-value.green { color: #34d399; }
  .cwa-metric-value.amber { color: #fbbf24; }

  /* CWA Panel Surface (Identical to /me & /cwa-settings) */
  .cwa-panel {
    background: rgba(18, 25, 38, 0.75);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 24px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.4);
  }

  .cwa-panel-title {
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.3px;
    color: #ffffff;
    margin: 0 0 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  }

  /* Sub-tabs inside CWA panel */
  .cwa-subnav {
    display: flex;
    gap: 10px;
    margin-bottom: 20px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    padding-bottom: 12px;
  }

  .cwa-tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 8px;
    color: #94a3b8;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .cwa-tab-btn:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.05);
  }

  .cwa-tab-btn.active {
    background: rgba(56, 189, 248, 0.12);
    border-color: rgba(56, 189, 248, 0.35);
    color: #38bdf8;
  }

  /* Search row */
  .cwa-search-bar-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
  }

  .search-input-wrap {
    flex: 1;
    min-width: 280px;
    position: relative;
  }

  .cwa-input, .cwa-select {
    height: 38px;
    background: rgba(11, 15, 25, 0.8) !important;
    border: 1px solid rgba(255, 255, 255, 0.12) !important;
    border-radius: 8px !important;
    color: #ffffff !important;
    font-size: 13.5px !important;
    padding: 0 12px !important;
    outline: none;
  }

  .cwa-input:focus, .cwa-select:focus {
    border-color: #38bdf8 !important;
    box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2) !important;
  }

  .clear-btn {
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
  }

  /* CWA Table Style (Matching /me tables) */
  .cwa-table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 0;
  }

  .cwa-table th {
    background: rgba(11, 15, 25, 0.6);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1) !important;
    color: #94a3b8;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 600;
    padding: 10px 14px !important;
  }

  .cwa-table td {
    padding: 12px 14px !important;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05) !important;
    vertical-align: middle !important;
    color: #e2e8f0;
    font-size: 13px;
  }

  .cwa-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .cwa-book-thumb {
    width: 32px;
    height: 46px;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .cwa-book-thumb-empty {
    width: 32px;
    height: 46px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
  }

  .cwa-row-title {
    font-weight: 600;
    color: #ffffff;
  }

  .cwa-row-author {
    font-size: 12px;
    color: #94a3b8;
  }

  .cwa-tag-format, .cwa-mono-cell {
    font-family: var(--font-mono, monospace);
    font-size: 12px;
  }

  /* Operation Cards matching /me */
  .cwa-op-card {
    background: rgba(19, 27, 42, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    padding: 14px 18px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    transition: all 0.2s ease;
    text-decoration: none;
    color: inherit;
  }

  .cwa-op-card:hover {
    background: rgba(30, 41, 59, 0.7);
    border-color: rgba(56, 189, 248, 0.3);
    text-decoration: none;
    transform: translateY(-1px);
  }

  .op-title {
    font-size: 14px;
    font-weight: 600;
    color: #ffffff;
    margin-bottom: 2px;
  }

  .op-desc {
    font-size: 12px;
    color: #94a3b8;
  }

  /* Settings rows matching /cwa-settings */
  .cwa-setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .cwa-setting-title {
    font-size: 14.5px;
    font-weight: 600;
    color: #ffffff;
    margin-bottom: 3px;
    display: flex;
    align-items: center;
  }

  .cwa-setting-desc {
    font-size: 12.5px;
    color: #94a3b8;
    line-height: 1.4;
  }

  /* CWA Native Switch Slider (Exact match to /cwa-settings) */
  .cwa-switch {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    margin: 0;
    cursor: pointer;
  }

  .cwa-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .cwa-slider {
    position: absolute;
    cursor: pointer;
    top: 0; left: 0; right: 0; bottom: 0;
    background-color: rgba(255, 255, 255, 0.15);
    transition: 0.25s ease;
    border-radius: 24px;
  }

  .cwa-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.25s ease;
    border-radius: 50%;
  }

  .cwa-switch input:checked + .cwa-slider {
    background-color: #3b82f6;
  }

  .cwa-switch input:checked + .cwa-slider:before {
    transform: translateX(20px);
  }

  /* Alert banners */
  .cwa-alert {
    padding: 10px 14px;
    border-radius: 8px;
    font-size: 13px;
    margin-bottom: 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .cwa-alert-success {
    background: rgba(34, 197, 94, 0.12);
    border: 1px solid rgba(34, 197, 94, 0.3);
    color: #4ade80;
  }

  .cwa-alert-warning {
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #fcd34d;
  }

  .cwa-alert-danger {
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
  }

  .cwa-alert-close {
    background: transparent;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 12px;
  }

  .cwa-hint-box {
    background: rgba(56, 189, 248, 0.08);
    border: 1px solid rgba(56, 189, 248, 0.2);
    border-radius: 6px;
    padding: 8px 12px;
    font-size: 12px;
    color: #bae6fd;
  }

  .cwa-fields-group {
    background: rgba(11, 15, 25, 0.4);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 10px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .cwa-field-row {
    max-width: 540px;
  }

  /* Buttons matching cwa.css */
  .btn-xs {
    padding: 2px 8px !important;
    font-size: 11px !important;
  }
</style>
