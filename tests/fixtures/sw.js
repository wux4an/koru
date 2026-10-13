self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (event) =>
  event.waitUntil(self.clients.claim()),
);
self.addEventListener("fetch", (event) => {
  if (new URL(event.request.url).pathname === "/sw-probe")
    event.respondWith(
      new Response("koru-worker-offline", {
        headers: { "Content-Type": "text/plain" },
      }),
    );
});
