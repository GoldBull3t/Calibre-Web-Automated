<script>
  // @ts-nocheck
  import { createEventDispatcher } from 'svelte';
  import { isOnline, canInstall, promptInstallPWA } from './pwa.js';
  export let activeTab = 'library';
  export let searchQuery = '';
  export let mamStatus = null;
  export let theme = 'dark';
  export let currentUser = null;

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

<header class="navbar cwa-navbar">
  <div class="nav-container">
    <!-- Brand matching CWA 2026 Design System -->
    <a href="/#" class="brand" on:click|preventDefault={() => setTab('library')}>
      <span class="brand-title">Calibre-Web Automated</span>
      <span class="cwa-badge-cwa">CWA</span>
    </a>

    <!-- Navigation Items -->
    <nav class="nav-links">
      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'library'} 
        on:click={() => setTab('library')}>
        <svg viewBox="0 0 24 24"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
        <span>Biblioteca</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'browse'} 
        on:click={() => setTab('browse')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
        <span>Explorar</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'shelves'} 
        on:click={() => setTab('shelves')}>
        <svg viewBox="0 0 24 24"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
        <span>Estantes</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'fts'} 
        on:click={() => setTab('fts')}>
        <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <span>Pesquisa FTS</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'sources'} 
        on:click={() => setTab('sources')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
        <span>Fontes</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'mam'} 
        on:click={() => setTab('mam')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"></path><path d="M2 12h20"></path></svg>
        <span>MAM</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'duplicates'} 
        on:click={() => setTab('duplicates')}>
        <svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
        <span>Duplicados</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'cwa_settings'} 
        on:click={() => setTab('cwa_settings')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
        <span>Automação CWA</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'plugins'} 
        on:click={() => setTab('plugins')}>
        <svg viewBox="0 0 24 24"><path d="M12 2v4m0 12v4M2 12h4m12 0h4m-3.17-6.83l-2.83 2.83m-8 8l-2.83 2.83m0-13.66l2.83 2.83m8 8l2.83 2.83"></path></svg>
        <span>Plugins</span>
      </button>

      <button 
        class="cwa-nav-btn" 
        class:active={activeTab === 'stats'} 
        on:click={() => setTab('stats')}>
        <svg viewBox="0 0 24 24"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
        <span>Sistema</span>
      </button>
    </nav>

    <!-- Search Input matching cwa.css -->
    <div class="search-wrap">
      <svg class="search-icon" viewBox="0 0 24 24">
        <circle cx="11" cy="11" r="8"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <input 
        type="text" 
        placeholder="Pesquisar Biblioteca..." 
        bind:value={searchQuery} 
        on:keydown={onSearch}
        on:input={() => dispatch('search', { query: searchQuery })}
      />
      {#if searchQuery}
        <button class="clear-btn" on:click={() => { searchQuery = ''; dispatch('search', { query: '' }); }}>✕</button>
      {/if}
    </div>

    <!-- Quick Command Palette Button -->
    <button class="cwa-nav-btn cwa-cmd-btn" on:click={() => dispatch('openPalette')} title="Command Palette (Ctrl+K)">
      <svg viewBox="0 0 24 24" width="13" height="13"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
      <span class="cmd-k-kbd">⌘K</span>
    </button>

    <!-- Right Actions -->
    <div class="nav-actions">
      {#if mamStatus && mamStatus.configured && mamStatus.account}
        <div class="cwa-mam-badge" title="Rácio MAM: {mamStatus.account.ratio}">
          <span>🐭 {mamStatus.account.ratio}</span>
        </div>
      {/if}

      {#if !$isOnline}
        <div class="cwa-offline-badge" title="Modo offline ativo">
          <span>Offline</span>
        </div>
      {/if}

      {#if $canInstall}
        <button class="cwa-nav-btn" on:click={promptInstallPWA} title="Instalar Aplicação">
          <svg viewBox="0 0 24 24" width="13" height="13"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          <span class="hidden-sm">Instalar</span>
        </button>
      {/if}

      <!-- User Profile Link to /me -->
      {#if currentUser && currentUser.authenticated}
        <a href="/me" class="cwa-nav-btn" title="Definições da Conta ({currentUser.name})">
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
          <span class="user-name hidden-sm">{currentUser.name}</span>
        </a>
        <a href="/logout" class="cwa-nav-btn cwa-nav-icon-btn" title="Terminar Sessão">
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
        </a>
      {:else}
        <a href="/login" class="cwa-nav-btn" title="Iniciar Sessão">
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path><polyline points="10 17 15 12 10 7"></polyline><line x1="15" y1="12" x2="3" y2="12"></line></svg>
          <span>Entrar</span>
        </a>
      {/if}

      <!-- Theme Switcher -->
      <button class="cwa-nav-btn cwa-nav-icon-btn" on:click={toggleTheme} title="Alternar Tema">
        {#if theme === 'dark'}
          <svg viewBox="0 0 24 24" width="14" height="14"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
        {:else}
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
        {/if}
      </button>

      <!-- Upload Button matching cwa.css lines 333-372 -->
      <button class="cwa-upload-btn" on:click={() => dispatch('openUpload')}>
        <svg viewBox="0 0 24 24" width="14" height="14"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
        <span>Upload</span>
      </button>
    </div>
  </div>
</header>

<!-- Mobile Bottom Navigation Bar (< 860px) -->
<nav class="mobile-bottom-nav" aria-label="Navegação móvel">
  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'library'} 
    on:click={() => setTab('library')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
    <span>Livros</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'browse'} 
    on:click={() => setTab('browse')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
    <span>Explorar</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'shelves'} 
    on:click={() => setTab('shelves')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
    <span>Estantes</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'sources'} 
    on:click={() => setTab('sources')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
    <span>Fontes</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'fts'} 
    on:click={() => setTab('fts')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
    <span>FTS</span>
  </button>

  <button 
    class="bottom-nav-item" 
    on:click={() => dispatch('openPalette')}>
    <svg viewBox="0 0 24 24" width="18" height="18"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
    <span>Mais</span>
  </button>
</nav>

<style>
  .cwa-navbar {
    position: sticky;
    top: 0;
    left: 0;
    right: 0;
    height: 60px;
    min-height: 60px;
    background: rgba(11, 15, 25, 0.95);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.35);
    padding: 0 20px;
    margin: 0;
    z-index: 1000;
  }

  .nav-container {
    width: 100%;
    max-width: 1800px;
    margin: 0 auto;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .brand {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: #ffffff;
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.2px;
    text-decoration: none;
    cursor: pointer;
    white-space: nowrap;
    flex-shrink: 0;
  }
  .brand:hover {
    color: #38bdf8;
  }

  .cwa-badge-cwa {
    background: linear-gradient(135deg, #0284c7, #0369a1);
    color: #ffffff;
    font-size: 10px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 6px;
    letter-spacing: 0.5px;
    border: 1px solid rgba(56, 189, 248, 0.4);
    vertical-align: middle;
    display: inline-block;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
    margin-left: 6px;
  }

  .nav-links {
    display: flex;
    align-items: center;
    gap: 5px;
    overflow-x: auto;
    padding: 2px 0;
  }

  /* Nav Buttons matching cwa.css lines 421-460 */
  .cwa-nav-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 36px;
    width: auto;
    min-width: 36px;
    padding: 0 12px;
    margin: 0;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    font-size: 13px;
    font-weight: 500;
    text-decoration: none;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    gap: 7px;
    cursor: pointer;
    white-space: nowrap;
    font-family: inherit;
    box-sizing: border-box;
  }

  .cwa-nav-btn svg {
    width: 14px;
    height: 14px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    flex-shrink: 0;
  }

  .cwa-nav-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.2);
    color: #ffffff;
    transform: translateY(-1px);
    box-shadow: 0 3px 10px rgba(0, 0, 0, 0.25);
  }

  .cwa-nav-btn.active {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.4);
    color: #38bdf8;
    font-weight: 600;
  }

  .cwa-nav-icon-btn {
    padding: 0;
    width: 36px;
  }

  /* Search Input matching cwa.css lines 154-227 */
  .search-wrap {
    position: relative;
    display: flex;
    align-items: center;
    width: 220px;
    flex-shrink: 1;
    transition: width 0.2s ease;
  }
  .search-wrap:focus-within {
    width: 270px;
  }

  .search-wrap input {
    width: 100%;
    height: 34px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 6px 28px 6px 34px;
    color: #ffffff;
    font-size: 13px;
    outline: none;
    transition: all 0.2s ease;
    font-family: inherit;
  }

  .search-wrap input::placeholder {
    color: #64748b;
    font-size: 13px;
  }

  .search-wrap input:focus {
    background: rgba(255, 255, 255, 0.08);
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.25);
  }

  .search-icon {
    position: absolute;
    left: 10px;
    width: 14px;
    height: 14px;
    fill: none;
    stroke: #94a3b8;
    stroke-width: 2;
    pointer-events: none;
  }

  .clear-btn {
    position: absolute;
    right: 8px;
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
    font-size: 11px;
  }
  .clear-btn:hover {
    color: #ffffff;
  }

  .cwa-cmd-btn {
    padding: 0 8px;
    gap: 4px;
  }
  .cmd-k-kbd {
    font-size: 10px;
    font-weight: 600;
    padding: 1px 4px;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #94a3b8;
  }

  .nav-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  /* Upload button matching cwa.css lines 333-372 */
  .cwa-upload-btn {
    color: #ffffff;
    background: linear-gradient(135deg, #2563eb, #1d4ed8);
    border: 1px solid rgba(59, 130, 246, 0.5);
    border-radius: 8px;
    font-weight: 600;
    font-size: 13px;
    padding: 7px 16px;
    height: 36px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    box-shadow: 0 2px 10px rgba(37, 99, 235, 0.3);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    white-space: nowrap;
    font-family: inherit;
    text-transform: none;
  }
  .cwa-upload-btn:hover {
    background: linear-gradient(135deg, #3b82f6, #2563eb);
    box-shadow: 0 4px 14px rgba(37, 99, 235, 0.45);
    transform: translateY(-1px);
    color: #ffffff;
  }
  .cwa-upload-btn svg {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .cwa-mam-badge {
    display: inline-flex;
    align-items: center;
    height: 36px;
    padding: 0 10px;
    border-radius: 8px;
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    font-size: 12px;
    font-weight: 600;
  }

  .cwa-offline-badge {
    display: inline-flex;
    align-items: center;
    height: 36px;
    padding: 0 10px;
    border-radius: 8px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.35);
    color: #f59e0b;
    font-size: 12px;
    font-weight: 600;
  }

  .user-name {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Responsive Rules */
  @media (max-width: 1240px) {
    .hidden-sm {
      display: none !important;
    }
    .search-wrap {
      width: 160px;
    }
    .search-wrap:focus-within {
      width: 200px;
    }
  }

  @media (max-width: 900px) {
    .nav-links {
      display: none;
    }
    .search-wrap {
      width: 140px;
    }
  }

  /* Mobile Bottom Navigation */
  .mobile-bottom-nav {
    display: none;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 56px;
    background: rgba(11, 15, 25, 0.98);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-top: 1px solid rgba(255, 255, 255, 0.08);
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
    color: #94a3b8;
    cursor: pointer;
    padding: 6px 10px;
    border-radius: 8px;
    transition: all 0.15s ease;
    flex: 1;
    font-family: inherit;
  }

  .bottom-nav-item svg {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .bottom-nav-item span {
    font-size: 10px;
    font-weight: 500;
  }

  .bottom-nav-item.active {
    color: #38bdf8;
  }

  .bottom-nav-item:active {
    transform: scale(0.92);
  }

  @media (max-width: 900px) {
    .mobile-bottom-nav {
      display: flex;
    }
    .cwa-cmd-btn .cmd-k-kbd {
      display: none;
    }
  }
</style>
