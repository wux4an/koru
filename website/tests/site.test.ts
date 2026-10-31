import { beforeAll, describe, expect, test } from "bun:test";
import { readFile } from "node:fs/promises";
import { resolve, relative } from "node:path";
import { createHash } from "node:crypto";
import {
  build,
  root,
  siteConfig,
  listFiles,
  normalizeLinks,
  archiveRawPath,
} from "../scripts/build";
import { locales, pages, route } from "../scripts/catalog";
const output = "test-results/unit-build";
const canonical = "https://docs.koru.test/handbook/";
beforeAll(async () => {
  await build({ siteURL: canonical, outDir: output });
});

describe("localized, crawlable documentation", () => {
  test("every catalog page has both locales, real text and reciprocal SEO metadata", async () => {
    for (const lang of locales) {
      for (const page of pages) {
        const folder = resolve(
          root,
          output,
          "docs",
          lang,
          page.slug === "overview" ? "" : page.slug,
        );
        const html = await readFile(resolve(folder, "index.html"), "utf8");
        const markdown = await readFile(resolve(folder, "README.md"), "utf8");
        expect(markdown.length).toBeGreaterThan(300);
        expect(html).toContain(`<html lang="${lang}">`);
        expect(html).toContain(
          `rel="canonical" href="${canonical}docs/${lang}/${page.slug === "overview" ? "" : page.slug + "/"}"`,
        );
        expect(html).toContain('content="index,follow"');
        expect(html.match(/<h1\b/g)?.length).toBe(1);
        expect(html).not.toContain("{{base}}");
        for (const other of locales)
          expect(html).toContain(
            `hreflang="${other}" href="${canonical}docs/${other}/${page.slug === "overview" ? "" : page.slug + "/"}"`,
          );
        expect(html).toContain('hreflang="x-default"');
        expect(html).toContain('"@type":"TechArticle"');
        expect(html).not.toMatch(
          /<script[^>]+src="https?:\/\/|<link[^>]+rel="stylesheet"[^>]+href="https?:\/\//,
        );
      }
    }
  });
  test("all repository documentation is preserved with correct hashes", async () => {
    const manifest = JSON.parse(
      await readFile(resolve(root, output, "original/manifest.json"), "utf8"),
    );
    const repo = resolve(root, "..");
    const markdown = (await listFiles(resolve(repo, "docs")))
      .filter((file) => file.endsWith(".md"))
      .map((file) => relative(repo, file));
    for (const file of [
      ...markdown,
      "README.md",
      "AGENTS.md",
      "tests/README.md",
      "runtime/native/README.md",
      "examples/web/README.md",
    ])
      expect(manifest[file]).toBeDefined();
    for (const [path, hash] of Object.entries(manifest)) {
      const original = await readFile(resolve(repo, path));
      const archived = await readFile(
        resolve(root, output, "original", archiveRawPath(path)),
      );
      expect(archived.equals(original)).toBe(true);
      expect(createHash("sha256").update(archived).digest("hex")).toBe(hash);
      expect(path).not.toMatch(/(?:\.koru|android-demo|\.pem$|\.jks$|\.p12$)/);
    }
  });
  test("sitemap covers every translated page once and excludes raw evidence", async () => {
    const sitemap = await readFile(
      resolve(root, output, "sitemap.xml"),
      "utf8",
    );
    expect(sitemap.match(/<loc>/g)?.length).toBe(
      pages.length * locales.length + locales.length + 1,
    );
    expect(sitemap).toContain(`<loc>${canonical}en/</loc>`);
    expect(sitemap).toContain(`<loc>${canonical}es/</loc>`);
    for (const lang of locales)
      for (const page of pages)
        expect(sitemap).toContain(
          new URL(route("/handbook/", lang, page.slug), canonical).href,
        );
    expect(sitemap).not.toContain("/original/");
    // Viewers carry noindex, so crawlers must be allowed to read it.
    const robots = await readFile(resolve(root, output, "robots.txt"), "utf8");
    expect(robots).not.toContain("Disallow");
    expect(robots).toContain(`Sitemap: ${canonical}sitemap.xml`);
  });
  test("every local document link points to a generated route or archived file", async () => {
    for (const file of (await listFiles(resolve(root, output))).filter(
      (file) => file.endsWith("index.html") && !file.includes("/original/"),
    )) {
      const html = await readFile(file, "utf8");
      const links = [
        ...html.matchAll(/href="(\/handbook\/[^"?#]*)[^"]*"/g),
      ].map((match) => match[1].slice("/handbook/".length));
      for (const link of links) {
        const target = resolve(
          root,
          output,
          link,
          link === "" || link.endsWith("/") ? "index.html" : "",
        );
        if (!(await Bun.file(target).exists()))
          throw new Error(`${relative(root, file)} links to missing ${link}`);
      }
    }
  });
  test("long guides include a localized, working section index", async () => {
    for (const lang of locales) {
      for (const page of pages) {
        const folder = resolve(
          root,
          output,
          "docs",
          lang,
          page.slug === "overview" ? "" : page.slug,
        );
        const html = await readFile(resolve(folder, "index.html"), "utf8");
        const headings = [...html.matchAll(/<h2 id="([^"]+)">/g)].map(
          ([, id]) => id,
        );
        const toc = html.match(
          /<details class="page-toc">([\s\S]*?)<\/details>/,
        );
        if (headings.length < 3) {
          expect(toc).toBeNull();
          continue;
        }
        expect(toc).not.toBeNull();
        expect(toc?.[1]).toContain(
          lang === "en" ? "On this page" : "En esta página",
        );
        const links = [...(toc?.[1] || "").matchAll(/href="#([^"]+)"/g)].map(
          ([, id]) => id,
        );
        expect(links).toEqual(headings);
      }
    }
  });
  test("early evidence is explicitly historical in both locales", async () => {
    for (const lang of locales) {
      const content = await readFile(
        resolve(root, output, "docs", lang, "dex-results/README.md"),
        "utf8",
      );
      expect(content).toContain(`/handbook/docs/${lang}/android-results/`);
      expect(content).toMatch(/Historical|histórico/);
    }
  });
});

