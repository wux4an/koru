# CLI commands

Run `./bin/koru --help` for the installed version's command list and `./bin/koru build --help` for flags on a command. Examples on this page assume the binary is at `bin/koru` and the shell is at the repository root. If you put the executable on `PATH`, replace `./bin/koru` with `koru`.

The CLI and help are in English. Interactive terminals use the documentation palette; redirected output and `NO_COLOR` have no ANSI escapes. Errors go to stderr and return a nonzero exit status. The two inspection commands produce JSON on stdout so they can be parsed.

## Project commands at a glance

| Command                      | Input                                    | Result                                                                                                  |
| ---------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `init [directory]`           | Optional project directory; default `.`. | Creates `koru.json`, `www/index.html` and `.koru/signing.pem`. Refuses to replace an existing config.   |
| `doctor [project]`           | Optional project directory; default `.`. | Validates config, signing key, web entry point and plugin manifests.                                    |
| `build [project]`            | Optional project directory; default `.`. | Builds and signs an APK. Debug by default; output defaults to `app.apk` in the current shell directory. |
| `verify <app.apk>`           | One APK path.                            | Validates the signature and package structure and prints JSON.                                          |
| `keygen --output <file.pem>` | New destination file.                    | Explicitly creates a signing identity; refuses to overwrite. `-o` is an alias for `--output`.           |
| `version`                    | No arguments.                            | Prints version and host platform.                                                                       |

All commands accept `--help` or `-h`. A compiled binary builds apps without Go, Java, Android SDK build tools or network access.

## `init`: files you can edit

```sh
./bin/koru init ./my-app
```

```text
Initialized ./my-app; persistent signing identity: .koru/signing.pem
```

`my-app/koru.json` contains app identity, SDK values, `webDir`, plugins and signing path. The sample page is `my-app/www/index.html`. The key is private; keep `my-app/.koru/signing.pem` out of version control and the web folder. `init` uses an existing directory if needed but refuses to overwrite existing project configuration.

## `doctor`: check before packaging

```sh
./bin/koru doctor ./my-app
```

```text
Build and verification use the self-contained Go implementation; no Java, SDK, or external build tools are required.
Project com.example.koru: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

The `appId` comes from your config. Running `doctor` without a project in a directory lacking `koru.json` prints a general environment note; give the project path to check an app. `doctor` does not build, install or execute an APK.

## `build`: debug, release and destinations

```sh
./bin/koru build ./my-app --output app-debug.apk
./bin/koru build ./my-app --release --output app.apk
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
./bin/koru verify app.apk
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
./bin/koru keygen --output ./my-app/alternate.pem
./bin/koru version
```

`keygen` creates a new PEM identity at the specified path and refuses to overwrite an existing file. It does not update `koru.json`. Set `signingKey` or pass `--keystore` explicitly if you intend to build with it; do not switch identities when updating an installed app.

## Standalone DEX commands

The DEX subcommands operate on the [DEX JSON intermediate representation]({{base}}en/dex-api/), separately from the Android APK builder.

```sh
./bin/koru dex build examples/dex/answer.json --output answer.dex
./bin/koru dex inspect answer.dex
```

```text
Wrote answer.dex (476 bytes); DEX 035 verified successfully.
```

`dex build` requires one IR JSON file. Its `--output` defaults to `classes.dex`, and the builder validates the produced DEX with the independent parser. `dex inspect` takes one file and prints an indented JSON object containing size, strings, types, prototypes, fields, methods and classes. The DEX input limit is 64 MiB. Structural validation is not full ART type verification.

```sh
./bin/koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` takes no IR file. It produces the historical `file://` Activity experiment for comparison only; normal APK builds use intercepted HTTPS.

[First-app walkthrough]({{base}}en/getting-started/) · [Troubleshooting]({{base}}en/troubleshooting/) · [DEX API]({{base}}en/dex-api/)
