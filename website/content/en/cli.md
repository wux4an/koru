# CLI commands

Run `koru --help` for the command list and `koru <command> --help` for a command’s flags and examples. Help is in English. Errors say what is wrong and how to fix it, and exit with a nonzero status. Set `NO_COLOR=1` for output without color.

Commands that take `[project]` default to the current directory, so inside a project you can simply run `koru dev`, `koru run` or `koru build`.

## Commands at a glance

| Command                      | What it does                                                                                                                                      |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `create [directory]`         | Creates an app from a template (`basic`, `vite`, `react`, `vue`, `svelte`, `vanilla-ts`) with `koru.json`, `.gitignore`, types and a signing key. |
| `dev [project]`              | Serves the app in a browser with live reload and web implementations of the JavaScript API.                                                       |
| `run [project]`              | Builds a debug APK, installs it on a connected device or emulator and opens it. `--logs` keeps streaming logs.                                    |
| `logs [project]`             | Streams the app’s logcat output and WebView console messages from a device or emulator.                                                           |
| `devices`                    | Lists the devices and emulators the adb server sees, with their state.                                                                            |
| `build [project]`            | Builds and signs an APK. Debug by default; `--release` for distribution; `--bump` raises the version first.                                       |
| `bump [level] [project]`     | Raises `versionCode` and the version in `koru.json`: `patch`, `minor`, `major` or `code`.                                                         |
| `verify <app.apk>`           | Checks an APK’s signature and structure and prints a JSON report.                                                                                 |
| `doctor [project]`           | Checks the configuration, key, web entry point and plugins, warns about common mistakes, and reports whether an adb server is reachable.          |
| `plugin list\|add\|remove`   | Shows, enables or disables the built-in plugins in `koru.json`.                                                                                   |
| `config [project]`           | Prints the resolved configuration as JSON, including the permissions it adds.                                                                     |
| `deeplinks [project]`        | Prints the `assetlinks.json` that verifies your `https` deep links.                                                                               |
| `init [directory]`           | Adds a minimal `koru.json`, sample page and key to a directory, for existing projects.                                                            |
| `js [directory]`             | Copies the JavaScript module and types into your sources, for bundlers. Default `src/koru`.                                                       |
| `keygen --output <file.pem>` | Creates a new signing identity. Refuses to overwrite.                                                                                             |
| `dex build` / `dex inspect`  | Compiles or inspects standalone DEX files.                                                                                                        |
| `version`                    | Prints the version and host platform.                                                                                                             |

A compiled binary needs no Go, Java, Android SDK or network access to build apps. `dev`, `run` and `devices` are development helpers; they are not part of building an APK.

## `create`: start from a template

```sh
koru create my-app
koru create my-app --template react --name "My App" --app-id com.example.myapp
```

| Flag               | Default                            | Meaning                                                    |
| ------------------ | ---------------------------------- | ---------------------------------------------------------- |
| `-t`, `--template` | `basic`                            | Starter template; see the table below.                     |
| `--name`           | From the directory: `My App`       | Name shown under the launcher icon.                        |
| `--app-id`         | From the name: `com.example.myapp` | Android application id. Choose your own before publishing. |

| Template     | What you get                                               | `webDir` |
| ------------ | ---------------------------------------------------------- | -------- |
| `basic`      | HTML, CSS and JavaScript. No build step or Node.js needed. | `www`    |
| `vite`       | A Vite project with hot reload.                            | `dist`   |
| `react`      | Vite with React 18 and JSX.                                | `dist`   |
| `vue`        | Vite with Vue 3 single-file components.                    | `dist`   |
| `svelte`     | Vite with Svelte 5.                                        | `dist`   |
| `vanilla-ts` | Vite with strict TypeScript and no framework.              | `dist`   |

Every template other than `basic` needs Node.js. Run `npm install` once, then use the scripts `create` writes: `npm run dev` (hot reload in the browser), `npm run android` (build, then `koru run`) and `npm run apk` (signed release APK). Each includes a copy of the koru JavaScript module in `src/koru/`; refresh it with `koru js src/koru --force` after upgrading koru. The framework templates were built with `npm run build`, `koru build` and `koru verify` (and `tsc --noEmit` for `vanilla-ts`); they have not been run on Android.

