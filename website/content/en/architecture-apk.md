# ADR 0002: Go APK, HTTPS origin and plugins

## Package and signature

AXML and a minimal ARSC table containing the mipmap icon are emitted directly. The custom ZIP32 packager computes CRC, compresses through Go's flate, fixes DOS timestamps and aligns stored resources to four bytes. Optional native libraries use 16 KiB alignment. Alignment and contents are finalized before inserting the v2 signature; the signed APK is not modified.

Signing uses RSA PKCS#1 v1.5/SHA-256 (v2 algorithm `0x0103`), 1 MiB digest blocks and Android's specified signer structure. The PEM identity contains the key and certificate, is explicitly created through init/keygen, has permissions `0600`, and is never regenerated during build. Release reuses that identity and disables debug and remote WebView inspection. The verifier checks signature, certificate, digest, ZIP, DEX, manifest, resources and icon.

## Local origin and content policy

The application is served at `https://app.koru.invalid`. A generated WebViewClient checks scheme/authority and uses an exact route allowlist. It does not concatenate received paths with filesystem paths. It only opens assets within its APK; read errors yield 404. SPA mode returns `index.html` for extensionless routes.

Service-worker callbacks use the same allowlist and Application context without retaining Activity. File/content access is disabled. Network is blocked unless the project explicitly declares INTERNET. CSP prevents external scripts and iframes, allows inline/eval required by certain web outputs, and allows HTTPS connections when the app authorizes network. Packaged scripts are trusted code.

## Bridge and request lifecycle

The bridge uses WebMessagePort and transfers it to the main frame with an explicit targetOrigin. It does not use addJavascriptInterface or require Java method-export annotations. The receiver checks its document origin, main-frame context, and a trusted native event without Window source. It does not interpret a native event's origin field as a sender page's origin.

Protocol v1 carries id, method, params and result/error. Enabled plugins compile into the dispatcher, and only their routes are accepted. Input sizes and types are validated; errors become structured codes. The client bounds pending requests, supports timeouts and pre-dispatch cancellation, and closes promises on page exit. A completed Android action cannot be undone by Promise cancellation.

Notification permission resolves asynchronously: the exact request ID and port are retained, and the response is sent from the permission callback or return from settings. A later navigation cannot receive an earlier page's response. API 33+ uses POST_NOTIFICATIONS; earlier APIs check notification availability and can open user settings.

## Plugin ABI and native code

DEX plugin ABI v1 requires a public static method receiving Activity/JSONObject and returning JSONObject. Manifests cannot replace Android, Java or runtime classes. Plugins are trusted developer code sharing the application's UID and permissions, not a sandbox for hostile native code.

JNI is optional. The six built-in plugins use DEX/framework and require no C. The freestanding x86_64 example has no libc/JVM dependencies and has loaded in ART. It demonstrates global references, local-reference cleanup, exceptions and attach/detach scopes. Bridge UI operations run on main.

## Runtime cache

The DEX cache depends on the koru binary, configuration, plugins and asset names/MIME types. Changing only web bytes preserves the runtime. SHA-256 and structural verification detect corruption, which produces an explicit error rather than fallback. Disable the cache with `--no-cache`.
