// koru self-test: every check that needs no human runs when the app opens.
// Results are shown on screen, logged as one `KORU_SELFTEST {...}` line (read
// it with `koru logs`) and exposed as `window.__KORU_SELFTEST__`, a promise of
// the report, for automation over the WebView DevTools protocol.
import { app, device, invoke, isNative, notifications, platform, share, storage, vibration } from "/__koru/core.mjs";

const native = isNative();
const expected = { appId: "dev.koru.selftest", name: "koru self-test", version: "1.0.0", versionCode: 1 };
const groups = [];
let current;

function group(name, fn) {
  current = { name, tests: [] };
  groups.push(current);
  fn();
}
/** A check passes when it returns; `skip(reason)` marks it as not applicable here. */
function test(name, fn) {
  current.tests.push({ name, fn });
}
class Skip extends Error {}
const skip = (reason) => {
  throw new Skip(reason);
};
function assert(condition, message) {
  if (!condition) throw new Error(message);
}
function equal(actual, wanted, what) {
  assert(Object.is(actual, wanted), `${what}: expected ${JSON.stringify(wanted)}, got ${JSON.stringify(actual)}`);
}
/** Resolves when `promise` rejects with a koru error whose code is one of `codes`. */
async function rejects(promise, ...codes) {
  try {
    await promise;
  } catch (error) {
    assert(error instanceof Error, "rejection is not an Error");
    assert(codes.includes(error.code), `expected ${codes.join(" or ")}, got ${error.code}: ${error.message}`);
    return error.code;
  }
  throw new Error(`resolved, expected ${codes.join(" or ")}`);
}
const withTimeout = (promise, ms, what) =>
  Promise.race([promise, new Promise((_, reject) => setTimeout(() => reject(new Error(`${what} took over ${ms} ms`)), ms))]);

group("Bridge", () => {
  test("window.Koru is ready", async () => {
    assert(window.Koru, "window.Koru is missing");
    await withTimeout(window.Koru.ready, 5000, "Koru.ready");
  });
  test("platform matches where the page runs", () => {
    equal(platform, window.Koru.platform, "platform");
    if (native) equal(location.origin, "https://app.koru.invalid", "origin");
    else assert(location.origin !== "https://app.koru.invalid", "web mode on the APK origin");
  });
  test("rejects an invalid method name", () => rejects(invoke("Not A Method"), "INVALID_REQUEST"));
  test("rejects non-object params", () => rejects(invoke("device.info", []), "INVALID_REQUEST"));
  test("rejects an out-of-range timeout", () => rejects(invoke("device.info", {}, { timeout: 0 }), "INVALID_REQUEST"));
  test("honours an aborted signal", () => {
    const controller = new AbortController();
    controller.abort();
    return rejects(invoke("device.info", {}, { signal: controller.signal }), "CANCELLED");
  });
  test("rejects a request over 64 KiB", () => rejects(invoke("storage.set", { key: "big", value: "x".repeat(70000) }), "INVALID_REQUEST"));
  test("unknown method of an enabled plugin", () => rejects(invoke("device.missing"), "METHOD_NOT_FOUND"));
  test("64 concurrent calls all resolve", async () => {
    const all = await Promise.all(Array.from({ length: 64 }, () => device.getInfo()));
    equal(all.length, 64, "replies");
    assert(all.every((info) => typeof info.model === "string"), "a reply has no model");
  });
});

group("app", () => {
  test("getInfo matches koru.json", async () => {
    const info = await app.getInfo();
    if (!native && !window.__KORU_APP__) skip("open it with koru dev or in the APK");
    for (const [field, value] of Object.entries(expected)) equal(info[field], value, field);
  });
});

group("device", () => {
  test("getInfo has the documented shape", async () => {
    const info = await device.getInfo();
    for (const field of ["model", "manufacturer", "osVersion"]) equal(typeof info[field], "string", field);
    assert(Number.isInteger(info.sdk), "sdk is not an integer");
    if (native) assert(info.sdk >= 26, `sdk ${info.sdk} is below minSdk 26`);
    else equal(info.sdk, 0, "sdk in a browser");
    return `${info.manufacturer} ${info.model}, ${native ? `API ${info.sdk}` : "browser"}`;
  });
});

