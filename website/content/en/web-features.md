# Files, location, camera and microphone

The optional `webFeatures` field in `koru.json` lets your pages use the file picker, location, camera and microphone through ordinary web APIs. This is **phase 2** of web features. It was tested in Chromium with `koru dev` (the web APIs) and its bytecode passes the `koru build` verifier, but it has **not been tested on an Android device yet**.

## Configure it

```json
"webFeatures": ["files", "geolocation", "camera", "microphone"]
```

This is a field fragment, not a complete `koru.json`. Use any subset, without duplicates.

| Value         | What it enables                                         | Permissions added                                |
| ------------- | ------------------------------------------------------- | ------------------------------------------------ |
| `files`       | `<input type="file">` opens the Android picker.         | None                                             |
| `geolocation` | `navigator.geolocation`, approximate or precise.        | `ACCESS_COARSE_LOCATION`, `ACCESS_FINE_LOCATION` |
| `camera`      | `navigator.mediaDevices.getUserMedia({ video: true })`. | `CAMERA`                                         |
| `microphone`  | `navigator.mediaDevices.getUserMedia({ audio: true })`. | `RECORD_AUDIO`, `MODIFY_AUDIO_SETTINGS`          |

Without `webFeatures` the runtime does not change, byte for byte. With it, the generated code goes through the same bytecode verifier as the rest of `koru build`.

## Files

`<input type="file">` opens the Android picker (documents or gallery, depending on `accept`). Add `multiple` to allow selecting several files. It works on any page loaded in the app and adds no permissions.

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

The `capture` attribute does **not** open the camera directly; the picker is used. To take a photo with the camera, use `getUserMedia` with `camera` (below).

## Location

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

The first time, Android asks the user for permission. Approximate or precise location is enough.

## Camera

```js
const stream = await navigator.mediaDevices.getUserMedia({ video: true });
document.querySelector("video").srcObject = stream;
await document.querySelector("video").play();
// When you are done:
stream.getTracks().forEach((track) => track.stop());
```

To take a photo, draw a video frame on a `<canvas>` and call `canvas.toBlob()`.

## Microphone

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

## Which pages get access

`geolocation`, `camera` and `microphone` are granted only to the app origin (`https://app.koru.invalid`) and to the origins in `server.bridge`. Any other page receives a rejection. A page that asks for a resource you did not enable (for example the microphone when only `camera` is listed), or for MIDI or protected media, is also rejected. Android asks only for the permissions that the page needs. See [Signing and security]({{base}}docs/en/security/) and [Remote content and hybrid apps]({{base}}docs/en/remote-content/).

## In the browser

With `koru dev` these APIs already work as on any website, because `localhost` is a secure context.

## Try it

The [`examples/selftest`](https://github.com/wux4an/koru/tree/main/examples/selftest) app enables all four values. Its **Web features** section has the buttons Pick files, Get location, Camera photo and Microphone, plus automatic checks that the APIs exist.

## Status

Tested in Chromium with `koru dev`, and the bytecode is verified. Not tested on an Android device yet, so do not treat it as working there until you have installed and run it. See [Configure your project]({{base}}docs/en/configuration/).
