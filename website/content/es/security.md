# Firmas y seguridad

Tu app incluye una identidad de firma y sus propios archivos web. Protege la clave y configura los permisos que necesita tu aplicación.

## Conserva tu clave de firma

`koru init` crea `.koru/signing.pem` con la clave privada y el certificado. Guarda una copia segura fuera del proyecto y excluye `.koru/` del control de versiones. Mantén la clave fuera de `webDir`.

Android exige el mismo certificado para actualizar una app instalada. Tanto debug como release usan `signingKey` de `koru.json`. Si la clave falta o es inválida, el build falla: restaura la identidad original antes de continuar.

`koru keygen --output nueva-identidad.pem` crea otra identidad, pero no sirve para reemplazar la de una app ya distribuida. Consulta [Distribuye tu app]({{base}}es/distribution/) para preparar actualizaciones.

## Incluye solo archivos de la app

Apunta `webDir` a la salida web que quieras distribuir, con su `index.html`. Evita incluir credenciales, archivos privados o contenido ajeno a la app. Koru rechaza claves privadas detectadas, enlaces simbólicos y rutas inseguras, pero revisa siempre qué estás empaquetando.

Los archivos se cargan desde `https://app.koru.invalid`. Ese origen pertenece al contenido local de la app; no requiere un dominio ni un servidor externo. Los scripts e iframes remotos están bloqueados. Incluye tus scripts en la carpeta web.

## Permite solo la red que necesitas

Tu contenido local funciona sin permiso de Internet. Si consumes una API externa, declara `android.permission.INTERNET`, usa HTTPS y configura CORS para el origen de la app. Encontrarás un ejemplo en [Usa tu proyecto web]({{base}}es/existing-project/).

Los plugins personalizados se ejecutan con los permisos de tu app. Incluye solo plugins y bibliotecas nativas de confianza.

## Comprueba el APK que compartes

```sh
./bin/koru build ./my-app --release --output app-release.apk
./bin/koru verify app-release.apk
```

Release desactiva debug y la inspección del WebView. `verify` comprueba la firma y la estructura del APK. Instálalo y prueba también las funciones de tu app en Android.

Para entender el origen HTTPS, MessagePort y los límites del WebView, consulta [APK y runtime]({{base}}es/architecture-apk/).