In a terminal, missing values are asked for interactively; press Enter to accept the suggestion. In scripts and CI (no terminal), the suggestions are used directly and the directory is required. `create` refuses a directory that already has files. Afterwards it prints the next commands to run. For `react`:

```text
Created my-app (com.example.myapp, react template)

Next steps:
  cd my-app
  npm install
  npm run dev        # browser preview with hot reload
  npm run android    # install and open on a device

Back up my-app/.koru/signing.pem: Android only accepts updates signed with the same key.
```

## `dev`: live preview

```sh
koru dev
koru dev --host 0.0.0.0
koru dev --host 0.0.0.0 --qr
koru dev --proxy http://localhost:5173
```

| Flag      | Default     | Meaning                                                                                          |
| --------- | ----------- | ------------------------------------------------------------------------------------------------ |
| `--port`  | `8100`      | Port to listen on; `0` picks a free one.                                                         |
| `--host`  | `127.0.0.1` | `0.0.0.0` makes it reachable from other devices on the network.                                  |
| `--proxy` | none        | Forward to your framework’s dev server and add the koru API.                                     |
| `--qr`    | off         | Print a QR code of the network URL. Shown by itself in a terminal when the host is not loopback. |

Stop it with Ctrl+C. See [Develop with live reload]({{base}}docs/en/develop/).

## `run` and `devices`

```sh
koru devices
koru run
koru run --device emulator-5554 --release
koru run --logs
```

| Flag         | Default                                     | Meaning                                                  |
| ------------ | ------------------------------------------- | -------------------------------------------------------- |
| `--device`   | The only ready device, or `$ANDROID_SERIAL` | Serial from `koru devices`.                              |
| `--release`  | off                                         | Install a release build (debugging disabled).            |
| `--logs`     | off                                         | After opening the app, stream its logs like `koru logs`. |
| `--no-cache` | off                                         | Regenerate the runtime DEX.                              |

`run`, `logs` and `devices` talk to the adb server at `127.0.0.1:5037` (or `ANDROID_ADB_SERVER_PORT` / `ADB_SERVER_SOCKET=tcp:host:port`). koru does not start it; run `adb start-server` or open Android Studio. See [Run on a device]({{base}}docs/en/run-on-device/).

## `logs`: the app’s logs in your terminal

```sh
koru logs
koru logs ./my-app --device emulator-5554
```

Streams what the running app logs: its logcat output, plus the `chromium` lines that carry `console.log`, `console.warn` and `console.error` from the WebView. The app must already be open (see `koru run`). Press Ctrl+C to stop. `logs` takes the same `--device` flag and `$ANDROID_SERIAL` as `run`. If the app is not running, it fails with “is not running on …: open it with `koru run` first”.

koru does this through the adb server’s network protocol, running `pidof` and `logcat` on the device; it does not run `adb` or `logcat` on your computer. This is **not validated on a device yet**: the commands and the line format follow Android’s documentation and are tested only against a simulated adb server. If the output looks wrong on your device, use `chrome://inspect` instead.

## `init`: add koru to a directory

```sh
cd your-web-app
koru init
```

```text
Initialized .; persistent signing identity: .koru/signing.pem
```

Writes `koru.json`, a sample `www/index.html` and `.koru/signing.pem`, and adds `.koru/` to `.gitignore` (creating it if needed, appending otherwise). It refuses to replace an existing `koru.json`. Edit `webDir` to point at your build output. For a brand-new app, `create` gives you a richer starting point.

## `js`: the module for bundlers

```sh
koru js src/koru
koru js src/koru --force   # after upgrading koru
```

Writes `core.js`, `core.mjs` and `core.d.mts`. Import `./koru/core.mjs` from your code; see [Frameworks]({{base}}docs/en/existing-project/).

## `doctor`: check before packaging

```sh
koru doctor ./my-app
```

