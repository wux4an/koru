# Distribute your app

Build a release APK to share and install directly on Android. Koru does not yet generate AABs for publishing on Google Play.

## Prepare the version

Before sharing your first APK, review `appId`, `name`, `version` and `icon` in `koru.json`. Test your screens, navigation and plugins on your target devices.

Keep a secure backup of `.koru/signing.pem` outside the project. Android uses this identity to recognize updates to your app. Keep it out of Git, the web folder and any files you distribute with the APK.

## Build a release

From the directory that contains your project `my-app`:

```sh
koru doctor ./my-app
koru build ./my-app --release --output app-release.apk
koru verify app-release.apk
```

`--release` disables debug mode and WebView inspection. It uses the key configured in `signingKey`; it does not create a new identity. The resulting file is `app-release.apk` in the current directory.

Install and open this APK too: `verify` checks its signature and structure, but does not run your app. `koru run --release` installs a release build on a connected device in one step.

## Share the file

Send the APK or host it as a download. The person installing it will need to allow their browser or file manager to install apps from that source. The permission wording and location depend on Android and the device.

To test with a connected device:

```sh
adb install -r app-release.apk
```

Open the app from the launcher and check its behavior before sharing it.

## Publish an update

Keep the same `appId` and signing key. Increase `versionCode` and change `version`. Let koru do it:

```sh
koru bump minor                     # version 1.0.0 -> 1.1.0, versionCode 1 -> 2
koru build --release --bump --output app-release.apk   # patch bump, then build
```

`bump` takes `patch` (the default), `minor`, `major` or `code` (only `versionCode`). Use `--bump` on `build` for a patch release, or run `bump` first for another level, but not both: that would raise the version twice. You can still edit the fields in `koru.json` by hand, for example:

```json
"version": "1.1.0",
"versionCode": 2
```

Rebuild your frontend, generate the release APK and install it over the previous version. Check that saved data is preserved.

If Android rejects the update because the signature differs, restore the original key. Creating another key with `keygen` does not fix an update. Uninstalling the old app removes its data; avoid using that as a solution for your users.

## Deep links

If your app should open from web links, add `deepLinks` to `koru.json` and publish the `assetlinks.json` that `koru deeplinks` prints on your domain, so Android can verify the `https` links. Details and limits, including that the link does not reach your JavaScript yet, are in [Configure your project]({{base}}en/configuration/). Nothing about deep links has been run on Android; test them on a device before you ship.

See [Signing and security]({{base}}en/security/) for identity management and [Troubleshooting]({{base}}en/troubleshooting/) for installation errors.
