import { mkdir, readFile, writeFile, readdir, cp, rm } from "node:fs/promises";
import { resolve, dirname, relative, extname, posix } from "node:path";
import { createHash } from "node:crypto";
import { Marked } from "marked";
import { locales, pages, route, text, type Locale } from "./catalog";

export const root = resolve(import.meta.dir, "..");
const repo = resolve(root, "..");
export const escape = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (char) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        char
      ]!,
  );
export function siteConfig(
  value = process.env.SITE_URL || "http://localhost:4173/",
) {
  const url = new URL(value);
  if (
    !["http:", "https:"].includes(url.protocol) ||
    url.search ||
    url.hash ||
    url.username ||
    url.password
  )
    throw new Error(
      "SITE_URL must be an HTTP(S) URL without credentials, query or fragment.",
    );
  url.pathname = url.pathname.replace(/\/+$/, "") + "/";
  const production = !["localhost", "127.0.0.1", "[::1]"].includes(
    url.hostname,
  );
  if (production && url.protocol !== "https:")
    throw new Error("A production SITE_URL must use HTTPS.");
  return { url: url.href, base: url.pathname, production };
}
const textExtensions = new Set([
  "",
  ".md",
  ".txt",
  ".json",
  ".jsonl",
  ".js",
  ".mjs",
  ".go",
  ".html",
  ".svg",
  ".xml",
  ".log",
  ".out",
  ".tape",
]);
export const archiveRawPath = (path: string) =>
  [".html", ".svg"].includes(extname(path)) ? path + ".txt" : path;
export const archiveViewPath = (path: string) =>
  textExtensions.has(extname(path)) ? path + ".html" : path;
