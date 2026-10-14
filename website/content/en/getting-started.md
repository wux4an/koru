# Build your first Android app

This guide starts with a clean checkout and ends with a signed APK running on Android. The commands assume you are at the repository root. Koru itself needs Go **1.23 or later** to compile once; the resulting CLI needs no Go installation, JDK, Android SDK or network to build an APK.

## 1. Get the CLI

```sh
git clone https://github.com/wux4an/koru.git
cd koru
go version
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

The last command prints the koru version and your host platform, for example:

```text
koru 0.2.0-dev (linux/amd64)
```

If `go` is missing, install the official Go toolchain first. In the recorded development workspace, `/usr/bin/go` is a game; check `command -v go` and `go version` before using it. The compiled `bin/koru` is the only build tool needed for normal APK builds. Cobra is fetched when compiling the CLI for the first time; it is not needed to run the binary.

## 2. Create a project

```sh
./bin/koru init ./my-app
```

Expected output:

```text
Initialized ./my-app; persistent signing identity: .koru/signing.pem
```

You now have:

```text
my-app/
├── koru.json             project settings
├── www/
│   └── index.html        runnable sample page
└── .koru/
    └── signing.pem       private key and certificate
```

`init` refuses to replace an existing `koru.json`. The key is created once with file mode `0600`. Back it up outside the web folder and version control: Android needs the same certificate for every update.

Open `my-app/www/index.html` in a browser to inspect the sample. Replace it with your own HTML, CSS, JavaScript, fonts and images before building. Keep all referenced files inside `webDir`; external scripts and iframes are not loaded by the packaged app.

## 3. Set the app identity before sharing it

`my-app/koru.json` starts with `appId: "com.example.koru"`, `name: "Koru"`, `version: "1.0.0"`, `versionCode: 1` and `webDir: "www"`. Change the identifier and display name before your first install. Android treats a different `appId` as a different app; later updates should retain the same identifier and signing key.

```json
{
  "appId": "com.example.hello",
  "name": "Hello Android",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "www",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app"],
  "signingKey": ".koru/signing.pem"
}
```

This is a complete configuration. `versionCode` must increase for each distributed update. The accepted SDK range is 26–36; the recorded device execution is Android 11/API 30, not a promise that every API level has been tested. [Every field and validation rule]({{base}}en/configuration/) is documented separately.

## 4. Check, build and verify

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
./bin/koru build ./my-app --release --output app.apk
./bin/koru verify app.apk
```

`doctor` checks config, key, web entry point and plugin manifests. A successful project check ends with a message like this:

```text
Project com.example.hello: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

The first `build` creates a debug APK, with WebView inspection available. The second creates a release APK and disables debug mode and DevTools. Both use the same configured signing identity; neither silently creates a new key. The output path is relative to your current shell directory, must end in `.apk` and must be outside `webDir`. Creating `app.apk` in the repository root satisfies that rule.

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

The numbers and fingerprint vary with your files and key. The second build may report `Runtime cache hit: true` because its runtime DEX can be reused; the final APK still reflects your current assets and mode.

`verify` prints a JSON report to stdout. Its fields are `appId`, `entries`, `bytes`, `certificateSha256` and `signatureScheme`. The signature and package structure passing verification are necessary, but they do not prove Android execution.

## 5. Install and launch on Android

For this step alone, use a device or emulator and Android's `adb`. It is a testing/install tool, not an APK build dependency. Android 8/API 26 is the configured minimum; test your own supported devices.

```sh
adb install -r app.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Use the exact `appId` from your `koru.json` before the slash. The runtime Activity is `com.koru.runtime.MainActivity`. Look for your page in WebView and exercise navigation, storage and any native methods your app calls. The [recorded Android matrix]({{base}}en/android-results/) covers API 30/x86_64; it does not validate your particular app on every device.

## 6. Update the app

Edit your web files, increment `versionCode` (and update `version` if you want a new visible label), then rebuild with the **same** `appId` and `signingKey`:

```sh
./bin/koru build ./my-app --release --output app-v2.apk
./bin/koru verify app-v2.apk
adb install -r app-v2.apk
```

Do not delete or replace `my-app/.koru/signing.pem`. If Android reports `INSTALL_FAILED_UPDATE_INCOMPATIBLE`, the installed app and new APK are signed by different certificates or use a different identity. Restore the original key instead of silently generating a new one.

## Using an existing web project

Build React, Vue, Svelte, Parcel or another tool to a **static folder** first. Point `webDir` at that folder and make sure it contains a regular `index.html` with local script, stylesheet and asset URLs. For a typical `dist/` next to `koru.json`:

```sh
npm --prefix ./my-app run build
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output app.apk
```

Set `"webDir": "dist"` in `my-app/koru.json` if `my-app/dist/index.html` is the generated entry. The web framework and its Node dependencies run before Koru; they are not embedded as build tools in Koru. Client-side routes can use `"spa": true`; unknown file paths still return 404. The [configuration guide]({{base}}en/configuration/) covers framework output, URL paths and network permissions.

If a command fails, the [troubleshooting guide]({{base}}en/troubleshooting/) gives the error, cause and next step. For flags and machine-readable outputs see the [CLI reference]({{base}}en/cli/).
