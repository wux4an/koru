# Solución de problemas

Comienza con `./bin/koru doctor ./my-app`. Comprueba la configuración, la identidad de firma, el archivo web inicial y los manifests de plugins antes del build. Los comandos parten de la raíz del repositorio y del proyecto de [Primeros pasos]({{base}}es/getting-started/).

## El comando de Go no funciona

Si `go version` abre un juego o dice que no encuentra el comando, revisa `command -v go` y coloca un compilador Go 1.23+ real al inicio de `PATH`. En el entorno de desarrollo registrado, `/usr/bin/go` es un juego. Go solo hace falta para compilar la CLI; el binario compilado genera el APK sin Go.

## `doctor` o `build` rechazan el proyecto

| Síntoma                             | Qué revisar                                                                                                                                                                 |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| No se puede leer `koru.json`        | Pasa la carpeta que contiene `koru.json`, por ejemplo `./my-app`. El JSON debe ser válido y usar solo campos documentados.                                                  |
| Falta `index.html`                  | Compila primero el frontend. `webDir` debe apuntar a una carpeta con `index.html` regular en la raíz.                                                                       |
| Falta la identidad de firma         | Recupera el mismo PEM con clave y certificado indicado por `signingKey`. `build` no lo reemplaza. Solo para una app nueva: `koru keygen --output my-app/.koru/signing.pem`. |
| El APK debe estar fuera de `webDir` | Mueve `--output` fuera de la carpeta de assets; desde la raíz, `--output app.apk`.                                                                                          |
| Plugin no disponible                | Activa su ID en `plugins`. Para uno personalizado, añade el manifest a `pluginManifests` y revisa ABI y handler.                                                            |
| Fallo al reunir assets              | Elimina symlinks, claves privadas y nombres inseguros de `webDir`. Máximo: 64 MiB por archivo, 256 MiB y 2.000 archivos en total.                                           |
| Error de configuración o permisos   | Usa identificadores Android completos como `android.permission.INTERNET`, IDs únicos y SDK configurados entre 26 y 36.                                                      |

`doctor` revisa el punto de entrada; el build inspecciona además todos los assets. Corrige la ruta indicada y repite el comando. No muevas la clave privada a `webDir`.

## Página en blanco o ruta rota

Abre el `index.html` generado y revisa las rutas a scripts, estilos e imágenes. Incluye todos estos archivos dentro de `webDir`. Koru intercepta `https://app.koru.invalid`; los scripts e iframes externos están bloqueados. Si el frontend genera rutas absolutas `/assets/...`, esos archivos deben existir en `webDir/assets/`. Si genera un prefijo `/nombre-del-proyecto/`, cambia la ruta base del frontend para el APK.

Para rutas cliente como `/ajustes`, usa `"spa": true` en `koru.json`. Una ruta con extensión que no exista sigue devolviendo 404; así un CSS o JS perdido no recibe HTML. Empaqueta también los service workers como assets locales permitidos.

Puedes ver la parte web en un navegador de escritorio, pero `/__koru/core.mjs` y los métodos nativos solo funcionan dentro del APK. Protege esas llamadas en la vista previa o pruébalas en Android.

## Fallo al instalar o actualizar

Comprueba el dispositivo con `adb devices`. Verifica e instala exactamente el APK que usarás:

```sh
./bin/koru verify app.apk
adb install -r app.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Cambia `com.example.hello` por tu `appId`. `INSTALL_FAILED_UPDATE_INCOMPATIBLE` indica que el ID o el certificado no coinciden con la app instalada. Para actualizar, conserva el `appId` y la clave `.koru/signing.pem` originales; aumentar `versionCode` no arregla una firma distinta. En una versión nueva incrementa `versionCode`.

`verify` comprueba estructura y firma v2, pero no renderizado WebView, resolución del framework ni plugins. Instala y ejercita la app en los Android que soportarás. La [matriz ejecutada]({{base}}es/android-results/) cubre Android 11/API 30/x86_64.

## Falla una llamada nativa

Comprueba que el plugin esté activo y que los argumentos coincidan con [la API JavaScript]({{base}}es/plugins/). Maneja las promesas rechazadas y consulta su código (`METHOD_NOT_FOUND`, `PERMISSION_DENIED`, `TIMEOUT` o `CHANNEL_ERROR`, por ejemplo). Un permiso en el manifest aún puede requerir la decisión del usuario en Android recientes. El diálogo de notificaciones de API 33+ no se ha probado en la matriz API 30.

Para reportar un fallo, incluye comando, código de salida, error exacto, `koru.json` sin datos privados, API/ABI Android y resultado de `koru verify`. Nunca compartas `signing.pem`.

[Referencia CLI]({{base}}es/cli/) · [Configuración]({{base}}es/configuration/) · [Seguridad]({{base}}es/security/)
