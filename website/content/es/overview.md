<div class="hero">
<div class="hero-copy">
<p class="release-note"><span class="status-dot"></span> Una toolchain Android independiente</p>
<h1>Tu app web.<br>Un poco más nativa.</h1>
<p class="hero-description">Dale un hogar en Android a tu HTML, CSS y JavaScript. koru genera un APK firmado directamente en Go, con un WebView local y un pequeño puente nativo.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}es/getting-started/">Crea tu primera app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}es/architecture-apk/">Cómo funciona</a></div>
</div>
<div class="clay-scene" role="img" aria-label="Un navegador web se convierte en un paquete Android">
<div class="clay-orbit"></div><div class="clay-browser"><div class="browser-dots"><b></b><b></b><b></b></div><i class="ph-duotone ph-code" aria-hidden="true"></i><span>HTML · CSS · JS</span></div>
<div class="clay-connector"><i class="ph-duotone ph-arrow-bend-down-right" aria-hidden="true"></i></div>
<div class="clay-android"><i class="ph-duotone ph-android-logo" aria-hidden="true"></i><span>app.apk</span></div>
<div class="clay-stamp"><i class="ph-duotone ph-seal-check" aria-hidden="true"></i> Firmado en Go</div>
</div>
</div>

## De una carpeta web a una app Android

Tu framework genera los archivos web. koru se encarga del siguiente paso.

```sh
koru init ./mi-app
koru build ./mi-app --release --output app.apk
koru verify app.apk
```

El binario compilado de koru construye apps sin conexión y sin tener Go instalado. El compilador DEX, XML binario, tabla de recursos, empaquetado APK alineado y firma APK v2 están implementados en Go con biblioteca estándar. El build normal no genera Java ni Kotlin, ni utiliza JVM, Gradle, herramientas de build del SDK Android o frameworks de wrappers.

## Un runtime pequeño, conexiones útiles

El WebView sirve los assets empaquetados en `https://app.koru.invalid`. Un puente MessagePort conecta JavaScript con seis plugins: device, app, storage, vibration, share y notifications. JNI y los plugins DEX personalizados son opcionales.

```js
import { device, storage } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("theme", "light");
```

[Explora la API JavaScript]({{base}}es/plugins/)

## Consulta la evidencia y conoce los límites

Los APK generados se han instalado, arrancado y actualizado en **Android 11, API 30, x86_64**. Las comprobaciones registradas cubren renderizado local sin Wi-Fi ni datos, plugins nativos, service workers, navegación, teclado y persistencia de storage.

Es una matriz concreta, no una garantía para todos los dispositivos Android. koru está en `0.2.0-dev`; AAB, firmas APK v3/v4, alarmas programadas y una matriz más amplia quedan fuera del alcance actual.

[Registros de ejecución Android]({{base}}es/android-results/) · [Hitos y alcance]({{base}}es/milestones/)
