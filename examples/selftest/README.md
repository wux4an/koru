# koru self-test app

A koru app that tests koru. When it opens it runs every check that needs no
human: the JavaScript bridge and its error codes, each integrated plugin
(`app`, `device`, `storage`, `vibration`, `share`, `notifications`), the custom
`toast` plugin from [`../toast`](../toast) and WebView features (assets over
the app origin, workers, IndexedDB, WebAssembly, `crypto.subtle`, CSS). The
screen lists every result with its time and error code.

A second section has the checks that need you: sharing, the notification
permission prompt, a long vibration and a long toast.

## Download it

Every [release](https://github.com/wux4an/koru/releases/latest) has
`selftest.apk`, built from this folder with that release, and every CI run
keeps it as the `selftest-apk` artifact. Install it with
`adb install -r selftest.apk` or by opening the file on the phone. It is a
debug build, so the WebView can be inspected with DevTools.

## Build and run it

```sh
cd examples/selftest
mkdir -p .koru && koru keygen --output .koru/signing.pem   # once; never commit it

koru dev            # in a browser, with the web implementations
koru run --logs     # on a phone or emulator, then watch the logs
```

The report is available three ways:

- **On screen**, with a PASS/FAIL summary, also in the page title.
- **In the logs**, as one line: `koru logs | grep KORU_SELFTEST` prints
  `KORU_SELFTEST {"platform":"android","total":37,"passed":…,"failures":[…]}`.
- **For automation**, as `window.__KORU_SELFTEST__`, a promise of the full
  report. Evaluate it over the WebView DevTools protocol, for example with
  `python3 tests/android-cdp-eval.py "window.__KORU_SELFTEST__"` after
  `adb forward tcp:9223 localabstract:webview_devtools_remote_$(adb shell pidof dev.koru.selftest)`.
  "Copy report" puts the same JSON on the clipboard.

## What a result means

- **pass**: the behaviour matches the documented API on this platform.
- **skip**: not applicable here, such as the custom plugin in a browser or a
  notification without permission (use the interactive check).
- **fail**: a bug in koru or a difference on this device. The detail shows the
  error code and message; please report it with the copied JSON.

Expectations depend on the platform: in the APK the origin must be
`https://app.koru.invalid` and `sdk` at least 26; in a browser the API uses web
implementations and custom plugins answer `METHOD_NOT_FOUND`.

## Status

Under `koru dev` in Chromium 153 all checks pass except the two that are
skipped by design. The APK builds, passes `koru verify` in CI and ships with every release. It has **not**
run on an Android device yet: that is what this app is for.

## Web features

The app enables `webFeatures` (`files`, `geolocation`, `camera`, `microphone`).
Its **Web features** section has the buttons Pick files, Get location, Camera
photo and Microphone, plus automatic checks that the APIs exist. Tested in
Chromium with `koru dev`; not yet on an Android device. See
[Files, location, camera and microphone](https://koru.js.org/docs/en/web-features/).

## Navigation and remote bridge

This app sets `server.allowNavigation` to `koru.js.org` and `server.bridge` to
`https://koru.js.org`, and its page has three links to try (not yet run on a
device):

- **Remote page**, `https://koru.js.org/selftest/remote/`: opens inside the app,
  checks that the koru bridge reaches that origin and writes one line to the
  log: `koru logs | grep KORU_SELFTEST_REMOTE`.
- **External link**: a site outside `allowNavigation`; it should open in the
  system browser, not in the app.
- **`mailto:` link**: should open the system mail app.

See [Remote content and hybrid apps](https://koru.js.org/docs/en/remote-content/).
