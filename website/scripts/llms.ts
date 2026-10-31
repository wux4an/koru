import { readFile, writeFile, mkdir, access } from "node:fs/promises";
import { resolve } from "node:path";
import { locales, pages, route, text, type Locale } from "./catalog";

type Config = { url: string; base: string; production: boolean };

/** Public location of the llms.txt index for a language. */
export function llmsRoute(base: string, lang: Locale) {
  return lang === "en" ? `${base}llms.txt` : `${base}${lang}/llms.txt`;
}

const copy = {
  en: {
    summary:
      "koru is a single Go binary that turns static HTML, CSS and JavaScript into a signed Android APK, with no JDK, Gradle or Android SDK.",
    commandsIntro: "Key commands:",
    notes: [
      "Pure Go and the standard library: DEX, binary manifest, resources, ZIP packaging and the APK v2 signature are written in Go. Builds need no JDK, Gradle or Android SDK.",
      "Android only. koru does not build iOS apps or Play Store AABs; it produces APKs for direct installation.",
      "Runtime: the WebView loads your files from an intercepted HTTPS origin and talks to Android through a MessagePort. There is no JavaScriptInterface.",
      "Experimental (0.2.0-dev): recorded execution covers only Android 11 (API 30) on an x86_64 emulator. Newer features are not validated on a device.",
    ],
    other: "Documentación en español: ",
    otherLabel: "índice llms.txt en español",
    optionalNote:
      "Architecture, internals and historical test records. Skip these when context is short.",
    full: "Full documentation in one file",
    fullNote: "every English page concatenated, for loading in one request.",
  },
  es: {
    summary:
      "koru es un único binario Go que convierte HTML, CSS y JavaScript estáticos en un APK Android firmado, sin JDK, Gradle ni Android SDK.",
    commandsIntro: "Comandos principales:",
    notes: [
      "Go puro y biblioteca estándar: DEX, manifiesto binario, recursos, empaquetado ZIP y la firma APK v2 están escritos en Go. Compilar no requiere JDK, Gradle ni Android SDK.",
      "Solo Android. koru no genera apps iOS ni AAB para Play Store; produce APK para instalación directa.",
      "Runtime: el WebView carga tus archivos desde un origen HTTPS interceptado y se comunica con Android mediante MessagePort. No hay JavaScriptInterface.",
      "Experimental (0.2.0-dev): la ejecución registrada solo cubre Android 11 (API 30) en un emulador x86_64. Las funciones más nuevas no están validadas en un dispositivo.",
    ],
    other: "English documentation: ",
    otherLabel: "English llms.txt index",
    optionalNote:
      "Arquitectura, internos y registros históricos de pruebas. Omítelos si el contexto es corto.",
    full: "Documentación completa en un archivo (inglés)",
    fullNote: "todas las páginas en inglés concatenadas.",
  },
};

const commands = `koru create my-app   # new project from a template
koru dev             # browser preview with live reload
koru run             # install and open on a device or emulator
koru build --release # signed release APK
koru verify app.apk  # check the APK signature`;

const guideSlugs = (groups: number[]) =>
  pages.filter((p) => p.slug !== "sources" && groups.includes(p.group));

function abs(config: Config, path: string) {
  return new URL(path, config.url).href;
}
export function pageMarkdownURL(config: Config, lang: Locale, slug: string) {
  return abs(config, route(config.base, lang, slug) + "README.md");
}

