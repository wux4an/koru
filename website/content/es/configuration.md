# Configura tu proyecto

Configura el nombre, el icono, los archivos web y los permisos de tu app en `koru.json`, en la raíz del proyecto que crearon `koru create` o `koru init`. Las rutas de este archivo parten de esa carpeta.

## Un punto de partida completo

```json
{
  "$schema": "https://koru.js.org/schema/koru.schema.json",
  "appId": "com.example.demo",
  "name": "Demo",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "./dist",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app"],
  "signingKey": ".koru/signing.pem",
  "spa": true
}
```

`create` completa `name` y `appId` con tus respuestas y define `webDir` según la plantilla (`www` o `dist`). `init` usa `www`, `com.example.koru` y `Koru`, y deja SPA desactivado. El ejemplo anterior selecciona la salida de un framework y activa las rutas SPA.

Después de editar ejecuta `koru doctor`. El resultado correcto muestra tu `appId`; el build también inspecciona todos los assets web. `koru.json` va en la raíz del proyecto, junto a la carpeta web. Sus rutas son relativas a esa raíz; el destino CLI `--output` es relativo al directorio actual de la terminal.

## Autocompletado en el editor

La línea `$schema` indica a editores como VS Code el [JSON Schema]({{base}}schema/koru.schema.json ":ignore") de koru. Obtienes autocompletado, descripciones al pasar el ratón y errores en valores no válidos mientras escribes. koru ignora el campo al compilar; los proyectos de `create` e `init` lo incluyen.

Cuando un valor no es válido, koru nombra el campo y cómo corregirlo, por ejemplo:

```text
koru: koru.json: invalid appId "myapp": use reverse-domain form such as com.example.myapp (letters, digits and _; at least two segments)
```

## Campos

