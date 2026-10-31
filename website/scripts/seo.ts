// SEO helpers shared by the documentation, landing and root pages.
import { resolve } from "node:path";
import { text, type Locale } from "./catalog";
import { koruVersion } from "./version";

export type SiteConfig = { url: string; base: string; production: boolean };

const repo = resolve(import.meta.dir, "../..");
export const githubURL = "https://github.com/wux4an/koru";

export const escapeHTML = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (char) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        char
      ]!,
  );
export const jsonLD = (value: unknown) =>
  JSON.stringify(value).replaceAll("<", "\\u003c");
export const absolute = (config: SiteConfig, path: string) =>
  new URL(path, config.url).href;

export { koruVersion };

const dates = new Map<string, string | undefined>();
/**
 * Committer date of the newest commit touching any of `files` (repo-relative).
 * Returns undefined when git or the history is unavailable: a made-up date
 * would be worse than none.
 */
export function lastModified(files: string[]) {
  const key = files.join("\0");
  if (dates.has(key)) return dates.get(key);
  let value: string | undefined;
  try {
    const run = Bun.spawnSync(
      ["git", "-C", repo, "log", "-1", "--format=%cI", "--", ...files],
      { stdout: "pipe", stderr: "pipe" },
    );
    const out = run.stdout.toString().trim();
    if (run.exitCode === 0 && /^\d{4}-\d\d-\d\dT/.test(out)) value = out;
  } catch {
    value = undefined;
  }
  dates.set(key, value);
  return value;
}

/** Search engine ownership tags, emitted only when the variables are set. */
export function verificationMeta(env: Record<string, string | undefined>) {
  return (
    [
      ["google-site-verification", env.GOOGLE_SITE_VERIFICATION],
      ["msvalidate.01", env.BING_SITE_VERIFICATION],
      ["yandex-verification", env.YANDEX_VERIFICATION],
    ] as const
  )
    .filter(([, value]) => value)
    .map(
      ([name, value]) =>
        `<meta name="${name}" content="${escapeHTML(value!)}">`,
    )
    .join("");
}

/** IndexNow key file name and contents, or undefined when no key is set. */
export function indexNowKey(env: Record<string, string | undefined>) {
  const key = env.INDEXNOW_KEY;
  if (!key) return undefined;
  if (!/^[a-zA-Z0-9-]{8,128}$/.test(key))
    throw new Error("INDEXNOW_KEY must match [a-zA-Z0-9-]{8,128}.");
  return key;
}

export const socialImage = (config: SiteConfig, lang: Locale) =>
  absolute(config, `${config.base}assets/og-${lang}.png`);

export function socialMeta(options: {
  config: SiteConfig;
  lang: Locale;
  title: string;
  description: string;
  url: string;
  type: "website" | "article";
  modified?: string;
}) {
  const { config, lang, title, description, url, type, modified } = options;
  const image = socialImage(config, lang);
  const alt = escapeHTML(text[lang].socialAlt);
  const t = escapeHTML(title);
  const d = escapeHTML(description);
  return `<meta property="og:type" content="${type}"><meta property="og:site_name" content="koru"><meta property="og:title" content="${t}"><meta property="og:description" content="${d}"><meta property="og:url" content="${url}"><meta property="og:locale" content="${lang === "en" ? "en_US" : "es_ES"}"><meta property="og:locale:alternate" content="${lang === "en" ? "es_ES" : "en_US"}">
<meta property="og:image" content="${image}"><meta property="og:image:type" content="image/png"><meta property="og:image:width" content="1200"><meta property="og:image:height" content="630"><meta property="og:image:alt" content="${alt}">${modified ? `<meta property="article:modified_time" content="${modified}">` : ""}
<meta name="twitter:card" content="summary_large_image"><meta name="twitter:title" content="${t}"><meta name="twitter:description" content="${d}"><meta name="twitter:image" content="${image}"><meta name="twitter:image:alt" content="${alt}">`;
}

const websiteID = (config: SiteConfig) => config.url + "#website";
const organizationID = (config: SiteConfig) => config.url + "#organization";

/** WebSite and Organization nodes, present in every page's @graph. */
export function siteNodes(config: SiteConfig) {
  return [
    {
      "@type": "WebSite",
      "@id": websiteID(config),
      name: "koru",
      url: config.url,
      inLanguage: ["en", "es"],
      publisher: { "@id": organizationID(config) },
    },
    {
      "@type": "Organization",
      "@id": organizationID(config),
      name: "koru",
      url: config.url,
      logo: absolute(config, `${config.base}assets/favicon.svg`),
      sameAs: [githubURL],
    },
  ];
}

export function docsGraph(options: {
  config: SiteConfig;
  lang: Locale;
  slug: string;
  title: string;
  description: string;
  url: string;
  modified?: string;
  docsURL: string;
  homeURL: string;
}) {
  const { config, lang, slug, title, description, url, modified } = options;
  const t = text[lang];
  const crumbs = [
    [t.home, options.homeURL],
    [t.docs, options.docsURL],
    ...(slug === "overview" ? [] : [[title, url]]),
  ];
  return {
    "@context": "https://schema.org",
    "@graph": [
      ...siteNodes(config),
      {
        "@type": "BreadcrumbList",
        itemListElement: crumbs.map(([name, item], index) => ({
          "@type": "ListItem",
          position: index + 1,
          name,
          item,
        })),
      },
      {
        "@type": "TechArticle",
        "@id": url + "#article",
        headline: title,
        description,
        inLanguage: lang,
        url,
        mainEntityOfPage: url,
        image: socialImage(config, lang),
        ...(modified ? { dateModified: modified } : {}),
        isPartOf: { "@id": websiteID(config) },
        publisher: { "@id": organizationID(config) },
      },
    ],
  };
}

export function landingGraph(options: {
  config: SiteConfig;
  lang: Locale;
  description: string;
  url: string;
}) {
  const { config, lang, description, url } = options;
  return {
    "@context": "https://schema.org",
    "@graph": [
      ...siteNodes(config),
      {
        "@type": "SoftwareApplication",
        "@id": url + "#software",
        name: "koru",
        applicationCategory: "DeveloperApplication",
        operatingSystem: "Linux, macOS, Windows",
        softwareVersion: koruVersion(),
        description,
        inLanguage: lang,
        url,
        image: socialImage(config, lang),
        sameAs: [githubURL],
        license: "https://opensource.org/licenses/MIT",
        offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
        publisher: { "@id": organizationID(config) },
      },
    ],
  };
}

export type SitemapEntry = {
  loc: string;
  lastmod?: string;
  alternates: [string, string][];
};
export function sitemapXML(entries: SitemapEntry[]) {
  const body = entries
    .map(
      (entry) =>
        `<url><loc>${escapeHTML(entry.loc)}</loc>${entry.lastmod ? `<lastmod>${entry.lastmod}</lastmod>` : ""}${entry.alternates
          .map(
            ([hreflang, href]) =>
              `<xhtml:link rel="alternate" hreflang="${hreflang}" href="${escapeHTML(href)}"/>`,
          )
          .join("")}</url>`,
    )
    .join("\n");
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">\n${body}\n</urlset>\n`;
}
