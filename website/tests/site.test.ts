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
      pages.length * locales.length + locales.length,
    );
    expect(sitemap).toContain(`<loc>${canonical}en/</loc>`);
    expect(sitemap).toContain(`<loc>${canonical}es/</loc>`);
    for (const lang of locales)
      for (const page of pages)
        expect(sitemap).toContain(
          new URL(route("/handbook/", lang, page.slug), canonical).href,
        );
    expect(sitemap).not.toContain("/original/");
    expect(
      await readFile(resolve(root, output, "robots.txt"), "utf8"),
    ).toContain("Disallow: /handbook/original/");
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
    // The domain root sends readers to a language's landing page.
    expect(index).toContain('content="1;url=/en/"');
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
