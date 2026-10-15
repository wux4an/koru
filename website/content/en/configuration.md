# Configure your project

Set your app’s name, icon, web files and permissions in `koru.json`. You will find it in the folder created by `koru init`. Paths in this file start at that folder.

## A complete starting point

```json
{
  "appId": "com.example.demo",
  "name": "Demo",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "./dist",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app"],
  "signingKey": ".koru/signing.pem",
  "spa": true
}
```

`init` uses `www` for `webDir`, `com.example.koru` for `appId`, `Koru` for `name`, and leaves SPA mode off. The example above selects a framework output folder and enables SPA routing.

Run `./bin/koru doctor ./my-app` after changing the file. A successful check names your configured `appId`; a full build additionally scans all web assets. Keep `koru.json` at the project root, next to the configured web folder. Paths in this file are relative to that project root, while CLI `--output` is relative to your current shell directory.

## Fields

| Field             | Meaning and validation                                                                                                           |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `appId`           | Dotted application identifier, at most 127 bytes. Each segment begins with a letter and contains letters, digits or underscores. |
| `name`            | Nonblank application label, at most 256 bytes.                                                                                   |
| `version`         | Nonempty version label, at most 128 bytes.                                                                                       |
| `versionCode`     | Integer from 1 to 2,100,000,000. Increase it for updates.                                                                        |
| `webDir`          | Required folder of static web assets containing `index.html`.                                                                    |
| `minSdk`          | Minimum Android API; at least 26.                                                                                                |
| `targetSdk`       | Target API; `minSdk ≤ targetSdk ≤ 36`. This is a configuration range, not the executed test matrix.                              |
| `permissions`     | Unique names such as `android.permission.INTERNET`. Use full Android permission identifiers.                                     |
| `plugins`         | Unique enabled plugin IDs. Defaults: `device`, `storage`, `app`. Other built-ins: `vibration`, `share`, `notifications`.         |
| `signingKey`      | Required existing PEM private key and certificate.                                                                               |
| `icon`            | Optional PNG icon path.                                                                                                          |
| `spa`             | Optional boolean; routes without an extension fall back to `index.html`.                                                         |
| `pluginManifests` | Optional paths to custom DEX plugin manifests. Enable their IDs in `plugins`.                                                    |
| `nativeLibraries` | Optional map from Android ABI to lists of precompiled ELF shared libraries.                                                      |

The defaults are written by `init`, so edit the generated file rather than guessing missing values. A custom plugin ID must appear in `plugins` and have a corresponding manifest in `pluginManifests`. The six built-in IDs and their methods are in [JavaScript and plugins]({{base}}en/plugins/).

## Point to frontend output

Build your frontend before invoking koru. For a Parcel or Vite project rooted at `my-app/` whose build writes `my-app/dist/index.html`, set `"webDir": "dist"`, then run:

```sh
npm --prefix ./my-app run build
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output app.apk
```

The `npm` command assumes that project has a `build` script; use the package manager and script your frontend actually provides. Koru does not install packages or build JavaScript. If the output HTML refers to `/assets/app.js`, the APK must contain `dist/assets/app.js`. Configure your frontend's public/base path for the APK origin, not a deployment subdirectory. Do not add `dist` itself to the URL.

## Network and permissions

Local content works without `INTERNET`. Add that permission explicitly if your app needs external HTTPS services, and configure the server for the app origin, `https://app.koru.invalid`. External scripts and iframes are blocked; bundle scripts locally.

```json
"permissions": ["android.permission.INTERNET"]
```

This is a field fragment, not a complete `koru.json`. Keep remote access in your own JavaScript fetch calls and account for CORS on the remote server. The app origin is a reserved local HTTPS origin, not a host you need to register in DNS.

Enabling `vibration` adds `android.permission.VIBRATE`. Enabling `notifications` adds `android.permission.POST_NOTIFICATIONS`. The notification permission dialog for API 33+ has not been covered by the API 30 execution matrix.

## Assets, routes and cache

Symlinks, private keys among assets, unsafe entries and oversized inputs are rejected. Keep the output APK outside `webDir`. Unknown paths with an extension return 404, including in SPA mode. Service workers use the same asset allowlist.

There can be at most 2,000 assets, 64 MiB per file and 256 MiB in total. The builder skips `.git`, `.koru` and `node_modules` directories. Put `signingKey` outside `webDir`; do not rely on the skip rule to protect a misplaced key. For client-side routes, enable `spa` so extensionless unknown paths load `index.html`; missing `.js`, `.css` or image paths remain 404. For multi-page sites, keep `spa` off and include every HTML page.

The runtime DEX cache tracks toolchain, configuration, plugins, paths and MIME types. Changing only web file bytes can reuse DEX. Integrity or structural corruption produces an error. Use `--no-cache` to disable it.

[Signing and security]({{base}}en/security/) · [Custom plugins]({{base}}en/plugins/)
