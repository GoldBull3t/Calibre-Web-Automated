package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/crocodilestick/calibre-web-go/internal/formats"
)

// getComicFilePath localiza o ficheiro CBZ ou CBR do livro especificado.
func (h *Handler) getComicFilePath(r *http.Request, bookID int64) (string, string, error) {
	book, err := h.db.GetBookByID(r.Context(), bookID)
	if err != nil || book == nil {
		return "", "", fmt.Errorf("livro não encontrado")
	}

	for _, f := range book.Formats {
		if strings.EqualFold(f.Format, "CBZ") || strings.EqualFold(f.Format, "CBR") {
			fileName := fmt.Sprintf("%s.%s", f.Name, strings.ToLower(f.Format))
			fullPath := filepath.Join(h.booksDir, filepath.FromSlash(book.Path), fileName)
			if _, statErr := os.Stat(fullPath); statErr == nil {
				return fullPath, strings.ToUpper(f.Format), nil
			}
		}
	}

	return "", "", fmt.Errorf("nenhum formato de banda desenhada (CBZ/CBR) encontrado para este livro")
}

// APIComicPages devolve a lista de páginas de uma banda desenhada / mangá.
func (h *Handler) APIComicPages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"ID inválido"}`, http.StatusBadRequest)
		return
	}

	comicPath, fmtName, err := h.getComicFilePath(r, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	info, err := formats.ListComicPages(comicPath)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	type PageItem struct {
		Index int    `json:"index"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	}

	pages := make([]PageItem, len(info.Pages))
	for i, p := range info.Pages {
		pages[i] = PageItem{
			Index: i,
			Name:  filepath.Base(p),
			URL:   fmt.Sprintf("/api/comic/%d/page/%d", id, i),
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"book_id":     id,
		"format":      fmtName,
		"total_pages": info.TotalPages,
		"pages":       pages,
	})
}

// APIComicPageImage entrega a imagem de uma página específica de um arquivo CBZ/CBR.
func (h *Handler) APIComicPageImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	pageStr := r.PathValue("page")
	pageIdx, err := strconv.Atoi(pageStr)
	if err != nil || pageIdx < 0 {
		http.NotFound(w, r)
		return
	}

	comicPath, _, err := h.getComicFilePath(r, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data, mimeType, err := formats.ExtractComicPage(comicPath, pageIdx)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ComicReaderWeb serve a página HTML do leitor especializado de Banda Desenhada e Mangá.
func (h *Handler) ComicReaderWeb(w http.ResponseWriter, r *http.Request, bookID int64) {
	book, err := h.db.GetBookByID(r.Context(), bookID)
	if err != nil || book == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, comicReaderHTML, book.Title, book.Title, strings.Join(book.Authors, ", "), book.ID)
}

