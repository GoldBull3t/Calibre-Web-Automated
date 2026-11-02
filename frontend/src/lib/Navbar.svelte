<script>
  // @ts-nocheck
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { isOnline, canInstall, promptInstallPWA } from './pwa.js';
  export let activeTab = 'library';
  export let searchQuery = '';
  export let mamStatus = null;
  export let theme = 'dark';
  export let currentUser = null;

  const dispatch = createEventDispatcher();

  let showMoreDropdown = false;
  let showMobileSheet = false;

  const MORE_TABS = ['fts', 'mam', 'duplicates', 'cwa_settings', 'plugins', 'stats'];
  $: isMoreActive = MORE_TABS.includes(activeTab);

  const TAB_NAMES = {
    fts: 'Pesquisa FTS',
    mam: 'MAM',
    duplicates: 'Duplicados',
    cwa_settings: 'Automação CWA',
    plugins: 'Plugins',
    stats: 'Sistema'
  };

  function onSearch(e) {
    if (e.key === 'Enter') {
      dispatch('search', { query: searchQuery });
    }
  }

  function toggleTheme() {
    dispatch('toggleTheme');
  }

  function setTab(tab) {
    showMoreDropdown = false;
    showMobileSheet = false;
    dispatch('navigate', { tab });
  }

  function toggleDropdown(e) {
    e.stopPropagation();
    showMoreDropdown = !showMoreDropdown;
  }

  function toggleSheet(e) {
    e.stopPropagation();
    showMobileSheet = !showMobileSheet;
  }

  function handleDocumentClick() {
    showMoreDropdown = false;
  }

  onMount(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('click', handleDocumentClick);
    }
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('click', handleDocumentClick);
    }
  });
</script>

