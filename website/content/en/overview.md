<div class="hero">
<div class="hero-copy">
<h1>Your web app, on Android.</h1>
<p class="hero-description">Create, preview and ship Android apps with HTML, CSS and JavaScript. koru is a single Go binary: it builds signed APKs without Java, Gradle or the Android SDK.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}docs/en/getting-started/">Create your first app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}docs/en/existing-project/">I have a web project</a></div>
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
<p>With koru installed, these four commands take you from nothing to an app on your phone:</p>

```sh
koru create my-app   # pick a name and a template
cd my-app
koru dev             # preview in the browser with live reload
koru run             # install and open on a device or emulator
```

New to koru? [Install it in one step]({{base}}docs/en/getting-started/).

</div>

## What would you like to do?

<div class="doc-paths">
<a class="doc-path" href="{{base}}docs/en/getting-started/"><strong>Create an app</strong><span>Install koru, start from a template and open it on Android.</span></a>
<a class="doc-path" href="{{base}}docs/en/existing-project/"><strong>Use your framework</strong><span>Package Vite, React, Vue, Svelte, Astro or any static output.</span></a>
<a class="doc-path" href="{{base}}docs/en/develop/"><strong>Develop faster</strong><span>Live reload in the browser with the same JavaScript API.</span></a>
<a class="doc-path" href="{{base}}docs/en/plugins/"><strong>Add Android features</strong><span>Store data, share text, vibrate and show notifications.</span></a>
<a class="doc-path" href="{{base}}docs/en/distribution/"><strong>Share your app</strong><span>Build a release APK and keep your key for updates.</span></a>
<a class="doc-path" href="{{base}}docs/en/comparison/"><strong>Compare</strong><span>How koru relates to Capacitor and Expo.</span></a>
</div>

## How it works

You keep your web tools. koru packages the static files in `webDir` into an APK whose WebView serves them from a private HTTPS origin. A small JavaScript module gives your code access to Android features through a secure message channel. During development, `koru dev` serves the same files in your browser with web implementations of that API, and `koru run` installs the real app on a device.

The APK toolchain is written in Go with its standard library: DEX bytecode, the binary manifest, resources, ZIP packaging and the v2 signature. Builds are offline and take milliseconds. Your app works offline too; to call an HTTPS service, add the [Internet permission]({{base}}docs/en/configuration/).

<details class="walkthrough"><summary>Watch a terminal example</summary>
<figure class="hero-recording">
<picture><source media="(prefers-reduced-motion: reduce)" srcset="{{base}}assets/hero-poster.png"><img src="{{base}}assets/hero-demo.gif" width="960" height="360" loading="lazy" alt="Create a project, build an APK and verify its signature"></picture>
<figcaption>An earlier recording of the init, build and verify commands in a real session.</figcaption>
</figure>
</details>

### Before you distribute

koru is in development. It produces APKs for direct installation; it does not yet produce AABs for Google Play, and it does not build iOS apps. Recorded execution covers an Android 11 (API 30), x86_64 emulator. Test your app on the devices you plan to support. See [tested compatibility]({{base}}docs/en/android-results/).
