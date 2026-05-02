const CACHE_NAME = 'dropsandgrinds-v2';
const STATIC_CACHE_NAME = 'static-v2';
const API_CACHE_NAME = 'api-v2';
const IMAGE_CACHE_NAME = 'images-v2';

const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/search.html',
  '/game.html',
  '/library.html',
  '/wishlist.html',
  '/savings.html',
  '/bundle.html',
  '/buy-timing.html',
  '/about.html',
  '/login.html',
  '/register.html',
  '/css/modern-theme.css',
  '/css/index.css',
  '/css/search.css',
  '/css/game.css',
  '/css/auth-modern.css',
  '/js/app.js',
  '/js/search.js',
  '/js/game.js',
  '/js/wishlist.js',
  '/js/auth.js',
  '/js/navbar.js',
  '/images/icon-192x192.png',
  '/images/icon-512x512.png'
];

// API endpoints that support offline caching
const CACHEABLE_API_ENDPOINTS = [
  '/api/games/search',
  '/api/games/',
  '/api/prices/',
  '/api/deals',
  '/api/sales/active'
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  console.log('[SW] Installing...');
  event.waitUntil(
    caches.open(STATIC_CACHE_NAME)
      .then((cache) => {
        console.log('[SW] Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .then(() => self.skipWaiting())
      .catch((err) => console.error('[SW] Install failed:', err))
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating...');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== STATIC_CACHE_NAME && 
              cacheName !== API_CACHE_NAME && 
              cacheName !== IMAGE_CACHE_NAME) {
            console.log('[SW] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim())
  );
});

// Helper: Check if URL is a cacheable API endpoint
function isCacheableAPI(url) {
  return CACHEABLE_API_ENDPOINTS.some(endpoint => url.includes(endpoint));
}

// Helper: Check if request is an image
function isImageRequest(request) {
  return request.destination === 'image' || 
         request.url.match(/\.(jpg|jpeg|png|gif|webp|svg|ico)$/i);
}

// Helper: Check if request is for static asset
function isStaticAsset(request) {
  return request.destination === 'style' || 
         request.destination === 'script' ||
         request.destination === 'font' ||
         request.destination === 'document';
}

// Fetch event with advanced caching strategies
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests for caching (but allow them through)
  if (request.method !== 'GET') {
    return;
  }

  // Handle API requests with Network-First strategy
  if (url.pathname.includes('/api/')) {
    if (isCacheableAPI(url.pathname)) {
      event.respondWith(networkFirstStrategy(request));
    }
    return;
  }

  // Handle image requests with Cache-First strategy
  if (isImageRequest(request)) {
    event.respondWith(cacheFirstStrategy(request, IMAGE_CACHE_NAME));
    return;
  }

  // Handle static assets with Cache-First strategy
  if (isStaticAsset(request)) {
    event.respondWith(cacheFirstStrategy(request, STATIC_CACHE_NAME));
    return;
  }

  // Default: Network with cache fallback
  event.respondWith(networkFirstStrategy(request));
});

// Network-First Strategy: Try network first, fall back to cache
async function networkFirstStrategy(request) {
  const cache = await caches.open(API_CACHE_NAME);
  
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      // Update cache with fresh data
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[SW] Network failed, trying cache:', request.url);
    const cachedResponse = await caches.match(request);
    
    if (cachedResponse) {
      console.log('[SW] Serving from cache:', request.url);
      // Add header to indicate offline mode
      const headers = new Headers(cachedResponse.headers);
      headers.set('X-SW-Offline', 'true');
      
      return new Response(cachedResponse.body, {
        status: 200,
        statusText: 'OK (Offline)',
        headers: headers
      });
    }
    
    // No cache available
    return new Response(
      JSON.stringify({ error: 'Offline - No cached data available' }),
      { 
        status: 503, 
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }
}

// Cache-First Strategy: Try cache first, then network
async function cacheFirstStrategy(request, cacheName) {
  const cache = await caches.open(cacheName);
  const cachedResponse = await cache.match(request);
  
  if (cachedResponse) {
    // Return cached but also update cache in background
    fetch(request).then((networkResponse) => {
      if (networkResponse.ok) {
        cache.put(request, networkResponse);
      }
    }).catch(() => {});
    
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone());
    }
    return networkResponse;
  } catch (error) {
    console.error('[SW] Fetch failed for:', request.url);
    throw error;
  }
}

// Background Sync for offline actions
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-wishlist-updates') {
    event.waitUntil(syncWishlistUpdates());
  }
});

async function syncWishlistUpdates() {
  // Placeholder for wishlist sync logic
  console.log('[SW] Syncing wishlist updates...');
}

// Push notification support (future enhancement)
self.addEventListener('push', (event) => {
  if (event.data) {
    const data = event.data.json();
    event.waitUntil(
      self.registration.showNotification(data.title, {
        body: data.body,
        icon: '/images/icon-192x192.png',
        badge: '/images/icon-192x192.png',
        data: data.data
      })
    );
  }
});

// Notification click handler
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(
    clients.openWindow(event.notification.data?.url || '/')
  );
});

// Message handling from main thread
self.addEventListener('message', (event) => {
  if (event.data === 'skipWaiting') {
    self.skipWaiting();
  }
});
