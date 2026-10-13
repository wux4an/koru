// Unit tests for the browser protocol. Android end-to-end tests live separately.
const { test } = require("node:test");
const assert = require("node:assert/strict");
const vm = require("node:vm");
const fs = require("node:fs");
const { MessageChannel } = require("node:worker_threads");
function setup() {
  const listeners = {};
  const w = { addEventListener: (name, fn) => (listeners[name] = fn) };
  w.top = w;
  const ctx = {
    window: w,
    location: { origin: "https://app.koru.invalid" },
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
