# Documentation for people building apps with koru

## Audience and editorial direction

Readers want to install koru, package their web project, use Android features and
share an APK. The README answers what koru does, how to start and where to find
help. The website guides those tasks with prerequisites, executable commands,
expected outcomes and next steps. Internal workspace paths and development-tool
notes belong in engineering documentation.

The information architecture separates first steps, app guides, reference and
“Inside koru”. DEX, JNI and runtime explanations live in the last section.
Historical test reports remain reachable at their existing URLs and through the
project archive; they do not occupy a main navigation group. Compatibility limits
remain visible where they affect installation or distribution decisions.

## Research

Reviewed on 2026-10-14:

- [Expo docs](https://docs.expo.dev/): task-oriented entry points and a clear path
  from starting a project to developing and distributing it.
- [Bun docs](https://bun.com/docs): short introductions with useful commands near
  the start, and a compact topic navigation.
- [Deno docs](https://docs.deno.com/runtime/): separate getting-started material,
  guides, examples and reference.
- [Parcel docs](https://parceljs.org/docs/): familiar task names and direct links
  into setup and feature guides.
- [GitHub README guidance](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes):
  introduce the project, explain getting started and point to help.
- [Diátaxis](https://www.diataxis.fr/start-here/): tutorials, task guides,
  reference and explanations serve different reader needs.

These are structural references. Koru's examples come from its CLI and runtime;
copy is written for this product rather than adapted from another tool's promises.

## Design plan

Six core colors: paper `#FFFFFF`, wash `#F7F7FA`, ink `#292B36`, secondary text
`#646674`, lavender `#6851A5`, and lavender wash `#EFEBF8`. Borders use `#E6E5ED`.
Dark mode uses neutral violet surfaces, pale text and light lavender links.

Nunito Variable retains koru's soft character at a quieter weight and scale.
JetBrains Mono is reserved for code. Fonts and icons stay local. No new package
is needed for this design.

```text
logo / documentation                  language / appearance / archive
navigation          introduction and first action
search              two commands to create an APK
first steps         task links with short descriptions
app guides          how the app fits an existing web project
reference           compatibility and next steps
inside koru
```

All reading content is left aligned, with a maximum article width of 880 px
including padding. Mobile uses the existing navigation drawer and one-column task
links. Rounded controls, lavender surfaces and generous line spacing provide
softness. Borders separate content only where they help scanning.

## Review before implementation

The first rework removed the browser-to-APK illustration to bring the first
commands forward. Reader feedback made clear that the playful illustration was a
favorite part of koru's character. Keep it beside the headline, where it explains
the product at a glance. The two-command start follows immediately below. Raised
clay surfaces remain limited to this one illustration; the rest of the site keeps
the quiet reading layout. The terminal recording stays behind a reader-opened
disclosure.

## Implementation and review

The CSS is replaced with a token-based light/dark system rather than another layer
of overrides. Existing Docsify routing, copy, search, language switching and theme
keyboard behavior remain. The search namespace changes with this content revision
to avoid showing stale indexed text. The long-page contents index works with and
without JavaScript. All original records remain in the generated archive.

Review uses the installed `google-chrome-stable` through Playwright: desktop and
mobile screenshots, all routes, theme controls, links, code copying, search,
keyboard access, reduced motion and axe checks. Results are recorded separately
in [VALIDATION.md](VALIDATION.md).
