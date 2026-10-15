# JavaScript API

Use Android features from your app’s JavaScript: read device information, store data, share text or show notifications. Calls return promises and work inside the APK installed on Android.

## Before you call an API

The generated `koru.json` enables `device`, `storage` and `app`. To use the other plugins, add their IDs and rebuild:

```json
"plugins": ["device", "storage", "app", "vibration", "share", "notifications"]
```

This is one field to put in your existing `koru.json`, not a complete file. Enabling `vibration` adds `android.permission.VIBRATE`; `notifications` adds `android.permission.POST_NOTIFICATIONS`. In a packaged HTML page, import the generated module:

```html
<script type="module">
  import { device } from "/__koru/core.mjs";
  document.querySelector("#model").textContent = (await device.getInfo()).model;
</script>
```

Add an element with `id="model"` to that page. Koru also injects `/__koru/core.js`, which exposes `window.Koru`. The packaged module exports `device`, `app`, `storage`, `vibration`, `share`, `notifications` and `invoke`. If your frontend bundles the JavaScript package from `packages/javascript`, build it before calling koru; the browser-facing examples below use the generated `/__koru/core.mjs` path. All methods return a Promise. Wrap calls in `try/catch` to show a useful error:

```js
try {
  const info = await device.getInfo();
  console.log(info.model);
} catch (error) {
  console.error(error.code, error.message);
}
```

## Method index

