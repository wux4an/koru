# Changelog

All notable changes to koru are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and koru uses
[Semantic Versioning](https://semver.org/). Before 1.0.0, minor versions may
break the CLI, `koru.json` or the JavaScript API.

## [Unreleased]

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

[Unreleased]: https://github.com/wux4an/koru/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/wux4an/koru/releases/tag/v0.0.1
