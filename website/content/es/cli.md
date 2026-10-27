# Comandos CLI

Ejecuta `koru --help` para ver la lista de comandos y `koru <comando> --help` para las opciones y ejemplos de un comando. La ayuda está en inglés. Los errores dicen qué está mal y cómo corregirlo, y terminan con un código de salida distinto de cero. Define `NO_COLOR=1` para obtener la salida sin colores.

Los comandos que aceptan `[project]` usan por defecto el directorio actual, así que dentro de un proyecto basta con `koru dev`, `koru run` o `koru build`.

## Comandos de un vistazo

| Comando                         | Qué hace                                                                                                                                          |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `create [directorio]`           | Crea una app desde una plantilla (`basic`, `vite`, `react`, `vue`, `svelte`, `vanilla-ts`) con `koru.json`, `.gitignore`, tipos y clave de firma. |
| `dev [proyecto]`                | Sirve la app en un navegador con recarga en vivo e implementaciones web de la API JavaScript.                                                     |
| `run [proyecto]`                | Compila un APK de depuración, lo instala en un dispositivo o emulador conectado y lo abre. Con `--logs` sigue mostrando los logs.                 |
| `logs [proyecto]`               | Muestra en directo el logcat de la app y los mensajes de consola del WebView desde un dispositivo o emulador.                                     |
| `devices`                       | Lista los dispositivos y emuladores que ve el servidor adb, con su estado.                                                                        |
| `build [proyecto]`              | Compila y firma un APK. Depuración por defecto; `--release` para distribuir; `--bump` sube la versión antes.                                      |
| `bump [nivel] [proyecto]`       | Sube `versionCode` y la versión de `koru.json`: `patch`, `minor`, `major` o `code`.                                                               |
| `verify <app.apk>`              | Comprueba la firma y la estructura de un APK e imprime un informe JSON.                                                                           |
| `doctor [proyecto]`             | Revisa la configuración, la clave, la entrada web y los plugins, avisa de errores comunes e indica si hay un servidor adb accesible.              |
| `plugin list\|add\|remove`      | Muestra, activa o desactiva los plugins integrados en `koru.json`.                                                                                |
| `config [proyecto]`             | Imprime la configuración resuelta como JSON, con los permisos que añade.                                                                          |
| `deeplinks [proyecto]`          | Imprime el `assetlinks.json` que verifica tus deep links `https`.                                                                                 |
| `init [directorio]`             | Añade a un directorio un `koru.json` mínimo, una página de ejemplo y una clave, para proyectos existentes.                                        |
| `js [directorio]`               | Copia el módulo JavaScript y sus tipos a tus fuentes, para bundlers. Por defecto `src/koru`.                                                      |
| `keygen --output <archivo.pem>` | Crea una identidad de firma nueva. No sobrescribe.                                                                                                |
| `dex build` / `dex inspect`     | Compila o inspecciona archivos DEX independientes.                                                                                                |
| `version`                       | Muestra la versión y la plataforma.                                                                                                               |

Un binario compilado no necesita Go, Java, el SDK de Android ni red para compilar apps. `dev`, `run` y `devices` son ayudas para desarrollar; no forman parte de la compilación del APK.

## `create`: parte de una plantilla

```sh
koru create mi-app
koru create mi-app --template react --name "Mi App" --app-id com.example.miapp
```

| Opción             | Valor por defecto               | Significado                                                       |
| ------------------ | ------------------------------- | ----------------------------------------------------------------- |
| `-t`, `--template` | `basic`                         | Plantilla inicial; consulta la tabla siguiente.                   |
| `--name`           | Del directorio: `Mi App`        | Nombre que aparece bajo el icono.                                 |
| `--app-id`         | Del nombre: `com.example.miapp` | Identificador de la app Android. Elige el tuyo antes de publicar. |