<header class="navbar cwa-navbar">
  <div class="nav-container">
    <!-- Brand -->
    <a href="/#" class="brand" on:click|preventDefault={() => setTab('library')}>
      <span class="brand-title">Calibre-Web</span>
      <span class="cwa-badge-cwa">CWA</span>
    </a>

    <!-- Desktop Navigation: 4 Core Tabs + Tools Dropdown -->
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
        class:active={activeTab === 'sources'} 
        on:click={() => setTab('sources')}>
        <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
        <span>Fontes</span>
      </button>

      <!-- Compact Tools Dropdown -->
      <div class="cwa-dropdown">
        <button 
          class="cwa-nav-btn" 
          class:active={isMoreActive} 
          on:click={toggleDropdown}
          title="Ferramentas e Configurações">
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
          <span>{isMoreActive ? (TAB_NAMES[activeTab] || 'Mais') : 'Mais'}</span>
          <svg class="chevron-icon" viewBox="0 0 24 24"><polyline points="6 9 12 15 18 9"></polyline></svg>
        </button>

        {#if showMoreDropdown}
          <div class="cwa-dropdown-menu" on:click|stopPropagation on:keydown={() => {}} role="menu" tabindex="-1">
            <button 
              class="dropdown-item" 
              class:active={activeTab === 'fts'} 
              on:click={() => setTab('fts')}>
              <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
              <span>Pesquisa FTS</span>
            </button>

            <button 
              class="dropdown-item" 
              class:active={activeTab === 'mam'} 
              on:click={() => setTab('mam')}>
              <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"></path><path d="M2 12h20"></path></svg>
              <span>MAM Torrents</span>
            </button>

            <button 
              class="dropdown-item" 
              class:active={activeTab === 'duplicates'} 
              on:click={() => setTab('duplicates')}>
              <svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
              <span>Duplicados</span>
            </button>

            <button 
              class="dropdown-item" 
              class:active={activeTab === 'cwa_settings'} 
              on:click={() => setTab('cwa_settings')}>
              <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
              <span>Automação CWA</span>
            </button>

            <button 
              class="dropdown-item" 
              class:active={activeTab === 'plugins'} 
              on:click={() => setTab('plugins')}>
              <svg viewBox="0 0 24 24"><path d="M12 2v4m0 12v4M2 12h4m12 0h4m-3.17-6.83l-2.83 2.83m-8 8l-2.83 2.83m0-13.66l2.83 2.83m8 8l2.83 2.83"></path></svg>
              <span>Plugins</span>
            </button>

            <button 
              class="dropdown-item" 
              class:active={activeTab === 'stats'} 
              on:click={() => setTab('stats')}>
              <svg viewBox="0 0 24 24"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
              <span>Sistema</span>
            </button>
          </div>
        {/if}
      </div>
    </nav>

    <!-- Compact Search Bar -->
    <div class="search-wrap">
      <svg class="search-icon" viewBox="0 0 24 24">
        <circle cx="11" cy="11" r="8"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <input 
        type="text" 
        placeholder="Pesquisar..." 
        bind:value={searchQuery} 
        on:keydown={onSearch}
        on:input={() => dispatch('search', { query: searchQuery })}
      />
      {#if searchQuery}
        <button class="clear-btn" on:click={() => { searchQuery = ''; dispatch('search', { query: '' }); }}>✕</button>
      {/if}
    </div>

    <!-- Right Header Actions (Compact, Zero Overflow) -->
    <div class="nav-actions">
      <!-- MAM Status Indicator (Compact) -->
      {#if mamStatus && mamStatus.configured && mamStatus.account}
        <div class="cwa-mam-badge hidden-xs" title="Rácio MAM: {mamStatus.account.ratio}">
          <span>🐭 {mamStatus.account.ratio}</span>
        </div>
      {/if}

      <!-- Offline Indicator -->
      {#if !$isOnline}
        <div class="cwa-offline-badge" title="Modo offline">
          <span>Offline</span>
        </div>
      {/if}

      <!-- User Profile Link to /me -->
      {#if currentUser && currentUser.authenticated}
        <a href="/me" class="cwa-btn-icon user-btn" title="Perfil ({currentUser.name})">
          <svg viewBox="0 0 24 24" width="15" height="15"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
          <span class="user-name hidden-sm">{currentUser.name}</span>
        </a>
        <a href="/logout" class="cwa-btn-icon hidden-xs" title="Terminar Sessão">
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
        </a>
      {:else}
        <a href="/login" class="cwa-btn-icon" title="Entrar">
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path><polyline points="10 17 15 12 10 7"></polyline><line x1="15" y1="12" x2="3" y2="12"></line></svg>
        </a>
      {/if}

      <!-- Theme Switcher -->
      <button class="cwa-btn-icon" on:click={toggleTheme} title="Alternar Tema">
        {#if theme === 'dark'}
          <svg viewBox="0 0 24 24" width="14" height="14"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
        {:else}
          <svg viewBox="0 0 24 24" width="14" height="14"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
        {/if}
      </button>

      <!-- Upload Button -->
      <button class="cwa-upload-btn" on:click={() => dispatch('openUpload')} title="Adicionar Livro">
        <svg viewBox="0 0 24 24" width="13" height="13"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        <span class="hidden-xs">Upload</span>
      </button>
    </div>
  </div>
</header>

<!-- Mobile Bottom Sheet for "Mais" options -->
{#if showMobileSheet}
  <div class="mobile-sheet-backdrop" on:click={() => { showMobileSheet = false; }} on:keydown={(e) => { if (e.key === 'Escape') showMobileSheet = false; }} role="presentation">
    <div class="mobile-sheet" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
      <div class="mobile-sheet-header">
        <span class="mobile-sheet-title">Ferramentas & Operações</span>
        <button class="sheet-close-btn" on:click={() => { showMobileSheet = false; }}>✕</button>
      </div>

      <div class="mobile-sheet-grid">
        <button 
          class="sheet-btn" 
          class:active={activeTab === 'fts'} 
          on:click={() => setTab('fts')}>
          <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
          <span>FTS</span>
        </button>

        <button 
          class="sheet-btn" 
          class:active={activeTab === 'mam'} 
          on:click={() => setTab('mam')}>
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"></path><path d="M2 12h20"></path></svg>
          <span>MAM</span>
        </button>

        <button 
          class="sheet-btn" 
          class:active={activeTab === 'duplicates'} 
          on:click={() => setTab('duplicates')}>
          <svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
          <span>Duplicados</span>
        </button>

        <button 
          class="sheet-btn" 
          class:active={activeTab === 'cwa_settings'} 
          on:click={() => setTab('cwa_settings')}>
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
          <span>Automação</span>
        </button>

        <button 
          class="sheet-btn" 
          class:active={activeTab === 'plugins'} 
          on:click={() => setTab('plugins')}>
          <svg viewBox="0 0 24 24"><path d="M12 2v4m0 12v4M2 12h4m12 0h4m-3.17-6.83l-2.83 2.83m-8 8l-2.83 2.83m0-13.66l2.83 2.83m8 8l2.83 2.83"></path></svg>
          <span>Plugins</span>
        </button>

        <button 
          class="sheet-btn" 
          class:active={activeTab === 'stats'} 
          on:click={() => setTab('stats')}>
          <svg viewBox="0 0 24 24"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>
          <span>Sistema</span>
        </button>
      </div>

      {#if $canInstall}
        <button class="sheet-action-btn" on:click={promptInstallPWA}>
          <svg viewBox="0 0 24 24" width="15" height="15"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          <span>Instalar Aplicação PWA</span>
        </button>
      {/if}
    </div>
  </div>
{/if}

<!-- Mobile Bottom Navigation Bar (Compact 50px) -->
<nav class="mobile-bottom-nav" aria-label="Navegação móvel">
  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'library'} 
    on:click={() => setTab('library')}>
    <svg viewBox="0 0 24 24"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg>
    <span>Livros</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'browse'} 
    on:click={() => setTab('browse')}>
    <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
    <span>Explorar</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'shelves'} 
    on:click={() => setTab('shelves')}>
    <svg viewBox="0 0 24 24"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></svg>
    <span>Estantes</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={activeTab === 'sources'} 
    on:click={() => setTab('sources')}>
    <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
    <span>Fontes</span>
  </button>

  <button 
    class="bottom-nav-item" 
    class:active={isMoreActive || showMobileSheet} 
    on:click={toggleSheet}>
    <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="1"></circle><circle cx="12" cy="5" r="1"></circle><circle cx="12" cy="19" r="1"></circle></svg>
    <span>Mais</span>
  </button>
</nav>

<style>
  /* Slim & Refined Header (Height: 50px) */
  .cwa-navbar {
    position: sticky;
    top: 0;
    left: 0;
    right: 0;
    height: 50px;
    min-height: 50px;
    background: rgba(11, 15, 25, 0.96);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
    padding: 0 16px;
    margin: 0;
    z-index: 1000;
  }

  .nav-container {
    width: 100%;
    max-width: 1750px;
    margin: 0 auto;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  /* Compact Brand */
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #ffffff;
    font-size: 14px;
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
    font-size: 9px;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 5px;
    letter-spacing: 0.5px;
    border: 1px solid rgba(56, 189, 248, 0.4);
    display: inline-block;
    line-height: normal;
  }

  /* Desktop Nav Links */
  .nav-links {
    display: flex;
    align-items: center;
    gap: 3px;
    flex-shrink: 0;
  }

  .cwa-nav-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    padding: 0 10px;
    border-radius: 6px;
    background: transparent;
    border: 1px solid transparent;
    color: #94a3b8;
    font-size: 12.5px;
    font-weight: 500;
    text-decoration: none;
    transition: all 0.15s ease;
    gap: 6px;
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

  .cwa-nav-btn .chevron-icon {
    width: 12px;
    height: 12px;
    opacity: 0.6;
    margin-left: -2px;
  }

  .cwa-nav-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
  }

  .cwa-nav-btn.active {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.35);
    color: #38bdf8;
    font-weight: 600;
  }

  /* Dropdown Menu */
  .cwa-dropdown {
    position: relative;
    display: inline-flex;
  }

  .cwa-dropdown-menu {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    min-width: 180px;
    background: #121926;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    padding: 5px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    z-index: 1200;
    animation: fadeIn 0.15s ease-out;
  }

  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px;
    border-radius: 5px;
    border: none;
    background: transparent;
    color: #94a3b8;
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    text-align: left;
    width: 100%;
    transition: all 0.15s ease;
    font-family: inherit;
  }

  .dropdown-item svg {
    width: 14px;
    height: 14px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .dropdown-item:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
  }

  .dropdown-item.active {
    background: rgba(59, 130, 246, 0.15);
    color: #38bdf8;
    font-weight: 600;
  }

  /* Compact Search Bar */
  .search-wrap {
    position: relative;
    display: flex;
    align-items: center;
    width: 180px;
    flex-shrink: 1;
    transition: width 0.2s ease;
  }
  .search-wrap:focus-within {
    width: 220px;
  }

  .search-wrap input {
    width: 100%;
    height: 32px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 6px;
    padding: 4px 24px 4px 30px;
    color: #ffffff;
    font-size: 12.5px;
    outline: none;
    transition: all 0.15s ease;
    font-family: inherit;
  }

  .search-wrap input::placeholder {
    color: #64748b;
    font-size: 12px;
  }

  .search-wrap input:focus {
    background: rgba(255, 255, 255, 0.08);
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  .search-icon {
    position: absolute;
    left: 8px;
    width: 13px;
    height: 13px;
    fill: none;
    stroke: #94a3b8;
    stroke-width: 2;
    pointer-events: none;
  }

  .clear-btn {
    position: absolute;
    right: 6px;
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
    font-size: 11px;
  }
  .clear-btn:hover {
    color: #ffffff;
  }

  /* Right Header Actions */
  .nav-actions {
    display: flex;
    align-items: center;
    gap: 5px;
    flex-shrink: 0;
  }

  .cwa-btn-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    width: 32px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #94a3b8;
    cursor: pointer;
    text-decoration: none;
    transition: all 0.15s ease;
    font-family: inherit;
  }
  .cwa-btn-icon:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
  }
  .cwa-btn-icon svg {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .user-btn {
    width: auto;
    padding: 0 8px;
    gap: 5px;
  }
  .user-name {
    font-size: 12px;
    font-weight: 500;
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Compact Upload Button */
  .cwa-upload-btn {
    color: #ffffff;
    background: linear-gradient(135deg, #2563eb, #1d4ed8);
    border: 1px solid rgba(59, 130, 246, 0.5);
    border-radius: 6px;
    font-weight: 600;
    font-size: 12px;
    padding: 0 12px;
    height: 32px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    box-shadow: 0 1px 6px rgba(37, 99, 235, 0.3);
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
    font-family: inherit;
  }
  .cwa-upload-btn:hover {
    background: linear-gradient(135deg, #3b82f6, #2563eb);
    box-shadow: 0 2px 10px rgba(37, 99, 235, 0.45);
    transform: translateY(-1px);
  }
  .cwa-upload-btn svg {
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
  }

  .cwa-mam-badge {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0 8px;
    border-radius: 6px;
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.25);
    color: #f59e0b;
    font-size: 11px;
    font-weight: 600;
  }

  .cwa-offline-badge {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0 8px;
    border-radius: 6px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    font-size: 11px;
    font-weight: 600;
  }

  /* Responsive Rules */
  @media (max-width: 1024px) {
    .hidden-sm {
      display: none !important;
    }
    .user-btn {
      width: 32px;
      padding: 0;
    }
  }

  @media (max-width: 768px) {
    .nav-links {
      display: none;
    }
    .hidden-xs {
      display: none !important;
    }
    .search-wrap {
      width: 130px;
    }
    .search-wrap:focus-within {
      width: 160px;
    }
    .cwa-upload-btn {
      padding: 0 8px;
      width: 32px;
    }
  }

  @media (max-width: 440px) {
    .brand-title {
      display: none;
    }
    .search-wrap {
      width: 110px;
    }
    .search-wrap:focus-within {
      width: 140px;
    }
  }

  /* Mobile Bottom Navigation Bar (Height: 50px) */
  .mobile-bottom-nav {
    display: none;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 50px;
    background: rgba(11, 15, 25, 0.98);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    z-index: 9990;
    align-items: center;
    justify-content: space-around;
    padding: 0 4px calc(env(safe-area-inset-bottom, 0px));
    box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.3);
  }

  .bottom-nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 4px 6px;
    border-radius: 6px;
    transition: all 0.15s ease;
    flex: 1;
    font-family: inherit;
  }

  .bottom-nav-item svg {
    width: 16px;
    height: 16px;
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

  @media (max-width: 768px) {
    .mobile-bottom-nav {
      display: flex;
    }
  }

  /* Mobile Sheet Modal */
  .mobile-sheet-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    z-index: 9995;
    display: flex;
    align-items: flex-end;
    animation: fadeIn 0.15s ease-out;
  }

  .mobile-sheet {
    width: 100%;
    background: #121926;
    border-top: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 16px 16px 0 0;
    padding: 16px 16px calc(65px + env(safe-area-inset-bottom, 0px)) 16px;
    box-shadow: 0 -8px 30px rgba(0, 0, 0, 0.6);
    display: flex;
    flex-direction: column;
    gap: 8px;
    animation: slideUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .mobile-sheet-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .mobile-sheet-title {
    font-size: 12px;
    font-weight: 700;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .sheet-close-btn {
    background: transparent;
    border: none;
    color: #64748b;
    font-size: 14px;
    cursor: pointer;
    padding: 4px;
  }

  .mobile-sheet-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    padding: 6px 0;
  }

  .sheet-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 12px 6px;
    color: #94a3b8;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .sheet-btn svg {
    width: 20px;
    height: 20px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
  }

  .sheet-btn span {
    font-size: 11px;
    font-weight: 500;
  }

  .sheet-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
  }

  .sheet-btn.active {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.4);
    color: #38bdf8;
    font-weight: 600;
  }

  .sheet-action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: rgba(59, 130, 246, 0.15);
    border: 1px solid rgba(59, 130, 246, 0.3);
    border-radius: 8px;
    padding: 10px;
    color: #38bdf8;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    font-family: inherit;
    margin-top: 4px;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes slideUp {
    from { transform: translateY(40px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>
