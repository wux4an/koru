# Archivos, ubicación, cámara y micrófono

El campo opcional `webFeatures` de `koru.json` permite que tus páginas usen el selector de archivos, la ubicación, la cámara y el micrófono con las APIs web de siempre. Es la **fase 2** de las funciones web. Se probó en Chromium con `koru dev` (las APIs web) y su bytecode pasa el verificador de `koru build`, pero **aún no se ha probado en un dispositivo Android**.

## Configúralo

```json
"webFeatures": ["files", "geolocation", "camera", "microphone"]
```

Es un fragmento, no un `koru.json` completo. Usa cualquier subconjunto, sin duplicados.

| Valor         | Qué activa                                              | Permisos que añade                               |
| ------------- | ------------------------------------------------------- | ------------------------------------------------ |
| `files`       | `<input type="file">` abre el selector de Android.      | Ninguno                                          |
| `geolocation` | `navigator.geolocation`, aproximada o precisa.          | `ACCESS_COARSE_LOCATION`, `ACCESS_FINE_LOCATION` |
| `camera`      | `navigator.mediaDevices.getUserMedia({ video: true })`. | `CAMERA`                                         |
| `microphone`  | `navigator.mediaDevices.getUserMedia({ audio: true })`. | `RECORD_AUDIO`, `MODIFY_AUDIO_SETTINGS`          |

Sin `webFeatures` el runtime no cambia, byte a byte. Con él, el código generado pasa por el mismo verificador de bytecode que el resto de `koru build`.

## Archivos

`<input type="file">` abre el selector de Android (documentos o galería, según `accept`). Añade `multiple` para permitir elegir varios archivos. Funciona en cualquier página cargada en la app y no añade permisos.

```html
<input id="pick" type="file" accept="image/*" multiple />
<script>
  document.getElementById("pick").addEventListener("change", (event) => {
    for (const file of event.target.files) {
      console.log(file.name, file.type, file.size);
    }
  });
</script>
```

El atributo `capture` **no** abre la cámara directamente; se usa el selector. Para hacer una foto con la cámara, usa `getUserMedia` con `camera` (más abajo).

## Ubicación

```js
navigator.geolocation.getCurrentPosition(
  (position) => {
    const { latitude, longitude, accuracy } = position.coords;
    console.log(latitude, longitude, accuracy);
  },
  (error) => console.error(error.code, error.message),
  { enableHighAccuracy: true },
);
```

La primera vez, Android pide permiso al usuario. Vale la ubicación aproximada o la precisa.

## Cámara

```js
const stream = await navigator.mediaDevices.getUserMedia({ video: true });
document.querySelector("video").srcObject = stream;
await document.querySelector("video").play();
// Al terminar:
stream.getTracks().forEach((track) => track.stop());
```

Para hacer una foto, dibuja un fotograma del vídeo en un `<canvas>` y llama a `canvas.toBlob()`.

## Micrófono

```js
const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
const recorder = new MediaRecorder(stream);
const chunks = [];
recorder.ondataavailable = (event) => chunks.push(event.data);
recorder.onstop = () => {
  const audio = new Blob(chunks, { type: recorder.mimeType });
  console.log(audio.size);
};
recorder.start();
```

## Qué páginas tienen acceso

`geolocation`, `camera` y `microphone` solo se conceden al origen de la app (`https://app.koru.invalid`) y a los orígenes de `server.bridge`. Cualquier otra página recibe un rechazo. Una página que pide un recurso que no activaste (por ejemplo el micrófono cuando solo figura `camera`), o MIDI o medios protegidos, también se rechaza. Android pide solo los permisos que la página necesita. Consulta [Firmas y seguridad]({{base}}docs/es/security/) y [Contenido remoto y apps híbridas]({{base}}docs/es/remote-content/).

## En el navegador

Con `koru dev` estas APIs ya funcionan como en cualquier web, porque `localhost` es un contexto seguro.

## Pruébalo

La app [`examples/selftest`](https://github.com/wux4an/koru/tree/main/examples/selftest) activa los cuatro valores. Su sección **Web features** tiene los botones Pick files, Get location, Camera photo y Microphone, además de comprobaciones automáticas de que existen las APIs.

## Estado

Probado en Chromium con `koru dev`, y el bytecode está verificado. Aún no se ha probado en un dispositivo Android, así que no lo des por funcional allí hasta instalarlo y ejecutarlo. Consulta [Configura tu proyecto]({{base}}docs/es/configuration/).
