# __APP_NAME__

A [Vite](https://vite.dev/) app packaged for Android with [koru](https://koru.js.org/).

```sh
npm install
npm run dev        # Vite with hot reload; the koru API uses web implementations
npm run android    # build the web app, then install and open it on a device
npm run apk        # signed release APK
```

To check that your code only calls plugins enabled in `koru.json`, run
`koru dev --proxy http://localhost:5173` in a second terminal and open the
address it prints.

`koru.json` points `webDir` at `dist`, the folder `vite build` writes.
`src/koru/` is a copy of the koru JavaScript module and its types. After
upgrading koru, refresh it with `koru js src/koru --force`.

`.koru/signing.pem` is your signing key. Back it up and never commit it.
