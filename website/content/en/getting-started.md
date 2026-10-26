# Create your first app

In about five minutes you will create an app from a template, preview it in your browser with live reload and open it on an Android phone or emulator. You need **Go 1.23 or later** to install koru. To run the app on Android you also need a phone with USB debugging, or an emulator.

## 1. Install koru

```sh
git clone https://github.com/wux4an/koru.git
cd koru
make install
koru version
```

`make install` builds a self-contained binary and copies it to `~/.local/bin`. If `koru version` says “command not found”, add that folder to your `PATH` (for example `export PATH="$HOME/.local/bin:$PATH"` in your shell profile). Without `make`, run `CGO_ENABLED=0 go build -trimpath -o ~/.local/bin/koru ./cmd/koru`.

The binary is all you need from now on: it builds APKs without Go, Java, Gradle, the Android SDK or an Internet connection.

## 2. Create a project

```sh
koru create my-app
```

In a terminal, koru asks three questions. Press Enter to accept each suggestion:

```text
? App name: (My App)
? Application id: (com.example.myapp)
  1. basic  HTML, CSS and JavaScript, no build step or Node.js needed
  2. vite   Vite project with hot reload (needs Node.js)
? Template: (basic)
```

You can answer them with flags instead, which is handy in scripts:

```sh
koru create my-app --name "My App" --app-id com.example.myapp --template basic
```

The `basic` template contains:

```text
my-app/
├── koru.json          app name, id, version, plugins
├── www/               the files packaged into the APK
│   ├── index.html
│   ├── app.js
│   └── style.css
├── types/koru.d.mts   types for the JavaScript API
├── jsconfig.json      lets your editor use those types
├── .gitignore         keeps .koru/ out of Git
└── .koru/signing.pem  your private signing key
```

> **Back up `.koru/signing.pem`.** Android only installs updates signed with the same key. The generated `.gitignore` keeps it out of version control.

## 3. Preview with live reload

```sh
cd my-app
koru dev
```

Open the address it prints, `http://127.0.0.1:8100`. Edit `www/index.html` or `www/app.js` and save: the page reloads by itself. The JavaScript API works here too, on web implementations: the counter is saved with `storage`, and **Share** uses your browser’s share sheet. The badge shows `web`; inside the APK it shows `android`. See [Develop with live reload]({{base}}en/develop/) for testing on a phone’s browser or using another dev server.

## 4. Run it on Android

Connect a phone with [USB debugging](https://developer.android.com/studio/debug/dev-options) enabled, or start an emulator from Android Studio’s Device Manager. Then:

```sh
koru run
```

```text
Built com.example.myapp 1.0.0 (20 KB) in 35ms
Installed on sdk gphone64 x86 64 (emulator-5554)
Opened My App on sdk gphone64 x86 64 (emulator-5554); inspect it at chrome://inspect
```

`koru run` builds a debug APK, installs it and opens it. It talks to the adb server that Android Studio and the platform tools run; if none is running, the error tells you how to start it. While the app is open, visit `chrome://inspect` in Chrome on your computer to see its console and DOM. More in [Run on a device]({{base}}en/run-on-device/).

Change something, then run `koru run` again. That is the whole loop: **edit → `koru dev` → `koru run`**.

## 5. Build a release

```sh
koru build --release --output my-app.apk
koru verify my-app.apk
```

A release build turns off WebView debugging and signs with your existing key. To publish an update later, keep the same `appId` and key and increase `versionCode` in `koru.json`. See [Distribute your app]({{base}}en/distribution/).

> koru is experimental. Recorded execution covers Android 11 (API 30) on an x86_64 emulator; the configurable range is API 26–36. Test on the devices you want to support.

## Next steps

- [Use React, Vue, Svelte or another framework]({{base}}en/existing-project/), or start with `koru create my-app --template vite`.
- [Call Android features from JavaScript]({{base}}en/plugins/), with TypeScript types.
- [See every command]({{base}}en/cli/) and [every `koru.json` option]({{base}}en/configuration/).

If a step fails, run `koru doctor` and see [Troubleshooting]({{base}}en/troubleshooting/).
