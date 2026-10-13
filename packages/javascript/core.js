/* koru v1 protocol: one origin-bound MessagePort owned by the main frame. */
(() => {
  "use strict";
  if (typeof window === "undefined" || window.Koru) return;
  const origin = "https://app.koru.invalid";
  const pending = new Map();
  let channel = null,
    nextId = 0;
  let readyResolve;
  const ready = new Promise((resolve) => {
    readyResolve = resolve;
  });
  function fail(code, message) {
    const error = new Error(message);
    error.code = code;
    return error;
  }
  function rejectAll(code, message) {
    for (const entry of pending.values()) {
      clearTimeout(entry.timer);
      entry.cleanup();
      entry.reject(fail(code, message));
    }
    pending.clear();
  }
  window.addEventListener("message", (event) => {
    if (
      location.origin !== origin ||
      window !== window.top ||
      !event.isTrusted ||
      event.source !== null ||
      event.data !== "koru-ready" ||
      event.ports.length !== 1
    )
      return;
    if (channel) {
      channel.close();
      rejectAll("CHANNEL_REPLACED", "Native channel replaced");
    }
    channel = event.ports[0];
    channel.onmessage = (event) => {
      if (typeof event.data !== "string" || event.data.length > 131072) return;
      let reply;
      try {
        reply = JSON.parse(event.data);
      } catch (ignored) {
        return;
      }
      if (!reply || typeof reply.id !== "string") return;
      const entry = pending.get(reply.id);
      if (!entry) return;
      pending.delete(reply.id);
      clearTimeout(entry.timer);
      entry.cleanup();
      if (reply.error && typeof reply.error.code === "string")
        entry.reject(
          fail(reply.error.code, String(reply.error.message || "Native error")),
        );
      else if (Object.prototype.hasOwnProperty.call(reply, "result"))
        entry.resolve(reply.result);
      else entry.reject(fail("INVALID_RESPONSE", "Invalid native response"));
    };
    channel.onmessageerror = () =>
      rejectAll("CHANNEL_ERROR", "Native channel error");
    channel.start();
    readyResolve();
  });
  const Koru = Object.freeze({
    ready,
    invoke(method, params = {}, options = {}) {
      if (
        typeof method !== "string" ||
        !/^[a-z][a-z0-9_]{0,31}\.[a-zA-Z][a-zA-Z0-9_]{0,63}$/.test(method) ||
        !params ||
        typeof params !== "object" ||
        Array.isArray(params)
      )
        return Promise.reject(
          fail("INVALID_REQUEST", "Invalid method or params"),
        );
      if (options.signal && options.signal.aborted)
        return Promise.reject(fail("CANCELLED", "Request cancelled"));
      const timeout = options.timeout === undefined ? 10000 : options.timeout;
      if (!Number.isFinite(timeout) || timeout < 1 || timeout > 120000)
        return Promise.reject(fail("INVALID_REQUEST", "Invalid timeout"));
      if (pending.size >= 256)
        return Promise.reject(fail("BUSY", "Too many pending requests"));
      nextId = (nextId % 2147483647) + 1;
      const id = String(nextId);
      let payload;
      try {
        payload = JSON.stringify({ id, method, params });
      } catch (ignored) {
        return Promise.reject(
          fail("INVALID_REQUEST", "Params must be serializable"),
        );
      }
      if (payload.length > 65535)
        return Promise.reject(fail("INVALID_REQUEST", "Request too large"));
      return new Promise((resolve, reject) => {
        const cancel = () => {
          const entry = pending.get(id);
          if (!entry) return;
          pending.delete(id);
          clearTimeout(entry.timer);
          cleanup();
          reject(
            fail(
              "CANCELLED",
              "Request cancelled; dispatched native actions cannot be undone",
            ),
          );
        };
        const cleanup = () =>
          options.signal && options.signal.removeEventListener("abort", cancel);
        const timer = setTimeout(() => {
          pending.delete(id);
          cleanup();
          reject(fail("TIMEOUT", "Native request timed out"));
        }, timeout);
        pending.set(id, { resolve, reject, timer, cleanup });
        options.signal &&
          options.signal.addEventListener("abort", cancel, { once: true });
        ready.then(() => {
          if (pending.has(id)) {
            try {
              channel.postMessage(payload);
            } catch (ignored) {
              pending.delete(id);
              clearTimeout(timer);
              cleanup();
              reject(fail("CHANNEL_ERROR", "Cannot send native request"));
            }
          }
        });
      });
    },
  });
  Object.defineProperty(window, "Koru", {
    value: Koru,
    writable: false,
    configurable: false,
  });
  window.addEventListener("pagehide", () => {
    rejectAll("PAGE_CLOSED", "Page closed");
    channel && channel.close();
    channel = null;
  });
})();
