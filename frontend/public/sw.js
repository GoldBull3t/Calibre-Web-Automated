// Calibre-Web Automated NextGen - Progressive Web App Service Worker
const CACHE_STATIC = 'cwa-static-v2';
const CACHE_COVERS = 'cwa-covers-v1';
const CACHE_BOOKS = 'cwa-books-v1';
const CACHE_API = 'cwa-api-v1';

const STATIC_SHELL = [
  '/',
  '/manifest.webmanifest',
  '/manifest.json',
  '/favicon.svg',
  '/static/js/libs/epub.min.js',
  '/static/js/compress/jszip_epub.min.js',
  '/static/js/libs/jquery.min.js'
];

self.addEventListener('install', event => {
  self.skipWaiting();
  event.waitUntil(
    caches.open(CACHE_STATIC).then(cache => {
      return cache.addAll(STATIC_SHELL).catch(err => {
        console.warn('[SW] Falha parcial no pré-cache estático:', err);
      });
    })
  );
});

self.addEventListener('activate', event => {
  event.waitUntil(
    Promise.all([
      self.clients.claim(),
      caches.keys().then(keys => {
        return Promise.all(
          keys.filter(k => ![CACHE_STATIC, CACHE_COVERS, CACHE_BOOKS, CACHE_API].includes(k))
              .map(k => caches.delete(k))
        );
      })
    ])
  );
});

self.addEventListener('fetch', event => {
  const req = event.request;
  const url = new URL(req.url);

  // Apenas gere requisições GET do mesmo host
  if (req.method !== 'GET' || url.origin !== self.location.origin) {
    return;
  }

  // 1. Capas (/cover/*) -> Cache-First com atualização em background
  if (url.pathname.startsWith('/cover/')) {
    event.respondWith(
      caches.open(CACHE_COVERS).then(async cache => {
        const cached = await cache.match(req);
        if (cached) return cached;
        try {
          const networkResp = await fetch(req);
          if (networkResp && networkResp.status === 200) {
            cache.put(req, networkResp.clone());
          }
          return networkResp;
        } catch (err) {
          return new Response('', { status: 404, statusText: 'Offline cover not found' });
        }
      })
    );
    return;
  }

  // 2. Livro EPUB (/show/*) e Leitor (/read/*) -> Cache Storage para leitura 100% offline
  if (url.pathname.startsWith('/show/') || url.pathname.startsWith('/read/')) {
    event.respondWith(
      caches.open(CACHE_BOOKS).then(async cache => {
        const cached = await cache.match(req);
        if (cached) {
          // Serve do cache imediatamente e atualiza em background se houver rede
          fetch(req).then(networkResp => {
            if (networkResp && networkResp.status === 200) {
              cache.put(req, networkResp);
            }
          }).catch(() => {});
          return cached;
        }
        try {
          const networkResp = await fetch(req);
          if (networkResp && networkResp.status === 200) {
            cache.put(req, networkResp.clone());
          }
          return networkResp;
        } catch (err) {
          if (cached) return cached;
          throw err;
        }
      })
    );
    return;
  }

  // 3. API de Livros (/api/books) -> Network-First com Fallback de Cache
  if (url.pathname === '/api/books') {
    event.respondWith(
      fetch(req).then(async networkResp => {
        if (networkResp && networkResp.status === 200) {
          const cache = await caches.open(CACHE_API);
          cache.put(req, networkResp.clone());
        }
        return networkResp;
      }).catch(async () => {
        const cache = await caches.open(CACHE_API);
        const cached = await cache.match(req);
        if (cached) return cached;
        return new Response(JSON.stringify({ books: [], offline: true }), {
          headers: { 'Content-Type': 'application/json' }
        });
      })
    );
    return;
  }

  // 4. Assets compilados (/assets/*) -> Cache-First
  if (url.pathname.startsWith('/assets/') || STATIC_SHELL.includes(url.pathname)) {
    event.respondWith(
      caches.open(CACHE_STATIC).then(async cache => {
        const cached = await cache.match(req);
        if (cached) return cached;
        try {
          const networkResp = await fetch(req);
          if (networkResp && networkResp.status === 200) {
            cache.put(req, networkResp.clone());
          }
          return networkResp;
        } catch (err) {
          return cached || new Response('Offline asset unavailable', { status: 503 });
        }
      })
    );
    return;
  }

  // 5. Navegações gerais SPA -> Network-First com fallback para index.html em cache
  if (req.mode === 'navigate') {
    event.respondWith(
      fetch(req).catch(async () => {
        const cache = await caches.open(CACHE_STATIC);
        const shell = await cache.match('/');
        if (shell) return shell;
        return new Response('<h1>Aplicação Offline</h1><p>Esta página não está em cache.</p>', {
          headers: { 'Content-Type': 'text/html; charset=utf-8' }
        });
      })
    );
  }
});

// Mensagens vindas da interface (Salvar Livro Offline, Verificar, etc.)
self.addEventListener('message', async event => {
  const data = event.data;
  if (!data || !data.type) return;

  const bookId = data.bookId;

  if (data.type === 'SAVE_BOOK_OFFLINE' && bookId) {
    try {
      const bookCache = await caches.open(CACHE_BOOKS);
      const coverCache = await caches.open(CACHE_COVERS);

      const epubURL = `/show/${bookId}/epub/book.epub`;
      const readerURL = `/read/${bookId}`;
      const coverURL = `/cover/${bookId}`;

      // Descarrega os 3 recursos em paralelo
      await Promise.all([
        fetch(epubURL).then(res => { if (res.ok) bookCache.put(epubURL, res); }),
        fetch(readerURL).then(res => { if (res.ok) bookCache.put(readerURL, res); }),
        fetch(coverURL).then(res => { if (res.ok) coverCache.put(coverURL, res); })
      ]);

      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_SAVED', bookId: bookId, success: true });
      }
    } catch (err) {
      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_SAVED', bookId: bookId, success: false, error: err.message });
      }
    }
  }

  if (data.type === 'REMOVE_BOOK_OFFLINE' && bookId) {
    try {
      const bookCache = await caches.open(CACHE_BOOKS);
      await bookCache.delete(`/show/${bookId}/epub/book.epub`);
      await bookCache.delete(`/read/${bookId}`);
      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_REMOVED', bookId: bookId, success: true });
      }
    } catch (err) {
      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_REMOVED', bookId: bookId, success: false });
      }
    }
  }

  if (data.type === 'CHECK_BOOK_OFFLINE' && bookId) {
    try {
      const bookCache = await caches.open(CACHE_BOOKS);
      const matched = await bookCache.match(`/show/${bookId}/epub/book.epub`);
      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_STATUS', bookId: bookId, isOffline: !!matched });
      }
    } catch (err) {
      if (event.source) {
        event.source.postMessage({ type: 'BOOK_OFFLINE_STATUS', bookId: bookId, isOffline: false });
      }
    }
  }

  if (data.type === 'GET_ALL_OFFLINE_BOOKS') {
    try {
      const bookCache = await caches.open(CACHE_BOOKS);
      const requests = await bookCache.keys();
      const ids = new Set();
      requests.forEach(req => {
        const url = new URL(req.url);
        const match = url.pathname.match(/\/show\/(\d+)\/epub/);
        if (match && match[1]) {
          ids.add(parseInt(match[1], 10));
        }
      });
      if (event.source) {
        event.source.postMessage({ type: 'ALL_OFFLINE_BOOKS', bookIds: Array.from(ids) });
      }
    } catch (err) {
      if (event.source) {
        event.source.postMessage({ type: 'ALL_OFFLINE_BOOKS', bookIds: [] });
      }
    }
  }
});

