# API JavaScript

Usa funciones de Android desde el JavaScript de tu app: consulta el dispositivo, guarda datos, comparte texto o muestra notificaciones. Las llamadas devuelven promesas. Dentro del APK se ejecutan de forma nativa en Android; en un navegador y con [`koru dev`]({{base}}es/develop/) las mismas llamadas usan implementaciones web, así que puedes desarrollar sin un dispositivo.

## Antes de llamar la API

`koru init` activa `device`, `storage` y `app`; la plantilla `basic` activa además `vibration` y `share`. Para usar otro plugin, añade su ID y reconstruye:

```json
"plugins": ["device", "storage", "app", "vibration", "share", "notifications"]
```

Es un campo para añadir a tu `koru.json`, no un archivo completo. Activar `vibration` añade `android.permission.VIBRATE`; `notifications` añade `android.permission.POST_NOTIFICATIONS`. En el HTML empaquetado importa el módulo generado:

```html
<script type="module">
  import { device } from "/__koru/core.mjs";
  document.querySelector("#model").textContent = (await device.getInfo()).model;
</script>
```

Añade a la página un elemento con `id="model"`. Koru también inyecta `/__koru/core.js`, que expone `window.Koru`. El módulo exporta `device`, `app`, `storage`, `vibration`, `share`, `notifications`, `invoke`, `platform` e `isNative`. Con un bundler como Vite, importa en su lugar la copia que escribe [`koru js`]({{base}}es/cli/) (`./koru/core.mjs`); se comporta igual. Todos los métodos devuelven una promesa. Maneja los errores para mostrar algo útil:

```js
try {
  const info = await device.getInfo();
  console.log(info.model);
} catch (error) {
  console.error(error.code, error.message);
}
```

## TypeScript y soporte del editor

El módulo incluye declaraciones de tipos. Los proyectos de `koru create` ya vienen configurados:

- **basic**: `types/koru.d.mts` y un `jsconfig.json` que asocia `/__koru/core.mjs` a ese archivo, para que VS Code y `tsc --checkJs` comprueben `www/app.js`.
- **vite**: `src/koru/core.d.mts` está junto a `core.mjs` y se detecta automáticamente.

En tu propio proyecto TypeScript, ejecuta `koru js src/koru` o asocia la ruta servida en `tsconfig.json`:

```json
{
  "compilerOptions": {
    "moduleResolution": "bundler",
    "paths": { "/__koru/core.mjs": ["./types/koru.d.mts"] }
  }
}
```

Los tipos cubren cada método, la forma de cada resultado, `CallOptions` y los códigos de `KoruError`:

```ts
import { storage, type KoruError } from "/__koru/core.mjs";

try {
  const theme: string | null = await storage.get("theme");
} catch (error) {
  console.warn((error as KoruError).code);
}
```

## Android o navegador

`platform` vale `"android"` dentro del APK y `"web"` en cualquier otro sitio. `isNative()` devuelve `platform === "android"`. Úsalos en los pocos lugares donde el comportamiento deba cambiar:

```js
import { isNative, share } from "/__koru/core.mjs";

const label = isNative() ? "Compartir" : "Copiar enlace";
```

En el navegador, `device.getInfo()` devuelve `model: "Web browser"` y `sdk: 0`; `storage` usa `localStorage` con el prefijo `koru:`; `share.text` usa la Web Share API o copia al portapapeles; las notificaciones usan la API Notification del navegador; `vibration.vibrate` usa `navigator.vibrate` donde existe. La validación de argumentos y los códigos de error son los mismos que en Android. Un navegador sin equivalente rechaza con `UNAVAILABLE`. Con `koru dev`, las llamadas a plugins no activados en `koru.json` se rechazan con `METHOD_NOT_FOUND`, como en el APK.

## Índice de métodos

