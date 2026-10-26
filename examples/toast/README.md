# Toast plugin (custom plugin example)

A complete custom koru plugin written only as DEX IR JSON: no Java, no JVM.
`plugin.json` defines the plugin `toast` (ABI 1, no permissions) and one handler,
`toast.show`, implemented by `Lcom/example/toast/ToastPlugin;.show`.

```js
import { invoke } from "/__koru/core.mjs";
await invoke("toast.show", { text: "Hello", long: true });
```

Parameters: `text` (non-empty string, required) and `long` (boolean, optional;
`true` = `Toast.LENGTH_LONG`, otherwise `LENGTH_SHORT`). It resolves with `{}`.
Invalid parameters make the handler return `null`; the dispatcher turns that into
an `INVALID_REQUEST` error. If Android throws, the dispatcher reports
`NATIVE_ERROR`.

Threading: the dispatcher runs on the thread that receives the `WebMessagePort`
message. The runtime registers the callback without a `Handler`, so Android delivers
it on the UI (main) thread, which `Toast.makeText` requires. This follows the
Android API reference and has not been checked on a device.

## Use it

In a project created with `koru init`, copy `plugin.json` to `plugins/toast/` and
`www/index.html` over your `www/index.html`, then edit `koru.json`:

```json
{
  "plugins": ["device", "storage", "app", "toast"],
  "pluginManifests": ["plugins/toast/plugin.json"]
}
```

(These two fields go into your existing `koru.json`; `koru plugin add` only
knows the integrated plugins, so edit the file by hand.) Then:

```sh
koru build --output app.apk
koru verify app.apk
unzip -p app.apk classes.dex > classes.dex && koru dex inspect classes.dex
koru run
```

## Status

`internal/app/toast_example_test.go` loads and validates the manifest, builds a
complete APK and checks the class and method in `classes.dex`. The plugin has
**not** been run on Android: there was no device. Under `koru dev` a custom
plugin has no web implementation, so `invoke("toast.show")` rejects with
`METHOD_NOT_FOUND`.
