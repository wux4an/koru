# Escribe un plugin

Un plugin añade un método nativo propio a la app, como `toast.show`, que JavaScript puede llamar con `invoke()`. Esta guía construye un ejemplo real, el plugin `toast` de [`examples/toast`]({{base}}original/examples/toast/README.md), paso a paso. Lee antes la [API JavaScript]({{base}}docs/es/plugins/) para conocer `invoke()` y los códigos de error.

## Qué es un plugin

- **Código de confianza dentro de tu app.** Se ejecuta con los permisos de Android de la app y puede hacer lo que ella hace. Incluye solo código que hayas escrito o revisado. No está aislado de la app.
- **Sin Java ni Kotlin.** koru no tiene JVM, `javac`, `d8` ni Gradle. Un plugin es una descripción JSON de clases DEX: el _IR_ (representación intermedia) que el ensamblador en Go de koru convierte en `classes.dex`. Escribes instrucciones Dalvik, un objeto JSON por instrucción. Consulta la [API DEX]({{base}}docs/es/dex-api/) para ver todas las operaciones admitidas.
- **Un handler es un método estático.** Cada nombre de método de JavaScript, como `toast.show`, corresponde a un método `public static` con una firma fija. El dispatcher de koru lo llama y empaqueta el resultado para JavaScript.

Escribir IR a mano es de bajo nivel. Mantén los handlers cortos y, si prefieres, genera el JSON con un script.

## 1. El manifiesto

Un plugin es un archivo JSON con un ID, la versión de ABI, permisos, un mapa de handlers y las clases. Este es el esquema de [`plugin.json`]({{base}}original/examples/toast/plugin.json ":ignore"):

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

Los `...` representan el resto de la clase. Reglas que koru comprueba al cargar el archivo:

- `id`: minúsculas, dígitos y `_`, empieza por letra, máximo 32 caracteres. No puede ser un ID integrado (`device`, `app`, `storage`, `vibration`, `share`, `notifications`).
- `abi`: debe ser `1`.
- `permissions`: permisos de Android que necesita el plugin, como `android.permission.VIBRATE`. koru los añade al manifiesto del APK. `toast` no necesita ninguno.
- `handlers`: cada clave empieza por el ID del plugin y un punto (`toast.show`), y la parte del método empieza por una letra. Cada valor apunta a un método `public static` de una clase `public` de `classes`.
- `classes`: tus clases. No pueden empezar por `Landroid/`, `Ljava/` ni `Lcom/koru/runtime/`, así que no pueden reemplazar el framework ni el runtime. Usa tu propio paquete, como `Lcom/example/toast/`.

## 2. La firma del handler

Todos los handlers tienen exactamente este descriptor; `Validate` rechaza cualquier otro:

```
(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;
```

En términos de Java: `public static JSONObject show(Activity activity, JSONObject params)`. La Activity te da un `Context`. `params` es el objeto que JavaScript pasó a `invoke()`; nunca es `null`. El valor devuelto es el objeto de resultado que recibe JavaScript.

Los flags de clases y métodos son enteros: `Public = 1`, `Static = 8`; por eso el handler tiene `"access": 9` y la clase `"access": 1`.

## 3. Registros: cómo leer el código del ejemplo

Dalvik no tiene variables locales, solo registros numerados. El método declara `"locals": 5`, así que `v0` a `v4` son registros de trabajo. Los parámetros se escriben con números negativos: `-1` es la primera palabra de parámetro (`activity`) y `-2` la segunda (`params`). Son `P(0)` y `P(1)` en la API de Go; el ensamblador los coloca después de los locales. Los métodos estáticos no tienen `this`.

Un objeto o un valor de 32 bits (`int`, `boolean`) ocupa un registro. Los saltos usan etiquetas: `{"op": "label", "label": "show"}` define una y `"target": "show"` salta a ella.

## 4. Leer los parámetros

`text` debe ser un string no vacío. Primero se lee el valor sin procesar con `JSONObject.opt`, que devuelve `null` si falta la clave en lugar de lanzar una excepción:

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

