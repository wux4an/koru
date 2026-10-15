# Hitos — estado de la versión funcional

| Hito           | Implementación y pruebas                                                                                                                                           | Resultado / límite                                                                                                               |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| 1 DEX mínimo   | Formato 035, IDs, mapa, LEB, MUTF, hashes, parser, método ejecutable                                                                                               | Pasa pruebas propias/AOSP; clases cargadas y ejecutadas en ART                                                                   |
| 2 DEX avanzado | Referencias, constructores, parámetros, campos, instrucciones, arrays, interfaces, conversiones y excepciones                                                      | Bytecode de Activity, bridge y plugins ejecutado en Android; no compilador Java completo                                         |
| 3 Activity     | Activity/WebView y ciclo de vida generados desde Go                                                                                                                | Arranque confirmado por am start, logcat y DevTools                                                                              |
| 4 APK          | AXML, ARSC/icono, ZIP32 alineado, identidad persistente, firma v2 y verificación propia                                                                            | Instalación y actualización con misma firma; aapt/apksigner externos aceptan la salida                                           |
| 5 WebView      | Origen HTTPS controlado, assets, CSP, SPA e interceptor de service workers                                                                                         | HTML/JS/CSS/SVG sin Wi‑Fi ni datos; navegación, Atrás y teclado comprobados                                                      |
| 6 Runtime      | Puente MessagePort, validación/promesas/errores; seis plugins y ABI DEX; JNI opcional                                                                              | Dispositivo/app/storage/vibración/notificaciones/chooser probados; JNI info ejecutado en ART                                     |
| 7 Toolchain    | CLI con ayuda en inglés (`init`, `build --release`, `dex`, `verify`, `doctor`, `version` y `keygen`), caché, ejemplos, documentación bilingüe en Docsify y binario | El build APK sigue siendo offline y está implementado en Go; Linux está validado y los demás hosts tienen pruebas de compilación |

La dependencia Cobra solo se usa en `cmd/koru`. Koru implementa DEX, AXML, ARSC,
el empaquetado APK y las firmas con Go y su biblioteca estándar. La ayuda de la CLI
se genera a partir de los comandos e incluye ejemplos. En terminales interactivos
usa la paleta del sitio; al redirigir la salida, conserva el texto sin códigos de
color. La documentación bilingüe se publica con Docsify y se compila por separado
con Bun; Bun no forma parte del build de APK.

La matriz de ejecución comprobada es Android 11 (API 30), x86_64, en un emulador
por software sin KVM. La prueba cubre el APK completo, no solo que el ZIP y el DEX
tengan una estructura válida. No demuestra compatibilidad con todos los dispositivos
ni permite evaluar el rendimiento en hardware real.

Se conservan logs de fallos iniciales y su diagnóstico. En particular se corrigió
la recepción de puertos nativos y la prueba de regreso de navegación: WebView puede
reemplazar su target DevTools al restaurar una página. La prueba verifica URL y
Activity y se reconecta al target restaurado.

Fuera del alcance inicial quedan las firmas APK v3/v4, AAB, las alarmas
programadas, las pruebas en Android 26–36 y ARM, los benchmarks comparativos,
las anotaciones generales, multidex y la verificación semántica completa de
registros. El parser rechaza estructuras ajenas al subconjunto implementado.
La biblioteca JNI precompilada de ejemplo es para x86_64; las apps que usan solo
los plugins DEX estándar no dependen de esa arquitectura.

## Documentación de uso — rework de octubre de 2026

El README y el sitio bilingüe priorizan instalación, primera app, integración de
un frontend existente, API JavaScript y distribución con actualizaciones firmadas.
La navegación sitúa arquitectura, DEX y JNI en «Por dentro de koru»; los registros
históricos mantienen sus rutas y el archivo conserva los documentos originales.

El diseño usa superficies suaves, tipografía local y una paleta lavanda con modos
claro y oscuro. La investigación y las decisiones están en `website/DESIGN.md`;
las comprobaciones reproducibles del sitio están en `website/VALIDATION.md`.
Este hito no amplía la matriz de compatibilidad Android ni modifica el compilador.
