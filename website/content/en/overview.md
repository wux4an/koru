<div class="hero">
<div class="hero-copy">
<p class="release-note"><span class="status-dot"></span> An independent Android toolchain</p>
<h1>Your web app.<br>A little more native.</h1>
<p class="hero-description">Give your HTML, CSS and JavaScript a home on Android. koru builds a signed APK directly in Go, with a local WebView and a small native bridge.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}en/getting-started/">Build your first app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}en/architecture-apk/">See how it works</a></div>
</div>
<div class="clay-scene" role="img" aria-label="A web browser becomes an Android package">
<div class="clay-orbit"></div><div class="clay-browser"><div class="browser-dots"><b></b><b></b><b></b></div><i class="ph-duotone ph-code" aria-hidden="true"></i><span>HTML · CSS · JS</span></div>
<div class="clay-connector"><i class="ph-duotone ph-arrow-bend-down-right" aria-hidden="true"></i></div>
<div class="clay-android"><i class="ph-duotone ph-android-logo" aria-hidden="true"></i><span>app.apk</span></div>
<div class="clay-stamp"><i class="ph-duotone ph-seal-check" aria-hidden="true"></i> Signed in Go</div>
</div>
</div>

<figure class="hero-recording">
<video controls playsinline preload="none" poster="{{base}}assets/hero-poster.png" aria-label="Real koru CLI demonstration: initialize, build and verify an Android APK"><source src="{{base}}assets/hero-demo.mp4" type="video/mp4"><a href="{{base}}assets/hero-demo.mp4">Download the terminal recording</a></video>
<figcaption>A real terminal session recorded with VHS: one project, one signed APK, one verification. The commands and their output are written below.</figcaption>
</figure>

## From a web folder to an Android app

Koru takes the static output of your web app and packages it as an APK. It does not run your web framework's build step. Start with the project created by `koru init`, or point `webDir` to your existing output folder.

### 1. Create a project, then edit its page

Run `./bin/koru init ./my-app` from the koru repository root. Replace the generated `my-app/www/index.html` with:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Hello Android</title>
    <style>
      body {
        font: 1.1rem system-ui;
        padding: 2rem;
      }
      button {
        padding: 0.8rem 1rem;
        font: inherit;
      }
    </style>
  </head>
  <body>
    <h1>Hello Android</h1>
    <button id="tap">Taps: 0</button>
    <script>
      let taps = 0;
      document.querySelector("#tap").addEventListener("click", (event) => {
        event.currentTarget.textContent = "Taps: " + ++taps;
      });
    </script>
  </body>
</html>
```

The button updates on each tap. Koru also accepts separate local CSS, JavaScript, images and fonts. The existing [web example]({{base}}en/web-example/) shows navigation, plugins and offline behavior.

### 2. Build and inspect the APK

Run these from the koru repository root after [building the CLI]({{base}}en/getting-started/):

```sh
./bin/koru build ./my-app --release --output app.apk
./bin/koru verify app.apk
```

The build prints its cache state, APK size, signature scheme and certificate fingerprint. `verify` returns JSON that tools can parse:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
{"appId":"com.example.koru","entries":7,"bytes":16785,
 "certificateSha256":"3ecd2de5…","signatureScheme":"APK v2 RSA/SHA-256"}
```

These values are illustrative: your HTML, assets and signing identity change the size and fingerprint. The recording above shows an unedited real run. `verify` checks structure and signature; install the APK on Android to check execution.

## JavaScript can reach Android

Packaged pages run at `https://app.koru.invalid` and can call the six included plugins through a MessagePort bridge. For example, in a module script inside your app:

```js
import { device, storage } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("last-device", info.model);
```

This API runs inside a koru APK; an ordinary browser preview has no native bridge. See [JavaScript and plugins]({{base}}en/plugins/) for each method, parameters and error handling.

## What is ready today

The compiled CLI builds offline and uses Go's standard library for DEX, binary XML, resources, APK packaging and v2 signing. It does not generate Java or Kotlin, and the normal APK build does not invoke a JDK, Gradle or Android SDK build tools.

Generated APKs have been installed, launched and updated on **Android 11, API 30, x86_64**. The recorded checks cover rendering without Wi-Fi or mobile data, native plugins, service workers, navigation, keyboard input and storage persistence. This is a specific matrix, so test your own app on the devices you support. Koru is `0.2.0-dev`; AAB, APK v3/v4 signatures and a broader device matrix remain outside the current scope.

[Complete first-app guide]({{base}}en/getting-started/) · [Configuration]({{base}}en/configuration/) · [Android evidence]({{base}}en/android-results/)
