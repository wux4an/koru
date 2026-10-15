# Crea tu primera app

Crea un proyecto, edita su página y ábrelo en Android. Necesitas Go 1.23 o posterior para obtener koru desde el código fuente y un dispositivo o emulador Android para probar tu app.

## Instala koru

Clona el repositorio y compila el ejecutable:

```sh
git clone https://github.com/wux4an/koru.git
cd koru
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

La primera compilación puede descargar dependencias de Go. Una vez creado, `bin/koru` funciona por sí solo y genera APK sin Go, Java, Gradle ni conexión a Internet.

Los ejemplos siguientes se ejecutan desde esta carpeta. Si añades el ejecutable a tu `PATH`, puedes usar `koru` en lugar de `./bin/koru`.

## Crea un proyecto

```sh
./bin/koru init ./my-app
```

Encontrarás estos archivos:

```text
my-app/
├── koru.json
├── www/
│   └── index.html
└── .koru/
    └── signing.pem
```

`www` contiene tu web. `koru.json` configura la app. `signing.pem` es su identidad de firma: guarda una copia segura y excluye `.koru/` de Git. Necesitarás esa misma clave para actualizar la app.

Abre `my-app/www/index.html` y cambia el título o añade tu contenido. Puedes previsualizar la parte web en un navegador; las funciones de Android se prueban dentro del APK.

## Dale un nombre

Sustituye el contenido de `my-app/koru.json` por:

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

`name` aparece en Android. `appId` identifica tu app; elige uno propio antes de distribuirla. Consulta [Configuración]({{base}}es/configuration/) cuando necesites cambiar el icono, los permisos o la carpeta web.

## Genera el APK

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
./bin/koru verify app-debug.apk
```

`doctor` revisa el proyecto. `build` genera `app-debug.apk` en la carpeta actual. `verify` devuelve los datos del APK y comprueba su firma y estructura. La salida del APK debe quedar fuera de `www`.

Este build permite inspeccionar el WebView durante las pruebas. Para compartir la app, usa después [un build release]({{base}}es/distribution/).

## Abre la app en Android

Conecta un dispositivo con depuración USB activada o inicia un emulador. Con `adb` instalado y el dispositivo conectado, ejecuta:

```sh
adb install -r app-debug.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

Si elegiste otro `appId`, sustituye `com.example.hello` antes de la barra. Verás la página de `www/index.html` dentro de tu app. `adb` se usa para instalar y probar; no hace falta para generar el APK.

Cambia algo en el HTML, repite `build` y vuelve a instalar. Ya tienes el recorrido de edición y prueba.

> La ejecución registrada de koru cubre Android 11 (API 30), x86_64. El rango configurable es API 26–36; comprueba tu app en los dispositivos que quieras admitir.

## Sigue con tu app

- [Usa tu frontend existente]({{base}}es/existing-project/).
- [Guarda datos y usa funciones de Android]({{base}}es/plugins/).
- [Prepara una versión para compartir]({{base}}es/distribution/).

Si un paso falla, consulta [Solución de problemas]({{base}}es/troubleshooting/).
