# Use your web project

Bring an HTML, CSS and JavaScript site, or your framework’s static output, to Android. You need a folder containing `index.html` and the files it uses.

If you do not have the executable yet, follow [getting started]({{base}}en/getting-started/). Run the commands in this guide from the koru repository root.

## Prepare your files

Build your frontend with your usual command. For a project with a `build` script:

```sh
npm --prefix /path/to/your-web-app run build
```

Find the output folder, usually `dist` or `build`. Koru packages this output after your frontend build. Apps that need a server to render pages must provide a static export or access their server as an HTTPS API.

## Create the Android project

Use a new directory for your koru configuration:

```sh
./bin/koru init ./my-app
```

Copy the contents of your frontend output into `my-app/www/`. Place `index.html` directly inside `www`, alongside its scripts, styles and images.

To avoid copying, change `webDir` in `my-app/koru.json` to point to your frontend output. Relative paths start at the directory containing `koru.json`. For example, `../your-web-app/dist` points to a sibling directory named `your-web-app`.

## Configure navigation

If your app uses client-side routing, add this field to `koru.json`:

```json
"spa": true
```

Routes such as `/settings` will open `index.html`, letting your router choose the screen. Missing files with extensions, such as `/assets/app.js`, still return 404.

For a site with several HTML files, leave `spa` disabled and link to the included pages. Configure your frontend to serve from the root: `/assets/app.js` should match `www/assets/app.js`, without a deployment prefix or domain.

## Connect your API

Packaged files are served from `https://app.koru.invalid`. To call external services, add the Internet permission to `permissions`:

```json
"permissions": ["android.permission.INTERNET"]
```

Use HTTPS URLs and allow `https://app.koru.invalid` in your server’s CORS settings. This origin is local to the app: you do not need a domain or DNS configuration. Bundle your JavaScript dependencies; remote scripts and iframes are blocked.

## Build and test the APK

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
```

Install the APK on Android and check the first screen, routes, Back button and network requests. After changing your frontend, build it again and rebuild the APK.

Continue with [Android features from JavaScript]({{base}}en/plugins/) or [distribute your app]({{base}}en/distribution/).
