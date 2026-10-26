<div class="hero">
<div class="hero-copy">
<h1>Tu app web, en Android.</h1>
<p class="hero-description">Crea, previsualiza y publica apps Android con HTML, CSS y JavaScript. koru es un solo binario de Go: genera APKs firmados sin Java, Gradle ni el SDK de Android.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}es/getting-started/">Crea tu primera app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}es/existing-project/">Ya tengo un proyecto web</a></div>
</div>
<div class="clay-scene" role="img" aria-label="Un navegador con HTML, CSS y JavaScript se transforma en una app APK de Android">
<div class="clay-orbit"></div>
<div class="clay-browser"><div class="browser-dots"><b></b><b></b><b></b></div><i class="ph-duotone ph-code" aria-hidden="true"></i><span>HTML · CSS · JS</span></div>
<div class="clay-connector"><i class="ph-duotone ph-arrow-bend-down-right" aria-hidden="true"></i></div>
<div class="clay-android"><i class="ph-duotone ph-android-logo" aria-hidden="true"></i><span>app.apk</span></div>
<div class="clay-stamp"><i class="ph-duotone ph-seal-check" aria-hidden="true"></i> Firmado en Go</div>
</div>
</div>

<div class="quick-start">
<p>Con koru instalado, estos cuatro comandos te llevan de cero a una app en tu móvil:</p>

```sh
koru create mi-app   # elige un nombre y una plantilla
cd mi-app
koru dev             # vista previa en el navegador con recarga en vivo
koru run             # instala y abre en un dispositivo o emulador
```

¿Es tu primera vez? [Instala koru en un paso]({{base}}es/getting-started/).

</div>

## ¿Qué quieres hacer?

<div class="doc-paths">
<a class="doc-path" href="{{base}}es/getting-started/"><strong>Crear una app</strong><span>Instala koru, parte de una plantilla y ábrela en Android.</span></a>
<a class="doc-path" href="{{base}}es/existing-project/"><strong>Usar tu framework</strong><span>Empaqueta Vite, React, Vue, Svelte, Astro o cualquier salida estática.</span></a>
<a class="doc-path" href="{{base}}es/develop/"><strong>Desarrollar más rápido</strong><span>Recarga en vivo en el navegador con la misma API JavaScript.</span></a>
<a class="doc-path" href="{{base}}es/plugins/"><strong>Añadir funciones de Android</strong><span>Guarda datos, comparte texto, vibra y muestra notificaciones.</span></a>
<a class="doc-path" href="{{base}}es/distribution/"><strong>Compartir tu app</strong><span>Genera un APK release y conserva tu clave para actualizar.</span></a>
<a class="doc-path" href="{{base}}es/comparison/"><strong>Comparar</strong><span>Cómo se relaciona koru con Capacitor y Expo.</span></a>
</div>

## Cómo funciona

Sigues usando tus herramientas web. koru empaqueta los archivos estáticos de `webDir` en un APK cuyo WebView los sirve desde un origen HTTPS privado. Un pequeño módulo JavaScript da a tu código acceso a funciones de Android mediante un canal de mensajes seguro. Mientras desarrollas, `koru dev` sirve los mismos archivos en tu navegador con implementaciones web de esa API, y `koru run` instala la app real en un dispositivo.

La cadena de compilación del APK está escrita en Go con su biblioteca estándar: bytecode DEX, manifiesto binario, recursos, empaquetado ZIP y firma v2. Las builds funcionan sin conexión y tardan milisegundos. Tu app también funciona sin conexión; para llamar a un servicio HTTPS, añade el [permiso de Internet]({{base}}es/configuration/).

<details class="walkthrough"><summary>Ver un ejemplo en la terminal</summary>
<figure class="hero-recording">
<picture><source media="(prefers-reduced-motion: reduce)" srcset="{{base}}assets/hero-poster.png"><img src="{{base}}assets/hero-demo.gif" width="960" height="360" loading="lazy" alt="Crear un proyecto, generar un APK y verificar su firma"></picture>
<figcaption>Una grabación anterior de los comandos init, build y verify en una sesión real.</figcaption>
</figure>
</details>

### Antes de distribuir

koru está en desarrollo. Genera APKs para instalación directa; todavía no genera AAB para Google Play ni apps iOS. La ejecución registrada cubre un emulador Android 11 (API 30) x86_64. Prueba tu app en los dispositivos que quieras soportar. Consulta la [compatibilidad probada]({{base}}es/android-results/).
