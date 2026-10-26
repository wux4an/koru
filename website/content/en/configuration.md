# Configure your project

Set your app’s name, icon, web files and permissions in `koru.json`, at the root of the project that `koru create` or `koru init` made. Paths in this file start at that folder.

## A complete starting point

```json
{
  "$schema": "https://koru.js.org/schema/koru.schema.json",
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

`create` fills in `name` and `appId` from your answers and sets `webDir` for the template (`www` or `dist`). `init` uses `www`, `com.example.koru` and `Koru`. The example above selects a framework output folder and enables SPA routing.

Run `koru doctor` after changing the file. A successful check names your configured `appId`; a full build additionally scans all web assets. Keep `koru.json` at the project root, next to the configured web folder. Paths in this file are relative to that project root, while CLI `--output` is relative to your current shell directory.

## Editor autocompletion

The `$schema` line points editors such as VS Code at koru’s [JSON Schema]({{base}}schema/koru.schema.json ":ignore"). You get completion, descriptions on hover and errors for invalid values while you type. koru ignores the field when building; projects created by `create` and `init` include it.

When a value is invalid, koru names the field and the fix, for example:

```text
koru: koru.json: invalid appId "myapp": use reverse-domain form such as com.example.myapp (letters, digits and _; at least two segments)
```

## Fields

| Field             | Meaning and validation                                                                                                           |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `$schema`         | Optional editor hint; ignored by koru.                                                                                           |
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
| `orientation`     | Optional: `portrait`, `landscape` or `sensor`. Omit it to keep Android’s default.                                                |
| `allowBackup`     | Optional boolean, default `false`. `true` opts the app into Android’s automatic backup.                                          |
| `keyboard`        | Optional: `resize` (default; the page shrinks) or `pan` (the page slides up) when the on-screen keyboard opens.                  |
| `deepLinks`       | Optional list of links that open the app. At most 32, no duplicates. See [Deep links](#deep-links) below.                        |

The defaults are written by `init`, so edit the generated file rather than guessing missing values. A custom plugin ID must appear in `plugins` and have a corresponding manifest in `pluginManifests`. The six built-in IDs and their methods are in [JavaScript and plugins]({{base}}en/plugins/).

## Screen, keyboard and backup

```json
{
  "orientation": "portrait",
  "keyboard": "pan",
  "allowBackup": false
}
```

This is a fragment to merge into your `koru.json`. `orientation` locks the screen: `portrait`, `landscape`, or `sensor` (follows the device sensor). `keyboard` decides what the on-screen keyboard does to a page with a text field near the bottom: `resize` shrinks the page above the keyboard, `pan` slides the whole page up instead. `allowBackup` is off unless you set it; turn it on only if you want Android to back up the app’s data to the user’s Google account.

An invalid value stops the build and says how to fix it:

```text
koru: koru.json: invalid orientation "upside": use "portrait", "landscape" or "sensor", or remove the field to keep Android's default
```

`koru build` writes these into the generated manifest, and the attribute IDs were checked against the Android sources. None of them has been run on a device yet: the recorded API 30 matrix does not cover rotation, the keyboard or backup.

## Deep links

`deepLinks` makes links open your app. Each entry is a `scheme`, and optionally a `host` and a `pathPrefix`:

```json
"deepLinks": [
  { "scheme": "https", "host": "example.com", "pathPrefix": "/app" },
  { "scheme": "myapp" }
]
```

| Entry                                                                | Opens the app for                                         |
| -------------------------------------------------------------------- | --------------------------------------------------------- |
| `{ "scheme": "myapp" }`                                              | Any `myapp://…` link. A custom scheme needs nothing else. |
| `{ "scheme": "https", "host": "example.com" }`                       | Any `https://example.com/…` link.                         |
| `{ "scheme": "https", "host": "example.com", "pathPrefix": "/app" }` | Only paths starting with `/app`. Needs a `host`.          |

Rules koru checks: the scheme is lowercase letters, digits, `+`, `.` or `-`, starting with a letter; `http`, `javascript`, `file` and `content` are rejected; an `https` link needs a full domain in `host` (a leading `*.` wildcard is allowed); `pathPrefix` starts with `/`, has no spaces, `?` or `#`, and needs a `host`. Errors name the entry, for example:

```text
koru: koru.json: forbidden deepLinks scheme "http": plain HTTP is not allowed: use "https"
```

**The link opens the app on its initial page. The URL is not passed to your JavaScript yet**, so the app cannot read which link was tapped. Today deep links are useful for launching the app from a web link, not for routing to a screen.

### Verify `https` links with assetlinks.json

For `https` links, Android opens the app without a chooser only after it verifies that your domain trusts the app. koru marks these links for verification (`autoVerify`) and prints the file the domain must serve:

```sh
koru deeplinks ./my-app --json > assetlinks.json
```

```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.example.myapp",
      "sha256_cert_fingerprints": [
        "86:CC:81:42:72:1E:56:ED:50:68:3F:39:CF:48:64:56:27:90:D1:6E:54:32:B1:C8:12:C4:23:4B:5C:70:C2:B1"
      ]
    }
  }
]
```

The fingerprint is the SHA-256 of the certificate in your `signingKey` (this one comes from a throwaway example key). Publish the file at `https://example.com/.well-known/assetlinks.json` for every `https` domain in `deepLinks`, over HTTPS with status 200, content type `application/json` and no redirects. A `*.example.com` wildcard needs the file on every subdomain that should open the app. If you replace the signing key, the fingerprint changes and verification fails; see [Signing and security]({{base}}en/security/). Without the file, or until Android has checked it, a link may show a chooser instead of opening the app directly.

`koru verify app.apk` lists the deep links found in the manifest. **Nothing here has been run on Android:** koru has not recorded an install, a verification or a tapped link. Test on a device before relying on it.

## Point to frontend output

Build your frontend before invoking koru. For a Parcel or Vite project rooted at `my-app/` whose build writes `my-app/dist/index.html`, set `"webDir": "dist"`, then run:

```sh
npm --prefix ./my-app run build
koru doctor ./my-app
koru build ./my-app --release --output app.apk
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
