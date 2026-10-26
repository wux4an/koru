# Write a plugin

A plugin adds your own native method to the app, such as `toast.show`, that JavaScript can call with `invoke()`. This guide builds one real example, the `toast` plugin in [`examples/toast`]({{base}}original/examples/toast/README.md), step by step. Read [JavaScript API]({{base}}en/plugins/) first for `invoke()` and the error codes.

## What a plugin is

- **Trusted code inside your app.** It runs with the app's Android permissions and can do what the app can do. Include only code you wrote or reviewed. It is not sandboxed from the app.
- **No Java or Kotlin.** koru has no JVM, `javac`, `d8` or Gradle. A plugin is a JSON description of DEX classes: the _IR_ (intermediate representation) that koru's Go assembler turns into `classes.dex`. You write Dalvik instructions, one JSON object per instruction. See the [DEX API]({{base}}en/dex-api/) for every supported operation.
- **A handler is a static method.** Each JavaScript method name, such as `toast.show`, maps to one `public static` method with a fixed signature. koru's dispatcher calls it and wraps the result for JavaScript.

Writing IR by hand is low level. Keep handlers short, and generate the JSON with a script if you prefer.

## 1. The manifest

A plugin is one JSON file with an ID, the ABI version, permissions, a handler map and the classes. This is the outline of [`plugin.json`]({{base}}original/examples/toast/plugin.json ":ignore"):

```json
{
  "id": "toast",
  "abi": 1,
  "permissions": [],
  "handlers": {
    "toast.show": {
      "owner": "Lcom/example/toast/ToastPlugin;",
      "name": "show",
      "descriptor": "(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;"
    }
  },
  "classes": [{ "descriptor": "Lcom/example/toast/ToastPlugin;", "...": "..." }]
}
```

The `...` stands for the rest of the class. The rules koru checks when it loads the file:

- `id`: lowercase letters, digits and `_`, starting with a letter, at most 32 characters. It cannot be a built-in ID (`device`, `app`, `storage`, `vibration`, `share`, `notifications`).
- `abi`: must be `1`.
- `permissions`: Android permissions the plugin needs, such as `android.permission.VIBRATE`. Koru adds them to the manifest of the APK. `toast` needs none.
- `handlers`: every key must start with the plugin ID and a dot (`toast.show`), and the method part must start with a letter. Each value points to a `public static` method of a `public` class in `classes`.
- `classes`: your classes. They cannot start with `Landroid/`, `Ljava/` or `Lcom/koru/runtime/`, so they cannot replace the framework or the runtime. Use your own package, like `Lcom/example/toast/`.

## 2. The handler signature

Every handler has exactly this descriptor, and `Validate` rejects any other:

```
(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;
```

In Java terms: `public static JSONObject show(Activity activity, JSONObject params)`. The Activity gives you a `Context`. `params` is the object JavaScript passed to `invoke()`; it is never `null`. The return value is the result object sent back to JavaScript.

Class and method flags are plain integers: `Public = 1`, `Static = 8`, so the handler has `"access": 9` and the class has `"access": 1`.

## 3. Registers: how the example reads its code

Dalvik has no local variables, only numbered registers. The method declares `"locals": 5`, so `v0` to `v4` are scratch registers. Parameters are written with negative numbers: `-1` is the first parameter word (`activity`) and `-2` is the second (`params`). That is `P(0)` and `P(1)` in the Go API; the assembler places them after the locals. Static methods have no `this`.

An object or a 32-bit value (`int`, `boolean`) takes one register. Jumps use labels: `{"op": "label", "label": "show"}` defines one, and `"target": "show"` jumps to it.

## 4. Read the parameters

`text` must be a non-empty string. First, read the raw value with `JSONObject.opt`, which returns `null` for a missing key instead of throwing:

```json
{ "op": "const-string", "regs": [0], "string": "text" },
{
  "op": "invoke-virtual",
  "regs": [-2, 0],
  "method": {
    "owner": "Lorg/json/JSONObject;",
    "name": "opt",
    "descriptor": "(Ljava/lang/String;)Ljava/lang/Object;"
  }
},
{ "op": "move-result-object", "regs": [1] }
```

