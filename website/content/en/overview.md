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

## From a web folder to an Android app

Your framework builds the web assets. koru takes it from there.

```sh
koru init ./my-app
koru build ./my-app --release --output app.apk
koru verify app.apk
```

A compiled koru binary builds apps offline, without Go installed. The DEX compiler, binary XML, resource table, aligned APK packaging and APK v2 signing are implemented in Go with the standard library. No generated Java or Kotlin, JVM, Gradle, Android SDK build tools or wrapper framework enters the normal build.

## Small runtime, useful connections

The WebView serves packaged assets at `https://app.koru.invalid`. A MessagePort bridge connects your JavaScript to six built-in plugins: device, app, storage, vibration, share and notifications. JNI and custom DEX plugins are optional.

```js
import { device, storage } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("theme", "light");
```

[Explore the JavaScript API]({{base}}en/plugins/)

## Read the evidence, know the limits

Generated APKs have been installed, launched and updated on **Android 11, API 30, x86_64**. Recorded checks cover local rendering without Wi-Fi or mobile data, native plugins, service workers, navigation, the keyboard and storage persistence.

That is a specific test matrix, not a guarantee for every Android device. koru is `0.2.0-dev`; AAB, APK v3/v4 signatures, scheduled alarms and a broader device matrix remain outside the current scope.

[Android execution records]({{base}}en/android-results/) · [Milestones and scope]({{base}}en/milestones/)