| Campo             | Significado y validación                                                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `$schema`         | Opcional; ayuda al editor y koru lo ignora.                                                                                 |
| `appId`           | Identificador con puntos, máximo 127 bytes. Cada segmento empieza con una letra y contiene letras, números o guiones bajos. |
| `name`            | Nombre visible no vacío, máximo 256 bytes.                                                                                  |
| `version`         | Etiqueta de versión no vacía, máximo 128 bytes.                                                                             |
| `versionCode`     | Entero entre 1 y 2.100.000.000. Increméntalo para actualizar.                                                               |
| `webDir`          | Carpeta obligatoria de assets estáticos con `index.html`.                                                                   |
| `minSdk`          | API Android mínima, al menos 26.                                                                                            |
| `targetSdk`       | API objetivo; `minSdk ≤ targetSdk ≤ 36`. Es el rango configurable, no la matriz ejecutada.                                  |
| `permissions`     | Nombres únicos como `android.permission.INTERNET`. Usa los identificadores Android completos.                               |
| `plugins`         | IDs únicos activados. Por defecto: `device`, `storage`, `app`. Otros integrados: `vibration`, `share`, `notifications`.     |
| `signingKey`      | PEM existente obligatorio con clave privada y certificado.                                                                  |
| `icon`            | Ruta opcional de icono PNG.                                                                                                 |
| `spa`             | Booleano opcional; rutas sin extensión devuelven `index.html`.                                                              |
| `pluginManifests` | Rutas opcionales a manifests de plugins DEX. Activa sus IDs en `plugins`.                                                   |
| `nativeLibraries` | Mapa opcional de ABI Android a listas de bibliotecas ELF precompiladas.                                                     |
| `orientation`     | Opcional: `portrait`, `landscape` o `sensor`. Omítelo para mantener el valor por defecto de Android.                        |
| `allowBackup`     | Booleano opcional, por defecto `false`. Con `true` la app usa la copia de seguridad automática de Android.                  |
| `keyboard`        | Opcional: `resize` (por defecto; la página se reduce) o `pan` (la página sube) cuando se abre el teclado en pantalla.       |
| `deepLinks`       | Lista opcional de enlaces que abren la app. Máximo 32, sin duplicados. Consulta [Deep links](#deep-links) más abajo.        |

`init` escribe los valores iniciales; edita el archivo generado. Un plugin personalizado debe estar en `plugins` y tener su manifest en `pluginManifests`. Los seis IDs integrados y sus métodos están en [JavaScript y plugins]({{base}}docs/es/plugins/).

## Pantalla, teclado y copia de seguridad

```json
{
  "orientation": "portrait",
  "keyboard": "pan",
  "allowBackup": false
}
```

Es un fragmento para integrar en tu `koru.json`. `orientation` bloquea la pantalla: `portrait`, `landscape` o `sensor` (sigue el sensor del dispositivo). `keyboard` decide qué hace el teclado en pantalla con una página que tiene un campo de texto cerca del borde inferior: `resize` reduce la página por encima del teclado y `pan` sube la página entera. `allowBackup` está desactivado si no lo indicas; actívalo solo si quieres que Android guarde los datos de la app en la cuenta de Google del usuario.

Un valor no válido detiene el build y explica cómo corregirlo:

```text
koru: koru.json: invalid orientation "upside": use "portrait", "landscape" or "sensor", or remove the field to keep Android's default
```

`koru build` los escribe en el manifiesto generado, y los IDs de los atributos se contrastaron con las fuentes de Android. Ninguno se ha ejecutado todavía en un dispositivo: la matriz API 30 registrada no cubre la rotación, el teclado ni la copia de seguridad.

## Deep links

`deepLinks` hace que ciertos enlaces abran tu app. Cada entrada es un `scheme` y, opcionalmente, un `host` y un `pathPrefix`:

```json
"deepLinks": [
  { "scheme": "https", "host": "example.com", "pathPrefix": "/app" },
  { "scheme": "myapp" }
]
```

| Entrada                                                              | Abre la app para                                                      |
| -------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `{ "scheme": "myapp" }`                                              | Cualquier enlace `myapp://…`. Un esquema propio no necesita nada más. |
| `{ "scheme": "https", "host": "example.com" }`                       | Cualquier enlace `https://example.com/…`.                             |
| `{ "scheme": "https", "host": "example.com", "pathPrefix": "/app" }` | Solo rutas que empiecen por `/app`. Requiere `host`.                  |

Reglas que comprueba koru: el esquema usa minúsculas, dígitos, `+`, `.` o `-` y empieza por una letra; se rechazan `http`, `javascript`, `file` y `content`; un enlace `https` necesita un dominio completo en `host` (se admite un comodín `*.` al inicio); `pathPrefix` empieza por `/`, no lleva espacios, `?` ni `#`, y requiere `host`. Los errores nombran la entrada, por ejemplo:

```text
koru: koru.json: forbidden deepLinks scheme "http": plain HTTP is not allowed: use "https"
```

**El enlace abre la app en su página inicial. La URL todavía no llega a tu JavaScript**, así que la app no puede leer qué enlace se pulsó. Hoy los deep links sirven para lanzar la app desde un enlace web, no para llevar a una pantalla concreta.

### Verifica los enlaces `https` con assetlinks.json

En los enlaces `https`, Android abre la app sin mostrar un selector solo cuando comprueba que tu dominio confía en ella. koru marca estos enlaces para verificación (`autoVerify`) e imprime el archivo que debe servir el dominio:

```sh
koru deeplinks ./my-app --json > assetlinks.json
```

```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.example.myapp",
      "sha256_cert_fingerprints": [
        "86:CC:81:42:72:1E:56:ED:50:68:3F:39:CF:48:64:56:27:90:D1:6E:54:32:B1:C8:12:C4:23:4B:5C:70:C2:B1"
      ]
    }
  }
]
```

La huella es el SHA-256 del certificado de tu `signingKey` (esta sale de una clave de ejemplo desechable). Publica el archivo en `https://example.com/.well-known/assetlinks.json` para cada dominio `https` de `deepLinks`, por HTTPS con estado 200, tipo de contenido `application/json` y sin redirecciones. Un comodín `*.example.com` necesita el archivo en cada subdominio que deba abrir la app. Si cambias la clave de firma, la huella cambia y la verificación falla; consulta [Firmas y seguridad]({{base}}docs/es/security/). Sin el archivo, o hasta que Android lo compruebe, un enlace puede mostrar un selector en lugar de abrir la app directamente.

`koru verify app.apk` lista los deep links que hay en el manifiesto. **Nada de esto se ha ejecutado en Android:** koru no ha registrado una instalación, una verificación ni un enlace pulsado. Prueba en un dispositivo antes de depender de ello.

## Usa la salida de tu frontend

Compila el frontend antes de koru. Si tu proyecto Parcel o Vite está en `my-app/` y genera `my-app/dist/index.html`, pon `"webDir": "dist"` y ejecuta:

```sh
npm --prefix ./my-app run build
koru doctor ./my-app
koru build ./my-app --release --output app.apk
```

El comando `npm` supone que existe un script `build`; usa el gestor y el script de tu proyecto. Koru no instala paquetes ni compila JavaScript. Si el HTML generado apunta a `/assets/app.js`, el APK debe incluir `dist/assets/app.js`. Configura la ruta base pública del frontend para el origen del APK, sin un subdirectorio de despliegue. No incluyas `dist` en la URL.

## Red y permisos

El contenido local funciona sin `INTERNET`. Declara ese permiso si necesitas servicios HTTPS externos y configura el servidor para `https://app.koru.invalid`. Scripts e iframes externos están bloqueados; empaqueta los scripts localmente.

```json
"permissions": ["android.permission.INTERNET"]
```

Es un fragmento de campo, no un `koru.json` completo. Haz las solicitudes remotas desde el JavaScript de tu app y configura CORS en el servidor. El origen local HTTPS está reservado; no necesitas registrarlo en DNS.

Activar `vibration` añade `android.permission.VIBRATE`. Activar `notifications` añade `android.permission.POST_NOTIFICATIONS`. La matriz API 30 no cubre el diálogo de permisos introducido en API 33.

## Assets, rutas y caché

Se rechazan symlinks, claves privadas entre assets, entradas inseguras y entradas demasiado grandes. El APK debe quedar fuera de `webDir`. Las rutas desconocidas con extensión devuelven 404 incluso en modo SPA. Los service workers usan la misma allowlist.

El límite es 2.000 assets, 64 MiB por archivo y 256 MiB en total. El builder omite `.git`, `.koru` y `node_modules`. Deja `signingKey` fuera de `webDir`; no dependas de la omisión para proteger una clave mal situada. Para rutas cliente, activa `spa`: rutas desconocidas sin extensión devuelven `index.html`, mientras que `.js`, `.css` o imágenes ausentes siguen devolviendo 404. Para sitios multipágina, deja `spa` apagado e incluye todos los HTML.

La caché DEX considera toolchain, configuración, plugins, rutas y tipos MIME. Cambiar solo bytes web puede reutilizar DEX. La corrupción de integridad o estructura produce un error. `--no-cache` permite desactivarla.

[Firmas y seguridad]({{base}}docs/es/security/) · [Plugins personalizados]({{base}}docs/es/plugins/)