/** llms.txt for one language, following https://llmstxt.org/. */
export function llmsIndex(config: Config, lang: Locale) {
  const c = copy[lang];
  const index = locales.indexOf(lang);
  const link = (p: (typeof pages)[number]) =>
    `- [${p.titles[index]}](${pageMarkdownURL(config, lang, p.slug)}): ${p.descriptions[index]}`;
  const section = (name: string, groups: number[]) =>
    `## ${name}\n\n${guideSlugs(groups).map(link).join("\n")}\n`;
  const other: Locale = lang === "en" ? "es" : "en";
  const head = [
    "# koru",
    "",
    `> ${c.summary}`,
    "",
    c.commandsIntro,
    "",
    "```sh",
    commands,
    "```",
    "",
    c.notes.map((n) => `- ${n}`).join("\n"),
    "",
    `${c.other}[${c.otherLabel}](${abs(config, llmsRoute(config.base, other))})`,
    "",
  ];
  const [a, b, r] = text[lang].groups;
  const optional = [
    ...guideSlugs([3, 4]).map(link),
    `- [${c.full}](${abs(config, `${config.base}llms-full.txt`)}): ${c.fullNote}`,
  ];
  return (
    head.join("\n") +
    "\n" +
    [
      section(a, [0]),
      section(b, [1]),
      section(r, [2]),
      `## Optional\n\n${c.optionalNote}\n\n${optional.join("\n")}\n`,
    ].join("\n")
  );
}

/** Remove decorative HTML from page Markdown, keeping every useful sentence. */
export function plainMarkdown(markdown: string, config: Config, shift = 0) {
  const origin = new URL(config.url).origin;
  return markdown
    .split(/(^```[\s\S]*?^```[^\n]*$)/m)
    .map((part, i) => {
      if (i % 2) return part;
      return part
        .replace(
          /<div class="clay-scene"[\s\S]*?clay-stamp[^\n]*\n<\/div>\n/g,
          "",
        )
        .replace(/<picture>[\s\S]*?<\/picture>/g, "")
        .replace(
          /<a class="doc-path" href="([^"]+)"><strong>(.*?)<\/strong><span>(.*?)<\/span><\/a>/g,
          "- [$2]($1): $3",
        )
        .replace(/<\/a><a /g, "</a>\n<a ")
        .replace(/<h1>(.*?)<\/h1>/g, "**$1**")
        .replace(/<i [^>]*><\/i>/g, "")
        .replace(/<a [^>]*href="([^"]+)"[^>]*>\s*(.*?)\s*<\/a>/g, "[$2]($1)")
        .replace(
          /<\/?(div|p|figure|figcaption|details|summary|span|picture|source|img)\b[^>]*>/g,
          "",
        )
        .replaceAll(`](${config.base}`, `](${origin}${config.base}`)
        .replace(
          /^(#{1,6}) /gm,
          (_, h) => "#".repeat(Math.min(6, h.length + shift)) + " ",
        )
        .replace(/\n{3,}/g, "\n\n");
    })
    .join("");
}

export async function writeLlms(dist: string, config: Config) {
  const checks: string[] = [];
  const out = async (path: string, body: string) => {
    await mkdir(resolve(dist, path, ".."), { recursive: true });
    await writeFile(resolve(dist, path), body);
    for (const m of body.matchAll(/\]\((https?:\/\/[^)\s]+)\)/g))
      checks.push(m[1]);
  };
  await out("llms.txt", llmsIndex(config, "en"));
  await out("es/llms.txt", llmsIndex(config, "es"));
  const parts: string[] = [
    `# koru\n\n> ${copy.en.summary}\n\nThis file concatenates every English documentation page. Index: ${abs(config, llmsRoute(config.base, "en"))}\n`,
  ];
  for (const p of guideSlugs([0, 1, 2, 3, 4])) {
    const file = resolve(
      dist,
      "docs",
      "en",
      p.slug === "overview" ? "" : p.slug,
      "README.md",
    );
    const md = plainMarkdown(await readFile(file, "utf8"), config, 2).trim();
    parts.push(
      `---\n\n## ${p.titles[0]}\n\nSource: ${pageMarkdownURL(config, "en", p.slug)}\n\n${md}\n`,
    );
  }
  await out("llms-full.txt", parts.join("\n"));
  // Fail the build rather than publish an index that points at nothing.
  const prefix = abs(config, config.base);
  for (const url of new Set(checks)) {
    if (!url.startsWith(prefix)) continue;
    const rel = url.slice(prefix.length).split(/[?#]/)[0];
    const path = resolve(dist, rel);
    if (rel.endsWith("/")) await access(resolve(path, "index.html"));
    else await access(path);
  }
}
