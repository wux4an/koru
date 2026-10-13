# Crea tu primera app Android

Empieza con una carpeta que contenga `index.html` y sus assets locales. También puedes usar la salida estática de React, Vue, Svelte u otro framework; genera esa salida antes de ejecutar koru.

## Compila el binario de koru

Usa un compilador Go real, versión 1.23 o posterior, desde la raíz del repositorio:

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

Algunas sesiones registradas colocan Go en `/workspace/tooling/go/bin`. Esa ruta depende del entorno. Comprueba `go version`; en este workspace `/usr/bin/go` es un juego, no el compilador. Una vez compilado, koru no necesita Go ni red para construir aplicaciones.

## Inicia un proyecto

```sh
./bin/koru init ./mi-app
```

Se crean `koru.json`, un ejemplo `www/index.html` y una identidad de firma persistente en `.koru/signing.pem`, con permisos `0600`. Guarda esa identidad: las actualizaciones necesitan el mismo certificado. `init` rechaza sobrescribir una configuración existente.

## Usa tus archivos web

Pon los archivos en `mi-app/www`, o cambia `webDir` en `mi-app/koru.json` por la carpeta de salida de tu framework. Debe contener un `index.html` regular. Empaqueta los scripts de la aplicación localmente. No pongas claves privadas entre los assets.

[Consulta todos los campos de configuración]({{base}}es/configuration/)

## Compila y verifica

```sh
./bin/koru doctor ./mi-app
./bin/koru build ./mi-app --output app-debug.apk
./bin/koru build ./mi-app --release --output app.apk
./bin/koru verify app.apk
```

Escribe el APK fuera de `webDir`. El build debug permite inspeccionar el WebView; `--release` deshabilita debug y DevTools y reutiliza la clave configurada. Una clave ausente o inválida detiene el build; koru nunca crea una sustituta silenciosamente.

`verify` comprueba firma y estructura. No ejecuta la app ni demuestra compatibilidad con dispositivos.

## Ejecuta en Android

Instala el APK en un dispositivo o emulador con Android 8/API 26 o posterior. ADB es una herramienta opcional para pruebas de ejecución, no una dependencia del build:

```sh
adb install -r app.apk
adb shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
```

Sustituye `com.example.koru` por tu `appId`. La matriz registrada es Android 11/API 30/x86_64. Prueba tu app en las versiones y dispositivos que quieras soportar.

[Reproduce las pruebas Android]({{base}}es/testing/) · [Consulta la evidencia]({{base}}es/android-results/)
