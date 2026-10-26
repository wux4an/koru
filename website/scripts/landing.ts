// The product landing page: one standalone HTML document per locale.
import { route, type Locale } from "./catalog";

const escape = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (char) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        char
      ]!,
  );

export const landingRoute = (base: string, lang: Locale) =>
  lang === "en" ? base : `${base}es/inicio/`;

const copy = {
  en: {
    title: "koru · Web apps on Android, built with one Go binary",
    description:
      "Create, preview and ship Android apps with HTML, CSS and JavaScript. Templates, live reload, one-command install and a typed JavaScript API, without Java, Gradle or the Android SDK.",
    skip: "Skip to content",
    nav: [
      ["Docs", "overview"],
      ["JavaScript API", "plugins"],
      ["CLI", "cli"],
    ],
    eyebrow: "Experimental · open source · MIT",
    h1: "Your web app, on Android.",
    lead: "Create, preview and ship Android apps with the HTML, CSS and JavaScript you already know. koru is one Go binary: no Java, no Gradle, no Android SDK.",
    primary: "Get started",
    secondary: "Compare with Capacitor and Expo",
    terminalLabel: "From zero to your phone in four commands",
    terminal: [
      ["koru create my-app", "start from a template"],
      ["cd my-app", ""],
      ["koru dev", "preview with live reload"],
      ["koru run", "install and open on a device"],
    ],
    copy: "Copy commands",
    copied: "Copied",
    phone: {
      badge: "android · v1.0.0",
      title: "My App",
      lead: "Edit www/index.html and save.",
      counter: "Saved counter",
      count: "Count: 3",
      device: "This device",
      rows: [
        ["Model", "Android emulator"],
        ["Android API", "30"],
      ],
      actions: ["Vibrate", "Share"],
    },
    phoneAlt:
      "The starter app on a phone: a saved counter, device details and native actions",
    featuresTitle: "Everything you need for the whole loop",
    features: [
      [
        "magic-wand",
        "Start in seconds",
        "koru create writes a ready project from a template: plain HTML, Vite, React, Vue, Svelte or TypeScript, with typed JavaScript, a signing key and a .gitignore that keeps it private.",
      ],
      [
        "arrows-clockwise",
        "Live reload in your browser",
        "koru dev serves your app with live reload and prints a QR code to open it on your phone. The same JavaScript API runs on web implementations, so storage, share and notifications work before you build.",
      ],
      [
        "device-mobile",
        "One command to your phone",
        "koru run builds, installs and opens the app on a connected phone or emulator, and explains what to fix when something is wrong. koru logs shows the app's logs in your terminal.",
      ],
      [
        "puzzle-piece",
        "Android features from JavaScript",
        "Device info, persistent storage, vibration, the share sheet and notifications, with promises, typed errors and TypeScript declarations.",
      ],
      [
        "lightning",
        "Fast, offline builds",
        "The starter template builds into a signed 21 KB APK in about 35 ms on our test machine. No downloads, daemons or cloud queue.",
      ],
      [
        "shield-check",
        "Signed and verified",
        "APK Signature Scheme v2 with a persistent key, a strict verifier, an HTTPS-only app origin and a MessagePort bridge with no addJavascriptInterface.",
      ],
    ],
    codeTitle: "Native features, plain JavaScript",
    codeLead:
      "Import the module koru serves inside your app. In a browser the same calls use web implementations, and platform tells you where you are.",
    code: `import { device, storage, share, platform } from "/__koru/core.mjs";

const { model, sdk } = await device.getInfo();
await storage.set("theme", "dark");

document.querySelector("#share").onclick = () =>
  share.text(\`Built with koru on \${platform}\`);`,
    codeLink: "Read the JavaScript API",
    frameworksTitle: "Bring your own frontend",
    frameworksLead:
      "Start from a React, Vue, Svelte or TypeScript template, or bring anything that outputs static files: point webDir at the build folder. Keep your framework's hot reload with koru dev --proxy.",
    frameworks: [
      "Vite",
      "React",
      "Vue",
      "Svelte",
      "Astro",
      "Solid",
      "Vanilla JS",
    ],
    frameworksLink: "Framework guide",
    compareTitle: "How koru compares",
    compareLead:
      "Capacitor and Expo are mature and cover far more ground. koru trades that breadth for a tiny, self-contained Android toolchain.",
    compareHead: ["", "koru", "Capacitor", "Expo"],
    compareRows: [
      ["You write", "HTML, CSS, JS", "HTML, CSS, JS", "React Native"],
      [
        "Android build needs",
        "the koru binary",
        "Android Studio, Gradle, JDK",
        "EAS cloud, or Android SDK locally",
      ],
      ["Platforms", "Android", "Android, iOS, web", "Android, iOS, web"],
      ["Output", "Signed APK", "APK, AAB", "APK, AAB"],
      ["Plugins", "6 built-in + DEX plugins", "Large ecosystem", "Large SDK"],
    ],
    compareLink: "Detailed comparison",
    statusTitle: "Honest status",
    status:
      "koru is experimental. Execution is recorded on an Android 11 (API 30) x86_64 emulator; AAB output for Google Play and iOS are not available. Test on the devices you support.",
    statusLink: "Tested compatibility",
    ctaTitle: "Build your first app now",
    ctaLead: "Install koru with Go 1.23+, then create a project.",
    install: `git clone https://github.com/wux4an/koru.git
cd koru && make install`,
    footer: "Made for the web. Built for Android.",
    donate: "Support koru",
    language: "Español",
    languageLabel: "Language",
    mainNav: "Main",
    footerNav: "Footer",
    tableLabel: "Comparison table",
    appearance: "Appearance",
    themes: ["System", "Light", "Dark"],
  },
  es: {
    title: "koru · Apps web en Android con un solo binario de Go",
    description:
      "Crea, previsualiza y publica apps Android con HTML, CSS y JavaScript. Plantillas, recarga en vivo, instalación con un comando y una API JavaScript tipada, sin Java, Gradle ni el SDK de Android.",
    skip: "Saltar al contenido",
    nav: [
      ["Documentación", "overview"],
      ["API JavaScript", "plugins"],
      ["CLI", "cli"],
    ],
    eyebrow: "Experimental · código abierto · MIT",
    h1: "Tu app web, en Android.",
    lead: "Crea, previsualiza y publica apps Android con el HTML, CSS y JavaScript que ya conoces. koru es un solo binario de Go: sin Java, sin Gradle, sin SDK de Android.",
    primary: "Empieza ahora",
    secondary: "Compárala con Capacitor y Expo",
    terminalLabel: "De cero a tu móvil en cuatro comandos",
    terminal: [
      ["koru create mi-app", "parte de una plantilla"],
      ["cd mi-app", ""],
      ["koru dev", "vista previa con recarga en vivo"],
      ["koru run", "instala y abre en un dispositivo"],
    ],
    copy: "Copiar comandos",
    copied: "Copiado",
    phone: {
      badge: "android · v1.0.0",
      title: "Mi App",
      lead: "Edita www/index.html y guarda.",
      counter: "Contador guardado",
      count: "Cuenta: 3",
      device: "Este dispositivo",
      rows: [
        ["Modelo", "Emulador Android"],
        ["API Android", "30"],
      ],
      actions: ["Vibrar", "Compartir"],
    },
    phoneAlt:
      "La app de la plantilla en un móvil: un contador guardado, datos del dispositivo y acciones nativas",
    featuresTitle: "Todo el ciclo, de principio a fin",
    features: [
      [
        "magic-wand",
        "Empieza en segundos",
        "koru create escribe un proyecto listo desde una plantilla: HTML simple, Vite, React, Vue, Svelte o TypeScript, con JavaScript tipado, una clave de firma y un .gitignore que la mantiene privada.",
      ],
      [
        "arrows-clockwise",
        "Recarga en vivo en el navegador",
        "koru dev sirve tu app con recarga en vivo e imprime un código QR para abrirla en tu móvil. La misma API JavaScript usa implementaciones web, así que storage, share y las notificaciones funcionan antes de compilar.",
      ],
      [
        "device-mobile",
        "Un comando hasta tu móvil",
        "koru run compila, instala y abre la app en un móvil o emulador conectado, y explica qué corregir cuando algo falla. koru logs muestra los logs de la app en tu terminal.",
      ],
      [
        "puzzle-piece",
        "Funciones de Android desde JavaScript",
        "Información del dispositivo, almacenamiento persistente, vibración, el menú de compartir y notificaciones, con promesas, errores con código y declaraciones TypeScript.",
      ],
      [
        "lightning",
        "Builds rápidos y offline",
        "La plantilla genera un APK firmado de 21 KB en unos 35 ms en nuestra máquina de pruebas. Sin descargas, daemons ni colas en la nube.",
      ],
      [
        "shield-check",
        "Firmado y verificado",
        "Esquema de firma APK v2 con clave persistente, un verificador estricto, origen HTTPS para la app y un puente MessagePort sin addJavascriptInterface.",
      ],
    ],
    codeTitle: "Funciones nativas con JavaScript normal",
    codeLead:
      "Importa el módulo que koru sirve dentro de tu app. En un navegador las mismas llamadas usan implementaciones web, y platform te dice dónde estás.",
    code: `import { device, storage, share, platform } from "/__koru/core.mjs";

const { model, sdk } = await device.getInfo();
await storage.set("theme", "dark");

document.querySelector("#share").onclick = () =>
  share.text(\`Hecho con koru en \${platform}\`);`,
    codeLink: "Lee la API JavaScript",
    frameworksTitle: "Trae tu propio frontend",
    frameworksLead:
      "Empieza con una plantilla de React, Vue, Svelte o TypeScript, o trae cualquier herramienta que genere archivos estáticos: apunta webDir a la carpeta de salida. Conserva la recarga de tu framework con koru dev --proxy.",
    frameworks: [
      "Vite",
      "React",
      "Vue",
      "Svelte",
      "Astro",
      "Solid",
      "JS sin framework",
    ],
    frameworksLink: "Guía de frameworks",
    compareTitle: "Cómo se compara koru",
    compareLead:
      "Capacitor y Expo son maduros y cubren mucho más terreno. koru cambia esa amplitud por una cadena de herramientas Android mínima y autónoma.",
    compareHead: ["", "koru", "Capacitor", "Expo"],
    compareRows: [
      ["Escribes", "HTML, CSS, JS", "HTML, CSS, JS", "React Native"],
      [
        "Compilar Android requiere",
        "el binario de koru",
        "Android Studio, Gradle, JDK",
        "EAS en la nube, o el SDK de Android en local",
      ],
      ["Plataformas", "Android", "Android, iOS, web", "Android, iOS, web"],
      ["Resultado", "APK firmado", "APK, AAB", "APK, AAB"],
      [
        "Plugins",
        "6 integrados + plugins DEX",
        "Ecosistema amplio",
        "SDK amplio",
      ],
    ],
    compareLink: "Comparación detallada",
    statusTitle: "Estado real",
    status:
      "koru es experimental. La ejecución está registrada en un emulador Android 11 (API 30) x86_64; no genera AAB para Google Play ni apps iOS. Prueba en los dispositivos que quieras soportar.",
    statusLink: "Compatibilidad probada",
    ctaTitle: "Crea tu primera app ahora",
    ctaLead: "Instala koru con Go 1.23 o posterior y crea un proyecto.",
    install: `git clone https://github.com/wux4an/koru.git
cd koru && make install`,
    footer: "Hecho para la web. Construido para Android.",
    donate: "Apoya a koru",
    language: "English",
    languageLabel: "Idioma",
    mainNav: "Principal",
    footerNav: "Pie de página",
    tableLabel: "Tabla comparativa",
    appearance: "Apariencia",
    themes: ["Sistema", "Claro", "Oscuro"],
  },
};