group("storage", () => {
  test("persists across launches", async () => {
    const launches = Number((await storage.get("selftest.launches")) ?? 0) + 1;
    await storage.set("selftest.launches", String(launches));
    equal(await storage.get("selftest.launches"), String(launches), "launch counter");
    return `launch #${launches}`;
  });
  test("set, get and remove round-trip", async () => {
    await storage.set("selftest.value", "hello");
    equal(await storage.get("selftest.value"), "hello", "stored value");
    await storage.remove("selftest.value");
    equal(await storage.get("selftest.value"), null, "after remove");
  });
  test("keeps Unicode, empty and 8 KiB values", async () => {
    for (const value of ["ñ ü 漢字 🚀 \u0000 \n \"quotes\"", "", "k".repeat(8192)]) {
      await storage.set("selftest.text", value);
      equal(await storage.get("selftest.text"), value, `value of ${value.length} chars`);
    }
    await storage.remove("selftest.text");
  });
  test("missing key reads as null", async () => equal(await storage.get("selftest.never-set"), null, "missing key"));
  test("rejects invalid keys", async () => {
    for (const key of ["", "has space", "slash/key", "k".repeat(129)]) await rejects(storage.get(key), "INVALID_REQUEST");
  });
  test("rejects non-string values", () => rejects(invoke("storage.set", { key: "selftest.n", value: 42 }), "INVALID_REQUEST"));
  test("32 concurrent writes", async () => {
    await Promise.all(Array.from({ length: 32 }, (_, i) => storage.set(`selftest.c${i}`, String(i))));
    const values = await Promise.all(Array.from({ length: 32 }, (_, i) => storage.get(`selftest.c${i}`)));
    values.forEach((value, i) => equal(value, String(i), `selftest.c${i}`));
    await Promise.all(Array.from({ length: 32 }, (_, i) => storage.remove(`selftest.c${i}`)));
  });
  test("clear removes every key", async () => {
    const launches = await storage.get("selftest.launches");
    await storage.set("selftest.a", "1");
    await storage.clear();
    equal(await storage.get("selftest.a"), null, "after clear");
    equal(await storage.get("selftest.launches"), null, "launch counter after clear");
    if (launches !== null) await storage.set("selftest.launches", launches);
  });
});

group("vibration", () => {
  test("vibrates for 30 ms", () => vibration.vibrate(30));
  test("rejects durations outside 1–10000 ms", async () => {
    for (const duration of [0, 10001, 1.5]) await rejects(vibration.vibrate(duration), "INVALID_REQUEST");
  });
});

group("share", () => {
  test("rejects non-string text", () => rejects(invoke("share.text", { text: 7 }), "INVALID_REQUEST"));
});

group("notifications", () => {
  test("permission reports granted or denied", async () => {
    const { status } = await notifications.permission();
    assert(status === "granted" || status === "denied", `status ${status}`);
    return status;
  });
  test("show and cancel", async () => {
    const { status } = await notifications.permission();
    if (status !== "granted") {
      await rejects(notifications.show({ id: 4242, title: "koru self-test", body: "Automatic check" }), "PERMISSION_DENIED");
      skip("permission not granted: use the interactive check");
    }
    await notifications.show({ id: 4242, title: "koru self-test", body: "Automatic check" });
    await notifications.cancel(4242);
  });
  test("cancel of an unknown id resolves", () => notifications.cancel(987654));
});

group("Custom plugin (toast)", () => {
  test("toast.show", async () => {
    if (!native) return rejects(invoke("toast.show", { text: "koru self-test" }), "METHOD_NOT_FOUND").then(() => "no web implementation, as documented");
    await invoke("toast.show", { text: "koru self-test" });
  });
  test("rejects missing text", async () => {
    if (!native) skip("custom plugins only run in the APK");
    await rejects(invoke("toast.show", {}), "INVALID_REQUEST");
  });
});

