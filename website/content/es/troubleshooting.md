# Solución de problemas

Comienza con `koru doctor` en el directorio de tu proyecto. Comprueba la configuración, la identidad de firma, el archivo web inicial y los manifests de plugins antes del build, e indica si hay un servidor adb y dispositivos disponibles. Cada mensaje de error nombra el campo o archivo implicado y cómo corregirlo.

## No se encuentra `koru` o Go

Si `go version` devuelve «command not found», instala Go 1.23 o posterior. Solo necesitas Go para compilar el ejecutable de koru. Si no se encuentra `koru` después de `make install`, añade `~/.local/bin` a tu `PATH`.

## `doctor` o `build` rechazan el proyecto

| Síntoma                             | Qué revisar                                                                                                                                                                 |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `no koru.json in …`                 | Ejecuta el comando dentro del proyecto o pasa su carpeta (`koru build ./mi-app`). Crea uno con `koru create` o `koru init`.                                                 |
| `koru.json: unknown field …`        | Corrige el nombre del campo; la línea `$schema` da a tu editor autocompletado de todos los campos válidos.                                                                  |
| `webDir … has no index.html`        | Compila primero el frontend (por ejemplo, `npm run build`). `webDir` debe apuntar a la carpeta con `index.html` en la raíz.                                                 |
| Falta la identidad de firma         | Recupera el mismo PEM con clave y certificado indicado por `signingKey`. `build` no lo reemplaza. Solo para una app nueva: `koru keygen --output mi-app/.koru/signing.pem`. |
| El APK debe estar fuera de `webDir` | Mueve `--output` fuera de la carpeta de assets; desde la raíz, `--output app.apk`.                                                                                          |
| Plugin no disponible                | Activa su ID en `plugins`. Para uno personalizado, añade el manifest a `pluginManifests` y revisa ABI y handler.                                                            |
| Fallo al reunir assets              | Elimina symlinks, claves privadas y nombres inseguros de `webDir`. Máximo: 64 MiB por archivo, 256 MiB y 2.000 archivos en total.                                           |
| Error de configuración o permisos   | Usa identificadores Android completos como `android.permission.INTERNET`, IDs únicos y SDK configurados entre 26 y 36.                                                      |

`doctor` revisa el punto de entrada; el build inspecciona además todos los assets. Corrige la ruta indicada y repite el comando. No muevas la clave privada a `webDir`.

## Página en blanco o ruta rota

Abre el `index.html` generado y revisa las rutas a scripts, estilos e imágenes. Incluye todos estos archivos dentro de `webDir`. Koru intercepta `https://app.koru.invalid`; los scripts e iframes externos están bloqueados. Si el frontend genera rutas absolutas `/assets/...`, esos archivos deben existir en `webDir/assets/`. Si genera un prefijo `/nombre-del-proyecto/`, cambia la ruta base del frontend para el APK.

Para rutas cliente como `/ajustes`, usa `"spa": true` en `koru.json`. Una ruta con extensión que no exista sigue devolviendo 404; así un CSS o JS perdido no recibe HTML. Empaqueta también los service workers como assets locales permitidos.

En un navegador, previsualiza con `koru dev` en lugar de abrir el archivo directamente: sirve `/__koru/core.mjs` con implementaciones web de la API. Un servidor estático cualquiera o `file://` no pueden cargarlo. El comportamiento nativo (permisos, menú de compartir de Android, notificaciones) sigue necesitando una prueba en Android con `koru run`.

## Problemas con `koru dev`

| Síntoma                             | Qué revisar                                                                                                                            |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `cannot listen on …`                | Otro programa usa el puerto. Ejecuta `koru dev --port 3000`, o `--port 0` para cualquier puerto libre.                                 |
| La página no se recarga             | Guarda dentro del directorio del proyecto; `node_modules`, `.git` y `.koru` se ignoran. Revisa si la terminal muestra errores `Watch`. |
| `METHOD_NOT_FOUND … is not enabled` | Añade el plugin a `plugins` en `koru.json`; la siguiente recarga lo aplica.                                                            |
| `--proxy` muestra `cannot reach …`  | Inicia antes el servidor de desarrollo de tu framework y usa su URL exacta, por ejemplo `http://localhost:5173`.                       |
| El móvil no abre la URL de red      | Usa `--host 0.0.0.0`, mantén ambos dispositivos en la misma red y permite el puerto en tu cortafuegos.                                 |

## Fallo al instalar o actualizar

Ejecuta `koru devices`. `unauthorized` significa que debes aceptar el aviso de depuración USB en el móvil; `offline`, que debes reconectarlo. Si koru indica que no hay servidor adb, inicia uno con `adb start-server` o abre Android Studio. `koru run` explica los errores de instalación habituales; consulta [Ejecuta en un dispositivo]({{base}}es/run-on-device/). Para instalar a mano, verifica antes el APK exacto:

```sh
koru verify app.apk
adb install -r app.apk
adb shell am start -W -n com.example.miapp/com.koru.runtime.MainActivity
```

Cambia `com.example.miapp` por tu `appId`. `INSTALL_FAILED_UPDATE_INCOMPATIBLE` indica que el ID o el certificado no coinciden con la app instalada. Para actualizar, conserva el `appId` y la clave `.koru/signing.pem` originales; aumentar `versionCode` no arregla una firma distinta. En una versión nueva incrementa `versionCode`.

`verify` comprueba estructura y firma v2, pero no renderizado WebView, resolución del framework ni plugins. Instala y ejercita la app en los Android que soportarás. La [matriz ejecutada]({{base}}es/android-results/) cubre Android 11/API 30/x86_64.

## Falla una llamada nativa

Comprueba que el plugin esté activo y que los argumentos coincidan con [la API JavaScript]({{base}}es/plugins/). Maneja las promesas rechazadas y consulta su código (`METHOD_NOT_FOUND`, `PERMISSION_DENIED`, `TIMEOUT` o `CHANNEL_ERROR`, por ejemplo). Un permiso en el manifest aún puede requerir la decisión del usuario en Android recientes. El diálogo de notificaciones de API 33+ no se ha probado en la matriz API 30.

Para reportar un fallo, incluye comando, código de salida, error exacto, `koru.json` sin datos privados, API/ABI Android y resultado de `koru verify`. Nunca compartas `signing.pem`.

[Referencia CLI]({{base}}es/cli/) · [Configuración]({{base}}es/configuration/) · [Seguridad]({{base}}es/security/)
