# Pruebas reproducibles

La compilación y las pruebas Go básicas no necesitan SDK, Java, Node ni Python:

```sh
go test ./...
go vet ./...
go test -race ./...
```

Cliente JavaScript: `node --test packages/javascript/core.test.cjs` (protocolo
nativo y modo web).

Flujo de desarrollo (sin red ni dispositivo): `internal/templates` genera ambas
plantillas, `internal/devserver` prueba estáticos, SPA, SSE, vigilancia y proxy
con `httptest`, e `internal/adb` usa el servidor adb falso de
`internal/adb/adbtest`. `cmd/koru` encadena `create` → `build` → `verify`,
`dev` y `run` contra ese servidor falso. Ninguna de estas pruebas sustituye una
instalación real en Android.

Sitio de documentación (Bun): `cd website && bun test` y
`CHROME_PATH=/ruta/chrome bun run test:browser`, que incluye la landing page.
JNI: `make native-test` prueba propiedad de referencias/hilos con una tabla JNI
controlada; no sustituye la ejecución real del ejemplo en Android.

## Comparación externa opcional

Estas herramientas **solo verifican el artefacto ya generado por Go**. No se
incluyen en la entrega ni participan en `koru build`.

```sh
export KORU_TEST_DEXDUMP=/ruta/dexdump
export KORU_TEST_AAPT=/ruta/aapt
export KORU_TEST_JAVA=/ruta/java
export KORU_TEST_APKSIGNER_JAR=/ruta/apksigner.jar
go test -tags android_external -v ./tests/external
```

El entorno debe resolver las bibliotecas compartidas de dexdump/aapt cuando el
prebuilt oficial las requiera.

## Ejecución Android

Instalar un APK debug con los seis plugins y el ejemplo `native.info` configurado
según docs/plugins.md. Para la muestra usada por android-smoke.py, el identificador
es `com.example.koru`, y webDir contiene examples/web más tests/fixtures/sw.js copiado como sw.js. Ese service
worker debe responder a `/sw-probe` con `koru-worker-offline` y reclamar sus clientes.
El APK de entrega koru-demo.apk es release y no expone DevTools.

```sh
export KORU_TEST_ADB=/ruta/adb
"$KORU_TEST_ADB" install -r app-debug.apk
"$KORU_TEST_ADB" shell svc wifi disable
"$KORU_TEST_ADB" shell svc data disable
"$KORU_TEST_ADB" shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
pid=$("$KORU_TEST_ADB" shell pidof com.example.koru | tr -d '\r')
"$KORU_TEST_ADB" forward tcp:9223 "localabstract:webview_devtools_remote_$pid"
python3 tests/android-smoke.py
python3 tests/android-navigation.py
python3 tests/android-worker.py
python3 tests/android-keyboard.py
python3 tests/android-share.py
```

Python requiere websockets para estas pruebas de QA. El puerto predeterminado es
9223; el script smoke admite --port y --output. Ejecutar secuencialmente y volver a
abrir index.html antes de smoke si el contador fue usado. Un force-stop y nuevo
arranque entre smoke y navigation permite comprobar persistencia de storage.

El teclado/navegación usan la pantalla 720×1280, densidad 320 y barra de estado
24dp del emulador de referencia. Adaptar coordenadas a otra pantalla. Ningún test
controla aplicaciones ajenas ni envía el contenido del chooser a un destinatario.

Resultados y matriz comprobada: docs/results/full/README.md. La ejecución en API
30 no prueba el diálogo de permisos de notificaciones introducido en API 33.

## Formato antes de commit

```sh
gofmt -w cmd internal pkg plugins runtime/dex examples/dex tests/external
black tests/*.py
clang-format -i runtime/native/koru_plugin.c runtime/native/koru_plugin.h runtime/native/scope_test.c
```

Los archivos JS/HTML/CSS/JSON se formatearon con Prettier 3.6.2. Los headers JNI de
AOSP conservan su formato y licencia original. Los formateadores son herramientas
de desarrollo, no requisitos para ejecutar Koru.