```text
Build and verification use the self-contained Go implementation; no Java, SDK, or external build tools are required.
Devices 1 ready of 1 connected (adb server 127.0.0.1:5037); see koru devices
Project com.example.koru: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

The `appId` comes from your config. `doctor` also reports whether an adb server is reachable and how many devices are ready; that line is informational, since only `koru run` needs a device. Without a project in the current directory it prints the environment notes only. `doctor` does not build, install or execute an APK.

When it finds something likely to bite you, it adds `Warning` lines. Warnings never make the check fail; the invalid configuration still does. This run is from a project with several problems, and the line about the adb server depends on your machine:

```text
Project com.example.myapp: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
Warning .koru/ is not listed in .gitignore; it holds your signing key. Add a line ".koru/" to .gitignore and keep a private backup of the key.
Warning index.html has no <meta name="viewport" content="width=device-width,initial-scale=1">; the page renders at desktop width on phones. Add it inside <head>.
Warning api.js calls fetch() with an https:// URL but android.permission.INTERNET is not in permissions (heuristic text match, not a full analysis). Add it to "permissions" in koru.json if the app needs the network.
Warning icon icon.png is 64x48; use a square PNG or Android will distort it.
Warning icon icon.png is 64x48; use at least 192x192 px so it stays sharp on high-density screens.
Note recorded Android execution covers API 30 (targetSdk here is 36); test on your devices before shipping.
```

| Warning      | Triggered when                                                                                                                                                   |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.gitignore` | The project has a `.git` folder or a `.gitignore` that does not list `.koru/`, where the key lives.                                                              |
| viewport     | `index.html` has no `<meta name="viewport">`.                                                                                                                    |
| `INTERNET`   | A script or HTML file calls `fetch()` with a literal `https://` URL and the permission is missing. This is a text match, so it can miss calls or flag dead code. |
| icon         | `icon` is not square, or is smaller than 192 × 192 px. An icon that is missing or not a PNG is an error.                                                         |

The final `Note` appears when `targetSdk` is not 30, which includes new projects (the default is 36): the recorded runs cover Android 11 (API 30) only, so test on the devices you support. See [Tested compatibility]({{base}}docs/en/android-results/).

## `build`: debug, release and destinations

```sh
koru build ./my-app --output app-debug.apk
koru build ./my-app --release --output app.apk
```

| Flag         | Default                  | Meaning                                                                                                      |
| ------------ | ------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `--output`   | `app.apk`                | Signed APK destination, relative to the current shell directory. Must end in `.apk` and be outside `webDir`. |
| `--release`  | off                      | Disables debug mode and WebView DevTools. Uses the configured identity.                                      |
| `--keystore` | `signingKey` from config | Existing PEM identity to use instead; relative paths resolve against the project directory.                  |
| `--no-cache` | off                      | Regenerates runtime DEX even when a valid cache entry exists.                                                |
| `--bump`     | off                      | Raises `versionCode` and the patch version in `koru.json` before building. See `bump` below.                 |

Typical output is two lines:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

The byte count and full 64-character SHA-256 fingerprint vary; the ellipsis here shortens the sample. A cache hit means the runtime DEX was reused, not that the APK or its web assets were skipped. An invalid cache entry returns an error. The output must not replace the signing identity.

`--release` does **not** create a key. Both debug and release use the same configured key unless `--keystore` points to another existing identity. Replacing the key breaks Android updates signed by the old certificate.

## `bump`: raise the version

```sh
koru bump              # patch: 1.0.0 -> 1.0.1
koru bump minor        # 1.0.1 -> 1.1.0
koru bump major ./my-app
koru bump code         # only versionCode
koru build --release --bump --output app-release.apk
```

```text
Bumped version 1.0.0 -> 1.0.1, versionCode 1 -> 2
Bumped version 1.0.1 -> 1.1.0, versionCode 2 -> 3
Bumped version 1.1.0 -> 1.1.0, versionCode 3 -> 4
Bumped version 1.1.0 -> 2.0.0, versionCode 4 -> 5
```

`versionCode` always goes up by one, which is what Android compares when you install an update. The level (default `patch`) also changes the `X.Y.Z` in `version`; `code` leaves `version` alone. If `version` is not three dot-separated numbers, such as `1.0`, only `versionCode` is raised and a warning asks you to edit `version` by hand. `bump` refuses to go past Android’s limit of 2,100,000,000.

`koru build --bump` does a patch bump and then builds, so the APK carries the new numbers; the build prints the same `Bumped …` line first. `bump` rewrites `koru.json` with two-space indentation in koru’s field order, so reformat it by hand afterwards if you care about its layout. See [Distribute your app]({{base}}docs/en/distribution/).

