# Web example

`examples/web` contains dependency-free HTML, CSS, SVG and JavaScript. It is the fixture used in the recorded Android tests to exercise rendering, a counter, native plugins, navigation and offline behavior.

Copy it into your configured `webDir` and build through koru. For the service-worker test, additionally copy `tests/fixtures/sw.js` as `sw.js`.

The original example README describes an earlier stage when the APK builder had not yet been implemented. That historical statement is retained in the source archive. Later execution is documented in the [Android evidence]({{base}}en/android-results/).

## Inspect the fixture

- [index.html]({{base}}original/examples/web/index.html ":ignore")
- [app.js]({{base}}original/examples/web/app.js ":ignore")
- [style.css]({{base}}original/examples/web/style.css ":ignore")
- [icon.svg]({{base}}original/examples/web/icon.svg ":ignore")
- [second.html]({{base}}original/examples/web/second.html ":ignore")

The archive serves HTML files as plain source text; it does not execute the Android example in your browser.
