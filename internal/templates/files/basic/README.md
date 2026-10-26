# __APP_NAME__

An Android app built from web files with [koru](https://koru.js.org/).

```sh
koru dev            # preview in your browser with live reload
koru run            # build, install and open on a connected device
koru build --release --output __APP_SLUG__.apk
```

- `www/` holds the web files that go into the APK.
- `koru.json` sets the app name, id, version and enabled plugins.
- `.koru/signing.pem` is your signing key. Back it up and never commit it:
  Android only accepts updates signed with the same key.