| Plantilla    | Qué obtienes                                                  | `webDir` |
| ------------ | ------------------------------------------------------------- | -------- |
| `basic`      | HTML, CSS y JavaScript. No necesita paso de build ni Node.js. | `www`    |
| `vite`       | Un proyecto Vite con recarga en caliente.                     | `dist`   |
| `react`      | Vite con React 18 y JSX.                                      | `dist`   |
| `vue`        | Vite con componentes de un solo archivo de Vue 3.             | `dist`   |
| `svelte`     | Vite con Svelte 5.                                            | `dist`   |
| `vanilla-ts` | Vite con TypeScript estricto y sin framework.                 | `dist`   |

Todas las plantillas salvo `basic` necesitan Node.js. Ejecuta `npm install` una vez y usa los scripts que escribe `create`: `npm run dev` (recarga en caliente en el navegador), `npm run android` (compila y luego `koru run`) y `npm run apk` (APK release firmado). Cada una incluye una copia del módulo JavaScript de koru en `src/koru/`; renuévala con `koru js src/koru --force` tras actualizar koru. Las plantillas de framework se compilaron con `npm run build`, `koru build` y `koru verify` (y `tsc --noEmit` en `vanilla-ts`); no se han ejecutado en Android.

En una terminal, los valores que faltan se preguntan; pulsa Intro para aceptar la sugerencia. En scripts y CI (sin terminal), se usan directamente las sugerencias y el directorio es obligatorio. `create` rechaza un directorio que ya tenga archivos. Al terminar muestra los siguientes comandos. Con `react`:

```text
Created mi-app (com.example.miapp, react template)

Next steps:
  cd mi-app
  npm install
  npm run dev        # browser preview with hot reload
  npm run android    # install and open on a device

Back up mi-app/.koru/signing.pem: Android only accepts updates signed with the same key.
```

## `dev`: vista previa en vivo

```sh
koru dev
koru dev --host 0.0.0.0
koru dev --host 0.0.0.0 --qr
koru dev --proxy http://localhost:5173
```

| Opción    | Valor por defecto | Significado                                                                                        |
| --------- | ----------------- | -------------------------------------------------------------------------------------------------- |
| `--port`  | `8100`            | Puerto de escucha; `0` elige uno libre.                                                            |
| `--host`  | `127.0.0.1`       | `0.0.0.0` lo hace accesible desde otros dispositivos de la red.                                    |
| `--proxy` | ninguno           | Reenvía al servidor de desarrollo de tu framework y añade la API de koru.                          |
| `--qr`    | desactivado       | Imprime un código QR de la URL de red. Aparece solo en una terminal cuando el host no es loopback. |

Detenlo con Ctrl+C. Consulta [Desarrolla con recarga en vivo]({{base}}docs/es/develop/).

## `run` y `devices`

```sh
koru devices
koru run
koru run --device emulator-5554 --release
koru run --logs
```

| Opción       | Valor por defecto                               | Significado                                           |
| ------------ | ----------------------------------------------- | ----------------------------------------------------- |
| `--device`   | El único dispositivo listo, o `$ANDROID_SERIAL` | Número de serie de `koru devices`.                    |
| `--release`  | desactivado                                     | Instala una build release (sin depuración).           |
| `--logs`     | desactivado                                     | Tras abrir la app, muestra sus logs como `koru logs`. |
| `--no-cache` | desactivado                                     | Regenera el DEX del runtime.                          |

`run`, `logs` y `devices` hablan con el servidor adb en `127.0.0.1:5037` (o en `ANDROID_ADB_SERVER_PORT` / `ADB_SERVER_SOCKET=tcp:host:puerto`). koru no lo inicia; ejecuta `adb start-server` o abre Android Studio. Consulta [Ejecuta en un dispositivo]({{base}}docs/es/run-on-device/).

## `logs`: los logs de la app en tu terminal

```sh
koru logs
koru logs ./my-app --device emulator-5554
```

