# Ejecuta el ejemplo web

Prueba una app pequeña de HTML, CSS y JavaScript antes de conectar tu proyecto. El ejemplo incluye un contador, una imagen SVG y una segunda página. No necesitas instalar dependencias web.

## Genera el APK

Desde la raíz del repositorio, después de [compilar koru]({{base}}es/getting-started/):

```sh
koru init ./my-app
cp -a examples/web/. my-app/www/
koru doctor ./my-app
koru build ./my-app --release --output example.apk
koru verify example.apk
```

`cp -a examples/web/. my-app/www/` reemplaza la página inicial con todos los archivos del ejemplo. El punto final importa: copia el contenido, sin crear `www/web/`. Si tu sistema no tiene `cp` de GNU, copia cada archivo a `my-app/www/`.

La salida del build se parece a esta; tamaño y certificado varían:

```text
Runtime cache hit: false
Wrote example.apk (… bytes), APK v2 RSA/SHA-256, certificate …
```

Antes de distribuir, cambia `appId` y `name` en `my-app/koru.json`. Conserva `.koru/signing.pem` para las actualizaciones.

## Compruébalo en Android

Instala y arranca con el ID inicial de `init`:

```sh
adb install -r example.apk
adb shell am start -W -n com.example.koru/com.koru.runtime.MainActivity
```

La primera página debe mostrar el SVG, el título «koru», una línea de texto, `Contador: 0` y `JavaScript ejecutado`. Toca dos veces: el botón pasa a `Contador: 2`. Abre `Probar navegación`: verás `Segunda página local`. El botón Atrás de Android debe volver a la primera página. Así se comprueban HTML, CSS, SVG, JavaScript y navegación local en WebView.

`app.js` no llama plugins nativos. Para probar el [puente JavaScript]({{base}}es/plugins/), añade un módulo que importe `/__koru/core.mjs`. La vista previa en navegador solo prueba la parte web; necesitas Android para verificar el puente.

## Explora los archivos

- [index.html]({{base}}original/examples/web/index.html ":ignore") — estructura y referencias a assets.
- [app.js]({{base}}original/examples/web/app.js ":ignore") — contador y estado.
- [style.css]({{base}}original/examples/web/style.css ":ignore") — estilos.
- [icon.svg]({{base}}original/examples/web/icon.svg ":ignore") — imagen local.
- [second.html]({{base}}original/examples/web/second.html ":ignore") — destino de navegación.

Estos enlaces muestran el código fuente. Para ver el ejemplo funcionando, instala el APK en Android.
