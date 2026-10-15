# Signing and security

Your app includes a signing identity and its own web files. Protect the key and configure the permissions your app needs.

## Keep your signing key

`koru init` creates `.koru/signing.pem` with a private key and certificate. Keep a secure backup outside the project and exclude `.koru/` from version control. Keep the key outside `webDir`.

Android requires the same certificate to update an installed app. Debug and release both use `signingKey` from `koru.json`. If the key is missing or invalid, the build fails: restore the original identity before continuing.

`koru keygen --output new-identity.pem` creates another identity, but cannot replace the identity of an app you have already distributed. See [Distribute your app]({{base}}en/distribution/) to prepare updates.

## Include only application files

Point `webDir` to the web output you want to distribute, containing `index.html`. Keep credentials, private files and unrelated content out of that folder. Koru rejects detected private keys, symbolic links and unsafe paths, but you should still review what you package.

Files load from `https://app.koru.invalid`. This origin belongs to the app’s local content; it does not require a domain or an external server. Remote scripts and iframes are blocked. Include your scripts in the web folder.

## Allow the network access you need

Local content works without Internet permission. If you consume an external API, declare `android.permission.INTERNET`, use HTTPS and configure CORS for the app origin. See [Use your web project]({{base}}en/existing-project/) for an example.

Custom plugins run with your app’s permissions. Include only trusted plugins and native libraries.

## Check the APK you share

```sh
./bin/koru build ./my-app --release --output app-release.apk
./bin/koru verify app-release.apk
```

Release disables debugging and WebView inspection. `verify` checks the signature and APK structure. Install it and test your app’s features on Android too.

For the HTTPS origin, MessagePort and WebView boundaries, see [APK and runtime]({{base}}en/architecture-apk/).