group("Navigation (server)", () => {
  test("user agent carries the bridge marker", () => {
    if (!native) skip("the marker is added by the APK");
    assert(/ KoruApp\/1(?:\s|$)/.test(navigator.userAgent), "no KoruApp/1 in the user agent");
  });
});

group("WebView", () => {
  test("secure context", () => {
    if (native) assert(window.isSecureContext, "the APK origin is not a secure context");
    return window.isSecureContext ? "secure" : "not secure (plain http outside localhost)";
  });
  test("serves app assets", async () => {
    const response = await fetch("/fixture.json", { cache: "no-store" });
    equal(response.status, 200, "status");
    equal((await response.json()).unicode, "ñ ü 漢字 🚀", "UTF-8 content");
  });
  test("missing asset is not served", async () => {
    const response = await fetch("/does-not-exist.txt", { cache: "no-store" });
    assert(!response.ok, `status ${response.status}`);
    return `status ${response.status}`;
  });
  test("Web Worker from the app origin", async () => {
    const worker = new Worker("/worker.js");
    const reply = await withTimeout(
      new Promise((resolve, reject) => {
        worker.onmessage = (event) => resolve(event.data);
        worker.onerror = () => reject(new Error("worker failed to load"));
        worker.postMessage("ping");
      }),
      5000,
      "worker",
    );
    worker.terminate();
    equal(reply.echo, "ping", "echo");
  });
  test("crypto.subtle SHA-256", async () => {
    if (!crypto.subtle) skip("crypto.subtle needs a secure context");
    const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode("abc"));
    const hex = [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
    equal(hex, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", "digest");
  });
  test("WebAssembly", async () => {
    const bytes = new Uint8Array([0, 97, 115, 109, 1, 0, 0, 0, 1, 7, 1, 96, 2, 127, 127, 1, 127, 3, 2, 1, 0, 7, 7, 1, 3, 97, 100, 100, 0, 0, 10, 9, 1, 7, 0, 32, 0, 32, 1, 106, 11]);
    const { instance } = await WebAssembly.instantiate(bytes);
    equal(instance.exports.add(20, 22), 42, "add(20, 22)");
  });
  test("IndexedDB", async () => {
    const db = await new Promise((resolve, reject) => {
      const open = indexedDB.open("koru-selftest", 1);
      open.onupgradeneeded = () => open.result.createObjectStore("kv");
      open.onsuccess = () => resolve(open.result);
      open.onerror = () => reject(open.error);
    });
    const request = (r) => new Promise((resolve, reject) => ((r.onsuccess = () => resolve(r.result)), (r.onerror = () => reject(r.error))));
    await request(db.transaction("kv", "readwrite").objectStore("kv").put("value", "key"));
    equal(await request(db.transaction("kv").objectStore("kv").get("key")), "value", "stored value");
    db.close();
  });
  test("localStorage", () => {
    localStorage.setItem("koru-selftest", "1");
    equal(localStorage.getItem("koru-selftest"), "1", "value");
    localStorage.removeItem("koru-selftest");
  });
  test("history.pushState keeps the page", () => {
    const before = location.pathname;
    history.pushState({ selftest: true }, "", "#selftest");
    equal(location.hash, "#selftest", "hash");
    history.replaceState(null, "", before);
  });
  test("modern CSS", () => {
    for (const [property, value] of [["display", "grid"], ["gap", "1rem"], ["color", "color-mix(in srgb, red, blue)"]])
      assert(CSS.supports(property, value), `${property}: ${value}`);
  });
});

const $ = (selector) => document.querySelector(selector);

async function runTest(entry) {
  const started = performance.now();
  try {
    const detail = await withTimeout(Promise.resolve().then(entry.fn), 15000, "test");
    return { name: entry.name, status: "pass", detail: typeof detail === "string" ? detail : "", ms: performance.now() - started };
  } catch (error) {
    const status = error instanceof Skip ? "skip" : "fail";
    const detail = error && error.code ? `${error.code}: ${error.message}` : String((error && error.message) || error);
    return { name: entry.name, status, detail, ms: performance.now() - started };
  }
}

function render(report) {
  const results = $("#results");
  results.replaceChildren();
  for (const g of report.groups) {
    const section = document.createElement("div");
    section.className = "group";
    const title = document.createElement("h3");
    title.textContent = g.name;
    section.append(title);
    for (const t of g.tests) {
      const row = document.createElement("div");
      row.className = `test ${t.status}`;
      row.innerHTML = '<span class="verdict"></span><span class="name"></span><span class="time"></span>';
      row.querySelector(".verdict").textContent = t.status.toUpperCase();
      row.querySelector(".name").textContent = t.name;
      row.querySelector(".time").textContent = `${Math.round(t.ms)} ms`;
      if (t.detail) {
        const detail = document.createElement("span");
        detail.className = "detail";
        detail.textContent = t.detail;
        row.append(detail);
      }
      section.append(row);
    }
    results.append(section);
  }
  const { passed, failed, skipped, total } = report.summary;
  const summary = $("#summary");
  summary.className = `summary ${failed ? "fail" : "pass"}`;
  summary.textContent = `${failed ? "FAIL" : "PASS"} · ${passed}/${total} passed, ${failed} failed, ${skipped} skipped · ${Math.round(report.ms)} ms`;
  document.title = `${failed ? "FAIL" : "PASS"} ${passed}/${total} · koru self-test`;
}

async function runAll() {
  $("#rerun").disabled = $("#copy").disabled = true;
  $("#summary").className = "summary";
  $("#summary").textContent = "Running…";
  const started = performance.now();
  const report = { platform, userAgent: navigator.userAgent, startedAt: new Date().toISOString(), groups: [] };
  for (const g of groups) {
    const done = { name: g.name, tests: [] };
    report.groups.push(done);
    for (const t of g.tests) done.tests.push(await runTest(t));
  }
  const all = report.groups.flatMap((g) => g.tests);
  report.ms = performance.now() - started;
  report.summary = {
    total: all.length,
    passed: all.filter((t) => t.status === "pass").length,
    failed: all.filter((t) => t.status === "fail").length,
    skipped: all.filter((t) => t.status === "skip").length,
  };
  render(report);
  console.info("KORU_SELFTEST " + JSON.stringify({ platform, ...report.summary, failures: all.filter((t) => t.status === "fail").map((t) => `${t.name}: ${t.detail}`) }));
  $("#rerun").disabled = $("#copy").disabled = false;
  lastReport = report;
  return report;
}

let lastReport;
$("#rerun").addEventListener("click", () => (window.__KORU_SELFTEST__ = runAll()));
$("#copy").addEventListener("click", async () => {
  const text = JSON.stringify(lastReport, null, 2);
  try {
    await navigator.clipboard.writeText(text);
    $("#copy").textContent = "Copied";
  } catch {
    await share.text(text).catch(() => {});
  }
  setTimeout(() => ($("#copy").textContent = "Copy report"), 1500);
});

const manual = {
  share: () => share.text("Sent from the koru self-test"),
  notify: async () => {
    const { status } = await notifications.requestPermission();
    if (status !== "granted") throw Object.assign(new Error("Permission denied"), { code: "PERMISSION_DENIED" });
    await notifications.show({ id: 7, title: "koru self-test", body: "Notifications work." });
  },
  vibrate: () => vibration.vibrate(400),
  toast: () => invoke("toast.show", { text: "The custom toast plugin works.", long: true }),
};
document.querySelectorAll("[data-manual]").forEach((button) =>
  button.addEventListener("click", async () => {
    const out = $("#manual-status");
    out.textContent = `${button.textContent}…`;
    try {
      await manual[button.dataset.manual]();
      out.textContent = `${button.textContent}: done.`;
    } catch (error) {
      out.textContent = `${button.textContent}: ${error.code ?? "ERROR"} ${error.message}`;
    }
  }),
);

$("#platform").textContent = native ? "Android APK" : "Browser (web implementations)";
window.__KORU_SELFTEST__ = runAll();
