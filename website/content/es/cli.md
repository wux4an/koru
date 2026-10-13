# Comandos CLI

Ejecuta `koru --help` o `koru <comando> --help` para consultar los flags de tu versión. Los comandos y su ayuda están en inglés. El binario construye y verifica paquetes sin herramientas de build Android.

## Comandos de proyecto

| Comando                                | Función                                                                                              |
| -------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `koru init [directorio]`               | Crea configuración, assets de ejemplo e identidad de firma persistente. Directorio por defecto: `.`. |
| `koru build [proyecto]`                | Genera un APK debug. Proyecto por defecto: `.`; salida: `app.apk`.                                   |
| `koru verify <app.apk>`                | Comprueba firma APK v2 y estructura; devuelve un reporte JSON.                                       |
| `koru doctor [proyecto]`               | Comprueba configuración, identidad existente, index web y manifests personalizados.                  |
| `koru keygen --output <identidad.pem>` | Crea explícitamente una identidad sin sobrescribir archivos.                                         |
| `koru version`                         | Muestra la versión de la toolchain.                                                                  |

## Opciones de build

```sh
koru build ./mi-app --release --output app-release.apk
koru build ./mi-app --keystore .koru/otra-existente.pem --no-cache
```

| Flag         | Significado                                              |
| ------------ | -------------------------------------------------------- |
| `--output`   | Destino firmado con extensión `.apk`, fuera de `webDir`. |
| `--release`  | Deshabilita debug y DevTools; reutiliza la identidad.    |
| `--keystore` | PEM existente; rutas relativas desde el proyecto.        |
| `--no-cache` | Desactiva la caché DEX del runtime.                      |

Conserva la clave y su certificado para actualizar la app. Release no crea una identidad nueva automáticamente.

## Comandos DEX independientes

```sh
koru dex build examples/dex/answer.json --output answer.dex
koru dex inspect answer.dex
koru dex build --activity-experiment --output experiment.dex
```

`dex build` acepta el IR DEX en JSON. `dex inspect` devuelve una inspección JSON decodificada después de validar. `--activity-experiment` conserva el experimento inicial con `file://` para comparación. El build APK normal usa HTTPS interceptado.

El límite de entrada DEX es 64 MiB. La validación estructural no ejecuta ART ni comprueba todos los tipos.

[API DEX]({{base}}es/dex-api/) · [Primer build]({{base}}es/getting-started/)
