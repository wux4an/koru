/* koru v1 protocol: one origin-bound MessagePort owned by the main frame.
   Outside the app origin (a desktop browser or `koru dev`) calls run against
   web implementations instead, so the same code works in both places. */
(() => {
  "use strict";
  if (typeof window === "undefined" || window.Koru) return;
  const origin = "https://app.koru.invalid";
  const native = location.origin === origin;
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
  if (!native) readyResolve();
  const storageKey = /^[A-Za-z0-9_.:-]{1,128}$/;
  function webStore() {
    try {
      return window.localStorage;
    } catch (ignored) {
      throw fail("STORAGE_ERROR", "Browser storage is unavailable");
    }
  }
  function key(params) {
    if (typeof params.key !== "string" || !storageKey.test(params.key))
      throw fail("INVALID_REQUEST", "Invalid storage key");
    return "koru:" + params.key;
  }
  function integer(value, fallback, min, max) {
    const n = value === undefined ? fallback : value;
    if (!Number.isInteger(n) || n < min || n > max)
      throw fail("INVALID_REQUEST", "Invalid integer parameter");
    return n;
  }
  function webPermission() {
    const N = window.Notification;
    return { status: N && N.permission === "granted" ? "granted" : "denied" };
  }
  const shown = new Map();
  const web = {
    "device.info": () => {
      const nav = window.navigator || {};
      return {
        model: "Web browser",
        manufacturer: String(nav.vendor || ""),
        osVersion: String(nav.userAgent || ""),
        sdk: 0,
      };
    },
    "app.info": () => {
      const dev = window.__KORU_APP__ || {};
      return {
        appId: String(dev.appId || "web"),
        name: String(dev.name || (window.document && document.title) || ""),
        version: String(dev.version || "0.0.0"),
        versionCode: Number(dev.versionCode || 0),
      };
    },
    "storage.get": (params) => {
      const value = webStore().getItem(key(params));
      return value === null ? {} : { value };
    },
    "storage.set": (params) => {
      const k = key(params);
      if (typeof params.value !== "string")
        throw fail("INVALID_REQUEST", "Storage values must be strings");
      webStore().setItem(k, params.value);
      return {};
    },
    "storage.remove": (params) => {
      webStore().removeItem(key(params));
      return {};
    },
    "storage.clear": () => {
      const store = webStore();
      for (let i = store.length - 1; i >= 0; i--) {
        const k = store.key(i);
        if (k && k.startsWith("koru:")) store.removeItem(k);
      }
      return {};
    },
    "vibration.vibrate": (params) => {
      const duration = integer(params.duration, 100, 1, 10000);
      const nav = window.navigator || {};
      if (typeof nav.vibrate === "function") nav.vibrate(duration);
      return {};
    },
    "share.text": async (params) => {
      if (typeof params.text !== "string")
        throw fail("INVALID_REQUEST", "Share text must be a string");
      const nav = window.navigator || {};
      if (typeof nav.share === "function") {
        try {
          await nav.share({ text: params.text });
        } catch (error) {
          if (!error || error.name !== "AbortError")
            throw fail("NATIVE_ERROR", "Browser share failed");
        }
      } else if (nav.clipboard && typeof nav.clipboard.writeText === "function")
        await nav.clipboard.writeText(params.text);
      else throw fail("UNAVAILABLE", "This browser cannot share text");
      return {};
    },
    "notifications.permission": () => webPermission(),
    "notifications.requestPermission": async () => {
      const N = window.Notification;
      if (N && N.permission === "default") await N.requestPermission();
      return webPermission();
    },
    "notifications.show": (params) => {
      const id = integer(params.id, 1, 0, 2147483647);
      if (typeof params.title !== "string" || typeof params.body !== "string")
        throw fail("INVALID_REQUEST", "Notification title and body required");
      if (webPermission().status !== "granted")
        throw fail("PERMISSION_DENIED", "Notifications are disabled");
      shown.get(id) && shown.get(id).close();
      shown.set(id, new window.Notification(params.title, { body: params.body }));
      return {};
    },
    "notifications.cancel": (params) => {
      const id = integer(params.id, 1, 0, 2147483647);
      shown.get(id) && shown.get(id).close();
      shown.delete(id);
      return {};
    },
  };
  function webInvoke(method, params) {
    const dev = window.__KORU_APP__;
    const plugin = method.slice(0, method.indexOf("."));
    const enabled =
      !dev || !Array.isArray(dev.plugins) || dev.plugins.includes(plugin);
    if (!enabled || !Object.prototype.hasOwnProperty.call(web, method))
      return Promise.reject(
        fail(
          "METHOD_NOT_FOUND",
          enabled
            ? "No web implementation for " + method
            : "Plugin " + plugin + " is not enabled in koru.json",
        ),
      );
    return new Promise((resolve) => resolve(web[method](params)));
  }
  const Koru = Object.freeze({
    ready,
    platform: native ? "android" : "web",
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
      if (!native) return webInvoke(method, params);
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
