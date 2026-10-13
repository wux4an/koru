# Configura tu proyecto

`koru.json` está en la raíz del proyecto. Las rutas se resuelven desde ese directorio salvo que sean absolutas. Se rechazan campos desconocidos, JSON adicional y valores inválidos.

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

## Red y permisos

El contenido local funciona sin `INTERNET`. Declara ese permiso si necesitas servicios HTTPS externos y configura el servidor para `https://app.koru.invalid`. Scripts e iframes externos están bloqueados; empaqueta los scripts localmente.

Activar `vibration` añade `android.permission.VIBRATE`. Activar `notifications` añade `android.permission.POST_NOTIFICATIONS`. La matriz API 30 no cubre el diálogo de permisos introducido en API 33.

## Assets, rutas y caché

Se rechazan symlinks, claves privadas entre assets, entradas inseguras y entradas demasiado grandes. El APK debe quedar fuera de `webDir`. Las rutas desconocidas con extensión devuelven 404 incluso en modo SPA. Los service workers usan la misma allowlist.

La caché DEX considera toolchain, configuración, plugins, rutas y tipos MIME. Cambiar solo bytes web puede reutilizar DEX. La corrupción de integridad o estructura produce un error. `--no-cache` permite desactivarla.

[Firmas y seguridad]({{base}}es/security/) · [Plugins personalizados]({{base}}es/plugins/)
