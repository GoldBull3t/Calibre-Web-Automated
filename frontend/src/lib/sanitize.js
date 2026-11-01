// @ts-nocheck
/**
 * Sanitiza HTML para permitir formatação segura de sinopses e descrições de livros.
 * Remove tags ativas e inseguras (<script>, <iframe>, <object>, etc.) e atributos de eventos (onclick, etc.).
 */
export function sanitizeHtml(html) {
  if (!html || typeof html !== 'string') return '';
  if (typeof DOMParser === 'undefined') return html;

  try {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');

    const disallowedTags = [
      'script', 'style', 'iframe', 'frame', 'object', 'embed',
      'form', 'input', 'button', 'select', 'textarea', 'link', 'meta', 'base'
    ];

    disallowedTags.forEach(tag => {
      const elements = doc.querySelectorAll(tag);
      elements.forEach(el => el.remove());
    });

    const allElements = doc.querySelectorAll('*');
    allElements.forEach(el => {
      for (let i = el.attributes.length - 1; i >= 0; i--) {
        const attr = el.attributes[i];
        const name = attr.name.toLowerCase();
        const value = attr.value.trim().toLowerCase();

        // Remove handlers inline on*
        if (name.startsWith('on')) {
          el.removeAttribute(attr.name);
        }
        // Remove esquemas javascript:, vbscript: e data: não-imagem
        else if (['href', 'src', 'action', 'formaction', 'xlink:href'].includes(name)) {
          if (value.startsWith('javascript:') || value.startsWith('vbscript:') || (value.startsWith('data:') && !value.startsWith('data:image/'))) {
            el.removeAttribute(attr.name);
          }
        }
      }
    });

    return doc.body.innerHTML;
  } catch (e) {
    console.error('Falha ao sanitizar HTML:', e);
    return '';
  }
}

/**
 * Sanitiza snippets de pesquisa Full-Text Search garantindo que APENAS
 * as tags de destaque <mark> e </mark> sejam mantidas, escapando todo o restante.
 */
export function sanitizeSnippet(snippet) {
  if (!snippet || typeof snippet !== 'string') return '';

  const MARK_OPEN = '___MARK_OPEN___';
  const MARK_CLOSE = '___MARK_CLOSE___';

  // Substitui mark válidos por sentinelas
  let safe = snippet
    .replace(/<mark>/gi, MARK_OPEN)
    .replace(/<\/mark>/gi, MARK_CLOSE);

  // Escapa entidades HTML no navegador
  if (typeof document !== 'undefined') {
    const div = document.createElement('div');
    div.textContent = safe;
    safe = div.innerHTML;
  } else {
    safe = safe
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // Restaura apenas os <mark> sem atributos
  return safe
    .replaceAll(MARK_OPEN, '<mark>')
    .replaceAll(MARK_CLOSE, '</mark>');
}

