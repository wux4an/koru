# Resultados reales — 2026-10-13

Todas las pruebas registradas se ejecutaron sobre código real en este workspace.
No hay pruebas de instalación/ejecución Android ni de un APK: esos resultados no
existen todavía. `summary.json` recoge casos, cobertura y hashes de artefactos.

## Validación

| Comprobación | Resultado y evidencia |
| --- | --- |
| Suite Go | Pasa; `unit-tests.txt`, `tests.jsonl` |
| Cobertura conjunta de paquetes | 84.0%; `coverage.out`, `coverage-summary.txt` |
| `go vet ./...` | Exit 0, sin hallazgos; `vet.txt` vacío corresponde a éxito |
| `go test -race ./...` | Pasa; `race.txt` |
| Fuzz ULEB128/SLEB128, 10 s, dos workers | 552944 ejecuciones, sin fallo; `fuzz-leb.txt` |
| Fuzz MUTF-8, 10 s, dos workers | 271686 ejecuciones, sin fallo; `fuzz-mutf8.txt` |
| Fuzz parser, 15 s, dos workers, semillas Answer y Activity | 256070 ejecuciones, sin fallo; `fuzz-parser.txt` |
| Mutación deliberada de la validación Adler-32 | La prueba detecta la mutación y falla; se restauró el código antes de la suite final; `mutation-check.txt` |
| `dexdump` oficial AOSP | Checksum, apertura y desensamblado de Answer/Activity pasan; `external-tests.txt`, `*-dexdump.txt` |
| Build sin proxy de módulos ni CGO | Pasa con `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0` |
| CLI sin herramientas en PATH | Generación/inspección pasan con `PATH=/nonexistent`; `offline-cli.txt` |
| Entrega ZIP | CRC correcto; reconstrucción desde extracción limpia sin proxy de módulos; binario y Answer DEX idénticos byte a byte (`cmp`, exit 0) |
| Dependencias | Cobra solo en `cmd/koru`; DEX, AXML, ARSC, APK y firmas siguen usando Go y la biblioteca estándar |
| Windows amd64/macOS arm64 | Compilación cruzada solamente; `cross-builds.txt`; no se ejecutaron en esos sistemas |

La cobertura conjunta usa `go test -coverpkg=./... -coverprofile=... ./...`:
instrumenta también llamadas entre paquetes. Es evidencia de alcance de las
pruebas, no una garantía de ausencia de errores.

Los registros de comandos y hashes de este informe corresponden al estado previo
a la incorporación de Cobra. La suite vigente también cubre la ayuda generada,
los comandos en inglés y la validación de argumentos. El CLI usa Cobra; el núcleo
de empaquetado y formatos Android sigue escrito en Go y la biblioteca estándar.

## Medidas DEX

Host Linux amd64, CPU reportada por Go: AMD EPYC 9V74. Los benchmarks tienen
`-benchmem` y usan el IR preparado fuera de la región medida.

| Caso | Tiempo medio | Asignaciones | Bytes asignados/op | Archivo DEX |
| --- | --- | --- | --- | --- |
| Constructor binario de una clase con método estático `answer()I` | 6176 ns/op | 70 | 3926 B/op | 396 bytes en este fixture de benchmark |
| Activity, build público con parser/verificación incluida | 165560 ns/op | 1420 | 98460 B/op | 1848 bytes |

El archivo `artifacts/answer.dex` incluye además un constructor y ocupa 476 bytes;
no es el mismo fixture que el benchmark mínimo. El binario Linux se compila con
`-trimpath`, sin stripping y sin CGO; su tamaño y hash exactos están en
`summary.json`. Los bytes asignados no son una medida de memoria pico.

No se midieron compilación APK, instalación, tiempo de arranque, rendimiento del
puente ni comparativas con Capacitor. No se afirma superioridad de rendimiento.

## Reproducir

Los comandos están en el README principal. El compilador Go de la sesión está en
`/workspace/tooling/go/bin`, y la herramienta oficial de contraste en
`/workspace/tooling/android-validation/dexdump`. La herramienta de contraste no
se redistribuye con koru ni es dependencia del build. Procedencia y hashes figuran
en `tooling.json`; fuentes oficiales consultadas en `../specs/sources.json`.

Para repetir el build/CLI sin disponibilidad de procesos auxiliares:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 go build -trimpath -o bin/koru ./cmd/koru
PATH=/nonexistent ./bin/koru dex build examples/dex/answer.json --output answer.dex
PATH=/nonexistent ./bin/koru dex build --activity-experiment --output activity.dex
PATH=/nonexistent ./bin/koru dex inspect activity.dex
```

El binario no importa `os/exec` ni clientes de red. `os/exec` solo aparece en la
prueba externa optativa. `dexdump` verifica estructura/checksum y muestra código;
no ejecuta los métodos en ART ni verifica su renderizado en Android.
