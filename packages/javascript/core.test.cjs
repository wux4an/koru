// Unit tests for the browser protocol. Android end-to-end tests live separately.
const { test } = require("node:test");
const assert = require("node:assert/strict");
const vm = require("node:vm");
const fs = require("node:fs");
const { MessageChannel } = require("node:worker_threads");
function setup(origin = "https://app.koru.invalid", extra = {}) {
  const listeners = {};
  const w = { addEventListener: (name, fn) => (listeners[name] = fn), ...extra };
  w.top = w;
  const ctx = {
    window: w,
    location: { origin },
    setTimeout,
    clearTimeout,
    Promise,
    Map,
    Error,
    JSON,
    Object,
    Number,
  };
  vm.createContext(ctx);
  vm.runInContext(fs.readFileSync(__dirname + "/core.js", "utf8"), ctx);
  ctx.Koru = w.Koru;
  return { ctx, listeners };
}
test("port handles promises, errors and rejects foreign/synthetic handshakes", async () => {
  const { ctx, listeners } = setup();
  const { port1, port2 } = new MessageChannel();
  listeners.message({
    isTrusted: false,
    source: null,
    data: "koru-ready",
    ports: [port1],
  });
  await assert.rejects(
    ctx.Koru.invoke("device.info", {}, { timeout: 10 }),
    (e) => e.code === "TIMEOUT",
  );
  listeners.message({
    isTrusted: true,
    source: null,
    data: "koru-ready",
    ports: [port1],
  });
  port2.on("message", (payload) => {
    const m = JSON.parse(payload);
    port2.postMessage(
      JSON.stringify(
        m.method === "unknown.call"
          ? {
              id: m.id,
              error: { code: "METHOD_NOT_FOUND", message: "unknown" },
            }
          : { id: m.id, result: { platform: "android" } },
      ),
    );
  });
  assert.equal((await ctx.Koru.invoke("device.info")).platform, "android");
  await assert.rejects(
    ctx.Koru.invoke("unknown.call"),
    (e) => e.code === "METHOD_NOT_FOUND",
  );
  port1.close();
  port2.close();
});
test("validates inputs, timeout and cancellation", async () => {
  const { ctx } = setup();
  await assert.rejects(
    ctx.Koru.invoke("bad", {}),
    (e) => e.code === "INVALID_REQUEST",
  );
  await assert.rejects(
    ctx.Koru.invoke("device.info", {}, { timeout: 0 }),
    (e) => e.code === "INVALID_REQUEST",
  );
  await assert.rejects(
    ctx.Koru.invoke("device.info", {}, { signal: { aborted: true } }),
    (e) => e.code === "CANCELLED",
  );
});

function memoryStorage() {
  const data = new Map();
  return {
    get length() {
      return data.size;
    },
    key: (i) => [...data.keys()][i] ?? null,
    getItem: (k) => (data.has(k) ? data.get(k) : null),
    setItem: (k, v) => data.set(k, String(v)),
    removeItem: (k) => data.delete(k),
    data,
  };
}
test("browser origin uses web implementations without waiting", async () => {
  const localStorage = memoryStorage();
  localStorage.setItem("other", "kept");
  const vibrations = [];
  const shared = [];
  const { ctx } = setup("http://127.0.0.1:5173", {
    localStorage,
    navigator: {
      vendor: "Test",
      userAgent: "UA",
      vibrate: (ms) => vibrations.push(ms),
      share: async (data) => shared.push(data.text),
    },
    __KORU_APP__: {
      appId: "com.example.demo",
      name: "Demo",
      version: "1.2.3",
      versionCode: 7,
      plugins: ["device", "storage", "app", "vibration", "share"],
    },
  });
  assert.equal(ctx.Koru.platform, "web");
  await ctx.Koru.ready;
  assert.equal((await ctx.Koru.invoke("device.info")).sdk, 0);
  assert.equal((await ctx.Koru.invoke("app.info")).versionCode, 7);
  assert.deepEqual(
    { ...(await ctx.Koru.invoke("storage.get", { key: "theme" })) },
    {},
  );
  await ctx.Koru.invoke("storage.set", { key: "theme", value: "dark" });
  assert.equal(
    (await ctx.Koru.invoke("storage.get", { key: "theme" })).value,
    "dark",
  );
  await assert.rejects(
    ctx.Koru.invoke("storage.set", { key: "bad key", value: "x" }),
    (e) => e.code === "INVALID_REQUEST",
  );
  await assert.rejects(
    ctx.Koru.invoke("storage.set", { key: "n", value: 5 }),
    (e) => e.code === "INVALID_REQUEST",
  );
  await ctx.Koru.invoke("storage.clear");
  assert.equal(localStorage.getItem("koru:theme"), null);
  assert.equal(localStorage.getItem("other"), "kept");
  await ctx.Koru.invoke("vibration.vibrate", { duration: 40 });
  await assert.rejects(
    ctx.Koru.invoke("vibration.vibrate", { duration: 0 }),
    (e) => e.code === "INVALID_REQUEST",
  );
  assert.deepEqual(vibrations, [40]);
  await ctx.Koru.invoke("share.text", { text: "hi" });
  assert.deepEqual(shared, ["hi"]);
  await assert.rejects(
    ctx.Koru.invoke("notifications.permission"),
    (e) => e.code === "METHOD_NOT_FOUND" && /not enabled/.test(e.message),
  );
  await assert.rejects(
    ctx.Koru.invoke("custom.call"),
    (e) => e.code === "METHOD_NOT_FOUND",
  );
});
test("web notifications follow the browser permission", async () => {
  const shown = [];
  class Notification {
    static permission = "default";
    static async requestPermission() {
      Notification.permission = "granted";
      return "granted";
    }
    constructor(title, options) {
      this.closed = false;
      shown.push([title, options.body, this]);
    }
    close() {
      this.closed = true;
    }
  }
  const { ctx } = setup("http://localhost:4000", { Notification });
  assert.equal(
    (await ctx.Koru.invoke("notifications.permission")).status,
    "denied",
  );
  await assert.rejects(
    ctx.Koru.invoke("notifications.show", { id: 1, title: "a", body: "b" }),
    (e) => e.code === "PERMISSION_DENIED",
  );
  assert.equal(
    (await ctx.Koru.invoke("notifications.requestPermission")).status,
    "granted",
  );
  await ctx.Koru.invoke("notifications.show", { id: 1, title: "a", body: "b" });
  await ctx.Koru.invoke("notifications.cancel", { id: 1 });
  assert.equal(shown.length, 1);
  assert.equal(shown[0][2].closed, true);
});
test("android origin reports the android platform", () => {
  const { ctx } = setup();
  assert.equal(ctx.Koru.platform, "android");
});
