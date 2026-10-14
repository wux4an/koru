# Crea tu primera app Android

Esta guía empieza en un checkout limpio y termina con un APK firmado ejecutándose en Android. Los comandos se ejecutan desde la raíz del repositorio. Koru necesita Go **1.23 o posterior** para compilar la CLI una vez; el binario resultante no necesita Go, JDK, SDK Android ni red para construir APK.

## 1. Obtén la CLI

```sh
git clone https://github.com/wux4an/koru.git
cd koru
go version
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

El último comando muestra la versión y plataforma, por ejemplo:

```text
koru 0.2.0-dev (linux/amd64)
```

Si falta `go`, instala primero la toolchain oficial. En el workspace de desarrollo registrado, `/usr/bin/go` es un juego; comprueba `command -v go` y `go version`. El binario `bin/koru` es la única herramienta necesaria para el build normal de APK. Cobra se descarga al compilar la CLI por primera vez; no se necesita para ejecutar el binario.

## 2. Crea un proyecto

```sh
./bin/koru init ./my-app
```

Salida esperada:

```text
Initialized ./my-app; persistent signing identity: .koru/signing.pem
```

Se crea esta estructura:

```text
my-app/
├── koru.json             configuración
├── www/
│   └── index.html        página de ejemplo
└── .koru/
    └── signing.pem       clave privada y certificado
```

`init` no reemplaza un `koru.json` existente. La clave se crea una sola vez con modo `0600`. Guárdala fuera de la carpeta web y del control de versiones: Android exige el mismo certificado en cada actualización.

Abre `my-app/www/index.html` en un navegador para ver el ejemplo. Sustitúyelo por tu HTML, CSS, JavaScript, fuentes e imágenes. Los archivos referenciados deben quedar dentro de `webDir`; la app empaquetada no carga scripts e iframes externos.

## 3. Define la identidad antes de distribuir

`my-app/koru.json` comienza con `appId: "com.example.koru"`, `name: "Koru"`, `version: "1.0.0"`, `versionCode: 1` y `webDir: "www"`. Cambia identificador y nombre antes de la primera instalación. Android considera un `appId` diferente como otra app; las actualizaciones deben conservar identificador y clave.

```json
{
  "appId": "com.example.hello",
  "name": "Hola Android",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "www",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app"],
  "signingKey": ".koru/signing.pem"
}
```

Es una configuración completa. `versionCode` debe subir en cada actualización distribuida. Se acepta el rango SDK 26–36; la ejecución registrada es Android 11/API 30, no todos los niveles. [Todos los campos y reglas]({{base}}es/configuration/) se explican aparte.

## 4. Comprueba, compila y verifica

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
./bin/koru build ./my-app --release --output app.apk
./bin/koru verify app.apk
```

`doctor` revisa configuración, clave, entrada web y manifiestos de plugins. Una comprobación correcta termina así:

```text
Project com.example.hello: configuration, persistent signing identity, web entry point, and plugin manifests are valid.
```

El primer `build` genera un APK debug con inspección WebView disponible. El segundo genera release y desactiva debug y DevTools. Ambos usan la identidad de firma configurada; ninguno crea una clave nueva en silencio. La ruta de salida es relativa al directorio actual de la terminal, debe terminar en `.apk` y quedar fuera de `webDir`. `app.apk` en la raíz del repositorio cumple esto.

```text
Runtime cache hit: false
Wrote app.apk (16785 bytes), APK v2 RSA/SHA-256, certificate 3ecd2de5…
```

El tamaño y la huella varían con tus archivos y clave. El segundo build puede decir `Runtime cache hit: true` porque reutiliza DEX del runtime; el APK final sigue usando tus assets y el modo actual.

`verify` imprime JSON en stdout. Sus campos son `appId`, `entries`, `bytes`, `certificateSha256` y `signatureScheme`. Que la firma y estructura pasen no demuestra ejecución en Android.

## 5. Instala y abre en Android

Solo para este paso necesitas un dispositivo o emulador y `adb`. Es una herramienta de instalación/prueba, no dependencia del build. Android 8/API 26 es el mínimo configurado; prueba tus dispositivos objetivo.

```sh
adb install -r app.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Usa tu `appId` exacto antes de la barra. La Activity del runtime es `com.koru.runtime.MainActivity`. Comprueba que el WebView muestra tu página y prueba navegación, almacenamiento y métodos nativos que uses. La [matriz Android registrada]({{base}}es/android-results/) cubre API 30/x86_64; no valida tu app en todos los equipos.

## 6. Actualiza la app

Edita tus archivos web, incrementa `versionCode` (y cambia `version` si quieres otra etiqueta visible), y recompila con **el mismo** `appId` y `signingKey`:

```sh
./bin/koru build ./my-app --release --output app-v2.apk
./bin/koru verify app-v2.apk
adb install -r app-v2.apk
```

No borres ni reemplaces `my-app/.koru/signing.pem`. Si Android muestra `INSTALL_FAILED_UPDATE_INCOMPATIBLE`, la app instalada y el APK nuevo tienen certificados distintos o cambió su identidad. Recupera la clave original; no generes otra en silencio.

## Usa un proyecto web existente

Compila React, Vue, Svelte, Parcel u otra herramienta a una **carpeta estática**. Configura `webDir` para esa carpeta y asegúrate de que contiene `index.html` regular y rutas locales para scripts, estilos y assets. Para `dist/` junto a `koru.json`:

```sh
npm --prefix ./my-app run build
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output app.apk
```

Pon `"webDir": "dist"` en `my-app/koru.json` si la entrada generada está en `my-app/dist/index.html`. El framework y sus dependencias Node se ejecutan antes de Koru; no se empaquetan como herramientas de build de Koru. Las rutas de cliente pueden usar `"spa": true`; archivos desconocidos siguen devolviendo 404. La [guía de configuración]({{base}}es/configuration/) explica salidas de frameworks, URLs y permisos de red.

Si un comando falla, la [guía de problemas]({{base}}es/troubleshooting/) relaciona error, causa y solución. La [referencia CLI]({{base}}es/cli/) enumera flags y salidas procesables.
