# Milestones and scope

| Milestone             | Implementation and tests                                                                                    | Result / limit                                                                         |
| --------------------- | ----------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| 1 Minimal DEX         | Format 035, IDs, map, LEB, MUTF, hashes, parser, executable method.                                         | Own/AOSP checks pass; classes loaded and executed in ART.                              |
| 2 Advanced DEX        | References, constructors, parameters, fields, instructions, arrays, interfaces, conversions and exceptions. | Activity, bridge and plugin bytecode executed on Android; not a full Java compiler.    |
| 3 Activity            | Go-generated Activity/WebView and lifecycle.                                                                | Startup confirmed with am start, logcat and DevTools.                                  |
| 4 APK                 | AXML, ARSC/icon, aligned ZIP32, persistent identity, v2 signature and own verifier.                         | Install/update under the same identity; external aapt/apksigner accept output.         |
| 5 WebView             | Controlled HTTPS origin, assets, CSP, SPA, service-worker interceptor.                                      | HTML/JS/CSS/SVG without Wi-Fi or data; navigation, Back and keyboard checked.          |
| 6 Runtime             | MessagePort, validation/promises/errors, six plugins and DEX ABI, optional JNI.                             | Device/app/storage/vibration/notifications/chooser tested; JNI info executed in ART.   |
| 7 Toolchain           | English Cobra CLI: init/build/release/dex/verify/doctor/version/keygen; cache, examples, docs and binary.   | APK build stays offline; Linux validated, other hosts limited to compilation checks.   |
| Documentation website | English/Spanish Docsify site, local assets, prerendered HTML and browser checks.                            | Website tooling is separate from the Android compiler; see website validation records. |

Cobra is confined to `cmd/koru`. DEX, AXML, ARSC, APK packaging and signatures remain implemented by koru using Go and the standard library. Help derives from the command tree and uses ANSI colors in interactive terminals.

## Executed matrix

Android 11, API 30, x86_64, software emulation without KVM. This demonstrates the complete APK, not only a well-formed ZIP/DEX. It does not prove compatibility with all devices or real-hardware performance.

Initial failure logs and diagnoses are retained. Native-port reception and Back navigation tests were fixed: WebView can replace its DevTools target when restoring a page. The test checks URL/Activity and reconnects to the restored target.

## Outside the initial scope

APK v3/v4, AAB, scheduled alarms, Android 26–36/ARM device matrix, comparative benchmarks, general annotations, multidex and full semantic register verification. The parser remains strict for the implemented subset. The JNI prebuilt example is x86_64; standard DEX apps do not depend on that architecture.

[Android evidence]({{base}}docs/en/android-results/) · [Documentation archive]({{base}}docs/en/sources/)