`invoke-virtual` toma primero el receptor (`-2`, los params) y después los argumentos (`v0`, la clave). Una llamada nunca escribe su resultado en un registro por sí sola: la instrucción siguiente, `move-result-object` (o `move-result` para `int` y `boolean`), lo copia a `v1`.

Después se comprueba el tipo. `instance-of` escribe `1` o `0` en `v2` (`null` no es instancia de nada) y `if-eqz` salta cuando vale `0`:

```json
{ "op": "instance-of", "regs": [2, 1], "type": "Ljava/lang/String;" },
{ "op": "if-eqz", "regs": [2], "target": "invalid" },
{ "op": "check-cast", "regs": [1], "type": "Ljava/lang/String;" }
```

`check-cast` indica al runtime que `v1` ya es un `String`, para poder llamar a métodos de String. El ejemplo llama luego a `trim()` y a `isEmpty()`, y salta a `invalid` con `if-nez` si el texto está en blanco.

La opción `long` sigue el mismo patrón. `v3` empieza en `0` (`Toast.LENGTH_SHORT`). Si falta `long`, salta a `show`; si tiene otro tipo, a `invalid`; en otro caso extrae el `Boolean` y pone `v3` en `1` (`LENGTH_LONG`) cuando es `true`:

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

Los `...` representan la lectura de `opt("long")` y la comprobación de tipo, que repiten el patrón anterior con `Boolean`.

Los números de JavaScript llegan como `Integer`, `Long` o `Double`, los strings como `String` y los booleanos como `Boolean`. Comprueba siempre el tipo antes de convertir: una conversión incorrecta lanza una excepción.

## 5. Hacer el trabajo y devolver

Con los argumentos listos, llama a la API de Android. `Toast.makeText` es un método estático, así que no hay receptor, solo los argumentos `activity` (un `Context`), `v1` (un `CharSequence`; un `String` lo es) y `v3`:

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

Un descriptor de método lista los tipos de parámetros entre paréntesis y el tipo de retorno después: `I` es `int`, `Z` es `boolean`, `V` es `void` y `L...;` es una clase. El resultado es un `JSONObject` nuevo y vacío:

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

Para devolver datos, llama a `JSONObject.put(String, Object)` o `put(String, int)` sobre ese objeto antes de `return-object`. JavaScript lo recibe sin cambios desde `invoke()`.

## 6. Errores: qué hace el dispatcher con tu handler

El dispatcher ya valida el sobre de la petición y llama a tu handler dentro de un `try`. Tienes dos formas de fallar:

| El handler                  | JavaScript ve                               |
| --------------------------- | ------------------------------------------- |
| Devuelve `null`             | `INVALID_REQUEST`, "Invalid native request" |
| Lanza cualquier `Throwable` | `NATIVE_ERROR`, "Android operation failed"  |
| Devuelve un objeto          | La promesa se resuelve con ese objeto       |

El ejemplo documenta y usa la primera: cuando `text` está vacío o no es un string, o `long` no es booleano, salta a una etiqueta que devuelve `null`:

```json
{ "op": "label", "label": "invalid" },
{ "op": "const", "regs": [0], "literal": 0 },
{ "op": "return-object", "regs": [0] }
```

(Un registro con cero es `null`.) Un handler no puede elegir su propio código ni mensaje de error; solo el dispatcher los produce. Si necesitas más información sobre el fallo, devuelve un resultado como `{ "ok": false, "reason": "..." }` y haz que tu envoltorio de JavaScript rechace la promesa.

## 7. Qué hilo ejecuta el handler

`Toast.makeText` necesita un hilo con `Looper`, que en la práctica es el hilo de interfaz. El runtime de koru crea un `WebMessagePort` y llama a `setWebMessageCallback(callback)` sin `Handler`; la referencia de Android indica que esa forma entrega los mensajes en el hilo de interfaz. El dispatcher se ejecuta dentro de ese callback, así que tu handler corre en el **hilo de interfaz**, sin `runOnUiThread` ni una clase `Runnable` adicional. El plugin integrado `notifications.requestPermission` se apoya en el mismo comportamiento cuando llama a `requestPermissions`.

