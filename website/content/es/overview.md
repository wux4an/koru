<div class="hero">
<div class="hero-copy">
<h1>Tu web, en Android.</h1>
<p class="hero-description">Convierte tu HTML, CSS y JavaScript en una app Android. koru empaqueta tus archivos web en un APK firmado, listo para instalar y probar en tu dispositivo.</p>
<div class="hero-actions"><a class="primary-link" href="{{base}}es/getting-started/">Crea tu primera app <i class="ph-duotone ph-rocket-launch" aria-hidden="true"></i></a><a href="{{base}}es/existing-project/">Ya tengo un proyecto web</a></div>
</div>
<div class="clay-scene" role="img" aria-label="Un navegador con HTML, CSS y JavaScript se transforma en una app APK para Android">
<div class="clay-orbit"></div>
<div class="clay-browser"><div class="browser-dots"><b></b><b></b><b></b></div><i class="ph-duotone ph-code" aria-hidden="true"></i><span>HTML · CSS · JS</span></div>
<div class="clay-connector"><i class="ph-duotone ph-arrow-bend-down-right" aria-hidden="true"></i></div>
<div class="clay-android"><i class="ph-duotone ph-android-logo" aria-hidden="true"></i><span>app.apk</span></div>
<div class="clay-stamp"><i class="ph-duotone ph-seal-check" aria-hidden="true"></i> Firmado en Go</div>
</div>
</div>

<div class="quick-start">
<p>Con koru compilado, empieza con una página de ejemplo:</p>

```sh
./bin/koru init ./my-app
./bin/koru build ./my-app --output app.apk
```

Edita `my-app/www/index.html` y vuelve a generar el APK para incluir tus cambios. ¿Aún no tienes koru? Sigue la [instalación]({{base}}es/getting-started/).

</div>

## ¿Qué quieres hacer?

<div class="doc-paths">
<a class="doc-path" href="{{base}}es/existing-project/"><strong>Usa tu proyecto web</strong><span>Conecta la salida de React, Vue, Svelte o cualquier frontend estático.</span></a>
<a class="doc-path" href="{{base}}es/plugins/"><strong>Añade funciones de Android</strong><span>Guarda datos, comparte texto y muestra notificaciones desde JavaScript.</span></a>
<a class="doc-path" href="{{base}}es/distribution/"><strong>Comparte tu app</strong><span>Prepara un APK release y conserva la firma de tus actualizaciones.</span></a>
<a class="doc-path" href="{{base}}es/troubleshooting/"><strong>Resuelve un problema</strong><span>Encuentra ayuda con archivos, instalación y llamadas a plugins.</span></a>
</div>

## Cómo encaja en tu proyecto

Sigue usando tus herramientas web. Si usas un framework, compila primero su salida estática y dile a koru dónde está con `webDir`. La app muestra esos archivos en el WebView de Android; puedes añadir funciones del dispositivo con la API JavaScript.

El ejecutable de koru genera APK sin instalar Java, Gradle ni Android Studio. Tus archivos locales también están disponibles sin conexión. Para llamar a un servicio HTTPS, configura el [permiso de Internet]({{base}}es/configuration/).

<details class="walkthrough"><summary>Ver un ejemplo en la terminal</summary>
<figure class="hero-recording">
<picture><source media="(prefers-reduced-motion: reduce)" srcset="{{base}}assets/hero-poster.png"><img src="{{base}}assets/hero-demo.gif" width="960" height="360" loading="lazy" alt="Crear un proyecto, generar un APK y verificar su firma"></picture>
<figcaption>Los comandos init, build y verify en una sesión real.</figcaption>
</figure>
</details>

### Antes de distribuir

koru está en desarrollo. Genera APK para instalación directa; todavía no genera AAB para Google Play. La ejecución comprobada corresponde a un emulador Android 11 (API 30), x86_64. Prueba tu app en los dispositivos que quieras admitir. Consulta la [compatibilidad comprobada]({{base}}es/android-results/) si necesitas más detalle.
