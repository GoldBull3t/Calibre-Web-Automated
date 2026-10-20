<script>
  import { onMount } from 'svelte';

  let plugins = [];
  let loading = true;
  let uploadFile = null;
  let uploading = false;
  let message = '';

  onMount(loadPlugins);

  async function loadPlugins() {
    loading = true;
    try {
      const res = await fetch('/api/admin/plugins');
      const data = await res.json();
      plugins = data.plugins || [];
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  async function togglePlugin(filename, currentEnabled) {
    try {
      const res = await fetch(`/api/admin/plugins/toggle/${encodeURIComponent(filename)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled: !currentEnabled })
      });
      const data = await res.json();
      if (data.status === 'success') {
        plugins = plugins.map(p => p.filename === filename ? { ...p, enabled: !currentEnabled } : p);
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function handleUpload() {
    if (!uploadFile) return;
    uploading = true;
    message = '';
    const formData = new FormData();
    formData.append('plugin_file', uploadFile);

    try {
      const res = await fetch('/api/admin/plugins/upload', {
        method: 'POST',
        body: formData
      });
      const data = await res.json();
      if (data.status === 'success') {
        message = '✓ ' + data.message;
        uploadFile = null;
        await loadPlugins();
      } else {
        message = 'Erro: ' + (data.error || 'Falha ao instalar');
      }
    } catch (e) {
      message = 'Erro ao enviar ficheiro do plugin.';
    } finally {
      uploading = false;
    }
  }
</script>

<div class="plugins-container">
  <!-- Header Card -->
  <div class="plugins-hero glass">
    <div class="hero-content">
      <h2>Plugins Calibre & Motor DeDRM</h2>
      <p>Gere e execute plugins oficiais do ecossistema Calibre diretamente pelo CWA. O motor DeDRM analisa e remove a encriptação de ficheiros protegidos no momento da ingestão.</p>
    </div>

    <!-- Upload Card -->
    <div class="upload-plugin-box">
      <label class="file-label">
        <input 
          type="file" 
          accept=".zip" 
          on:change={(e) => { uploadFile = e.target.files[0]; }} 
        />
        <span>{uploadFile ? uploadFile.name : 'Selecionar ficheiro .zip de plugin...'}</span>
      </label>
      <button 
        class="btn btn-primary" 
        on:click={handleUpload} 
        disabled={!uploadFile || uploading}>
        {uploading ? 'A instalar...' : 'Instalar Plugin'}
      </button>
    </div>

    {#if message}
      <div class="msg-box">{message}</div>
    {/if}
  </div>

  <!-- DeDRM Banner -->
  <div class="dedrm-banner glass">
    <div class="banner-icon">
      <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
    </div>
    <div class="banner-text">
      <h4>Proteção DeDRM Ativa</h4>
      <p>Ficheiros EPUB com esquemas Adobe DRM / FairPlay passam pela inspeção de encriptação em <code>META-INF/encryption.xml</code> antes da indexação no Calibre.</p>
    </div>
  </div>

  <!-- Plugins List -->
  <div class="plugins-list-section">
    <h3>Plugins Instalados</h3>

    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>A carregar plugins...</p>
      </div>
    {:else if plugins.length === 0}
      <div class="empty-plugins glass">
        <p>Nenhum plugin adicional instalado na pasta <code>plugins/</code>.</p>
      </div>
    {:else}
      <div class="plugins-grid">
        {#each plugins as p}
          <div class="plugin-card glass glass-hover">
            <div class="plugin-top">
              <div>
                <h4 class="plugin-name">{p.name || p.filename}</h4>
                <span class="plugin-ver">v{p.version || '1.0'}</span>
              </div>
              
              <!-- Toggle Switch -->
              <label class="switch">
                <input 
                  type="checkbox" 
                  checked={p.enabled} 
                  on:change={() => togglePlugin(p.filename, p.enabled)} 
                />
                <span class="slider"></span>
              </label>
            </div>

            <p class="plugin-desc">{p.description || 'Plugin oficial para Calibre.'}</p>
            <div class="plugin-author">por {p.author || 'Calibre Community'}</div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .plugins-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .plugins-hero {
    padding: 28px;
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .hero-content h2 {
    font-size: 22px;
    font-weight: 700;
    margin-bottom: 6px;
    color: var(--text-primary);
  }
  .hero-content p {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .upload-plugin-box {
    display: flex;
    gap: 12px;
    max-width: 550px;
  }

  .file-label {
    flex: 1;
    background: rgba(255, 255, 255, 0.05);
    border: 1px dashed var(--border-secondary);
    border-radius: var(--radius-sm);
    padding: 8px 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    color: var(--text-secondary);
    font-size: 13px;
    overflow: hidden;
  }
  .file-label input {
    display: none;
  }
  .file-label span {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .msg-box {
    font-size: 13px;
    color: var(--accent);
  }

  .dedrm-banner {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 18px 24px;
    border-radius: var(--radius-md);
    background: rgba(56, 189, 248, 0.06);
    border-color: rgba(56, 189, 248, 0.2);
  }

  .banner-icon {
    font-size: 28px;
  }
  .banner-text h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--accent);
    margin-bottom: 2px;
  }
  .banner-text p {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .plugins-list-section h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text-primary);
  }

  .plugins-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
  }

  .plugin-card {
    padding: 18px;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .plugin-top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }

  .plugin-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .plugin-ver {
    font-size: 11px;
    color: var(--text-muted);
  }

  .plugin-desc {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .plugin-author {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: auto;
  }

  /* Switch */
  .switch {
    position: relative;
    display: inline-block;
    width: 38px;
    height: 22px;
  }
  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: rgba(255, 255, 255, 0.15);
    transition: 0.2s;
    border-radius: 22px;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: 16px;
    width: 16px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.2s;
    border-radius: 50%;
  }
  input:checked + .slider {
    background-color: var(--accent);
  }
  input:checked + .slider:before {
    transform: translateX(16px);
  }

  .empty-plugins {
    padding: 32px;
    text-align: center;
    border-radius: var(--radius-md);
    color: var(--text-secondary);
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
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>

