# Comandos CLI

Ejecuta `./bin/koru --help` para ver los comandos de tu versión y `./bin/koru build --help` para los flags de un comando. Los ejemplos suponen que el binario está en `bin/koru` y que la terminal está en la raíz del repositorio. Si el ejecutable está en `PATH`, sustituye `./bin/koru` por `koru`.

La ayuda está en inglés. Los errores indican el problema y devuelven un código de salida distinto de cero. Usa `NO_COLOR=1` si prefieres salida sin color.

## Comandos de proyecto

| Comando                      | Entrada                               | Resultado                                                                                           |
| ---------------------------- | ------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `init [directorio]`          | Directorio opcional; por defecto `.`. | Crea `koru.json`, `www/index.html` y `.koru/signing.pem`. No reemplaza una configuración existente. |
| `doctor [proyecto]`          | Directorio opcional; por defecto `.`. | Valida configuración, clave de firma, entrada web y manifiestos de plugins.                         |
| `build [proyecto]`           | Directorio opcional; por defecto `.`. | Genera y firma un APK. Debug por defecto; salida `app.apk` en el directorio actual de la terminal.  |
| `verify <app.apk>`           | Una ruta de APK.                      | Valida firma y estructura e imprime JSON.                                                           |
| `keygen --output <file.pem>` | Archivo nuevo.                        | Crea explícitamente una identidad; no sobrescribe. `-o` equivale a `--output`.                      |
| `version`                    | Sin argumentos.                       | Muestra versión y plataforma.                                                                       |

Todos aceptan `--help` o `-h`. El binario compilado construye apps sin Go, Java, herramientas de build del SDK Android ni red.

## `init`: archivos que puedes editar

```sh
./bin/koru init ./my-app
```

```text
Initialized ./my-app; persistent signing identity: .koru/signing.pem
```

`my-app/koru.json` contiene identidad, SDK, `webDir`, plugins y ruta de firma. La página de muestra está en `my-app/www/index.html`. La clave es privada: deja `my-app/.koru/signing.pem` fuera de Git y de la carpeta web. `init` crea el directorio si falta, pero no sobrescribe la configuración.

## `doctor`: revisa antes de empaquetar

```sh
./bin/koru doctor ./my-app
```

```text
Build and verification use the self-contained Go implementation; no Java, SDK, or external build tools are required.
Project com.example.koru: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

El `appId` viene de tu configuración. Si llamas `doctor` sin proyecto desde una carpeta sin `koru.json`, imprime una nota general del entorno; pasa la ruta del proyecto para revisarlo. `doctor` no construye, instala ni ejecuta el APK.

## `build`: debug, release y destinos

```sh
./bin/koru build ./my-app --output app-debug.apk
./bin/koru build ./my-app --release --output app.apk
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
./bin/koru verify app.apk
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
./bin/koru keygen --output ./my-app/alternate.pem
./bin/koru version
```

`keygen` crea una identidad PEM nueva en la ruta indicada y no sobrescribe archivos. No actualiza `koru.json`. Configura `signingKey` o pasa `--keystore` explícitamente para usarla; no cambies identidades al actualizar una app instalada.

## Comandos DEX independientes

Los subcomandos DEX operan sobre la [representación intermedia JSON de DEX]({{base}}es/dex-api/) por separado del builder APK.

```sh
./bin/koru dex build examples/dex/answer.json --output answer.dex
./bin/koru dex inspect answer.dex
```

```text
Wrote answer.dex (476 bytes); DEX 035 verified successfully.
```

`dex build` requiere un archivo IR JSON. `--output` usa `classes.dex` por defecto y valida el DEX producido con el parser independiente. `dex inspect` toma un archivo e imprime JSON indentado con tamaño, strings, tipos, prototipos, campos, métodos y clases. Límite de entrada DEX: 64 MiB. Validación estructural no equivale a verificación de tipos de ART.

```sh
./bin/koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` no acepta archivo IR. Produce el experimento histórico de Activity con `file://` para comparación; el APK normal usa HTTPS interceptado.

[Primer APK]({{base}}es/getting-started/) · [Problemas frecuentes]({{base}}es/troubleshooting/) · [API DEX]({{base}}es/dex-api/)
