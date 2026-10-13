# Security and signing

koru packages trusted application code. Its checks establish clear boundaries for assets, navigation and bridge messages; they do not sandbox a malicious signed plugin from the application.

## Keep one signing identity

`init` and `keygen` explicitly create a PEM private key and certificate, with permissions `0600`. `build` requires an existing, valid identity. `--release` reuses it and disables debugging. Back up the identity securely, outside version control and web assets. Android updates need the same signing certificate.

APK v2 uses RSA PKCS#1 v1.5/SHA-256, algorithm `0x0103`, and 1 MiB content digest blocks. Resources are aligned before signing; native libraries use 16 KiB alignment. The signed APK is not subsequently changed. `verify` checks signature, certificate, digest, ZIP, DEX, manifest, resources and icon.

## One local HTTPS origin

The runtime serves only packaged, allowlisted assets at `https://app.koru.invalid`. It checks scheme and authority rather than concatenating incoming paths onto host filesystem paths. File and content access are disabled. Unknown assets return 404; SPA fallback only applies to paths without an extension.

Service workers use the same allowlist through Application context and do not retain the Activity. Network access is blocked unless the app explicitly declares `INTERNET`. CSP blocks external scripts and frames, while allowing packaged application scripts, inline/eval needed by some web outputs, and authorized HTTPS connections. This is not a strict no-inline CSP.

## MessagePort bridge

The native port is transferred to the main frame with an explicit target origin. There is no `addJavascriptInterface`. The receiver checks its document origin, main-frame context, native trusted event and absent Window source.

Requests use protocol v1, IDs, method names and parameters. Only enabled plugin handlers are accepted. The client limits pending requests, applies timeouts, rejects outstanding promises on page exit, and can cancel before dispatch. Cancellation does not undo an action Android already executed.

Custom DEX and JNI plugins share the application's UID and permissions. They are trusted developer code. Validate parameters, respect UI-thread requirements and provide the correct native ABI.

## What verification proves

A structurally valid APK is not proof of successful ART resolution, lifecycle, rendering or safe behavior for arbitrary application code. The recorded execution checks apply to the tested runtime and API 30/x86_64 matrix. Validate your own artifacts on Android.

[Runtime decisions]({{base}}en/architecture-apk/) · [Android execution evidence]({{base}}en/android-results/)
