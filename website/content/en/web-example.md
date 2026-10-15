# Run the web example

Try a small HTML, CSS and JavaScript app before connecting your project. The example includes a counter, an SVG image and a second page. You do not need to install web dependencies.

## Make the APK

From the repository root, after [compiling koru]({{base}}en/getting-started/):

```sh
./bin/koru init ./my-app
cp -a examples/web/. my-app/www/
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output example.apk
./bin/koru verify example.apk
```

`cp -a examples/web/. my-app/www/` copies all files over the initial sample page. The dot matters: it copies the contents of the directory, including `index.html`, rather than making `www/web/`. On systems without GNU `cp`, copy each file into `my-app/www/`.

A successful build finishes with output shaped like this; bytes and certificate differ on your machine:

```text
Runtime cache hit: false
Wrote example.apk (… bytes), APK v2 RSA/SHA-256, certificate …
```

Before distributing the app, change `appId` and `name` in `my-app/koru.json`. Keep the generated `.koru/signing.pem` for updates.

## What to check on Android

Install and launch with the package ID currently configured by `init`:

```sh
adb install -r example.apk
adb shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
```

The first page should show the SVG icon, "koru" heading, a Spanish text line, `Contador: 0` and `JavaScript ejecutado`. Tap the counter twice: its label becomes `Contador: 2`. Open `Probar navegación`: the second page should display `Segunda página local`. Android Back should return to the first page. This checks local HTML, CSS, SVG, JavaScript and document navigation in the WebView.

The example's `app.js` does not call native plugins. To test the [JavaScript bridge]({{base}}en/plugins/), add a module script that imports `/__koru/core.mjs`. The browser preview tests only the web portion; Android is required to verify the bridge.

For the recorded service-worker check, copy `tests/examples/sw.js` to `my-app/www/sw.js` and register it from your page. The recorded [Android evidence]({{base}}en/android-results/) also covers rendering with network radios disabled.

## Read the files

- [index.html]({{base}}original/examples/web/index.html ":ignore") — markup and local asset references.
- [app.js]({{base}}original/examples/web/app.js ":ignore") — counter and status update.
- [style.css]({{base}}original/examples/web/style.css ":ignore") — styling.
- [icon.svg]({{base}}original/examples/web/icon.svg ":ignore") — local vector image.
- [second.html]({{base}}original/examples/web/second.html ":ignore") — navigation target.

These links show the source code. Install the APK on Android to run the example.
