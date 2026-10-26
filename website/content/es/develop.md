# Desarrolla con recarga en vivo

`koru dev` sirve tu proyecto a un navegador mientras trabajas. Recarga la página al guardar y ejecuta la API JavaScript de koru con implementaciones web, así que puedes construir y depurar la mayor parte de tu app antes de generar un APK.

```sh
koru dev            # en el directorio del proyecto
koru dev ./mi-app   # o indicando el proyecto
```

```text
koru dev live reload on save; press Ctrl+C to stop
  Local:    http://127.0.0.1:8100
  The koru API uses web implementations here; run `koru run` for the native app.
```

## Qué pasa al guardar

koru vigila el directorio del proyecto (sin contar `.git`, `.koru` ni `node_modules`). Cuando cambia un archivo, todas las páginas abiertas se recargan. Los cambios en `koru.json`, como activar un plugin, también se aplican en la siguiente recarga.

Las páginas se sirven desde `webDir` con las mismas reglas que el APK: nunca se sirven archivos ocultos ni enlaces simbólicos, y `"spa": true` devuelve `index.html` para rutas sin extensión. koru añade dos scripts a cada página HTML: `/__koru/dev.js` (recarga y datos de tu app) y `/__koru/core.js` (la API).

## La API JavaScript en el navegador

Las llamadas funcionan igual que en Android, con equivalentes del navegador:

| API             | En el navegador                                                                        |
| --------------- | -------------------------------------------------------------------------------------- |
| `device`        | `model` es `"Web browser"` y `sdk` es `0`.                                             |
| `app`           | Tu `appId`, `name`, `version` y `versionCode` de `koru.json`.                          |
| `storage`       | `localStorage`, con claves con el prefijo `koru:`. Se aplican las mismas reglas.       |
| `vibration`     | `navigator.vibrate()` donde exista (Chrome en Android); si no, la promesa se resuelve. |
| `share`         | El menú de compartir del navegador o, si no existe, copia el texto al portapapeles.    |
| `notifications` | La API Notification del navegador y su solicitud de permiso.                           |

Usa `platform` para distinguirlos:

```js
import { platform, isNative } from "/__koru/core.mjs";

if (isNative()) {
  // Se ejecuta dentro del APK en Android.
}
console.log(platform); // "android" o "web"
```

Con `koru dev`, una llamada a un plugin que **no** esté activado en `koru.json` se rechaza con `METHOD_NOT_FOUND`, igual que en el APK. Así detectas el error antes de compilar.

> Las implementaciones web son para desarrollar. Prueba siempre el comportamiento real con [`koru run`]({{base}}es/run-on-device/): los permisos, el menú de compartir de Android, las notificaciones y la vibración no son iguales en un navegador.

## Pruébala en el navegador del móvil

```sh
koru dev --host 0.0.0.0
```

koru también muestra direcciones `Network:`, como `http://192.168.1.20:8100`. Abre una en el navegador del móvil con ambos dispositivos en la misma red. Sirve para revisar el diseño y el tacto; la API sigue usando implementaciones web. Usa `0.0.0.0` solo en redes de confianza: cualquiera en la red puede leer los archivos servidos.

## Usa el servidor de desarrollo de tu framework

Vite, Astro, SvelteKit y herramientas similares ya tienen recarga de módulos en caliente. Consérvala y deja que koru añada su API y los datos de tu app delante:

```sh
npm run dev                                  # p. ej., Vite en http://localhost:5173
koru dev --proxy http://localhost:5173       # en una segunda terminal
```

Abre `http://127.0.0.1:8100`. koru reenvía todas las peticiones salvo `/__koru/…` al servidor del framework, añade sus scripts a las páginas HTML y deja pasar las conexiones WebSocket, así que la recarga en caliente sigue funcionando.

La [plantilla `vite`]({{base}}es/existing-project/) incluye una copia del módulo JavaScript en `src/koru/`, de modo que `npm run dev` por sí solo ya tiene la API con implementaciones web. Usa `--proxy` cuando quieras además comprobar los plugins contra `koru.json` y los datos reales de tu app.

## Opciones

| Opción          | Valor por defecto | Significado                                                        |
| --------------- | ----------------- | ------------------------------------------------------------------ |
| `--port <n>`    | `8100`            | Puerto de escucha. `0` elige uno libre.                            |
| `--host <dir>`  | `127.0.0.1`       | Dirección de escucha. `0.0.0.0` la expone en la red.               |
| `--proxy <url>` | ninguno           | Reenvía a otro servidor de desarrollo en lugar de servir `webDir`. |

`koru dev` nunca genera un APK y no forma parte de la cadena de compilación. Siguiente paso: [ejecuta en un dispositivo]({{base}}es/run-on-device/).
