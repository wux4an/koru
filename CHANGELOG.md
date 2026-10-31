# Changelog

All notable changes to koru are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and koru uses
[Semantic Versioning](https://semver.org/). Before 1.0.0, minor versions may
break the CLI, `koru.json` or the JavaScript API.

## [Unreleased]

## [0.1.0] - 2026-10-31

Remote content, web APIs and new native plugins. All of it is opt-in in
`koru.json`: without the new fields the runtime bytecode and the manifest are
byte-identical to 0.0.3. None of it has run on an Android device yet; the
self-test app (`selftest.apk`) covers each feature so it can be checked on a
phone.

### Added

- `server` in `koru.json`: `url` opens a website as the app, `allowNavigation`
  lists the hosts that open inside the app (hybrid apps), and `bridge` gives
  the koru API to exact https origins. Links to any other site open in the
  system browser or app (https, http, mailto, tel, sms, geo, market).
- `webFeatures` in `koru.json`: `files` (`<input type="file">`, with multiple
  selection), `geolocation`, `camera` and `microphone` (`getUserMedia`).
  Location, camera and microphone ask Android for the permission and only
  answer the app origin and `server.bridge` origins.
- Plugins `alarms` (scheduled notifications that survive a reboot),
  `browser` (`browser.open`), `screen` (`screen.keepAwake`) and `clipboard`
  (`writeText`, `readText`), with web versions under `koru dev`.
- `koru build` type-checks all the bytecode, runtime and custom plugins, with
  a verifier modelled on ART's, and refuses an APK that Android would reject
  when it loads the class.
- The self-test app checks navigation, the remote bridge (through
  `https://koru.js.org/selftest/remote/`), the web features and the new
  plugins.

## [0.0.3] - 2026-10-31

### Added

- `examples/selftest`, a koru app that runs every automatic check of the
  bridge, the integrated plugins, a custom plugin and the WebView when it
  opens, and reports the result on screen, in `koru logs`
  (`KORU_SELFTEST …`) and as `window.__KORU_SELFTEST__`.
- Every release publishes it as `selftest.apk`, and CI keeps the APK of each
  build as an artifact.

## [0.0.2] - 2026-10-31

### Added

- Homebrew and Scoop packages, served from this repository:
  `brew tap wux4an/koru https://github.com/wux4an/koru && brew install koru`
  and `scoop bucket add koru https://github.com/wux4an/koru && scoop install koru`.
  Every release updates them.

### Changed

- The Go module is now `github.com/wux4an/koru`, so
  `go install github.com/wux4an/koru/cmd/koru@latest` works.

### Fixed

- `koru build` on Windows no longer rejects the signing identity because of
  Unix file permissions, which Windows does not have.

## [0.0.1] - 2026-10-31

First public release. Experimental: APK installation and execution are
recorded only on Android 11 (API 30) on an x86_64 emulator.

### Added

- `koru build` turns a static web project into a signed APK (APK Signature
  Scheme v2) with pure Go DEX, AXML, ARSC and ZIP writers: no Java, Gradle or
  Android SDK. `koru verify` checks the result.
- Runtime that serves the app from an intercepted HTTPS origin and talks to
  JavaScript over a `MessagePort`, without `addJavascriptInterface`.
- Plugins for device information, storage, sharing, notifications and
  vibration, with a typed JavaScript API (`core.d.ts`) that falls back to web
  implementations in the browser.
- `koru create` with the `basic`, `vite`, `react`, `vue`, `svelte` and
  `vanilla-ts` templates.
- `koru dev` with live reload, a proxy for framework dev servers and a QR code
  for phones.
- `koru run`, `koru devices` and `koru logs`, speaking the adb server protocol
  directly.
- `koru bump`, `koru plugin`, `koru config`, `koru deeplinks`, `koru doctor`,
  `koru init`, `koru keygen` and `koru js`.
- `koru.json` options for orientation, backups, keyboard behaviour and deep
  links, with a JSON Schema for editors.
- Release archives for Linux, macOS and Windows (amd64 and arm64) with SHA-256
  checksums and build provenance attestations, and installers for
  `curl … | sh` and PowerShell.

[Unreleased]: https://github.com/wux4an/koru/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/wux4an/koru/releases/tag/v0.1.0
[0.0.3]: https://github.com/wux4an/koru/releases/tag/v0.0.3
[0.0.2]: https://github.com/wux4an/koru/releases/tag/v0.0.2
[0.0.1]: https://github.com/wux4an/koru/releases/tag/v0.0.1