Por eso un handler **bloquea la interfaz** mientras se ejecuta. Hazlo breve y no hagas trabajo de red ni de disco en él. Esta lectura de la documentación de Android no se ha confirmado en un dispositivo en este proyecto; si fuera incorrecta, `Toast.makeText` probablemente lanzaría una excepción y verías `NATIVE_ERROR`.

## 8. Actívalo en koru.json

Guarda el archivo como `plugins/toast/plugin.json` en tu proyecto y añade dos campos a tu `koru.json` existente (es un fragmento, no un archivo completo):

```json
{
  "plugins": ["device", "storage", "app", "toast"],
  "pluginManifests": ["plugins/toast/plugin.json"]
}
```

El ID de `plugins` debe coincidir con el `id` de un manifiesto; si no, el build se detiene con `plugin "toast" has no manifest`. `koru plugin add` y `koru plugin remove` solo manejan plugins integrados, así que edita estos campos a mano. `koru plugin list` muestra `toast` como `on` una vez configurado. Los permisos del manifiesto se añaden al APK automáticamente.

## 9. Llámalo desde JavaScript

Los métodos propios no tienen envoltorio, así que usa `invoke()`. Define tu propia función tipada:

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

Úsala como cualquier otra llamada y gestiona los fallos con los códigos de error:

```js
try {
  await invoke("toast.show", { text: "Hola", long: true });
} catch (error) {
  console.warn(error.code, error.message); // INVALID_REQUEST, NATIVE_ERROR...
}
```

La página completa es [`examples/toast/www/index.html`]({{base}}original/examples/toast/www/index.html ":ignore").

## 10. Pruébalo

```sh
koru build --output app.apk
koru verify app.apk
unzip -p app.apk classes.dex > classes.dex
koru dex inspect classes.dex
koru run
```

`koru build` carga y valida cada manifiesto, así que los errores de estructura fallan aquí, con el mensaje del ensamblador DEX. `koru verify` comprueba la firma y que `classes.dex` se pueda leer. `koru dex inspect` imprime en JSON los tipos, métodos y strings decodificados: busca `Lcom/example/toast/ToastPlugin;`, su método `show` y la cadena de ruta `toast.show`. `koru run` instala y abre la app en un dispositivo conectado.

La prueba del repositorio `internal/app/toast_example_test.go` automatiza los tres primeros pasos para el ejemplo.

## Límites que debes conocer

- **El plugin `toast` no se ha ejecutado en Android.** Este entorno no tenía dispositivo. Las pruebas demuestran que el manifiesto es válido, que el APK se construye y se verifica, y que la clase y el método están en `classes.dex`. No demuestran que ART acepte el bytecode, que el toast aparezca ni cuál es el hilo usado. Ejecuta `koru run` en un teléfono o emulador antes de confiar en él. dexdump y el parser de koru no sustituyen esa prueba.
- **No hay versión web.** Con `koru dev` y en un navegador, `invoke("toast.show")` se rechaza con `METHOD_NOT_FOUND`, porque solo los plugins integrados tienen implementación web. Protege la llamada:

  ```js
  import { invoke, isNative } from "/__koru/core.mjs";

  async function notify(text) {
    if (isNative()) await invoke("toast.show", { text });
    else console.info(text); // o muestra tu propio mensaje en la página
  }
  ```

- **Los errores son genéricos.** `INVALID_REQUEST` y `NATIVE_ERROR` llevan mensajes fijos (sección 6).
- **El IR escrito a mano es fácil de equivocar.** El ensamblador comprueba estructura, etiquetas, referencias y destinos de salto, pero no todos los tipos de registro. Un error que pase el build aún puede fallar cuando ART verifique la clase.
- **Solo código de confianza.** Un plugin comparte los permisos de la app. Cualquier página que cargues puede llamar a cualquier handler activo, así que valida todos los parámetros. Para código nativo `.so`, consulta el [adaptador JNI opcional]({{base}}docs/es/native/).

[API JavaScript]({{base}}docs/es/plugins/) · [API DEX]({{base}}docs/es/dex-api/) · [Configuración]({{base}}docs/es/configuration/)