| Plugin         | Métodos                                                                                                                                                                        | ¿Activo con `init`?           |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| Dispositivo    | [`getInfo()`](#api-device-info)                                                                                                                                                | Sí                            |
| App            | [`getInfo()`](#api-app-info)                                                                                                                                                   | Sí                            |
| Almacenamiento | [`get()`](#api-storage-get), [`set()`](#api-storage-set), [`remove()`](#api-storage-remove), [`clear()`](#api-storage-clear)                                                   | Sí                            |
| Vibración      | [`vibrate()`](#api-vibration)                                                                                                                                                  | No                            |
| Compartir      | [`text()`](#api-share)                                                                                                                                                         | No                            |
| Notificaciones | [`permission()`](#api-notification-permission), [`requestPermission()`](#api-notification-request), [`show()`](#api-notification-show), [`cancel()`](#api-notification-cancel) | No                            |
| Puente directo | [`invoke()`](#api-invoke)                                                                                                                                                      | Disponible en páginas del APK |

Cada acordeón se abre por separado. Los ejemplos presuponen importaciones desde `/__koru/core.mjs` y funcionan dentro del APK.

## Dispositivo y aplicación

<details class="api-detail" id="api-device-info"><summary><code>device.getInfo()</code><span>Consulta el dispositivo Android</span></summary>

**Firma:** `device.getInfo(options?) → Promise<DeviceInfo>`. Sin parámetros nativos ni permisos.

**Resultado:** `{ model: string, manufacturer: string, osVersion: string, sdk: number, platform: "android" }`. Los valores vienen de los campos `Build` de Android y cambian según el dispositivo; `sdk` es su nivel de API.

```js
import { device } from "/__koru/core.mjs";
const info = await device.getInfo();
console.log(
  `${info.manufacturer} ${info.model} · Android ${info.osVersion} (API ${info.sdk})`,
);
```

`options` admite `timeout` y `signal`, explicados en `invoke()`. Un plugin desactivado rechaza con `METHOD_NOT_FOUND`; el canal y el tiempo de espera también pueden producir errores.

</details>

<details class="api-detail" id="api-app-info"><summary><code>app.getInfo()</code><span>Consulta la identidad de la app</span></summary>

**Firma:** `app.getInfo(options?) → Promise<AppInfo>`. Sin parámetros nativos ni permisos.

**Resultado:** `{ appId: string, name: string, version: string, versionCode: number }`, tomado del `koru.json` con el que se generó el APK.

```js
import { app } from "/__koru/core.mjs";
const info = await app.getInfo();
document.querySelector("#version").textContent = `${info.name} ${info.version}`;
```

El objeto describe la app empaquetada. No consulta una tienda ni versiones disponibles.

</details>

## Almacenamiento privado

Storage usa `SharedPreferences` privado de la app. Las claves tienen 1–128 caracteres: `A–Z`, `a–z`, `0–9`, `_`, `.`, `:` y `-`. Los valores son strings. Los datos persisten entre arranques y actualizaciones firmadas con la misma identidad. Se eliminan al desinstalar. No pongas secretos en una página web ni supongas que este almacenamiento está cifrado.

<details class="api-detail" id="api-storage-get"><summary><code>storage.get(key)</code><span>Lee un string o null</span></summary>

**Firma:** `storage.get(key: string, options?) → Promise<string | null>`.

**Argumentos:** `key` debe ser una clave válida. **Resultado:** el string guardado, o `null` si nunca existió o fue eliminado.

```js
import { storage } from "/__koru/core.mjs";
const theme = (await storage.get("theme")) ?? "system";
console.log(theme);
```

Una clave ausente, de tipo incorrecto o inválida rechaza con `INVALID_REQUEST`. Este wrapper extrae por ti el valor del objeto nativo `{ value }`.

</details>

<details class="api-detail" id="api-storage-set"><summary><code>storage.set(key, value)</code><span>Guarda un string</span></summary>

**Firma:** `storage.set(key: string, value: string, options?) → Promise<{}>`.

**Argumentos:** clave válida y valor string. Android confirma el guardado antes de resolver la promesa. **Resultado:** objeto vacío `{}`.

```js
import { storage } from "/__koru/core.mjs";
await storage.set("theme", "dark");
await storage.set("visits", String(3)); // convierte números a string
```

Tipos incorrectos rechazan con `INVALID_REQUEST`; un guardado fallido, con `STORAGE_ERROR`.

</details>

<details class="api-detail" id="api-storage-remove"><summary><code>storage.remove(key)</code><span>Elimina una clave</span></summary>

**Firma:** `storage.remove(key: string, options?) → Promise<{}>`. **Resultado:** objeto vacío una vez confirmada la eliminación. Eliminar una clave inexistente es válido.

```js
import { storage } from "/__koru/core.mjs";
await storage.remove("theme");
console.log(await storage.get("theme")); // null
```

Una clave inválida rechaza con `INVALID_REQUEST`; un guardado fallido, con `STORAGE_ERROR`.

</details>

<details class="api-detail" id="api-storage-clear"><summary><code>storage.clear()</code><span>Elimina todas las claves koru</span></summary>

**Firma:** `storage.clear(options?) → Promise<{}>`. Elimina todas las claves del espacio privado de koru para esta app y resuelve con `{}`. No borra los assets web empaquetados ni otros datos Android de la app.

```js
import { storage } from "/__koru/core.mjs";
await storage.clear();
```

Un guardado fallido rechaza con `STORAGE_ERROR`. Úsalo cuando el usuario realmente quiera borrar todas las claves.

</details>

## Acciones del dispositivo

<details class="api-detail" id="api-vibration"><summary><code>vibration.vibrate(duration?)</code><span>Vibra durante unos milisegundos</span></summary>

**Configuración:** activa `vibration` en `koru.json`; koru añade `android.permission.VIBRATE` al APK.

**Firma:** `vibration.vibrate(duration = 100, options?) → Promise<{}>`. `duration` debe ser entero entre 1 y 10.000 milisegundos. La promesa resuelve con `{}` al iniciar la acción Android; no espera a que termine la vibración.

```js
import { vibration } from "/__koru/core.mjs";
await vibration.vibrate(150);
```

Duraciones inválidas rechazan con `INVALID_REQUEST`. Un fallo de hardware o Android puede rechazar con `NATIVE_ERROR`; un dispositivo sin vibrador no ofrece respuesta táctil.

</details>

<details class="api-detail" id="api-share"><summary><code>share.text(text)</code><span>Abre el panel Compartir de Android</span></summary>

**Configuración:** activa `share` en `koru.json`. No añade permisos.

**Firma:** `share.text(text: string, options?) → Promise<{}>`. `text` es el string a compartir. Koru abre el selector `ACTION_SEND` con tipo `text/plain` y resuelve con `{}` tras lanzarlo. Eso no significa que el usuario haya elegido un destino o completado el envío.

```js
import { share } from "/__koru/core.mjs";
await share.text("Prueba mi app koru");
```

Un valor que no sea string rechaza con `INVALID_REQUEST`; si no se puede abrir el selector puede aparecer `NATIVE_ERROR`.

</details>

## Notificaciones locales

Activa `notifications` en `koru.json` y reconstruye. Koru añade `android.permission.POST_NOTIFICATIONS`. Android 13/API 33+ puede pedir una decisión durante la ejecución; las versiones compatibles anteriores usan el ajuste de notificaciones del sistema. El diálogo de API 33+ no se ha probado en la matriz registrada de API 30. Las notificaciones son locales e inmediatas y usan el canal `koru-local`; esta API no programa alarmas futuras ni push remoto.

<details class="api-detail" id="api-notification-permission"><summary><code>notifications.permission()</code><span>Consulta el permiso actual</span></summary>

**Firma:** `notifications.permission(options?) → Promise<{ status: "granted" | "denied" }>`.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.permission();
console.log(status); // "granted" o "denied"
```

Lee el permiso y el ajuste actuales de Android. No muestra un diálogo. Un plugin desactivado rechaza con `METHOD_NOT_FOUND`.

</details>

<details class="api-detail" id="api-notification-request"><summary><code>notifications.requestPermission()</code><span>Pide acceso a notificaciones</span></summary>

**Firma:** `notifications.requestPermission(options?) → Promise<{ status: "granted" | "denied" }>`.

En Android 13/API 33+, si está denegado abre el diálogo del permiso. En Android 8–12/API 26–32, abre los ajustes de notificación de la app y resuelve al regresar. Si ya está concedido, resuelve de inmediato. El timeout predeterminado del wrapper es **120.000 ms** para este método; `options.timeout` puede cambiarlo hasta ese límite. Solo puede haber una solicitud pendiente.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.requestPermission();
if (status === "granted") console.log("Notificaciones activas");
```

`BUSY` indica otra solicitud pendiente; `TIMEOUT`, que no llegó respuesta a tiempo. `denied` es un resultado normal, no un error lanzado.

</details>

<details class="api-detail" id="api-notification-show"><summary><code>notifications.show({ id, title, body })</code><span>Publica ahora una notificación</span></summary>

**Firma:** `notifications.show({ id: number, title: string, body: string }, options?) → Promise<{}>`.

`id` es un identificador entero; usa el mismo para reemplazar o cancelar una notificación. El valor nativo predeterminado es `1` si no pasas `id`, pero especificarlo aclara la cancelación posterior. `title` y `body` son strings obligatorios. La promesa resuelve con `{}` tras publicar, no cuando el usuario abre la notificación.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.permission();
if (status === "granted") {
  await notifications.show({
    id: 7,
    title: "Listo",
    body: "La exportación terminó.",
  });
}
```

Si las notificaciones están desactivadas o falta el permiso en Android 13+, rechaza con `PERMISSION_DENIED`. Tipos incorrectos rechazan con `INVALID_REQUEST`.

</details>

<details class="api-detail" id="api-notification-cancel"><summary><code>notifications.cancel(id)</code><span>Retira una notificación</span></summary>

**Firma:** `notifications.cancel(id: number, options?) → Promise<{}>`. Usa el mismo ID entero de `show`; resuelve con `{}` después de pedirle a Android que la cancele. Cancelar un ID ausente es válido. No abre un diálogo de permiso.

```js
import { notifications } from "/__koru/core.mjs";
await notifications.cancel(7);
```

Un ID que no sea entero rechaza con `INVALID_REQUEST`. Un fallo nativo Android puede rechazar con `NATIVE_ERROR`.

</details>

## Puente directo y errores

<details class="api-detail" id="api-invoke"><summary><code>invoke(method, params?, options?)</code><span>Llama directamente a un handler nativo</span></summary>

**Firma:** `invoke(method: string, params = {}, options = {}) → Promise<object>`. Es la función exportada por `/__koru/core.mjs`; `window.Koru.invoke` es el equivalente global. El método es un string como `"device.info"`; `params` debe ser un objeto simple serializable a JSON. Úsalo para un plugin personalizado o si necesitas el resultado nativo sin transformar.

```js
import { invoke } from "/__koru/core.mjs";
const info = await invoke("device.info", {});
const raw = await invoke("storage.get", { key: "theme" });
console.log(info.model, raw.value); // raw.value es string o null
```

`options.timeout` es un número finito de milisegundos entre 1 y 120.000; por defecto 10.000. El temporizador también incluye la espera del canal nativo. `options.signal` admite `AbortSignal`; cancelar rechaza con `CANCELLED`. La cancelación impide una petición que aún no se envió, pero no deshace una acción nativa ya iniciada. Máximo 256 llamadas pendientes. El JSON serializado de una petición debe medir menos de 65.536 caracteres.

**Resultado:** el objeto devuelto por el handler nativo. Las acciones integradas devuelven `{}`; el wrapper `storage.get()` convierte el resultado bruto `{ value }` en string o `null`. Los handlers personalizados definen su objeto de resultado.

</details>

Los errores rechazan con un `Error` cuyo `code` y `message` puedes mostrar o registrar. `INVALID_REQUEST` significa método o argumentos inválidos; `METHOD_NOT_FOUND`, método o plugin no disponible; `STORAGE_ERROR`, fallo de guardado; `PERMISSION_DENIED`, permiso o ajuste de notificaciones desactivado; `NATIVE_ERROR`, excepción Android; `UNAVAILABLE`, el navegador no tiene equivalente (solo implementaciones web). `BUSY`, `TIMEOUT`, `CANCELLED`, `CHANNEL_ERROR`, `CHANNEL_REPLACED`, `INVALID_RESPONSE` y `PAGE_CLOSED` describen el ciclo de vida del puente. Un timeout o cancelación no demuestra que una acción nativa ya iniciada se haya revertido.

## Amplía tu app

Para añadir funciones propias, consulta el [adaptador nativo]({{base}}es/native/) y la [referencia de plugins personalizados]({{base}}original/docs/plugins.md).

[Configuración]({{base}}es/configuration/) · [Solución de problemas]({{base}}es/troubleshooting/)