function sourceViewer(path: string, bytes: Buffer, base: string) {
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="robots" content="noindex,follow"><title>${escape(path)} · koru source</title><style>:root{color-scheme:light dark}body{font:15px/1.6 system-ui;margin:24px;max-width:1100px}a{color:LinkText}pre{white-space:pre-wrap;overflow-wrap:anywhere;font:13px/1.7 monospace}h1{font-size:22px;overflow-wrap:anywhere}</style></head><body><p><a href="${base}en/sources/">Documentation archive</a> · <a href="${base}es/sources/">Archivo de documentación</a></p><h1>${escape(path)}</h1><p>Original recorded source; its language and historical claims are preserved. <a href="${base}original/${archiveRawPath(path)}" download="${escape(posix.basename(path))}">Download original file</a></p><pre>${escape(bytes.toString("utf8"))}</pre></body></html>`;
}
function archiveLinks(markdown: string, base: string, originals: Set<string>) {
  return markdown.replace(
    /(\]\(|href=")(\/[^\s"')]+)(?=[\s"')])/g,
    (all, prefix, href) => {
      const start = base + "original/";
      if (!href.startsWith(start)) return all;
      const path = href.slice(start.length);
      return originals.has(path) ? prefix + start + archiveViewPath(path) : all;
    },
  );
}
export const headingId = (value: string) =>
  value
    .toLowerCase()
    .replace(/<[^>]*>/g, "")
    .replace(/[^\p{L}\p{N}\s-]/gu, "")
    .trim()
    .replace(/\s+/g, "-");
export async function listFiles(path: string): Promise<string[]> {
  const entries = await readdir(path, { withFileTypes: true });
  const nested = await Promise.all(
    entries.map((entry) =>
      entry.isDirectory()
        ? listFiles(resolve(path, entry.name))
        : [resolve(path, entry.name)],
    ),
  );
  return nested.flat().sort();
}
export async function originalFiles() {
  const docs = await listFiles(resolve(repo, "docs"));
  return [
    ...docs,
    ...[
      "README.md",
      "AGENTS.md",
      "LICENSE",
      "tests/README.md",
      "runtime/native/README.md",
      "examples/web/README.md",
      "examples/web/index.html",
      "examples/web/second.html",
      "examples/web/app.js",
      "examples/web/style.css",
      "examples/web/icon.svg",
      "examples/dex/answer.json",
      "examples/dex/main.go",
      "examples/native/plugin.json",
      "packages/javascript/core.js",
      "packages/javascript/core.mjs",
      "website/README.md",
      "website/DESIGN.md",
      "website/VALIDATION.md",
      "website/DEPLOYMENT.md",
    ].map((path) => resolve(repo, path)),
  ].sort();
}
export function normalizeLinks(
  markdown: string,
  source: string,
  base: string,
  lang: Locale,
  originals: Set<string>,
) {
  return markdown
    .replace(
      /(!?\[[^\]]*\]\()(<[^>]+>|[^\s)]+)([^)]*\))/g,
      (all, start, href, suffix) => {
        if (href.startsWith("{{base}}")) return all.replace("{{base}}", base);
        if (/^<?(?:[a-z]+:|\/|#)/i.test(href)) return all;
        href = href.replace(/^<|>$/g, "");
        const [file, anchor] = href.split("#");
        const target = posix.normalize(posix.join(posix.dirname(source), file));
        const page = pages.find(
          (page) => "source" in page && page.source === target,
        );
        if (page)
          return `${start}${route(base, lang, page.slug)}${anchor ? "?id=" + anchor : ""}${suffix}`;
        if (!originals.has(target))
          throw new Error(`Unresolved documentation link ${href} in ${source}`);
        return `${start}${base}original/${target}${anchor ? "#" + anchor : ""}${suffix.replace(/\)$/, " ':ignore')")}`;
      },
    )
    .replaceAll("{{base}}", base);
}
function sidebar(base: string, lang: Locale) {
  const index = locales.indexOf(lang);
  return (
    text[lang].groups
      .map(
        (group, groupIndex) =>
          `- ${group}\n${pages
            .filter((page) => page.group === groupIndex)
            .map(
              (page) =>
                `  - [<i class="ph-duotone ph-${page.icon}" aria-hidden="true"></i> ${page.titles[index]}](${route(base, lang, page.slug)})`,
            )
            .join("\n")}`,
      )
      .join("\n\n") +
    `\n\n<div class="sidebar-donate-section"><a class="sidebar-donate" href="https://ko-fi.com/R5R81YESSI" target="_blank" rel="noopener noreferrer"><span class="donate-icon" aria-hidden="true"><i class="ph-duotone ph-coffee"></i><i class="ph-duotone ph-heart"></i></span><span>${text[lang].donate}</span></a></div>`
  );
}
const renderer = new Marked({
  renderer: {
    heading({ tokens, depth }) {
      const html = this.parser.parseInline(tokens);
      return `<h${depth} id="${escape(headingId(html))}">${html}</h${depth}>\n`;
    },
  },
});
function addPageContents(html: string, label: string) {
  const headings = [...html.matchAll(/<h2 id="([^"]+)">([\s\S]*?)<\/h2>/g)];
  if (headings.length < 3) return html;
  const links = headings
    .map(([, id, title]) => `<li><a href="#${id}">${title}</a></li>`)
    .join("");
  const contents = `<details class="page-toc"><summary>${escape(label)}</summary><nav aria-label="${escape(label)}"><ul>${links}</ul></nav></details>`;
  return html.replace(headings[0][0], contents + headings[0][0]);
}
function template(
  content: string,
  sidebarHTML: string,
  lang: Locale,
  slug: string,
  title: string,
  description: string,
  config: ReturnType<typeof siteConfig>,
  manifest: unknown,
) {
  const t = text[lang];
  const canonical = new URL(route(config.base, lang, slug), config.url).href;
  const alternates = locales
    .map(
      (locale) =>
        `<link rel="alternate" hreflang="${locale}" href="${new URL(route(config.base, locale, slug), config.url).href}">`,
    )
    .join("\n");
  const localeLinks = locales
    .map(
      (locale) =>
        `<a href="${route(config.base, locale, slug)}" lang="${locale}" data-locale="${locale}" ${lang === locale ? 'aria-current="true"' : ""}>${locale === "en" ? "English" : "Español"}</a>`,
    )
    .join("");
  const data = JSON.stringify({
    base: config.base,
    siteURL: config.url,
    production: config.production,
    lang,
    slug,
    pages: manifest,
    text,
  }).replaceAll("<", "\\u003c");
  const structured = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "TechArticle",
    headline: title,
    description,
    inLanguage: lang,
    url: canonical,
    isPartOf: {
      "@type": "WebSite",
      name: "koru documentation",
      url: config.url,
    },
  }).replaceAll("<", "\\u003c");
  return `<!doctype html>
