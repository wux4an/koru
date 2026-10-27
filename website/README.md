# koru documentation website

English and Spanish documentation for people creating Android apps with koru. Built with Docsify 4 and Docsify Themeable, with a soft lavender reading theme. Existing Docsify navigation, search and copy-code controls are reused. Phosphor duotone icons, fonts and application scripts are served locally. The sidebar links to the Ko-fi donation page.

## Develop

Requires Bun (tested with 1.2.13) and a fresh checkout of the repository:

```sh
cd website
bun install --frozen-lockfile
bun run dev
```

Visit http://localhost:4173/docs/en/ or http://localhost:4173/docs/es/. The site builds before serving. After editing Markdown, CSS or scripts, rebuild and refresh; this intentionally small server does not provide hot reload.

## Build and deploy

```sh
SITE_URL=https://koru.js.org/ bun run build
bun run preview
```

Replace the URL with the real public HTTPS domain and optional subpath. Deploy the contents of `dist/` at that path. Ordinary static directory indexes are sufficient: every localized route has its own `index.html` and `README.md`. Do not configure an SPA catch-all that turns missing URLs into HTTP 200. Serve `404.html` with HTTP 404.

Without `SITE_URL`, the build uses localhost and marks preview pages `noindex`. With a real HTTPS `SITE_URL`, it emits indexable HTML, absolute self-canonicals, reciprocal English/Spanish/x-default `hreflang`, Open Graph metadata, TechArticle JSON-LD, robots.txt and a sitemap with every localized page plus both landing pages. No domain or ranking is fabricated. The public URL must be set correctly before deployment. When deploying under a subpath, merge the generated robots directives into the host’s domain-root `/robots.txt`; crawlers do not use a robots file inside a subdirectory. The sitemap can stay under the subpath.

The landing pages (`scripts/landing.ts`) live at `/en/` and `/es/`, the domain root redirects to one by browser language, and the documentation lives under `/docs/en/` and `/docs/es/` (old `/<lang>/<page>/` URLs redirect there); its copy lives in that file rather than in Markdown. Language switching in the documentation preserves the current document. The build also publishes `internal/config/koru.schema.json` at `/schema/koru.schema.json` for editors. Other languages can be added to `scripts/catalog.ts`, with matching content and UI translations. Missing content fails the build rather than silently using another language.

Text records in `original/` have noindex HTML viewers, with download links preserving their original bytes. Example HTML/SVG sources are stored as `.txt` downloads so they cannot execute as another app on the documentation origin. This works on GitHub Pages, which cannot set custom response headers. `robots.txt` excludes the raw archive; its records remain accessible to readers.

See [DEPLOYMENT.md](DEPLOYMENT.md) for the Pages workflow and the live `koru.js.org` domain. Rebuild after changing the Pages custom domain so generated routes and canonical URLs use the new base path.

## Appearance

Use the appearance menu in the header to choose System, Light or Dark (Sistema, Claro u Oscuro in Spanish). System follows changes to the OS preference; explicit choices persist across pages, language switches and tabs. The theme initializes before CSS to avoid a light flash. If preference storage is unavailable, the menu still works for the current page. With JavaScript disabled, the site retains its light reading theme and hides the inactive control.

Long guides with at least three main sections include a collapsible, localized “On this page” index above the first section. It is part of the prerendered HTML for readers without JavaScript and is rebuilt after Docsify navigation so links always match the rendered headings.

## Content and provenance

English guides are in `content/en/`; original Spanish reference pages are read directly from the repository at build time unless a richer localized page exists in `content/es/`. New Spanish guides live in `content/es/`. The public navigation prioritizes first steps, app guides and reference. Architecture, DEX and JNI live under Inside koru. Historical reports keep their routes and are reachable through the archive. The build includes every Markdown file from `docs/`, plus the root README, engineering instructions, testing/native/example READMEs, raw results and example sources. `original/manifest.json` records SHA-256 hashes. Original documents retain their historical language and claims; the reading guides explicitly distinguish early DEX-only evidence from later APK execution.

When changing a reference, update the English counterpart. Add every new page to `scripts/catalog.ts`; both locales must be present. The build rejects unresolved relative documentation links. The tests check locale parity, coverage, source hashes and metadata. Run the build again after changing source docs or validation notes.

## Validate

```sh
bun run build
bun test
bun run test:browser
bun run format:check
```

Browser tests launch the installed `google-chrome-stable` with Playwright, rather than downloading a separate Chromium. Set `CHROME_PATH` if the executable is elsewhere. They start their own preview server; port 4183 must be available. Checks cover every document route, the landing pages, locale switching, locale-scoped search, copying, keyboard focus and section-index navigation, responsive overflow, axe accessibility, HTTP 404, no-JavaScript reading, local-only assets and metadata updates. Screenshots and the JSON report go to ignored `test-results/`. See [VALIDATION.md](VALIDATION.md) for recorded results.

Website tooling is separate from koru's Go compiler and never enters an APK build. Upstream licenses are copied into `dist/assets/licenses/`. See [DESIGN.md](DESIGN.md) for design decisions and official research references.

## Hero terminal recording

`assets/hero-demo.gif` is a real CLI run recorded from [scripts/hero.tape](scripts/hero.tape) with [Charm VHS](https://github.com/charmbracelet/vhs). It initializes a temporary project, builds a signed APK, and verifies it. The recording's private signing identity stays in the temporary directory; it is never a website asset. A static frame is served when the reader prefers reduced motion. To regenerate, compile `bin/koru`, install VHS and its `ttyd`/FFmpeg dependencies, then run from the repository root:

```sh
KORU_BIN="$PWD/bin/koru" vhs website/scripts/hero.tape
gifsicle --resize-width 960 -O3 --colors 128 website/assets/hero-demo.gif -o /tmp/koru-animation.gif
gifsicle --unoptimize /tmp/koru-animation.gif '#-1' -o /tmp/koru-final.gif
gifsicle --delay=220 /tmp/koru-final.gif --delay=8 /tmp/koru-animation.gif -o /tmp/koru-loop.gif
gifsicle -O3 --colors 128 /tmp/koru-loop.gif -o website/assets/hero-demo.gif
ffmpeg -y -i website/assets/hero-demo.gif -frames:v 1 website/assets/hero-poster.png
```

The first GIF frame holds the completed output briefly before replaying the command sequence. A `<picture>` source substitutes the still poster under `prefers-reduced-motion: reduce`. The same commands and representative output appear as accessible text below it. It is documentation media, never part of an APK.
