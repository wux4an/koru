# Troubleshooting

Start with `koru doctor` in your project directory. It validates the configuration, signing identity, web entry point and plugin manifests before a full build, and reports whether an adb server and devices are available. Every error message names the field or file involved and how to fix it.

## `koru` or Go is not found

If `go version` returns “command not found”, install Go 1.23 or later. You only need Go to build the koru executable. If `koru` is not found after `make install`, add `~/.local/bin` to your `PATH`.

## `doctor` or `build` rejects the project

| Symptom                                | What to check                                                                                                                                                                        |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `no koru.json in …`                    | Run the command inside the project, or pass its directory (`koru build ./my-app`). Create one with `koru create` or `koru init`.                                                     |
| `koru.json: unknown field …`           | Fix the field name; the `$schema` line gives your editor completion for every valid field.                                                                                           |
| `webDir … has no index.html`           | Build the web frontend first (for example `npm run build`). Check that `webDir` points to the folder with `index.html` at its root.                                                  |
| Signing identity is missing or invalid | Restore the same PEM key and certificate named by `signingKey`. `build` does not silently replace it. For a brand-new app only, run `koru keygen --output my-app/.koru/signing.pem`. |
| APK output must be outside webDir      | Move `--output` outside the asset folder; for example `--output app.apk` from the repository root.                                                                                   |
| A plugin is unavailable                | Enable its ID in `plugins`. For custom plugins, also supply its manifest in `pluginManifests` and validate the manifest's declared ABI and handler.                                  |
| Asset collection fails                 | Remove symlinks, private keys and unsafe names from `webDir`. A file can be at most 64 MiB; the total at most 256 MiB and 2,000 files.                                               |
| Permission/configuration error         | Check Android permission names are complete (for example `android.permission.INTERNET`), IDs are unique, and `minSdk`/`targetSdk` stay in the supported 26–36 configuration range.   |

`doctor` checks the project entry point, but the full build also scans every asset. Fix the reported path and rerun the command; do not move a private key into `webDir` to make it readable.

## The web page is blank or a route fails

Open the generated `index.html` and inspect its script, stylesheet and image URLs. Place those files inside `webDir` and use local paths. Koru intercepts the app origin at `https://app.koru.invalid`; external scripts and iframes are blocked. If your frontend emits root-absolute `/assets/...` URLs, ensure those files exist under `webDir/assets/`. If it emits a `/project-name/` prefix, change the frontend's base/public URL for the APK build.

For client-side routes like `/settings`, set `"spa": true` in `koru.json`. Routes containing a file extension still return 404 when no matching asset exists; this keeps missing CSS and JavaScript from receiving HTML. Service workers must also be packaged as permitted local assets.

In a browser, preview with `koru dev` rather than opening the file directly: it serves `/__koru/core.mjs` with web implementations of the API. A plain static server or `file://` cannot load it. Native behavior (permissions, Android share sheet, notifications) still needs a test on Android with `koru run`.

## `koru dev` problems

| Symptom                             | What to check                                                                                                             |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `cannot listen on …`                | Another program uses the port. Run `koru dev --port 3000`, or `--port 0` for any free port.                               |
| The page does not reload            | Save inside the project directory; `node_modules`, `.git` and `.koru` are ignored. Check the terminal for `Watch` errors. |
| `METHOD_NOT_FOUND … is not enabled` | Add the plugin to `plugins` in `koru.json`; the next reload picks it up.                                                  |
| `--proxy` shows `cannot reach …`    | Start your framework’s dev server first and use its exact URL, for example `http://localhost:5173`.                       |
| A phone cannot open the Network URL | Use `--host 0.0.0.0`, keep both devices on the same network and allow the port in your firewall.                          |

## Installation or update fails

Run `koru devices`. `unauthorized` means you must accept the USB debugging prompt on the phone; `offline` means reconnecting it. If koru reports that no adb server is running, start one with `adb start-server` or open Android Studio. `koru run` explains the common install errors; see [Run on a device]({{base}}docs/en/run-on-device/). To install by hand, build and inspect the exact APK first:

```sh
koru verify app.apk
adb install -r app.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Replace `com.example.hello` with your configured `appId`. If Android reports `INSTALL_FAILED_UPDATE_INCOMPATIBLE`, the installed app and APK have different package IDs or signing certificates. An update needs the original `appId` and `.koru/signing.pem`; changing `versionCode` alone cannot repair a key mismatch. For a newly distributed version, increment `versionCode`.

`verify` confirms APK structure and its v2 signature, not WebView rendering, framework resolution or plugin behavior. Install and exercise the app on the Android versions you support. The [recorded execution matrix]({{base}}docs/en/android-results/) covers Android 11/API 30/x86_64.

## Native call fails

Check that the plugin is enabled and that its arguments match [the JavaScript API]({{base}}docs/en/plugins/). Handle rejected Promises and inspect the error code (`METHOD_NOT_FOUND`, `PERMISSION_DENIED`, `TIMEOUT` or `CHANNEL_ERROR`, for example). A permission declared in the manifest can still require a user decision on a newer Android version. The API 33+ notification dialog has not been exercised in the recorded API 30 matrix.

Still stuck? Capture the exact command, exit status, error text, `koru.json` with private values removed, Android API/ABI and the result of `koru verify`. Never include `signing.pem` in a report.

[CLI reference]({{base}}docs/en/cli/) · [Configuration]({{base}}docs/en/configuration/) · [Security]({{base}}docs/en/security/)
