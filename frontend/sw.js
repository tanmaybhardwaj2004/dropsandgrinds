const CACHE_NAME = 'dropsandgrinds-v2';
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/css/index.css',
  '/css/search.css',
  '/js/app.js',
  '/js/search.js',
  '/js/auth.js',
  '/images/icon-192x192.png',
  '/images/icon-512x512.png'
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(STATIC_ASSETS);
    })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

// Fetch event - keep API and app shell fresh during active development/deploys.
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  if (url.origin !== self.location.origin || url.pathname.startsWith('/api/') || url.pathname.startsWith('/health') || url.pathname.startsWith('/metrics')) {
    return;
  }

  event.respondWith(
    fetch(event.request).then((networkResponse) => {
        if (networkResponse.ok) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, networkResponse.clone());
          });
        }
        return networkResponse;
    }).catch(() => {
      return caches.match(event.request);
    })
  );
});
