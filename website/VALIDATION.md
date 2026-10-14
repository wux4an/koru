# Website validation

Recorded in this workspace on 2026-10-13 and rechecked on 2026-10-14 using Bun 1.2.13 and the installed
`google-chrome-stable`, Google Chrome 136.0.7103.113. Playwright launched that
executable; it did not download its own Chromium.

| Check                                   | Recorded result                                                                                                                                                                                                                                                                                                |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bun test`                              | 9 tests passed; 2,291 assertions for locale parity, prerendered content, document coverage, source hashes, links, sitemap, custom-domain root routes, favicon and broken links.                                                                                                                                |
| `bun run test:browser`                  | 20 checks passed, covering all 36 English/Spanish routes.                                                                                                                                                                                                                                                      |
| axe WCAG 2.2 AA and best-practice rules | Zero violations in 17 scans: 10 light-theme scans of overview, plugins, configuration, DEX API and Android evidence in each locale, plus 6 dark-theme scans of overview, configuration and DEX API and a scan with the custom theme menu open. This is an automated check, not an accessibility certification. |
| Responsive layout                       | No document-level horizontal overflow at 320, 390, 768, 1024 or 1440 px on overview, configuration and DEX API. The native mobile drawer opens and closes.                                                                                                                                                     |
| Locale behavior                         | Switch preserves the document; navigation, search and content change language. Search results stay in their locale.                                                                                                                                                                                            |
| Navigation metadata                     | Canonical, description, hreflang and TechArticle JSON-LD update on Docsify navigation; browser Back restores the preceding document.                                                                                                                                                                           |
| JavaScript disabled                     | All 36 documents render text from the initial HTML. Native document and locale links work.                                                                                                                                                                                                                     |
| Static hosting under a subpath          | Direct loading, navigation, search, locales, canonical URL and HTTP 404 pass under `/handbook/`.                                                                                                                                                                                                               |
| Network and console                     | No external asset requests, page errors or failed requests during the checked flows. Fonts, icons, theme and scripts are local.                                                                                                                                                                                |
| Copy and keyboard                       | Code copies to the clipboard; keyboard skip link focuses content. Hidden navigation is excluded from keyboard focus.                                                                                                                                                                                           |
| HTTP behavior                           | Missing pages return 404. Archived HTML/SVG is served as plain .txt downloads; source viewers carry noindex metadata and preserve original bytes. The published domain serves `/en/`, `/es/` and `/favicon.ico` with HTTP 200.                                                                                 |
| Appearance                              | System/light/dark selection, OS changes, persistence across reloads and locales, tab synchronization, invalid preferences and blocked storage pass. Dark layout has no overflow at 320, 390 and 768 px.                                                                                                        |
| Reduced motion                          | Sidebar transitions are disabled under the reduced-motion preference.                                                                                                                                                                                                                                          |
| Hero interaction and media              | Dark-mode visited CTA retains at least 4.5:1 text contrast on hover. Both localized overview pages include the VHS recording; MP4 and poster return HTTP 200.                                                                                                                                                  |
| `bun run format:check`                  | Prettier 3.6.2 passes.                                                                                                                                                                                                                                                                                         |
| `git diff --check`                      | Passes. No keys or private Android demo directories added.                                                                                                                                                                                                                                                     |

The 2026-10-14 live Chrome check loaded both `https://koru.js.org/en/` and
`https://koru.js.org/es/`, confirmed their canonical URLs and rendered content,
and fetched `https://koru.js.org/favicon.ico`; all returned HTTP 200 without
page errors. GitHub Pages deployment run 38632921430 completed successfully.

Reproduce with the commands in [README.md](README.md). Ignored artifacts in
`test-results/` contain `browser-report.json`, desktop/mobile/reference screenshots,
dark desktop/mobile screenshots, a JavaScript-free screenshot and isolated unit/subpath builds. Original repository
records are copied into the site's archive with SHA-256 hashes.

Chrome review found and corrected Themeable's interaction with the language nav,
layout movement on mobile, low contrast in syntax comments, and an invalid
`aria-hidden` value in the existing copy-code plugin. The integration preserves
those upstream controls rather than replacing them.

The repository publication is authorized. Google indexing has not been tested. Set the actual public
HTTPS `SITE_URL` before building for publication. For subpath hosting, merge the
generated crawler directives into the domain-root `/robots.txt` as described in [DEPLOYMENT.md](DEPLOYMENT.md).

Website browser results do not extend koru's Android execution matrix. Existing
Android evidence remains limited to Android 11/API 30/x86_64, as documented in
`docs/results/full`. The CLI agent separately reran Go tests, vet and race tests
for the Cobra migration; those checks do not substitute for Android execution.

The browser search bootstrap race was reproduced on repeated reloads and fixed by
loading enhancement scripts at the end of the document while the parser is still
loading. A delayed search-plugin regression check now passes. The theme menu uses
our markup and styles, checked radio menu items and keyboard navigation; it no
longer opens a browser select.

Before publishing this checkout, `go test ./...`, `go vet ./...` and
`go test -race ./...` passed using Go 1.25.8 from the Nix user profile. `gofmt -l`
reported no files. The untracked user SHA256SUMS file is excluded from the commit.

The updated first-app flow, CLI reference, configuration, JavaScript plugin example, web fixture and troubleshooting pages are available in both languages. The overview shows a real VHS recording of init/build/verify plus accessible textual commands and illustrative output. This website-only change does not alter the Android execution evidence.
