// Calibre-Web Automated NextGen - Progressive Web App Service Worker
const CACHE_STATIC = 'cwa-static-v2026-03';
const CACHE_COVERS = 'cwa-covers-v1';
const CACHE_BOOKS = 'cwa-books-v1';
const CACHE_API = 'cwa-api-v1';

const STATIC_SHELL = [
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
              .map(k => {
                console.log('[SW] Apagando cache legada:', k);
                return caches.delete(k);
              })
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

  // 4. Navegações gerais SPA ('/' e req.mode === 'navigate') -> Network-First (SEMPRE obtém a versão mais recente quando online)
  if (req.mode === 'navigate' || url.pathname === '/' || url.pathname === '/index.html') {
    event.respondWith(
      fetch(req).then(async networkResp => {
        if (networkResp && networkResp.status === 200) {
          const cache = await caches.open(CACHE_STATIC);
          cache.put(req, networkResp.clone());
        }
        return networkResp;
      }).catch(async () => {
        const cache = await caches.open(CACHE_STATIC);
        const cached = await cache.match(req) || await cache.match('/');
        if (cached) return cached;
        return new Response('<h1>Aplicação Offline</h1><p>Esta página não está em cache.</p>', {
          headers: { 'Content-Type': 'text/html; charset=utf-8' }
        });
      })
    );
    return;
  }

  // 5. Assets estáticos com hash (/assets/*) e bibliotecas -> Cache-First
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
});

// Mensagens vindas da interface (Salvar Livro Offline, Verificar, etc.)
self.addEventListener('message', async event => {
  const data = event.data;
  if (!data || !data.type) return;

  if (data.type === 'CACHE_BOOK') {
    const { bookId, format } = data;
    try {
      const bookCache = await caches.open(CACHE_BOOKS);
      const coverCache = await caches.open(CACHE_COVERS);

      const bookUrl = `/read/${bookId}`;
      const coverUrl = `/cover/${bookId}`;

      const [bookRes, coverRes] = await Promise.all([
        fetch(bookUrl),
        fetch(coverUrl)
      ]);

      if (bookRes.ok) await bookCache.put(bookUrl, bookRes);
      if (coverRes.ok) await coverCache.put(coverUrl, coverRes);

      event.source.postMessage({
        type: 'BOOK_CACHED_SUCCESS',
        bookId
      });
    } catch (err) {
      event.source.postMessage({
        type: 'BOOK_CACHED_ERROR',
        bookId,
        error: err.message
      });
    }
  }

  if (data.type === 'REMOVE_CACHED_BOOK') {
    const { bookId } = data;
    const bookCache = await caches.open(CACHE_BOOKS);
    const coverCache = await caches.open(CACHE_COVERS);

    await Promise.all([
      bookCache.delete(`/read/${bookId}`),
      coverCache.delete(`/cover/${bookId}`)
    ]);

    event.source.postMessage({
      type: 'BOOK_REMOVED_SUCCESS',
      bookId
    });
  }

  if (data.type === 'GET_ALL_OFFLINE_BOOKS') {
    const bookCache = await caches.open(CACHE_BOOKS);
    const keys = await bookCache.keys();
    const bookIds = [];

    keys.forEach(k => {
      const u = new URL(k.url);
      const parts = u.pathname.split('/');
      if (parts[1] === 'read' && parts[2]) {
        const idNum = parseInt(parts[2], 10);
        if (!isNaN(idNum)) bookIds.push(idNum);
      }
    });

    event.source.postMessage({
      type: 'ALL_OFFLINE_BOOKS',
      bookIds
    });
  }
});