describe("deployment configuration and broken input", () => {
  test("custom-domain artifact uses root paths and serves a root favicon", async () => {
    const domainOutput = "test-results/domain-build";
    await build({ siteURL: "https://koru.js.org/", outDir: domainOutput });
    const index = await readFile(
      resolve(root, domainOutput, "index.html"),
      "utf8",
    );
    const english = await readFile(
      resolve(root, domainOutput, "docs/en/index.html"),
      "utf8",
    );
    const notFound = await readFile(
      resolve(root, domainOutput, "404.html"),
      "utf8",
    );
    const icon = await readFile(resolve(root, domainOutput, "favicon.ico"));
    // The domain root is a real page and never redirects by language.
    expect(index).not.toContain("http-equiv");
    expect(index).toContain('href="/en/"');
    expect(index).toContain('href="/es/"');
    const landingEN = await readFile(
      resolve(root, domainOutput, "en/index.html"),
      "utf8",
    );
    expect(landingEN).toContain(
      'rel="canonical" href="https://koru.js.org/en/"',
    );
    expect(landingEN).toContain('hreflang="es" href="https://koru.js.org/es/"');
    expect(landingEN.match(/<h1\b/g)?.length).toBe(1);
    expect(landingEN).toContain('href="/docs/en/getting-started/"');
    const spanish = await readFile(
      resolve(root, domainOutput, "es/index.html"),
      "utf8",
    );
    expect(spanish).toContain('<html lang="es">');
    expect(spanish).toContain('href="/docs/es/getting-started/"');
    const schema = JSON.parse(
      await readFile(
        resolve(root, domainOutput, "schema/koru.schema.json"),
        "utf8",
      ),
    );
    expect(schema.$id).toBe("https://koru.js.org/schema/koru.schema.json");
    expect(english).toContain(
      'rel="canonical" href="https://koru.js.org/docs/en/"',
    );
    expect(english).toContain('href="/assets/favicon.svg"');
    expect(notFound).toContain('href="/docs/en/"');
    expect(notFound).toContain('name="viewport"');
    expect(notFound).toContain('name="robots" content="noindex"');
    expect(notFound).toContain('href="/docs/es/"');
    expect(index).not.toContain("/koru/en/");
    expect(icon.subarray(0, 4).equals(Buffer.from([0, 0, 1, 0]))).toBe(true);
  });
  test("normalizes the domain and subpath without guessing a production domain", () => {
    expect(siteConfig("https://docs.koru.test/manual").base).toBe("/manual/");
    expect(siteConfig("http://localhost:4173").production).toBe(false);
    expect(siteConfig("https://docs.koru.test").production).toBe(true);
  });
  test("rejects non-HTTPS public sites and ambiguous canonical URLs", () => {
    for (const url of [
      "http://docs.koru.test/",
      "file:///tmp/site",
      "https://user:pass@docs.koru.test",
      "https://docs.koru.test/?x=1",
      "https://docs.koru.test/#foo",
    ])
      expect(() => siteConfig(url)).toThrow();
  });
  test("broken relative source links fail instead of receiving a fallback", () => {
    expect(() =>
      normalizeLinks(
        "[broken](missing.md)",
        "docs/guide.md",
        "/",
        "en",
        new Set(),
      ),
    ).toThrow("Unresolved documentation link");
  });
});

