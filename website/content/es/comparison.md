# koru, Capacitor y Expo

Capacitor y Expo son herramientas maduras y muy usadas. koru hace mucho menos, a propósito. Esta página te ayuda a decidir cuál encaja con tu proyecto.

## En pocas palabras

- Elige **koru** si tienes una app web, solo necesitas Android y quieres una cadena de herramientas mínima: un binario de Go que genera un APK firmado sin conexión en milisegundos, sin Java, Gradle ni el SDK de Android.
- Elige **Capacitor** si tu app web necesita iOS además de Android, un ecosistema amplio de plugins, paquetes para Google Play (AAB) o código nativo que escribes en Kotlin, Java o Swift.
- Elige **Expo** si quieres escribir la app en React Native, con componentes de interfaz nativos, iOS y Android, actualizaciones remotas y builds en la nube.

## Comparación

|                            | koru                                                                                             | Capacitor                                                                | Expo                                                         |
| -------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------ |
| Escribes                   | HTML, CSS, JavaScript; cualquier frontend estático                                               | HTML, CSS, JavaScript; cualquier framework web                           | React Native (JavaScript/TypeScript)                         |
| La interfaz se muestra en  | El WebView del sistema Android                                                                   | WebView (Android, iOS)                                                   | Vistas nativas                                               |
| Compilar Android requiere  | Solo el binario de koru                                                                          | Android Studio o Gradle, JDK y SDK de Android                            | EAS Build (en la nube), o JDK y SDK de Android en local      |
| Plataformas                | Android                                                                                          | Android, iOS, web                                                        | Android, iOS, web                                            |
| Resultado                  | APK firmado (v2)                                                                                 | APK, AAB                                                                 | APK, AAB                                                     |
| Empezar un proyecto        | `koru create`: plantillas `basic`, Vite, React, Vue, Svelte o TypeScript                         | `npm init @capacitor/app`                                                | `npx create-expo-app`                                        |
| Vista previa al programar  | `koru dev`: navegador, recarga en vivo, implementaciones web de la API                           | El servidor de tu framework; recarga en el dispositivo con `server.url`  | Expo Go o una development build, con fast refresh            |
| Ejecutar en un dispositivo | `koru run`                                                                                       | `npx cap run android`                                                    | `npx expo run:android`                                       |
| Logs del dispositivo       | `koru logs` en tu terminal (aún sin validar en un dispositivo)                                   | Logcat de Android Studio, o `chrome://inspect` para el WebView           | Terminal y herramientas de desarrollo con el servidor activo |
| Deep links                 | `deepLinks` en `koru.json`; abre la app, la URL aún no llega a JavaScript; sin probar en Android | Compatibles, con la URL en JavaScript (plugin App)                       | Compatibles mediante Expo Router y linking                   |
| Versionado                 | `koru bump`, `koru build --bump`                                                                 | Editar los archivos de proyecto de Gradle o Xcode                        | `version` en la configuración; EAS puede autoincrementar     |
| Funciones nativas          | 6 plugins integrados; plugins propios como manifiestos DEX o bibliotecas JNI                     | Plugins oficiales y de la comunidad; puedes escribir los tuyos en nativo | Módulos del SDK de Expo; puedes escribir los tuyos en nativo |
| Madurez                    | Experimental; registrado en Android 11 (API 30) x86_64                                           | Muy usado en producción                                                  | Muy usado en producción                                      |

## Qué hace koru de forma distinta

**Sin cadena de herramientas nativa.** koru escribe por sí mismo, en Go, el bytecode DEX, el manifiesto binario, la tabla de recursos, el ZIP y la firma del APK. Una build no descarga nada ni inicia una JVM, por eso es rápida y funciona sin conexión. A cambio, no puedes añadir código Java o Kotlin arbitrario: los plugins propios se describen como representación intermedia DEX o llaman a una biblioteca JNI precompilada.

**Un runtime pequeño y estricto.** Tus archivos se sirven desde un origen HTTPS privado, `https://app.koru.invalid`, y el puente JavaScript es un MessagePort ligado a ese origen. No usa `addJavascriptInterface`. El acceso a la red está desactivado salvo que añadas el permiso de Internet.

**La misma API en el navegador.** Como las implementaciones web de Capacitor, la API JavaScript de koru funciona en un navegador de escritorio durante `koru dev`, así que puedes construir la mayor parte de la app antes de instalarla.

## Qué no hace koru todavía

- Apps para iOS.
- Android App Bundles (AAB) para Google Play. koru genera APKs para instalación directa y tiendas que acepten APKs.
- Pasar la URL de un deep link a tu JavaScript. El enlace abre la app en su primera página, y ni el flujo de enlaces verificados ni nada de esto se ha ejecutado en Android.
- Notificaciones programadas, cámara, geolocalización, selectores de archivos y otros plugins fuera de los seis integrados.
- Recarga en vivo dentro del APK instalado. Usa `koru dev` en un navegador y luego `koru run`.
- Ejecución verificada más allá de la [matriz registrada]({{base}}docs/es/android-results/).

Si hoy necesitas alguna de esas funciones, Capacitor o Expo son mejor opción. Si necesitas una cadena de herramientas Android pequeña y auditable para una app web, prueba koru: [crea tu primera app]({{base}}docs/es/getting-started/).

_Capacitor y Expo cambian rápido; consulta la documentación de [Capacitor](https://capacitorjs.com/docs) y de [Expo](https://docs.expo.dev/) para conocer los detalles actuales._