## `plugin` and `config`: edit and inspect `koru.json`

```sh
koru plugin list
koru plugin add vibration notifications
koru plugin remove share --project ./my-app
koru config
```

`plugin list` shows the six built-in plugins, whether each is on, and the Android permission it adds. This is the list for a project created from `basic`:

```text
  on   device         -
  on   storage        -
  on   vibration      android.permission.VIBRATE
  on   app            -
  on   share          -
  off  notifications  android.permission.POST_NOTIFICATIONS
```

`plugin add <id>...` and `plugin remove <id>...` edit `plugins` in `koru.json` and print `Enabled …` or `Disabled …`. `list` takes the project as an argument; `add` and `remove` take it with `--project` (default `.`). An unknown ID is an error that lists the valid ones. Custom DEX plugins still go in `pluginManifests`; see [JavaScript and plugins]({{base}}docs/en/plugins/).

`config` prints what koru will use, as JSON: the `config` as read from `koru.json`, `effectivePermissions` (the ones your plugins add, such as `android.permission.VIBRATE`), the absolute `webRoot` and the absolute `signingKey` path. It validates first, so it doubles as a quick check of your edits:

```text
koru: koru.json: invalid orientation "upside": use "portrait", "landscape" or "sensor", or remove the field to keep Android's default
```

## `deeplinks`: verify your `https` links

```sh
koru deeplinks
koru deeplinks ./my-app --json > assetlinks.json
```

Prints the Digital Asset Links file (`assetlinks.json`) for the `https` entries of `deepLinks` in `koru.json`. With `--json` only the JSON is printed, ready to save; `--keystore` uses another existing PEM identity instead of `signingKey`. Without `https` entries it fails and says so. The full flow is in [Configure your project]({{base}}docs/en/configuration/).

## `verify`: structured report

```sh
koru verify app.apk
```

```json
{
  "appId": "com.example.koru",
  "entries": 7,
  "bytes": 16785,
  "certificateSha256": "3ecd2de568b127fef17621e1cce54a68a0587c4b3de22ebcd06224bf5f052122",
  "signatureScheme": "APK v2 RSA/SHA-256"
}
```

When `koru.json` has `deepLinks`, the report also lists the ones found in the manifest, as `"deepLinks": ["https://example.com/app", "myapp://"]`. The example is pretty-printed for reading; the CLI emits compact one-line JSON. `appId` is the manifest package; `entries` counts ZIP members; `bytes` is the APK length; `certificateSha256` identifies the signing certificate; `signatureScheme` names the verified APK signature. The sample values came from a test build and will differ for your app. Verification checks the APK's signature, ZIP, DEX, manifest, resources and icon. It does not install the app or prove ART/WebView execution. The APK input is limited to 320 MiB.

To extract a field in another program, parse JSON; keep the raw `verify` output intact. Human-readable build status goes to stdout, so do not concatenate it with `verify` when parsing.

## `keygen` and `version`

```sh
koru keygen --output ./my-app/alternate.pem
koru version
```

`keygen` creates a new PEM identity at the specified path and refuses to overwrite an existing file. It does not update `koru.json`. Set `signingKey` or pass `--keystore` explicitly if you intend to build with it; do not switch identities when updating an installed app.

## Standalone DEX commands

The DEX subcommands operate on the [DEX JSON intermediate representation]({{base}}docs/en/dex-api/), separately from the Android APK builder.

```sh
koru dex build examples/dex/answer.json --output answer.dex
koru dex inspect answer.dex
```

```text
Wrote answer.dex (476 bytes); DEX 035 verified successfully.
```

`dex build` requires one IR JSON file. Its `--output` defaults to `classes.dex`, and the builder validates the produced DEX with the independent parser. `dex inspect` takes one file and prints an indented JSON object containing size, strings, types, prototypes, fields, methods and classes. The DEX input limit is 64 MiB. Structural validation is not full ART type verification.

```sh
koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` takes no IR file. It produces the historical `file://` Activity experiment for comparison only; normal APK builds use intercepted HTTPS.

[First-app walkthrough]({{base}}docs/en/getting-started/) · [Troubleshooting]({{base}}docs/en/troubleshooting/) · [DEX API]({{base}}docs/en/dex-api/)
