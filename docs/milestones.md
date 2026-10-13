# Hitos — estado de la versión funcional

| Hito | Implementación y pruebas | Resultado / límite |
| --- | --- | --- |
| 1 DEX mínimo | Formato 035, IDs, mapa, LEB, MUTF, hashes, parser, método ejecutable | Pasa pruebas propias/AOSP; clases cargadas y ejecutadas en ART |
| 2 DEX avanzado | Referencias, constructores, parámetros, campos, instrucciones, arrays, interfaces, conversiones y excepciones | Bytecode de Activity, bridge y plugins ejecutado en Android; no compilador Java completo |
| 3 Activity | Activity/WebView y ciclo de vida generados desde Go | Arranque confirmado por am start, logcat y DevTools |
| 4 APK | AXML, ARSC/icono, ZIP32 alineado, identidad persistente, firma v2 y verificación propia | Instalación y actualización con misma firma; aapt/apksigner externos aceptan la salida |
| 5 WebView | Origen HTTPS controlado, assets, CSP, SPA e interceptor de service workers | HTML/JS/CSS/SVG sin Wi‑Fi ni datos; navegación, Atrás y teclado comprobados |
| 6 Runtime | Puente MessagePort, validación/promesas/errores; seis plugins y ABI DEX; JNI opcional | Dispositivo/app/storage/vibración/notificaciones/chooser probados; JNI info ejecutado en ART |
| 7 Toolchain | CLI Cobra con ayuda/salida en inglés para init/build/release/dex/verify/doctor/version/keygen, caché, ejemplos, docs bilingües Docsify y binario | El núcleo del build APK sigue offline y en Go; Linux validado; otros hosts tienen pruebas de compilación |

La dependencia Cobra se limita a `cmd/koru`; DEX, AXML, ARSC, empaquetado APK y
firmas siguen implementados por koru con Go y la biblioteca estándar. La ayuda
se genera desde el árbol de comandos con ejemplos, alineación y la paleta del sitio
en terminales interactivos; los estados mantienen su texto exacto cuando se redirigen.
La web Docsify ofrece contenido en inglés y español y se compila por separado
con Bun; no forma parte del compilador APK.

La matriz real es Android 11, API 30, x86_64, con emulación software sin KVM.
Esto demuestra el APK completo, no solamente un ZIP/DEX bien formado. No demuestra
compatibilidad con todos los dispositivos ni rendimiento en hardware real.

Se conservan logs de fallos iniciales y su diagnóstico. En particular se corrigió
la recepción de puertos nativos y la prueba de regreso de navegación: WebView puede
reemplazar su target DevTools al restaurar una página. La prueba verifica URL y
Activity y se reconecta al target restaurado.

Pendientes fuera del primer alcance: APK v3/v4, AAB, alarmas programadas, matriz
Android 26–36 y dispositivos ARM, benchmarks comparativos, anotaciones generales,
multidex y verificador semántico completo de registros. El parser sigue siendo
estricto para el subconjunto implementado. El ejemplo JNI precompilado es x86_64;
las apps estándar DEX no dependen de esa arquitectura.
