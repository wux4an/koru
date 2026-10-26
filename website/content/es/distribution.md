# Distribuye tu app

Genera un APK release para compartirlo e instalarlo directamente en Android. Koru todavía no genera AAB para publicar en Google Play.

## Prepara la versión

Antes de compartir el primer APK, revisa `appId`, `name`, `version` e `icon` en `koru.json`. Prueba las pantallas, la navegación y los plugins en tus dispositivos objetivo.

Conserva una copia segura de `.koru/signing.pem` fuera del proyecto. Es la identidad que Android usa para reconocer las actualizaciones de tu app. No la incluyas en Git, en la carpeta web ni junto al APK que compartes.

## Genera un release

Desde el directorio que contiene tu proyecto `my-app`:

```sh
koru doctor ./my-app
koru build ./my-app --release --output app-release.apk
koru verify app-release.apk
```

`--release` desactiva el modo debug y la inspección del WebView. Usa la clave configurada en `signingKey`; no crea otra identidad. El archivo resultante es `app-release.apk` en el directorio actual.

Instala y abre también este APK: `verify` comprueba la firma y la estructura, pero no ejecuta tu app. `koru run --release` instala una build release en un dispositivo conectado en un solo paso.

## Comparte el archivo

Puedes enviar el APK o alojarlo como descarga. La persona que lo instala deberá autorizar a su navegador o gestor de archivos para instalar apps de esa fuente. El texto y la ubicación de ese permiso dependen de Android y del dispositivo.

Para probarlo con un dispositivo conectado:

```sh
adb install -r app-release.apk
```

Abre la app desde el lanzador y comprueba su comportamiento antes de compartirla.

## Publica una actualización

Conserva el mismo `appId` y la misma clave. Aumenta `versionCode` y cambia `version`. Deja que koru lo haga:

```sh
koru bump minor                     # version 1.0.0 -> 1.1.0, versionCode 1 -> 2
koru build --release --bump --output app-release.apk   # bump de patch y luego build
```

`bump` acepta `patch` (por defecto), `minor`, `major` o `code` (solo `versionCode`). Usa `--bump` en `build` para una versión de parche, o ejecuta antes `bump` con otro nivel, pero no ambos: subirías la versión dos veces. También puedes editar los campos de `koru.json` a mano, por ejemplo:

```json
"version": "1.1.0",
"versionCode": 2
```

Vuelve a compilar tu frontend, genera el APK release e instálalo sobre la versión anterior. Comprueba que se mantienen los datos guardados.

Si Android rechaza la actualización por una firma distinta, recupera la clave original. Crear otra clave con `keygen` no soluciona la actualización. Desinstalar la app anterior borra sus datos; evita usarlo como solución para tus usuarios.

## Deep links

Si tu app debe abrirse desde enlaces web, añade `deepLinks` a `koru.json` y publica en tu dominio el `assetlinks.json` que imprime `koru deeplinks`, para que Android pueda verificar los enlaces `https`. Los detalles y límites, incluido que el enlace todavía no llega a tu JavaScript, están en [Configura tu proyecto]({{base}}es/configuration/). Nada de los deep links se ha ejecutado en Android; pruébalos en un dispositivo antes de publicar.

Consulta [Firmas y seguridad]({{base}}es/security/) para gestionar la identidad y [Solución de problemas]({{base}}es/troubleshooting/) si la instalación falla.