const icon = (name: string) =>
  `<i class="ph-duotone ph-${name}" aria-hidden="true"></i>`;

function phone(lang: Locale) {
  const p = copy[lang].phone;
  return `<figure class="phone" aria-label="${escape(copy[lang].phoneAlt)}" role="img">
<div class="phone-screen" aria-hidden="true">
<span class="phone-badge">${escape(p.badge)}</span>
<strong class="phone-title">${escape(p.title)}</strong>
<span class="phone-lead">${escape(p.lead)}</span>
<span class="phone-card"><b>${escape(p.counter)}</b><span class="phone-button">${escape(p.count)}</span></span>
<span class="phone-card"><b>${escape(p.device)}</b>${p.rows.map(([k, v]) => `<span class="phone-row"><span>${escape(k)}</span><span>${escape(v)}</span></span>`).join("")}</span>
<span class="phone-card phone-actions">${p.actions.map((a) => `<span class="phone-button">${escape(a)}</span>`).join("")}</span>
</div>
</figure>`;
}

export function landing(
  lang: Locale,
  config: { url: string; base: string; production: boolean },
) {
  const t = copy[lang];
  const other: Locale = lang === "en" ? "es" : "en";
  const base = config.base;
  const doc = (slug: string) => route(base, lang, slug);
  const canonical = new URL(landingRoute(base, lang), config.url).href;
  const commands = t.terminal.map(([c]) => c).join("\n");
  const structured = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "koru",
    applicationCategory: "DeveloperApplication",
    operatingSystem: "Linux, macOS, Windows",
    description: t.description,
    inLanguage: lang,
    url: canonical,
    license: "https://opensource.org/licenses/MIT",
    offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
  }).replaceAll("<", "\\u003c");
  return `<!doctype html>
<html lang="${lang}"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>${escape(t.title)}</title><meta name="description" content="${escape(t.description)}">
<meta name="theme-color" content="#F7F9FD"><meta name="robots" content="${config.production ? "index,follow" : "noindex,follow"}">
<link rel="canonical" href="${canonical}">
<link rel="alternate" hreflang="en" href="${new URL(landingRoute(base, "en"), config.url).href}">
<link rel="alternate" hreflang="es" href="${new URL(landingRoute(base, "es"), config.url).href}">
<link rel="alternate" hreflang="x-default" href="${new URL(landingRoute(base, "en"), config.url).href}">
<meta property="og:type" content="website"><meta property="og:site_name" content="koru"><meta property="og:title" content="${escape(t.title)}"><meta property="og:description" content="${escape(t.description)}"><meta property="og:url" content="${canonical}"><meta property="og:locale" content="${lang === "en" ? "en_US" : "es_ES"}">
<meta name="twitter:card" content="summary"><meta name="twitter:title" content="${escape(t.title)}"><meta name="twitter:description" content="${escape(t.description)}">
<link rel="icon" href="${base}assets/favicon.svg" type="image/svg+xml">
<link rel="preload" as="font" type="font/woff2" href="${base}assets/fonts/nunito-latin-wght-normal.woff2" crossorigin>
<script src="${base}assets/theme.js"></script>
<link rel="stylesheet" href="${base}assets/vendor/phosphor/style.css"><link rel="stylesheet" href="${base}assets/theme.css"><link rel="stylesheet" href="${base}assets/landing.css">
<script type="application/ld+json">${structured}</script>
</head><body class="landing">
<a class="skip-link" href="#main">${t.skip}</a>
<header class="site-header landing-header"><a class="brand" href="${landingRoute(base, lang)}" aria-label="koru"><i class="ph-duotone ph-spiral" aria-hidden="true"></i><span>koru</span></a>
<nav class="landing-nav" aria-label="${t.mainNav}">${t.nav.map(([label, slug]) => `<a href="${doc(slug)}">${escape(label)}</a>`).join("")}<a href="https://github.com/wux4an/koru" rel="noopener">${icon("github-logo")}<span>GitHub</span></a></nav>
<nav class="language-nav" aria-label="${t.languageLabel}"><a href="${landingRoute(base, other)}" lang="${other}" hreflang="${other}">${t.language}</a></nav>
<div class="theme-control"><button class="theme-button" data-theme-button data-preference="system" type="button" aria-label="${t.appearance}" aria-haspopup="menu" aria-expanded="false" aria-controls="theme-menu"><i class="ph-duotone ph-desktop" data-theme-icon aria-hidden="true"></i><span data-theme-label>${t.themes[0]}</span><i class="ph-duotone ph-caret-down theme-caret" aria-hidden="true"></i></button><div class="theme-menu" id="theme-menu" data-theme-menu role="menu" aria-label="${t.appearance}" hidden>${(
    [
      ["system", t.themes[0], "desktop"],
      ["light", t.themes[1], "sun"],
      ["dark", t.themes[2], "moon"],
    ] as const
  )
    .map(
      ([value, label, name]) =>
        `<button type="button" role="menuitemradio" tabindex="-1" aria-checked="${value === "system"}" data-theme-choice="${value}" data-label="${label}">${icon(name)}<span>${label}</span><i class="ph-duotone ph-check theme-check" aria-hidden="true"></i></button>`,
    )
    .join("")}</div></div></header>
<main id="main" tabindex="-1">
<section class="landing-hero">
<div class="landing-hero-copy">
<p class="eyebrow">${escape(t.eyebrow)}</p>
<h1>${escape(t.h1)}</h1>
<p class="landing-lead">${escape(t.lead)}</p>
<div class="landing-actions"><a class="button button-primary" href="${doc("getting-started")}">${escape(t.primary)} ${icon("arrow-right")}</a><a class="button" href="${doc("comparison")}">${escape(t.secondary)}</a></div>
<div class="terminal" aria-label="${escape(t.terminalLabel)}">
<div class="terminal-bar"><span class="terminal-dots" aria-hidden="true"><b></b><b></b><b></b></span><span>${escape(t.terminalLabel)}</span><button class="copy-button" type="button" data-copy="${escape(commands)}" data-copied="${escape(t.copied)}">${icon("copy")}<span>${escape(t.copy)}</span></button></div>
<pre tabindex="0"><code>${t.terminal.map(([c, note]) => `<span class="prompt" aria-hidden="true">$</span> ${escape(c)}${note ? `<span class="comment">  # ${escape(note)}</span>` : ""}`).join("\n")}</code></pre>
</div>
</div>
${phone(lang)}
</section>

<section class="landing-section" aria-labelledby="features">
<h2 id="features">${escape(t.featuresTitle)}</h2>
<div class="feature-grid">${t.features.map(([name, title, body]) => `<article class="feature">${icon(name)}<h3>${escape(title)}</h3><p>${escape(body)}</p></article>`).join("")}</div>
</section>

<section class="landing-section landing-split" aria-labelledby="api">
<div><h2 id="api">${escape(t.codeTitle)}</h2><p>${escape(t.codeLead)}</p><a class="text-link" href="${doc("plugins")}">${escape(t.codeLink)} ${icon("arrow-right")}</a></div>
<pre class="code-card" tabindex="0" data-lang="js"><code>${escape(t.code)}</code></pre>
</section>

<section class="landing-section" aria-labelledby="frameworks">
<h2 id="frameworks">${escape(t.frameworksTitle)}</h2>
<p class="section-lead">${escape(t.frameworksLead)}</p>
<ul class="chips">${t.frameworks.map((f) => `<li>${escape(f)}</li>`).join("")}</ul>
<a class="text-link" href="${doc("existing-project")}">${escape(t.frameworksLink)} ${icon("arrow-right")}</a>
</section>

<section class="landing-section" aria-labelledby="compare">
<h2 id="compare">${escape(t.compareTitle)}</h2>
<p class="section-lead">${escape(t.compareLead)}</p>
<div class="table-wrap" tabindex="0" role="region" aria-label="${t.tableLabel}"><table class="compare">
<thead><tr>${t.compareHead.map((h, i) => (i === 0 ? `<td></td>` : `<th scope="col">${escape(h)}</th>`)).join("")}</tr></thead>
<tbody>${t.compareRows.map(([head, ...cells]) => `<tr><th scope="row">${escape(head)}</th>${cells.map((c) => `<td>${escape(c)}</td>`).join("")}</tr>`).join("")}</tbody>
</table></div>
<a class="text-link" href="${doc("comparison")}">${escape(t.compareLink)} ${icon("arrow-right")}</a>
</section>

<section class="landing-section status" aria-labelledby="status">
${icon("flask")}<div><h2 id="status">${escape(t.statusTitle)}</h2><p>${escape(t.status)} <a href="${doc("android-results")}">${escape(t.statusLink)}</a>.</p></div>
</section>

<section class="landing-section landing-cta" aria-labelledby="cta">
<h2 id="cta">${escape(t.ctaTitle)}</h2>
<p>${escape(t.ctaLead)}</p>
<pre class="code-card" tabindex="0" data-lang="sh"><code>${escape(t.install)}
koru create my-app</code></pre>
<div class="landing-actions"><a class="button button-primary" href="${doc("getting-started")}">${escape(t.primary)} ${icon("arrow-right")}</a></div>
</section>
</main>
<footer class="landing-footer"><span>${escape(t.footer)}</span><nav aria-label="${t.footerNav}">${t.nav.map(([label, slug]) => `<a href="${doc(slug)}">${escape(label)}</a>`).join("")}<a href="https://ko-fi.com/R5R81YESSI" rel="noopener">${icon("coffee")} ${escape(t.donate)}</a></nav></footer>
<script src="${base}assets/landing.js"></script>
</body></html>`;
}
