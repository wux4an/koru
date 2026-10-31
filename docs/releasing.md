# Publicar una versión

Las versiones se publican desde GitHub Actions al subir un tag `vX.Y.Z`
(`.github/workflows/release.yml`). Nadie compila binarios a mano.

## Pasos

1. Actualiza `version` en `cmd/koru/main.go` y `packages/javascript/package.json`.
2. Mueve lo de `## [Unreleased]` de `CHANGELOG.md` a una sección
   `## [X.Y.Z] - AAAA-MM-DD` y actualiza los enlaces del final.
3. Comprueba en local:

   ```sh
   scripts/release.sh vX.Y.Z      # archivos en dist/ (borra dist/ antes)
   scripts/test-install.sh dist   # instalador contra esos archivos
   ```

4. Commit en `main`, y después:

   ```sh
   git tag -a vX.Y.Z -m "koru vX.Y.Z"
   git push origin vX.Y.Z
   ```

## Qué hace el workflow

1. Ejecuta el CI completo (`ci.yml`): gofmt, vet, tests con `-race`, build en
   Linux, macOS y Windows, módulo JS, shellcheck y prueba del instalador.
2. `scripts/release.sh` compila con `CGO_ENABLED=0 -trimpath` para
   linux, darwin y windows en amd64 y arm64, e inyecta versión y commit con
   `-ldflags`. Falla si el tag no coincide con `cmd/koru/main.go`.
3. Publica por cada plataforma `koru_X.Y.Z_<os>_<arch>.tar.gz` (`.zip` en
   Windows) con su `.sha256`, un `SHA256SUMS` con todos, e `install.sh` /
   `install.ps1`.
4. Firma la procedencia con `actions/attest-build-provenance` (Sigstore). Se
   verifica con `gh attestation verify <archivo> --repo wux4an/koru`.
5. Crea la release con las notas de la sección de `CHANGELOG.md` (falla si no
   existe). Los tags con sufijo (`v1.0.0-rc.1`) salen como prerelease; el resto
   como última versión, que es la que leen los instaladores y la web.
6. Instala la versión publicada con `install.sh` en Linux x86_64 y arm64 y en
   macOS, y con `install.ps1` en Windows, y comprueba `koru version`.

Para rehacer una release existente: *Actions → Release → Run workflow* con el
tag. Sustituye los archivos y las notas.

## Builds reproducibles

Mismo commit y misma versión de Go dan los mismos bytes: `-trimpath`,
`-buildid=` vacío, y archivos con fecha del commit, orden fijo y sin
usuario. Para comprobar una release, recompila el tag con la versión de Go del
workflow y compara `SHA256SUMS`.

## Instaladores

- `scripts/install.sh` (POSIX `sh`, Linux y macOS): resuelve la última versión
  siguiendo la redirección de `/releases/latest` (sin la API ni su límite),
  descarga por HTTPS, verifica el SHA-256 antes de instalar y escribe en
  `~/.local/bin` de forma atómica. Opciones: `--version`, `--dir`,
  `KORU_VERSION`, `KORU_INSTALL_DIR`.
- `scripts/install.ps1` (Windows): igual, instala en
  `%LOCALAPPDATA%\Programs\koru` y lo añade al `PATH` del usuario.

El sitio publica ambos en `https://koru.js.org/install.sh` y
`https://koru.js.org/install.ps1` al desplegarse.
