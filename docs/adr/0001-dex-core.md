# ADR 0001: núcleo DEX independiente y verificable

Fecha: 2026-10-13. Estado: adoptado para el incremento inicial.

El objetivo exige emitir DEX real sin herramientas Java ni empaquetadores Android.
Se implementa en Go con biblioteca estándar exclusivamente. La salida es DEX 035:
no necesitamos invoke-custom, default methods ni versiones container. El runtime
experimental se orienta a Android API 26 o posterior; todavía no hay configuración
APK que aplique ese mínimo.

Se separan modelo simbólico, codificadores, ensamblador, constructor, parser y
runtime. La API pública está en `pkg/dex` para que un consumidor externo pueda
usar el generador sin importar paquetes `internal`. Los alias públicos evitan dos
representaciones incompatibles del mismo IR. La CLI admite JSON del mismo modelo
para permitir construir sin compilar un programa Go de usuario.

La compilación tiene dos pasadas lógicas: recolección y orden de símbolos, seguida
de resolución de instrucciones y emisión con fixups de offsets. El escritor es
append-only con fixups de campos reservados; no cambia la longitud de secciones
ya posicionadas. Registros de parámetros simbólicos permiten añadir locales sin
reescribir manualmente los índices físicos de argumentos.

El salto goto usa siempre la forma /32 para evitar un algoritmo de relajación de
saltos en este primer incremento. Las ramas condicionales conservan los límites
oficiales de 16 bits y fallan explícitamente si no caben. No hay trampolines
silenciosos. `const-string` selecciona jumbo cuando lo exige el índice.

Se comparte cada type_list exacta, se ordenan clases por dependencias y se ordenan
los miembros por identificadores. La misma entrada produce los mismos bytes;
reordenar declaraciones equivalentes tampoco cambia el archivo. No se incorporan
fechas, paths de host ni nombres de archivos de código inventados.

El parser no importa el escritor, builder ni ensamblador. Enumera físicamente las
secciones a partir de map_list y comprueba que las referencias apunten a comienzos
de ítems reconocidos. Comparte solo las primitivas pequeñas y las reglas de nombres
con el emisor: esas primitivas tienen pruebas golden independientes. Rechaza
secciones fuera del subconjunto soportado. Este enfoque hace explícito qué puede
afirmar la herramienta; no sustituye la resolución ni el análisis de tipos de ART.

La Activity se emite únicamente mediante `--activity-experiment`. Usa assets de
archivo para una prueba inicial admitida por la especificación del proyecto; no
se presenta como solución de producción. No hay AndroidX implícito ni puente
addJavascriptInterface. El siguiente runtime requiere una clase WebViewClient
generada con interceptación HTTPS local y una política de origen/navegación.

No se añaden aún C, JNI, plugins ni módulos vacíos de firmas o APK para simular
progreso. El bloqueo es la aceptación real en Android y las capacidades que todavía
no están implementadas, no una dependencia externa disimulada.

Fuentes: [formato DEX](https://source.android.com/docs/core/runtime/dex-format),
[bytecode](https://source.android.com/docs/core/runtime/dalvik-bytecode),
[formatos de instrucciones](https://source.android.com/docs/core/runtime/instruction-formats),
[Activity](https://developer.android.com/reference/android/app/Activity),
[WebView](https://developer.android.com/reference/android/webkit/WebView),
[WebSettings](https://developer.android.com/reference/android/webkit/WebSettings#setAllowFileAccess(boolean)).
La firma APK se diseñará a partir de las [especificaciones v2](https://source.android.com/docs/security/features/apksigning/v2)
y [v3](https://source.android.com/docs/security/features/apksigning/v3) ya consultadas.
