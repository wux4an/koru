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

<figure class="hero-recording">
<video controls playsinline preload="none" poster="{{base}}assets/hero-poster.png" aria-label="Demostración real de la CLI: iniciar, compilar y verificar un APK Android"><source src="{{base}}assets/hero-demo.mp4" type="video/mp4"><a href="{{base}}assets/hero-demo.mp4">Descargar la grabación de terminal</a></video>
<figcaption>Sesión de terminal real grabada con VHS: un proyecto, un APK firmado y una verificación. Los comandos y su salida aparecen abajo.</figcaption>
</figure>

## De una carpeta web a una app Android

Koru toma la salida estática de tu aplicación web y la empaqueta como APK. No ejecuta el build de tu framework. Comienza con el proyecto de `koru init` o configura `webDir` para usar una carpeta ya compilada.

### 1. Inicia el proyecto y edita su página

Ejecuta `./bin/koru init ./my-app` desde la raíz del repositorio. Reemplaza el archivo generado `my-app/www/index.html` por:

```html
<!doctype html>
<html lang="es">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Hola Android</title>
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
    <h1>Hola Android</h1>
    <button id="tap">Toques: 0</button>
    <script>
      let taps = 0;
      document.querySelector("#tap").addEventListener("click", (event) => {
        event.currentTarget.textContent = "Toques: " + ++taps;
      });
    </script>
  </body>
</html>
```

El botón cambia con cada toque. Koru también acepta CSS, JavaScript, imágenes y fuentes locales en archivos separados. El [ejemplo web]({{base}}es/web-example/) muestra navegación, plugins y uso sin red.

### 2. Genera e inspecciona el APK

Ejecuta estos comandos desde la raíz del repositorio tras [compilar la CLI]({{base}}es/getting-started/):

```sh
./bin/koru build ./my-app --release --output app.apk
./bin/koru verify app.apk
```

El build muestra el estado de caché, tamaño, tipo de firma y huella del certificado. `verify` devuelve JSON procesable:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
{"appId":"com.example.koru","entries":7,"bytes":16785,
 "certificateSha256":"3ecd2de5…","signatureScheme":"APK v2 RSA/SHA-256"}
```

Son valores ilustrativos: tu HTML, assets e identidad de firma cambian el tamaño y la huella. La grabación anterior enseña una ejecución real sin modificar. `verify` comprueba estructura y firma; instala el APK en Android para comprobar su ejecución.

## JavaScript puede llamar a Android

Las páginas empaquetadas viven en `https://app.koru.invalid` y acceden a seis plugins mediante MessagePort. Por ejemplo, dentro de un script de módulo de tu app:

```js
import { device, storage } from "/__koru/core.mjs";

const info = await device.getInfo();
await storage.set("last-device", info.model);
```

Esta API funciona dentro del APK de koru; un navegador común no tiene el puente nativo. En [JavaScript y plugins]({{base}}es/plugins/) están los métodos, parámetros y errores.

## Alcance actual

La CLI compilada construye sin red y usa Go estándar para DEX, XML binario, recursos, empaquetado APK y firma v2. No genera Java ni Kotlin; el build normal no invoca JDK, Gradle ni herramientas de build del SDK Android.

Los APK generados se han instalado, arrancado y actualizado en **Android 11, API 30, x86_64**. Las pruebas registradas cubren renderizado sin Wi-Fi ni datos, plugins, service workers, navegación, teclado y persistencia de datos. Es una matriz concreta: prueba tu aplicación en los dispositivos que vas a soportar. Koru está en `0.2.0-dev`; AAB, firmas APK v3/v4 y una matriz mayor quedan fuera del alcance actual.

[Guía completa del primer APK]({{base}}es/getting-started/) · [Configuración]({{base}}es/configuration/) · [Evidencia Android]({{base}}es/android-results/)
