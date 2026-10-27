import { chromium, type Page } from "playwright";
import AxeBuilder from "@axe-core/playwright";
import assert from "node:assert/strict";
import { mkdir, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { build, root } from "./build";
import { pages, locales, route } from "./catalog";

const chrome = process.env.CHROME_PATH || Bun.which("google-chrome-stable");
if (!chrome)
  throw new Error(
    "google-chrome-stable not found. Set CHROME_PATH to the installed executable.",
  );
await build({ siteURL: "http://localhost:4173/" });
await mkdir(resolve(root, "test-results"), { recursive: true });
const port = 4183;
const origin = `http://127.0.0.1:${port}`;
const server = Bun.spawn(["bun", "scripts/serve.ts"], {
  cwd: root,
  env: { ...process.env, PORT: String(port) },
  stdout: "pipe",
  stderr: "pipe",
});
const report: {
  chrome?: string;
  passed: string[];
  pageCount: number;
  accessibility: object[];
  errors: string[];
} = { passed: [], pageCount: 0, accessibility: [], errors: [] };
const browser = await chromium.launch({
  executablePath: chrome,
  headless: true,
  args: ["--no-sandbox"],
});
report.chrome = browser.version();
const context = await browser.newContext({
  viewport: { width: 1440, height: 1050 },
  permissions: ["clipboard-read", "clipboard-write"],
});
const page = await context.newPage();
page.on("pageerror", (error) => report.errors.push(error.message));
page.on("requestfailed", (request) => {
  if (request.failure()?.errorText !== "net::ERR_ABORTED")
    report.errors.push(`${request.url()}: ${request.failure()?.errorText}`);
});
page.on("request", (request) => {
  assert.equal(
    new URL(request.url()).origin,
    origin,
    `Unexpected external asset: ${request.url()}`,
  );
});
const ready = async (page: Page) => {
  await page.locator(".search input").waitFor();
  await page.locator(".markdown-section h1").waitFor();
  await page.waitForFunction(
    () =>
      !!document.querySelector(".doc-footer") &&
      document.body.classList.contains("ready"),
  );
  await page.evaluate(() => document.fonts.ready);
};
const chooseTheme = async (page: Page, choice: string) => {
  await page.locator("[data-theme-button]").click();
  await page.locator(`[data-theme-choice="${choice}"]`).click();
};
const check = async (name: string, task: () => Promise<void>) => {
  await task();
  report.passed.push(name);
  console.log(`PASS ${name}`);
};
try {
  for (let attempt = 0; attempt < 60; attempt++) {
    try {
      if ((await fetch(origin + "/docs/en/")).ok) break;
    } catch {}
    await Bun.sleep(100);
    if (attempt === 59) throw new Error("Preview server did not start.");
  }
  await check(
    `${pages.length * locales.length} routes render, with one H1, correct language and valid internal links`,
    async () => {
      const links = new Set<string>();
      for (const lang of locales) {
        for (const item of pages) {
          const url = origin + route("/", lang, item.slug);
          assert.equal((await page.goto(url))?.status(), 200, url);
          await ready(page);
          assert.equal(await page.locator(".markdown-section h1").count(), 1);
          assert.equal(await page.locator("html").getAttribute("lang"), lang);
          assert.equal(
            await page.title(),
            `${item.titles[lang === "en" ? 0 : 1]} · koru`,
          );
          const hrefs = await page
            .locator("a[href]")
            .evaluateAll((elements) =>
              elements
                .map((el) => el.getAttribute("href") || "")
                .filter(
                  (href) => href.startsWith("/") && !href.startsWith("//"),
                ),
            );
          hrefs.forEach((href) => links.add(href.split("?")[0].split("#")[0]));
          report.pageCount++;
        }
      }
      for (const href of links)
        assert.equal(
          (await fetch(origin + href)).status,
          200,
          `Broken link: ${href}`,
        );
    },
  );
  await check(
    "search initializes even when the plugin loads slowly",
    async () => {
      await page.route("**/assets/vendor/search.min.js", async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 350));
        await route.continue();
      });
      await page.goto(origin + "/docs/en/");
      await ready(page);
      await page.unroute("**/assets/vendor/search.min.js");
    },
  );
  await check(
    "language switch preserves document, translates navigation and search",
    async () => {
      await page.goto(origin + "/docs/en/plugins/");
      await ready(page);
      await page.locator('[data-locale="es"]').click();
      await page.waitForURL("**/es/plugins/");
      await ready(page);
      assert.equal(
        await page.locator(".markdown-section h1").innerText(),
        "API JavaScript",
      );
      assert.equal(
        await page.locator(".search input").getAttribute("placeholder"),
        "Buscar documentación",
      );
      await page.locator('[data-locale="en"]').click();
      await page.waitForURL("**/en/plugins/");
      await ready(page);
    },
  );
  await check(
    "client navigation updates canonical, hreflang, JSON-LD and browser history",
    async () => {
      await page
        .locator('.sidebar-nav a[href="/docs/en/configuration/"]')
        .click();
      await page.waitForURL("**/en/configuration/");
      await page.waitForFunction(
        () => document.title === "Configuration · koru",
      );
      assert.equal(
        await page.locator('link[rel="canonical"]').getAttribute("href"),
        "http://localhost:4173/docs/en/configuration/",
      );
      assert.equal(
        await page.locator('link[hreflang="es"]').getAttribute("href"),
        "http://localhost:4173/docs/es/configuration/",
      );
      const jsonld = JSON.parse(
        (await page.locator("#structured-data").textContent()) || "{}",
      );
      assert.equal(jsonld.headline, "Configuration");
      await page.goBack();
      await page.waitForURL("**/en/plugins/");
      await page.waitForFunction(
        () => document.title === "JavaScript API · koru",
      );
    },
  );
  for (const lang of locales)
    await check(`${lang} search is indexed and locale-scoped`, async () => {
      await page.goto(origin + route("/", lang, "overview"));
      await ready(page);
      await page.waitForFunction(
        (locale) =>
          !!localStorage.getItem(
            `docsify.search.index/koru-docs-v2/docs/${locale}`,
          ),
        lang,
      );
      const input = page.locator(".search input");
      await input.fill("MessagePort");
      await page.locator(".matching-post a").first().waitFor();
      const hrefs = await page
        .locator(".matching-post a")
        .evaluateAll((links) => links.map((link) => link.getAttribute("href")));
      assert(hrefs.length > 0);
      assert(
        hrefs.every((href) => href?.startsWith(`/docs/${lang}/`)),
        JSON.stringify(hrefs),
      );
      await input.fill("zzzzkorunomatchzzzz");
      await page.waitForFunction(
        () => document.querySelector(".results-panel .empty") !== null,
      );
      assert.equal(
        await page.locator(".results-panel .empty").innerText(),
        lang === "en" ? "No results found." : "No se encontraron resultados.",
      );
      await input.fill("");
    });
  await check(
    "copy code uses the existing plugin and copies the complete command",
    async () => {
      await page.goto(origin + "/docs/en/");
      await ready(page);
      const code = await page.locator("pre code").first().textContent();
      await page.locator(".docsify-copy-code-button").first().click();
      const clipboard = await page.evaluate(() =>
        navigator.clipboard.readText(),
      );
      assert.equal(clipboard.trim(), code?.trim());
    },
  );
  await check(
    "keyboard skip link reaches content with visible focus",
    async () => {
      await page.goto(origin + "/docs/en/");
      await ready(page);
      await page.keyboard.press("Tab");
      assert.equal(
        await page.locator(":focus").getAttribute("class"),
        "skip-link",
      );
      await page.keyboard.press("Enter");
      assert.equal(await page.locator(":focus").getAttribute("id"), "main");
    },
  );
  await check(
    "section index is localized, keyboard accessible and jumps to its heading",
    async () => {
      await page.goto(origin + "/docs/es/configuration/");
      await ready(page);
      const toc = page.locator(".page-toc");
      assert.equal(await toc.locator("summary").innerText(), "En esta página");
      await toc.locator("summary").click();
      const link = toc.locator("a").nth(1);
      const target = await link.getAttribute("href");
      assert(target?.startsWith("#"));
      await link.click();
      assert.equal(new URL(page.url()).hash, target);
      await page.locator(target!).waitFor();
      // The brand leads to the landing page with a full page load.
      await page.locator(".site-header .brand").click();
      await page.waitForURL("**/es/");
      await page.locator(".landing-hero h1").waitFor();
      await page.goBack();
      await ready(page);
      // The link must scroll within the page, not route away from it.
      assert.equal(new URL(page.url()).pathname, "/docs/es/configuration/");
      const top = await page
        .locator(target!)
        .evaluate((heading) => heading.getBoundingClientRect().top);
      assert(top >= 60 && top < 200, `section heading at ${top}px`);
    },
  );
  await check(
    "overview, API reference and mobile layouts have no horizontal overflow",
    async () => {
      for (const width of [1440, 1024, 768, 390, 320]) {
        await page.setViewportSize({ width, height: width < 800 ? 844 : 1050 });
        for (const slug of [
          "overview",
          "configuration",
          "dex-api",
          "plugins",
        ]) {
          await page.goto(origin + route("/", "en", slug));
          await ready(page);
          await page.waitForTimeout(350);
          if (slug === "plugins" && width <= 390)
            await page.locator("#api-notification-show summary").click();
          assert(
            await page.evaluate(
              () => document.documentElement.scrollWidth <= innerWidth + 1,
            ),
            `Overflow at ${width}px on ${slug}`,
          );
        }
      }
      await page.setViewportSize({ width: 1440, height: 1050 });
      await page.goto(origin + "/docs/en/");
      await ready(page);
      await page.screenshot({
        path: resolve(root, "test-results/desktop-en.png"),
        fullPage: true,
      });
      await page.goto(origin + "/docs/es/plugins/");
      await ready(page);
      await page.screenshot({
        path: resolve(root, "test-results/reference-es.png"),
        fullPage: true,
      });
      await page.setViewportSize({ width: 390, height: 844 });
      await page.goto(origin + "/docs/es/");
      await ready(page);
      await page.screenshot({
        path: resolve(root, "test-results/mobile-es.png"),
        fullPage: true,
      });
      await page.locator(".sidebar-toggle").click();
      await page.waitForFunction(() =>
        document.body.classList.contains("close"),
      );
      await page.waitForTimeout(350);
      assert.equal(
        await page.locator(".sidebar-toggle").getAttribute("aria-expanded"),
        "true",
      );
      assert(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth + 1,
        ),
      );
      await page
        .locator('.sidebar-nav a[href="/docs/es/configuration/"]')
        .click();
      await page.waitForURL("**/es/configuration/");
      await page.waitForFunction(
        () => !document.body.classList.contains("close"),
      );
    },
  );
  await check(
    "axe WCAG 2.2 AA checks pass for overview and reference in both locales",
    async () => {
      for (const lang of locales)
        for (const slug of [
          "overview",
          "plugins",
          "configuration",
          "dex-api",
          "android-results",
        ]) {
          await page.setViewportSize({
            width: slug === "overview" ? 390 : 1440,
            height: 1050,
          });
          await page.goto(origin + route("/", lang, slug));
          await ready(page);
          const result = await new AxeBuilder({ page })
            .withTags([
              "wcag2a",
              "wcag2aa",
              "wcag21aa",
              "wcag22aa",
              "best-practice",
            ])
            .analyze();
          report.accessibility.push({
            lang,
            slug,
            violations: result.violations,
          });
          assert.equal(
            result.violations.length,
            0,
            `${lang}/${slug}: ${JSON.stringify(result.violations.map((v) => ({ id: v.id, nodes: v.nodes.map((n) => n.target) })))}`,
          );
        }
    },
  );
  await check(
    "all documents remain readable and linked without JavaScript",
    async () => {
      const noJS = await browser.newContext({
        javaScriptEnabled: false,
        viewport: { width: 390, height: 844 },
      });
      const reader = await noJS.newPage();
      for (const lang of locales)
        for (const item of pages) {
          await reader.goto(origin + route("/", lang, item.slug));
          assert.equal(await reader.locator("#main h1").count(), 1);
          assert((await reader.locator("#main").innerText()).length > 300);
          assert(
            await reader.evaluate(
              () => document.documentElement.scrollWidth <= innerWidth + 1,
            ),
          );
        }
      await reader.goto(origin + "/docs/en/configuration/");
      await reader.locator('[data-locale="es"]').click();
      await reader.waitForURL("**/es/configuration/");
      assert.equal(await reader.locator("html").getAttribute("lang"), "es");
      await reader.screenshot({
        path: resolve(root, "test-results/nojs-mobile-es.png"),
        fullPage: true,
      });
      await noJS.close();
    },
  );
  await check(
    "unknown routes return HTTP 404 and archived HTML is source text",
    async () => {
      assert.equal(
        (await fetch(origin + "/docs/en/not-a-document/")).status,
        404,
      );
      assert.equal(
        (await fetch(origin + "/docs/es/missing/README.md")).status,
        404,
      );
      const raw = await fetch(origin + "/original/examples/web/index.html.txt");
      assert.match(raw.headers.get("content-type") || "", /^text\/plain/);
      assert.equal(raw.headers.get("x-robots-tag"), "noindex");
    },
  );
  await check("reduced motion is respected", async () => {
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.goto(origin + "/docs/en/");
    await ready(page);
    assert.equal(
      await page
        .locator(".sidebar")
        .evaluate((el) => getComputedStyle(el).transitionDuration),
      "0s",
    );
  });
  await check(
    "production subpath supports direct loads, routing, search and reciprocal locales",
    async () => {
      const folder = "test-results/subpath-build";
      await build({
        siteURL: "https://docs.koru.test/handbook/",
        outDir: folder,
      });
      const nestedOrigin = "http://127.0.0.1:4184";
      const nestedServer = Bun.spawn(["bun", "scripts/serve.ts"], {
        cwd: root,
        env: { ...process.env, PORT: "4184", DIST_DIR: folder },
        stdout: "pipe",
        stderr: "pipe",
      });
      const nestedContext = await browser.newContext({
        viewport: { width: 1280, height: 900 },
      });
      const nestedPage = await nestedContext.newPage();
      nestedPage.on("pageerror", (error) => report.errors.push(error.message));
      nestedPage.on("request", (request) =>
        assert.equal(new URL(request.url()).origin, nestedOrigin),
      );
      try {
        for (let attempt = 0; attempt < 60; attempt++) {
          try {
            if ((await fetch(nestedOrigin + "/handbook/docs/en/")).ok) break;
          } catch {}
          await Bun.sleep(100);
          if (attempt === 59) throw new Error("Subpath server did not start.");
        }
        await nestedPage.goto(
          nestedOrigin + "/handbook/docs/en/configuration/",
        );
        await ready(nestedPage);
        assert.equal(
          await nestedPage
            .locator('meta[name="robots"]')
            .getAttribute("content"),
          "index,follow",
        );
        assert.equal(
          await nestedPage
            .locator('link[rel="canonical"]')
            .getAttribute("href"),
          "https://docs.koru.test/handbook/docs/en/configuration/",
        );
        await nestedPage
          .locator('.sidebar-nav a[href="/handbook/docs/en/plugins/"]')
          .click();
        await nestedPage.waitForURL("**/handbook/docs/en/plugins/");
        await nestedPage.waitForFunction(
          () => document.title === "JavaScript API · koru",
        );
        await nestedPage.locator('[data-locale="es"]').click();
        await nestedPage.waitForURL("**/handbook/docs/es/plugins/");
        await ready(nestedPage);
        assert.equal(
          await nestedPage.locator(".markdown-section h1").innerText(),
          "API JavaScript",
        );
        await nestedPage.waitForFunction(
          () =>
            !!localStorage.getItem(
              "docsify.search.index/koru-docs-v2/handbook/docs/es",
            ),
        );
        await nestedPage.locator(".search input").fill("MessagePort");
        await nestedPage.locator(".matching-post a").first().waitFor();
        const hrefs = await nestedPage
          .locator(".matching-post a")
          .evaluateAll((elements) =>
            elements.map((el) => el.getAttribute("href")),
          );
        assert(hrefs.every((href) => href?.startsWith("/handbook/docs/es/")));
        assert.equal(
          (await fetch(nestedOrigin + "/handbook/docs/es/missing/")).status,
          404,
        );
      } finally {
        await nestedContext.close();
        nestedServer.kill();
      }
    },
  );
  await check(
    "visited hero action stays legible on dark hover and the VHS GIF is available",
    async () => {
      const themed = await browser.newContext({ colorScheme: "dark" });
      const hero = await themed.newPage();
      try {
        await hero.goto(origin + "/docs/en/");
        await ready(hero);
        const action = hero.locator(".hero-actions .primary-link");
        await action.click();
        await hero.waitForURL("**/en/getting-started/");
        await hero.goBack();
        await ready(hero);
        await action.hover();
        const contrast = await action.evaluate((el) => {
          const rgb = (value: string) =>
            (value.match(/[\d.]+/g) || []).slice(0, 3).map(Number);
          const luminance = (channels: number[]) =>
            channels
              .map((channel) => channel / 255)
              .map((channel) =>
                channel <= 0.04045
                  ? channel / 12.92
                  : ((channel + 0.055) / 1.055) ** 2.4,
              )
              .reduce(
                (sum, channel, index) =>
                  sum + channel * [0.2126, 0.7152, 0.0722][index],
                0,
              );
          const style = getComputedStyle(el);
          const foreground = luminance(rgb(style.color));
          const background = luminance(rgb(style.backgroundColor));
          return (
            (Math.max(foreground, background) + 0.05) /
            (Math.min(foreground, background) + 0.05)
          );
        });
        assert(contrast >= 4.5, `Hero CTA contrast: ${contrast}`);
        const recording = hero.locator(".hero-recording picture");
        assert.equal(await recording.count(), 1);
        const animation = await recording.locator("img").getAttribute("src");
        const poster = await recording.locator("source").getAttribute("srcset");
        assert(animation?.endsWith(".gif"));
        assert.equal((await fetch(origin + animation)).status, 200);
        assert.equal((await fetch(origin + poster)).status, 200);
        await hero.emulateMedia({ reducedMotion: "reduce" });
        assert.equal(
          await recording
            .locator("source")
            .evaluate((el) => matchMedia(el.media).matches),
          true,
        );
      } finally {
        await themed.close();
      }
    },
  );
  await check(
    "all 13 API accordions explain methods and work with keyboard in both locales",
    async () => {
      for (const lang of locales) {
        await page.goto(origin + `/${lang}/plugins/`);
        await ready(page);
        const accordions = page.locator(".api-detail");
        assert.equal(await accordions.count(), 13);
        const first = accordions.first();
        await first.locator("summary").focus();
        await page.keyboard.press("Enter");
        assert.equal(await first.getAttribute("open"), "");
        assert(await first.locator("pre code").count());
        assert(await first.innerText().then((value) => value.length > 200));
        await page.keyboard.press("Enter");
        assert.equal(await first.getAttribute("open"), null);
      }
    },
  );
  await check(
    "appearance follows the system and remembers manual choices across reloads, locales and tabs",
    async () => {
      const themes = await browser.newContext({
        colorScheme: "dark",
        viewport: { width: 1440, height: 1050 },
      });
      const themed = await themes.newPage();
      const assertTheme = async (expected: string) => {
        await themed.waitForFunction(
          (theme) => document.documentElement.dataset.theme === theme,
          expected,
        );
        assert.equal(
          await themed
            .locator("html")
            .evaluate((el) => getComputedStyle(el).colorScheme),
          expected,
        );
      };
      try {
        await themed.goto(origin + "/docs/en/");
        await ready(themed);
        await assertTheme("dark");
        assert.equal(
          await themed
            .locator("[data-theme-button]")
            .getAttribute("data-preference"),
          "system",
        );
        await themed.emulateMedia({ colorScheme: "light" });
        await assertTheme("light");
        await chooseTheme(themed, "dark");
        await assertTheme("dark");
        await themed.reload();
        await ready(themed);
        await assertTheme("dark");
        assert.equal(
          await themed
            .locator("[data-theme-button]")
            .getAttribute("data-preference"),
          "dark",
        );
        await themed.locator('[data-locale="es"]').click();
        await themed.waitForURL("**/es/");
        await ready(themed);
        await assertTheme("dark");
        assert.equal(
          await themed
            .locator("[data-theme-button]")
            .getAttribute("aria-label"),
          "Apariencia",
        );
        const second = await themes.newPage();
        await second.goto(origin + "/docs/es/configuration/");
        await ready(second);
        await chooseTheme(themed, "light");
        await assertTheme("light");
        await second.waitForFunction(
          () => document.documentElement.dataset.theme === "light",
        );
        await themed.emulateMedia({ colorScheme: "dark" });
        await assertTheme("light");
        await chooseTheme(themed, "system");
        await assertTheme("dark");
        await themed.emulateMedia({ colorScheme: "light" });
        await assertTheme("light");
        await themed.evaluate(() =>
          localStorage.setItem("koru-theme", "invalid"),
        );
        await themed.reload();
        await ready(themed);
        assert.equal(
          await themed
            .locator("[data-theme-button]")
            .getAttribute("data-preference"),
          "system",
        );
        await assertTheme("light");
      } finally {
        await themes.close();
      }
    },
  );
  await check(
    "appearance initializes before Docsify and works when preference storage is blocked",
    async () => {
      const restricted = await browser.newContext({ colorScheme: "dark" });
      await restricted.addInitScript(() => {
        Object.defineProperty(window, "localStorage", {
          get() {
            throw new DOMException("Storage is disabled", "SecurityError");
          },
        });
      });
      const reader = await restricted.newPage();
      // Stop enhancement scripts: the blocking theme bootstrap must still run.
      await reader.route("**/assets/site.js", (request) => request.abort());
      try {
        await reader.goto(origin + "/docs/en/");
        assert.equal(
          await reader.locator("html").getAttribute("data-theme"),
          "dark",
        );
        await chooseTheme(reader, "light");
        assert.equal(
          await reader.locator("html").getAttribute("data-theme"),
          "light",
        );
      } finally {
        await restricted.close();
      }
    },
  );
  await check(
    "dark overview and references pass accessibility and responsive checks in both locales",
    async () => {
      const dark = await browser.newContext({
        colorScheme: "dark",
        viewport: { width: 1440, height: 1050 },
      });
      const darkPage = await dark.newPage();
      darkPage.on("pageerror", (error) => report.errors.push(error.message));
      try {
        for (const lang of locales)
          for (const slug of ["overview", "configuration", "dex-api"]) {
            await darkPage.goto(origin + route("/", lang, slug));
            await ready(darkPage);
            assert.equal(
              await darkPage.locator("html").getAttribute("data-theme"),
              "dark",
            );
            const result = await new AxeBuilder({ page: darkPage })
              .withTags([
                "wcag2a",
                "wcag2aa",
                "wcag21aa",
                "wcag22aa",
                "best-practice",
              ])
              .analyze();
            report.accessibility.push({
              lang,
              slug,
              theme: "dark",
              violations: result.violations,
            });
            assert.equal(
              result.violations.length,
              0,
              `${lang}/${slug} dark: ${JSON.stringify(result.violations.map((v) => ({ id: v.id, nodes: v.nodes.map((n) => n.target) })))}`,
            );
          }
        await darkPage.goto(origin + "/docs/en/");
        await ready(darkPage);
        await darkPage.screenshot({
          path: resolve(root, "test-results/desktop-dark-en.png"),
          fullPage: true,
        });
        for (const width of [768, 390, 320]) {
          await darkPage.setViewportSize({ width, height: 844 });
          await darkPage.goto(origin + "/docs/es/");
          await ready(darkPage);
          await darkPage.waitForTimeout(350);
          assert(
            await darkPage.evaluate(
              () => document.documentElement.scrollWidth <= innerWidth + 1,
            ),
            `Dark overflow at ${width}px`,
          );
        }
        await darkPage.setViewportSize({ width: 390, height: 844 });
        await darkPage.screenshot({
          path: resolve(root, "test-results/mobile-dark-es.png"),
          fullPage: true,
        });
        await darkPage.locator("[data-theme-button]").focus();
        await darkPage.keyboard.press("ArrowDown");
        await darkPage.keyboard.press("Home");
        await darkPage.keyboard.press("ArrowDown");
        await darkPage.keyboard.press("Enter");
        await darkPage.waitForFunction(
          () => document.documentElement.dataset.theme === "light",
        );
      } finally {
        await dark.close();
      }
    },
  );
  await check(
    "custom appearance menu supports keyboard selection, Escape, outside dismissal and accessible checked states",
    async () => {
      await page.setViewportSize({ width: 1440, height: 1050 });
      await page.goto(origin + "/docs/en/");
      await ready(page);
      const trigger = page.locator("[data-theme-button]");
      assert.equal(await page.locator(".theme-control select").count(), 0);
      await trigger.focus();
      await page.keyboard.press("ArrowDown");
      assert.equal(
        await page.locator(":focus").getAttribute("data-theme-choice"),
        "system",
      );
      await page.keyboard.press("End");
      await page.keyboard.press("Enter");
      assert.equal(await trigger.getAttribute("data-preference"), "dark");
      assert.equal(await trigger.getAttribute("aria-expanded"), "false");
      assert.equal(
        await page.locator(":focus").getAttribute("data-theme-button"),
        "",
      );
      await trigger.click();
      assert.equal(
        await page
          .locator('[data-theme-choice="dark"]')
          .getAttribute("aria-checked"),
        "true",
      );
      const audit = await new AxeBuilder({ page })
        .withTags([
          "wcag2a",
          "wcag2aa",
          "wcag21aa",
          "wcag22aa",
          "best-practice",
        ])
        .analyze();
      report.accessibility.push({
        lang: "en",
        slug: "overview",
        theme: "dark",
        menu: "open",
        violations: audit.violations,
      });
      assert.equal(
        audit.violations.length,
        0,
        JSON.stringify(audit.violations.map((v) => v.id)),
      );
      await page.keyboard.press("Escape");
      assert.equal(await trigger.getAttribute("aria-expanded"), "false");
      await trigger.click();
      await page.locator(".hero h1").click();
      assert.equal(await trigger.getAttribute("aria-expanded"), "false");
      await trigger.focus();
      await page.keyboard.press("ArrowUp");
      assert.equal(
        await page.locator(":focus").getAttribute("data-theme-choice"),
        "dark",
      );
      await page.keyboard.press("Home");
      await page.keyboard.press("ArrowDown");
      await page.keyboard.press("Enter");
      assert.equal(await trigger.getAttribute("data-preference"), "light");
      await trigger.click();
      await page.keyboard.press("Tab");
      assert.equal(await trigger.getAttribute("aria-expanded"), "false");
    },
  );
  await check(
    "landing pages pass accessibility, fit every width and copy the quick start in both themes",
    async () => {
      for (const [lang, path] of [
        ["en", "/en/"],
        ["es", "/es/"],
      ] as const) {
        for (const theme of ["light", "dark"]) {
          await page.setViewportSize({ width: 1440, height: 1050 });
          await page.goto(origin + path);
          await page.evaluate(() => document.fonts.ready);
          await chooseTheme(page, theme);
          assert.equal(
            await page.evaluate(() => document.documentElement.dataset.theme),
            theme,
          );
          assert.equal(await page.locator("h1").count(), 1);
          assert.equal(await page.getAttribute("html", "lang"), lang);
          for (const width of [1440, 1024, 768, 390, 320]) {
            await page.setViewportSize({
              width,
              height: width < 800 ? 844 : 1050,
            });
            assert(
              await page.evaluate(
                () => document.documentElement.scrollWidth <= innerWidth + 1,
              ),
              `Landing overflow at ${width}px (${lang}, ${theme})`,
            );
            if (width === 390 || width === 1440)
              await page.screenshot({
                path: resolve(
                  root,
                  `test-results/landing-${lang}-${theme}-${width}.png`,
                ),
                fullPage: true,
              });
            const result = await new AxeBuilder({ page })
              .withTags([
                "wcag2a",
                "wcag2aa",
                "wcag21aa",
                "wcag22aa",
                "best-practice",
              ])
              .analyze();
            report.accessibility.push({
              lang,
              slug: `landing-${theme}-${width}`,
              violations: result.violations,
            });
            assert.equal(
              result.violations.length,
              0,
              `landing ${lang} ${theme} ${width}: ${JSON.stringify(result.violations.map((v) => ({ id: v.id, nodes: v.nodes.map((n) => n.target) })))}`,
            );
          }
        }
        await page.setViewportSize({ width: 1440, height: 1050 });
        await page.locator(".copy-button").click();
        assert.match(
          await page.evaluate(() => navigator.clipboard.readText()),
          lang === "en"
            ? /^koru create my-app\ncd my-app/
            : /^koru create mi-app/,
        );
        await page.locator(".landing-actions .button-primary").first().click();
        await page.waitForURL(`**/${lang}/getting-started/`);
        await ready(page);
        assert.equal(
          await page.locator(".site-header .brand").getAttribute("href"),
          path,
        );
      }
    },
  );
  assert.deepEqual(report.errors, [], "Browser errors");
  report.passed.push("no page errors or external asset requests");
} finally {
  await writeFile(
    resolve(root, "test-results/browser-report.json"),
    JSON.stringify(report, null, 2) + "\n",
  );
  await context.close();
  await browser.close();
  server.kill();
}
console.log(
  `Google Chrome ${report.chrome}: ${report.passed.length} checks, ${report.pageCount} localized routes.`,
);