| Plugin           | Methods                                                                                                                                                                        | Enabled by `init`?          |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------- |
| Device           | [`getInfo()`](#api-device-info)                                                                                                                                                | Yes                         |
| App              | [`getInfo()`](#api-app-info)                                                                                                                                                   | Yes                         |
| Storage          | [`get()`](#api-storage-get), [`set()`](#api-storage-set), [`remove()`](#api-storage-remove), [`clear()`](#api-storage-clear)                                                   | Yes                         |
| Vibration        | [`vibrate()`](#api-vibration)                                                                                                                                                  | No                          |
| Share            | [`text()`](#api-share)                                                                                                                                                         | No                          |
| Notifications    | [`permission()`](#api-notification-permission), [`requestPermission()`](#api-notification-request), [`show()`](#api-notification-show), [`cancel()`](#api-notification-cancel) | No                          |
| Low-level bridge | [`invoke()`](#api-invoke)                                                                                                                                                      | Available to packaged pages |

Each method below opens independently. The examples assume imports from `/__koru/core.mjs` and run only inside the APK.

## Device and app

<details class="api-detail" id="api-device-info"><summary><code>device.getInfo()</code><span>Read the Android device</span></summary>

**Signature:** `device.getInfo(options?) → Promise<DeviceInfo>`. No native parameters or permissions.

**Result:** `{ model: string, manufacturer: string, osVersion: string, sdk: number, platform: "android" }`. The values come from Android `Build` fields and differ by device; `sdk` is its API level.

```js
import { device } from "/__koru/core.mjs";
const info = await device.getInfo();
console.log(
  `${info.manufacturer} ${info.model} · Android ${info.osVersion} (API ${info.sdk})`,
);
```

`options` may contain the common `timeout` and `signal` described under `invoke()`. A disabled plugin rejects with `METHOD_NOT_FOUND`; channel and timeout failures can also reject.

</details>

<details class="api-detail" id="api-app-info"><summary><code>app.getInfo()</code><span>Read the installed app identity</span></summary>

**Signature:** `app.getInfo(options?) → Promise<AppInfo>`. No native parameters or permissions.

**Result:** `{ appId: string, name: string, version: string, versionCode: number }`, from the `koru.json` used to build this APK.

```js
import { app } from "/__koru/core.mjs";
const info = await app.getInfo();
document.querySelector("#version").textContent = `${info.name} ${info.version}`;
```

The object describes the packaged app. It does not query an app store or available updates.

</details>

## Private storage

Storage uses the app's private Android `SharedPreferences`. Keys are 1–128 characters from `A–Z`, `a–z`, `0–9`, `_`, `.`, `:` and `-`; values are strings. Data persists across app restarts and updates signed with the same identity. It is removed when the app is uninstalled. Do not put secrets in a web page or assume storage is encrypted.

<details class="api-detail" id="api-storage-get"><summary><code>storage.get(key)</code><span>Read a string or null</span></summary>

**Signature:** `storage.get(key: string, options?) → Promise<string | null>`.

**Arguments:** `key` is a valid storage key. **Result:** the stored string, or `null` when the key has never been set or was removed.

```js
import { storage } from "/__koru/core.mjs";
const theme = (await storage.get("theme")) ?? "system";
console.log(theme);
```

A missing, wrong-type or invalid key rejects with `INVALID_REQUEST`. This wrapper unwraps the native `{ value }` object for you.

</details>

<details class="api-detail" id="api-storage-set"><summary><code>storage.set(key, value)</code><span>Persist a string</span></summary>

**Signature:** `storage.set(key: string, value: string, options?) → Promise<{}>`.

**Arguments:** a valid key and a string value. The operation uses a synchronous Android commit before the Promise resolves. **Result:** an empty object `{}`.

```js
import { storage } from "/__koru/core.mjs";
await storage.set("theme", "dark");
await storage.set("visits", String(3)); // convert numbers to strings
```

Invalid key/value types reject with `INVALID_REQUEST`; an unsuccessful commit rejects with `STORAGE_ERROR`.

</details>

<details class="api-detail" id="api-storage-remove"><summary><code>storage.remove(key)</code><span>Delete one key</span></summary>

**Signature:** `storage.remove(key: string, options?) → Promise<{}>`. **Result:** an empty object after Android commits the removal. Removing a key that is absent is safe.

```js
import { storage } from "/__koru/core.mjs";
await storage.remove("theme");
console.log(await storage.get("theme")); // null
```

Invalid keys reject with `INVALID_REQUEST`; a failed commit rejects with `STORAGE_ERROR`.

</details>

<details class="api-detail" id="api-storage-clear"><summary><code>storage.clear()</code><span>Delete all koru storage keys</span></summary>

**Signature:** `storage.clear(options?) → Promise<{}>`. It clears all keys in koru's private storage namespace for this app, then resolves with `{}`. It does not erase packaged web assets or unrelated Android app data.

```js
import { storage } from "/__koru/core.mjs";
await storage.clear();
```

A failed commit rejects with `STORAGE_ERROR`. Use this only when the user intends to clear every key.

</details>

## Device actions

<details class="api-detail" id="api-vibration"><summary><code>vibration.vibrate(duration?)</code><span>Vibrate for milliseconds</span></summary>

**Setup:** enable `vibration` in `koru.json`; koru adds `android.permission.VIBRATE` to the APK.

**Signature:** `vibration.vibrate(duration = 100, options?) → Promise<{}>`. `duration` must be an integer from 1 to 10,000 milliseconds. The Promise resolves with `{}` after starting the Android action; it does not wait for the vibration to finish.

```js
import { vibration } from "/__koru/core.mjs";
await vibration.vibrate(150);
```

A wrong duration rejects with `INVALID_REQUEST`. Hardware or Android failures may reject with `NATIVE_ERROR`; a device without a vibrator cannot provide tactile feedback.

</details>

<details class="api-detail" id="api-share"><summary><code>share.text(text)</code><span>Open Android's share sheet</span></summary>

**Setup:** enable `share` in `koru.json`. No added permission.

**Signature:** `share.text(text: string, options?) → Promise<{}>`. `text` is the string to share. Koru opens an Android `ACTION_SEND` chooser with MIME type `text/plain` and resolves with `{}` after launching it. Resolution does not mean the user selected a destination or completed a share.

```js
import { share } from "/__koru/core.mjs";
await share.text("Try my koru app");
```

Non-string text rejects with `INVALID_REQUEST`; failure to open the chooser may reject with `NATIVE_ERROR`.

</details>

## Local notifications

Enable `notifications` in `koru.json` and rebuild. Koru adds `android.permission.POST_NOTIFICATIONS`. Android 13/API 33+ may require a runtime decision; older supported Android versions use the system notification setting. The API 33+ dialog has not been exercised in the recorded API 30 matrix. Notifications are immediate, local, and use koru's `koru-local` channel; this API does not schedule future alarms or remote push.

<details class="api-detail" id="api-notification-permission"><summary><code>notifications.permission()</code><span>Read notification status</span></summary>

**Signature:** `notifications.permission(options?) → Promise<{ status: "granted" | "denied" }>`.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.permission();
console.log(status); // "granted" or "denied"
```

This reads the current Android permission and notification setting. It does not show a prompt. A disabled plugin rejects with `METHOD_NOT_FOUND`.

</details>

<details class="api-detail" id="api-notification-request"><summary><code>notifications.requestPermission()</code><span>Ask for notification access</span></summary>

**Signature:** `notifications.requestPermission(options?) → Promise<{ status: "granted" | "denied" }>`.

On Android 13/API 33+, a denied status opens the runtime permission prompt. On Android 8–12/API 26–32, it opens app notification settings and resolves when the user returns. If access was already granted, it resolves immediately. The wrapper's default timeout is **120,000 ms** for this method; `options.timeout` can override it up to that limit. Only one request can be pending.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.requestPermission();
if (status === "granted") console.log("Notifications enabled");
```

`BUSY` means another request is pending; `TIMEOUT` means the decision did not arrive in time. A returned `denied` status is a normal result, not a thrown error.

</details>

<details class="api-detail" id="api-notification-show"><summary><code>notifications.show({ id, title, body })</code><span>Post a notification now</span></summary>

**Signature:** `notifications.show({ id: number, title: string, body: string }, options?) → Promise<{}>`.

`id` is an integer notification identifier; use the same ID to replace or cancel a notification. The native default is `1` if `id` is omitted, but passing it explicitly makes later cancellation clear. `title` and `body` are required strings. The Promise resolves with `{}` after posting the notification, not after the user opens it.

```js
import { notifications } from "/__koru/core.mjs";
const { status } = await notifications.permission();
if (status === "granted") {
  await notifications.show({
    id: 7,
    title: "Ready",
    body: "Your export finished.",
  });
}
```

If notifications are disabled or the Android 13+ permission is absent, it rejects with `PERMISSION_DENIED`. Wrong parameter types reject with `INVALID_REQUEST`.

</details>

<details class="api-detail" id="api-notification-cancel"><summary><code>notifications.cancel(id)</code><span>Remove one notification</span></summary>

**Signature:** `notifications.cancel(id: number, options?) → Promise<{}>`. Pass the same integer ID used for `show`; it resolves with `{}` after asking Android to cancel that notification. An absent ID is safe to cancel. No permission prompt is shown.

```js
import { notifications } from "/__koru/core.mjs";
await notifications.cancel(7);
```

A non-integer ID rejects with `INVALID_REQUEST`. A native Android failure may reject with `NATIVE_ERROR`.

</details>

## Low-level bridge and errors

<details class="api-detail" id="api-invoke"><summary><code>invoke(method, params?, options?)</code><span>Call a native handler directly</span></summary>

**Signature:** `invoke(method: string, params = {}, options = {}) → Promise<object>`. It is the exported function in `/__koru/core.mjs`; `window.Koru.invoke` is the equivalent global. The method is a string such as `"device.info"`; `params` must be a JSON-serializable plain object. Use this for a custom plugin or when you need the unwrapped native result.

```js
import { invoke } from "/__koru/core.mjs";
const info = await invoke("device.info", {});
const raw = await invoke("storage.get", { key: "theme" });
console.log(info.model, raw.value); // raw.value is string or null
```

`options.timeout` is a finite number of milliseconds from 1 to 120,000; default 10,000. The timer also covers waiting for the native channel. `options.signal` accepts an `AbortSignal`; a cancelled request rejects with `CANCELLED`. Cancellation prevents an unsent request, but cannot undo a native action already dispatched. At most 256 calls can be pending. A request's serialized JSON must stay below 65,536 characters.

**Result:** the native result object. Built-in action methods return `{}`; the wrapper `storage.get()` converts the raw `{ value }` result into a string or `null`. Custom handlers define their own result object.

</details>

Errors reject with an `Error` whose `code` and `message` can be displayed or logged. `INVALID_REQUEST` means invalid method/arguments; `METHOD_NOT_FOUND` means the method or plugin is unavailable; `STORAGE_ERROR` means a storage commit failed; `PERMISSION_DENIED` means a required notification permission or setting is off; `NATIVE_ERROR` means Android threw an error. `BUSY`, `TIMEOUT`, `CANCELLED`, `CHANNEL_ERROR`, `CHANNEL_REPLACED`, `INVALID_RESPONSE` and `PAGE_CLOSED` describe the bridge lifecycle. A timeout or cancellation does not prove an already dispatched native action was rolled back.

## Extend your app

To add your own features, see the [native adapter]({{base}}en/native/) and the [custom plugin reference]({{base}}original/docs/plugins.md).

[Configuration]({{base}}en/configuration/) · [Troubleshooting]({{base}}en/troubleshooting/)
