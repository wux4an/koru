# Frameworks & existing projects

koru packages static files: an `index.html` and the scripts, styles and images it uses. Any framework that can export a static site works. Point `webDir` at the build output and koru does the rest.

## Start from a framework template

The quickest path is a Vite-based template. They all need Node.js:

```sh
koru create my-app --template react
cd my-app
npm install
npm run dev        # Vite with hot reload
npm run android    # vite build, then koru run
npm run apk        # vite build, then a signed release APK
```

| `--template` | You get                                           |
| ------------ | ------------------------------------------------- |
| `react`      | Vite with React 18 and JSX.                       |
| `vue`        | Vite with Vue 3 single-file components.           |
| `svelte`     | Vite with Svelte 5.                               |
| `vanilla-ts` | Vite with strict TypeScript and no framework.     |
| `vite`       | A plain JavaScript Vite app, for any other setup. |

Each template has a `koru.json` with `"webDir": "dist"` and a copy of the koru JavaScript module in `src/koru/`, so imports work in Vite’s dev server and in production builds. For Solid or another framework, start from `vite` and add it the usual way, for example `npm install solid-js vite-plugin-solid` and a `vite.config.js`.

These templates were built with `npm run build`, `koru build` and `koru verify` (plus `tsc --noEmit` for `vanilla-ts`). They have not been installed or run on Android, so test one on a device before you build on it.

## Add koru to an existing project

From your project’s root, create a `koru.json` and signing key in place:

```sh
cd your-web-app
koru init
```

`init` also adds `.koru/` to your `.gitignore`. Then edit `koru.json`: set `name`, choose your own `appId` and point `webDir` at your build output. Delete the sample `www/` folder that `init` created if you do not use it.

```json
{
  "name": "My App",
  "appId": "com.example.myapp",
  "webDir": "dist",
  "spa": true
}
```

These are the fields to change; keep the others `init` wrote. Each build is then two steps:

```sh
npm run build
koru run           # or: koru build --release --output my-app.apk
```

## Output folders by framework

| Framework                        | Static build                                              | `webDir`                 | Client-side routes?                                          |
| -------------------------------- | --------------------------------------------------------- | ------------------------ | ------------------------------------------------------------ |
| Vite (React, Vue, Svelte, Solid) | `vite build`                                              | `dist`                   | Set `"spa": true` if you use a router                        |
| Create React App                 | `react-scripts build`                                     | `build`                  | `"spa": true` with React Router                              |
| Angular                          | `ng build`                                                | `dist/<project>/browser` | `"spa": true`                                                |
| Next.js                          | `output: "export"` in `next.config.js`, then `next build` | `out`                    | Leave off; pages are exported as HTML                        |
| Nuxt                             | `nuxi generate`                                           | `.output/public`         | Leave off for generated routes                               |
| SvelteKit                        | `@sveltejs/adapter-static`, then `vite build`             | `build`                  | Use the adapter’s `fallback: "index.html"` and `"spa": true` |
| Astro                            | `astro build` (static output, the default)                | `dist`                   | Leave off                                                    |

Server-side rendering, API routes and image optimization services do not run inside the APK. Use their static export, and call your server as an HTTPS API.

## Routing and paths

With `"spa": true`, any path without a file extension (such as `/settings/profile`) returns `index.html`, so your router can choose the screen. Missing files with extensions, such as `/assets/app.js`, still return 404 so a typo does not load HTML as JavaScript.

Files are served from the root of `https://app.koru.invalid/`. Build with a base path of `/` (the default in most tools). If your site is configured for a sub-path such as `/my-repo/`, set the build’s `base`/`publicPath` back to `/` for the APK.

## Use the JavaScript API from a bundler

Bundlers resolve imports at build time, so they cannot import `/__koru/core.mjs`, which only exists inside the app. Write a copy of the module into your sources:

```sh
koru js src/koru            # after upgrading koru: koru js src/koru --force
```

Then `import { device, storage } from "./koru/core.mjs"`. In the APK it uses the native bridge; in a browser or `vite dev` it uses web implementations. `core.d.mts` next to it gives TypeScript and editors the types. For apps without a bundler, import `/__koru/core.mjs` directly, as the `basic` template does.

## Call your API

Your app’s origin is `https://app.koru.invalid`. To call an HTTPS service, add the Internet permission and allow that origin in your server’s CORS settings:

```json
"permissions": ["android.permission.INTERNET"]
```

Bundle your JavaScript dependencies: remote scripts and iframes are blocked by the app’s content security policy. Plain `http://` requests are blocked too.

## Keep hot reload while you work

Run your framework’s dev server and put `koru dev` in front of it to get the koru API with your real app details and plugin checks:

```sh
npm run dev
koru dev --proxy http://localhost:5173
```

See [Develop with live reload]({{base}}docs/en/develop/). Then [run on a device]({{base}}docs/en/run-on-device/) and [distribute your app]({{base}}docs/en/distribution/).
