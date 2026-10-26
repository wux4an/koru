# Crea tu primera app

En unos cinco minutos vas a crear una app desde una plantilla, previsualizarla en el navegador con recarga en vivo y abrirla en un móvil o emulador Android. Necesitas **Go 1.23 o posterior** para instalar koru. Para ejecutar la app en Android necesitas además un móvil con depuración USB o un emulador.

## 1. Instala koru

```sh
git clone https://github.com/wux4an/koru.git
cd koru
make install
koru version
```

`make install` compila un binario autónomo y lo copia en `~/.local/bin`. Si `koru version` responde «command not found», añade esa carpeta a tu `PATH` (por ejemplo, `export PATH="$HOME/.local/bin:$PATH"` en el perfil de tu shell). Sin `make`, ejecuta `CGO_ENABLED=0 go build -trimpath -o ~/.local/bin/koru ./cmd/koru`.

A partir de aquí solo necesitas ese binario: genera APKs sin Go, Java, Gradle, el SDK de Android ni conexión a Internet.

## 2. Crea un proyecto

```sh
koru create mi-app
```

En una terminal, koru hace tres preguntas. Pulsa Intro para aceptar cada sugerencia:

```text
? App name: (Mi App)
? Application id: (com.example.miapp)
  1. basic  HTML, CSS and JavaScript, no build step or Node.js needed
  2. vite   Vite project with hot reload (needs Node.js)
? Template: (basic)
```

También puedes responderlas con opciones, algo útil en scripts:

```sh
koru create mi-app --name "Mi App" --app-id com.example.miapp --template basic
```

La plantilla `basic` contiene:

```text
mi-app/
├── koru.json          nombre, id, versión y plugins de la app
├── www/               los archivos que se empaquetan en el APK
│   ├── index.html
│   ├── app.js
│   └── style.css
├── types/koru.d.mts   tipos de la API JavaScript
├── jsconfig.json      permite que tu editor use esos tipos
├── .gitignore         deja .koru/ fuera de Git
└── .koru/signing.pem  tu clave privada de firma
```

> **Haz una copia de seguridad de `.koru/signing.pem`.** Android solo instala actualizaciones firmadas con la misma clave. El `.gitignore` generado la mantiene fuera del control de versiones.

## 3. Previsualiza con recarga en vivo

```sh
cd mi-app
koru dev
```

Abre la dirección que muestra, `http://127.0.0.1:8100`. Edita `www/index.html` o `www/app.js` y guarda: la página se recarga sola. La API JavaScript también funciona aquí, con implementaciones web: el contador se guarda con `storage` y **Share** usa el menú de compartir del navegador. La etiqueta muestra `web`; dentro del APK muestra `android`. En [Desarrolla con recarga en vivo]({{base}}es/develop/) verás cómo probar en el navegador del móvil o usar otro servidor de desarrollo.

## 4. Ejecútala en Android

Conecta un móvil con la [depuración USB](https://developer.android.com/studio/debug/dev-options) activada, o inicia un emulador desde el Device Manager de Android Studio. Después:

```sh
koru run
```

```text
Built com.example.miapp 1.0.0 (20 KB) in 35ms
Installed on sdk gphone64 x86 64 (emulator-5554)
Opened Mi App on sdk gphone64 x86 64 (emulator-5554); inspect it at chrome://inspect
```

`koru run` compila un APK de depuración, lo instala y lo abre. Se comunica con el servidor adb que ejecutan Android Studio y las platform tools; si no hay ninguno activo, el error te dice cómo iniciarlo. Con la app abierta, visita `chrome://inspect` en Chrome en tu ordenador para ver su consola y su DOM. Más detalles en [Ejecuta en un dispositivo]({{base}}es/run-on-device/).

Cambia algo y vuelve a ejecutar `koru run`. Ese es el ciclo completo: **editar → `koru dev` → `koru run`**.

## 5. Genera una versión release

```sh
koru build --release --output mi-app.apk
koru verify mi-app.apk
```

Una build release desactiva la depuración del WebView y firma con tu clave existente. Para publicar una actualización, conserva el mismo `appId` y la misma clave, e incrementa `versionCode` en `koru.json`. Consulta [Distribuye tu app]({{base}}es/distribution/).

> koru es experimental. La ejecución registrada cubre Android 11 (API 30) en un emulador x86_64; el rango configurable es API 26–36. Prueba en los dispositivos que quieras soportar.

## Siguientes pasos

- [Usa React, Vue, Svelte u otro framework]({{base}}es/existing-project/), o empieza con `koru create mi-app --template vite`.
- [Usa funciones de Android desde JavaScript]({{base}}es/plugins/), con tipos TypeScript.
- [Consulta todos los comandos]({{base}}es/cli/) y [todas las opciones de `koru.json`]({{base}}es/configuration/).

Si un paso falla, ejecuta `koru doctor` y consulta [Solución de problemas]({{base}}es/troubleshooting/).
