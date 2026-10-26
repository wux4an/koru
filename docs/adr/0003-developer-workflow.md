# ADR 0003: flujo de desarrollo sin procesos externos

Fecha: 2026-10-26. Spec: `docs/specs/developer-experience.md`.

## Contexto

Capacitor y Expo ofrecen un recorrido corto: crear desde plantilla, previsualizar
con recarga en vivo y ejecutar en un dispositivo con un comando. koru solo tenía
`init`, `build` y una instalación manual con `adb`. AGENTS.md prohíbe añadir
procesos externos a `cmd/koru`, al builder y al runtime.

## Decisión

1. **`koru create`** escribe plantillas embebidas con `embed.FS` (`basic`, `vite`)
   y crea la identidad de firma con el código existente. La plantilla Vite copia
   el módulo JS a `src/koru/` porque los bundlers resuelven imports al compilar.
2. **`koru dev`** es un servidor `net/http` local. Sirve `webDir` con las mismas
   reglas que el colector de assets (sin rutas ocultas ni enlaces simbólicos),
   inyecta `/__koru/dev.js` y `/__koru/core.js`, y recarga por Server-Sent Events.
   El vigilante sondea tamaños y fechas: no depende de notificaciones del SO.
   `--proxy` usa `httputil.ReverseProxy` delante del servidor del framework.
3. **Modo web de `core.js`.** Fuera de `https://app.koru.invalid` las llamadas
   usan implementaciones web (localStorage, Notification, navigator.share...),
   como las "web implementations" de Capacitor. En el origen de la app el código
   nativo no cambia. `Koru.platform` vale `android` o `web`.
4. **`koru run` y `koru devices`** hablan el protocolo smart-socket del servidor
   adb por TCP (`host:devices-l`, `host:transport:<serial>`,
   `exec:cmd package install -r -S <n>` con el APK por streaming y
   `exec:am start -W -n`). koru no inicia el servidor ni ningún programa: si no
   responde, el error indica `adb start-server`. Es un cliente de red, igual que
   el navegador hacia DevTools; no ejecuta herramientas del SDK.

## Consecuencias

- El build APK sigue siendo Go + stdlib; ninguna de estas piezas entra en él.
- La recarga en vivo dentro del APK queda fuera: exigiría tráfico en claro y un
  runtime DEX distinto, y no puede validarse en ART en este entorno.
- `koru run` está probado contra un servidor adb falso que implementa el
  protocolo (`internal/adb/adbtest`), no contra un dispositivo real. Hasta
  registrarlo en `docs/results`, la documentación lo marca como no validado en
  hardware.
- El cambio de `core.js` se ejecutó en Node y Chromium (modo web), no en el
  WebView de Android; el camino nativo no se modificó salvo la propiedad
  `platform` y la resolución inmediata de `ready` fuera del origen de la app.
