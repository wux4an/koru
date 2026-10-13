# Ejemplo web

`examples/web` contiene HTML, CSS, SVG y JavaScript sin dependencias. Es el fixture usado en las pruebas Android registradas para comprobar renderizado, contador, plugins nativos, navegación y funcionamiento offline.

Cópialo en `webDir` y compila con koru. Para la prueba de service worker, copia además `tests/fixtures/sw.js` como `sw.js`.

El README original describe una etapa anterior en la que el builder APK aún no existía. Esa afirmación histórica se conserva en el archivo de fuentes. La ejecución posterior está documentada en la [evidencia Android]({{base}}es/android-results/).

## Consulta el fixture

- [index.html]({{base}}original/examples/web/index.html ":ignore")
- [app.js]({{base}}original/examples/web/app.js ":ignore")
- [style.css]({{base}}original/examples/web/style.css ":ignore")
- [icon.svg]({{base}}original/examples/web/icon.svg ":ignore")
- [second.html]({{base}}original/examples/web/second.html ":ignore")

El archivo sirve los HTML como texto fuente; no ejecuta el ejemplo Android en tu navegador.