const comicReaderHTML = `<!DOCTYPE html>
<html lang="pt">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>%s | Leitor de Banda Desenhada & Mangá</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body, html {
      width: 100%%; height: 100%%;
      background: #090e15;
      color: #f1f5f9;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      overflow: hidden;
      user-select: none;
    }

    /* Top HUD */
    .hud {
      position: fixed;
      left: 0; right: 0;
      background: rgba(15, 23, 42, 0.88);
      backdrop-filter: blur(12px);
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      z-index: 100;
      transition: transform 0.25s ease, opacity 0.25s ease;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 20px;
    }
    .hud-top { top: 0; }
    .hud-bottom {
      top: auto;
      bottom: 0;
      border-bottom: none;
      border-top: 1px solid rgba(255, 255, 255, 0.08);
      flex-direction: column;
      gap: 8px;
      padding: 12px 24px;
    }
    .hud.hidden {
      opacity: 0;
      pointer-events: none;
      transform: translateY(-100%%);
    }
    .hud-bottom.hidden {
      transform: translateY(100%%);
    }

    .hud-title {
      font-size: 14px;
      font-weight: 600;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 40%%;
    }
    .hud-sub {
      font-size: 11px;
      color: #94a3b8;
      display: block;
    }

    .btn-group {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .c-btn {
      background: rgba(255, 255, 255, 0.06);
      border: 1px solid rgba(255, 255, 255, 0.12);
      color: #e2e8f0;
      padding: 6px 12px;
      border-radius: 6px;
      font-size: 12px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 6px;
      transition: all 0.15s;
    }
    .c-btn:hover {
      background: rgba(56, 189, 248, 0.15);
      border-color: #38bdf8;
      color: #38bdf8;
    }
    .c-btn.active {
      background: #0284c7;
      border-color: #38bdf8;
      color: #fff;
    }

    /* Scrubber */
    .scrubber-row {
      display: flex;
      align-items: center;
      gap: 14px;
      width: 100%%;
      max-width: 800px;
      margin: 0 auto;
    }
    .scrubber-row input[type="range"] {
      flex: 1;
      accent-color: #38bdf8;
      cursor: pointer;
    }
    .page-tag {
      font-size: 12px;
      font-weight: 600;
      color: #38bdf8;
      min-width: 80px;
      text-align: center;
    }

    /* Main Viewer Area */
    #viewer-stage {
      width: 100%%;
      height: 100%%;
      display: flex;
      align-items: center;
      justify-content: center;
      position: relative;
      overflow: hidden;
    }

    /* Single / Double Page Mode */
    .paged-view {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%%;
      height: 100%%;
      gap: 8px;
    }
    .paged-view img {
      max-height: 100%%;
      max-width: 100%%;
      object-fit: contain;
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
      border-radius: 4px;
    }

    /* Fit Modes */
    .fit-height img { max-height: 100%%; max-width: auto; }
    .fit-width img { max-width: 100%%; max-height: auto; }
    .fit-screen img { max-width: 100%%; max-height: 100%%; }

    /* Webtoon / Vertical Continuous Scroll Mode */
    .webtoon-view {
      width: 100%%;
      height: 100%%;
      overflow-y: auto;
      overflow-x: hidden;
      display: flex;
      flex-direction: column;
      align-items: center;
      scroll-behavior: smooth;
    }
    .webtoon-view img {
      width: 100%%;
      max-width: 900px;
      height: auto;
      display: block;
      margin: 0 auto;
    }

    /* Hotzones */
    .zone {
      position: absolute;
      top: 0; bottom: 0;
      width: 25%%;
      z-index: 50;
      cursor: pointer;
    }
    .zone-left { left: 0; }
    .zone-right { right: 0; }
    .zone-mid { left: 25%%; right: 25%%; width: 50%%; }

    /* Spinner */
    .loader {
      position: absolute;
      width: 40px; height: 40px;
      border: 3px solid rgba(56, 189, 248, 0.2);
      border-top-color: #38bdf8;
      border-radius: 50%%;
      animation: spin 0.8s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
  </style>
</head>
<body>
  <!-- Top HUD -->
  <header class="hud hud-top" id="hud-top">
    <div>
      <div class="hud-title">%s</div>
      <span class="hud-sub">%s</span>
    </div>

    <div class="btn-group">
      <!-- Mode buttons -->
      <button class="c-btn active" id="btn-mode-single" title="Página Única">1 Pág</button>
      <button class="c-btn" id="btn-mode-double" title="Página Dupla (Spread)">2 Págs</button>
      <button class="c-btn" id="btn-mode-webtoon" title="Webtoon (Scroll Contínuo)">Webtoon</button>

      <!-- Direction button -->
      <button class="c-btn" id="btn-dir" title="Alternar Sentido de Leitura (LTR / RTL)">LTR</button>

      <!-- Fit button -->
      <button class="c-btn" id="btn-fit" title="Ajustar Altura / Largura">Ajustar</button>

      <!-- Fullscreen button -->
      <button class="c-btn" id="btn-fs" title="Ecrã Inteiro (F)">⛶</button>
      <button class="c-btn" onclick="window.close()" title="Fechar Leitor">✕</button>
    </div>
  </header>

  <!-- Viewer Stage -->
  <div id="viewer-stage">
    <div class="loader" id="loader"></div>
    <div class="zone zone-left" id="zone-left"></div>
    <div class="zone zone-mid" id="zone-mid"></div>
    <div class="zone zone-right" id="zone-right"></div>
    <div id="content-wrap" class="paged-view fit-height"></div>
  </div>

  <!-- Bottom HUD -->
  <footer class="hud hud-bottom" id="hud-bottom">
    <div class="scrubber-row">
      <button class="c-btn" id="btn-prev">Anterior</button>
      <input type="range" id="page-slider" min="1" max="1" value="1">
      <span class="page-tag" id="page-display">1 / 1</span>
      <button class="c-btn" id="btn-next">Seguinte</button>
    </div>
  </footer>

  <script>
    const bookId = %d;
    let totalPages = 0;
    let pages = [];
    let currentPage = 1;
    let viewMode = 'single'; // 'single', 'double', 'webtoon'
    let direction = 'ltr';   // 'ltr' (ocidental), 'rtl' (mangá)
    let fitMode = 'height';  // 'height', 'width', 'screen'
    let hudVisible = true;
    let hideTimer = null;

    const stage = document.getElementById('viewer-stage');
    const content = document.getElementById('content-wrap');
    const loader = document.getElementById('loader');
    const slider = document.getElementById('page-slider');
    const pageDisplay = document.getElementById('page-display');
    const hudTop = document.getElementById('hud-top');
    const hudBottom = document.getElementById('hud-bottom');

    async function init() {
      try {
        const res = await fetch('/api/comic/' + bookId + '/pages');
        const data = await res.json();
        pages = data.pages || [];
        totalPages = pages.length;

        if (totalPages === 0) {
          loader.style.display = 'none';
          content.innerHTML = '<p style="color: #ef4444;">Nenhuma página encontrada no arquivo de banda desenhada.</p>';
          return;
        }

        slider.max = totalPages;
        loader.style.display = 'none';

        // Carrega progresso salvo
        try {
          const progRes = await fetch('/api/reader/progress/' + bookId);
          if (progRes.ok) {
            const pData = await progRes.json();
            if (pData.progress && parseInt(pData.progress) > 0) {
              currentPage = Math.min(parseInt(pData.progress), totalPages);
            }
          }
        } catch(e) {}

        render();
        resetHudTimer();
      } catch (err) {
        loader.style.display = 'none';
        content.innerHTML = '<p style="color: #ef4444;">Erro ao carregar páginas: ' + err.message + '</p>';
      }
    }

    function render() {
      slider.value = currentPage;
      pageDisplay.innerText = currentPage + ' / ' + totalPages;

      if (viewMode === 'webtoon') {
        content.className = 'webtoon-view';
        content.innerHTML = pages.map((p, idx) => 
          '<img src="' + p.url + '" alt="Pág ' + (idx + 1) + '" id="webtoon-p-' + (idx + 1) + '" loading="lazy" />'
        ).join('');

        // Rola até a página atual
        setTimeout(() => {
          const target = document.getElementById('webtoon-p-' + currentPage);
          if (target) target.scrollIntoView();
        }, 100);
        return;
      }

      content.className = 'paged-view fit-' + fitMode;
      const idx = currentPage - 1;

      if (viewMode === 'double' && idx + 1 < totalPages) {
        const first = direction === 'rtl' ? pages[idx + 1] : pages[idx];
        const second = direction === 'rtl' ? pages[idx] : pages[idx + 1];
        content.innerHTML = 
          '<img src="' + first.url + '" alt="Página ' + (idx + 1) + '" />' +
          '<img src="' + second.url + '" alt="Página ' + (idx + 2) + '" />';
      } else {
        content.innerHTML = '<img src="' + pages[idx].url + '" alt="Página ' + (idx + 1) + '" />';
      }

      saveProgress();
    }

    function saveProgress() {
      fetch('/api/reader/progress/' + bookId, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          progress: currentPage.toString(),
          percentage: (currentPage / totalPages) * 100
        })
      }).catch(() => {});
    }

    function nextPage() {
      const step = viewMode === 'double' ? 2 : 1;
      if (currentPage + step <= totalPages) {
        currentPage += step;
        render();
      } else if (currentPage < totalPages) {
        currentPage = totalPages;
        render();
      }
    }

    function prevPage() {
      const step = viewMode === 'double' ? 2 : 1;
      if (currentPage - step >= 1) {
        currentPage -= step;
        render();
      } else if (currentPage > 1) {
        currentPage = 1;
        render();
      }
    }

    function stepForward() {
      if (direction === 'rtl') prevPage();
      else nextPage();
    }

    function stepBackward() {
      if (direction === 'rtl') nextPage();
      else prevPage();
    }

    function toggleHud() {
      hudVisible = !hudVisible;
      hudTop.classList.toggle('hidden', !hudVisible);
      hudBottom.classList.toggle('hidden', !hudVisible);
    }

    function resetHudTimer() {
      hudVisible = true;
      hudTop.classList.remove('hidden');
      hudBottom.classList.remove('hidden');
      clearTimeout(hideTimer);
      hideTimer = setTimeout(() => {
        hudVisible = false;
        hudTop.classList.add('hidden');
        hudBottom.classList.add('hidden');
      }, 4000);
    }

    // Eventos de clique nas zonas
    document.getElementById('zone-left').onclick = (e) => {
      e.stopPropagation();
      stepBackward();
    };
    document.getElementById('zone-right').onclick = (e) => {
      e.stopPropagation();
      stepForward();
    };
    document.getElementById('zone-mid').onclick = (e) => {
      e.stopPropagation();
      toggleHud();
    };

    document.onmousemove = resetHudTimer;

    // Botões HUD
    document.getElementById('btn-prev').onclick = stepBackward;
    document.getElementById('btn-next').onclick = stepForward;

    slider.oninput = (e) => {
      currentPage = parseInt(e.target.value);
      render();
    };

    // Modos de visualização
    document.getElementById('btn-mode-single').onclick = () => {
      viewMode = 'single';
      updateModeButtons('single');
      render();
    };
    document.getElementById('btn-mode-double').onclick = () => {
      viewMode = 'double';
      updateModeButtons('double');
      render();
    };
    document.getElementById('btn-mode-webtoon').onclick = () => {
      viewMode = 'webtoon';
      updateModeButtons('webtoon');
      render();
    };

    function updateModeButtons(active) {
      document.getElementById('btn-mode-single').classList.toggle('active', active === 'single');
      document.getElementById('btn-mode-double').classList.toggle('active', active === 'double');
      document.getElementById('btn-mode-webtoon').classList.toggle('active', active === 'webtoon');
    }

    // Sentido de leitura
    const btnDir = document.getElementById('btn-dir');
    btnDir.onclick = () => {
      direction = direction === 'ltr' ? 'rtl' : 'ltr';
      btnDir.innerText = direction.toUpperCase();
      btnDir.title = direction === 'rtl' ? 'Mangá (Direita para a Esquerda)' : 'Ocidental (Esquerda para a Direita)';
      render();
    };

    // Ajuste de ecrã
    const btnFit = document.getElementById('btn-fit');
    btnFit.onclick = () => {
      if (fitMode === 'height') {
        fitMode = 'width';
        btnFit.innerText = 'Largura';
      } else if (fitMode === 'width') {
        fitMode = 'screen';
        btnFit.innerText = 'Ecrã';
      } else {
        fitMode = 'height';
        btnFit.innerText = 'Altura';
      }
      render();
    };

    // Ecrã inteiro
    document.getElementById('btn-fs').onclick = () => {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().catch(() => {});
      } else {
        document.exitFullscreen().catch(() => {});
      }
    };

    // Teclado
    window.onkeydown = (e) => {
      resetHudTimer();
      if (e.key === 'ArrowRight' || e.key === 'PageDown' || e.key === ' ') {
        stepForward();
      } else if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
        stepBackward();
      } else if (e.key === 'f' || e.key === 'F') {
        document.getElementById('btn-fs').click();
      } else if (e.key === 'Escape') {
        if (!hudVisible) toggleHud();
      }
    };

    init();
  </script>
</body>
</html>
`
