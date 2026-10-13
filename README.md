# koru

Toolchain Android independiente que transforma una salida web estática en un APK
firmado. El build normal está implementado en Go con biblioteca estándar: DEX,
AXML, ARSC, ZIP alineado y firma APK v2. No genera Java/Kotlin ni usa JVM/JDK,
Gradle, Android Studio, d8/dx, aapt, apksigner, zipalign o frameworks de wrappers.

**Se ha instalado y ejecutado en Android 11/ART**: WebView muestra HTML, CSS e
imágenes locales y ejecuta JavaScript con Wi‑Fi y datos desactivados. El puente,
los plugins estándar y el ejemplo JNI opcional tienen pruebas reales. Evidencia:
[docs/results/full](docs/results/full), además de los resultados históricos DEX.

## Uso

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru init ./mi-app
./bin/koru build ./mi-app --output app.apk
./bin/koru verify app.apk
./bin/koru build ./mi-app --release --output app-release.apk
./bin/koru doctor ./mi-app
```

El binario no requiere Go ni red para compilar aplicaciones. Para compilar **koru**
en este workspace, el Go real está en `/workspace/tooling/go/bin`:

```sh
export PATH=/workspace/tooling/go/bin:$PATH
```

`init` crea configuración, ejemplo HTML e identidad persistente en `.koru/signing.pem`
con permisos 0600. Guardar esa identidad de forma segura para futuras actualizaciones.
`build` nunca crea ni cambia una clave si falta o es inválida. `--release` reutiliza
la clave configurada y deshabilita debug/DevTools; se puede seleccionar otra identidad
existente con `--keystore`. `keygen --output firma.pem` crea una identidad nueva
explícitamente y nunca sobrescribe un archivo existente.

## Proyectos existentes y configuración

Koru acepta HTML/CSS/JS y carpetas de salida de React, Vue, Svelte u otros frameworks.
Los compiladores de esos frameworks se ejecutan antes; koru no los implementa.
Después de init, cambiar `webDir` por la carpeta estática deseada:

```json
{
  "appId": "com.example.demo",
  "name": "Demo",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "./dist",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app", "vibration", "share", "notifications"],
  "signingKey": ".koru/signing.pem",
  "spa": true
}
```

Se validan SDKs (26 ≤ minSdk ≤ targetSdk ≤ 36), nombres, versiones, permisos,
plugins y rutas. `icon` puede apuntar a un PNG. Los plugins agregan solamente sus
permisos declarados: VIBRATE y POST_NOTIFICATIONS para los casos correspondientes.
Para servicios externos HTTPS, declarar INTERNET explícitamente y configurar el
servidor para el origen de la app. Los scripts e iframes externos están bloqueados;
los scripts de la aplicación se empaquetan localmente.

El origen es `https://app.koru.invalid`, servido por clases DEX que interceptan
assets. No se usa file:// en el runtime normal ni AndroidX. Service workers usan
la misma allowlist desde Application context. Existe una opción SPA; desconocidos
con extensión devuelven 404. Se rechazan symlinks, claves privadas dentro de assets,
entradas inseguras y límites de tamaño. El APK debe escribirse fuera de webDir.

La caché de runtime evita regenerar DEX si solo cambia contenido web. Incluye
fingerprint del toolchain/configuración/rutas/plugins y SHA-256 del DEX. Corrupción
produce un error explícito; `--no-cache` desactiva la caché.

## API y plugins

```js
import { device, storage } from '/__koru/core.mjs';
console.log(await device.getInfo());
await storage.set('theme', 'dark');
```

El paquete `packages/javascript` se llama `@koru/core`. También está disponible
`Koru.invoke("device.info", {})`. El puente usa MessagePort, targetOrigin explícito y frame principal; no usa
addJavascriptInterface. Tiene IDs, errores estructurados, promesas, timeouts,
validación y cancelación antes del dispatch.

Los seis plugins iniciales y la ABI extensible se documentan en
[docs/plugins.md](docs/plugins.md). Plugins personalizados se suministran como IR
DEX y manifests, sin Java. Las bibliotecas JNI son opcionales, precompiladas por
ABI y alineadas a 16 KiB; el ejemplo x86_64 y su código C están incluidos. Los
plugins estándar no requieren C ni bibliotecas propias.

## CLI

`init`, `build`, `build --release`, `verify`, `doctor`, `version` y `keygen`.
El compilador DEX también puede utilizarse de forma independiente:

```sh
./bin/koru dex build examples/dex/answer.json --output answer.dex
./bin/koru dex inspect answer.dex
./bin/koru dex build --activity-experiment --output experiment.dex
```

`--activity-experiment` conserva el primer experimento file:// para comparación;
**no se utiliza en build**. La API Go está en [pkg/dex](pkg/dex), con un
[ejemplo](examples/dex/main.go) y [documentación](docs/dex-api.md).

## Pruebas

```sh
go test ./...
go vet ./...
go test -race ./...
node --test packages/javascript/core.test.cjs  # solo pruebas del cliente JS
```

El build no requiere Node. Los fuzz targets están en encoding, parser DEX, AXML,
ARSC y signing. Las pruebas externas Android quedan en tests y nunca en el build:
`dexdump`, `aapt`, `apksigner`, ADB, emulador y un JRE de pruebas se usaron únicamente
para contraste/ejecución. La entrega no contiene esas herramientas ni un JDK.

Las pruebas reales WebView usan `tests/android-smoke.py`, `android-navigation.py`,
`android-worker.py` y `android-keyboard.py`, con un APK debug instalado y un puerto
DevTools reenviado por ADB. Python/websockets son herramientas de QA opcionales.

## Arquitectura y alcance

DEX format/encoding/model/assembler/builder/parser son módulos separados. AXML,
ARSC, APK y signing no dependen de herramientas Android. Config/app integran esos
módulos; runtime/dex genera clases Android; runtime/native es opcional. Decisiones:
[ADR DEX](docs/adr/0001-dex-core.md) y [ADR APK/runtime](docs/adr/0002-apk-runtime.md).

Es una primera versión funcional, con una matriz Android limitada al emulador
Android 11 x86_64. No implica certificación de todos los fabricantes/API ni
aceptación en Play. La distribución actual es APK para instalación directa; no
se generan AAB. Las notificaciones son inmediatas, no alarmas futuras. Firma v3/v4,
multidex, anotaciones generales y verificación completa de tipos ART permanecen
fuera del subconjunto inicial. La ejecución en ART complementa la validación
estructural del parser propio. Windows/macOS tienen pruebas de compilación,
pero Linux es el host operativo validado. No hay comparativas con Capacitor.

Los [agent-skills de Addy Osmani](https://github.com/addyosmani/agent-skills) están
instalados como plugin Codex `agent-skills@agent-skills` y se aplicaron al desarrollo.
Son herramientas del agente, no dependencias de koru. Código propio: MIT; header
JNI AOSP: Apache-2.0 con licencia incluida.

Entrega reproducible: `python3 scripts/package.py --output koru-delivery.zip` después
de compilar bin/koru y generar los APKs de ejemplo. Solo incluye archivos
versionados, el ejecutable y dos APKs; rechaza claves y directorios privados.
