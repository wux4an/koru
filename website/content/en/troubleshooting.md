# Troubleshooting

Start with `./bin/koru doctor ./my-app`. It validates the configuration, signing identity, web entry point and plugin manifests before a full build. All commands below assume the koru repository root and the project shown in [Getting started]({{base}}en/getting-started/).

## The compiler command is wrong

If `go version` launches a game or says command not found, inspect `command -v go` and put an actual Go 1.23+ compiler first on `PATH`. In the recorded development workspace, `/usr/bin/go` is a game. Go is needed to compile the CLI once; it is not needed when running the compiled CLI to make an APK.

## `doctor` or `build` rejects the project

| Symptom                                | What to check                                                                                                                                                                        |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `koru.json` cannot be read             | Pass the directory containing `koru.json`, such as `./my-app`. Use valid JSON with only documented fields.                                                                           |
| `index.html` is missing                | Build the web frontend first. Check that `webDir` points to a directory with a regular `index.html` at its root.                                                                     |
| Signing identity is missing or invalid | Restore the same PEM key and certificate named by `signingKey`. `build` does not silently replace it. For a brand-new app only, run `koru keygen --output my-app/.koru/signing.pem`. |
| APK output must be outside webDir      | Move `--output` outside the asset folder; for example `--output app.apk` from the repository root.                                                                                   |
| A plugin is unavailable                | Enable its ID in `plugins`. For custom plugins, also supply its manifest in `pluginManifests` and validate the manifest's declared ABI and handler.                                  |
| Asset collection fails                 | Remove symlinks, private keys and unsafe names from `webDir`. A file can be at most 64 MiB; the total at most 256 MiB and 2,000 files.                                               |
| Permission/configuration error         | Check Android permission names are complete (for example `android.permission.INTERNET`), IDs are unique, and `minSdk`/`targetSdk` stay in the supported 26–36 configuration range.   |

`doctor` checks the project entry point, but the full build also scans every asset. Fix the reported path and rerun the command; do not move a private key into `webDir` to make it readable.

## The web page is blank or a route fails

Open the generated `index.html` and inspect its script, stylesheet and image URLs. Place those files inside `webDir` and use local paths. Koru intercepts the app origin at `https://app.koru.invalid`; external scripts and iframes are blocked. If your frontend emits root-absolute `/assets/...` URLs, ensure those files exist under `webDir/assets/`. If it emits a `/project-name/` prefix, change the frontend's base/public URL for the APK build.

For client-side routes like `/settings`, set `"spa": true` in `koru.json`. Routes containing a file extension still return 404 when no matching asset exists; this keeps missing CSS and JavaScript from receiving HTML. Service workers must also be packaged as permitted local assets.

A normal desktop browser can preview the page and its pure web behavior, but `/__koru/core.mjs` and native methods work only inside the APK. Guard native calls in a browser preview or test them on Android.

## Installation or update fails

Confirm a device is available with `adb devices`. Build and inspect the exact APK you plan to install:

```sh
./bin/koru verify app.apk
adb install -r app.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Replace `com.example.hello` with your configured `appId`. If Android reports `INSTALL_FAILED_UPDATE_INCOMPATIBLE`, the installed app and APK have different package IDs or signing certificates. An update needs the original `appId` and `.koru/signing.pem`; changing `versionCode` alone cannot repair a key mismatch. For a newly distributed version, increment `versionCode`.

`verify` confirms APK structure and its v2 signature, not WebView rendering, framework resolution or plugin behavior. Install and exercise the app on the Android versions you support. The [recorded execution matrix]({{base}}en/android-results/) covers Android 11/API 30/x86_64.

## Native call fails

Check that the plugin is enabled and that its arguments match [the JavaScript API]({{base}}en/plugins/). Handle rejected Promises and inspect the error code (`METHOD_NOT_FOUND`, `PERMISSION_DENIED`, `TIMEOUT` or `CHANNEL_ERROR`, for example). A permission declared in the manifest can still require a user decision on a newer Android version. The API 33+ notification dialog has not been exercised in the recorded API 30 matrix.

Still stuck? Capture the exact command, exit status, error text, `koru.json` with private values removed, Android API/ABI and the result of `koru verify`. Never include `signing.pem` in a report.

[CLI reference]({{base}}en/cli/) · [Configuration]({{base}}en/configuration/) · [Security]({{base}}en/security/)
