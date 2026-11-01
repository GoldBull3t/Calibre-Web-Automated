<script>
  // @ts-nocheck
  import { onMount } from 'svelte';

  let sources = [];
  let isAdmin = false;
  let query = '';
  let sourceFilter = '';
  let results = [];
  let errors = {};
  let loading = false;
  let searched = false;
  let downloadingKey = null;
  let message = '';
  let messageKind = 'ok';
  let editing = null; // id da fonte em edição
  let form = {};
  let saving = false;

  onMount(loadSources);

  async function loadSources() {
    try {
      const res = await fetch('/api/sources');
      if (res.status === 403) { isAdmin = false; return; }
      const data = await res.json();
      sources = data.sources || [];
      isAdmin = true;
    } catch (e) { console.error(e); }
  }

  function notify(text, kind = 'ok') { message = text; messageKind = kind; }

  function startEdit(s) {
    editing = s.id;
    form = { enabled: s.enabled || !s.configured, values: { ...s.values } };
  }

  async function saveSource(s) {
    saving = true;
    try {
      const res = await fetch(`/api/sources/${s.id}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ values: form.values, enabled: form.enabled })
      });
      const data = await res.json();
      if (res.ok) {
        sources = data.sources || sources;
        editing = null;
        notify(`Configuração de ${s.name} guardada.`);
      } else {
        notify(data.error || 'Falha ao guardar.', 'err');
      }
    } catch (e) {
      notify('Erro de comunicação com o servidor.', 'err');
    } finally { saving = false; }
  }

  async function search() {
    if (query.trim().length < 2) return;
    loading = true; searched = true; message = '';
    try {
      const qs = new URLSearchParams({ q: query.trim() });
      if (sourceFilter) qs.set('source', sourceFilter);
      const res = await fetch('/api/sources/search?' + qs);
      const data = await res.json();
      if (!res.ok) { notify(data.error || 'Falha na pesquisa.', 'err'); results = []; errors = {}; return; }
      results = data.results || [];
      errors = data.errors || {};
    } catch (e) {
      notify('Erro de rede ao pesquisar.', 'err');
    } finally { loading = false; }
  }

  async function download(r) {
    const key = r.source + ':' + r.id;
    downloadingKey = key;
    try {
      const res = await fetch('/api/sources/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source: r.source, id: r.id })
      });
      const data = await res.json();
      if (res.ok) notify(`✓ "${r.title}" enviado para a pasta de ingestão.`);
      else notify('Erro: ' + (data.error || 'Falha no download'), 'err');
    } catch (e) {
      notify('Erro de rede ao descarregar.', 'err');
    } finally { downloadingKey = null; }
  }

  function sourceName(id) { return sources.find(s => s.id === id)?.name || id; }
  $: active = sources.filter(s => s.enabled);
  function onKey(e) { if (e.key === 'Enter') search(); }
</script>

<div class="src-container">
  <div class="src-hero glass">
    <h2>Fontes de livros</h2>
    <p>Pesquise e descarregue livros de várias fontes diretamente para a sua biblioteca.</p>

    <div class="search-row">
      <input type="text" placeholder="Título, autor ou ISBN..." bind:value={query} on:keydown={onKey} />
      <select bind:value={sourceFilter}>
        <option value="">Todas as fontes</option>
        {#each active as s}<option value={s.id}>{s.name}</option>{/each}
      </select>
      <button class="btn btn-primary" on:click={search} disabled={loading}>
        {loading ? 'A pesquisar...' : 'Pesquisar'}
      </button>
    </div>
  </div>

  {#if message}
    <div class="alert-box glass" class:err={messageKind === 'err'}>{message}</div>
  {/if}

  {#each Object.entries(errors) as [id, err]}
    <div class="alert-box glass err">{sourceName(id)}: {err}</div>
  {/each}

  {#if loading}
    <div class="loading-state"><div class="spinner"></div><p>A consultar as fontes...</p></div>
  {:else if results.length > 0}
    <div class="results glass">
      <table>
        <thead><tr><th></th><th>Título &amp; Autor</th><th>Fonte</th><th>Formato</th><th>Ação</th></tr></thead>
        <tbody>
          {#each results as r (r.source + ':' + r.id)}
            <tr>
              <td class="cover-cell">
                {#if r.cover}<img src={r.cover} alt="" loading="lazy" referrerpolicy="no-referrer" />{/if}
              </td>
              <td>
                <div class="t-title">{r.title}</div>
                <div class="t-author">{r.author || '—'}{r.year ? ' · ' + r.year : ''}</div>
              </td>
              <td><span class="badge">{sourceName(r.source)}</span></td>
              <td><span class="mono">{(r.format || '').toUpperCase()} {r.size || ''}</span></td>
              <td>
                <button class="btn btn-primary btn-sm" on:click={() => download(r)}
                  disabled={downloadingKey === r.source + ':' + r.id}>
                  {downloadingKey === r.source + ':' + r.id ? 'A descarregar...' : 'Descarregar'}
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if searched && Object.keys(errors).length === 0}
    <div class="loading-state"><p>Sem resultados.</p></div>
  {/if}

  {#if isAdmin}
    <div class="cfg glass">
      <h3>Configuração das fontes</h3>
      {#each sources as s}
        <div class="cfg-item">
          <div class="cfg-head">
            <div>
              <strong>{s.name}</strong>
              <span class="status" class:on={s.enabled}>
                {s.enabled ? 'Ativa' : s.configured ? 'Desativada' : 'Por configurar'}
              </span>
              <div class="desc">{s.description}</div>
            </div>
            {#if s.fields && s.fields.length > 0 && editing !== s.id}
              <button class="btn btn-sm" on:click={() => startEdit(s)}>Configurar</button>
            {:else if !s.fields || s.fields.length === 0}
              <span class="desc">Sem configuração necessária</span>
            {/if}
          </div>

          {#if editing === s.id}
            <div class="cfg-form">
              {#each s.fields as f}
                <label>
                  <span>{f.label}</span>
                  <input type={f.secret ? 'password' : 'text'} bind:value={form.values[f.key]}
                    placeholder={f.default || ''} autocomplete="off" />
                </label>
              {/each}
              <label class="chk"><input type="checkbox" bind:checked={form.enabled} /> Ativa</label>
              <div class="cfg-actions">
                <button class="btn btn-primary btn-sm" on:click={() => saveSource(s)} disabled={saving}>
                  {saving ? 'A guardar...' : 'Guardar'}
                </button>
                <button class="btn btn-sm" on:click={() => (editing = null)}>Cancelar</button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .src-container { max-width: 1400px; margin: 0 auto; padding: 24px; display: flex; flex-direction: column; gap: 20px; }
  .src-hero { padding: 28px; border-radius: var(--radius-lg); display: flex; flex-direction: column; gap: 12px; }
  .src-hero h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); }
  .src-hero p { font-size: 13px; color: var(--text-secondary); }
  .search-row { display: flex; gap: 10px; flex-wrap: wrap; max-width: 800px; }
  input[type='text'], input[type='password'], select {
    background: rgba(255, 255, 255, 0.05); border: 1px solid var(--border-primary);
    border-radius: var(--radius-sm); padding: 10px 14px; color: var(--text-primary); font-size: 13px; outline: none;
  }
  .search-row input { flex: 1; min-width: 200px; }
  input:focus, select:focus { border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-glow); }
  .alert-box { padding: 12px 18px; border-radius: var(--radius-md); background: rgba(34, 197, 94, 0.12);
    border: 1px solid rgba(34, 197, 94, 0.3); color: #22c55e; font-size: 13px; }
  .alert-box.err { background: rgba(248, 113, 113, 0.12); border-color: rgba(248, 113, 113, 0.3); color: #f87171; }
  .results { border-radius: var(--radius-md); overflow: hidden; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; text-align: left; }
  th { padding: 12px 16px; background: rgba(255, 255, 255, 0.03); color: var(--text-muted); font-size: 11px;
    text-transform: uppercase; border-bottom: 1px solid var(--border-primary); }
  td { padding: 12px 16px; border-bottom: 1px solid var(--border-primary); color: var(--text-secondary); }
  .cover-cell { width: 56px; }
  .cover-cell img { width: 40px; height: 56px; object-fit: cover; border-radius: 4px; }
  .t-title { font-weight: 600; color: var(--text-primary); }
  .t-author { font-size: 11px; color: var(--text-muted); }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .cfg { padding: 24px; border-radius: var(--radius-lg); display: flex; flex-direction: column; gap: 14px; }
  .cfg h3 { font-size: 16px; color: var(--text-primary); }
  .cfg-item { padding: 14px; border: 1px solid var(--border-secondary); border-radius: var(--radius-md); background: rgba(0, 0, 0, 0.15); }
  .cfg-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; }
  .desc { font-size: 12px; color: var(--text-secondary); margin-top: 4px; }
  .status { margin-left: 8px; font-size: 11px; color: var(--text-muted); }
  .status.on { color: #22c55e; }
  .cfg-form { margin-top: 12px; display: flex; flex-direction: column; gap: 10px; max-width: 520px; }
  .cfg-form label { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--text-secondary); }
  .cfg-form label.chk { flex-direction: row; align-items: center; gap: 8px; }
  .cfg-actions { display: flex; gap: 8px; }
  .loading-state { text-align: center; padding: 40px; display: flex; flex-direction: column; align-items: center; gap: 12px; color: var(--text-secondary); }
  .spinner { width: 32px; height: 32px; border: 3px solid rgba(56, 189, 248, 0.15); border-top-color: var(--accent);
    border-radius: 50%; animation: spin 0.8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>

