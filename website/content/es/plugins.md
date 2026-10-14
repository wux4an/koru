# Plugins y API JavaScript

El builder integra `/__koru/core.js` y `/__koru/core.mjs`, e incluye el primero en
las páginas HTML. La API también se puede consumir desde `@koru/core` al añadir
`packages/javascript` al proyecto que compila sus archivos estáticos.

## Una página funcional

Agrega este botón y módulo dentro de `<body>` en un HTML de `webDir`:

```html
<button id="device">Mostrar dispositivo</button>
<p id="result" role="status">Listo</p>
<script type="module">
  import { device, storage } from "/__koru/core.mjs";

  document.querySelector("#device").addEventListener("click", async () => {
    const result = document.querySelector("#result");
    try {
      const info = await device.getInfo();
      await storage.set("last-model", info.model);
      result.textContent = "Dispositivo: " + info.model;
    } catch (error) {
      result.textContent = error.code + ": " + error.message;
    }
  });
</script>
```

Dentro del APK, el botón muestra el modelo Android y guarda el texto en almacenamiento privado. El modelo depende del dispositivo. En una vista previa de navegador común no existe el puente; instala el APK para probarlo. Los IDs `device` y `storage` ya están activados en el `koru.json` generado.

El código siguiente es una muestra de métodos, no una secuencia obligatoria:

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

| Plugin        | Métodos                                  | Permisos añadidos                                                 |
| ------------- | ---------------------------------------- | ----------------------------------------------------------------- |
| device        | device.info                              | Ninguno: modelo, fabricante, SO, SDK                              |
| app           | app.info                                 | Ninguno                                                           |
| storage       | get/set/remove/clear                     | Ninguno; SharedPreferences privado; keys limitadas, values string |
| vibration     | vibration.vibrate                        | VIBRATE; duración entera 1–10000 ms                               |
| share         | share.text                               | Ninguno; chooser ACTION_SEND text/plain                           |
| notifications | show/cancel/permission/requestPermission | POST_NOTIFICATIONS; canal local desde API 26                      |

Todos devuelven promesas. `storage.get(key)` resuelve a un string o `null`; `storage.set(key, value)` recibe strings. `vibration.vibrate(duration)` usa milisegundos enteros entre 1 y 10.000. `share.text(text)` abre el selector Android para compartir texto. `notifications.show({ id, title, body })` publica inmediatamente; `notifications.cancel(id)` la retira. Activa los plugins opcionales en `koru.json` antes de llamarlos. Un método ausente o desactivado rechaza la promesa.

Estos métodos también se invocan con `Koru.invoke(method, params, {timeout, signal})`.
Notificaciones `show` publica inmediatamente; no es un planificador de alarmas
futuras. `requestPermission` espera la decisión o el regreso desde ajustes. En
Android 33+ se solicita permiso; en Android 26–32 se comprueba el control de
notificaciones del sistema. La API no evita restricciones impuestas por el usuario.

Cancelación rechaza el Promise y evita una petición todavía no enviada. No revierte
escrituras, vibración, notificación o UI que ya se hayan ejecutado. Códigos principales:
INVALID_REQUEST, METHOD_NOT_FOUND, NATIVE_ERROR, PERMISSION_DENIED, STORAGE_ERROR,
BUSY, TIMEOUT, CANCELLED, CHANNEL_ERROR y PAGE_CLOSED.

## Plugin personalizado

Declarar `pluginManifests` en koru.json y activar su `id` en plugins. El manifiesto
contiene ABI=1, permisos, clases del IR DEX y un mapa de handlers. Cada handler es
public static con descriptor:

`(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;`

El dispatcher valida el protocolo, llama al handler y convierte excepciones en un
error estructurado. El plugin debe validar sus params y respetar el hilo UI. Sus
clases no pueden reemplazar el runtime/framework. Es código firmado de confianza,
con acceso a los permisos de su aplicación.

`examples/native/plugin.json` usa una clase DEX con clinit que invoca
System.loadLibrary y un método native. No contiene Java. Se integra así:

```json
{
  "plugins": ["device", "native"],
  "pluginManifests": ["plugins/native/plugin.json"],
  "nativeLibraries": { "x86_64": ["plugins/native/libkoru_native.so"] }
}
```

Son fragmentos que deben añadirse a la configuración completa, no un koru.json
completo. Las librerías se validan como ELF shared objects de la ABI configurada;
no se compilan durante koru build. El binario de ejemplo es x86_64 y fue probado
en Android 11. Para teléfonos ARM se debe proporcionar una librería de su ABI.
Las apps con los seis plugins estándar no requieren librerías nativas y son
independientes de la arquitectura del teléfono.

El código C y la API de scopes JNI están en runtime/native. Cada scope posee su
JNIEnv del hilo actual; solo se hace detach si el scope hizo attach. Referencias
locales no salen de su frame; referencias de clase compartidas son globales y se
liberan en JNI_OnUnload. `scope_test.c` prueba ownership del attach/detach mediante
una tabla JNI de prueba; el plugin real se ejecutó adicionalmente en ART.
