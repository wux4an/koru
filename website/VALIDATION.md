# Documentation rework validation

Validated locally on 2026-10-14 with Bun 1.2.13 and the installed
`google-chrome-stable` (Google Chrome 136.0.7103.113), launched through Playwright.
No Playwright browser download was used.

## Website checks

- `bun test`: 10 tests passed with 2,489 assertions, covering all 40 localized
  routes, source archive hashes, translated metadata, links, section indexes,
  sitemap and invalid input.
- `bun run test:browser`: all 22 checks passed on 40 English and Spanish routes.
  Checks include locale-scoped search, language switching, copy controls,
  keyboard focus, the mobile drawer, theme persistence, reduced motion, direct
  routes, browser history, subpath hosting and reading without JavaScript.
- Axe checks passed with zero violations in 17 light/dark scans, including the
  appearance menu. These automated results are not an accessibility certification.
- Responsive checks passed at 320, 390, 768, 1024 and 1440 px on the overview and
  reference pages; code and wide tables scroll within their containers.
- All 13 JavaScript API accordions work in both languages.
- No external asset requests, JavaScript errors or failed requests occurred in
  the checked browser flows. Missing routes return HTTP 404.
- `bun run format:check` and `git diff --check` pass.

The Chrome review covered desktop, phone and dark screenshots. It caught an icon
alignment issue, nested code font sizing and insufficient JSON syntax contrast in
dark mode; each was corrected. Final navigation surface adjustments were also
checked on the rebuilt artifact with Chrome and axe.

Ignored review artifacts live in `test-results/`, including `browser-report.json`,
`rework-desktop.png`, `rework-mobile.png`, `rework-dark.png` and the standard
browser-suite screenshots. The source archive preserves original documents,
including historical engineering results.

## CLI examples and repository checks

The documented source-build, `init`, `doctor`, debug `build`, release `build` and
`verify` commands succeeded in a temporary project. The generated private key
stayed in that temporary directory and was removed with it. The debug and release
APKs used the same signing certificate.

`go test ./...`, `go vet ./...` and `go test -race ./...` passed using the actual
Go compiler at `/etc/profiles/per-user/wux4an/bin/go`. The workspace's documented
`/workspace/tooling/go/bin/go` path is absent in this session. No Go source changed.
No keys or `.koru/android-demo` directories were added.

This documentation work does not add Android execution evidence. APK installation
and device behavior remain covered only by the existing matrix in
[the Android results](../docs/results/full/README.md). The rework is local; no new
production deployment was performed.