Muestra lo que registra la app en ejecución: su salida de logcat y las líneas `chromium` que llevan `console.log`, `console.warn` y `console.error` del WebView. La app debe estar abierta (consulta `koru run`). Pulsa Ctrl+C para detenerlo. `logs` acepta el mismo `--device` y `$ANDROID_SERIAL` que `run`. Si la app no está en ejecución, falla con “is not running on …: open it with `koru run` first”.

koru lo hace con el protocolo de red del servidor adb, ejecutando `pidof` y `logcat` en el dispositivo; no ejecuta `adb` ni `logcat` en tu ordenador. **Todavía no se ha validado en un dispositivo**: los comandos y el formato de línea siguen la documentación de Android y solo se han probado contra un servidor adb simulado. Si la salida no se ve bien en tu dispositivo, usa `chrome://inspect`.

## `init`: añade koru a un directorio

```sh
cd tu-app-web
koru init
```

```text
Initialized .; persistent signing identity: .koru/signing.pem
```

Escribe `koru.json`, una página de ejemplo `www/index.html` y `.koru/signing.pem`, y añade `.koru/` a `.gitignore` (lo crea si no existe y, si existe, añade la línea al final). No reemplaza un `koru.json` existente. Edita `webDir` para que apunte a la salida de tu build. Para una app nueva, `create` te da un punto de partida más completo.

## `js`: el módulo para bundlers

```sh
koru js src/koru
koru js src/koru --force   # tras actualizar koru
```

Escribe `core.js`, `core.mjs` y `core.d.mts`. Importa `./koru/core.mjs` desde tu código; consulta [Frameworks]({{base}}docs/es/existing-project/).

## `doctor`: revisa antes de empaquetar

```sh
koru doctor ./my-app
```

```text
Build and verification use the self-contained Go implementation; no Java, SDK, or external build tools are required.
Devices 1 ready of 1 connected (adb server 127.0.0.1:5037); see koru devices
Project com.example.koru: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

El `appId` viene de tu configuración. `doctor` también indica si hay un servidor adb accesible y cuántos dispositivos están listos; esa línea es informativa, porque solo `koru run` necesita un dispositivo. Sin proyecto en el directorio actual, solo muestra las notas del entorno. `doctor` no construye, instala ni ejecuta el APK.

Cuando detecta algo que probablemente te dará problemas, añade líneas `Warning`. Los avisos nunca hacen fallar la comprobación; una configuración inválida sí. Esta salida es de un proyecto con varios problemas, y la línea del servidor adb depende de tu equipo:

```text
Project com.example.myapp: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
Warning .koru/ is not listed in .gitignore; it holds your signing key. Add a line ".koru/" to .gitignore and keep a private backup of the key.
Warning index.html has no <meta name="viewport" content="width=device-width,initial-scale=1">; the page renders at desktop width on phones. Add it inside <head>.
Warning api.js calls fetch() with an https:// URL but android.permission.INTERNET is not in permissions (heuristic text match, not a full analysis). Add it to "permissions" in koru.json if the app needs the network.
Warning icon icon.png is 64x48; use a square PNG or Android will distort it.
Warning icon icon.png is 64x48; use at least 192x192 px so it stays sharp on high-density screens.
Note recorded Android execution covers API 30 (targetSdk here is 36); test on your devices before shipping.
```

| Aviso        | Se muestra cuando                                                                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.gitignore` | El proyecto tiene una carpeta `.git` o un `.gitignore` que no incluye `.koru/`, donde está la clave.                                                                      |
| viewport     | `index.html` no tiene `<meta name="viewport">`.                                                                                                                           |
| `INTERNET`   | Un script o HTML llama a `fetch()` con una URL `https://` literal y falta el permiso. Es una coincidencia de texto: puede pasar por alto llamadas o marcar código muerto. |
| icono        | `icon` no es cuadrado o mide menos de 192 × 192 px. Un icono que falta o que no es PNG es un error.                                                                       |