describe("llms.txt for language models", () => {
  const read = (path: string) => readFile(resolve(root, output, path), "utf8");
  const exists = async (url: string) => {
    expect(url.startsWith(canonical)).toBe(true);
    const path = url.slice(canonical.length);
    await readFile(
      resolve(root, output, path.endsWith("/") ? path + "index.html" : path),
    );
  };
  test("follows the llmstxt.org shape with an H1 first line and a summary", async () => {
    for (const file of ["llms.txt", "es/llms.txt"]) {
      const body = await read(file);
      expect(body.split("\n")[0]).toBe("# koru");
      expect(body).toMatch(/^> .+/m);
      expect(body).toContain("```sh\nkoru create");
      expect(body).toContain("\n## Optional\n");
      expect(body).toContain("## ");
    }
  });
  test("every URL is absolute and points at a file in the build", async () => {
    for (const file of ["llms.txt", "es/llms.txt", "llms-full.txt"]) {
      const urls = [...(await read(file)).matchAll(/\]\(([^)\s]+)\)/g)].map(
        (m) => m[1],
      );
      expect(urls.length).toBeGreaterThan(10);
      for (const url of urls) {
        if (file === "llms-full.txt") {
          expect(url.startsWith("/handbook/")).toBe(false);
          if (!url.startsWith(canonical)) continue;
        } else expect(url).toMatch(/^https:\/\//);
        if (url.startsWith(canonical)) await exists(url.split(/[?#]/)[0]);
      }
    }
  });
  test("both indexes list every catalog page except the archive", async () => {
    for (const [file, lang] of [
      ["llms.txt", "en"],
      ["es/llms.txt", "es"],
    ] as const) {
      const body = await read(file);
      for (const page of pages) {
        const url = `${canonical}${route("/", lang, page.slug).slice(1)}README.md`;
        expect(body.includes(`(${url})`)).toBe(page.slug !== "sources");
      }
    }
  });
  test("llms-full.txt carries the page text and Optional lists internals", async () => {
    const full = await read("llms-full.txt");
    const guide = await read("docs/en/getting-started/README.md");
    expect(full).toContain("Install koru");
    expect(full).toContain(
      guide
        .split("\n")
        .find((l) => l.startsWith("## 1."))!
        .replace(/^## /, "### "),
    );
    expect(full).not.toContain('<div class="hero">');
    const optional = (await read("llms.txt")).split("\n## Optional\n")[1];
    expect(optional).toContain("/docs/en/architecture-dex/README.md");
    expect(optional).toContain("/docs/en/milestones/README.md");
    expect(optional).not.toContain("/docs/en/cli/README.md");
  });
  test("pages and landings advertise llms.txt", async () => {
    expect(await read("en/index.html")).toContain(
      `rel="alternate" type="text/plain" href="${canonical}llms.txt" title="llms.txt"`,
    );
    expect(await read("es/index.html")).toContain(
      `href="${canonical}es/llms.txt"`,
    );
    expect(await read("docs/es/cli/index.html")).toContain(
      `href="${canonical}es/llms.txt"`,
    );
  });
});

describe("SEO lint on the production build", () => {
  const site = "https://koru.js.org/";
  const out = "test-results/seo-build";
  const hasGit = Bun.spawnSync(["git", "-C", root, "log", "-1"]).exitCode === 0;
  type Page = {
    file: string;
    url: string;
    html: string;
    robots: string;
    canonical?: string;
  };
  const meta = (html: string, attr: string, name: string) =>
    html.match(new RegExp(`<meta ${attr}="${name}" content="([^"]*)"`))?.[1];
  const unescape = (value = "") =>
    value.replaceAll("&amp;", "&").replaceAll("&quot;", '"');
  const pagesByURL = new Map<string, Page>();
  const legacy: Page[] = [];
  beforeAll(async () => {
    await build({ siteURL: site, outDir: out });
    const dist = resolve(root, out);
    for (const file of await listFiles(dist)) {
      if (!file.endsWith("index.html") || file.includes("/original/")) continue;
      const html = await readFile(file, "utf8");
      const path = relative(dist, file).replace(/index\.html$/, "");
      const page: Page = {
        file,
        url: site + path,
        html,
        robots: meta(html, "name", "robots") || "",
        canonical: html.match(/<link rel="canonical" href="([^"]+)"/)?.[1],
      };
      if (html.includes('http-equiv="refresh"')) legacy.push(page);
      else pagesByURL.set(page.url, page);
    }
  });
  const indexable = () =>
    [...pagesByURL.values()].filter((page) => !page.robots.includes("noindex"));
  test("titles and descriptions are unique and a sensible length", () => {
    const titles = new Set<string>();
    const descriptions = new Set<string>();
    for (const page of indexable()) {
      const title = unescape(page.html.match(/<title>([^<]*)<\/title>/)?.[1]);
      const description = unescape(meta(page.html, "name", "description"));
      expect(title.length, page.url).toBeGreaterThan(8);
      expect(title.length, page.url).toBeLessThan(60);
      expect(description.length, page.url).toBeGreaterThanOrEqual(50);
      expect(description.length, page.url).toBeLessThanOrEqual(160);
      expect(titles.has(title), `duplicate title ${title}`).toBe(false);
      expect(descriptions.has(description), `duplicate ${page.url}`).toBe(
        false,
      );
      titles.add(title);
      descriptions.add(description);
    }
    expect(titles.size).toBe(pages.length * locales.length + 3);
  });
  test("indexable pages have one h1 and a self-referential canonical", () => {
    for (const page of indexable()) {
      expect(page.html.match(/<h1\b/g)?.length, page.url).toBe(1);
      expect(page.canonical, page.url).toBe(page.url);
    }
  });
  test("hreflang annotations are reciprocal and point at indexable pages", () => {
    for (const page of indexable()) {
      const alternates = [
        ...page.html.matchAll(
          /<link rel="alternate" hreflang="([^"]+)" href="([^"]+)"/g,
        ),
      ].map(([, lang, href]) => [lang, href]);
      expect(alternates.length, page.url).toBeGreaterThanOrEqual(3);
      expect(alternates.map(([lang]) => lang).sort(), page.url).toEqual([
        "en",
        "es",
        "x-default",
      ]);
      for (const [lang, href] of alternates) {
        const target = pagesByURL.get(href);
        expect(target, `${page.url} -> ${href}`).toBeDefined();
        expect(target!.robots).not.toContain("noindex");
        // The root is a language selector, not a translation of either page.
        if (lang === "x-default" || page.url === site) continue;
        // The target must list this page back under its own language.
        const back = [
          ...target!.html.matchAll(
            /<link rel="alternate" hreflang="([^"]+)" href="([^"]+)"/g,
          ),
        ].map(([, l, h]) => `${l} ${h}`);
        expect(back, `${href} lacks a link to ${page.url}`).toContain(
          `${page.html.match(/<html lang="(\w+)"/)![1]} ${page.url}`,
        );
      }
    }
    expect(
      pagesByURL
        .get(site)!
        .html.includes('hreflang="x-default" href="https://koru.js.org/"'),
    ).toBe(true);
    for (const lang of locales)
      expect(pagesByURL.get(`${site}${lang}/`)!.html).toContain(
        'hreflang="x-default" href="https://koru.js.org/"',
      );
  });
  test("every indexable page has an absolute, existing og:image and twitter card", async () => {
    for (const page of indexable()) {
      const image = meta(page.html, "property", "og:image")!;
      expect(image, page.url).toMatch(
        /^https:\/\/koru\.js\.org\/assets\/og-(en|es)\.png$/,
      );
      expect(meta(page.html, "name", "twitter:image")).toBe(image);
      expect(meta(page.html, "name", "twitter:card")).toBe(
        "summary_large_image",
      );
      expect(meta(page.html, "property", "og:image:width")).toBe("1200");
      expect(meta(page.html, "property", "og:image:height")).toBe("630");
      expect(meta(page.html, "property", "og:image:alt")).toBeTruthy();
      const file = Bun.file(
        resolve(root, out, new URL(image).pathname.slice(1)),
      );
      expect(await file.exists(), image).toBe(true);
      expect(file.size).toBeLessThan(150 * 1024);
    }
  });
  test("JSON-LD parses into an @graph with the expected nodes", () => {
    for (const page of indexable()) {
      const raw = page.html.match(
        /<script type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/,
      )?.[1];
      expect(raw, page.url).toBeDefined();
      const graph = JSON.parse(raw!)["@graph"] as Record<string, any>[];
      const types = graph.map((node) => node["@type"]);
      expect(types, page.url).toContain("WebSite");
      expect(types, page.url).toContain("Organization");
      const organization = graph.find((n) => n["@type"] === "Organization")!;
      expect(organization.logo).toMatch(/^https:\/\/koru\.js\.org\/assets\//);
      if (page.url.includes("/docs/")) {
        expect(types).toContain("BreadcrumbList");
        const article = graph.find((n) => n["@type"] === "TechArticle")!;
        expect(article.inLanguage).toBe(page.url.split("/")[4]);
        expect(article.dateModified === undefined).toBe(!hasGit);
        if (hasGit)
          expect(meta(page.html, "property", "article:modified_time")).toBe(
            article.dateModified,
          );
        const crumbs = graph.find((n) => n["@type"] === "BreadcrumbList")!;
        expect(crumbs.itemListElement[0].position).toBe(1);
        expect(crumbs.itemListElement.length).toBe(
          page.url.endsWith("/docs/en/") || page.url.endsWith("/docs/es/")
            ? 2
            : 3,
        );
      } else if (/\/(en|es)\/$/.test(page.url)) {
        const app = graph.find((n) => n["@type"] === "SoftwareApplication")!;
        expect(app.softwareVersion).toMatch(/^\d+\.\d+\.\d+/);
        expect(app.aggregateRating).toBeUndefined();
        expect(app.sameAs).toContain("https://github.com/wux4an/koru");
      }
    }
  });
  test("the sitemap lists only indexable canonical pages, with lastmod and alternates", async () => {
    const sitemap = await readFile(resolve(root, out, "sitemap.xml"), "utf8");
    const entries = sitemap.split("<url>").slice(1);
    expect(entries.length).toBe(indexable().length);
    for (const entry of entries) {
      const loc = entry.match(/<loc>([^<]+)<\/loc>/)![1];
      const page = pagesByURL.get(loc);
      expect(page, loc).toBeDefined();
      expect(page!.robots).not.toContain("noindex");
      expect(page!.canonical).toBe(loc);
      const lastmod = entry.match(/<lastmod>([^<]+)<\/lastmod>/)?.[1];
      if (hasGit) {
        expect(lastmod, loc).toBeDefined();
        expect(Number.isNaN(Date.parse(lastmod!)), loc).toBe(false);
      } else expect(lastmod).toBeUndefined();
      expect(entry.match(/<xhtml:link /g)?.length, loc).toBe(3);
    }
    for (const page of legacy)
      expect(sitemap).not.toContain(`<loc>${page.url}`);
    expect(sitemap).not.toContain("<loc>https://koru.js.org/original/");
  });
  test("legacy URLs redirect to their canonical page and are not noindex", () => {
    expect(legacy.length).toBe((pages.length - 1) * locales.length + 1);
    for (const page of legacy) {
      expect(page.robots, page.url).toBe("");
      const target = page.canonical!;
      expect(target, page.url).toMatch(/^https:\/\/koru\.js\.org\//);
      expect(page.html).toContain(
        `content="0;url=${new URL(target).pathname}"`,
      );
      expect(page.html).toContain(
        `location.replace(${JSON.stringify(new URL(target).pathname)})`,
      );
      expect(pagesByURL.has(target), `${page.url} -> ${target}`).toBe(true);
    }
    expect(legacy.map((page) => page.url)).toContain(`${site}es/inicio/`);
  });
  test("the root is an indexable x-default page with visible language links", () => {
    const root_ = pagesByURL.get(site)!;
    expect(root_.robots).toBe("index,follow");
    expect(root_.html).not.toContain("http-equiv");
    for (const href of ["/en/", "/es/", "/docs/en/", "/docs/es/"])
      expect(root_.html).toContain(`<a class="button" href="${href}"`);
    expect(root_.html).not.toContain("location.replace");
    expect(root_.html).not.toContain("location.href");
  });
  test("verification tags, IndexNow key and robots follow the environment", async () => {
    const env = process.env;
    const saved = [
      env.GOOGLE_SITE_VERIFICATION,
      env.BING_SITE_VERIFICATION,
      env.YANDEX_VERIFICATION,
      env.INDEXNOW_KEY,
    ];
    try {
      env.GOOGLE_SITE_VERIFICATION = "g-token";
      env.BING_SITE_VERIFICATION = "b-token";
      env.YANDEX_VERIFICATION = "y-token";
      env.INDEXNOW_KEY = "abcdef1234567890";
      await build({ siteURL: site, outDir: "test-results/seo-verified" });
      const dist = resolve(root, "test-results/seo-verified");
      for (const file of ["index.html", "en/index.html", "es/index.html"]) {
        const html = await readFile(resolve(dist, file), "utf8");
        expect(html).toContain(
          'name="google-site-verification" content="g-token"',
        );
        expect(html).toContain('name="msvalidate.01" content="b-token"');
        expect(html).toContain('name="yandex-verification" content="y-token"');
      }
      expect(
        await readFile(resolve(dist, "abcdef1234567890.txt"), "utf8"),
      ).toBe("abcdef1234567890");
      env.INDEXNOW_KEY = "bad key/";
      await expect(
        build({ siteURL: site, outDir: "test-results/seo-bad" }),
      ).rejects.toThrow("INDEXNOW_KEY");
    } finally {
      [
        env.GOOGLE_SITE_VERIFICATION,
        env.BING_SITE_VERIFICATION,
        env.YANDEX_VERIFICATION,
        env.INDEXNOW_KEY,
      ] = saved;
      for (const [key, value] of Object.entries({
        GOOGLE_SITE_VERIFICATION: saved[0],
        BING_SITE_VERIFICATION: saved[1],
        YANDEX_VERIFICATION: saved[2],
        INDEXNOW_KEY: saved[3],
      }))
        if (value === undefined) delete env[key];
        else env[key] = value;
    }
    const plain = await readFile(resolve(root, out, "index.html"), "utf8");
    expect(plain).not.toContain("site-verification");
  });
});
