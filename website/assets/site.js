(() => {
  const site = JSON.parse(document.getElementById("site-config").textContent);
  const t = site.text[site.lang];
  const path = (lang, slug) =>
    `${site.base}${lang}/${slug === "overview" ? "" : slug + "/"}`;
  const update = (vm) => {
    const parts = vm.route.path
      .slice(site.base.length)
      .split("/")
      .filter(Boolean);
    const lang = ["en", "es"].includes(parts[0]) ? parts[0] : site.lang;
    const slug = parts[1] || "overview";
    const page = site.pages[slug];
    if (!page) {
      document.querySelector('meta[name="robots"]').content = "noindex";
      return;
    }
    const index = lang === "en" ? 0 : 1;
    document.documentElement.lang = lang;
    document.querySelector('meta[name="robots"]').content = site.production
      ? "index,follow"
      : "noindex,follow";
    document.body.dataset.page = slug;
    document.title = `${page.titles[index]} · koru`;
    const url = new URL(path(lang, slug), site.siteURL).href;
    document.querySelector('meta[name="description"]').content =
      page.descriptions[index];
    document.querySelector('link[rel="canonical"]').href = url;
    document.querySelector('meta[property="og:url"]').content = url;
    document.querySelector('meta[property="og:title"]').content =
      document.title;
    document.querySelector('meta[property="og:description"]').content =
      page.descriptions[index];
    document.querySelector('meta[name="twitter:title"]').content =
      document.title;
    document.querySelector('meta[name="twitter:description"]').content =
      page.descriptions[index];
    document.querySelectorAll("link[hreflang]").forEach((link) => {
      link.href = new URL(
        path(link.hreflang === "x-default" ? "en" : link.hreflang, slug),
        site.siteURL,
      ).href;
    });
    document.querySelectorAll("[data-locale]").forEach((link) => {
      link.href = path(link.dataset.locale, slug);
      if (link.dataset.locale === lang)
        link.setAttribute("aria-current", "true");
      else link.removeAttribute("aria-current");
    });
    const structured = JSON.parse(
      document.getElementById("structured-data").textContent,
    );
    Object.assign(structured, {
      headline: page.titles[index],
      description: page.descriptions[index],
      inLanguage: lang,
      url,
    });
    document.getElementById("structured-data").textContent =
      JSON.stringify(structured);
    const main = document.getElementById("main");
    main.setAttribute("tabindex", "-1");
    if (!main.querySelector(".page-toc")) {
      const headings = [...main.querySelectorAll("h2[id]")];
      if (headings.length >= 3) {
        const contents = document.createElement("details");
        contents.className = "page-toc";
        const summary = document.createElement("summary");
        summary.textContent = t.contents;
        const nav = document.createElement("nav");
        nav.setAttribute("aria-label", t.contents);
        const list = document.createElement("ul");
        for (const heading of headings) {
          const item = document.createElement("li");
          const link = document.createElement("a");
          link.href = `#${heading.id}`;
          link.textContent = heading.textContent;
          item.append(link);
          list.append(item);
        }
        nav.append(list);
        contents.append(summary, nav);
        headings[0].before(contents);
      }
    }
    // Docsify's existing widgets retain their behavior, with localized names.
    const toggle = document.querySelector(".sidebar-toggle");
    toggle?.setAttribute("aria-label", t.menu);
    const search = document.querySelector(".search input");
    search?.setAttribute("aria-label", t.search);
    const clear = document.querySelector(".clear-button");
    if (clear) {
      clear.setAttribute("role", "button");
      clear.setAttribute("tabindex", "0");
      clear.setAttribute(
        "aria-label",
        lang === "en" ? "Clear search" : "Borrar búsqueda",
      );
      clear.onkeydown = (event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          clear.click();
          search.focus();
        }
      };
    }
    document.querySelector(".sidebar")?.setAttribute("aria-label", t.docs);
    document.querySelector(".content")?.removeAttribute("role");
    main
      .querySelectorAll("pre code, table")
      .forEach((el) => el.setAttribute("tabindex", "0"));
    const fixCopyARIA = () =>
      main
        .querySelectorAll('[aria-hidden="hidden"]')
        .forEach((el) => el.setAttribute("aria-hidden", "true"));
    queueMicrotask(fixCopyARIA);
    document.querySelectorAll(".sidebar-nav a").forEach((link) => {
      const current =
        link.getAttribute("href")?.split("?")[0] === path(lang, slug);
      if (current) link.setAttribute("aria-current", "page");
      else link.removeAttribute("aria-current");
    });
    if (!main.querySelector(".doc-footer")) {
      const footer = document.createElement("footer");
      footer.className = "doc-footer";
      footer.append(document.createTextNode(t.footer));
      const version = document.createElement("span");
      version.textContent = t.note;
      footer.append(version);
      main.append(footer);
    }
    // Change language through a full document navigation so plugins and search
    // index are initialized in the requested locale, including browser Back.
    document.querySelectorAll("[data-locale]").forEach((link) => {
      link.onclick = (event) => {
        event.preventDefault();
        window.location.assign(link.href);
      };
    });
    document.dispatchEvent(new CustomEvent("koru:page-ready"));
  };
  const sidebarAlias = {};
  sidebarAlias[`${site.base}(en|es)/.*/_sidebar.md`] =
    `${site.base}$1/_sidebar.md`;
  window.$docsify = {
    el: "#app",
    name: "",
    routerMode: "history",
    basePath: "/",
    loadSidebar: true,
    subMaxLevel: 0,
    maxLevel: 3,
    auto2top: true,
    relativePath: false,
    alias: sidebarAlias,
    notFoundPage: `${site.base}_404.md`,
    externalLinkTarget: "_self",
    executeScript: false,
    search: {
      paths: "auto",
      pathNamespaces: [site.base + "en", site.base + "es"],
      namespace: "koru-docs-v2",
      placeholder: t.search,
      noData: t.empty,
      depth: 3,
      maxAge: 86400000,
      hideOtherSidebarContent: false,
    },
    copyCode: {
      buttonText: t.copy,
      successText: t.copied,
      errorText: t.copyError,
    },
    plugins: [
      (hook, vm) => {
        hook.doneEach(() => update(vm));
        hook.ready(() => {
          update(vm);
          // A skip link should focus the reading area without becoming a route.
          document.querySelector(".skip-link").onclick = (event) => {
            event.preventDefault();
            const main = document.getElementById("main");
            main.focus();
            main.scrollIntoView();
          };
          const toggle = document.querySelector(".sidebar-toggle");
          if (toggle) {
            const sync = () => {
              const expanded =
                window.innerWidth < 768
                  ? document.body.classList.contains("close")
                  : !document.body.classList.contains("close");
              toggle.setAttribute("aria-expanded", String(expanded));
              const sidebar = document.querySelector(".sidebar");
              sidebar.inert = !expanded;
              sidebar.setAttribute("aria-hidden", String(!expanded));
            };
            window.addEventListener("resize", sync);
            document.addEventListener("keydown", (event) => {
              if (
                event.key === "Escape" &&
                window.innerWidth < 768 &&
                document.body.classList.contains("close")
              ) {
                toggle.click();
                toggle.focus();
              }
            });
            new MutationObserver(sync).observe(document.body, {
              attributes: true,
              attributeFilter: ["class"],
            });
            sync();
          }
        });
      },
    ],
  };
})();
