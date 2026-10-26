# CLI commands

Run `koru --help` for the command list and `koru <command> --help` for a command’s flags and examples. Help is in English. Errors say what is wrong and how to fix it, and exit with a nonzero status. Set `NO_COLOR=1` for output without color.

Commands that take `[project]` default to the current directory, so inside a project you can simply run `koru dev`, `koru run` or `koru build`.

## Commands at a glance

| Command                      | What it does                                                                                                |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `create [directory]`         | Creates an app from a template with `koru.json`, `.gitignore`, types and a signing key. Asks in a terminal. |
| `dev [project]`              | Serves the app in a browser with live reload and web implementations of the JavaScript API.                 |
| `run [project]`              | Builds a debug APK, installs it on a connected device or emulator and opens it.                             |
| `devices`                    | Lists the devices and emulators the adb server sees, with their state.                                      |
| `build [project]`            | Builds and signs an APK. Debug by default; `--release` for distribution.                                    |
| `verify <app.apk>`           | Checks an APK’s signature and structure and prints a JSON report.                                           |
| `doctor [project]`           | Checks the configuration, key, web entry point, plugins and whether an adb server is reachable.             |
| `init [directory]`           | Adds a minimal `koru.json`, sample page and key to a directory, for existing projects.                      |
| `js [directory]`             | Copies the JavaScript module and types into your sources, for bundlers. Default `src/koru`.                 |
| `keygen --output <file.pem>` | Creates a new signing identity. Refuses to overwrite.                                                       |
| `dex build` / `dex inspect`  | Compiles or inspects standalone DEX files.                                                                  |
| `version`                    | Prints the version and host platform.                                                                       |

A compiled binary needs no Go, Java, Android SDK or network access to build apps. `dev`, `run` and `devices` are development helpers; they are not part of building an APK.

## `create`: start from a template

```sh
koru create my-app
koru create my-app --template vite --name "My App" --app-id com.example.myapp
```

| Flag               | Default                            | Meaning                                                    |
| ------------------ | ---------------------------------- | ---------------------------------------------------------- |
| `-t`, `--template` | `basic`                            | `basic` (HTML, CSS, JS, no build step) or `vite`.          |
| `--name`           | From the directory: `My App`       | Name shown under the launcher icon.                        |
| `--app-id`         | From the name: `com.example.myapp` | Android application id. Choose your own before publishing. |

In a terminal, missing values are asked for interactively; press Enter to accept the suggestion. In scripts and CI (no terminal), the suggestions are used directly and the directory is required. `create` refuses a directory that already has files. Afterwards it prints the next commands to run.

## `dev`: live preview

```sh
koru dev
koru dev --host 0.0.0.0
koru dev --proxy http://localhost:5173
```

| Flag      | Default     | Meaning                                                         |
| --------- | ----------- | --------------------------------------------------------------- |
| `--port`  | `8100`      | Port to listen on; `0` picks a free one.                        |
| `--host`  | `127.0.0.1` | `0.0.0.0` makes it reachable from other devices on the network. |
| `--proxy` | none        | Forward to your framework’s dev server and add the koru API.    |

Stop it with Ctrl+C. See [Develop with live reload]({{base}}en/develop/).

## `run` and `devices`

```sh
koru devices
koru run
koru run --device emulator-5554 --release
```

| Flag         | Default                                     | Meaning                                       |
| ------------ | ------------------------------------------- | --------------------------------------------- |
| `--device`   | The only ready device, or `$ANDROID_SERIAL` | Serial from `koru devices`.                   |
| `--release`  | off                                         | Install a release build (debugging disabled). |
| `--no-cache` | off                                         | Regenerate the runtime DEX.                   |

Both talk to the adb server at `127.0.0.1:5037` (or `ANDROID_ADB_SERVER_PORT` / `ADB_SERVER_SOCKET=tcp:host:port`). koru does not start it; run `adb start-server` or open Android Studio. See [Run on a device]({{base}}en/run-on-device/).

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

Writes `core.js`, `core.mjs` and `core.d.mts`. Import `./koru/core.mjs` from your code; see [Frameworks]({{base}}en/existing-project/).

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

Typical output is two lines:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

The byte count and full 64-character SHA-256 fingerprint vary; the ellipsis here shortens the sample. A cache hit means the runtime DEX was reused, not that the APK or its web assets were skipped. An invalid cache entry returns an error. The output must not replace the signing identity.

`--release` does **not** create a key. Both debug and release use the same configured key unless `--keystore` points to another existing identity. Replacing the key breaks Android updates signed by the old certificate.

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

The example is pretty-printed for reading; the CLI emits compact one-line JSON. `appId` is the manifest package; `entries` counts ZIP members; `bytes` is the APK length; `certificateSha256` identifies the signing certificate; `signatureScheme` names the verified APK signature. The sample values came from a test build and will differ for your app. Verification checks the APK's signature, ZIP, DEX, manifest, resources and icon. It does not install the app or prove ART/WebView execution. The APK input is limited to 320 MiB.

To extract a field in another program, parse JSON; keep the raw `verify` output intact. Human-readable build status goes to stdout, so do not concatenate it with `verify` when parsing.

## `keygen` and `version`

```sh
koru keygen --output ./my-app/alternate.pem
koru version
```

`keygen` creates a new PEM identity at the specified path and refuses to overwrite an existing file. It does not update `koru.json`. Set `signingKey` or pass `--keystore` explicitly if you intend to build with it; do not switch identities when updating an installed app.

## Standalone DEX commands

The DEX subcommands operate on the [DEX JSON intermediate representation]({{base}}en/dex-api/), separately from the Android APK builder.

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

[First-app walkthrough]({{base}}en/getting-started/) · [Troubleshooting]({{base}}en/troubleshooting/) · [DEX API]({{base}}en/dex-api/)
