<script>
  // @ts-nocheck
  import { onMount, onDestroy } from 'svelte';
  import Navbar from './lib/Navbar.svelte';
  import BookGrid from './lib/BookGrid.svelte';
  import BookModal from './lib/BookModal.svelte';
  import BrowseView from './lib/BrowseView.svelte';
  import ShelvesView from './lib/ShelvesView.svelte';
  import FTSView from './lib/FTSView.svelte';
  import MAMView from './lib/MAMView.svelte';
  import DuplicatesView from './lib/DuplicatesView.svelte';
  import CWASettingsView from './lib/CWASettingsView.svelte';
  import PluginsView from './lib/PluginsView.svelte';
  import StatsView from './lib/StatsView.svelte';
  import Toast from './lib/Toast.svelte';
  import { toast } from './lib/toast.js';
  import CommandPalette from './lib/CommandPalette.svelte';

  let activeTab = 'library';
  let searchQuery = '';
  let theme = localStorage.getItem('cwa_theme') || 'dark';

  // Command Palette & SSE state
  let showCommandPalette = false;
  let sseSource = null;

  // Library query parameters
  let books = [];
  let totalBooks = 0;
  let currentPage = 1;
  let currentSort = 'new';
  let currentFilter = 'all'; // 'all', 'read', 'unread', 'hot', 'random'
  let currentShelf = '';
  let currentAuthorId = 0;
  let currentSeriesId = 0;
  let currentTagId = 0;
  let currentFormat = '';
  let currentFilterTitle = '';
  let loadingBooks = true;

  let selectedBookId = null;
  let mamStatus = null;
  let showUploadModal = false;
  let uploadFiles = null;
  let uploading = false;
  let uploadMsg = '';

  // Theme application
  $: {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('cwa_theme', theme);
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
  }

  onMount(async () => {
    await Promise.all([
      fetchBooks(),
      fetchMAMStatus()
    ]);
    initSSE();
  });

  onDestroy(() => {
    if (sseSource) {
      sseSource.close();
    }
  });

  function initSSE() {
    try {
      sseSource = new EventSource('/api/events');
      sseSource.addEventListener('book.ingested', (e) => {
        try {
          const data = JSON.parse(e.data);
          toast.success(data.message || 'Novo livro adicionado à biblioteca!');
          fetchBooks();
        } catch (_) {}
      });
      sseSource.addEventListener('book.cover_updated', (e) => {
        try {
          const data = JSON.parse(e.data);
          toast.info(data.message || 'Capa do livro atualizada!');
          fetchBooks();
        } catch (_) {}
      });
      sseSource.addEventListener('backup.created', (e) => {
        try {
          const data = JSON.parse(e.data);
          toast.success(data.message || 'Backup criado com sucesso!');
        } catch (_) {}
      });
      sseSource.addEventListener('convert.progress', (e) => {
        try {
          const data = JSON.parse(e.data);
          toast.info(data.message || 'Conversão atualizada!');
          fetchBooks();
        } catch (_) {}
      });
      sseSource.addEventListener('fixer.progress', (e) => {
        try {
          const data = JSON.parse(e.data);
          toast.info(data.message || 'Reparação de EPUB atualizada!');
          fetchBooks();
        } catch (_) {}
      });
    } catch (e) {
      console.warn('SSE broker connection failed:', e);
    }
  }

  async function fetchBooks() {
    loadingBooks = true;
    try {
      const url = new URL('/api/books', window.location.origin);
      url.searchParams.set('page', currentPage.toString());
      url.searchParams.set('limit', '30');
      url.searchParams.set('sort', currentSort);

      if (searchQuery.trim()) {
        url.searchParams.set('query', searchQuery.trim());
      }
      if (currentFilter === 'read') {
        url.searchParams.set('read_status', '1');
      } else if (currentFilter === 'unread') {
        url.searchParams.set('read_status', '2');
      } else if (currentFilter === 'hot') {
        url.searchParams.set('sort', 'hotdesc');
      } else if (currentFilter === 'random') {
        url.searchParams.set('sort', 'random');
      }

      if (currentShelf) {
        url.searchParams.set('shelf', currentShelf);
      }
      if (currentAuthorId > 0) {
        url.searchParams.set('author_id', currentAuthorId.toString());
      }
      if (currentSeriesId > 0) {
        url.searchParams.set('series_id', currentSeriesId.toString());
      }
      if (currentTagId > 0) {
        url.searchParams.set('tag_id', currentTagId.toString());
      }
      if (currentFormat) {
        url.searchParams.set('format', currentFormat);
      }

      const res = await fetch(url);
      const data = await res.json();
      books = data.items || [];
      totalBooks = data.total || 0;
    } catch (e) {
      console.error('Erro ao carregar livros:', e);
    } finally {
      loadingBooks = false;
    }
  }

  async function fetchMAMStatus() {
    try {
      const res = await fetch('/api/mam/status');
      mamStatus = await res.json();
    } catch (e) {
      console.error(e);
    }
  }

  function handleNavigate(event) {
    activeTab = event.detail.tab;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  function handleSearch(event) {
    searchQuery = event.detail.query;
    currentPage = 1;
    if (activeTab !== 'library') {
      activeTab = 'library';
    }
    fetchBooks();
  }

  function handleSortChange(event) {
    currentSort = event.detail.sort;
    currentPage = 1;
    fetchBooks();
  }

  function handleFilterChange(event) {
    currentFilter = event.detail.filter;
    currentPage = 1;
    fetchBooks();
  }

  function handlePageChange(event) {
    currentPage = event.detail.page;
    fetchBooks();
  }

  function handleSelectBook(event) {
    if (event.detail && event.detail.book) {
      selectedBookId = event.detail.book.id;
    }
  }

  function handleFilterSelect(event) {
    const { type, item } = event.detail;
    clearCustomFilters();
    if (type === 'author') {
      currentAuthorId = item.id;
      currentFilterTitle = `Autor: ${item.name}`;
    } else if (type === 'series') {
      currentSeriesId = item.id;
      currentFilterTitle = `Série: ${item.name}`;
    } else if (type === 'tag') {
      currentTagId = item.id;
      currentFilterTitle = `Categoria: ${item.name}`;
    } else if (type === 'format') {
      currentFormat = item.name || item.format;
      currentFilterTitle = `Formato: ${currentFormat}`;
    }
    activeTab = 'library';
    currentPage = 1;
    fetchBooks();
  }

  function handleShelfSelect(event) {
    const { id, name, isMagic } = event.detail;
    clearCustomFilters();
    if (isMagic) {
      currentShelf = `magic_${id}`;
      currentFilterTitle = `Estante Mágica: ${name}`;
    } else {
      currentShelf = id.toString();
      currentFilterTitle = `Estante: ${name}`;
    }
    activeTab = 'library';
    currentPage = 1;
    fetchBooks();
  }

  function clearCustomFilters() {
    currentShelf = '';
    currentAuthorId = 0;
    currentSeriesId = 0;
    currentTagId = 0;
    currentFormat = '';
    currentFilterTitle = '';
  }

  async function handleUploadSubmit(e) {
    e.preventDefault();
    if (!uploadFiles || uploadFiles.length === 0) return;
    uploading = true;
    uploadMsg = '';

    const formData = new FormData();
    formData.append('file', uploadFiles[0]);

    try {
      const res = await fetch('/upload', {
        method: 'POST',
        body: formData
      });
      if (res.ok) {
        uploadMsg = '✓ Livro enviado com sucesso!';
        setTimeout(() => {
          showUploadModal = false;
          uploadFiles = null;
          uploadMsg = '';
          fetchBooks();
        }, 1200);
      } else {
        uploadMsg = 'Erro ao enviar o ficheiro.';
      }
    } catch (err) {
      uploadMsg = 'Falha de comunicação no envio.';
    } finally {
      uploading = false;
    }
  }
</script>

<div class="app-root">
  <!-- Toast Notifications -->
  <Toast />

  <!-- Global Command Palette (Ctrl+K / Cmd+K) -->
  <CommandPalette 
    isOpen={showCommandPalette}
    on:close={() => { showCommandPalette = false; }}
    on:open={() => { showCommandPalette = true; }}
    on:navigate={handleNavigate}
    on:openUpload={() => { showUploadModal = true; }}
    on:toggleTheme={toggleTheme}
    on:refresh={fetchBooks}
    on:selectBook={handleSelectBook}
  />

  <!-- Top Navigation Bar -->
  <Navbar 
    {activeTab} 
    {searchQuery} 
    {mamStatus} 
    {theme}
    on:navigate={handleNavigate}
    on:search={handleSearch}
    on:toggleTheme={toggleTheme}
    on:openUpload={() => { showUploadModal = true; }}
    on:openPalette={() => { showCommandPalette = true; }}
  />

  <!-- Main View Container -->
  <main class="main-content">
    {#if activeTab === 'library'}
      <BookGrid 
        {books} 
        {totalBooks} 
        {currentPage} 
        {currentSort} 
        {currentFilter}
        {currentFilterTitle}
        loading={loadingBooks}
        on:sortChange={handleSortChange}
        on:filterChange={handleFilterChange}
        on:pageChange={handlePageChange}
        on:selectBook={handleSelectBook}
        on:clearCustomFilter={() => { clearCustomFilters(); fetchBooks(); }}
        on:refresh={fetchBooks}
      />
    {:else if activeTab === 'browse'}
      <BrowseView on:filterSelect={handleFilterSelect} />
    {:else if activeTab === 'shelves'}
      <ShelvesView on:shelfSelect={handleShelfSelect} />
    {:else if activeTab === 'fts'}
      <FTSView on:selectBook={handleSelectBook} />
    {:else if activeTab === 'mam'}
      <MAMView {mamStatus} />
    {:else if activeTab === 'duplicates'}
      <DuplicatesView />
    {:else if activeTab === 'cwa_settings'}
      <CWASettingsView />
    {:else if activeTab === 'plugins'}
      <PluginsView />
    {:else if activeTab === 'stats'}
      <StatsView />
    {/if}
  </main>

  <!-- Book Detail Modal -->
  {#if selectedBookId}
    <BookModal 
      bookId={selectedBookId} 
      on:close={() => { selectedBookId = null; }} 
      on:refresh={fetchBooks}
    />
  {/if}

  <!-- Upload Modal -->
  {#if showUploadModal}
    <div class="modal-overlay" on:click={() => { showUploadModal = false; }} on:keydown={(e) => { if (e.key === 'Escape') showUploadModal = false; }} role="presentation">
      <div class="modal-content upload-modal" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
        <h3>Adicionar Livro à Biblioteca</h3>
        <p class="upload-desc">Formatos suportados: .epub, .pdf, .mobi, .azw3, .cbr, .cbz. O ficheiro passará automaticamente pelo motor de ingestão, normalização de metadados e DeDRM.</p>

        <form on:submit={handleUploadSubmit} class="upload-form">
          <label class="upload-dropzone">
            <input 
              type="file" 
              accept=".epub,.pdf,.mobi,.azw3,.cbr,.cbz" 
              on:change={(e) => { uploadFiles = e.target.files; }} 
            />
            <div class="dropzone-inner">
              <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
              <span>{uploadFiles && uploadFiles.length ? uploadFiles[0].name : 'Clica para escolher ou arrasta um livro para aqui'}</span>
            </div>
          </label>

          {#if uploadMsg}
            <div class="upload-msg">{uploadMsg}</div>
          {/if}

          <div class="upload-actions">
            <button type="button" class="btn btn-secondary" on:click={() => { showUploadModal = false; }}>
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" disabled={!uploadFiles || uploading}>
              {uploading ? 'A enviar...' : 'Enviar para Ingest'}
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>

<style>
  .app-root {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background-color: var(--bg-primary);
  }

  .main-content {
    flex: 1;
    padding-bottom: 60px;
  }

  @media (max-width: 768px) {
    .main-content {
      padding-bottom: 90px;
    }
  }

  .upload-modal {
    max-width: 500px;
    padding: 28px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .upload-modal h3 {
    font-size: 18px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .upload-desc {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .upload-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 6px;
  }

  .upload-dropzone {
    border: 2px dashed var(--border-secondary);
    border-radius: var(--radius-md);
    padding: 32px 20px;
    cursor: pointer;
    background: rgba(255, 255, 255, 0.02);
    transition: all 0.2s ease;
  }

  .upload-dropzone:hover {
    border-color: var(--accent);
    background: var(--accent-glow);
  }

  .upload-dropzone input {
    display: none;
  }

  .dropzone-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    color: var(--text-secondary);
    font-size: 13px;
    text-align: center;
  }

  .upload-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }

  .upload-msg {
    font-size: 13px;
    color: var(--accent);
    text-align: center;
  }
</style>
