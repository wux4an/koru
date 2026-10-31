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

## Experiencia de desarrollo — octubre de 2026

Spec: `docs/specs/developer-experience.md`. Decisión: ADR 0003.

| Pieza                | Implementación y pruebas                                                                                  | Resultado / límite                                                                 |
| -------------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `koru create`        | Plantillas `basic` y `vite` embebidas, preguntas en TTY, `$schema`, `.gitignore`, tipos y clave           | Ambas generan APKs verificados; la de Vite se validó con `npm install` y `vite build` reales |
| `koru dev`           | Servidor `net/http`, recarga SSE, vigilancia por sondeo, SPA, `--proxy` con `httputil.ReverseProxy`        | Probado en Chromium (modo web, persistencia, recarga) y con el HMR real de Vite vía proxy |
| Modo web de `core.js` | Implementaciones web fuera de `https://app.koru.invalid`; `Koru.platform`                                | Probado en Node y Chromium; no reejecutado en el WebView de API 30                 |
| `koru run`/`devices` | Cliente del protocolo smart-socket de adb, sin procesos externos                                          | Probado contra `internal/adb/adbtest`; **no registrado en hardware ni emulador**   |
| Tipos y schema       | `core.d.ts`, `koru.schema.json` publicado en `/schema/`, mensajes de error con la corrección               | `tsc` 5.6 comprueba ambas plantillas; un test enlaza schema y struct `Config`       |
| `koru js`            | Copia el módulo y sus tipos para bundlers                                                                 | Pruebas Go                                                                         |
| Sitio                | Landing en `/` y `/es/inicio/`; guías de desarrollo, dispositivo, frameworks y comparación en en/es       | `bun test` y 23 comprobaciones de navegador, con axe WCAG 2.2 AA en la landing     |

Este hito no amplía la matriz Android registrada: el APK generado no cambia salvo
el contenido de `core.js` (modo web y `platform`). Hasta reejecutar la matriz de
API 30 con este `core.js`, el camino nativo se considera validado solo por las
pruebas unitarias del protocolo.

`koru bump`, `koru build --bump`, `koru plugin list|add|remove`, `koru config` y los avisos de `koru doctor` (.gitignore, viewport, INTERNET por heurística, icono, targetSdk) se cubren con pruebas Go en `cmd/koru`; son edición de `koru.json` y diagnóstico, sin cambios en el APK generado.

`koru logs`, `koru run --logs` y el QR de `koru dev --qr`: `adb.Stream` (sin límite de 1 MiB), `pidof` y `logcat --pid` más las líneas `chromium`, y la URL de red como QR con `rsc.io/qr` (BSD-3); probados contra `internal/adb/adbtest` y con comprobación de la matriz QR, **sin validar el formato real de logcat ni el escaneo en un dispositivo**.

Plantillas `react`, `vue`, `svelte` y `vanilla-ts` de `koru create`: cada una se validó con `npm install`, `npm run build` (y `tsc --noEmit` en `vanilla-ts`), `koru build` y `koru verify`; no se ejecutaron en Android.

## Manifiesto configurable — octubre de 2026

Campos opcionales de `koru.json` emitidos en el AXML generado por Go: `orientation`
(`android:screenOrientation`), `allowBackup`, `keyboard` (`android:windowSoftInputMode`)
y `deepLinks` (un `<intent-filter>` VIEW/DEFAULT/BROWSABLE por enlace, con `autoVerify`
para https). `koru deeplinks` imprime el `assetlinks.json` con la huella SHA-256 de la
identidad de firma, y `koru verify` lista los `deepLinks` del manifiesto. Los IDs de
atributo se contrastaron con `public-final.xml` de AOSP; hay pruebas golden del AXML,
de configuración inválida y del comando, y un caso `aapt dump xmltree` bajo
`tests/external` (build tag `android_external`) que no se ha ejecutado.

**No se ha ejecutado en Android** y la matriz API 30 de `docs/results/full` no cubre
estos campos: ni la verificación de App Links, ni la rotación, ni el teclado. Los
enlaces abren la app en su página inicial; la URL todavía no llega a JavaScript
(requeriría cambios en el bytecode del runtime, fuera de este hito).

## Ejemplo de plugin propio y guía — octubre de 2026

`examples/toast` es un plugin propio completo (`toast.show`, ABI 1, sin permisos)
escrito solo como IR JSON, con README y una página web de ejemplo. La guía «Write a
plugin» / «Escribe un plugin» del sitio lo explica paso a paso. El handler valida
`text` y `long`; devuelve `null` (`INVALID_REQUEST`) ante parámetros inválidos y
`{}` si todo va bien. `internal/app/toast_example_test.go` carga el manifiesto,
construye y verifica un APK y comprueba la clase y el método en `classes.dex`.

**No se ha ejecutado en Android** (sin dispositivo): ART, el toast visible y el hilo
de ejecución (se asume el hilo de interfaz por `setWebMessageCallback` sin `Handler`)
siguen sin comprobarse. En `koru dev` no hay versión web de los plugins propios.

- Sitio de documentación: `llms.txt`, `es/llms.txt` y `llms-full.txt` generados en el build desde `website/scripts/catalog.ts`, con enlaces absolutos validados y `<link rel="alternate">` en landings y páginas (probado con `bun test` y `test:browser`; sin cambios en el compilador ni en Android).
- Sitio de documentación: SEO de indexación (og:image 1200x630 por idioma, sitemap con `lastmod` desde git y `xhtml:link` hreflang, raíz `/` indexable como x-default, redirecciones legacy sin `noindex`, JSON-LD `@graph`, verificación e IndexNow opcionales) probado con `bun test` y `test:browser`; sin cambios en el compilador ni en Android.

## Releases — v0.0.1 (octubre de 2026)

- `scripts/release.sh` compila builds reproducibles (`CGO_ENABLED=0`, `-trimpath`, versión y commit por `-ldflags`) para linux, darwin y windows en amd64 y arm64, con `.sha256` por archivo y `SHA256SUMS`. Comprobado en local: dos builds del mismo commit dan el mismo binario.
- `.github/workflows/release.yml` (tag `vX.Y.Z`): CI completo, archivos, attestations de procedencia, notas desde `CHANGELOG.md` e instalación de prueba de la release publicada, todo en una sola máquina Linux con compilación cruzada (`scripts/ci.sh`, también para `ci.yml` en `main` y PR).
- Instaladores `scripts/install.sh` (POSIX, verifica SHA-256; probado con `scripts/test-install.sh` y `dash`) e `install.ps1` (sin ejecutar: solo se comprueba su sintaxis con `pwsh`), servidos en `koru.js.org/install.sh` y `/install.ps1`.
- La web muestra la versión de `main.go` y la actualiza en el navegador con la última release de la API de GitHub (caché de 1 h).
- Proceso documentado en `docs/releasing.md`. Sin cambios en el compilador ni en el runtime Android.
- v0.0.2: módulo `github.com/wux4an/koru` (`go install …@latest`), fórmula de Homebrew en `Formula/` y manifiesto de Scoop en `bucket/`, generados en cada release por `scripts/package-managers.sh`. La fórmula se instala con Homebrew en el workflow; el manifiesto de Scoop solo se valida como JSON, no se instala en Windows.
- App de autodiagnóstico `examples/selftest`: 37 comprobaciones automáticas al abrirse (puente, plugins, plugin propio y WebView) más comprobaciones interactivas. Pasa en `koru dev` con Chromium (35 pass, 2 skip esperados); el APK se construye y verifica en CI. Pendiente: ejecutarla en Android.
