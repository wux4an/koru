import { readFile, writeFile, mkdir, access } from "node:fs/promises";
import { resolve } from "node:path";
import { locales, pages, route, text, type Locale } from "./catalog";
import { versionTag } from "./version";

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
      "Templates: basic, vite, react, vue, svelte, vanilla-ts. Plugins you can enable in koru.json: device, app, storage, vibration, share, notifications.",
      `Experimental (${versionTag()}): recorded execution covers only Android 11 (API 30) on an x86_64 emulator. Framework templates, deep links and the run, devices and logs commands on real hardware are not validated on a device. Do not call an APK working until it has been installed and run on Android.`,
    ],
    when: "Read this when",
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
      "Plantillas: basic, vite, react, vue, svelte, vanilla-ts. Plugins que se activan en koru.json: device, app, storage, vibration, share, notifications.",
      `Experimental (${versionTag()}): la ejecución registrada solo cubre Android 11 (API 30) en un emulador x86_64. Las plantillas de frameworks, los deep links y los comandos run, devices y logs en hardware real no están validados en un dispositivo. No des un APK por funcional hasta instalarlo y ejecutarlo en Android.`,
    ],
    when: "Lee esto si",
    other: "English documentation: ",
    otherLabel: "English llms.txt index",
    optionalNote:
      "Arquitectura, internos y registros históricos de pruebas. Omítelos si el contexto es corto.",
    full: "Documentación completa en un archivo (inglés)",
    fullNote: "todas las páginas en inglés concatenadas.",
  },
};

const commands = {
  en: `koru create my-app   # new project from a template
koru dev             # browser preview with live reload
koru run             # install and open on a device or emulator
koru build --release # signed release APK
koru verify app.apk  # check the APK signature`,
  es: `koru create my-app   # proyecto nuevo desde una plantilla
koru dev             # vista previa en el navegador con recarga en vivo
koru run             # instala y abre en un dispositivo o emulador
koru build --release # APK release firmado
koru verify app.apk  # comprueba la firma del APK`,
};

/** When to read each page, as a short phrase appended to the page summary. */
const when: Record<string, [string, string]> = {
  overview: [
    "you need the big picture and where to go next.",
    "necesitas la visión general y saber adónde ir.",
  ],
  "getting-started": [
    "installing koru or creating your first app.",
    "instalas koru o creas tu primera app.",
  ],
  comparison: [
    "choosing between koru, Capacitor and Expo, or someone needs iOS or AAB.",
    "eliges entre koru, Capacitor y Expo, o se necesita iOS o AAB.",
  ],
  develop: [
    "previewing in the browser or using a framework dev server.",
    "previsualizas en el navegador o usas el servidor de un framework.",
  ],
  "run-on-device": [
    "installing on a phone or emulator, reading logs or debugging the WebView.",
    "instalas en un móvil o emulador, lees logs o depuras el WebView.",
  ],
  "existing-project": [
    "packaging Vite, React, Vue, Svelte, Astro or another existing frontend.",
    "empaquetas Vite, React, Vue, Svelte, Astro u otro frontend existente.",
  ],
  distribution: [
    "building a release APK, bumping versions or updating an installed app.",
    "generas un APK release, subes la versión o actualizas una app instalada.",
  ],
  security: [
    "handling signing keys or judging the WebView trust boundary.",
    "gestionas claves de firma o evalúas los límites de confianza del WebView.",
  ],
  troubleshooting: [
    "a build, install, signing or bridge error appears.",
    "aparece un error de build, instalación, firma o puente.",
  ],
  cli: [
    "you need the exact command, flag or exit behavior.",
    "necesitas el comando, la flag o el comportamiento exacto.",
  ],
  configuration: [
    "editing koru.json: permissions, SDK, deep links, assets or signing.",
    "editas koru.json: permisos, SDK, deep links, assets o firma.",
  ],
  plugins: [
    "calling Android features (device, storage, share, notifications) from JavaScript.",
    "llamas a funciones de Android (dispositivo, storage, share, notificaciones) desde JavaScript.",
  ],
  "write-a-plugin": [
    "adding a native capability beyond the built-in plugins.",
    "añades una capacidad nativa más allá de los plugins incluidos.",
  ],
  "dex-api": [
    "using or extending the Go DEX library.",
    "usas o extiendes la biblioteca DEX en Go.",
  ],
  "architecture-dex": [
    "you need the reasoning behind the DEX core design.",
    "necesitas el razonamiento del diseño del núcleo DEX.",
  ],
  "architecture-apk": [
    "you need to know how APK packaging and the runtime work internally.",
    "necesitas saber cómo funcionan por dentro el empaquetado APK y el runtime.",
  ],
  native: [
    "building the optional JNI adapter.",
    "compilas el adaptador JNI opcional.",
  ],
  "android-results": [
    "checking what has actually been run on Android.",
    "compruebas qué se ha ejecutado realmente en Android.",
  ],
  testing: [
    "reproducing or extending the test suites.",
    "reproduces o amplías las pruebas.",
  ],
  "dex-results": [
    "you need historical DEX fuzz and benchmark records.",
    "necesitas registros históricos de fuzz y benchmarks DEX.",
  ],
  milestones: [
    "you need project status, scope and open work.",
    "necesitas el estado, el alcance y el trabajo pendiente.",
  ],
  demo: [
    "reproducing the CLI recording.",
    "reproduces la grabación de la CLI.",
  ],
  "web-example": [
    "you need the web fixture used to validate the runtime.",
    "necesitas el fixture web que valida el runtime.",
  ],
};

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
    `- [${p.titles[index]}](${pageMarkdownURL(config, lang, p.slug)}): ${p.descriptions[index]} ${c.when} ${when[p.slug][index]}`;
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
    commands[lang],
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
export function plainMarkdown(
  markdown: string,
  config: Config,
  shift = 0,
  pageURL = "",
) {
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
        .replace(/<details class="walkthrough">[\s\S]*?<\/details>\n?/g, "")
        .replace(/<picture>[\s\S]*?<\/picture>/g, "")
        .replace(
          /<details class="api-detail"[^>]*><summary><code>(.*?)<\/code><span>(.*?)<\/span><\/summary>/g,
          "**`$1`**: $2",
        )
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
          new RegExp(
            `\\]\\((${origin}${config.base}docs/[a-z]{2}/(?:[a-z-]+/)?)(#[^)]*)?\\)`,
            "g",
          ),
          (_, dir, hash = "") => `](${dir}README.md${hash})`,
        )
        .replace(/\]\((#[^)]*)\)/g, `](${pageURL}$1)`)
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
    const md = plainMarkdown(
      await readFile(file, "utf8"),
      config,
      2,
      pageMarkdownURL(config, "en", p.slug),
    ).trim();
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
