<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { toast } from './toast.js';
  import { offlineBookIds, saveBookOffline, removeBookOffline } from './pwa.js';
  export let bookId;

  const dispatch = createEventDispatcher();
  let book = null;
  let loading = true;
  let error = null;

  // Cover Drag & Drop state
  let isDraggingCover = false;
  let coverTimestamp = Date.now();

  // Edit Mode state
  let isEditing = false;
  let editTitle = '';
  let editAuthors = '';
  let editSeries = '';
  let editSeriesIndex = 1.0;
  let editTags = '';
  let editRating = 0;
  let editComment = '';
  let saving = false;
  let editMsg = '';

  // Send to eReader modal state
  let showSendModal = false;
  let sendEmail = '';
  let sendFormat = 'EPUB';
  let sending = false;
  let sendMsg = '';

  // Auto-Enrichment Studio state
  let showEnrichStudio = false;
  let enrichQuery = '';
  let searchingEnrich = false;
  let candidates = [];
  let selectedCandidateIndex = 0;
  let selectedCoverURL = '';
  let applyingEnrich = false;
  let applyCover = true;
  let applyDescription = true;
  let applyTags = true;
  let applyTitleAuthor = true;

  onMount(loadBook);

  async function loadBook() {
    loading = true;
    try {
      const res = await fetch(`/api/books/${bookId}`);
      if (!res.ok) throw new Error('Falha ao carregar detalhes do livro.');
      book = await res.json();
      populateEditForm();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  function populateEditForm() {
    if (!book) return;
    editTitle = book.title || '';
    editAuthors = book.authors ? book.authors.join(', ') : '';
    editSeries = book.series || '';
    editSeriesIndex = book.series_index || 1.0;
    editTags = book.tags ? book.tags.join(', ') : '';
    editRating = book.rating || 0;
    editComment = book.comment || '';
  }

  async function uploadCoverFile(file) {
    if (!file || !file.type.startsWith('image/')) {
      toast.error('Por favor seleciona uma imagem válida (.jpg, .png, .webp).');
      return;
    }
    const formData = new FormData();
    formData.append('cover', file);
    try {
      const res = await fetch(`/api/books/${book.id}/cover`, {
        method: 'POST',
        body: formData
      });
      if (res.ok) {
        coverTimestamp = Date.now();
        toast.success('Capa do livro atualizada com sucesso!');
        dispatch('refresh');
      } else {
        const errData = await res.json().catch(() => ({}));
        toast.error(errData.error || 'Falha ao atualizar capa.');
      }
    } catch (err) {
      toast.error('Erro de rede ao enviar capa.');
    }
  }

  function close() {
    dispatch('close');
  }

  function onRead() {
    window.location.href = `/read/${bookId}`;
  }

  function onDownload(format) {
    window.location.href = `/download/${bookId}/${format.toLowerCase()}`;
  }

  async function toggleRead() {
    if (!book) return;
    try {
      await fetch(`/ajax/toggleread/${bookId}`, { method: 'POST' });
      book.read_status = !book.read_status;
    } catch (e) {
      console.error(e);
    }
  }

  async function saveMetadata() {
    saving = true;
    editMsg = '';
    const payload = {
      title: editTitle.trim(),
      authors: editAuthors.split(',').map(s => s.trim()).filter(Boolean),
      series: editSeries.trim(),
      series_index: parseFloat(editSeriesIndex) || 1.0,
      tags: editTags.split(',').map(s => s.trim()).filter(Boolean),
      rating: parseInt(editRating, 10) || 0,
      comment: editComment
    };

    try {
      const res = await fetch(`/api/books/${bookId}/edit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        editMsg = '✓ Metadados atualizados com sucesso!';
        setTimeout(async () => {
          isEditing = false;
          await loadBook();
          dispatch('refresh');
        }, 800);
      } else {
        editMsg = 'Erro ao guardar metadados.';
      }
    } catch (e) {
      editMsg = 'Erro de comunicação.';
    } finally {
      saving = false;
    }
  }

  async function sendToEReader() {
    if (!sendEmail.trim()) return;
    sending = true;
    sendMsg = '';
    try {
      const res = await fetch(`/send/${bookId}/${sendFormat.toLowerCase()}/0`, {
        method: 'POST'
      });
      if (res.ok) {
        sendMsg = '✓ Livro enviado com sucesso para o teu e-Reader!';
        setTimeout(() => { showSendModal = false; sendMsg = ''; }, 1500);
      } else {
        sendMsg = 'Erro ao enviar por email.';
      }
    } catch (e) {
      sendMsg = 'Falha no envio.';
    } finally {
      sending = false;
    }
  }

  async function runFixerOnBook() {
    try {
      const res = await fetch(`/api/fixer/start?book_id=${bookId}`, { method: 'POST' });
      const data = await res.json();
      toast.success(data.message || 'Reparação de EPUB concluída!');
    } catch (e) {
      toast.error('Erro ao reparar livro.');
    }
  }

  async function deleteBook() {
    if (!confirm(`Tens a certeza que desejas eliminar "${book.title}" da biblioteca? Ficheiros físicos e metadados serão removidos.`)) {
      return;
    }
    try {
      const res = await fetch(`/delete/${bookId}`, { method: 'POST' });
      if (res.ok) {
        close();
        dispatch('refresh');
      }
    } catch (e) {
      console.error(e);
    }
  }

  // --- Offline PWA Storage Functions ---
  $: isSavedOffline = Boolean($offlineBookIds && $offlineBookIds.has(parseInt(bookId, 10)));
  let savingOffline = false;

  async function toggleOfflineStatus() {
    savingOffline = true;
    try {
      if (isSavedOffline) {
        await removeBookOffline(bookId);
        toast.info('Livro removido do armazenamento offline.');
      } else {
        await saveBookOffline(bookId);
        toast.success('💾 Livro guardado para leitura offline no teu dispositivo!');
      }
    } catch (e) {
      toast.error('Erro ao atualizar armazenamento offline.');
    } finally {
      savingOffline = false;
    }
  }

  // --- Auto-Enrichment Studio Functions ---
  function openEnrichStudio() {
    showEnrichStudio = true;
    const author = (book.authors && book.authors.length > 0) ? book.authors[0] : '';
    enrichQuery = `${book.title} ${author}`.trim();
    searchEnrichment();
  }

  async function searchEnrichment() {
    if (!enrichQuery.trim()) return;
    searchingEnrich = true;
    candidates = [];
    selectedCandidateIndex = 0;
    selectedCoverURL = '';
    try {
      const res = await fetch(`/api/books/${bookId}/enrich/search?query=${encodeURIComponent(enrichQuery)}`);
      if (res.ok) {
        const data = await res.json();
        candidates = data.candidates || [];
        if (candidates.length > 0) {
          const first = candidates[0];
          if (first.covers && first.covers.length > 0) {
            selectedCoverURL = first.covers[0].url;
          }
        }
      } else {
        toast.error('Erro ao pesquisar metadados online.');
      }
    } catch (e) {
      toast.error('Erro de conexão ao pesquisar.');
    } finally {
      searchingEnrich = false;
    }
  }

  async function quickOneClickEnrich() {
    applyingEnrich = true;
    try {
      const res = await fetch(`/api/books/${bookId}/enrich/auto`, { method: 'POST' });
      if (res.ok) {
        toast.success('✨ Livro enriquecido com sucesso com Capa HD e Metadados!');
        showEnrichStudio = false;
        coverTimestamp = Date.now();
        await loadBook();
        dispatch('refresh');
      } else {
        const errData = await res.json().catch(() => ({}));
        toast.error(errData.error || 'Falha no enriquecimento automático.');
      }
    } catch (e) {
      toast.error('Erro de rede ao enriquecer livro.');
    } finally {
      applyingEnrich = false;
    }
  }

  async function applySelectedCandidate() {
    if (!candidates || candidates.length === 0) return;
    const cand = candidates[selectedCandidateIndex];
    if (!cand) return;

    applyingEnrich = true;
    const payload = {
      cover_url: applyCover ? selectedCoverURL : '',
      title: applyTitleAuthor ? cand.title : book.title,
      authors: applyTitleAuthor ? cand.authors : book.authors,
      description: applyDescription ? cand.description : book.comment,
      tags: applyTags ? cand.tags : book.tags,
      publisher: cand.publisher,
      publish_date: cand.publish_date,
      rating: cand.rating ? Math.round(cand.rating * 2) : 0,
      identifiers: cand.identifiers || {}
    };

    try {
      const res = await fetch(`/api/books/${bookId}/enrich/apply`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        toast.success('✨ Metadados e Capa HD aplicados com sucesso!');
        showEnrichStudio = false;
        coverTimestamp = Date.now();
        await loadBook();
        dispatch('refresh');
      } else {
        const errData = await res.json().catch(() => ({}));
        toast.error(errData.error || 'Erro ao aplicar enriquecimento.');
      }
    } catch (e) {
      toast.error('Erro de conexão ao aplicar enriquecimento.');
    } finally {
      applyingEnrich = false;
    }
  }
</script>

<div class="modal-overlay" on:click={close} on:keydown={(e) => { if (e.key === 'Escape') close(); }} role="presentation">
  <div class="modal-content glass-card" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
    {#if loading}
      <div class="modal-loading">
        <div class="spinner"></div>
        <p>A carregar detalhes da obra...</p>
      </div>
    {:else if error}
      <div class="modal-error">
        <div class="error-icon">
          <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        </div>
        <p>{error}</p>
        <button class="btn btn-secondary" on:click={close}>Fechar</button>
      </div>
    {:else if book}
      <div class="modal-body">
        <!-- Close Button -->
        <button class="close-btn" on:click={close} title="Fechar modal (Esc)">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>

        <!-- Ambient Backdrop Glow -->
        <div class="ambient-glow" style="background-image: url('/cover/{book.id}?t={coverTimestamp}');"></div>

        <!-- Left Column: 3D Cover & Action Center -->
        <div class="modal-left">
          <div 
            class="cover-box" 
            class:dragging={isDraggingCover}
            on:dragover|preventDefault={() => { isDraggingCover = true; }}
            on:dragleave={() => { isDraggingCover = false; }}
            on:drop|preventDefault={(e) => {
              isDraggingCover = false;
              if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files[0]) {
                uploadCoverFile(e.dataTransfer.files[0]);
              }
            }}
            role="region"
            aria-label="Capa do livro"
          >
            <img 
              src="/cover/{book.id}?t={coverTimestamp}" 
              alt={book.title} 
              class="modal-cover-img"
              on:error={(e) => { e.target.src = '/static/images/cover_placeholder.svg'; }}
            />

            <!-- Subtle hover action pill to replace cover -->
            <label class="cover-change-pill" title="Clique ou arraste uma nova imagem para substituir a capa">
              <input 
                type="file" 
                accept="image/jpeg,image/png,image/webp" 
                style="display: none;" 
                on:change={(e) => { if (e.target.files && e.target.files[0]) uploadCoverFile(e.target.files[0]); }} 
              />
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path>
                <circle cx="12" cy="13" r="4"></circle>
              </svg>
              <span>Substituir</span>
            </label>
          </div>

          <!-- Hero Action: Ler Agora -->
          <button class="btn btn-hero-read" on:click={onRead}>
            <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
            </svg>
            <span>Ler Agora</span>
          </button>

          <!-- Auto-Enrich Action Button -->
          <button class="btn btn-enrich-hero" on:click={openEnrichStudio} title="Enriquecer com Capas HD e Metadados do OpenLibrary">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
            </svg>
            <span>Auto-Enriquecer ✨</span>
          </button>

          <!-- Action Grid (2x2 Balanced) -->
          <div class="action-grid">
            <button 
              class="btn btn-action-card" 
              class:active={book.read_status} 
              on:click={toggleRead}
              title={book.read_status ? 'Marcar como não lido' : 'Marcar como lido'}
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                {#if book.read_status}
                  <polyline points="20 6 9 17 4 12"></polyline>
                {:else}
                  <circle cx="12" cy="12" r="9"></circle>
                {/if}
              </svg>
              <span>{book.read_status ? 'Lido' : 'Por Ler'}</span>
            </button>

            <button 
              class="btn btn-action-card" 
              class:active={isSavedOffline} 
              on:click={toggleOfflineStatus}
              disabled={savingOffline}
              title={isSavedOffline ? 'Remover livro do armazenamento offline' : 'Guardar livro para ler em viagens ou sem internet'}
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                {#if isSavedOffline}
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                  <polyline points="22 4 12 14.01 9 11.01"></polyline>
                {:else}
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                {/if}
              </svg>
              <span>{isSavedOffline ? 'Offline ✓' : 'Ler Offline'}</span>
            </button>

            <button 
              class="btn btn-action-card" 
              on:click={() => { showSendModal = true; }}
              title="Enviar para dispositivo Kindle ou Kobo"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
              <span>e-Reader</span>
            </button>

            <button 
              class="btn btn-action-card" 
              class:active={isEditing}
              on:click={() => { isEditing = !isEditing; }}
              title="Editar dados da obra"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 20h9"></path>
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
              </svg>
              <span>{isEditing ? 'Sinopse' : 'Editar'}</span>
            </button>

            <button 
              class="btn btn-action-card" 
              on:click={runFixerOnBook}
              title="Validar e reparar estrutura do EPUB"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
              </svg>
              <span>Reparar</span>
            </button>
          </div>

          <!-- Downloads Section -->
          <div class="downloads-section">
            <span class="section-label">Descarregar</span>
            <div class="formats-row">
              {#if book.formats && book.formats.length > 0}
                {#each book.formats as f}
                  <button class="btn btn-format-pill" on:click={() => onDownload(f.format)} title="Descarregar ficheiro {f.format}">
                    <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                      <polyline points="7 10 12 15 17 10"></polyline>
                      <line x1="12" y1="15" x2="12" y2="3"></line>
                    </svg>
                    <span>{f.format}</span>
                  </button>
                {/each}
              {:else}
                <span class="no-formats">Sem ficheiros associados</span>
              {/if}
            </div>
          </div>

          <!-- Danger Action: Delete -->
          <button class="btn-delete-link" on:click={deleteBook} title="Eliminar livro da biblioteca">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
            <span>Eliminar obra</span>
          </button>
        </div>

        <!-- Right Column: Metadata / Edit Form -->
        <div class="modal-right">
          {#if isEditing}
            <!-- Metadata Edit Form -->
            <div class="edit-form">
              <div class="edit-header">
                <h3>Editar Metadados</h3>
                <span class="edit-sub">Altere título, autores e detalhes bibliográficos</span>
              </div>

              <div class="form-group">
                <label for="edit-title">Título</label>
                <input id="edit-title" type="text" bind:value={editTitle} placeholder="Título da obra" />
              </div>

              <div class="form-group">
                <label for="edit-authors">Autores (separados por vírgula)</label>
                <input id="edit-authors" type="text" bind:value={editAuthors} placeholder="Ex: Sun Tzu, Arthur Conan Doyle" />
              </div>

              <div class="form-row">
                <div class="form-group" style="flex: 2;">
                  <label for="edit-series">Série / Coleção</label>
                  <input id="edit-series" type="text" bind:value={editSeries} placeholder="Ex: Clássicos da Literatura" />
                </div>
                <div class="form-group" style="flex: 1;">
                  <label for="edit-series-idx">Índice (#)</label>
                  <input id="edit-series-idx" type="number" step="0.1" bind:value={editSeriesIndex} />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group" style="flex: 2;">
                  <label for="edit-tags">Categorias / Tags</label>
                  <input id="edit-tags" type="text" bind:value={editTags} placeholder="Ficção, Clássicos, Filosofia" />
                </div>
                <div class="form-group" style="flex: 1;">
                  <label for="edit-rating">Avaliação (0 a 10)</label>
                  <input id="edit-rating" type="number" min="0" max="10" bind:value={editRating} />
                </div>
              </div>

              <div class="form-group">
                <label for="edit-comment">Sinopse / Descrição</label>
                <textarea id="edit-comment" rows="5" bind:value={editComment} placeholder="Introdução ou sinopse da obra..."></textarea>
              </div>

              {#if editMsg}
                <div class="form-msg">{editMsg}</div>
              {/if}

              <div class="edit-actions">
                <button class="btn btn-secondary" on:click={() => { isEditing = false; }}>Cancelar</button>
                <button class="btn btn-hero-read" on:click={saveMetadata} disabled={saving}>
                  {saving ? 'A guardar...' : 'Guardar Alterações'}
                </button>
              </div>
            </div>
          {:else}
            <!-- View Mode -->
            <div class="view-header">
              <h2 class="book-title">{book.title}</h2>
              <div class="book-author">
                <span>por</span>
                <strong>{book.authors ? book.authors.join(', ') : 'Desconhecido'}</strong>
              </div>

              <!-- Metadata Badges Bar -->
              <div class="badges-bar">
                {#if book.series}
                  <span class="badge badge-series">
                    <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
                    <span>{book.series} #{book.series_index}</span>
                  </span>
                {/if}

                {#if book.rating > 0}
                  <span class="badge badge-rating">
                    <span class="stars">{'★'.repeat(Math.min(5, Math.max(1, Math.round(book.rating / 2))))}</span>
                    <span class="rating-val">{book.rating}/10</span>
                  </span>
                {/if}

                <span class="badge" class:badge-read={book.read_status} class:badge-unread={!book.read_status}>
                  {book.read_status ? '✓ Lido' : 'Não Lido'}
                </span>
              </div>
            </div>

            <!-- Tags Wrap -->
            {#if book.tags && book.tags.length > 0}
              <div class="tags-cluster">
                {#each book.tags as tag}
                  <span class="tag-chip">{tag}</span>
                {/each}
              </div>
            {/if}

            <!-- Synopsis Card -->
            <div class="synopsis-card">
              <span class="card-caption">Sinopse</span>
              <div class="synopsis-content">
                {#if book.comment}
                  {@html book.comment}
                {:else}
                  <p class="synopsis-empty">Sem descrição ou sinopse registada para esta obra.</p>
                {/if}
              </div>
            </div>

            <!-- Specs Grid (2x2 Cards) -->
            <div class="specs-grid">
              <div class="spec-card">
                <span class="spec-label">Editora</span>
                <span class="spec-value">{book.publishers && book.publishers.length > 0 ? book.publishers.join(', ') : '—'}</span>
              </div>
              <div class="spec-card">
                <span class="spec-label">Data de Publicação</span>
                <span class="spec-value">{book.pubdate ? new Date(book.pubdate).toLocaleDateString('pt-PT') : '—'}</span>
              </div>
              <div class="spec-card">
                <span class="spec-label">Idioma</span>
                <span class="spec-value">{book.languages && book.languages.length > 0 ? book.languages.join(', ') : 'Português'}</span>
              </div>
              <div class="spec-card">
                <span class="spec-label">Ficheiros</span>
                <span class="spec-value">{book.formats ? book.formats.map(f => f.format).join(', ') : 'Nenhum'}</span>
              </div>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- Send to eReader Sub-modal -->
    {#if showSendModal}
      <div class="sub-modal-backdrop" on:click={() => { showSendModal = false; }} role="presentation">
        <div class="sub-modal-card" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
          <div class="sub-modal-header">
            <h4>Enviar para e-Reader</h4>
            <button class="close-btn-sub" on:click={() => { showSendModal = false; }}>✕</button>
          </div>
          <p class="sub-modal-desc">
            Envia este livro diretamente para o teu dispositivo Kindle ou Kobo associado.
          </p>

          <div class="form-group">
            <label for="send-email">Endereço de Email do Dispositivo</label>
            <input 
              id="send-email" 
              type="email" 
              placeholder="ex: o-teu-kindle@kindle.com" 
              bind:value={sendEmail} 
            />
          </div>

          <div class="form-group">
            <label for="send-format">Formato Desejado</label>
            <select id="send-format" bind:value={sendFormat}>
              <option value="EPUB">EPUB (Moderno / Padrão)</option>
              <option value="MOBI">MOBI (Kindle Clássico)</option>
            </select>
          </div>

          {#if sendMsg}
            <div class="form-msg">{sendMsg}</div>
          {/if}

          <div class="edit-actions">
            <button class="btn btn-secondary" on:click={() => { showSendModal = false; }}>Fechar</button>
            <button class="btn btn-hero-read" on:click={sendToEReader} disabled={!sendEmail || sending}>
              {sending ? 'A enviar...' : 'Enviar Agora'}
            </button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Auto-Enrichment Studio Submodal -->
    {#if showEnrichStudio}
      <div class="sub-modal-backdrop" on:click={() => { showEnrichStudio = false; }} role="presentation">
        <div class="enrich-studio-card" on:click|stopPropagation on:keydown={() => {}} role="dialog" aria-modal="true" tabindex="-1">
          <div class="enrich-header">
            <div class="enrich-title-box">
              <span class="sparkle-icon">✨</span>
              <div>
                <h4>Auto-Enriquecimento Inteligente</h4>
                <p>Pesquisa online em múltiplos provedores (OpenLibrary, Google Books) com Capas HD e Sinopse</p>
              </div>
            </div>
            <button class="close-btn-sub" on:click={() => { showEnrichStudio = false; }}>✕</button>
          </div>

          <!-- Quick Action Bar -->
          <div class="enrich-quick-bar">
            <form class="enrich-search-box" on:submit|preventDefault={searchEnrichment}>
              <input type="text" bind:value={enrichQuery} placeholder="Título e autor para pesquisar..." />
              <button type="submit" class="btn btn-secondary btn-sm" disabled={searchingEnrich}>
                {searchingEnrich ? 'A pesquisar...' : 'Pesquisar'}
              </button>
            </form>
            <button class="btn btn-quick-auto" on:click={quickOneClickEnrich} disabled={applyingEnrich}>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              <span>{applyingEnrich ? 'A processar...' : '1-Clique Rápido'}</span>
            </button>
          </div>

          <!-- Candidates List or Loading -->
          {#if searchingEnrich}
            <div class="enrich-loading">
              <div class="spinner"></div>
              <p>A recolher capas em alta definição e metadados no OpenLibrary...</p>
            </div>
          {:else if candidates.length === 0}
            <div class="enrich-empty">
              <p>Nenhum resultado online encontrado para "{enrichQuery}". Tente ajustar os termos de pesquisa acima.</p>
            </div>
          {:else}
            <!-- Candidate Selector Tabs -->
            <div class="candidate-tabs">
              {#each candidates as cand, idx}
                <button 
                  class="candidate-tab-pill" 
                  class:active={selectedCandidateIndex === idx}
                  on:click={() => { 
                    selectedCandidateIndex = idx;
                    if (cand.covers && cand.covers.length > 0) {
                      selectedCoverURL = cand.covers[0].url;
                    }
                  }}
                >
                  <span class="tab-provider">{cand.provider === 'openlibrary' ? 'OpenLib' : 'Google'}</span>
                  <span class="tab-title">{cand.title}</span>
                </button>
              {/each}
            </div>

            <!-- Active Candidate Inspection -->
            {@const cand = candidates[selectedCandidateIndex]}
            {#if cand}
              <div class="candidate-detail-box">
                <!-- Cover Gallery (HD) -->
                {#if cand.covers && cand.covers.length > 0}
                  <div class="cover-gallery-section">
                    <span class="section-sublabel">Galeria de Capas HD ({cand.covers.length} disponíveis — clique para escolher):</span>
                    <div class="cover-gallery-scroll">
                      {#each cand.covers as cov, cIdx}
                        <div 
                          class="gallery-cover-item" 
                          class:selected={selectedCoverURL === cov.url}
                          on:click={() => { selectedCoverURL = cov.url; }}
                          role="button"
                          tabindex="0"
                          on:keydown={() => {}}
                        >
                          <img src={cov.thumb_url || cov.url} alt="Capa {cIdx + 1}" loading="lazy" />
                          {#if selectedCoverURL === cov.url}
                            <div class="selected-badge">✓</div>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Candidate Info Preview -->
                <div class="candidate-info-grid">
                  <div class="info-group">
                    <span class="info-k">Título Proposto:</span>
                    <span class="info-v">{cand.title}</span>
                  </div>
                  <div class="info-group">
                    <span class="info-k">Autor(es):</span>
                    <span class="info-v">{cand.authors ? cand.authors.join(', ') : '—'}</span>
                  </div>
                  {#if cand.publish_date}
                    <div class="info-group">
                      <span class="info-k">Ano / Data:</span>
                      <span class="info-v">{cand.publish_date}</span>
                    </div>
                  {/if}
                  {#if cand.publisher}
                    <div class="info-group">
                      <span class="info-k">Editora:</span>
                      <span class="info-v">{cand.publisher}</span>
                    </div>
                  {/if}
                </div>

                <!-- Synopsis Preview -->
                {#if cand.description}
                  <div class="synopsis-preview-box">
                    <span class="info-k">Sinopse Encontrada:</span>
                    <div class="synopsis-preview-text">{cand.description}</div>
                  </div>
                {/if}

                <!-- Tags / Genres Preview -->
                {#if cand.tags && cand.tags.length > 0}
                  <div class="tags-preview-row">
                    <span class="info-k">Géneros / Tags:</span>
                    <div class="tags-chips-list">
                      {#each cand.tags as t}
                        <span class="tag-chip-small">{t}</span>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Application Options Checkboxes -->
                <div class="enrich-options-row">
                  <label class="option-check">
                    <input type="checkbox" bind:checked={applyCover} disabled={!selectedCoverURL} />
                    <span>Substituir Capa HD</span>
                  </label>
                  <label class="option-check">
                    <input type="checkbox" bind:checked={applyDescription} disabled={!cand.description} />
                    <span>Aplicar Sinopse</span>
                  </label>
                  <label class="option-check">
                    <input type="checkbox" bind:checked={applyTags} disabled={!cand.tags || cand.tags.length === 0} />
                    <span>Adicionar Tags</span>
                  </label>
                  <label class="option-check">
                    <input type="checkbox" bind:checked={applyTitleAuthor} />
                    <span>Normalizar Título e Autor</span>
                  </label>
                </div>

                <!-- Bottom Action -->
                <div class="enrich-actions">
                  <button class="btn btn-secondary" on:click={() => { showEnrichStudio = false; }}>Cancelar</button>
                  <button class="btn btn-hero-read" on:click={applySelectedCandidate} disabled={applyingEnrich}>
                    {applyingEnrich ? 'A aplicar dados e descarregar capa...' : 'Aplicar Metadados ao Livro'}
                  </button>
                </div>
              </div>
            {/if}
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Glass Card Modal Frame */
  .glass-card {
    position: relative;
    max-width: 840px;
    width: 100%;
    background: rgba(18, 20, 26, 0.94);
    backdrop-filter: blur(28px);
    -webkit-backdrop-filter: blur(28px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 20px;
    box-shadow: 0 32px 70px -15px rgba(0, 0, 0, 0.8), 0 0 0 1px rgba(255, 255, 255, 0.05);
    overflow: hidden;
  }

  .ambient-glow {
    position: absolute;
    top: -50px;
    left: -50px;
    width: 320px;
    height: 400px;
    background-size: cover;
    background-position: center;
    filter: blur(60px);
    opacity: 0.18;
    pointer-events: none;
    z-index: 0;
  }

  .modal-body {
    position: relative;
    z-index: 1;
    display: grid;
    grid-template-columns: 230px 1fr;
    gap: 32px;
    padding: 32px;
  }

  /* Close Button */
  .close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
    z-index: 10;
  }
  .close-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #ffffff;
    transform: scale(1.05);
  }

  /* Left Column */
  .modal-left {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* 3D Realistic Book Cover */
  .cover-box {
    position: relative;
    width: 100%;
    aspect-ratio: 2 / 3;
    border-radius: 8px;
    overflow: hidden;
    background: #0d0e12;
    box-shadow: 
      -3px 0 6px rgba(0, 0, 0, 0.4),
      inset 3px 0 4px rgba(255, 255, 255, 0.22),
      inset -2px 0 6px rgba(0, 0, 0, 0.5),
      0 20px 36px -8px rgba(0, 0, 0, 0.75);
    border: 1px solid rgba(255, 255, 255, 0.1);
    transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .cover-box.dragging {
    border-color: #6366f1;
    box-shadow: 0 0 25px rgba(99, 102, 241, 0.5);
    transform: scale(1.02);
  }

  .modal-cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  /* Subtle Replace Cover Pill on Hover */
  .cover-change-pill {
    position: absolute;
    bottom: 10px;
    right: 10px;
    left: 10px;
    background: rgba(18, 20, 26, 0.85);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 20px;
    padding: 6px 12px;
    color: #f4f4f5;
    font-size: 11px;
    font-weight: 500;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    cursor: pointer;
    opacity: 0;
    transform: translateY(4px);
    transition: all 0.2s ease;
    user-select: none;
  }

  .cover-box:hover .cover-change-pill,
  .cover-box.dragging .cover-change-pill {
    opacity: 1;
    transform: translateY(0);
  }

  .cover-change-pill:hover {
    background: rgba(99, 102, 241, 0.9);
    border-color: rgba(99, 102, 241, 1);
    color: #ffffff;
  }

  /* Hero Button: Ler Agora */
  .btn-hero-read {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    padding: 10px 16px;
    border-radius: 10px;
    background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #ffffff;
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.01em;
    cursor: pointer;
    box-shadow: 0 6px 20px -3px rgba(99, 102, 241, 0.45);
    transition: all 0.2s ease;
  }
  .btn-hero-read:hover {
    background: linear-gradient(135deg, #4f46e5 0%, #4338ca 100%);
    box-shadow: 0 8px 24px -2px rgba(99, 102, 241, 0.6);
    transform: translateY(-1px);
  }

  /* Auto-Enrich Hero Button */
  .btn-enrich-hero {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    width: 100%;
    padding: 8px 14px;
    border-radius: 9px;
    background: linear-gradient(135deg, rgba(236, 72, 153, 0.18) 0%, rgba(99, 102, 241, 0.25) 100%);
    border: 1px solid rgba(236, 72, 153, 0.4);
    color: #f472b6;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  .btn-enrich-hero:hover {
    background: linear-gradient(135deg, rgba(236, 72, 153, 0.3) 0%, rgba(99, 102, 241, 0.4) 100%);
    border-color: #ec4899;
    color: #ffffff;
    box-shadow: 0 4px 16px rgba(236, 72, 153, 0.35);
    transform: translateY(-1px);
  }

  /* 2x2 Action Grid */
  .action-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .btn-action-card {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 10px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .btn-action-card:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.15);
    color: #ffffff;
  }
  .btn-action-card.active {
    background: rgba(34, 197, 94, 0.12);
    border-color: rgba(34, 197, 94, 0.35);
    color: #22c55e;
  }

  /* Downloads */
  .downloads-section {
    margin-top: 4px;
    padding-top: 10px;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .section-label {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .formats-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .btn-format-pill {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.09);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.02em;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .btn-format-pill:hover {
    background: rgba(99, 102, 241, 0.2);
    border-color: rgba(99, 102, 241, 0.4);
    color: #818cf8;
  }

  .no-formats {
    font-size: 11px;
    color: var(--text-muted);
    font-style: italic;
  }

  /* Danger Delete Link */
  .btn-delete-link {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    margin-top: 4px;
    padding: 6px 0;
    transition: color 0.15s ease;
  }
  .btn-delete-link:hover {
    color: #ef4444;
  }

  /* Right Column (Details) */
  .modal-right {
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-width: 0;
  }

  .book-title {
    font-size: 26px;
    font-weight: 700;
    color: #ffffff;
    line-height: 1.25;
    letter-spacing: -0.02em;
    margin-bottom: 4px;
    padding-right: 32px;
  }

  .book-author {
    font-size: 14px;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .book-author strong {
    color: #c7d2fe;
    font-weight: 600;
  }

  .badges-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }

  .badge-series {
    background: rgba(99, 102, 241, 0.15);
    border: 1px solid rgba(99, 102, 241, 0.3);
    color: #a5b4fc;
    font-size: 11px;
    font-weight: 600;
    padding: 3px 9px;
    border-radius: 6px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
  }

  .badge-rating {
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.25);
    color: #fbbf24;
    font-size: 11px;
    padding: 3px 9px;
    border-radius: 6px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
  }
  .stars {
    font-size: 12px;
    letter-spacing: 1px;
  }
  .rating-val {
    font-weight: 600;
  }

  .badge-read {
    background: rgba(34, 197, 94, 0.12);
    border: 1px solid rgba(34, 197, 94, 0.28);
    color: #4ade80;
    font-size: 11px;
    font-weight: 600;
    padding: 3px 8px;
    border-radius: 6px;
  }
  .badge-unread {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.09);
    color: var(--text-muted);
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 6px;
  }

  /* Tags */
  .tags-cluster {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .tag-chip {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text-secondary);
    font-size: 11px;
    padding: 3px 9px;
    border-radius: 9999px;
    transition: all 0.15s ease;
  }
  .tag-chip:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
  }

  /* Synopsis Card */
  .synopsis-card {
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .card-caption {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .synopsis-content {
    font-size: 13px;
    line-height: 1.65;
    color: #d1d5db;
    max-height: 180px;
    overflow-y: auto;
    padding-right: 4px;
  }

  .synopsis-empty {
    color: var(--text-muted);
    font-style: italic;
    font-size: 12px;
  }

  /* Specs Grid (2x2) */
  .specs-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    padding-top: 4px;
  }

  .spec-card {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 10px 14px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .spec-label {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .spec-value {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Edit Form */
  .edit-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .edit-header h3 {
    font-size: 18px;
    font-weight: 700;
    color: #ffffff;
    margin: 0;
  }
  .edit-sub {
    font-size: 12px;
    color: var(--text-muted);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .form-group label {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-muted);
  }

  .form-group input, 
  .form-group textarea, 
  .form-group select {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 8px 12px;
    color: #ffffff;
    font-size: 13px;
    outline: none;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .form-group input:focus, 
  .form-group textarea:focus, 
  .form-group select:focus {
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.22);
    background: rgba(255, 255, 255, 0.07);
  }

  .form-row {
    display: flex;
    gap: 10px;
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 6px;
  }

  .form-msg {
    font-size: 12px;
    color: #22c55e;
    font-weight: 500;
  }

  /* Submodals */
  .sub-modal-backdrop {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
    padding: 20px;
  }

  .sub-modal-card {
    max-width: 420px;
    width: 100%;
    background: rgba(24, 26, 32, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 16px;
    box-shadow: 0 24px 48px rgba(0, 0, 0, 0.6);
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .sub-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .sub-modal-header h4 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    color: #ffffff;
  }
  .close-btn-sub {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
  }
  .sub-modal-desc {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.5;
    margin: 0;
  }

  /* Auto-Enrichment Studio Card */
  .enrich-studio-card {
    max-width: 760px;
    width: 100%;
    max-height: 85vh;
    overflow-y: auto;
    background: rgba(22, 24, 30, 0.98);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 18px;
    box-shadow: 0 32px 64px rgba(0, 0, 0, 0.7);
    padding: 26px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .enrich-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  .enrich-title-box {
    display: flex;
    gap: 12px;
    align-items: center;
  }
  .sparkle-icon {
    font-size: 24px;
  }
  .enrich-title-box h4 {
    margin: 0;
    font-size: 18px;
    font-weight: 700;
    color: #ffffff;
  }
  .enrich-title-box p {
    margin: 2px 0 0;
    font-size: 12px;
    color: var(--text-muted);
  }

  /* Quick Bar */
  .enrich-quick-bar {
    display: flex;
    gap: 12px;
    align-items: center;
  }
  .enrich-search-box {
    flex: 1;
    display: flex;
    gap: 8px;
  }
  .enrich-search-box input {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 8px;
    padding: 8px 12px;
    color: #ffffff;
    font-size: 13px;
    outline: none;
  }
  .enrich-search-box input:focus {
    border-color: #6366f1;
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
  }

  .btn-quick-auto {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    border-radius: 8px;
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s ease;
  }
  .btn-quick-auto:hover {
    background: linear-gradient(135deg, #059669 0%, #047857 100%);
    transform: translateY(-1px);
    box-shadow: 0 4px 14px rgba(16, 185, 129, 0.4);
  }

  /* Candidate Tabs */
  .candidate-tabs {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding-bottom: 4px;
  }
  .candidate-tab-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.15s ease;
  }
  .candidate-tab-pill.active {
    background: rgba(99, 102, 241, 0.18);
    border-color: #6366f1;
    color: #ffffff;
  }
  .tab-provider {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    background: rgba(255, 255, 255, 0.1);
    padding: 2px 6px;
    border-radius: 4px;
  }
  .tab-title {
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Candidate Detail Box */
  .candidate-detail-box {
    display: flex;
    flex-direction: column;
    gap: 16px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 16px;
  }

  .cover-gallery-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .section-sublabel {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .cover-gallery-scroll {
    display: flex;
    gap: 10px;
    overflow-x: auto;
    padding: 6px 2px 10px;
  }
  .gallery-cover-item {
    position: relative;
    width: 90px;
    height: 135px;
    flex-shrink: 0;
    border-radius: 6px;
    overflow: hidden;
    cursor: pointer;
    border: 2px solid transparent;
    box-shadow: 0 4px 10px rgba(0,0,0,0.5);
    transition: all 0.2s ease;
  }
  .gallery-cover-item img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .gallery-cover-item:hover {
    transform: scale(1.04);
  }
  .gallery-cover-item.selected {
    border-color: #6366f1;
    box-shadow: 0 0 16px rgba(99, 102, 241, 0.6);
  }
  .selected-badge {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #6366f1;
    color: #fff;
    font-size: 11px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .candidate-info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }
  .info-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .info-k {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
  }
  .info-v {
    font-size: 12px;
    color: #ffffff;
    font-weight: 500;
  }

  .synopsis-preview-box {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .synopsis-preview-text {
    font-size: 12px;
    line-height: 1.55;
    color: #cbd5e1;
    max-height: 100px;
    overflow-y: auto;
    background: rgba(0, 0, 0, 0.2);
    padding: 8px 10px;
    border-radius: 6px;
  }

  .tags-preview-row {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .tags-chips-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .tag-chip-small {
    font-size: 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 2px 8px;
    color: var(--text-secondary);
  }

  .enrich-options-row {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
  }
  .option-check {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #f1f5f9;
    cursor: pointer;
  }
  .option-check input {
    accent-color: #6366f1;
    cursor: pointer;
  }

  .enrich-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 4px;
  }

  .enrich-loading, .enrich-empty {
    padding: 40px 20px;
    text-align: center;
    color: var(--text-muted);
    font-size: 13px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  /* Loading & Error States */
  .modal-loading, .modal-error {
    padding: 80px 20px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    color: var(--text-secondary);
  }

  .error-icon {
    color: #ef4444;
  }

  .spinner {
    width: 36px;
    height: 36px;
    border: 3px solid rgba(99, 102, 241, 0.2);
    border-top-color: #6366f1;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Responsiveness */
  @media (max-width: 720px) {
    .modal-body {
      grid-template-columns: 1fr;
      padding: 24px;
      gap: 20px;
    }
    .modal-left {
      max-width: 220px;
      margin: 0 auto;
    }
    .specs-grid {
      grid-template-columns: 1fr;
    }
    .enrich-options-row {
      flex-direction: column;
      gap: 8px;
    }
  }
</style>
