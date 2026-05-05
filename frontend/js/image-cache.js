/**
 * IndexedDB Image & Metadata Cache
 * Caches game covers and basic game data to avoid repeated fetches.
 * Images persist indefinitely; metadata TTL = 24 hours.
 */

const DB_NAME = 'DropsAndGrindsCache';
const DB_VERSION = 1;
const META_TTL_MS = 24 * 60 * 60 * 1000; // 24 hours

let dbPromise = null;

function openCacheDB() {
    if (dbPromise) return dbPromise;
    dbPromise = new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, DB_VERSION);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);
        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('images')) {
                db.createObjectStore('images', { keyPath: 'url' });
            }
            if (!db.objectStoreNames.contains('gameMeta')) {
                db.createObjectStore('gameMeta', { keyPath: 'gameId' });
            }
        };
    });
    return dbPromise;
}

function cacheImage(url, blob) {
    return new Promise(async (resolve, reject) => {
        try {
            const db = await openCacheDB();
            const tx = db.transaction('images', 'readwrite');
            const store = tx.objectStore('images');
            store.put({ url, blob, cachedAt: Date.now() });
            tx.oncomplete = () => resolve();
            tx.onerror = () => reject(tx.error);
        } catch (e) {
            console.warn('Failed to cache image:', e);
            reject(e);
        }
    });
}

async function getCachedImage(url) {
    try {
        const db = await openCacheDB();
        const tx = db.transaction('images', 'readonly');
        const store = tx.objectStore('images');
        const request = store.get(url);
        return new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    } catch (e) {
        console.warn('Failed to read cached image:', e);
        return null;
    }
}

function cacheGameMeta(gameId, data) {
    return new Promise(async (resolve, reject) => {
        try {
            const db = await openCacheDB();
            const tx = db.transaction('gameMeta', 'readwrite');
            const store = tx.objectStore('gameMeta');
            store.put({ gameId: String(gameId), data, cachedAt: Date.now() });
            tx.oncomplete = () => resolve();
            tx.onerror = () => reject(tx.error);
        } catch (e) {
            console.warn('Failed to cache game meta:', e);
            reject(e);
        }
    });
}

async function getCachedGameMeta(gameId) {
    try {
        const db = await openCacheDB();
        const tx = db.transaction('gameMeta', 'readonly');
        const store = tx.objectStore('gameMeta');
        const request = store.get(String(gameId));
        const result = await new Promise((resolve, reject) => {
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
        if (!result) return null;
        if (Date.now() - result.cachedAt > META_TTL_MS) return null;
        return result.data;
    } catch (e) {
        console.warn('Failed to read cached game meta:', e);
        return null;
    }
}

/**
 * Fetch an image URL, caching the blob for reuse.
 * Returns an object URL suitable for img.src.
 */
async function fetchCachedImage(url) {
    if (!url) return '';

    // Check IndexedDB cache first
    const cached = await getCachedImage(url);
    if (cached && cached.blob) {
        return URL.createObjectURL(cached.blob);
    }

    // Fetch from network
    try {
        const response = await fetch(url);
        if (!response.ok) throw new Error('Failed to fetch image');
        const blob = await response.blob();
        cacheImage(url, blob); // Fire and forget
        return URL.createObjectURL(blob);
    } catch (e) {
        console.warn('Image fetch failed, falling back to proxy:', e);
        return url;
    }
}

window.imageCache = {
    fetchCachedImage,
    cacheGameMeta,
    getCachedGameMeta,
};
