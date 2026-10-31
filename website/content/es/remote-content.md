# Contenido remoto y apps híbridas

Por defecto una app koru muestra solo los archivos de `webDir`. El bloque opcional `server` de `koru.json` le permite cargar una página de tu propio sitio, abrir algunos sitios dentro de la app y enviar los demás enlaces al sistema. Es la **fase 1** del contenido remoto y **todavía no está validada en un dispositivo Android**; consulta [Estado](#estado).

## Configúralo

```json
"server": {
  "url": "https://example.com/app",
  "allowNavigation": ["example.com", "*.example.com"],
  "bridge": ["https://example.com"]
}
```

Es un fragmento, no un `koru.json` completo. Los tres campos son opcionales.

| Campo             | Significado y validación                                                                                                                                                                |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `url`             | Primera página que se abre en lugar de `webDir/index.html`. Debe ser `https`. No puede ser `app.koru.invalid`.                                                                          |
| `allowNavigation` | Hosts que se abren **dentro** de la app, máximo 32. `*.example.com` coincide con subdominios como `app.example.com`, no con `example.com`. El host de `url` se incluye automáticamente. |
| `bridge`          | Orígenes `https` exactos, máximo 32, sin comodines y sin ruta, cuyas páginas reciben la API de koru. Cada uno debe cargarse en la app: estar en `url` o en `allowNavigation`.           |

`webDir` sigue siendo obligatorio. Sigue sirviendo tus páginas locales en `https://app.koru.invalid/`. Usar `server` añade `android.permission.INTERNET` automáticamente. Sin `server`, el bytecode del runtime es idéntico byte a byte al de antes; una prueba con hashes fijados lo comprueba.

## Tres formas de usarlo

### Una app que es una web

Apunta `url` a tu sitio. Añade `bridge` solo si sus páginas deben llamar a la API nativa.

```json
"server": {
  "url": "https://example.com/app",
  "bridge": ["https://example.com"]
}
```

### Una app local con tu sitio o tu login dentro

Mantén tus páginas locales como primera pantalla y deja que algunos hosts se abran en la app en lugar de salir de ella.

```json
"server": {
  "allowNavigation": ["login.example.com", "example.com"]
}
```

### Enlaces externos que por fin abren el navegador

Sin `server`, un enlace a otro sitio no hace nada. Con cualquier campo de `server`, incluso un bloque vacío, se abre con la app del sistema. Esta es toda la configuración:

```json
"server": {}
```

## Cómo funciona la navegación

Con `server` presente, la navegación principal se decide por su URL:

- Una URL `https` cuyo host está en `url` o en `allowNavigation`, o el origen local de la app, se carga dentro de la app. Solo `https` navega dentro de la app.
- Cualquier otro enlace se abre fuera de la app, con la app del sistema, para los esquemas `https`, `http`, `mailto`, `tel`, `sms`, `geo` y `market`.
- Los demás esquemas, como `intent:` o `javascript:`, se ignoran.
- Los iframes cargan donde están; esta regla es solo para la navegación principal.

Sin `server`, el comportamiento no cambia: los enlaces externos no hacen nada.

## Da la API de koru a las páginas remotas

`bridge` lista los orígenes cuyas páginas reciben los mismos plugins que la app local, incluido el `storage` compartido. El servidor remoto es tuyo, así que esas páginas deben cargar `core.js` ellas mismas; cópialo a tu sitio con `koru js`.

Con `bridge`:

- La app añade ` KoruApp/1` al user agent del WebView.
- Al terminar de cargar cada página, el origen de la app y los orígenes de `bridge` reciben el canal nativo. El MessagePort se entrega solo a ese origen exacto.
- El resto de páginas recibe `koru-none` y `core.js` usa las implementaciones web.
- En una página remota con el marcador, `Koru.platform` es `"android"` desde el principio y las llamadas esperan a que la página termine de cargar.

## Seguridad

`bridge` da a ese sitio el **mismo acceso nativo que tu app**. Cualquier XSS o compromiso de ese sitio puede usar la API. Lista solo orígenes que controles, sírvelos por HTTPS y define una Content-Security-Policy. Las páginas remotas no reciben la CSP que koru pone a los archivos locales; esa cabecera la envías tú. Consulta [Firmas y seguridad]({{base}}docs/es/security/#origenes-remotos-y-el-bridge).

## Estado

No validado en un dispositivo. Hay dos comprobaciones sin uno: el runtime construido sin `server` es idéntico al de antes, y `koru build` ahora pasa todo el bytecode (el runtime y tus plugins propios) por un verificador de tipos de registros parecido al de ART (`internal/dex/verify`). Se niega a construir un APK que ART rechazaría al cargar la clase. Eso no es ejecución en Android.

La app de pruebas en `examples/selftest` usa `allowNavigation: koru.js.org` y `bridge: https://koru.js.org`, con enlaces para probarlo: una página remota en `https://koru.js.org/selftest/remote/` que comprueba el bridge y escribe `KORU_SELFTEST_REMOTE` en logcat, un enlace externo y un enlace `mailto:`. Su README los explica.
