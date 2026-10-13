# ADR 0002: APK Go-only, origen HTTPS y plugins

Se emiten directamente AXML y una tabla ARSC mínima que contiene el icono mipmap.
El empaquetador ZIP32 propio calcula CRC, comprime con flate de Go, fija tiempos
DOS y alinea recursos almacenados a cuatro bytes. Las bibliotecas nativas opcionales
se almacenan con alineación de 16 KiB. La alineación y los contenidos se cierran
antes de insertar la firma v2; no se modifica el APK firmado.

La firma usa RSA PKCS#1 v1.5/SHA-256 (algoritmo v2 0x0103), digest de bloques de
1 MiB y la estructura de signers especificada por Android. La identidad PEM contiene
la clave y su certificado, se crea explícitamente con init/keygen, tiene permisos
0600 y nunca se regenera durante build. Release reutiliza la identidad; deshabilita
debug y la inspección remota del WebView. El verificador propio contrasta firma,
certificado, digest, ZIP, DEX, manifiesto, recursos e icono.

La aplicación se sirve desde `https://app.koru.invalid`. Un WebViewClient generado
comprueba esquema/autoridad y usa una allowlist exacta de rutas. No concatena paths
recibidos con el filesystem. Solo abre assets del propio APK; errores de lectura
producen un 404. La opción SPA devuelve index.html para rutas sin extensión.

Los callbacks de service workers usan la misma allowlist y Application context.
No retienen Activity. Se deshabilitan accesos file/content; la red se bloquea salvo
que el proyecto declare INTERNET explícitamente. CSP impide scripts e iframes externos;
permite scripts inline/eval necesarios para ciertas salidas web y conexiones HTTPS
cuando la aplicación autoriza red. Los scripts empaquetados son código de confianza.

El puente usa WebMessagePort y transferencia al frame principal con targetOrigin
explícito. No se usa addJavascriptInterface ni hacen falta anotaciones para exportar
métodos Java. El receptor comprueba el origen de su documento, que es el frame
principal y que el evento nativo es trusted y sin Window source. No se interpreta el
campo origin de un evento nativo como origen de una página remitente.

El protocolo v1 lleva id, method, params y result/error. Los plugins habilitados se
compilan en el dispatcher, y solo esas rutas se aceptan. Inputs tienen límites de
tamaño y validación de tipos; errores se convierten en códigos estructurados. El
cliente tiene límite de requests pendientes, timeouts, cancelación antes de dispatch
y cierre de promesas al salir de la página. Una acción Android ya ejecutada no puede
deshacerse mediante cancelación del Promise.

El permiso de notificaciones se resuelve asincrónicamente: se conserva el id y el
puerto exacto de la petición, y se responde al callback de permisos o al regresar
de ajustes. Así una navegación posterior no recibe la respuesta de una página
anterior. API 33+ usa el permiso POST_NOTIFICATIONS; API anteriores comprueban
si las notificaciones están habilitadas y pueden abrir los ajustes del usuario.

La ABI DEX de plugins es v1: método public static que recibe Activity/JSONObject
y devuelve JSONObject. Los manifiestos de plugins no pueden reemplazar clases
Android, Java ni del runtime. Plugins son código de confianza del desarrollador y
comparten UID/permisos de la aplicación; no son un sandbox para código nativo hostil.

La biblioteca JNI es opcional. Los seis plugins básicos usan DEX/framework y no
requieren C. El ejemplo freestanding x86_64 no tiene dependencias libc/JVM y se ha
cargado en ART; muestra referencias globales, limpieza de referencias locales,
excepciones y scopes de attach/detach. Las operaciones UI del puente corren en main.

La caché DEX depende del binario koru, configuración, plugins y nombres/MIME de
assets. Modificar solo bytes web conserva el runtime. La caché tiene SHA-256 y
verificación estructural; corrupción produce error explícito, no fallback. Se
puede desactivar con --no-cache.
