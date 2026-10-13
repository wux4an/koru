# Demostración de Koru con VHS

La grabación usa [VHS de Charm](https://github.com/charmbracelet/vhs), versión 0.12.1,
y ejecuta realmente el binario Koru. No simula la salida de los comandos.

Muestra init, HTML/CSS/JS del ejemplo, build release con PATH sin herramientas,
verificación v2, recompilación con caché, igualdad byte a byte y doctor. El comando
Python de la escena de verificación solo da formato al JSON para hacerlo legible;
no participa en la compilación ni en la firma.

## Reproducir

Instalar VHS, ttyd, FFmpeg y Chromium y compilar Koru siguiendo el README principal.
Desde la raíz del repositorio:

```sh
mkdir -p artifacts
KORU_BIN="$PWD/bin/koru" \
KORU_WEB="$PWD/examples/web" \
VHS_BROWSER_PATH=/usr/bin/chromium \
vhs docs/demo/koru.tape
```

El tape crea un directorio temporal nuevo en /tmp/koru-vhs-\* y una identidad de
firma nueva mediante init. Las compilaciones posteriores de esa misma aplicación
reutilizan la identidad. Las claves nunca aparecen en pantalla ni se incorporan
al vídeo o al repositorio. El directorio temporal es material privado de prueba.

Salida: artifacts/koru-demo.mp4, 1440×900 a 25 fps. El poster se extrae del vídeo
con FFmpeg después de la grabación, eligiendo el frame de verificación de firma:

```sh
ffmpeg -y -ss 30 -i artifacts/koru-demo.mp4 -frames:v 1 artifacts/koru-demo-poster.png
```

La entrega MP4 se remultiplexó con `-c copy -movflags +faststart` para permitir
reproducción progresiva. Se decodificaron todos los frames sin errores y se
comprobaron los APKs creados en la grabación. Evidencia: validation.json.
La grabación es del CLI; la evidencia de instalación y ejecución Android está
separada en docs/results/full.

VHS y sus dependencias son herramientas de documentación y no son requisitos de
koru build. No se publica el vídeo automáticamente en servicios externos.
