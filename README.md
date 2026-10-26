<p align="center">
  <img src="website/assets/favicon.svg" width="88" height="88" alt="koru spiral logo" />
</p>

<h1 align="center">koru</h1>

<p align="center"><strong>Your web app, on Android.</strong><br />Package HTML, CSS and JavaScript into a signed APK with koru.</p>

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

You need **Go 1.23 or later** to build koru from source. The first build may download Go dependencies.

```sh
git clone https://github.com/wux4an/koru.git
cd koru
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru init ./my-app
./bin/koru build ./my-app --output app.apk
```

Edit `my-app/www/index.html` and rebuild to include your changes. Your app settings are in `my-app/koru.json`.

To install and open the sample app on a connected Android device or emulator, use `adb`:

```sh
adb install -r app.apk
adb shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
```

`adb` is needed for this installation method, not for building APKs. See the [first app guide](https://koru.js.org/en/getting-started/) for the complete walkthrough.

Keep a secure backup of `my-app/.koru/signing.pem`. `koru init` adds `.koru/` to the project's `.gitignore` so the key stays out of version control. Android needs the same signing identity for future updates.

## Bring your frontend

Build your web project with its usual tools, then set `webDir` in `koru.json` to the folder containing the generated `index.html`. Enable `spa` if your app uses client-side routes.

Koru packages static files. Apps that render on a server need a static export or an HTTPS API. [Use an existing project](https://koru.js.org/en/existing-project/).

## Use Android features

The generated app includes a JavaScript module for device information, persistent storage, app information, vibration, sharing and local notifications:

```js
import { device, storage } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("theme", "dark");
console.log(info.model);
```

Run these calls inside the Android app. `device`, `storage` and `app` are enabled by default; enable other plugins in `koru.json`. See the [JavaScript API](https://koru.js.org/en/plugins/) for arguments, return values and error handling.

## Share your app

```sh
./bin/koru build ./my-app --release --output app-release.apk
./bin/koru verify app-release.apk
```

A release build disables debugging and uses your existing signing identity. Install and test the release APK before sharing it. To update an app, keep its `appId` and key, and increase `versionCode`. [Distribution guide](https://koru.js.org/en/distribution/).

## Current support

Koru is in development. APK installation and execution have been tested on an **Android 11 (API 30), x86_64 emulator**, including local web content, offline use and the standard plugins. The configurable SDK range is 26–36; this does not establish compatibility with every device or API level.

Distribution currently uses APKs for direct installation. AAB output for Google Play and scheduled notifications are not available. Linux is the validated operational host; Windows and macOS have compile checks. See the [recorded Android results](docs/results/full/README.md) for the tested scope.

## Learn more

- [Configuration](https://koru.js.org/en/configuration/): app identity, files, icons and permissions.
- [CLI reference](https://koru.js.org/en/cli/): commands and options.
- [Troubleshooting](https://koru.js.org/en/troubleshooting/): build, installation and plugin errors.
- [Inside koru](https://koru.js.org/en/architecture-apk/): how the Go toolchain builds and runs an APK.

For work on koru itself, see the [test instructions](tests/README.md), [milestones](docs/milestones.md) and [documentation website guide](website/README.md). The Android formats and signing core use Go and its standard library; documentation tooling is separate.

## License

Original code is [MIT licensed](LICENSE). The AOSP JNI header includes its Apache-2.0 license.
