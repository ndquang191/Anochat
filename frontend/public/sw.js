const CACHE_NAME = "anochat-shell-v1";
const OFFLINE_URL = "/offline";
const PRECACHE = [OFFLINE_URL, "/manifest.webmanifest", "/icon.svg"];

self.addEventListener("install", (event) => {
	event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE)));
	self.skipWaiting();
});

self.addEventListener("activate", (event) => {
	event.waitUntil(
		Promise.all([
			caches.keys().then((keys) => Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key)))),
			self.clients.claim(),
		]),
	);
});

self.addEventListener("fetch", (event) => {
	const request = event.request;
	if (request.method !== "GET" || request.mode !== "navigate") return;
	event.respondWith(fetch(request).catch(() => caches.match(OFFLINE_URL)));
});

self.addEventListener("push", (event) => {
	if (!event.data) return;
	event.waitUntil((async () => {
		let payload;
		try { payload = event.data.json(); } catch { return; }
		if (payload?.type !== "match_found") return;
		const windows = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
		const visible = windows.find((client) => client.visibilityState === "visible");
		if (visible) {
			windows.forEach((client) => client.postMessage({ type: "match_found" }));
			return;
		}
		await self.registration.showNotification(payload.title, {
			body: payload.body,
			icon: "/icons/icon-192.png",
			badge: "/icons/badge.png",
			tag: payload.tag || "match-found",
			data: { url: payload.url || "/" },
		});
	})());
});

self.addEventListener("notificationclick", (event) => {
	event.notification.close();
	const target = new URL(event.notification.data?.url || "/", self.location.origin).href;
	event.waitUntil((async () => {
		const windows = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
		for (const client of windows) {
			if ("focus" in client) {
				await client.focus();
				if ("navigate" in client) await client.navigate(target);
				return;
			}
		}
		await self.clients.openWindow(target);
	})());
});
