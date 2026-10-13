# Publish koru documentation

The repository is `wux4an/koru`. The GitHub Pages workflow in
`.github/workflows/pages.yml` builds `website/dist` with Bun 1.2.13, runs content
and formatting checks, and publishes the static artifact. Action versions are
pinned to reviewed release commits. Node/Bun exist only in this documentation
workflow and are unrelated to the Go APK compiler.

## GitHub Pages

Enable **GitHub Actions** as the Pages source in the repository settings. The
workflow runs on pushes to main that change documentation, and can also be run
manually. It takes the canonical URL from `actions/configure-pages`, so the first
publication works at `https://wux4an.github.io/koru/`. After configuring a custom
domain, a subsequent deployment automatically switches the site's base path,
canonical URLs, language links and sitemap to that domain.

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

## Request koru.js.org

The intended domain is `koru.js.org`. At the time of preparation, `koru` was absent
from the active JS.ORG CNAME list. This is a lookup, not a reservation or approval.
The project is a JavaScript application toolchain with a JavaScript bridge/API;
the documentation describes that purpose and includes substantive content.

1. Check that the GitHub Pages site has deployed and contains the documentation.
2. Set `koru.js.org` as the custom domain in the repository's Pages settings.
   For Actions deployments, GitHub uses this setting; a CNAME file in the artifact
   is not sufficient.
3. Submit your request in `js-org/js.org`, adding the following alphabetically to
   `cnames_active.js`:

   ```js
   "koru": "wux4an.github.io",
   ```

4. Follow any requested changes from the JS.ORG reviewers. DNS becomes usable
   after their approval and propagation; neither is controlled by this repository.
5. Rerun the documentation workflow after setting the domain, then enable
   **Enforce HTTPS** when GitHub has provisioned the certificate.

Do not claim the js.org URL is live until DNS and HTTPS have been checked. The
external domain request is left to the repository owner.

Official references:

- [JS.ORG instructions and content requirements](https://github.com/js-org/js.org/blob/master/README.md)
- [JS.ORG contribution guide](https://github.com/js-org/js.org/blob/master/CONTRIBUTING.md)
- [GitHub Pages workflows](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages)
- [GitHub custom domains](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site)
