// Shows the latest published koru release. The server-rendered value (the
// version in cmd/koru/main.go at build time) stays when the GitHub API fails.
(() => {
  const api = "https://api.github.com/repos/wux4an/koru/releases/latest";
  const releases = "https://github.com/wux4an/koru/releases/tag/";
  const key = "koru-latest-release";
  const maxAge = 60 * 60 * 1000;
  const valid = /^v\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$/;
  let latest = null;
  const apply = () => {
    if (!latest) return;
    document.querySelectorAll("[data-koru-version]").forEach((el) => {
      el.textContent = latest;
      if (el instanceof HTMLAnchorElement) el.href = releases + latest;
    });
  };
  const cached = () => {
    try {
      const entry = JSON.parse(localStorage.getItem(key) || "null");
      if (
        entry &&
        valid.test(entry.tag) &&
        Date.now() - entry.at >= 0 &&
        Date.now() - entry.at < maxAge
      )
        return entry.tag;
    } catch {}
    return null;
  };
  const store = (tag) => {
    try {
      localStorage.setItem(key, JSON.stringify({ tag, at: Date.now() }));
    } catch {}
  };
  // The docs footer is created after Docsify renders, so apply again then.
  document.addEventListener("koru:page-ready", apply);
  latest = cached();
  apply();
  if (latest) return;
  fetch(api, { headers: { Accept: "application/vnd.github+json" } })
    .then((response) => {
      if (!response.ok) throw new Error(String(response.status));
      return response.json();
    })
    .then((release) => {
      if (!release || !valid.test(release.tag_name)) return;
      latest = release.tag_name;
      store(latest);
      apply();
    })
    .catch(() => {});
})();
