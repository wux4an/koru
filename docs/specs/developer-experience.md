# Spec: experiencia de desarrollo al nivel de Capacitor y Expo

Fecha: 2026-10-26. Estado: aprobado por delegación (AGENTS.md autoriza specs).

## Objetivo

Que una persona que conoce HTML/JS pase de cero a su app abierta en un móvil
sin leer la arquitectura interna, con el mismo recorrido que ofrecen Capacitor
(`npx cap run android`) y Expo (`npx create-expo-app`, `expo start`):

1. **Crear** un proyecto con plantilla (`koru create`).
2. **Desarrollar** en el navegador con recarga en vivo y la misma API JS
   (`koru dev`), también delante de Vite u otro servidor (`--proxy`).
3. **Ejecutar** en un dispositivo o emulador con un comando (`koru run`).
4. **Distribuir** un APK release firmado (`koru build --release`), ya existente.

Sin cambiar la filosofía: Go propio + biblioteca estándar para DEX, AXML, ARSC,
APK y firmas; sin Java/Kotlin, JVM, Gradle ni herramientas del SDK en el build.

## Supuestos

1. `koru run` habla el protocolo del servidor ADB por TCP (`127.0.0.1:5037`).
   No lanza procesos: si el servidor no está activo, falla con una instrucción.
   Ver ADR 0003.
2. La API JS funciona en el navegador con implementaciones web (como los
   "web implementations" de Capacitor). En el origen de la app
   (`https://app.koru.invalid`) el comportamiento nativo no cambia.
3. No hay publicación en npm: las plantillas copian el módulo JS al proyecto.
4. Sin dispositivo en este entorno: `run` se prueba con un servidor ADB falso.

## Comandos

```sh
koru create my-app --template basic|vite [--name "My App"] [--app-id com.example.myapp]
koru dev [project] [--port 5173] [--host 127.0.0.1] [--proxy http://localhost:5173]
koru run [project] [--device SERIAL] [--release]
koru devices
koru init [dir]            # se mantiene (proyecto mínimo)
```

## Estructura

- `cmd/koru/` — comandos Cobra nuevos (`create.go`, `dev.go`, `run.go`).
- `internal/templates/` — plantillas embebidas (`embed.FS`).
- `internal/devserver/` — servidor HTTP, recarga SSE, proxy inverso, vigilante.
- `internal/adb/` — cliente del protocolo smart-socket ADB.
- `packages/javascript/` — `core.js` con modo web, `core.d.ts`, `koru-env.d.ts`.
- `internal/config/koru.schema.json` — JSON Schema de `koru.json` (`$schema`).
- `website/` — landing page y documentación reorganizada (en/es).

## Estilo

Igual que el código actual: errores explícitos, sin fallbacks silenciosos,
límites de tamaño, `DisallowUnknownFields`, pruebas en el mismo paquete.

## Pruebas

- Go: unitarias por paquete (servidor ADB falso, `httptest` para dev server,
  plantillas generan proyectos que `koru build` acepta y `Verify` valida).
- JS: `node --test packages/javascript/core.test.cjs` cubre el modo web.
- Sitio: `bun run build` y `bun test`.

## Límites

- Siempre: gofmt, vet, test, race; documentar en milestones.
- Nunca: `os/exec` en `cmd/koru`, builder o runtime; claves versionadas.
- Fuera de alcance: AAB, iOS, recarga en vivo dentro del APK (exigiría tráfico
  en claro y cambiar el runtime DEX sin poder validarlo en ART aquí).

## Criterios de éxito

- `koru create x && koru build x` produce un APK verificado con ambas plantillas.
- `koru dev` sirve la app, inyecta el puente y recarga al editar un archivo.
- La API JS resuelve en navegador (`platform === "web"`) sin esperar timeouts.
- `koru run` instala por streaming y lanza la Activity contra el servidor falso.
- El sitio tiene landing en `/` y `/es/inicio/`, y guías nuevas en ambos idiomas.
