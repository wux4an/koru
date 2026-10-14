# A small workshop for Android apps

Koru turns static web assets into a signed Android package. Its documentation is
for developers who need runnable examples, precise format references, and honest
Android evidence. The design makes that transformation tangible.

## Tokens

- Porcelain `#F7F9FD`: reading surface.
- Ink `#263B66`: text and code.
- Iris `#A79CE8`: molded hero objects and active navigation.
- Mint `#BCE4CC`: Android object and success context.
- Peach `#F2B9A5`: small supporting accents.
- White `#FFFFFF`: inset paper and highlights.

Nunito Variable supplies rounded, readable headings and body text. JetBrains Mono
is reserved for commands and code. Both are self-hosted. Phosphor duotone icons
share the rounded geometry; they replace decorative emoji.

## Layout

Left-aligned documentation lives on a quiet reading sheet, with Docsify's native
sidebar, search and copy controls. The overview pairs a large, plain headline with
a tactile browser-to-Android illustration. The illustration is the one bold area.

```text
┌ logo / documentation ───────────────── language / source ┐
│ search       │ headline                browser → Android │
│              │ introduction           molded illustration│
│ guide        │ get started / architecture                │
│ reference    │ VHS CLI recording                          │
│              │ runnable examples + expected output      │
│ evidence     │ readable documentation                    │
└──────────────┴───────────────────────────────────────────┘
```

On phones the illustration gets smaller and the native sidebar becomes a drawer.
Texture is confined to the molded objects, not behind text. Shadows establish
physical depth; they are not applied to every paragraph or section. Motion only
responds to user actions and respects reduced motion.

## Review against the brief

The first sketch spread raised clay cards across every section. That impeded
scanning and looked like a generic card grid. The final plan keeps Docsify's
existing reading and navigation structures and concentrates the stop-motion
reference in the overview. Cold porcelain, iris and mint replace the predictable
warm terracotta palette. No custom component framework is introduced.

## Research

- [Docsify configuration](https://docsify.js.org/#/configuration): history routing,
  locale sidebars and plugins. Docsify 4.13.1 is pinned to match Themeable's v4
  integration; a v5 migration needs a separate compatibility review.
- [Docsify Themeable](https://jhildenbiddle.github.io/docsify-themeable/#/): reuse
  the existing theme and override its documented design tokens.
- [Google JavaScript SEO](https://developers.google.com/search/docs/crawling-indexing/javascript/javascript-seo-basics):
  serve document text in HTML and use real URLs rather than hash routes.
- [Google localized pages](https://developers.google.com/search/docs/specialty/international/localized-versions):
  bidirectional, absolute `hreflang` links and self-canonical pages.
- [Google robots.txt location](https://developers.google.com/crawling/docs/robots-txt/create-robots-txt):
  subpath deployments must put crawler directives at the domain root.
- [Phosphor Web](https://github.com/phosphor-icons/web): reusable duotone icons,
  self-hosted and licensed under MIT.

- [Parcel](https://parceljs.org/): concise examples adjacent to the feature they
  explain, with immediate commands and visible output. Koru uses that teaching
  pattern while keeping Docsify's native code blocks and copy controls.
- [Charm VHS](https://github.com/charmbracelet/vhs): a reproducible terminal tape
  records real CLI commands as a compact GIF for the overview. A static frame
  replaces the animation under the reader's reduced-motion preference.
- [Expo documentation](https://docs.expo.dev/versions/latest/sdk/): per-method
  reference pages make arguments, return values and examples easy to find. Koru
  uses native HTML details for the equivalent API reference, keeping all method
  text in the initial HTML and available without JavaScript.

SEO metadata belongs to each generated HTML document. Docsify enhances that
document with search and navigation. A JavaScript-free reader receives the same
content and working document links. Missing routes return HTTP 404.

## Dark appearance

The night palette uses midnight blue `#171F32`, slate paper `#202B41`, pale ink
`#E2E8F5`, secondary ink `#B9C5DC`, lavender links `#C3B5F5` and mint accents.
Clay objects retain their pastel surfaces but cast darker shadows. Code, tables,
search results, controls and focus rings share the same palette. A custom menu button with checked options
reuses the header's inset treatment and keeps the mobile header compact.

The initial sketch merely inverted the reading surface, leaving bright search
fields and weak syntax contrast. The final palette covers those surfaces and
uses separate readable syntax colors.

Browser behavior follows [MDN color-scheme](https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/Properties/color-scheme)
and [matchMedia](https://developer.mozilla.org/en-US/docs/Web/API/Window/matchMedia).
The preference bootstrap runs before CSS, with system preference as the default
and optional storage for explicit choices.

The custom menu replaces the browser select at the user's request. It follows the
[WAI-ARIA menu button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/)
with radio menu items, arrow keys, Home/End, Escape, Tab dismissal and returned
focus. The appearance choices are styled entirely with the site theme.
