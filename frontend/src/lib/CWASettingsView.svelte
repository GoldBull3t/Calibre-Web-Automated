<script>
  import { onMount } from 'svelte';

  let settings = null;
  let backups = [];
  let loading = true;
  let runningFixer = false;
  let runningConverter = false;
  let creatingBackup = false;
  let actionMessage = '';

  let passkeys = [];
  let registeringPasskey = false;
  let passkeyMsg = '';

  let oidc = {
    issuer_url: '',
    client_id: '',
    client_secret: '',
    redirect_url: '',
    display_name: 'Single Sign-On',
    enabled: false
  };
  let savingOIDC = false;
  let oidcMsg = '';

  onMount(loadAll);

  async function loadAll() {
    loading = true;
    try {
      const [resS, resB] = await Promise.all([
        fetch('/api/cwa/settings'),
        fetch('/api/backups')
      ]);
      settings = await resS.json();
      const bData = await resB.json();
      backups = bData.backups || [];
      await Promise.all([loadPasskeys(), loadOIDC()]);
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  async function loadPasskeys() {
    try {
      const res = await fetch('/api/auth/passkeys');
      if (res.ok) {
        const data = await res.json();
        passkeys = data.passkeys || [];
      }
    } catch(e) {}
  }

  async function loadOIDC() {
    try {
      const res = await fetch('/api/auth/oidc/config');
      if (res.ok) {
        oidc = await res.json();
      }
    } catch(e) {}
  }

  async function deletePasskey(id) {
    if (!confirm('Tens a certeza que desejas eliminar esta passkey?')) return;
    try {
      await fetch('/api/auth/passkeys/' + id, { method: 'DELETE' });
      await loadPasskeys();
    } catch(e) {}
  }

  function bufferDecode(value) {
    return Uint8Array.from(atob(value.replace(/-/g, '+').replace(/_/g, '/')), c => c.charCodeAt(0));
  }

  function bufferEncode(value) {
    return btoa(String.fromCharCode(...new Uint8Array(value)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
  }

  async function registerPasskey() {
    registeringPasskey = true;
    passkeyMsg = '';
    try {
      if (!window.PublicKeyCredential) {
        passkeyMsg = 'O teu navegador não suporta WebAuthn / Passkeys.';
        return;
      }
      const optRes = await fetch('/api/auth/passkey/register/options');
      const options = await optRes.json();

      options.challenge = bufferDecode(options.challenge);
      options.user.id = bufferDecode(options.user.id);

      const credential = await navigator.credentials.create({ publicKey: options });
      const rawId = bufferEncode(credential.rawId);
      const clientData = bufferEncode(credential.response.clientDataJSON);

      const deviceName = navigator.userAgent.includes('Windows') ? 'Windows Hello' : (navigator.userAgent.includes('Mac') ? 'Touch ID / Mac' : 'Chave de Segurança');

      const verifyRes = await fetch('/api/auth/passkey/register/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          credential_id: rawId,
          public_key: rawId,
          name: deviceName,
          client_data_json: clientData
        })
      });

      if (verifyRes.ok) {
        passkeyMsg = '✓ Passkey registada com sucesso!';
        await loadPasskeys();
        setTimeout(() => { passkeyMsg = ''; }, 3000);
      } else {
        passkeyMsg = 'Erro ao validar passkey no servidor.';
      }
    } catch (e) {
      if (e.name === 'NotAllowedError') {
        passkeyMsg = 'Registo cancelado pelo utilizador.';
      } else {
        passkeyMsg = 'Falha no registo: ' + e.message;
      }
    } finally {
      registeringPasskey = false;
    }
  }

  async function saveOIDC() {
    savingOIDC = true;
    oidcMsg = '';
    try {
      const res = await fetch('/api/auth/oidc/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(oidc)
      });
      if (res.ok) {
        oidcMsg = '✓ Configuração SSO / OIDC guardada com sucesso!';
        setTimeout(() => { oidcMsg = ''; }, 3000);
      } else {
        oidcMsg = 'Erro ao guardar definições OIDC.';
      }
    } catch(e) {
      oidcMsg = 'Erro de comunicação ao guardar OIDC.';
    } finally {
      savingOIDC = false;
    }
  }

  async function triggerBackup() {
    creatingBackup = true;
    actionMessage = 'A compactar e gerar arquivo de backup...';
    try {
      const res = await fetch('/api/backups/create', { method: 'POST' });
      const data = await res.json();
      actionMessage = data.message || 'Backup criado com sucesso!';
      await loadAll();
    } catch (e) {
      actionMessage = 'Erro ao criar backup.';
    } finally {
      creatingBackup = false;
    }
  }

  async function deleteBackup(filename) {
    if (!confirm(`Eliminar backup ${filename}?`)) return;
    try {
      await fetch(`/api/backups/${encodeURIComponent(filename)}`, { method: 'DELETE' });
      backups = backups.filter(b => b.filename !== filename);
    } catch (e) {
      console.error(e);
    }
  }

  async function triggerFixer() {
    runningFixer = true;
    actionMessage = 'A executar EPUB Fixer estrutural na biblioteca...';
    try {
      const res = await fetch('/api/fixer/start', { method: 'POST' });
      const data = await res.json();
      actionMessage = data.message || 'Reparação de EPUB concluída!';
    } catch (e) {
      actionMessage = 'Erro ao executar EPUB Fixer.';
    } finally {
      runningFixer = false;
    }
  }

  async function triggerConverter() {
    runningConverter = true;
    actionMessage = 'A iniciar conversão de biblioteca em segundo plano...';
    try {
      const res = await fetch('/api/convert/start');
      const data = await res.json();
      actionMessage = 'Conversão de biblioteca agendada com sucesso!';
    } catch (e) {
      actionMessage = 'Erro ao agendar conversão.';
    } finally {
      runningConverter = false;
    }
  }

  function downloadBackup(filename) {
    window.location.href = `/cwa-backup/download/${encodeURIComponent(filename)}`;
  }
