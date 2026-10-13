# koru — instrucciones de ingeniería

El proyecto se llama koru. Implementar por hitos comprobables y mantener
`docs/milestones.md` y los resultados de pruebas coherentes con el código.
Las decisiones técnicas ordinarias y specs ya están autorizadas por el usuario;
no pedir aprobación rutinaria entre hitos.

## Restricciones

El build normal solo puede usar código Go propio y biblioteca estándar para DEX,
AXML, ARSC, empaquetado APK y firmas. No añadir dependencias que sustituyan estos
núcleos. No generar Java/Kotlin ni invocar JVM/JDK, Gradle, javac, kotlinc, d8,
dx, aapt/aapt2, apksigner, zipalign, Ant o frameworks que escondan esas herramientas.

Herramientas Android oficiales únicamente en pruebas externas explícitas, bajo
`tests/external` y build tag `android_external`. Nunca añadir procesos externos a
`cmd/koru`, `pkg/dex`, el builder o el runtime de compilación.

## Entorno y pruebas

En este workspace `/usr/bin/go` es un juego, no el compilador. Usar:

```sh
export PATH=/workspace/tooling/go/bin:$PATH
go test ./...
go vet ./...
go test -race ./...
```

Las skills de Addy Osmani están instaladas como plugin Codex y disponibles en
`/workspace/tooling/agent-skills`. Las referencias oficiales consultadas están en
`/workspace/tooling/android-specs` y su procedencia en `docs/specs/sources.json`.
`/workspace/tooling/android-validation/dexdump` solo sirve para validación externa.

Añadir pruebas golden y de entradas inválidas al extender formatos. El parser debe
seguir sin depender del builder o ensamblador. No silenciar errores con fallbacks.

No declarar funcional un APK hasta instalarlo y ejecutarlo en Android. dexdump y
el parser de koru no prueban ejecución ART, resolución de framework, renderizado,
ciclo de vida ni seguridad del WebView. El experimento file:// no es producción. El runtime normal usa HTTPS interceptado
y MessagePort, sin JavaScriptInterface. La matriz real API 30 está en docs/results/full.

Antes de commit: ejecutar gofmt, revisar git diff --check y comprobar que no se
versionan claves ni directorios .koru/android-demo. Los formateadores adicionales
se documentan en tests/README.md.
