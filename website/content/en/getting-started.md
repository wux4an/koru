# Create your first app

Create a project, edit its page and open it on Android. You need Go 1.23 or later to build koru from source, and an Android device or emulator to test your app.

## Install koru

Clone the repository and build the executable:

```sh
git clone https://github.com/wux4an/koru.git
cd koru
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/koru ./cmd/koru
./bin/koru version
```

The first build may download Go dependencies. Once built, `bin/koru` runs on its own and produces APKs without Go, Java, Gradle or an Internet connection.

Run the following examples from this directory. If you add the executable to your `PATH`, you can use `koru` instead of `./bin/koru`.

## Create a project

```sh
./bin/koru init ./my-app
```

You will find these files:

```text
my-app/
├── koru.json
├── www/
│   └── index.html
└── .koru/
    └── signing.pem
```

`www` holds your web files. `koru.json` configures your app. `signing.pem` is its signing identity: keep a secure backup and exclude `.koru/` from Git. You will need the same key to update your app.

Open `my-app/www/index.html` and change the title or add your content. You can preview the web content in a browser; Android features need to be tested inside the APK.

## Give it a name

Replace the contents of `my-app/koru.json` with:

```json
{
  "appId": "com.example.hello",
  "name": "Hello Android",
  "version": "1.0.0",
  "versionCode": 1,
  "webDir": "www",
  "minSdk": 26,
  "targetSdk": 36,
  "permissions": [],
  "plugins": ["device", "storage", "app"],
  "signingKey": ".koru/signing.pem"
}
```

`name` is displayed on Android. `appId` identifies your app; choose your own before distributing it. See [Configuration]({{base}}en/configuration/) when you need to change the icon, permissions or web folder.

## Build the APK

```sh
./bin/koru doctor ./my-app
./bin/koru build ./my-app --output app-debug.apk
./bin/koru verify app-debug.apk
```

`doctor` checks your project. `build` writes `app-debug.apk` to the current directory. `verify` returns APK details and checks its signature and structure. Keep the APK output outside `www`.

This build lets you inspect the WebView during testing. When you are ready to share the app, make [a release build]({{base}}en/distribution/).

## Open the app on Android

Connect a device with USB debugging enabled, or start an emulator. With `adb` installed and the device connected, run:

```sh
adb install -r app-debug.apk
adb shell am start -W -n com.example.hello/com.koru.runtime.MainActivity
```

If you chose another `appId`, replace `com.example.hello` before the slash. You will see `www/index.html` inside your app. `adb` is used for installation and testing; you do not need it to build the APK.

Change something in the HTML, repeat `build` and install again. This is your edit and test workflow.

> Recorded koru execution covers Android 11 (API 30), x86_64. The configurable range is API 26–36; test your app on the devices you want to support.

## Keep building

- [Use your existing frontend]({{base}}en/existing-project/).
- [Store data and use Android features]({{base}}en/plugins/).
- [Prepare a version to share]({{base}}en/distribution/).

If a step fails, see [Troubleshooting]({{base}}en/troubleshooting/).