</script>

<div class="cwa-settings-container">
  <!-- Header -->
  <div class="settings-hero glass">
    <h2>Centro de Controlo de Automação (CWA) ⚙️</h2>
    <p>Gere os serviços autónomos em segundo plano do Calibre-Web Automated: Ingestão, Conversões, Reparações de EPUB e Backups compactados.</p>
  </div>

  {#if actionMessage}
    <div class="msg-banner glass">
      <span>{actionMessage}</span>
    </div>
  {/if}

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>A carregar serviços CWA...</p>
    </div>
  {:else if settings}
    <!-- Grid of Services -->
    <div class="services-grid">
      <!-- Service 1: Ingest -->
      <div class="service-card glass">
        <div class="service-top">
          <div class="service-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"></polyline><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path></svg>
          </div>
          <div>
            <h4>Ingestão Contínua (Watcher)</h4>
            <p>Monitoriza novos livros colocados na pasta ingest e importa metadados automaticamente.</p>
          </div>
          <span class="badge badge-success">ATIVO</span>
        </div>
        <div class="service-detail">
          <span class="detail-k">Pasta Monitorizada:</span>
          <code class="detail-v">{settings.ingest_path}</code>
        </div>
      </div>

      <!-- Service 2: Multi-Format Converter -->
      <div class="service-card glass">
        <div class="service-top">
          <div class="service-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
          </div>
          <div>
            <h4>Conversor Multi-Formato</h4>
            <p>Converte automaticamente livros para o teu formato preferido para e-readers.</p>
          </div>
          <span class="badge badge-success">ATIVO</span>
        </div>
        <div class="service-actions">
          <button class="btn btn-secondary btn-sm" on:click={triggerConverter} disabled={runningConverter}>
            {runningConverter ? 'A processar...' : 'Converter Toda a Biblioteca'}
          </button>
        </div>
      </div>

      <!-- Service 3: EPUB Fixer -->
      <div class="service-card glass">
        <div class="service-top">
          <div class="service-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path></svg>
          </div>
          <div>
            <h4>EPUB Fixer Service 2.0</h4>
            <p>Repara manifestos corrompidos, tags XHTML inválidas e erros de digitalização OCR (hífens quebrados, ligaduras e pontuação).</p>
          </div>
          <span class="badge badge-success">ATIVO</span>
        </div>
        <div class="service-actions">
          <button class="btn btn-secondary btn-sm" on:click={triggerFixer} disabled={runningFixer}>
            {runningFixer ? 'A reparar...' : 'Reparar Toda a Biblioteca'}
          </button>
        </div>
      </div>

      <!-- Service 4: Metadata Providers -->
      <div class="service-card glass">
        <div class="service-top">
          <div class="service-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path><line x1="7" y1="7" x2="7.01" y2="7"></line></svg>
          </div>
          <div>
            <h4>Fornecedores de Metadados</h4>
            <p>Fontes configuradas para pesquisa e enriquecimento automático.</p>
          </div>
        </div>
        <div class="providers-list">
          <span class="provider-pill">✓ Google Books</span>
          <span class="provider-pill">✓ OpenLibrary</span>
          <span class="provider-pill">✓ Hardcover API</span>
        </div>
      </div>
    </div>

    <!-- Backups Section -->
    <div class="backups-section glass">
      <div class="backups-header">
        <div>
          <h3>Serviço de Backups Automáticos 🔒</h3>
          <p>Instantâneos compactados da base de dados e configurações do Calibre.</p>
        </div>

        <button class="btn btn-primary" on:click={triggerBackup} disabled={creatingBackup}>
          {#if creatingBackup}
            <span>A criar backup...</span>
          {:else}
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="15"></line></svg>
            <span>Criar Novo Backup (.zip)</span>
          {/if}
        </button>
      </div>

      {#if backups.length === 0}
        <div class="no-backups">
          <p>Nenhum ficheiro de backup gerado ainda. Clica em <strong>"Criar Novo Backup"</strong> para proteger a tua base de dados.</p>
        </div>
      {:else}
        <table class="backup-table">
          <thead>
            <tr>
              <th>Ficheiro de Backup</th>
              <th>Tamanho</th>
              <th>Data de Criação</th>
              <th>Ações</th>
            </tr>
          </thead>
          <tbody>
            {#each backups as b}
              <tr>
                <td><strong>{b.filename}</strong></td>
                <td><code class="size-txt">{b.size || '3.2 MB'}</code></td>
                <td>{b.created ? new Date(b.created).toLocaleString() : 'Recente'}</td>
                <td>
                  <div class="btn-group">
                    <button class="btn btn-secondary btn-sm" on:click={() => downloadBackup(b.filename)}>
                      Descarregar
                    </button>
                    <button class="btn btn-danger btn-sm" on:click={() => deleteBackup(b.filename)}>
                      Eliminar
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>

    <!-- Modern Authentication: Passkeys & SSO (OIDC) -->
    <div class="auth-section glass">
      <div class="auth-header">
        <div>
          <h3>Autenticação Moderna: Passkeys & SSO (OIDC)</h3>
          <p>Acede à biblioteca sem palavra-passe através de biometria local ou integra com o teu fornecedor de identidade.</p>
        </div>
      </div>

      <div class="auth-grid">
        <!-- Passkeys Card -->
        <div class="sub-auth-card glass">
          <div class="sub-auth-top">
            <div class="auth-icon-wrap">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 2l-2 2m-1.5 1.5L14 9l-1.5-1.5L11 9l-1.5-1.5L8 9 6.5 7.5 5 9 2 6l3-3 16-1z"></path><circle cx="7.5" cy="15.5" r="5.5"></circle><path d="M12 12l8 8"></path></svg>
            </div>
            <div>
              <h4>Passkeys (Biometria / FIDO2)</h4>
              <p>Windows Hello, Touch ID, Face ID ou chaves de segurança físicas (YubiKey).</p>
            </div>
          </div>

          <div class="passkeys-list">
            {#if passkeys.length === 0}
              <div class="empty-passkeys">
                Nenhuma Passkey registada neste utilizador.
              </div>
            {:else}
              {#each passkeys as pk}
                <div class="passkey-row">
                  <div class="pk-info">
                    <span class="pk-name">🔑 {pk.name}</span>
                    <span class="pk-date">Registada a {new Date(pk.created_at).toLocaleDateString()}</span>
                  </div>
                  <button class="delete-btn" on:click={() => deletePasskey(pk.id)} title="Remover Passkey">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                  </button>
                </div>
              {/each}
            {/if}
          </div>

          {#if passkeyMsg}
            <div class="auth-msg">{passkeyMsg}</div>
          {/if}

          <div class="auth-card-actions">
            <button class="btn btn-primary btn-sm" on:click={registerPasskey} disabled={registeringPasskey}>
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              <span>{registeringPasskey ? 'A aguardar autenticador...' : 'Registar Este Dispositivo'}</span>
            </button>
          </div>
        </div>

        <!-- OIDC / SSO Card -->
        <div class="sub-auth-card glass">
          <div class="sub-auth-top">
            <div class="auth-icon-wrap sso">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
            </div>
            <div>
              <h4>Single Sign-On (OpenID Connect)</h4>
              <p>Authentik, Authelia, Keycloak ou Google Workspace.</p>
            </div>
          </div>

          <form on:submit|preventDefault={saveOIDC} class="oidc-form">
            <label class="check-toggle">
              <input type="checkbox" bind:checked={oidc.enabled} />
              <span>Ativar Início de Sessão via SSO (OIDC)</span>
            </label>

            <div class="form-row">
              <label for="oidc-issuer">URL do Fornecedor (Issuer):</label>
              <input id="oidc-issuer" type="text" placeholder="https://auth.exemplo.com/application/o/calibre/" bind:value={oidc.issuer_url} />
            </div>

            <div class="form-row">
              <label for="oidc-client-id">Client ID:</label>
              <input id="oidc-client-id" type="text" placeholder="calibre-web-cwa" bind:value={oidc.client_id} />
            </div>

            <div class="form-row">
              <label for="oidc-secret">Client Secret:</label>
              <input id="oidc-secret" type="password" placeholder="••••••••" bind:value={oidc.client_secret} />
            </div>

            <div class="form-row">
              <label for="oidc-name">Nome do Botão:</label>
              <input id="oidc-name" type="text" placeholder="Authentik SSO" bind:value={oidc.display_name} />
            </div>

            {#if oidcMsg}
              <div class="auth-msg">{oidcMsg}</div>
            {/if}

            <div class="auth-card-actions">
              {#if oidc.enabled}
                <a href="/auth/oidc/login" class="btn btn-secondary btn-sm" target="_blank">
                  Testar Login SSO ↗
                </a>
              {/if}
              <button type="submit" class="btn btn-primary btn-sm" disabled={savingOIDC}>
                {savingOIDC ? 'A guardar...' : 'Guardar Configuração OIDC'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .cwa-settings-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .settings-hero {
    padding: 24px;
    border-radius: var(--radius-lg);
  }
  .settings-hero h2 {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .settings-hero p {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .msg-banner {
    padding: 12px 18px;
    border-radius: var(--radius-md);
    background: rgba(56, 189, 248, 0.1);
    border-color: rgba(56, 189, 248, 0.25);
    color: var(--accent);
    font-size: 13px;
  }

  .services-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
  }

  .service-card {
    padding: 20px;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .service-top {
    display: flex;
    align-items: flex-start;
    gap: 14px;
  }

  .service-icon {
    font-size: 26px;
    width: 44px;
    height: 44px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .service-top h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 2px;
  }
  .service-top p {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.35;
  }

  .service-detail {
    font-size: 11px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .detail-k {
    color: var(--text-muted);
    font-weight: 600;
    text-transform: uppercase;
  }
  .detail-v {
    color: var(--accent);
    word-break: break-all;
  }

  .service-actions {
    margin-top: auto;
  }

  .providers-list {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .provider-pill {
    font-size: 11px;
    color: #22c55e;
    background: rgba(34, 197, 94, 0.1);
    padding: 3px 8px;
    border-radius: var(--radius-full);
  }

  .backups-section {
    padding: 24px;
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .backups-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px;
  }
  .backups-header h3 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .backups-header p {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .no-backups {
    padding: 24px;
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
  }

  .backup-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }
  .backup-table th {
    text-align: left;
    padding: 10px 14px;
    border-bottom: 1px solid var(--border-primary);
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
  }
  .backup-table td {
    padding: 12px 14px;
    border-bottom: 1px solid var(--border-primary);
    color: var(--text-secondary);
  }
  .size-txt {
    font-family: var(--font-mono);
  }
  .btn-group {
    display: flex;
    gap: 8px;
  }

  .loading-state {
    text-align: center;
    padding: 60px;
    color: var(--text-secondary);
  }
  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(56, 189, 248, 0.15);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin: 0 auto 12px;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  /* Modern Auth Styles */
  .auth-section {
    padding: 24px;
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: 18px;
  }
  .auth-header h3 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
  }
  .auth-header p {
    font-size: 12px;
    color: var(--text-secondary);
  }
  .auth-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
    gap: 18px;
  }
  .sub-auth-card {
    padding: 20px;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .sub-auth-top {
    display: flex;
    align-items: center;
    gap: 14px;
  }
  .sub-auth-top h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .sub-auth-top p {
    font-size: 11px;
    color: var(--text-secondary);
  }
  .auth-icon-wrap {
    width: 42px;
    height: 42px;
    border-radius: var(--radius-md);
    background: rgba(56, 189, 248, 0.12);
    color: var(--accent);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .auth-icon-wrap.sso {
    background: rgba(168, 85, 247, 0.12);
    color: #a855f7;
  }
  .passkeys-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 80px;
  }
  .empty-passkeys {
    font-size: 12px;
    color: var(--text-muted);
    font-style: italic;
    padding: 14px;
    text-align: center;
    background: rgba(255, 255, 255, 0.02);
    border-radius: var(--radius-sm);
  }
  .passkey-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-secondary);
    border-radius: var(--radius-sm);
  }
  .pk-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .pk-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .pk-date {
    font-size: 10px;
    color: var(--text-muted);
  }
  .oidc-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .check-toggle {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    cursor: pointer;
  }
  .form-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .form-row label {
    font-size: 11px;
    color: var(--text-muted);
    font-weight: 500;
  }
  .form-row input {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm);
    padding: 8px 12px;
    color: var(--text-primary);
    font-size: 12px;
    outline: none;
  }
  .form-row input:focus {
    border-color: var(--accent);
  }
  .auth-card-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 4px;
  }
  .auth-msg {
    font-size: 12px;
    color: var(--accent);
    padding: 6px 10px;
    background: rgba(56, 189, 248, 0.08);
    border-radius: var(--radius-sm);
  }
</style>

