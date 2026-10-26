# Run on a device

`koru run` builds your app, installs it on a connected phone or emulator and opens it. Use it whenever you want to see the real Android behavior: native plugins, permissions, the share sheet and notifications.

```sh
koru run
```

```text
Built com.example.myapp 1.0.0 (20 KB) in 35ms
Installed on sdk gphone64 x86 64 (emulator-5554)
Opened My App on sdk gphone64 x86 64 (emulator-5554); inspect it at chrome://inspect
```

> `koru run` and `koru devices` are new. They are tested against a program that implements the adb server protocol, not yet recorded on physical hardware. If something does not work on your setup, the manual commands at the end of this page always work.

## Prepare a device

**Phone.** Turn on [Developer options and USB debugging](https://developer.android.com/studio/debug/dev-options), connect the cable and accept the “Allow USB debugging?” prompt on the phone.

**Emulator.** Create and start a virtual device from Android Studio’s Device Manager. Images with API 26 or later work with koru’s `minSdk` range.

koru communicates with the **adb server**, the background service from Android’s platform tools that Android Studio starts for you. koru does not start programs itself. If no server is running:

```sh
adb start-server
```

Check what koru sees:

```sh
koru devices
```

```text
emulator-5554            device        sdk gphone64 x86 64
R5CT1234567              unauthorized
```

`device` means ready. `unauthorized` means the phone is waiting for you to accept the debugging prompt. `offline` usually means reconnecting the cable or restarting the emulator.

## Choose a device

With one ready device, `koru run` uses it. With several, choose by serial:

```sh
koru run --device emulator-5554
```

You can also set `ANDROID_SERIAL=emulator-5554`. koru honors `ANDROID_ADB_SERVER_PORT` and `ADB_SERVER_SOCKET=tcp:host:port`, the same variables as `adb`, for example to use a server on another machine.

## Debug the WebView

Debug builds (the default for `koru run`) enable WebView debugging. With the app open, visit `chrome://inspect` in Chrome on your computer and select your app to use DevTools: console, network, elements and breakpoints.

`koru run --release` installs a release build instead. Debugging is disabled, as in the APK you distribute.

## When installation fails

koru explains the common Android errors:

| Message                                 | Fix                                                                                                                         |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| …installed with a different signing key | The device has an app with the same `appId` signed by another key. Uninstall it on the device, or restore the original key. |
| …the device has a newer versionCode     | Increase `versionCode` in `koru.json`, or uninstall the newer app.                                                          |
| …older than minSdk                      | The device’s Android version is below `minSdk`. Use a newer device or emulator image.                                       |
| no adb server …                         | Run `adb start-server`, or open Android Studio.                                                                             |
| several devices are connected           | Add `--device <serial>`.                                                                                                    |

## Install manually

Every APK from `koru build` can also be installed with the platform tools directly:

```sh
koru build --output app.apk
adb install -r app.apk
adb shell am start -W -n com.example.myapp/com.koru.runtime.MainActivity
```

Replace `com.example.myapp` with your `appId`. To share the app with other people, see [Distribute your app]({{base}}en/distribution/).
