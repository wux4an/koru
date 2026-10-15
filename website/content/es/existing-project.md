# Usa tu proyecto web

Puedes llevar a Android una web hecha con HTML, CSS y JavaScript o la salida estática de tu framework. Necesitas una carpeta con `index.html` y los archivos que usa la página.

Si aún no tienes el ejecutable, sigue los [primeros pasos]({{base}}es/getting-started/). Los comandos de esta guía se ejecutan desde la raíz del repositorio de koru.

## Prepara tus archivos

Compila tu frontend con el comando habitual de tu proyecto. Por ejemplo, si tiene un script `build`:

```sh
npm --prefix /ruta/a/tu-web run build
```

Localiza la carpeta resultante, normalmente `dist` o `build`. Koru empaqueta esa salida; la compilación del frontend se hace antes. Una app que necesita un servidor para renderizar sus páginas debe ofrecer una exportación estática o consumir su servidor como una API HTTPS.

## Crea el proyecto Android

Usa una carpeta nueva para la configuración de koru:

```sh
./bin/koru init ./my-app
```

Copia el contenido de la salida de tu frontend a `my-app/www/`. `index.html` debe quedar directamente dentro de `www`, junto con sus scripts, estilos e imágenes.

Si prefieres evitar la copia, cambia `webDir` en `my-app/koru.json` para apuntar a la salida de tu frontend. Las rutas relativas parten del directorio que contiene `koru.json`. Por ejemplo, `../tu-web/dist` apunta a una carpeta hermana llamada `tu-web`.

## Configura la navegación

Si tu app usa rutas del lado del cliente, añade este campo a `koru.json`:

```json
"spa": true
```

Así, rutas como `/ajustes` abren `index.html` y tu router decide qué pantalla mostrar. Los archivos ausentes con extensión, como `/assets/app.js`, siguen devolviendo 404.

Para una web con varios HTML, deja `spa` desactivado y enlaza los archivos incluidos. Configura el frontend para servirse desde la raíz: `/assets/app.js` debe corresponder a `www/assets/app.js`, sin prefijos de dominio o de despliegue.

## Conecta tu API

Los archivos empaquetados se sirven desde `https://app.koru.invalid`. Para llamar a servicios externos, añade el permiso a la lista `permissions`:

```json
"permissions": ["android.permission.INTERNET"]
```

Usa una URL HTTPS y permite el origen `https://app.koru.invalid` en el CORS de tu servidor. Este origen es local a la app: no necesitas comprar un dominio ni configurar DNS. Empaqueta las dependencias JavaScript; los scripts e iframes remotos están bloqueados.

## Genera y prueba el APK

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
```

Instala el APK en Android y comprueba la pantalla inicial, las rutas, el botón Atrás y las solicitudes de red. Si cambias el frontend, vuelve a compilarlo y a generar el APK.

Continúa con [funciones de Android desde JavaScript]({{base}}es/plugins/) o [distribuye tu app]({{base}}es/distribution/).
