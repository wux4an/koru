# Ejecuta en un dispositivo

`koru run` compila tu app, la instala en un móvil o emulador conectado y la abre. Úsalo cada vez que quieras ver el comportamiento real de Android: plugins nativos, permisos, el menú de compartir y las notificaciones.

```sh
koru run
```

```text
Built com.example.miapp 1.0.0 (20 KB) in 35ms
Installed on sdk gphone64 x86 64 (emulator-5554)
Opened Mi App on sdk gphone64 x86 64 (emulator-5554); inspect it at chrome://inspect
```

> `koru run` y `koru devices` son nuevos. Están probados contra un programa que implementa el protocolo del servidor adb, pero todavía no se han registrado en hardware físico. Si algo no funciona en tu equipo, los comandos manuales del final de esta página siempre funcionan.

## Prepara un dispositivo

**Móvil.** Activa las [opciones para desarrolladores y la depuración USB](https://developer.android.com/studio/debug/dev-options), conecta el cable y acepta en el móvil el aviso «¿Permitir depuración USB?».

**Emulador.** Crea e inicia un dispositivo virtual desde el Device Manager de Android Studio. Las imágenes con API 26 o posterior funcionan con el rango de `minSdk` de koru.

koru se comunica con el **servidor adb**, el servicio en segundo plano de las platform tools de Android que Android Studio inicia por ti. koru no inicia programas por su cuenta. Si no hay ningún servidor activo:

```sh
adb start-server
```

Comprueba qué ve koru:

```sh
koru devices
```

```text
emulator-5554            device        sdk gphone64 x86 64
R5CT1234567              unauthorized
```

`device` significa listo. `unauthorized` indica que el móvil espera que aceptes el aviso de depuración. `offline` suele resolverse reconectando el cable o reiniciando el emulador.

## Elige un dispositivo

Con un solo dispositivo listo, `koru run` lo usa. Si hay varios, elige uno por su número de serie:

```sh
koru run --device emulator-5554
```

También puedes definir `ANDROID_SERIAL=emulator-5554`. koru respeta `ANDROID_ADB_SERVER_PORT` y `ADB_SERVER_SOCKET=tcp:host:puerto`, las mismas variables que `adb`, por ejemplo para usar un servidor en otra máquina.

## Depura el WebView

Las builds de depuración (las que usa `koru run` por defecto) activan la depuración del WebView. Con la app abierta, visita `chrome://inspect` en Chrome en tu ordenador y elige tu app para usar DevTools: consola, red, elementos y puntos de interrupción.

`koru run --release` instala en cambio una build release. La depuración queda desactivada, como en el APK que distribuyes.

## Si la instalación falla

koru explica los errores habituales de Android:

| Mensaje                                 | Solución                                                                                                                               |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| …installed with a different signing key | El dispositivo tiene una app con el mismo `appId` firmada con otra clave. Desinstálala en el dispositivo o recupera la clave original. |
| …the device has a newer versionCode     | Incrementa `versionCode` en `koru.json` o desinstala la app más nueva.                                                                 |
| …older than minSdk                      | La versión de Android del dispositivo es inferior a `minSdk`. Usa un dispositivo o una imagen de emulador más reciente.                |
| no adb server …                         | Ejecuta `adb start-server` o abre Android Studio.                                                                                      |
| several devices are connected           | Añade `--device <serie>`.                                                                                                              |

## Instala a mano

Cualquier APK de `koru build` se puede instalar también directamente con las platform tools:

```sh
koru build --output app.apk
adb install -r app.apk
adb shell am start -W -n com.example.miapp/com.koru.runtime.MainActivity
```

Sustituye `com.example.miapp` por tu `appId`. Para compartir la app con otras personas, consulta [Distribuye tu app]({{base}}es/distribution/).