`invoke-virtual` takes the receiver first (`-2`, the params), then the arguments (`v0`, the key). A call never writes its result to a register by itself: the next instruction, `move-result-object` (or `move-result` for `int` and `boolean`), copies it to `v1`.

Then check the type. `instance-of` writes `1` or `0` into `v2` (`null` is not an instance), and `if-eqz` jumps when that is `0`:

```json
{ "op": "instance-of", "regs": [2, 1], "type": "Ljava/lang/String;" },
{ "op": "if-eqz", "regs": [2], "target": "invalid" },
{ "op": "check-cast", "regs": [1], "type": "Ljava/lang/String;" }
```

`check-cast` tells the runtime that `v1` is now a `String`, so that it can call String methods. The example then calls `trim()` and `isEmpty()` and jumps to `invalid` with `if-nez` when the text is blank.

The optional `long` flag follows the same pattern. `v3` starts as `0` (`Toast.LENGTH_SHORT`). If `long` is missing, the code jumps to `show`; if it has another type, to `invalid`; otherwise it unboxes the `Boolean` and sets `v3` to `1` (`LENGTH_LONG`) when it is `true`:

```json
{ "op": "const", "regs": [3], "literal": 0 },
...
{
  "op": "invoke-virtual",
  "regs": [2],
  "method": {
    "owner": "Ljava/lang/Boolean;",
    "name": "booleanValue",
    "descriptor": "()Z"
  }
},
{ "op": "move-result", "regs": [4] },
{ "op": "if-eqz", "regs": [4], "target": "show" },
{ "op": "const", "regs": [3], "literal": 1 }
```

The `...` stands for the `opt("long")` lookup and the type check, which repeat the pattern above with `Boolean`.

JavaScript numbers arrive as `Integer`, `Long` or `Double` values, strings as `String`, and booleans as `Boolean`. Always check the type before casting: a wrong cast throws.

## 5. Do the work and return

With the arguments ready, call the Android API. `Toast.makeText` is a static method, so there is no receiver, only the arguments `activity` (a `Context`), `v1` (a `CharSequence`; a `String` is one) and `v3`:

```json
{
  "op": "invoke-static",
  "regs": [-1, 1, 3],
  "method": {
    "owner": "Landroid/widget/Toast;",
    "name": "makeText",
    "descriptor": "(Landroid/content/Context;Ljava/lang/CharSequence;I)Landroid/widget/Toast;"
  }
},
{ "op": "move-result-object", "regs": [0] },
{
  "op": "invoke-virtual",
  "regs": [0],
  "method": {
    "owner": "Landroid/widget/Toast;",
    "name": "show",
    "descriptor": "()V"
  }
}
```

A method descriptor lists parameter types in parentheses and the return type after them: `I` is `int`, `Z` is `boolean`, `V` is `void`, and `L...;` is a class. The result is a new empty `JSONObject`:

```json
{ "op": "new-instance", "regs": [0], "type": "Lorg/json/JSONObject;" },
{
  "op": "invoke-direct",
  "regs": [0],
  "method": {
    "owner": "Lorg/json/JSONObject;",
    "name": "<init>",
    "descriptor": "()V"
  }
},
{ "op": "return-object", "regs": [0] }
```

To return data, call `JSONObject.put(String, Object)` or `put(String, int)` on that object before `return-object`. JavaScript receives it unchanged from `invoke()`.

## 6. Errors: what the dispatcher does with your handler

The dispatcher already validates the request envelope and calls your handler inside a `try`. You have two ways to fail:

| Handler does           | JavaScript sees                             |
| ---------------------- | ------------------------------------------- |
| Returns `null`         | `INVALID_REQUEST`, "Invalid native request" |
| Throws any `Throwable` | `NATIVE_ERROR`, "Android operation failed"  |
| Returns an object      | The promise resolves with that object       |

The example documents and uses the first: when `text` is empty or not a string, or `long` is not a boolean, it jumps to a label that returns `null`:

```json
{ "op": "label", "label": "invalid" },
{ "op": "const", "regs": [0], "literal": 0 },
{ "op": "return-object", "regs": [0] }
```

(A zero register is `null`.) Handlers cannot choose their own error code or message; the dispatcher only produces these. If you need rich failure information, return a result such as `{ "ok": false, "reason": "..." }` and let your JavaScript wrapper reject.

