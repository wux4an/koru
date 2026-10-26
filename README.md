<p align="center">
  <img src="website/assets/favicon.svg" width="88" height="88" alt="koru spiral logo" />
</p>

<h1 align="center">koru</h1>

<p align="center"><strong>Your web app, on Android.</strong><br />Create, preview and ship Android apps with HTML, CSS and JavaScript — one Go binary, no Java, Gradle or Android SDK.</p>

<p align="center">
  <img alt="Go 1.23+" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white" />
  <img alt="Android APK" src="https://img.shields.io/badge/Android-APK-3DDC84?logo=android&logoColor=white" />
  <img alt="Experimental" src="https://img.shields.io/badge/status-experimental-e8a838" />
  <img alt="MIT license" src="https://img.shields.io/badge/license-MIT-806bc8" />
</p>

<p align="center">
  <a href="https://koru.js.org/en/getting-started/">Quick start</a> ·
  <a href="https://koru.js.org/en/">Documentation</a> ·
  <a href="https://koru.js.org/es/">Español</a> ·
  <a href="https://koru.js.org/en/plugins/">JavaScript API</a>
</p>

<p align="center">
  <a href="https://ko-fi.com/R5R81YESSI">☕️ Support koru 💛</a>
</p>

## Get started

You need **Go 1.23 or later** to install koru. After that, the `koru` binary is all you need.

```sh
git clone https://github.com/wux4an/koru.git
cd koru && make install          # installs to ~/.local/bin

koru create my-app               # pick a name and a template
cd my-app
koru dev                         # preview in your browser with live reload
koru run                         # install and open on a phone or emulator
```

`koru run` talks to the adb server that Android Studio or the platform tools run; koru never starts external programs. The [first app guide](https://koru.js.org/en/getting-started/) walks through every step.

Keep a secure backup of `.koru/signing.pem`. `koru create` and `koru init` add `.koru/` to the project's `.gitignore` so the key stays out of version control. Android needs the same signing identity for future updates.

## Why koru

- **One binary, no native toolchain.** DEX bytecode, the binary manifest, resources, ZIP and the APK v2 signature are written in Go with its standard library. No Java, Gradle or Android SDK; builds are offline and take milliseconds.
- **The loop you know from Capacitor and Expo.** `create` from a template, `dev` with live reload, `run` on a device, `build --release` to ship.
- **The same API in the browser.** Under `koru dev`, the JavaScript API runs on web implementations, so you can build most of your app before installing it.
- **Typed and strict.** TypeScript declarations, a JSON Schema for `koru.json`, error messages that say how to fix the input, an HTTPS-only app origin and a MessagePort bridge with no `addJavascriptInterface`.

Need iOS, Google Play bundles or a large plugin ecosystem? See the honest [comparison with Capacitor and Expo](https://koru.js.org/en/comparison/).

## Bring your frontend

Start with `koru create my-app --template vite`, or add koru to an existing project with `koru init` and set `webDir` to the folder containing the generated `index.html`. Enable `spa` if your app uses client-side routes. Keep your framework's hot reload with `koru dev --proxy http://localhost:5173`.

Koru packages static files. Apps that render on a server need a static export or an HTTPS API. [Frameworks guide](https://koru.js.org/en/existing-project/).

## Use Android features

The generated app includes a JavaScript module for device information, persistent storage, app information, vibration, sharing and local notifications:

```js
import { device, storage, platform } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("theme", "dark");
console.log(platform, info.model); // "android" in the APK, "web" under koru dev
```

`device`, `storage` and `app` are enabled by default; enable other plugins in `koru.json`. With a bundler, `koru js src/koru` writes the module and its types into your sources. See the [JavaScript API](https://koru.js.org/en/plugins/) for arguments, return values and error handling.

## Share your app

```sh
koru build --release --output app-release.apk
koru verify app-release.apk
```

A release build disables debugging and uses your existing signing identity. Install and test the release APK before sharing it (`koru run --release`). To update an app, keep its `appId` and key, and increase `versionCode`. [Distribution guide](https://koru.js.org/en/distribution/).

## Current support

Koru is in development. APK installation and execution have been tested on an **Android 11 (API 30), x86_64 emulator**, including local web content, offline use and the standard plugins. The configurable SDK range is 26–36; this does not establish compatibility with every device or API level.

Distribution currently uses APKs for direct installation. AAB output for Google Play and scheduled notifications are not available. Linux is the validated operational host; Windows and macOS have compile checks. See the [recorded Android results](docs/results/full/README.md) for the tested scope. `koru run` and `koru devices` are tested against an in-process implementation of the adb server protocol; they are not yet recorded on physical hardware.

## Learn more

- [Configuration](https://koru.js.org/en/configuration/): app identity, files, icons and permissions.
- [Develop with live reload](https://koru.js.org/en/develop/) and [run on a device](https://koru.js.org/en/run-on-device/).
- [CLI reference](https://koru.js.org/en/cli/): commands and options.
- [Troubleshooting](https://koru.js.org/en/troubleshooting/): build, installation and plugin errors.
- [Inside koru](https://koru.js.org/en/architecture-apk/): how the Go toolchain builds and runs an APK.

For work on koru itself, see the [test instructions](tests/README.md), [milestones](docs/milestones.md) and [documentation website guide](website/README.md). The Android formats and signing core use Go and its standard library; documentation tooling is separate.

## License

Original code is [MIT licensed](LICENSE). The AOSP JNI header includes its Apache-2.0 license.
