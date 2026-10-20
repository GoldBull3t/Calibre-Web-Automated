<script>
  // @ts-nocheck
  import { createEventDispatcher } from 'svelte';
  import { isOnline, canInstall, promptInstallPWA } from './pwa.js';
  export let activeTab = 'library';
  export let searchQuery = '';
  export let mamStatus = null;
  export let theme = 'dark';

  const dispatch = createEventDispatcher();

  function onSearch(e) {
    if (e.key === 'Enter') {
      dispatch('search', { query: searchQuery });
    }
  }

  function toggleTheme() {
    dispatch('toggleTheme');
  }

  function setTab(tab) {
    dispatch('navigate', { tab });
  }
</script>

<header class="navbar glass">
  <div class="nav-container">
    <!-- Brand -->
    <div class="brand" on:click={() => setTab('library')} on:keydown={(e) => { if (e.key === 'Enter') setTab('library'); }} role="button" tabindex="0">
      <div class="logo-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
          <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
        </svg>
      </div>
      <div class="brand-text">
        <span class="brand-title">Calibre-Web</span>
        <span class="badge badge-accent">NextGen</span>
      </div>
    </div>

    <!-- Navigation Pills -->
    <nav class="nav-links">
      <button 
        class="nav-pill" 
        class:active={activeTab === 'library'} 
        on:click={() => setTab('library')}>
        <svg viewBox="0 0 24 24"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
        <span>Biblioteca</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'browse'} 
        on:click={() => setTab('browse')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
        <span>Explorar</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'shelves'} 
        on:click={() => setTab('shelves')}>
        <svg viewBox="0 0 24 24"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
        <span>Estantes</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'fts'} 
        on:click={() => setTab('fts')}>
        <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <span>Pesquisa FTS</span>
      </button>

      <button 
        class="nav-pill mam-pill" 
        class:active={activeTab === 'mam'} 
        on:click={() => setTab('mam')}>
        <svg viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
        <span>MAM</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'duplicates'} 
        on:click={() => setTab('duplicates')}>
        <svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
        <span>Duplicados</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'cwa_settings'} 
        on:click={() => setTab('cwa_settings')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
        <span>Automação CWA</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'plugins'} 
        on:click={() => setTab('plugins')}>
        <svg viewBox="0 0 24 24"><path d="M12 2v4m0 12v4M2 12h4m12 0h4m-3.17-6.83l-2.83 2.83m-8 8l-2.83 2.83m0-13.66l2.83 2.83m8 8l2.83 2.83"></path></svg>
        <span>Plugins</span>
      </button>

      <button 
        class="nav-pill" 
        class:active={activeTab === 'stats'} 
        on:click={() => setTab('stats')}>
        <svg viewBox="0 0 24 24"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
        <span>Sistema</span>
      </button>
    </nav>

    <!-- Search Input -->
    <div class="search-wrap">
      <svg class="search-icon" viewBox="0 0 24 24">
        <circle cx="11" cy="11" r="8"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <input 
        type="text" 
        placeholder="Pesquisar livros, autores..." 
        bind:value={searchQuery} 
        on:keydown={onSearch}
        on:input={() => dispatch('search', { query: searchQuery })}
      />
      {#if searchQuery}
        <button class="clear-btn" on:click={() => { searchQuery = ''; dispatch('search', { query: '' }); }}>✕</button>
      {/if}
    </div>

    <!-- Command Palette Quick Trigger -->
    <button class="palette-quick-btn" on:click={() => dispatch('openPalette')} title="Abrir Command Palette (Ctrl+K)">
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
      <span class="cmd-k-kbd">⌘K</span>
    </button>

    <!-- Actions -->
    <div class="nav-actions">
      <!-- MAM Status Indicator -->
      {#if mamStatus && mamStatus.configured && mamStatus.account}
        <div class="mam-badge" title="Rácio MAM: {mamStatus.account.ratio}">
          <span class="ratio-label">🐭 {mamStatus.account.ratio}</span>
          {#if mamStatus.account.vip}
            <span class="vip-tag">VIP</span>
          {/if}
        </div>
      {/if}

      <!-- Offline Network Status Badge -->
      {#if !$isOnline}
        <div class="offline-pill" title="Sem ligação à internet. A biblioteca está a funcionar em modo offline.">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="1" y1="1" x2="23" y2="23"></line>
            <path d="M16.72 11.06A10.94 10.94 0 0 1 19 12.55"></path>
            <path d="M5 12.55a10.94 10.94 0 0 1 5.17-2.39"></path>
            <path d="M10.71 5.05A16 16 0 0 1 22.58 9"></path>
            <path d="M1.42 9a15.91 15.91 0 0 1 4.7-2.88"></path>
            <path d="M8.53 16.11a6 6 0 0 1 6.95 0"></path>
            <line x1="12" y1="20" x2="12.01" y2="20"></line>
          </svg>
          <span>Offline</span>
        </div>
      {/if}

      <!-- PWA Install Button -->
      {#if $canInstall}
        <button class="btn btn-secondary btn-sm pwa-install-btn" on:click={promptInstallPWA} title="Instalar como Aplicação Desktop / Telemóvel">
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          <span>Instalar App</span>
        </button>
      {/if}

      <!-- Theme Switcher -->
      <button class="btn btn-secondary btn-icon-only" on:click={toggleTheme} title="Alternar Tema">
        {#if theme === 'dark'}
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
        {:else}
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
        {/if}
      </button>

      <!-- Upload Button -->
      <button class="btn btn-primary" on:click={() => dispatch('openUpload')}>
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
        <span>Upload</span>
      </button>
    </div>
  </div>
</header>

<!-- Mobile Bottom Navigation Bar (< 768px) -->
<nav class="mobile-bottom-nav glass" aria-label="Navegação móvel">
  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'library'} 
    on:click={() => setTab('library')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
    <span>Livros</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'browse'} 
    on:click={() => setTab('browse')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
    <span>Explorar</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'shelves'} 
    on:click={() => setTab('shelves')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
    <span>Estantes</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'fts'} 
    on:click={() => setTab('fts')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
    <span>FTS</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'mam'} 
    on:click={() => setTab('mam')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
    <span>MAM</span>
  </button>

  <button 
    class="bottom-nav-item" 
    on:click={() => dispatch('openPalette')}>
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
    <span>Mais</span>
  </button>
</nav>

<style>
  .navbar {
    position: sticky;
    top: 0;
    left: 0;
    right: 0;
    height: 64px;
    z-index: 100;
    border-bottom: 1px solid var(--border-primary);
  }

  .nav-container {
    max-width: 1750px;
    margin: 0 auto;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    gap: 12px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    user-select: none;
    text-decoration: none;
    flex-shrink: 0;
  }

  .logo-icon {
    width: 36px;
    height: 36px;
    background: linear-gradient(135deg, #0284c7 0%, #38bdf8 100%);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 4px 12px rgba(56, 189, 248, 0.3);
  }

  .logo-icon svg {
    width: 20px;
    height: 20px;
  }

  .brand-text {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .brand-title {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--text-primary);
  }

  .nav-links {
    display: flex;
    align-items: center;
    gap: 4px;
    background: rgba(255, 255, 255, 0.03);
    padding: 3px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border-primary);
    overflow-x: auto;
  }

  .nav-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 11px;
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
    background: transparent;
    border: none;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
  }

  .nav-pill svg {
    width: 14px;
    height: 14px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .nav-pill:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .nav-pill.active {
    color: var(--accent);
    background: var(--accent-glow);
    font-weight: 600;
  }

  .search-wrap {
    flex: 1;
    max-width: 280px;
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-icon {
    position: absolute;
    left: 12px;
    width: 15px;
    height: 15px;
    fill: none;
    stroke: var(--text-muted);
    stroke-width: 2;
    pointer-events: none;
  }

  .search-wrap input {
    width: 100%;
    height: 36px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-full);
    padding: 0 32px 0 36px;
    color: var(--text-primary);
    font-size: 12px;
    outline: none;
    transition: all 0.15s ease;
  }

  .search-wrap input:focus {
    background: rgba(255, 255, 255, 0.08);
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow);
  }

  .clear-btn {
    position: absolute;
    right: 10px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 11px;
  }

  .nav-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .mam-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.3);
    padding: 4px 8px;
    border-radius: var(--radius-full);
    font-size: 11px;
    font-weight: 600;
    color: #f59e0b;
  }

  .vip-tag {
    background: #f59e0b;
    color: #000;
    font-size: 9px;
    padding: 1px 4px;
    border-radius: 4px;
    font-weight: 800;
  }

  .palette-quick-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-primary, rgba(255, 255, 255, 0.1));
    border-radius: var(--radius-full, 9999px);
    padding: 6px 10px;
    color: var(--text-secondary, #8b949e);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .palette-quick-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-primary, #ffffff);
    border-color: var(--accent, #6366f1);
  }

  .cmd-k-kbd {
    font-size: 10px;
    font-weight: 600;
    padding: 1px 4px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
  }

  /* Mobile Bottom Navigation */
  .mobile-bottom-nav {
    display: none;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 60px;
    background: var(--bg-card, rgba(22, 27, 34, 0.92));
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
    z-index: 9990;
    align-items: center;
    justify-content: space-around;
    padding: 0 8px;
    box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.3);
  }

  .bottom-nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    background: transparent;
    border: none;
    color: var(--text-secondary, #8b949e);
    cursor: pointer;
    padding: 6px 10px;
    border-radius: 8px;
    transition: all 0.15s ease;
    flex: 1;
  }

  .bottom-nav-item span {
    font-size: 10px;
    font-weight: 500;
  }

  .bottom-nav-item.active {
    color: var(--accent, #6366f1);
  }

  .bottom-nav-item:active {
    transform: scale(0.92);
  }

  .offline-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 10px;
    border-radius: 20px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.35);
    color: #f59e0b;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.02em;
    animation: pulse 2s infinite ease-in-out;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.65; }
  }

  .pwa-install-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: rgba(99, 102, 241, 0.15);
    border-color: rgba(99, 102, 241, 0.3);
    color: #a5b4fc;
    font-size: 11px;
    font-weight: 600;
    padding: 6px 12px;
    border-radius: 8px;
    transition: all 0.2s ease;
  }
  .pwa-install-btn:hover {
    background: rgba(99, 102, 241, 0.3);
    color: #ffffff;
    border-color: #6366f1;
  }

  @media (max-width: 1150px) {
    .nav-links {
      display: none;
    }
  }

  @media (max-width: 768px) {
    .mobile-bottom-nav {
      display: flex;
    }
    .palette-quick-btn .cmd-k-kbd {
      display: none;
    }
  }
</style>
