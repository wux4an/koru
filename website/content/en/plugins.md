# JavaScript and plugins

The builder integrates `/__koru/core.js` and `/__koru/core.mjs`, and includes the first in HTML pages. You can also consume the API through `@koru/core` by adding `packages/javascript` to the project that builds your static assets.

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
await share.text("Hello");
await notifications.requestPermission({ timeout: 120000 });
await notifications.show({
  id: 1,
  title: "Reminder",
  body: "Local notification",
});
```

## Built-in methods

| Plugin        | Methods                                                  | Added permissions                                             |
| ------------- | -------------------------------------------------------- | ------------------------------------------------------------- |
| device        | `device.info`                                            | None; model, manufacturer, OS and SDK.                        |
| app           | `app.info`                                               | None.                                                         |
| storage       | `storage.get/set/remove/clear`                           | None; private SharedPreferences, bounded keys, string values. |
| vibration     | `vibration.vibrate`                                      | VIBRATE; integer duration from 1 to 10,000 ms.                |
| share         | `share.text`                                             | None; ACTION_SEND text/plain chooser.                         |
| notifications | `notifications.show/cancel/permission/requestPermission` | POST_NOTIFICATIONS; local channel from API 26.                |

You can invoke these methods with `Koru.invoke(method, params, { timeout, signal })` or the exported `invoke` function. The module's `storage.get` returns the stored string or `null` when absent.

`notifications.show` publishes immediately, rather than scheduling a future alarm. `requestPermission` waits for the decision or return from settings. Android 33+ requests the runtime permission; Android 26–32 checks system notification controls. The API respects restrictions imposed by the user. The API 33+ dialog has not been tested in the recorded API 30 matrix.

## Errors and cancellation

Cancellation rejects the Promise and prevents a request that has not yet been sent. It does not undo writes, vibration, notifications or UI actions already executed. Main codes: `INVALID_REQUEST`, `METHOD_NOT_FOUND`, `NATIVE_ERROR`, `PERMISSION_DENIED`, `STORAGE_ERROR`, `BUSY`, `TIMEOUT`, `CANCELLED`, `CHANNEL_ERROR` and `PAGE_CLOSED`.

## Custom plugin

Declare `pluginManifests` in `koru.json` and enable the manifest's `id` in `plugins`. A manifest contains ABI=1, permissions, DEX IR classes and a handler map. Each handler is public static with descriptor:

```text
(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;
```

The dispatcher validates the protocol, calls the handler and converts exceptions into structured errors. The plugin must validate its parameters and respect the UI thread. Its classes cannot replace runtime or framework classes. It is trusted, signed code with access to the application's permissions.

`examples/native/plugin.json` uses a DEX class whose `clinit` calls `System.loadLibrary` and a native method. It contains no Java. Integration:

```json
{
  "plugins": ["device", "native"],
  "pluginManifests": ["plugins/native/plugin.json"],
  "nativeLibraries": { "x86_64": ["plugins/native/libkoru_native.so"] }
}
```

This is a fragment to add to the complete configuration, not a complete `koru.json`. Libraries are validated as ELF shared objects for the configured ABI; koru does not compile them during build. The sample binary is x86_64 and was tested on Android 11. ARM phones need a library for their ABI. Apps using the six standard plugins require no native libraries and impose no phone architecture dependency.

## JNI ownership

The C code and JNI scope API live in `runtime/native`. Each scope owns the current thread's JNIEnv; detach occurs only when that scope attached. Local references do not escape their frame. Shared class references are global and released in `JNI_OnUnload`. `scope_test.c` checks attach/detach ownership using a controlled JNI table; the real plugin was additionally executed in ART.

[Optional JNI adapter]({{base}}en/native/) · [Bridge security]({{base}}en/security/)
