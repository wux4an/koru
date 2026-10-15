# Configura tu proyecto

Configura el nombre, el icono, los archivos web y los permisos de tu app en `koru.json`. Lo encontrarás en la carpeta que creaste con `koru init`. Las rutas de este archivo parten de esa carpeta.

## Un punto de partida completo

```json
{
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

`init` usa `www` en `webDir`, `com.example.koru` en `appId`, `Koru` en `name` y deja SPA desactivado. El ejemplo anterior selecciona la salida de un framework y activa las rutas SPA.

Después de editar ejecuta `./bin/koru doctor ./my-app`. El resultado correcto muestra tu `appId`; el build también inspecciona todos los assets web. `koru.json` va en la raíz del proyecto, junto a la carpeta web. Sus rutas son relativas a esa raíz; el destino CLI `--output` es relativo al directorio actual de la terminal.

## Campos

| Campo             | Significado y validación                                                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------- |
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

`init` escribe los valores iniciales; edita el archivo generado. Un plugin personalizado debe estar en `plugins` y tener su manifest en `pluginManifests`. Los seis IDs integrados y sus métodos están en [JavaScript y plugins]({{base}}es/plugins/).

## Usa la salida de tu frontend

Compila el frontend antes de koru. Si tu proyecto Parcel o Vite está en `my-app/` y genera `my-app/dist/index.html`, pon `"webDir": "dist"` y ejecuta:

```sh
npm --prefix ./my-app run build
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output app.apk
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

[Firmas y seguridad]({{base}}es/security/) · [Plugins personalizados]({{base}}es/plugins/)