La `Note` final aparece cuando `targetSdk` no es 30, lo que incluye los proyectos nuevos (por defecto es 36): las ejecuciones registradas cubren solo Android 11 (API 30), así que prueba en los dispositivos que quieras soportar. Consulta [Compatibilidad probada]({{base}}docs/es/android-results/).

## `build`: debug, release y destinos

```sh
koru build ./my-app --output app-debug.apk
koru build ./my-app --release --output app.apk
```

| Flag         | Predeterminado        | Función                                                                                                     |
| ------------ | --------------------- | ----------------------------------------------------------------------------------------------------------- |
| `--output`   | `app.apk`             | Destino del APK firmado, relativo al directorio actual. Debe terminar en `.apk` y quedar fuera de `webDir`. |
| `--release`  | apagado               | Deshabilita debug y WebView DevTools. Usa la identidad configurada.                                         |
| `--keystore` | `signingKey` del JSON | PEM existente alternativo; rutas relativas desde el proyecto.                                               |
| `--no-cache` | apagado               | Regenera DEX del runtime aunque exista una caché válida.                                                    |
| `--bump`     | apagado               | Sube `versionCode` y la versión patch de `koru.json` antes de compilar. Consulta `bump` más abajo.          |

Salida típica:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

Tamaño y huella SHA-256 completa de 64 caracteres varían; los puntos acortan el ejemplo. Un cache hit significa que se reutilizó DEX, no que se omitieron el APK o los assets web. Una caché corrupta devuelve error. El destino no puede reemplazar la clave de firma.

`--release` **no** crea una clave. Debug y release usan la identidad configurada salvo que `--keystore` seleccione otra existente. Cambiar de clave impide actualizar apps instaladas con el certificado anterior.

## `bump`: sube la versión

```sh
koru bump              # patch: 1.0.0 -> 1.0.1
koru bump minor        # 1.0.1 -> 1.1.0
koru bump major ./my-app
koru bump code         # solo versionCode
koru build --release --bump --output app-release.apk
```

```text
Bumped version 1.0.0 -> 1.0.1, versionCode 1 -> 2
Bumped version 1.0.1 -> 1.1.0, versionCode 2 -> 3
Bumped version 1.1.0 -> 1.1.0, versionCode 3 -> 4
Bumped version 1.1.0 -> 2.0.0, versionCode 4 -> 5
```

`versionCode` siempre sube una unidad, que es lo que compara Android al instalar una actualización. El nivel (por defecto `patch`) cambia además el `X.Y.Z` de `version`; `code` no toca `version`. Si `version` no son tres números separados por puntos, como `1.0`, solo sube `versionCode` y un aviso te pide editar `version` a mano. `bump` se niega a superar el límite de Android, 2.100.000.000.

`koru build --bump` hace un bump de patch y después compila, así que el APK lleva los números nuevos; la compilación imprime antes la misma línea `Bumped …`. `bump` reescribe `koru.json` con sangría de dos espacios y el orden de campos de koru; si te importa el formato, reordénalo a mano después. Consulta [Distribuye tu app]({{base}}docs/es/distribution/).

## `plugin` y `config`: edita e inspecciona `koru.json`

```sh
koru plugin list
koru plugin add vibration notifications
koru plugin remove share --project ./my-app
koru config
```

`plugin list` muestra los seis plugins integrados, si cada uno está activo y el permiso de Android que añade. Esta es la lista de un proyecto creado con `basic`:

```text
  on   device         -
  on   storage        -
  on   vibration      android.permission.VIBRATE
  on   app            -
  on   share          -
  off  notifications  android.permission.POST_NOTIFICATIONS
```

