# Build your first Android app

Start with a folder containing `index.html` and its local assets. You can also use the static output of React, Vue, Svelte or another framework; build that output before running koru.

## Build the koru binary

Use a real Go compiler, version 1.23 or later, from the repository root:

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

Some recorded development sessions place Go in `/workspace/tooling/go/bin`. That path is environment-specific. Check `go version`; in this workspace `/usr/bin/go` is a game, not the compiler. Once compiled, koru needs neither Go nor network access to build applications.

## Initialize a project

```sh
./bin/koru init ./my-app
```

This creates `koru.json`, a `www/index.html` example and a persistent signing identity at `.koru/signing.pem`, with permissions `0600`. Keep that identity safe: future updates need the same certificate. `init` refuses to overwrite an existing configuration.

## Use your web assets

Put the files in `my-app/www`, or set `webDir` in `my-app/koru.json` to your framework's output directory. The directory must contain a regular `index.html`. Bundle application scripts locally. Do not place private keys in web assets.

[See all configuration fields]({{base}}en/configuration/)

## Build and verify

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
./bin/koru build ./my-app --release --output app.apk
./bin/koru verify app.apk
```

Write the APK outside `webDir`. The debug build allows WebView inspection; `--release` disables debug and DevTools and reuses the configured signing key. A missing or invalid key stops the build; koru never silently creates a replacement.

`verify` checks the APK signature and package structure. It does not execute the app or establish device compatibility.

## Run on Android

Install the APK on a device or emulator with Android 8/API 26 or later. ADB is an optional execution-testing tool, not a build dependency:

```sh
adb install -r app.apk
adb shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
```

Replace `com.example.koru` with your configured `appId`. The recorded execution matrix is Android 11/API 30/x86_64. Test your own app on the Android versions and devices you intend to support.

[Reproduce the Android tests]({{base}}en/testing/) · [Review the recorded evidence]({{base}}en/android-results/)
