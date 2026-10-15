# Plugins y API JavaScript

El builder incorpora `/__koru/core.js` y `/__koru/core.mjs` al APK e incluye el
primero en las páginas HTML. También puedes usar el paquete `@koru/core` desde
el proyecto que genera tus archivos estáticos; su código está en
`packages/javascript`.

```js
import {
  device,
  app,
  storage,
  vibration,
  share,
  notifications,
} from "/__koru/core.mjs";
console.log(await device.getInfo());
console.log(await app.getInfo());
await storage.set("theme", "dark");
console.log(await storage.get("theme"));
await vibration.vibrate(100);
await share.text("Hola");
await notifications.requestPermission({ timeout: 120000 });
await notifications.show({
  id: 1,
  title: "Recordatorio",
  body: "Notificación local",
});
```

| Plugin        | Métodos JavaScript públicos                                 | Permisos que añade                                                    |
| ------------- | ----------------------------------------------------------- | --------------------------------------------------------------------- |
| device        | `device.getInfo()`                                          | Ninguno: modelo, fabricante, SO y SDK                                 |
| app           | `app.getInfo()`                                             | Ninguno                                                               |
| storage       | `storage.get()`, `set()`, `remove()`, `clear()`             | Ninguno; SharedPreferences privado; claves limitadas y valores string |
| vibration     | `vibration.vibrate()`                                       | `VIBRATE`; duración entera de 1 a 10.000 ms                           |
| share         | `share.text()`                                              | Ninguno; selector `ACTION_SEND` con `text/plain`                      |
| notifications | `show()`, `cancel()`, `permission()`, `requestPermission()` | `POST_NOTIFICATIONS`; canal local desde API 26                        |

También puedes invocar los métodos directamente con
`Koru.invoke(method, params, {timeout, signal})`. `notifications.show` publica
una notificación de inmediato; no programa alarmas futuras. `requestPermission`
espera la respuesta del usuario o su regreso desde los ajustes. En Android 13
(API 33) y posteriores, solicita el permiso en tiempo de ejecución; en Android
8–12 (API 26–32), comprueba el ajuste de notificaciones de la app. La API respeta
las restricciones que haya establecido el usuario.

Al cancelar, la promesa se rechaza y se evita enviar la petición si todavía no se
ha despachado. La cancelación no revierte escrituras, vibraciones, notificaciones
ni cambios de interfaz que ya se hayan ejecutado. Los códigos de error principales
son `INVALID_REQUEST`, `METHOD_NOT_FOUND`, `NATIVE_ERROR`, `PERMISSION_DENIED`,
`STORAGE_ERROR`, `BUSY`, `TIMEOUT`, `CANCELLED`, `CHANNEL_ERROR` y `PAGE_CLOSED`.

## Plugin personalizado

Para añadir un plugin propio, declara su manifiesto en `pluginManifests` y activa
su `id` en `plugins`. El manifiesto especifica ABI 1, permisos, clases del IR DEX
y un mapa de handlers. Cada handler debe ser un método `public static` con este
descriptor:

`(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;`

El dispatcher valida el protocolo, llama al handler y devuelve un error
estructurado si se produce una excepción. El plugin debe validar sus parámetros
y respetar los requisitos del hilo de interfaz. Sus clases no pueden reemplazar
clases del runtime ni del framework Android. Un plugin es código de confianza
incluido en la app y comparte sus permisos.

`examples/native/plugin.json` muestra una clase DEX cuyo inicializador estático
invoca `System.loadLibrary` y declara un método nativo. No incluye código Java.
Añade estos campos a la configuración del proyecto:

```json
{
  "plugins": ["device", "native"],
  "pluginManifests": ["plugins/native/plugin.json"],
  "nativeLibraries": { "x86_64": ["plugins/native/libkoru_native.so"] }
}
```

Este ejemplo es un fragmento; no reemplaza el archivo `koru.json` completo. Koru
valida que cada biblioteca sea un objeto compartido ELF de la ABI declarada, pero
no la compila durante el build. El binario de ejemplo es para x86_64 y se probó
en Android 11. Para un dispositivo ARM, proporciona una biblioteca compatible con
su ABI. Las apps que usan los seis plugins estándar no necesitan bibliotecas
nativas y no dependen de la arquitectura del dispositivo.

El código C y la API de scopes JNI están en `runtime/native`. Cada scope usa el
`JNIEnv` del hilo actual y solo separa (`detach`) el hilo si ese scope lo había
conectado (`attach`). Las referencias locales no salen de su frame; las referencias
de clase compartidas son globales y se liberan en `JNI_OnUnload`. `scope_test.c`
comprueba la responsabilidad de conectar y separar hilos mediante una tabla JNI
de prueba. Además, el plugin real se ejecutó en ART.