`plugin add <id>...` y `plugin remove <id>...` editan `plugins` en `koru.json` e imprimen `Enabled …` o `Disabled …`. `list` recibe el proyecto como argumento; `add` y `remove` lo reciben con `--project` (por defecto `.`). Un ID desconocido es un error que enumera los válidos. Los plugins DEX personalizados siguen yendo en `pluginManifests`; consulta [JavaScript y plugins]({{base}}docs/es/plugins/).

`config` imprime lo que usará koru, como JSON: el `config` leído de `koru.json`, `effectivePermissions` (los que añaden tus plugins, como `android.permission.VIBRATE`), el `webRoot` absoluto y la ruta absoluta de `signingKey`. Valida antes de imprimir, así que sirve también para comprobar tus cambios:

```text
koru: koru.json: invalid orientation "upside": use "portrait", "landscape" or "sensor", or remove the field to keep Android's default
```

## `deeplinks`: verifica tus enlaces `https`

```sh
koru deeplinks
koru deeplinks ./my-app --json > assetlinks.json
```

Imprime el archivo Digital Asset Links (`assetlinks.json`) para las entradas `https` de `deepLinks` en `koru.json`. Con `--json` imprime solo el JSON, listo para guardar; `--keystore` usa otra identidad PEM existente en lugar de `signingKey`. Sin entradas `https` falla y lo indica. El flujo completo está en [Configura tu proyecto]({{base}}docs/es/configuration/).

## `verify`: reporte estructurado

```sh
koru verify app.apk
```

```json
{
  "appId": "com.example.koru",
  "entries": 7,
  "bytes": 16785,
  "certificateSha256": "3ecd2de568b127fef17621e1cce54a68a0587c4b3de22ebcd06224bf5f052122",
  "signatureScheme": "APK v2 RSA/SHA-256"
}
```

Cuando `koru.json` tiene `deepLinks`, el reporte también lista los que hay en el manifiesto, como `"deepLinks": ["https://example.com/app", "myapp://"]`. Se ha aplicado formato al ejemplo; la CLI emite una sola línea JSON. `appId` es el paquete del manifiesto; `entries` cuenta miembros ZIP; `bytes` es la longitud del APK; `certificateSha256` identifica el certificado; `signatureScheme` nombra la firma validada. Los valores de ejemplo vienen de un build de prueba y cambiarán. La verificación revisa firma, ZIP, DEX, manifiesto, recursos e icono. No instala la app ni prueba ejecución ART/WebView. Límite de entrada APK: 320 MiB.

Para extraer un campo desde otro programa, procesa el JSON; no combines la salida humana de `build` con la de `verify`.

## `keygen` y `version`

```sh
koru keygen --output ./my-app/alternate.pem
koru version
```

`keygen` crea una identidad PEM nueva en la ruta indicada y no sobrescribe archivos. No actualiza `koru.json`. Configura `signingKey` o pasa `--keystore` explícitamente para usarla; no cambies identidades al actualizar una app instalada.

## Comandos DEX independientes

Los subcomandos DEX operan sobre la [representación intermedia JSON de DEX]({{base}}docs/es/dex-api/) por separado del builder APK.

```sh
koru dex build examples/dex/answer.json --output answer.dex
koru dex inspect answer.dex
```

```text
Wrote answer.dex (476 bytes); DEX 035 verified successfully.
```

`dex build` requiere un archivo IR JSON. `--output` usa `classes.dex` por defecto y valida el DEX producido con el parser independiente. `dex inspect` toma un archivo e imprime JSON indentado con tamaño, strings, tipos, prototipos, campos, métodos y clases. Límite de entrada DEX: 64 MiB. Validación estructural no equivale a verificación de tipos de ART.

```sh
koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` no acepta archivo IR. Produce el experimento histórico de Activity con `file://` para comparación; el APK normal usa HTTPS interceptado.

[Primer APK]({{base}}docs/es/getting-started/) · [Problemas frecuentes]({{base}}docs/es/troubleshooting/) · [API DEX]({{base}}docs/es/dex-api/)
