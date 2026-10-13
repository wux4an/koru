# Seguridad y firmas

koru empaqueta código de aplicación de confianza. Sus comprobaciones delimitan assets, navegación y mensajes del puente; no aíslan un plugin firmado malicioso de la aplicación.

## Conserva una identidad de firma

`init` y `keygen` crean explícitamente un PEM con clave y certificado, con permisos `0600`. `build` exige una identidad válida existente. `--release` la reutiliza y deshabilita debug. Guarda una copia segura fuera del control de versiones y de los assets. Las actualizaciones Android necesitan el mismo certificado.

APK v2 usa RSA PKCS#1 v1.5/SHA-256, algoritmo `0x0103`, y bloques digest de 1 MiB. Los recursos se alinean antes de firmar; las bibliotecas nativas usan 16 KiB. El APK firmado no se modifica después. `verify` comprueba firma, certificado, digest, ZIP, DEX, manifiesto, recursos e icono.

## Un origen HTTPS local

El runtime sirve solo assets empaquetados de su allowlist en `https://app.koru.invalid`. Comprueba esquema y autoridad; no concatena rutas recibidas con el filesystem del host. Deshabilita file/content. Los assets desconocidos devuelven 404; SPA solo aplica a rutas sin extensión.

Los service workers usan la misma allowlist mediante Application context, sin retener Activity. La red se bloquea salvo permiso `INTERNET` explícito. CSP bloquea scripts y frames externos, pero permite scripts empaquetados, inline/eval que ciertas salidas web necesitan y conexiones HTTPS autorizadas. No es una CSP estricta sin inline.

## Puente MessagePort

El puerto se transfiere al frame principal con targetOrigin explícito. No existe `addJavascriptInterface`. El receptor comprueba origen del documento, frame principal, evento nativo trusted y ausencia de Window source.

Las peticiones usan protocolo v1, IDs, métodos y parámetros. Solo se aceptan handlers activados. El cliente limita peticiones pendientes, aplica timeouts, cierra promesas al salir y permite cancelar antes del dispatch. Cancelar no deshace una acción ya ejecutada en Android.

Los plugins DEX y JNI comparten UID y permisos. Son código de confianza del desarrollador. Valida parámetros, respeta el hilo UI y aporta la ABI nativa correcta.

## Qué demuestra la verificación

Un APK estructuralmente válido no demuestra resolución ART, ciclo de vida, renderizado o seguridad de código arbitrario. Las comprobaciones registradas corresponden al runtime probado y la matriz API 30/x86_64. Valida tus propios artefactos en Android.

[Decisiones del runtime]({{base}}es/architecture-apk/) · [Evidencia de ejecución]({{base}}es/android-results/)
