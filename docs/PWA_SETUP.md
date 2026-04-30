# PWA (Progressive Web App) Setup Guide

## Overview
Convert DropsAndGrinds into a Progressive Web App for improved mobile experience, offline support, and installability.

## PWA Benefits

- **Installable**: Users can install the app on their devices
- **Offline Support**: Works without internet connection
- **Push Notifications**: Notify users of new deals
- **Fast Loading**: Cached assets load instantly
- **Better UX**: Native app-like experience

## Implementation Steps

### 1. Create Web App Manifest
Create `frontend/manifest.json`:
```json
{
  "name": "DropsAndGrinds",
  "short_name": "DropsAndGrinds",
  "description": "Smart Game Deal Tracker with India-specific pricing",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0d1117",
  "theme_color": "#58a6ff",
  "orientation": "portrait",
  "icons": [
    {
      "src": "/images/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/images/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ],
  "categories": ["games", "shopping"],
  "screenshots": [
    {
      "src": "/images/screenshot-wide.png",
      "sizes": "1280x720",
      "type": "image/png",
      "form_factor": "wide"
    },
    {
      "src": "/images/screenshot-narrow.png",
      "sizes": "750x1334",
      "type": "image/png",
      "form_factor": "narrow"
    }
  ]
}
```

### 2. Link Manifest in HTML
Add to `frontend/index.html`:
```html
<link rel="manifest" href="/manifest.json">
<link rel="apple-touch-icon" href="/images/icon-192x192.png">
<meta name="theme-color" content="#58a6ff">
```

### 3. Create Service Worker
Create `frontend/sw.js`:
```javascript
const CACHE_NAME = 'dropsandgrinds-v1';
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

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
  // Skip API calls - always go to network
  if (event.request.url.includes('/api/')) {
    return;
  }

  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        return cachedResponse;
      }

      return fetch(event.request).then((networkResponse) => {
        // Cache successful responses
        if (networkResponse.ok) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, networkResponse.clone());
          });
        }

        return networkResponse;
      });
    })
  );
});
```

### 4. Register Service Worker
Add to `frontend/js/app.js`:
```javascript
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js')
      .then((registration) => {
        console.log('Service Worker registered:', registration);
      })
      .catch((error) => {
        console.log('Service Worker registration failed:', error);
      });
  });
}
```

### 5. Add Install Prompt
Add to `frontend/js/app.js`:
```javascript
let deferredPrompt;

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredPrompt = e;
  
  // Show install button
  const installBtn = document.getElementById('install-btn');
  if (installBtn) {
    installBtn.style.display = 'block';
    installBtn.addEventListener('click', () => {
      deferredPrompt.prompt();
      deferredPrompt.userChoice.then((choiceResult) => {
        if (choiceResult.outcome === 'accepted') {
          console.log('User accepted install prompt');
        }
        deferredPrompt = null;
        installBtn.style.display = 'none';
      });
    });
  }
});
```

### 6. Add Install Button to UI
Add to `frontend/index.html`:
```html
<button id="install-btn" class="btn btn-primary" style="display: none;">
  Install App
</button>
```

## Offline Strategy

### Cache Strategy
- **Static assets**: Cache first, network fallback
- **API calls**: Network only (for real-time data)
- **HTML pages**: Cache first, network fallback

### Offline Fallback
Create `frontend/offline.html`:
```html
<!DOCTYPE html>
<html>
<head>
  <title>Offline - DropsAndGrinds</title>
  <style>
    body {
      font-family: sans-serif;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
      margin: 0;
      background: #0d1117;
      color: #c9d1d9;
    }
  </style>
</head>
<body>
  <div>
    <h1>You're Offline</h1>
    <p>Please check your internet connection.</p>
  </div>
</body>
</html>
```

Update service worker to handle offline:
```javascript
self.addEventListener('fetch', (event) => {
  if (event.request.url.includes('/api/')) {
    event.respondWith(
      fetch(event.request).catch(() => {
        return new Response(JSON.stringify({ error: 'Offline' }), {
          headers: { 'Content-Type': 'application/json' }
        });
      })
    );
    return;
  }

  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        return cachedResponse;
      }

      return fetch(event.request).catch(() => {
        return caches.match('/offline.html');
      });
    })
  );
});
```

## Push Notifications (Optional)

### 1. Subscribe to Push
```javascript
async function subscribeToPush() {
  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY)
  });
  
  await fetch('/api/push/subscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(subscription)
  });
}
```

### 2. Send Push Notifications (Backend)
Use a library like `webpush` in Go to send notifications.

## Testing

### 1. Lighthouse Audit
Run Lighthouse audit in Chrome DevTools:
- Open DevTools → Lighthouse
- Select Progressive Web App
- Run audit

### 2. Test Offline
- Open DevTools → Network
- Select "Offline"
- Reload page

### 3. Test Install
- Open in Chrome
- Look for install icon in address bar
- Click to install

## Environment Variables
```
VAPID_PUBLIC_KEY=your-vapid-public-key
VAPID_PRIVATE_KEY=your-vapid-private-key
VAPID_SUBJECT=mailto:admin@dropsandgrinds.com
```

## Troubleshooting

### Service Worker Not Registering
- Check file path is correct
- Verify HTTPS is enabled (required for SW)
- Check browser console for errors

### Cache Not Updating
- Update CACHE_NAME version
- Clear old caches in activate event
- Force refresh (Ctrl+Shift+R)

### Install Prompt Not Showing
- Verify manifest is valid
- Check PWA criteria are met
- Ensure site is served over HTTPS
- Test on mobile device
