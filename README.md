# koru

Koru is a self-contained Android toolchain that turns a static web project into a signed APK. The normal build is written in Go and the standard library: DEX, AXML, ARSC, aligned ZIP packaging, and APK v2 signing. It does not generate Java or Kotlin and does not use a JVM/JDK, Gradle, Android Studio, d8/dx, aapt, apksigner, zipalign, or wrapper frameworks.

**Koru has been installed and run on Android 11/ART.** WebView displayed local HTML, CSS, and images and ran JavaScript with Wi-Fi and mobile data disabled. The bridge, standard plugins, and optional JNI example have been tested on a real Android emulator. Evidence is available in [docs/results/full](docs/results/full), alongside the historical DEX results.

## Quick start

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru init ./my-app
./bin/koru build ./my-app --output app.apk
./bin/koru verify app.apk
./bin/koru build ./my-app --release --output app-release.apk
./bin/koru doctor ./my-app
```

The `koru` binary does not need Go or a network connection to build apps. Check that your shell has a real Go compiler before rebuilding Koru: `/usr/bin/go` is a game in this workspace. Use `command -v go` and `go version` to verify the selected compiler. Here, the real Go compiler is exposed by the Nix user profile at `/etc/profiles/per-user/wux4an/bin/go`. Some other managed workspaces provide it at `/workspace/tooling/go/bin`; check that path exists before using it.

```sh
command -v go
go version
# If needed, prepend the directory containing your verified Go compiler:
export PATH=/path/to/go/bin:$PATH
```

Rebuilding Koru itself also needs the Cobra module listed in `go.mod`. Go can fetch it during the first build, or use the local module cache for an offline rebuild. Cobra handles only command dispatch and help; the Android format and signing cores do not depend on it. The compiled `koru` binary can build apps offline.

`init` creates a project configuration, sample HTML, and a persistent signing identity at `.koru/signing.pem` with mode 0600. Keep that identity safe for future app updates. `build` never creates or changes a key when it is missing or invalid. `--release` reuses the configured key and disables debug mode and DevTools. Set `--keystore` to select another existing identity. `keygen --output signing.pem` explicitly creates a new identity and never overwrites an existing file.

## Existing projects and configuration

Koru accepts HTML, CSS, and JavaScript, including static output folders from React, Vue, Svelte, and other frameworks. Run those frameworks' compilers separately; Koru does not implement them. After `init`, set `webDir` to the static output directory:

```json
{
  "appId": "com.example.demo",
  "name": "Demo",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "./dist",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app", "vibration", "share", "notifications"],
  "signingKey": ".koru/signing.pem",
  "spa": true
}
```

Koru validates SDK levels (`26 ≤ minSdk ≤ targetSdk ≤ 36`), identifiers, versions, permissions, plugins, and paths. `icon` can point to a PNG. Plugins add only their declared permissions, including VIBRATE and POST_NOTIFICATIONS when needed. To use external HTTPS services, declare INTERNET and configure the server for the app origin. External scripts and iframes are blocked; application scripts are packaged locally.

The runtime uses the `https://app.koru.invalid` origin and DEX classes that intercept app assets. The normal runtime does not use `file://` or AndroidX. Service workers use the same allowlist from the Application context. SPA support is optional; unknown paths with file extensions return 404. Koru rejects symlinks, private keys inside assets, unsafe archive entries, and inputs over configured limits. The APK output must be outside `webDir`.

The runtime cache avoids regenerating DEX when only web content changes. Its fingerprint includes the toolchain, configuration, paths, plugins, and DEX SHA-256. Corruption returns an explicit error; `--no-cache` disables the cache.

## JavaScript API and plugins

```js
import { device, storage } from '/__koru/core.mjs';
console.log(await device.getInfo());
await storage.set('theme', 'dark');
```

