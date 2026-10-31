# Remote content and hybrid apps

By default a koru app shows only the files in `webDir`. The optional `server` block in `koru.json` lets it load a page from your own website, open some sites inside the app and send other links to the system. This is **phase 1** of remote content, and it has **not been validated on an Android device yet**; see [Status](#status).

## Configure it

```json
"server": {
  "url": "https://example.com/app",
  "allowNavigation": ["example.com", "*.example.com"],
  "bridge": ["https://example.com"]
}
```

This is a field fragment, not a complete `koru.json`. All three fields are optional.

| Field             | Meaning and validation                                                                                                                                                               |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `url`             | First page to open instead of `webDir/index.html`. Must be `https`. It cannot be `app.koru.invalid`.                                                                                 |
| `allowNavigation` | Hosts that open **inside** the app, at most 32. `*.example.com` matches subdomains such as `app.example.com`, not `example.com` itself. The host of `url` is included automatically. |
| `bridge`          | Exact `https` origins, at most 32, with no wildcards and no path, whose pages receive the koru API. Each one must be loaded by the app: be in `url` or in `allowNavigation`.         |

`webDir` stays required. It still serves your local pages at `https://app.koru.invalid/`. Using `server` adds `android.permission.INTERNET` automatically. Without `server`, the runtime bytecode is byte for byte the same as before; a test with pinned hashes checks it.

## Three ways to use it

### An app that is a website

Point `url` at your site. Add `bridge` only if its pages should call the native API.

```json
"server": {
  "url": "https://example.com/app",
  "bridge": ["https://example.com"]
}
```

### A local app with your own site or login inside

Keep your local pages as the first screen and let some hosts open in the app instead of leaving it.

```json
"server": {
  "allowNavigation": ["login.example.com", "example.com"]
}
```

### External links that finally open the browser

Without `server`, a link to another site does nothing. With any `server` field, even an empty block, it opens in the system app. This is the whole configuration:

```json
"server": {}
```

## How navigation works

When `server` is present, a main-frame navigation is decided by its URL:

- An `https` URL whose host is in `url` or `allowNavigation`, or the local app origin, loads inside the app. Only `https` navigates inside the app.
- Any other link opens outside the app, with the system app, for the schemes `https`, `http`, `mailto`, `tel`, `sms`, `geo` and `market`.
- Other schemes, such as `intent:` or `javascript:`, are ignored.
- Iframes load where they are; this rule applies to main-frame navigation only.

Without `server`, behavior is unchanged: external links do nothing.

## Give remote pages the koru API

`bridge` lists the origins whose pages get the same plugins as the local app, including shared `storage`. The remote server is yours, so those pages must load `core.js` themselves; copy it into your site with `koru js`.

When `bridge` is set:

- The app adds ` KoruApp/1` to the WebView user agent.
- When each page finishes loading, the app origin and the `bridge` origins receive the native channel. The MessagePort is delivered only to that exact origin.
- Every other page receives `koru-none`, and `core.js` uses its web implementations.
- On a remote page with the marker, `Koru.platform` is `"android"` from the start, and calls wait until the page has finished loading.

## Security

`bridge` gives that site the **same native access as your app**. Any XSS or compromise of that site can use the API. List only origins you control, serve them over HTTPS and set a Content-Security-Policy. Remote pages do not get the CSP koru adds to local files; that header is yours to send. See [Signing and security]({{base}}docs/en/security/#remote-origins-and-the-bridge).

## Status

Not validated on a device. Two checks exist without one: the runtime built without `server` is identical to before, and `koru build` now passes all bytecode (the runtime and your own plugins) through a register type verifier similar to ART's (`internal/dex/verify`). It refuses to build an APK that ART would reject when loading the class. That is not execution on Android.

The self-test app in `examples/selftest` uses `allowNavigation: koru.js.org` and `bridge: https://koru.js.org`, with links to try it: a remote page at `https://koru.js.org/selftest/remote/` that checks the bridge and writes `KORU_SELFTEST_REMOTE` to logcat, an external link and a `mailto:` link. Its README explains them.
