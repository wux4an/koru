# Frameworks y proyectos existentes

koru empaqueta archivos estáticos: un `index.html` y los scripts, estilos e imágenes que usa. Sirve cualquier framework que pueda exportar un sitio estático. Apunta `webDir` a la carpeta de salida y koru hace el resto.

## Empieza desde una plantilla de framework

El camino más rápido es una plantilla basada en Vite. Todas necesitan Node.js:

```sh
koru create mi-app --template react
cd mi-app
npm install
npm run dev        # Vite con recarga en caliente
npm run android    # vite build y después koru run
npm run apk        # vite build y después un APK release firmado
```

| `--template` | Obtienes                                                     |
| ------------ | ------------------------------------------------------------ |
| `react`      | Vite con React 18 y JSX.                                     |
| `vue`        | Vite con componentes de un solo archivo de Vue 3.            |
| `svelte`     | Vite con Svelte 5.                                           |
| `vanilla-ts` | Vite con TypeScript estricto y sin framework.                |
| `vite`       | Una app Vite en JavaScript simple, para cualquier otro caso. |

Cada plantilla trae un `koru.json` con `"webDir": "dist"` y una copia del módulo JavaScript de koru en `src/koru/`, para que los imports funcionen tanto en el servidor de desarrollo de Vite como en las builds de producción. Para Solid u otro framework, parte de `vite` y añádelo como de costumbre, por ejemplo `npm install solid-js vite-plugin-solid` y un `vite.config.js`.

Estas plantillas se compilaron con `npm run build`, `koru build` y `koru verify` (y `tsc --noEmit` en `vanilla-ts`). No se han instalado ni ejecutado en Android, así que pruébala en un dispositivo antes de construir sobre ella.

## Añade koru a un proyecto existente

Desde la raíz de tu proyecto, crea ahí mismo un `koru.json` y una clave de firma:

```sh
cd tu-app-web
koru init
```

`init` también añade `.koru/` a tu `.gitignore`. Después edita `koru.json`: define `name`, elige tu propio `appId` y apunta `webDir` a la salida de tu build. Borra la carpeta de ejemplo `www/` que creó `init` si no la usas.

```json
{
  "name": "Mi App",
  "appId": "com.example.miapp",
  "webDir": "dist",
  "spa": true
}
```

Estos son los campos que debes cambiar; conserva el resto de los que escribió `init`. Cada build queda en dos pasos:

```sh
npm run build
koru run           # o bien: koru build --release --output mi-app.apk
```

## Carpetas de salida por framework

| Framework                        | Build estática                                                | `webDir`                  | ¿Rutas en el cliente?                                      |
| -------------------------------- | ------------------------------------------------------------- | ------------------------- | ---------------------------------------------------------- |
| Vite (React, Vue, Svelte, Solid) | `vite build`                                                  | `dist`                    | `"spa": true` si usas un router                            |
| Create React App                 | `react-scripts build`                                         | `build`                   | `"spa": true` con React Router                             |
| Angular                          | `ng build`                                                    | `dist/<proyecto>/browser` | `"spa": true`                                              |
| Next.js                          | `output: "export"` en `next.config.js` y después `next build` | `out`                     | Desactivado; las páginas se exportan como HTML             |
| Nuxt                             | `nuxi generate`                                               | `.output/public`          | Desactivado para rutas generadas                           |
| SvelteKit                        | `@sveltejs/adapter-static` y después `vite build`             | `build`                   | Usa `fallback: "index.html"` del adaptador y `"spa": true` |
| Astro                            | `astro build` (salida estática, la predeterminada)            | `dist`                    | Desactivado                                                |

El renderizado en servidor, las rutas de API y los servicios de optimización de imágenes no se ejecutan dentro del APK. Usa su exportación estática y llama a tu servidor como una API HTTPS.

## Rutas y paths

Con `"spa": true`, cualquier ruta sin extensión (como `/settings/profile`) devuelve `index.html`, para que tu router elija la pantalla. Los archivos con extensión que no existen, como `/assets/app.js`, siguen devolviendo 404, para que una errata no cargue HTML como JavaScript.

Los archivos se sirven desde la raíz de `https://app.koru.invalid/`. Compila con una ruta base `/` (la predeterminada en casi todas las herramientas). Si tu sitio está configurado para una subruta como `/mi-repo/`, vuelve a poner `base`/`publicPath` en `/` para el APK.

## Usa la API JavaScript con un bundler

Los bundlers resuelven los imports al compilar, así que no pueden importar `/__koru/core.mjs`, que solo existe dentro de la app. Escribe una copia del módulo en tus fuentes:

```sh
koru js src/koru            # tras actualizar koru: koru js src/koru --force
```

Después, `import { device, storage } from "./koru/core.mjs"`. En el APK usa el puente nativo; en un navegador o con `vite dev`, las implementaciones web. El archivo `core.d.mts` que lo acompaña da los tipos a TypeScript y a los editores. En apps sin bundler, importa `/__koru/core.mjs` directamente, como hace la plantilla `basic`.

## Llama a tu API

El origen de tu app es `https://app.koru.invalid`. Para llamar a un servicio HTTPS, añade el permiso de Internet y permite ese origen en la configuración CORS de tu servidor:

```json
"permissions": ["android.permission.INTERNET"]
```

Empaqueta tus dependencias JavaScript: la política de seguridad de contenido de la app bloquea scripts e iframes remotos. Las peticiones `http://` sin cifrar también se bloquean.

## Conserva la recarga en caliente mientras trabajas

Ejecuta el servidor de desarrollo de tu framework y pon `koru dev` delante para tener la API de koru con los datos reales de tu app y la comprobación de plugins:

```sh
npm run dev
koru dev --proxy http://localhost:5173
```

Consulta [Desarrolla con recarga en vivo]({{base}}docs/es/develop/). Después, [ejecuta en un dispositivo]({{base}}docs/es/run-on-device/) y [distribuye tu app]({{base}}docs/es/distribution/).