The `packages/javascript` package is named `@koru/core`. You can also call `Koru.invoke("device.info", {})`. The bridge uses MessagePort, an explicit target origin, and the main frame; it does not use `addJavascriptInterface`. It provides request IDs, structured errors, promises, timeouts, validation, and cancellation before dispatch.

The six initial plugins and extensible ABI are described in [docs/plugins.md](docs/plugins.md). Custom plugins are supplied as DEX IR and manifests, without Java. JNI libraries are optional, prebuilt per ABI, and aligned to 16 KiB; an x86_64 example and its C source are included. Standard plugins do not require C or custom libraries.

## CLI

The Cobra-based CLI provides `init`, `build`, `verify`, `doctor`, `version`, and `keygen`. The DEX compiler can also be used independently:

```sh
./bin/koru dex build examples/dex/answer.json --output answer.dex
./bin/koru dex inspect answer.dex
./bin/koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` preserves the initial `file://` experiment for comparison; it is not used by APK builds. The Go API is in [pkg/dex](pkg/dex), with an [example](examples/dex/main.go) and [documentation](docs/dex-api.md).

## Documentation site

The bilingual Docsify site is developed separately from the Go toolchain. Install its documentation-only dependencies and start the local server with Bun:

```sh
cd website
bun install --frozen-lockfile
bun run dev
```

Open [http://localhost:4173/en/](http://localhost:4173/en/) for English or [http://localhost:4173/es/](http://localhost:4173/es/) for Spanish. To build for production, set `SITE_URL` to the real canonical site URL, including any optional subpath, then deploy `website/dist/` to static hosting:

```sh
SITE_URL=https://docs.example.com/koru/ bun run build
```

Replace the example URL with the actual domain and path. Bun and the website dependencies are for documentation development only; they are not needed to build the Go CLI or Android APKs.

## Tests

```sh
go test ./...
go vet ./...
go test -race ./...
node --test packages/javascript/core.test.cjs  # JavaScript client tests only
```

The build does not require Node. Fuzz targets cover encoding, DEX parsing, AXML, ARSC, and signing. External Android tests live under `tests` and are never part of the build: `dexdump`, `aapt`, `apksigner`, ADB, an emulator, and a test JRE were used only for comparison or execution. The delivered toolchain does not include those tools or a JDK.

Real WebView tests use `tests/android-smoke.py`, `android-navigation.py`, `android-worker.py`, and `android-keyboard.py`, with a debug APK installed and a DevTools port forwarded through ADB. Python and websockets are optional QA tools.

## Architecture and scope

DEX format, encoding, model, assembler, builder, and parser are separate modules. AXML, ARSC, APK packaging, and signing do not depend on Android tools. Config and app packages connect these modules; `runtime/dex` generates Android classes from Go; `runtime/native` is optional. See [ADR 0001: DEX core](docs/adr/0001-dex-core.md) and [ADR 0002: APK and runtime](docs/adr/0002-apk-runtime.md).

This is an initial functional release, with Android validation limited to an Android 11 x86_64 emulator. It does not establish compatibility with every device or API level, or acceptance in Google Play. Distribution is currently an APK for direct installation; Koru does not generate AABs. Notifications are immediate rather than scheduled. APK v3/v4 signing, multidex, general annotations, and full ART type verification are outside the initial subset. Running on ART complements structural validation by Koru's own parser. Windows and macOS have compile checks, while Linux is the validated operational host. There are no comparisons with Capacitor.

The [Addy Osmani agent skills](https://github.com/addyosmani/agent-skills) are installed as the Codex plugin `agent-skills@agent-skills` and were used during development. They are agent tools, not Koru dependencies. Original code is MIT licensed; the AOSP JNI header is Apache-2.0 and includes its license.

To create a reproducible delivery archive, run `python3 scripts/package.py --output koru-delivery.zip` after building `bin/koru` and generating the example APKs. The archive includes only versioned files, the executable, and two example APKs; it rejects keys and private directories.
