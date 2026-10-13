# Configure your project

`koru.json` lives at the project root. Paths are resolved against that directory unless absolute. Unknown fields, trailing JSON and invalid values are rejected.

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

## Network and permissions

Local content works without `INTERNET`. Add that permission explicitly if your app needs external HTTPS services, and configure the server for the app origin, `https://app.koru.invalid`. External scripts and iframes are blocked; bundle scripts locally.

Enabling `vibration` adds `android.permission.VIBRATE`. Enabling `notifications` adds `android.permission.POST_NOTIFICATIONS`. The notification permission dialog for API 33+ has not been covered by the API 30 execution matrix.

## Assets, routes and cache

Symlinks, private keys among assets, unsafe entries and oversized inputs are rejected. Keep the output APK outside `webDir`. Unknown paths with an extension return 404, including in SPA mode. Service workers use the same asset allowlist.

The runtime DEX cache tracks toolchain, configuration, plugins, paths and MIME types. Changing only web file bytes can reuse DEX. Integrity or structural corruption produces an error. Use `--no-cache` to disable it.

[Signing and security]({{base}}en/security/) · [Custom plugins]({{base}}en/plugins/)
