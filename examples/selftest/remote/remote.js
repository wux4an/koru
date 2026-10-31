// Checks that a bridged remote page gets the native API inside the app.
// Results also go to logcat as one `KORU_SELFTEST_REMOTE {...}` line.
const results = [];
const check = async (name, fn) => {
  try {
    const detail = await fn();
    results.push({ name, status: "pass", detail: detail ?? "" });
  } catch (error) {
    results.push({ name, status: "fail", detail: `${error.code ?? ""} ${error.message}`.trim() });
  }
};
const assert = (ok, message) => {
  if (!ok) throw new Error(message);
};
const inApp = / KoruApp\/1(?:\s|$)/.test(navigator.userAgent);

await check("opened inside the koru app", () => assert(inApp, "the user agent has no KoruApp marker: open this page from the self-test app"));
await check("platform is android", () => assert(Koru.platform === "android", `platform ${Koru.platform}`));
await check("native channel arrives", async () => {
  await Promise.race([Koru.ready, new Promise((_, reject) => setTimeout(() => reject(new Error("no channel after 10 s")), 10000))]);
});
const channel = results.at(-1).status === "pass";
const native = (name, fn) =>
  channel ? check(name, fn) : results.push({ name, status: "fail", detail: "not run: no native channel" });

await native("app.info is the self-test app", async () => {
  const info = await Koru.invoke("app.info");
  assert(info.appId === "dev.koru.selftest", `appId ${info.appId}`);
  return info.appId;
});
await native("device.info comes from Android", async () => {
  const info = await Koru.invoke("device.info");
  assert(info.sdk >= 26, `sdk ${info.sdk}`);
  return `${info.manufacturer} ${info.model}, API ${info.sdk}`;
});
await native("storage is shared with the app", async () => {
  const value = `remote-${Date.now()}`;
  await Koru.invoke("storage.set", { key: "selftest.remote", value });
  const read = await Koru.invoke("storage.get", { key: "selftest.remote" });
  assert(read.value === value, `read ${JSON.stringify(read)}`);
});

const failed = results.filter((r) => r.status === "fail").length;
const summary = document.querySelector("#summary");
summary.textContent = `${failed ? "FAIL" : "PASS"} · ${results.length - failed}/${results.length} passed`;
summary.className = failed ? "fail" : "pass";
for (const r of results) {
  const li = document.createElement("li");
  li.className = r.status;
  li.textContent = `${r.status.toUpperCase()} ${r.name}${r.detail ? ` — ${r.detail}` : ""}`;
  document.querySelector("#results").append(li);
}
console.info("KORU_SELFTEST_REMOTE " + JSON.stringify({ passed: results.length - failed, failed, results }));