## 7. Which thread runs the handler

The Android `Toast.makeText` call needs a thread with a `Looper`, which in practice is the UI thread. koru's runtime creates a `WebMessagePort` and calls `setWebMessageCallback(callback)` without a `Handler`; Android's reference says that form delivers messages on the UI thread. The dispatcher runs inside that callback, so your handler runs on the **UI thread**, with no `runOnUiThread` or extra `Runnable` class needed. The integrated `notifications.requestPermission` relies on the same behavior when it calls `requestPermissions`.

Because of that, a handler **blocks the UI** while it runs. Keep it short and do not do network or disk work in it. This reading of the Android documentation has not been confirmed on a device in this project; if it were wrong, `Toast.makeText` would likely throw and you would see `NATIVE_ERROR`.

## 8. Enable it in koru.json

Save the file as `plugins/toast/plugin.json` in your project and add two fields to your existing `koru.json` (a fragment, not a whole file):

```json
{
  "plugins": ["device", "storage", "app", "toast"],
  "pluginManifests": ["plugins/toast/plugin.json"]
}
```

The ID in `plugins` must match a manifest's `id`; otherwise the build stops with `plugin "toast" has no manifest`. `koru plugin add` and `koru plugin remove` only handle integrated plugins, so edit these fields by hand. `koru plugin list` shows `toast` as `on` once configured. The permissions listed by the manifest are added to the APK automatically.

## 9. Call it from JavaScript

Custom methods have no wrapper, so use `invoke()`. Give it a typed function of your own:

```ts
import { invoke, type CallOptions } from "/__koru/core.mjs";

export interface ToastParams {
  text: string;
  long?: boolean;
}

export function showToast(params: ToastParams, options?: CallOptions) {
  return invoke<Record<string, never>>("toast.show", { ...params }, options);
}
```

Use it like any other call and handle failures with the error codes:

```js
try {
  await invoke("toast.show", { text: "Hello", long: true });
} catch (error) {
  console.warn(error.code, error.message); // INVALID_REQUEST, NATIVE_ERROR...
}
```

The complete page is [`examples/toast/www/index.html`]({{base}}original/examples/toast/www/index.html ":ignore").

## 10. Test it

```sh
koru build --output app.apk
koru verify app.apk
unzip -p app.apk classes.dex > classes.dex
koru dex inspect classes.dex
koru run
```

`koru build` loads and validates every manifest, so structural mistakes fail here, with the DEX assembler's message. `koru verify` checks the signature and that `classes.dex` parses. `koru dex inspect` prints the decoded types, methods and strings as JSON: look for `Lcom/example/toast/ToastPlugin;`, its `show` method and the route string `toast.show`. `koru run` installs and opens the app on a connected device.

The repository test `internal/app/toast_example_test.go` automates the first three steps for the example.

## Limits you should know

- **The `toast` plugin has not been run on Android.** This environment had no device. The tests prove that the manifest is valid, that the APK builds and verifies, and that the class and method are in `classes.dex`. They do not prove that ART accepts the bytecode, that the toast appears, or the thread used. Run `koru run` on a phone or emulator before relying on it. dexdump and koru's parser do not replace that.
- **No web version.** Under `koru dev` and in a browser, `invoke("toast.show")` rejects with `METHOD_NOT_FOUND`, because only the integrated plugins have web implementations. Guard the call:

  ```js
  import { invoke, isNative } from "/__koru/core.mjs";

  async function notify(text) {
    if (isNative()) await invoke("toast.show", { text });
    else console.info(text); // or show your own in-page message
  }
  ```

- **Errors are generic.** `INVALID_REQUEST` and `NATIVE_ERROR` carry fixed messages (section 6).
- **Hand-written IR is easy to get wrong.** The assembler checks structure, labels, references and jump targets, but not every register type. A mistake that passes the build can still crash when ART verifies the class.
- **Trusted code only.** A plugin shares the app's permissions. A page you load can call any enabled handler, so validate every parameter. For native `.so` code, see [Optional JNI adapter]({{base}}en/native/).

[JavaScript API]({{base}}en/plugins/) · [DEX API]({{base}}en/dex-api/) · [Configuration]({{base}}en/configuration/)
