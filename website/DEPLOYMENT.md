# Publish koru documentation

The repository is `wux4an/koru`. The GitHub Pages workflow in
`.github/workflows/pages.yml` builds `website/dist` with Bun 1.2.13, runs content
and formatting checks, and publishes the static artifact. Action versions are
pinned to reviewed release commits. Node/Bun exist only in this documentation
workflow and are unrelated to the Go APK compiler.

## GitHub Pages

Enable **GitHub Actions** as the Pages source in the repository settings. The
workflow runs on pushes to main that change documentation, and can also be run
manually. It takes the host and path from `actions/configure-pages` and builds
canonical URLs with HTTPS. The site was first published at
`https://wux4an.github.io/koru/` and now uses
`https://koru.js.org/`. After changing a custom domain, run the workflow again:
the artifact must be rebuilt to update its base path, canonical URLs, language
links and sitemap. Changing the Pages setting alone does not rebuild it.

[Pages settings](https://github.com/wux4an/koru/settings/pages) ·
[Documentation workflow](https://github.com/wux4an/koru/actions/workflows/pages.yml)

Pages serves the generated directory indexes directly. The 404 page is a real
404 response; there is no SPA catch-all. `robots.txt` within the initial `/koru/`
subpath does not control the whole github.io host. For a dedicated custom domain,
the generated robots file is at the required domain root.

GitHub Pages does not provide custom response headers. For that host, the website
build archives example HTML/SVG as plain `.txt` downloads and wraps original
Markdown/source/log files with a noindex HTML text viewer. Original bytes remain
available inside those viewers as downloads; source code is never executed.

## Custom domain: koru.js.org

The approved domain is `koru.js.org`. The project is a JavaScript application
toolchain with a JavaScript bridge/API; the documentation describes that purpose
and includes substantive content.

1. Check that the GitHub Pages site has deployed and contains the documentation.
2. Set `koru.js.org` as the custom domain in the repository's Pages settings.
   For Actions deployments, GitHub uses this setting; a CNAME file in the artifact
   is not sufficient.
3. Submit your request in `js-org/js.org`, adding the following alphabetically to
   `cnames_active.js`:

   ```js
   "koru": "wux4an.github.io",
   ```

4. Rerun the documentation workflow after setting the domain, then enable
   **Enforce HTTPS** when GitHub has provisioned the certificate.

The expected documentation paths are `https://koru.js.org/en/` and
`https://koru.js.org/es/`, with the icon at `https://koru.js.org/favicon.ico`.

Official references:

- [JS.ORG instructions and content requirements](https://github.com/js-org/js.org/blob/master/README.md)
- [JS.ORG contribution guide](https://github.com/js-org/js.org/blob/master/CONTRIBUTING.md)
- [GitHub Pages workflows](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages)
- [GitHub custom domains](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site)

## Search engines and variables

Settings, in Settings > Secrets and variables > Actions of the repository:

| Name                       | Kind     | Purpose                                                                              |
| -------------------------- | -------- | ------------------------------------------------------------------------------------ |
| `SITE_URL`                 | variable | Public URL, for example `https://koru.js.org/`. If set it wins; the workflow logs it |
| `GOOGLE_SITE_VERIFICATION` | variable | Content of the Search Console `google-site-verification` meta tag                    |
| `BING_SITE_VERIFICATION`   | variable | Content of the Bing Webmaster `msvalidate.01` meta tag                               |
| `YANDEX_VERIFICATION`      | variable | Content of the Yandex Webmaster `yandex-verification` meta tag                       |
| `INDEXNOW_KEY`             | secret   | 8-128 characters `[a-zA-Z0-9-]`; published as `/<key>.txt`                           |

The checkout uses `fetch-depth: 0` so the sitemap's `lastmod` reflects real commit dates. Run the workflow again after changing any of these.

Manual steps that cannot be automated:

1. **Google Search Console**: add a "URL prefix" property for `https://koru.js.org/` (js.org does not let you create DNS TXT records). Choose the HTML tag method, copy the `content` value into `GOOGLE_SITE_VERIFICATION`, rerun the workflow, then verify. Under Sitemaps submit `sitemap.xml`; use URL inspection to request indexing of `/en/` and `/es/` (`/` only redirects to `/en/`; the verification meta tags are emitted on the landings). Check the Pages and International reports over the following weeks.
2. **Bing Webmaster Tools**: import the site from Search Console, or verify with the meta tag (`BING_SITE_VERIFICATION`). Submit `sitemap.xml`. This also covers DuckDuckGo.
3. **IndexNow**: generate a key (for example `openssl rand -hex 16`), store it as the `INDEXNOW_KEY` secret and rerun the workflow. The `notify` job submits the sitemap after each deployment to Bing, Yandex, Naver and Seznam; a failure only warns.
4. **Yandex Webmaster** (optional): verify with `YANDEX_VERIFICATION` and submit the sitemap.
5. Check a shared link's preview (`assets/og-en.png`, `assets/og-es.png`) with a debugger such as opengraph.xyz, and validate structured data with validator.schema.org.