<html lang="${lang}"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>${escape(title)} · koru</title><meta name="description" content="${escape(description)}">
<meta name="theme-color" content="#F7F9FD"><meta name="robots" content="${config.production ? "index,follow" : "noindex,follow"}">
<link rel="canonical" href="${canonical}">${alternates}
<link rel="alternate" hreflang="x-default" href="${new URL(route(config.base, "en", slug), config.url).href}">
<meta property="og:type" content="article"><meta property="og:site_name" content="koru"><meta property="og:title" content="${escape(title)} · koru"><meta property="og:description" content="${escape(description)}"><meta property="og:url" content="${canonical}"><meta property="og:locale" content="${lang === "en" ? "en_US" : "es_ES"}">
<meta name="twitter:card" content="summary"><meta name="twitter:title" content="${escape(title)} · koru"><meta name="twitter:description" content="${escape(description)}">
<link rel="icon" href="${config.base}assets/favicon.svg" type="image/svg+xml">
<link rel="preload" as="font" type="font/woff2" href="${config.base}assets/fonts/nunito-latin-wght-normal.woff2" crossorigin>
<script src="${config.base}assets/theme.js"></script>
<link rel="stylesheet" href="${config.base}assets/vendor/theme.css"><link rel="stylesheet" href="${config.base}assets/vendor/phosphor/style.css"><link rel="stylesheet" href="${config.base}assets/theme.css">
<script type="application/ld+json" id="structured-data">${structured}</script>
<script id="site-config" type="application/json">${data}</script>
</head><body data-page="${slug}">
<a class="skip-link" href="#main">${lang === "en" ? "Skip to content" : "Saltar al contenido"}</a>
<header class="site-header"><a class="brand" href="${route(config.base, lang, "overview")}" aria-label="koru ${t.docs}"><i class="ph-duotone ph-spiral" aria-hidden="true"></i><span>koru</span></a><span class="header-label">${t.docs}</span><nav class="language-nav" aria-label="${t.language}">${localeLinks}</nav><div class="theme-control"><button class="theme-button" data-theme-button data-preference="system" type="button" aria-label="${t.appearance}" aria-haspopup="menu" aria-expanded="false" aria-controls="theme-menu"><i class="ph-duotone ph-desktop" data-theme-icon aria-hidden="true"></i><span data-theme-label>${t.systemTheme}</span><i class="ph-duotone ph-caret-down theme-caret" aria-hidden="true"></i></button><div class="theme-menu" id="theme-menu" data-theme-menu role="menu" aria-label="${t.appearance}" hidden>${[
    ["system", t.systemTheme, "desktop"],
    ["light", t.lightTheme, "sun"],
    ["dark", t.darkTheme, "moon"],
  ]
    .map(
      ([value, label, icon]) =>
        `<button type="button" role="menuitemradio" tabindex="-1" aria-checked="${value === "system"}" data-theme-choice="${value}" data-label="${label}"><i class="ph-duotone ph-${icon}" aria-hidden="true"></i><span>${label}</span><i class="ph-duotone ph-check theme-check" aria-hidden="true"></i></button>`,
    )
    .join(
      "",
    )}</div></div><a class="source-link" href="${route(config.base, lang, "sources")}"><i class="ph-duotone ph-code-block" aria-hidden="true"></i><span>${t.source}</span></a></header>
