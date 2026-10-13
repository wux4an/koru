# ADR 0001: independent, verifiable DEX core

Date: 2026-10-13. Status: adopted for the initial increment.

> This is the initial architectural decision. Its APK and runtime work was completed in later milestones. The file-based Activity described here remains an experiment; the normal runtime uses intercepted HTTPS. See [ADR 0002]({{base}}en/architecture-apk/).

The goal requires real DEX emission without Java tools or Android packagers. The implementation uses Go and the standard library exclusively. Output is DEX 035: invoke-custom, default methods and container versions are unnecessary. The experimental runtime targets Android API 26 or later; this initial increment did not yet include APK configuration enforcing that minimum.

## Separate layers

Symbolic model, encoders, assembler, builder, parser and runtime are separate. The public API lives in `pkg/dex` so outside consumers can use the generator without importing internal packages. Public aliases avoid incompatible representations of the same IR. The CLI accepts JSON in the same model, enabling builds without compiling a user Go program.

Compilation has two logical passes: collect and sort symbols, then resolve instructions and emit with offset fixups. The writer is append-only with fixups for reserved fields; already positioned section lengths do not change. Symbolic parameter registers allow additional locals without manually rewriting physical argument indices.

## Deterministic assembly

`goto` always uses the /32 form, avoiding a branch-relaxation algorithm in the first increment. Conditional branches retain official 16-bit limits and fail explicitly if they do not fit. No silent trampolines. `const-string` chooses jumbo when required by the index.

Each exact type_list is shared. Classes sort by dependency and members by identifiers. Identical inputs produce identical bytes; reordering equivalent declarations also preserves output. Dates, host paths and invented source filenames are not included.

## Independent parser

The parser does not import writer, builder or assembler. It physically enumerates sections through map_list and checks references against recognized item starts. It shares only small primitives and naming rules with the emitter; those primitives have independent golden tests. Unsupported sections are rejected. This makes the tool's claims explicit; it does not replace ART's resolution and type analysis.

## Initial Activity experiment

The Activity is emitted only by `--activity-experiment`. It uses file assets for the initial test permitted by the project specification and is not presented as production. There is no implicit AndroidX or addJavascriptInterface bridge. The next runtime requires a generated WebViewClient with local HTTPS interception and an origin/navigation policy.

C, JNI, plugins and empty signature/APK modules were not added merely to simulate progress. The remaining acceptance requirement was actual Android execution and unimplemented capabilities, not a disguised external dependency.

## Specification references

- [DEX format](https://source.android.com/docs/core/runtime/dex-format)
- [Dalvik bytecode](https://source.android.com/docs/core/runtime/dalvik-bytecode)
- [Instruction formats](https://source.android.com/docs/core/runtime/instruction-formats)
- [Activity](https://developer.android.com/reference/android/app/Activity)
- [WebView](https://developer.android.com/reference/android/webkit/WebView)
- [WebSettings file access](<https://developer.android.com/reference/android/webkit/WebSettings#setAllowFileAccess(boolean)>)

APK signing would be designed from the already consulted [v2](https://source.android.com/docs/security/features/apksigning/v2) and [v3](https://source.android.com/docs/security/features/apksigning/v3) specifications. The implemented current scheme is v2.
