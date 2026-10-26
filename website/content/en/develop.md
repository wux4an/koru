# Develop with live reload

`koru dev` serves your project to a browser while you work. It reloads the page when you save and runs the koru JavaScript API on web implementations, so most of your app can be built and debugged before you create an APK.

```sh
koru dev            # in the project directory
koru dev ./my-app   # or pass the project
```

```text
koru dev live reload on save; press Ctrl+C to stop
  Local:    http://127.0.0.1:8100
  The koru API uses web implementations here; run `koru run` for the native app.
```

## What happens on save

koru watches the project directory (ignoring `.git`, `.koru` and `node_modules`). When a file changes, every open page reloads. Edits to `koru.json`, such as enabling a plugin, apply on the next reload too.

Pages are served from `webDir` with the same rules as the APK: hidden files and symbolic links are never served, and `"spa": true` returns `index.html` for routes without a file extension. koru adds two scripts to each HTML page: `/__koru/dev.js` (reload and your app’s details) and `/__koru/core.js` (the API).

## The JavaScript API in a browser

Calls work the same as on Android, with browser equivalents:

| API             | In the browser                                                                  |
| --------------- | ------------------------------------------------------------------------------- |
| `device`        | `model` is `"Web browser"`, `sdk` is `0`.                                       |
| `app`           | Your `appId`, `name`, `version` and `versionCode` from `koru.json`.             |
| `storage`       | `localStorage`, with keys prefixed `koru:`. The same key and value rules apply. |
| `vibration`     | `navigator.vibrate()` where available (Android Chrome); otherwise it resolves.  |
| `share`         | The browser’s share sheet, or copies the text to the clipboard as a fallback.   |
| `notifications` | The browser’s Notification API and permission prompt.                           |

Use `platform` to tell the two apart:

```js
import { platform, isNative } from "/__koru/core.mjs";

if (isNative()) {
  // Running inside the APK on Android.
}
console.log(platform); // "android" or "web"
```

Under `koru dev`, a call to a plugin that is **not** enabled in `koru.json` rejects with `METHOD_NOT_FOUND`, just like the APK. You find the mistake before building.

> Web implementations are for development. Always test the real behavior with [`koru run`]({{base}}en/run-on-device/): permissions, the Android share sheet, notifications and vibration differ between a browser and Android.

## Try it on a phone’s browser

```sh
koru dev --host 0.0.0.0
```

koru also prints `Network:` addresses, such as `http://192.168.1.20:8100`. Open one in the phone’s browser while both devices are on the same network. This is useful for layout and touch checks; the API still uses web implementations. Only use `0.0.0.0` on networks you trust: anyone on the network can read the served files.

## Use your framework’s dev server

Vite, Astro, SvelteKit and similar tools already have hot module replacement. Keep it, and let koru add its API and app details in front:

```sh
npm run dev                                  # e.g. Vite on http://localhost:5173
koru dev --proxy http://localhost:5173       # in a second terminal
```

Open `http://127.0.0.1:8100`. koru forwards every request except `/__koru/…` to the framework server, adds its scripts to HTML pages and passes WebSocket upgrades through, so hot reload keeps working.

The [`vite` template]({{base}}en/existing-project/) includes a copy of the JavaScript module in `src/koru/`, so `npm run dev` alone already has the API on web implementations. Use `--proxy` when you also want plugin checks against `koru.json` and your real app details.

## Options

| Flag            | Default     | Meaning                                                    |
| --------------- | ----------- | ---------------------------------------------------------- |
| `--port <n>`    | `8100`      | Port to listen on. `0` picks a free port.                  |
| `--host <addr>` | `127.0.0.1` | Address to listen on. `0.0.0.0` exposes it to the network. |
| `--proxy <url>` | none        | Forward to another dev server instead of serving `webDir`. |

`koru dev` never builds an APK and is not part of the APK toolchain. Next: [run on a device]({{base}}en/run-on-device/).