<div id="app"><aside class="sidebar" aria-label="${t.docs}"><div class="sidebar-nav">${sidebarHTML}</div></aside><main class="content"><article class="markdown-section" id="main" tabindex="-1">${content}<footer class="doc-footer">${t.footer}<span>${t.note}</span></footer></article></main></div>
<noscript><style>.sidebar{padding-top:28px}.sidebar-toggle,.theme-control{display:none}@media(max-width:767px){.sidebar{position:static!important;width:auto;height:auto!important;transform:none!important}.content{margin-left:0}.sidebar-nav{display:none}.sidebar::before{content:'${lang === "en" ? "JavaScript is off. Document links and language switching still work." : "JavaScript está desactivado. Los enlaces y el cambio de idioma siguen funcionando."}';font-size:14px}.site-header{position:sticky}.content{padding-top:15px!important}}</style></noscript>
<script src="${config.base}assets/site.js"></script>
<script src="${config.base}assets/vendor/docsify.min.js"></script>
<script src="${config.base}assets/vendor/search.min.js"></script>
<script src="${config.base}assets/vendor/copy.min.js"></script>
<script src="${config.base}assets/vendor/prism-go.min.js"></script>
<script src="${config.base}assets/vendor/prism-json.min.js"></script>
<script src="${config.base}assets/vendor/prism-bash.min.js"></script>
</body></html>`;
}
export async function build(
  options: { siteURL?: string; outDir?: string } = {},
) {
  const config = siteConfig(options.siteURL);
  const dist = resolve(root, options.outDir || "dist");
  await rm(dist, { recursive: true, force: true });
  await mkdir(dist, { recursive: true });
  await writeFile(
    resolve(dist, "go.mod"),
    "module koru.dev/documentation-artifact\n\ngo 1.23\n",
  );
  await writeFile(resolve(dist, ".nojekyll"), "");
  const originals = await originalFiles();
  const originalSet = new Set(originals.map((file) => relative(repo, file)));
  const hashes: Record<string, string> = {};
  for (const file of originals) {
    const path = relative(repo, file);
    const data = await readFile(file);
    hashes[path] = createHash("sha256").update(data).digest("hex");
    await mkdir(dirname(resolve(dist, "original", path)), { recursive: true });
    await writeFile(resolve(dist, "original", archiveRawPath(path)), data);
    if (archiveViewPath(path) !== path)
      await writeFile(
        resolve(dist, "original", archiveViewPath(path)),
        sourceViewer(path, data, config.base),
      );
  }
  await writeFile(
    resolve(dist, "original/manifest.json"),
    JSON.stringify(hashes, null, 2) + "\n",
  );
  const assets = resolve(dist, "assets");
  await cp(resolve(root, "assets"), assets, { recursive: true });
  await cp(resolve(root, "assets/favicon.ico"), resolve(dist, "favicon.ico"));
  const vendors: Record<string, string> = {
    "docsify/lib/docsify.min.js": "docsify.min.js",
    "docsify/lib/plugins/search.min.js": "search.min.js",
    "docsify-copy-code/dist/docsify-copy-code.min.js": "copy.min.js",
    "docsify-themeable/dist/css/theme-simple.css": "theme.css",
    "prismjs/components/prism-go.min.js": "prism-go.min.js",
    "prismjs/components/prism-json.min.js": "prism-json.min.js",
    "prismjs/components/prism-bash.min.js": "prism-bash.min.js",
  };
  await mkdir(resolve(assets, "vendor"), { recursive: true });
  for (const [source, destination] of Object.entries(vendors))
    await cp(
      resolve(root, "node_modules", source),
      resolve(assets, "vendor", destination),
    );
  const phosphor = resolve(
    root,
    "node_modules/@phosphor-icons/web/src/duotone",
  );
  await mkdir(resolve(assets, "vendor/phosphor"), { recursive: true });
  await cp(
    resolve(phosphor, "Phosphor-Duotone.woff2"),
    resolve(assets, "vendor/phosphor/Phosphor-Duotone.woff2"),
  );
  // Keep the upstream icon definitions, shipping only the modern font format.
  const iconCSS = (
    await readFile(resolve(phosphor, "style.css"), "utf8")
  ).replace(
    /src:[\s\S]*?;/,
    'src: url("./Phosphor-Duotone.woff2") format("woff2");',
  );
  await writeFile(resolve(assets, "vendor/phosphor/style.css"), iconCSS);
  await mkdir(resolve(assets, "fonts"), { recursive: true });
  for (const name of [
    "nunito-latin-wght-normal.woff2",
    "nunito-latin-ext-wght-normal.woff2",
  ])
    await cp(
      resolve(root, "node_modules/@fontsource-variable/nunito/files", name),
      resolve(assets, "fonts", name),
    );
  await cp(
    resolve(
      root,
      "node_modules/@fontsource/jetbrains-mono/files/jetbrains-mono-latin-400-normal.woff2",
    ),
    resolve(assets, "fonts/jetbrains-mono.woff2"),
  );
  const licenses = {
    docsify: "LICENSE",
    "docsify-themeable": "LICENSE",
    "docsify-copy-code": "LICENSE",
    "@phosphor-icons/web": "LICENSE",
    "@fontsource-variable/nunito": "LICENSE",
    "@fontsource/jetbrains-mono": "LICENSE",
    marked: "LICENSE.md",
    prismjs: "LICENSE",
  };
  await mkdir(resolve(assets, "licenses"), { recursive: true });
  for (const [pkg, file] of Object.entries(licenses))
    await cp(
      resolve(root, "node_modules", pkg, file),
      resolve(assets, "licenses", pkg.replaceAll("/", "-") + ".txt"),
    );
  const manifest = Object.fromEntries(
    pages.map((page) => [
      page.slug,
      { titles: page.titles, descriptions: page.descriptions },
    ]),
  );
  for (const lang of locales) {
    const sidebarMD = sidebar(config.base, lang);
    await mkdir(resolve(dist, lang), { recursive: true });
    await writeFile(resolve(dist, lang, "_sidebar.md"), sidebarMD);
    for (const page of pages) {
      let markdown: string;
      if (page.slug === "sources") {
        const documentation = originals.filter(
          (file) =>
            file.endsWith(".md") &&
            !relative(repo, file).startsWith("website/"),
        );
        const raw = originals.filter((file) => !file.endsWith(".md"));
        markdown =
          lang === "en"
            ? "# Documentation archive\n\nEvery original project document and validation artifact is available below. The guides provide English and Spanish reading paths; originals preserve their recorded language and historical context. They are copied during the website build, with SHA-256 provenance.\n\n## Original documents\n\n"
            : "# Archivo de documentación\n\nAquí están todos los documentos originales y artefactos de validación. Las guías ofrecen recorridos en inglés y español; los originales conservan su idioma y contexto histórico. Se copian al construir el sitio, con procedencia SHA-256.\n\n## Documentos originales\n\n";
        markdown += documentation
          .map((file) => {
            const path = relative(repo, file);
            return `- [${path}](${config.base}original/${path} ':ignore')`;
          })
          .join("\n");
        markdown +=
          lang === "en"
            ? "\n\n## Provenance and official specifications\n\n"
            : "\n\n## Procedencia y especificaciones oficiales\n\n";
        markdown += `[SHA-256 manifest](${config.base}original/manifest.json ':ignore') · [Android specification sources](${config.base}original/docs/specs/sources.json ':ignore')\n\n`;
        const specs = JSON.parse(
          await readFile(resolve(repo, "docs/specs/sources.json"), "utf8"),
        );
        markdown += specs
          .map(
            (spec: { name: string; url: string }) =>
              `- [${spec.name}](${spec.url})`,
          )
          .join("\n");
        markdown +=
          lang === "en"
            ? "\n\n## Raw validation records and examples\n\n"
            : "\n\n## Registros de validación y ejemplos\n\n";
        markdown += raw
          .map((file) => {
            const path = relative(repo, file);
            return `- [${path}](${config.base}original/${path} ':ignore')`;
          })
          .join("\n");
      } else {
        const source = "source" in page ? page.source : undefined;
        const local = resolve(root, "content", lang, page.slug + ".md");
        if (lang === "es" && source && !(await Bun.file(local).exists()))
          markdown = await readFile(resolve(repo, source), "utf8");
        else markdown = await readFile(local, "utf8");
        if (lang === "es" && page.slug === "dex-results")
          markdown =
            "> Registro histórico anterior al builder APK. Las afirmaciones de ausencia de ejecución corresponden a esa etapa. Consulta la [evidencia Android posterior]({{base}}es/android-results/).\n\n" +
            markdown;
        if (lang === "es" && page.slug === "architecture-dex")
          markdown =
            "> Decisión inicial conservada como registro histórico. El runtime normal posterior usa HTTPS interceptado; véase [ADR 0002]({{base}}es/architecture-apk/).\n\n" +
            markdown;
        if (source)
          markdown = normalizeLinks(
            markdown,
            source,
            config.base,
            lang,
            originalSet,
          );
        else markdown = markdown.replaceAll("{{base}}", config.base);
      }
      markdown = archiveLinks(markdown, config.base, originalSet);
      const index = locales.indexOf(lang);
      const parsed = await renderer.parse(markdown);
      const html = addPageContents(parsed, text[lang].contents)
        .replace(
          /<pre><code class="language-([^"]+)">/g,
          '<pre tabindex="0" data-lang="$1"><code class="language-$1">',
        )
        .replaceAll("<pre>", '<pre tabindex="0">')
        .replaceAll("<table>", '<table tabindex="0">');
      const folder = resolve(
        dist,
        lang,
        page.slug === "overview" ? "" : page.slug,
      );
      await mkdir(folder, { recursive: true });
      // Docsify's history router requests README.md for directory-shaped routes.
      await writeFile(resolve(folder, "README.md"), markdown);
      await writeFile(
        resolve(folder, "index.html"),
        template(
          html,
          await renderer.parse(sidebarMD),
          lang,
          page.slug,
          page.titles[index],
          page.descriptions[index],
          config,
          manifest,
        ),
      );
    }
  }
  const landing = await readFile(resolve(dist, "en/index.html"), "utf8");
  await writeFile(
    resolve(dist, "index.html"),
    landing.replace(
      "<head>",
      `<head><meta http-equiv="refresh" content="0;url=${route(config.base, "en", "overview")}">`,
    ),
  );
  await writeFile(
    resolve(dist, "README.md"),
    await readFile(resolve(dist, "en/README.md")),
  );
  await writeFile(resolve(dist, "_sidebar.md"), sidebar(config.base, "en"));
  const notFound =
    '<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="robots" content="noindex"><title>Page not found · koru</title></head><body><h1>Page not found / Página no encontrada</h1><p><a href="' +
    route(config.base, "en", "overview") +
    '">English documentation</a> · <a href="' +
    route(config.base, "es", "overview") +
    '">Documentación en español</a></p></body></html>';
  await writeFile(resolve(dist, "404.html"), notFound);
  await writeFile(
    resolve(dist, "_404.md"),
    "# Page not found / Página no encontrada\n\n[English](" +
      route(config.base, "en", "overview") +
      ") · [Español](" +
      route(config.base, "es", "overview") +
      ")",
  );
  const sitemap = locales
    .flatMap((lang) =>
      pages.map(
        (page) =>
          `<url><loc>${escape(new URL(route(config.base, lang, page.slug), config.url).href)}</loc></url>`,
      ),
    )
    .join("\n");
  await writeFile(
    resolve(dist, "sitemap.xml"),
    `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${sitemap}\n</urlset>\n`,
  );
  await writeFile(
    resolve(dist, "robots.txt"),
    config.production
      ? `# Subpath hosts must merge these directives into the domain-root /robots.txt.\nUser-agent: *\nAllow: /\nDisallow: ${config.base}original/\nSitemap: ${new URL("sitemap.xml", config.url).href}\n`
      : "User-agent: *\nDisallow: /\n",
  );
  await writeFile(
    resolve(dist, "site.json"),
    JSON.stringify(
      { ...config, pages: pages.length * locales.length },
      null,
      2,
    ),
  );
  console.log(
    `Built ${pages.length * locales.length} localized pages and ${originals.length} original files in website/${relative(root, dist)}. ${config.production ? "Production SEO enabled." : "Local preview is noindex; set SITE_URL for production SEO."}`,
  );
}
if (import.meta.main) await build();
