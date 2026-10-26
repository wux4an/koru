# koru, Capacitor and Expo

Capacitor and Expo are mature, widely used tools. koru does much less, on purpose. This page helps you decide which one fits your project.

## The short version

- Choose **koru** when you have a web app, you only need Android, and you want a tiny toolchain: one Go binary that builds a signed APK offline in milliseconds, with no Java, Gradle or Android SDK.
- Choose **Capacitor** when your web app needs iOS as well as Android, a large plugin ecosystem, Google Play bundles (AAB) or native code you write in Kotlin, Java or Swift.
- Choose **Expo** when you want to write the app in React Native, with native UI components, iOS and Android, over-the-air updates and cloud builds.

## Side by side

|                      | koru                                                                                             | Capacitor                                                           | Expo                                              |
| -------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------- | ------------------------------------------------- |
| You write            | HTML, CSS, JavaScript; any static frontend                                                       | HTML, CSS, JavaScript; any web framework                            | React Native (JavaScript/TypeScript)              |
| UI renders in        | Android System WebView                                                                           | WebView (Android, iOS)                                              | Native views                                      |
| Android build needs  | The koru binary only                                                                             | Android Studio or Gradle, JDK, Android SDK                          | EAS Build (cloud), or JDK and Android SDK locally |
| Platforms            | Android                                                                                          | Android, iOS, web                                                   | Android, iOS, web                                 |
| Output               | Signed APK (v2)                                                                                  | APK, AAB                                                            | APK, AAB                                          |
| Start a project      | `koru create`: `basic`, Vite, React, Vue, Svelte or TypeScript templates                         | `npm init @capacitor/app`                                           | `npx create-expo-app`                             |
| Preview while coding | `koru dev`: browser, live reload, web implementations of the API                                 | Your framework’s dev server; live reload on device via `server.url` | Expo Go or a development build, with fast refresh |
| Run on a device      | `koru run`                                                                                       | `npx cap run android`                                               | `npx expo run:android`                            |
| Device logs          | `koru logs` in your terminal (not yet validated on a device)                                     | Android Studio Logcat, or `chrome://inspect` for the WebView        | Terminal and dev tools while the dev server runs  |
| Deep links           | `deepLinks` in `koru.json`; opens the app, URL not passed to JavaScript yet; untested on Android | Supported, including the URL in JavaScript (App plugin)             | Supported through Expo Router and linking         |
| Versioning           | `koru bump`, `koru build --bump`                                                                 | Edit Gradle or Xcode project files                                  | `version` in app config; EAS can auto-increment   |
| Native features      | 6 built-in plugins; custom plugins as DEX manifests or JNI libraries                             | Official and community plugins; write your own natively             | Expo SDK modules; write your own natively         |
| Maturity             | Experimental; recorded on Android 11 (API 30) x86_64                                             | Widely used in production                                           | Widely used in production                         |

## What koru does differently

**No native toolchain.** koru writes the DEX bytecode, binary manifest, resource table, ZIP and APK signature itself in Go. A build does not download anything or start a JVM, so it is fast and works offline. The trade-off is that you cannot add arbitrary Java or Kotlin code: custom plugins are described as DEX intermediate representation, or call a prebuilt JNI library.

**A small, strict runtime.** Your files are served from a private HTTPS origin, `https://app.koru.invalid`, and the JavaScript bridge is a MessagePort bound to that origin. There is no `addJavascriptInterface`. Network access is off unless you add the Internet permission.

**The same API in the browser.** Like Capacitor’s web implementations, koru’s JavaScript API runs in a desktop browser during `koru dev`, so you can build most of the app before installing it.

## What koru does not do yet

- iOS apps.
- Android App Bundles (AAB) for Google Play. koru produces APKs for direct installation and stores that accept APKs.
- Pass a deep link’s URL to your JavaScript. The link opens the app on its first page, and neither the verified-link flow nor any of this has been run on Android.
- Scheduled notifications, camera, geolocation, file pickers and other plugins outside the six built-in ones.
- Live reload inside the installed APK. Use `koru dev` in a browser, then `koru run`.
- Verified execution beyond the [recorded matrix]({{base}}en/android-results/).

If you need those today, Capacitor or Expo are the better choice. If you need a small, auditable Android toolchain for a web app, try koru: [create your first app]({{base}}en/getting-started/).

_Capacitor and Expo change quickly; check their [Capacitor](https://capacitorjs.com/docs) and [Expo](https://docs.expo.dev/) documentation for current details._
