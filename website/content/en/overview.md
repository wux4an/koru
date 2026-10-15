<div class="hero">
<div class="hero-copy">
<h1>Your web app, on Android.</h1>
<p class="hero-description">Turn your HTML, CSS and JavaScript into an Android app. koru packages your web files into a signed APK that you can install and test on your device.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}en/getting-started/">Create your first app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}en/existing-project/">I have a web project</a></div>
</div>
<div class="clay-scene" role="img" aria-label="A browser with HTML, CSS and JavaScript transforms into an Android APK app">
<div class="clay-orbit"></div>
<div class="clay-browser"><div class="browser-dots"><b></b><b></b><b></b></div><i class="ph-duotone ph-code" aria-hidden="true"></i><span>HTML · CSS · JS</span></div>
<div class="clay-connector"><i class="ph-duotone ph-arrow-bend-down-right" aria-hidden="true"></i></div>
<div class="clay-android"><i class="ph-duotone ph-android-logo" aria-hidden="true"></i><span>app.apk</span></div>
<div class="clay-stamp"><i class="ph-duotone ph-seal-check" aria-hidden="true"></i> Go signed</div>
</div>
</div>

<div class="quick-start">
<p>Once koru is built, start with a sample page:</p>

```sh
./bin/koru init ./my-app
./bin/koru build ./my-app --output app.apk
```

Edit `my-app/www/index.html` and rebuild the APK to include your changes. Need koru first? Follow the [installation steps]({{base}}en/getting-started/).

</div>

## What would you like to do?

<div class="doc-paths">
<a class="doc-path" href="{{base}}en/existing-project/"><strong>Use your web project</strong><span>Connect the output of React, Vue, Svelte or any static frontend.</span></a>
<a class="doc-path" href="{{base}}en/plugins/"><strong>Add Android features</strong><span>Store data, share text and show notifications from JavaScript.</span></a>
<a class="doc-path" href="{{base}}en/distribution/"><strong>Share your app</strong><span>Prepare a release APK and keep the signing identity for updates.</span></a>
<a class="doc-path" href="{{base}}en/troubleshooting/"><strong>Solve a problem</strong><span>Find help with assets, installation and plugin calls.</span></a>
</div>

## How it fits your project

Keep using your web tools. If you use a framework, build its static output first and point koru to it with `webDir`. Your app displays those files in Android’s WebView; the JavaScript API gives you access to device features.

The koru executable builds APKs without Java, Gradle or Android Studio. Your local files also work offline. To call an HTTPS service, configure the [Internet permission]({{base}}en/configuration/).

<details class="walkthrough"><summary>Watch a terminal example</summary>
<figure class="hero-recording">
<picture><source media="(prefers-reduced-motion: reduce)" srcset="{{base}}assets/hero-poster.png"><img src="{{base}}assets/hero-demo.gif" width="960" height="360" loading="lazy" alt="Create a project, build an APK and verify its signature"></picture>
<figcaption>The init, build and verify commands in a real session.</figcaption>
</figure>
</details>

### Before you distribute

koru is in development. It produces APKs for direct installation; it does not yet produce AABs for Google Play. Recorded execution covers an Android 11 (API 30), x86_64 emulator. Test your app on the devices you plan to support. See [tested compatibility]({{base}}en/android-results/) for details.
