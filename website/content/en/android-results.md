# Android execution evidence

Results recorded on 2026-10-13 for koru `0.2.0-dev`. These are real outputs, not fixtures simulating Android execution. Official tools are used only in QA, outside the normal build.

## Compiler and toolchain

The original full-runtime snapshot recorded:

- 114 Go cases/subcases passing in 16 packages: [go-tests.jsonl]({{base}}original/docs/results/full/go-tests.jsonl ":ignore").
- 74.2% overall coverage, including the then-new CLI and validators: [coverage.txt]({{base}}original/docs/results/full/coverage.txt ":ignore").
- String-pool decoded cost bounded to 64 MiB, accounting for each reference and UTF-16 units, preventing amplification through shared offsets. Adversarial tests and byte-identical APK output are recorded.
- `go vet` and `go test -race` without errors: [vet.txt]({{base}}original/docs/results/full/vet.txt ":ignore") and [race.txt]({{base}}original/docs/results/full/race.txt ":ignore").
- Two Node protocol tests: [js-tests.txt]({{base}}original/docs/results/full/js-tests.txt ":ignore").
- JNI C reference and attach/detach ownership tests: [native-unit.txt]({{base}}original/docs/results/full/native-unit.txt ":ignore").
- AXML/ARSC/signature fuzz campaigns without failures. The initial XML campaign with only 12 executions is also preserved. These short campaigns are not exhaustive exploration.
- Official dexdump, aapt and apksigner acceptance: [official-tests.txt]({{base}}original/docs/results/full/official-tests.txt ":ignore").
- Static Linux binary with CGO=0, trimpath and buildvcs=false, reproducible from a delivery without Git metadata. Builds with empty PATH produced byte-identical APKs: [offline-isolated.txt]({{base}}original/docs/results/full/offline-isolated.txt ":ignore"). The app build downloads no tools and launches no subprocesses.
- Cross-compilation without execution on other hosts: [host-portability.txt]({{base}}original/docs/results/full/host-portability.txt ":ignore").

These numbers describe the recorded snapshot. Later CLI changes have their own validation records; do not treat historical coverage as a measurement of the current checkout.

## Actual Android execution

Official emulator, Android 11/API 30, x86_64, default image r11, software execution without KVM. APKs generated in Go were installed, launched and updated under the same identity. Sample certificate:

```text
bfee8015a6fbf4e44d5ac5f01d59470a08e33ea3534b5da1e3cbe82da6f719f5
```

The private key is not distributed.

| Record                                                                                                                                                                                              | What was exercised                                                                                                                                                                                         |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [android-smoke.json]({{base}}original/docs/results/full/android-smoke.json ":ignore")                                                                                                               | Local HTTPS origin, HTML/JS/CSS/SVG, ES modules, structured errors, device/app/storage/vibration/notification and native.info JNI on main. Wi-Fi and mobile data disabled, no INTERNET permission.         |
| [android-back.json]({{base}}original/docs/results/full/android-back.json ":ignore")                                                                                                                 | Navigation to second.html, Back to index.html, storage after restart. The test reconnects when WebView replaces the DevTools target after history restoration. Earlier back-diagnosis failures were fixed. |
| [android-worker.json]({{base}}original/docs/results/full/android-worker.json ":ignore")                                                                                                             | Registered service worker and offline local response.                                                                                                                                                      |
| [android-keyboard.json]({{base}}original/docs/results/full/android-keyboard.json ":ignore")                                                                                                         | Real KoruKeyboard input, visible IME and dismissal with Back; reconnects after inspector replacement on resize.                                                                                            |
| [android-share.json]({{base}}original/docs/results/full/android-share.json ":ignore")                                                                                                               | Native chooser shown and cancelled, without sending to recipients.                                                                                                                                         |
| [external-navigation-result.json]({{base}}original/docs/results/full/external-navigation-result.json ":ignore")                                                                                     | External-origin navigation blocked.                                                                                                                                                                        |
| [demo-install.txt]({{base}}original/docs/results/full/demo-install.txt ":ignore") / [demo-launch.txt]({{base}}original/docs/results/full/demo-launch.txt ":ignore")                                 | Portable release APK installed and Activity launched.                                                                                                                                                      |
| [storage-after-restart.json]({{base}}original/docs/results/full/storage-after-restart.json ":ignore")                                                                                               | Value preserved after force-stop and fresh launch.                                                                                                                                                         |
| [update-install.txt]({{base}}original/docs/results/full/update-install.txt ":ignore") / [update-preserved-storage.json]({{base}}original/docs/results/full/update-preserved-storage.json ":ignore") | versionCode 3 installed with the same certificate and data preserved.                                                                                                                                      |

Final Wi-Fi and mobile-data states were both 0. The final screenshot shows CSS, SVG, JavaScript and “Contador: 1” after a real tap:

![The release app running in the API 30 Android emulator]({{base}}original/docs/results/full/demo-release-screen.png)

Initial logs include fixed development failures and emulator System UI startup dialogs. First/controlled/all screenshots are historical; demo-release-screen corresponds to the final delivery.

## Reproducible measurements

[benchmarks.txt]({{base}}original/docs/results/full/benchmarks.txt ":ignore"), Linux amd64, AMD EPYC 9V74. One recorded run:

| Operation                     |       Time |   Bytes/op | Allocations/op |
| ----------------------------- | ---------: | ---------: | -------------: |
| Minimal DEX                   |   6.414 µs |      3,926 |             70 |
| Minimal experimental Activity | 239.823 µs |     98,662 |          1,422 |
| Cached APK build              |  11.602 ms | 13,266,993 |          6,266 |

The Activity benchmark measures the original experiment, not the entire plugin runtime. The portable release APK in this sample is 22,852 bytes. The x86_64 JNI debug APK has another size and is not an equivalent measurement. bridgeMillis is a single software-emulation latency, not a performance comparison.

## Limits of the evidence

Executed matrix: API 30/x86_64 only. ARM hardware, other APIs/manufacturers and the API 33+ POST_NOTIFICATIONS dialog remain untested. The portable APK contains no native libraries and imposes no ABI, but that does not mean it has run on every architecture. There is no Capacitor comparison.

Unimplemented: APK v3/v4, AAB, scheduled alarms, multidex, general DEX annotations and complete per-register type analysis. The six built-in plugins use framework/DEX. JNI is optional and its included prebuilt is x86_64 only.
