# Evidencia de Koru 0.2.0-dev

Resultados obtenidos el 2026-10-13. Los archivos son salidas reales; no fixtures que
simulen ejecución Android. Herramientas oficiales solo se usan en QA, fuera del
build normal.

## Compilador y toolchain

- 114 casos/subcasos Go aprobados en 16 paquetes: go-tests.jsonl.
- Cobertura medida: 74.2% global, incluyendo CLI/validadores nuevos: coverage.txt.
- El lector de string pools limita coste decodificado a 64 MiB, contabilizando
  cada referencia y unidades UTF-16; evita amplificación por offsets compartidos.
  La prueba adversarial y el APK byte a byte inalterado están registrados.
- go vet y go test -race: sin errores, vet.txt y race.txt.
- Node: 2 pruebas del protocolo aprobadas, js-tests.txt.
- JNI C: referencias y attach/detach comprobados con tabla controlada, native-unit.txt.
- Fuzz AXML/ARSC/firma: sin fallos en campañas registradas. Se conserva también la
  campaña XML inicial de solo 12 ejecuciones; estas campañas cortas no representan
  exploración exhaustiva. Ver fuzz-*.txt.
- dexdump, aapt y apksigner oficiales aceptan salida propia: official-tests.txt.
- Binario Linux estático CGO=0, trimpath y buildvcs=false para reproducirlo
  desde una entrega sin metadata Git. Build con PATH vacío y APK byte a byte idéntico:
  offline-isolated.txt. El build no descarga herramientas ni lanza subprocesses.
- Compilación cruzada, sin ejecución en otros hosts: host-portability.txt.

## Android real

Emulador oficial, Android 11/API 30, x86_64, imagen default r11. Ejecución software
sin KVM. Se instalaron y arrancaron APKs generados por Go, y se actualizaron con
la misma identidad. El certificado de las muestras es
bfee8015a6fbf4e44d5ac5f01d59470a08e33ea3534b5da1e3cbe82da6f719f5.
La clave privada no se distribuye.

- android-smoke.json: origen HTTPS local, HTML/JS/CSS/SVG, módulo ES, errores
  estructurados, dispositivo/app/storage/vibración/notificación y native.info JNI
  en el hilo principal. Red Wi-Fi y datos desactivados, sin permiso INTERNET.
- android-back.json: navegación a second.html, Atrás a index.html y almacenamiento
  después de reinicio. WebView reemplaza el target DevTools al restaurar historial;
  la prueba se reconecta. Los logs back-diagnosis son fallos anteriores corregidos.
- android-worker.json: service worker registrado y respuesta local offline.
- android-keyboard.json: entrada real KoruKeyboard, IME visible y cierre con Atrás.
  La prueba se reconecta también cuando WebView reemplaza el inspector al redimensionar.
- android-share.json: chooser nativo mostrado y cancelado, sin enviar a destinatarios.
- external-navigation-result.json: navegación a un origen externo bloqueada.
- demo-install/demo-launch: APK release portable instalado y Activity iniciada.
  demo-release-screen.png muestra CSS, SVG, JavaScript y Contador: 1 tras toque real.
- wifi-state/mobile-state: ambos 0 en la prueba final.
- storage-after-restart.json: valor conservado después de force-stop/nuevo arranque.
- update-* registra instalación de versionCode 3 y conservación de datos con la misma firma.

Los logs iniciales contienen fallos de desarrollo corregidos y diálogos de System
UI durante el arranque del emulador. Las capturas first/controlled/all son evidencia
histórica; la captura demo-release-screen corresponde a la entrega final.

## Medidas reproducibles

benchmarks.txt (Linux amd64, AMD EPYC 9V74; resultado puntual):

| Operación | Tiempo | Bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| DEX mínimo | 6.414 µs | 3,926 | 70 |
| Activity experimental mínima | 239.823 µs | 98,662 | 1,422 |
| Build APK con caché | 11.602 ms | 13,266,993 | 6,266 |

El benchmark Activity mide el experimento inicial, no todo el runtime con plugins.
El APK release portable ocupa 22,852 bytes en esta muestra. El APK debug con JNI
x86_64 tiene otro tamaño y no es una medición equivalente. bridgeMillis es una
latencia puntual en emulación software, no una comparación de rendimiento.

## Límites de la evidencia

La matriz ejecutada solo cubre API 30/x86_64. Faltan pruebas en hardware ARM, otros
API y fabricantes, y del diálogo POST_NOTIFICATIONS de API 33+. El APK portable
no incluye bibliotecas nativas y no impone una ABI, pero eso no equivale a haberlo
probado en todas las arquitecturas. No se ha comparado con Capacitor.

No implementados: firmas v3/v4, AAB, alarmas programadas, multidex, anotaciones DEX
generales y análisis completo de tipos por registro. Los seis plugins iniciales
usan framework/DEX; JNI es opcional y su prebuilt incluido solo es x86_64.
