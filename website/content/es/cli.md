# Comandos CLI

Ejecuta `koru --help` para ver la lista de comandos y `koru <comando> --help` para las opciones y ejemplos de un comando. La ayuda está en inglés. Los errores dicen qué está mal y cómo corregirlo, y terminan con un código de salida distinto de cero. Define `NO_COLOR=1` para obtener la salida sin colores.

Los comandos que aceptan `[project]` usan por defecto el directorio actual, así que dentro de un proyecto basta con `koru dev`, `koru run` o `koru build`.

## Comandos de un vistazo

| Comando                         | Qué hace                                                                                                          |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `create [directorio]`           | Crea una app desde una plantilla con `koru.json`, `.gitignore`, tipos y clave de firma. Pregunta en una terminal. |
| `dev [proyecto]`                | Sirve la app en un navegador con recarga en vivo e implementaciones web de la API JavaScript.                     |
| `run [proyecto]`                | Compila un APK de depuración, lo instala en un dispositivo o emulador conectado y lo abre.                        |
| `devices`                       | Lista los dispositivos y emuladores que ve el servidor adb, con su estado.                                        |
| `build [proyecto]`              | Compila y firma un APK. Depuración por defecto; `--release` para distribuir.                                      |
| `verify <app.apk>`              | Comprueba la firma y la estructura de un APK e imprime un informe JSON.                                           |
| `doctor [proyecto]`             | Revisa la configuración, la clave, la entrada web, los plugins y si hay un servidor adb accesible.                |
| `init [directorio]`             | Añade a un directorio un `koru.json` mínimo, una página de ejemplo y una clave, para proyectos existentes.        |
| `js [directorio]`               | Copia el módulo JavaScript y sus tipos a tus fuentes, para bundlers. Por defecto `src/koru`.                      |
| `keygen --output <archivo.pem>` | Crea una identidad de firma nueva. No sobrescribe.                                                                |
| `dex build` / `dex inspect`     | Compila o inspecciona archivos DEX independientes.                                                                |
| `version`                       | Muestra la versión y la plataforma.                                                                               |

Un binario compilado no necesita Go, Java, el SDK de Android ni red para compilar apps. `dev`, `run` y `devices` son ayudas para desarrollar; no forman parte de la compilación del APK.

## `create`: parte de una plantilla

```sh
koru create mi-app
koru create mi-app --template vite --name "Mi App" --app-id com.example.miapp
```

| Opción             | Valor por defecto               | Significado                                                       |
| ------------------ | ------------------------------- | ----------------------------------------------------------------- |
| `-t`, `--template` | `basic`                         | `basic` (HTML, CSS y JS, sin paso de build) o `vite`.             |
| `--name`           | Del directorio: `Mi App`        | Nombre que aparece bajo el icono.                                 |
| `--app-id`         | Del nombre: `com.example.miapp` | Identificador de la app Android. Elige el tuyo antes de publicar. |

En una terminal, los valores que faltan se preguntan; pulsa Intro para aceptar la sugerencia. En scripts y CI (sin terminal), se usan directamente las sugerencias y el directorio es obligatorio. `create` rechaza un directorio que ya tenga archivos. Al terminar muestra los siguientes comandos.

## `dev`: vista previa en vivo

```sh
koru dev
koru dev --host 0.0.0.0
koru dev --proxy http://localhost:5173
```

| Opción    | Valor por defecto | Significado                                                               |
| --------- | ----------------- | ------------------------------------------------------------------------- |
| `--port`  | `8100`            | Puerto de escucha; `0` elige uno libre.                                   |
| `--host`  | `127.0.0.1`       | `0.0.0.0` lo hace accesible desde otros dispositivos de la red.           |
| `--proxy` | ninguno           | Reenvía al servidor de desarrollo de tu framework y añade la API de koru. |

Detenlo con Ctrl+C. Consulta [Desarrolla con recarga en vivo]({{base}}es/develop/).

## `run` y `devices`

```sh
koru devices
koru run
koru run --device emulator-5554 --release
```

| Opción       | Valor por defecto                               | Significado                                 |
| ------------ | ----------------------------------------------- | ------------------------------------------- |
| `--device`   | El único dispositivo listo, o `$ANDROID_SERIAL` | Número de serie de `koru devices`.          |
| `--release`  | desactivado                                     | Instala una build release (sin depuración). |
| `--no-cache` | desactivado                                     | Regenera el DEX del runtime.                |

Ambos hablan con el servidor adb en `127.0.0.1:5037` (o en `ANDROID_ADB_SERVER_PORT` / `ADB_SERVER_SOCKET=tcp:host:puerto`). koru no lo inicia; ejecuta `adb start-server` o abre Android Studio. Consulta [Ejecuta en un dispositivo]({{base}}es/run-on-device/).

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

Escribe `core.js`, `core.mjs` y `core.d.mts`. Importa `./koru/core.mjs` desde tu código; consulta [Frameworks]({{base}}es/existing-project/).

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

Salida típica:

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

Tamaño y huella SHA-256 completa de 64 caracteres varían; los puntos acortan el ejemplo. Un cache hit significa que se reutilizó DEX, no que se omitieron el APK o los assets web. Una caché corrupta devuelve error. El destino no puede reemplazar la clave de firma.

`--release` **no** crea una clave. Debug y release usan la identidad configurada salvo que `--keystore` seleccione otra existente. Cambiar de clave impide actualizar apps instaladas con el certificado anterior.

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

Se ha aplicado formato al ejemplo; la CLI emite una sola línea JSON. `appId` es el paquete del manifiesto; `entries` cuenta miembros ZIP; `bytes` es la longitud del APK; `certificateSha256` identifica el certificado; `signatureScheme` nombra la firma validada. Los valores de ejemplo vienen de un build de prueba y cambiarán. La verificación revisa firma, ZIP, DEX, manifiesto, recursos e icono. No instala la app ni prueba ejecución ART/WebView. Límite de entrada APK: 320 MiB.

Para extraer un campo desde otro programa, procesa el JSON; no combines la salida humana de `build` con la de `verify`.

## `keygen` y `version`

```sh
koru keygen --output ./my-app/alternate.pem
koru version
```

`keygen` crea una identidad PEM nueva en la ruta indicada y no sobrescribe archivos. No actualiza `koru.json`. Configura `signingKey` o pasa `--keystore` explícitamente para usarla; no cambies identidades al actualizar una app instalada.

## Comandos DEX independientes

Los subcomandos DEX operan sobre la [representación intermedia JSON de DEX]({{base}}es/dex-api/) por separado del builder APK.

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

[Primer APK]({{base}}es/getting-started/) · [Problemas frecuentes]({{base}}es/troubleshooting/) · [API DEX]({{base}}es/dex-api/)
