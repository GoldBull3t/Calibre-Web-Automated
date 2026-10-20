import { writable } from 'svelte/store';

export const isOnline = writable(typeof navigator !== 'undefined' ? navigator.onLine : true);
export const canInstall = writable(false);
export const offlineBookIds = writable(new Set());

let deferredInstallPrompt = null;

// Inicializa Service Worker e escuta de estado de rede
export function initPWA() {
  if (typeof window === 'undefined') return;

  // 1. Escuta Online / Offline
  window.addEventListener('online', () => isOnline.set(true));
  window.addEventListener('offline', () => isOnline.set(false));

  // 2. Evento de Instalação PWA (BeforeInstallPrompt)
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    deferredInstallPrompt = e;
    canInstall.set(true);
  });

  window.addEventListener('appinstalled', () => {
    canInstall.set(false);
    deferredInstallPrompt = null;
  });

  // 3. Registo do Service Worker
  if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
      navigator.serviceWorker.register('/sw.js', { scope: '/' })
        .then(reg => {
          // Verifica se há livros guardados offline
          refreshOfflineBookIds();
        })
        .catch(err => {
          console.warn('[PWA] Falha ao registar Service Worker:', err);
        });
    });

    // Escuta mensagens vindas do SW
    navigator.serviceWorker.addEventListener('message', (event) => {
      const data = event.data;
      if (!data) return;

      if (data.type === 'ALL_OFFLINE_BOOKS' && Array.isArray(data.bookIds)) {
        offlineBookIds.set(new Set(data.bookIds));
      }

      if (data.type === 'BOOK_OFFLINE_SAVED' && data.success) {
        offlineBookIds.update(set => {
          const next = new Set(set);
          next.add(parseInt(data.bookId, 10));
          return next;
        });
      }

      if (data.type === 'BOOK_OFFLINE_REMOVED' && data.success) {
        offlineBookIds.update(set => {
          const next = new Set(set);
          next.delete(parseInt(data.bookId, 10));
          return next;
        });
      }
    });
  }
}

// Dispara o diálogo nativo de instalação PWA
export async function promptInstallPWA() {
  if (!deferredInstallPrompt) return false;
  deferredInstallPrompt.prompt();
  const choice = await deferredInstallPrompt.userChoice;
  deferredInstallPrompt = null;
  canInstall.set(false);
  return choice.outcome === 'accepted';
}

// Pede ao SW para guardar o livro em cache offline (EPUB, Leitor, Capa)
export async function saveBookOffline(bookId) {
  if (!('serviceWorker' in navigator) || !navigator.serviceWorker.controller) {
    // Fallback: se controller não estiver pronto, tenta buscar para o Cache Storage diretamente
    if ('caches' in window) {
      const cache = await caches.open('cwa-books-v1');
      await Promise.all([
        cache.add(`/show/${bookId}/epub/book.epub`),
        cache.add(`/read/${bookId}`)
      ]);
      offlineBookIds.update(s => new Set(s).add(parseInt(bookId, 10)));
      return true;
    }
    return false;
  }

  navigator.serviceWorker.controller.postMessage({
    type: 'SAVE_BOOK_OFFLINE',
    bookId: bookId
  });
  return true;
}

// Remove o livro da cache offline
export async function removeBookOffline(bookId) {
  if (!('serviceWorker' in navigator) || !navigator.serviceWorker.controller) {
    if ('caches' in window) {
      const cache = await caches.open('cwa-books-v1');
      await cache.delete(`/show/${bookId}/epub/book.epub`);
      await cache.delete(`/read/${bookId}`);
      offlineBookIds.update(s => {
        const next = new Set(s);
        next.delete(parseInt(bookId, 10));
        return next;
      });
      return true;
    }
    return false;
  }

  navigator.serviceWorker.controller.postMessage({
    type: 'REMOVE_BOOK_OFFLINE',
    bookId: bookId
  });
  return true;
}

// Atualiza a lista de livros disponíveis offline
export function refreshOfflineBookIds() {
  if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
    navigator.serviceWorker.controller.postMessage({
      type: 'GET_ALL_OFFLINE_BOOKS'
    });
  } else if ('caches' in window) {
    caches.open('cwa-books-v1').then(cache => {
      cache.keys().then(requests => {
        const ids = new Set();
        requests.forEach(req => {
          const url = new URL(req.url);
          const match = url.pathname.match(/\/show\/(\d+)\/epub/);
          if (match && match[1]) {
            ids.add(parseInt(match[1], 10));
          }
        });
        offlineBookIds.set(ids);
      });
    }).catch(() => {});
  }
}

