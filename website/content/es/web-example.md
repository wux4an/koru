# Ejecuta el ejemplo web

`examples/web` es una app pequeña sin dependencias, usada en las pruebas Android. Tiene un contador, una segunda página local, CSS y un icono SVG. Es sencilla para que puedas comprobar cada comportamiento del WebView. No necesitas un bundler.

## Genera el APK

Desde la raíz del repositorio, después de [compilar koru]({{base}}es/getting-started/):

```sh
./bin/koru init ./my-app
cp -a examples/web/. my-app/www/
./bin/koru doctor ./my-app
./bin/koru build ./my-app --release --output example.apk
./bin/koru verify example.apk
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

Para la prueba registrada de service worker, copia `tests/fixtures/sw.js` como `my-app/www/sw.js` y regístralo desde tu página. La [evidencia Android]({{base}}es/android-results/) también cubre renderizado con radios de red apagadas.

## Explora los archivos

- [index.html]({{base}}original/examples/web/index.html ":ignore") — estructura y referencias a assets.
- [app.js]({{base}}original/examples/web/app.js ":ignore") — contador y estado.
- [style.css]({{base}}original/examples/web/style.css ":ignore") — estilos.
- [icon.svg]({{base}}original/examples/web/icon.svg ":ignore") — imagen local.
- [second.html]({{base}}original/examples/web/second.html ":ignore") — destino de navegación.

El archivo de fuentes sirve estos documentos como texto; abrir un enlace HTML en la documentación no ejecuta la app Android. El README original contiene una afirmación histórica anterior al builder APK; los [resultados actuales]({{base}}es/android-results/) actualizan ese estado.
