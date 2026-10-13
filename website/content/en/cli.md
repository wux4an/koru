# CLI commands

Run `koru --help` or `koru <command> --help` for the installed version's flags. Commands and help are in English. The compiled binary builds and verifies packages without Android build tools.

## Project commands

| Command                               | What it does                                                                                         |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `koru init [directory]`               | Creates configuration, example web assets and a persistent signing identity. Default directory: `.`. |
| `koru build [project]`                | Builds a debug APK. Default project: `.`; default output: `app.apk`.                                 |
| `koru verify <app.apk>`               | Checks APK v2 signature and package structure; emits a JSON report.                                  |
| `koru doctor [project]`               | Checks config, existing signing identity, web index and custom manifests.                            |
| `koru keygen --output <identity.pem>` | Explicitly creates a new signing identity without overwriting an existing file.                      |
| `koru version`                        | Prints the toolchain version.                                                                        |

## Build options

```sh
koru build ./my-app --release --output app-release.apk
koru build ./my-app --keystore .koru/other-existing.pem --no-cache
```

| Flag         | Meaning                                                            |
| ------------ | ------------------------------------------------------------------ |
| `--output`   | Signed APK destination, with `.apk` extension, outside `webDir`.   |
| `--release`  | Disables debug and WebView DevTools; reuses the signing identity.  |
| `--keystore` | Existing PEM identity; relative paths resolve against the project. |
| `--no-cache` | Disables runtime DEX caching.                                      |

Keep the key and certificate for future updates. A release build does not generate a new identity automatically.

## Standalone DEX commands

```sh
koru dex build examples/dex/answer.json --output answer.dex
koru dex inspect answer.dex
koru dex build --activity-experiment --output experiment.dex
```

`dex build` accepts the JSON form of the DEX IR. `dex inspect` emits a decoded JSON inspection after validation. `--activity-experiment` retains the original `file://` Activity experiment for comparison. The normal APK build uses intercepted HTTPS and does not use this experiment.

The DEX input limit is 64 MiB. Structural validation is not ART execution or full type verification.

[DEX API]({{base}}en/dex-api/) · [First build]({{base}}en/getting-started/)
