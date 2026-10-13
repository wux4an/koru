# Early DEX results

> Historical DEX-only snapshot, recorded 2026-10-13 before APK/runtime completion. Its “no Android execution yet” statements refer to that stage. Current APK evidence is on the [Android execution page]({{base}}en/android-results/).

All recorded tests ran against real workspace code. At this stage there were no Android installation/execution tests or APK results. `summary.json` records cases, coverage and artifact hashes.

## Validation

| Check                                                 | Result and evidence                                                                                                          |
| ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Go suite                                              | Passed; unit-tests.txt, tests.jsonl.                                                                                         |
| Combined package coverage                             | 84.0%; coverage.out, coverage-summary.txt.                                                                                   |
| `go vet ./...`                                        | Exit 0; empty vet.txt means success.                                                                                         |
| `go test -race ./...`                                 | Passed; race.txt.                                                                                                            |
| ULEB128/SLEB128 fuzz, 10 s, two workers               | 552,944 executions without failure; fuzz-leb.txt.                                                                            |
| MUTF-8 fuzz, 10 s, two workers                        | 271,686 executions without failure; fuzz-mutf8.txt.                                                                          |
| Parser fuzz, 15 s, two workers, Answer/Activity seeds | 256,070 executions without failure; fuzz-parser.txt.                                                                         |
| Deliberate Adler-32 validator mutation                | Test detected mutation and failed; code restored before the final suite; mutation-check.txt.                                 |
| Official AOSP dexdump                                 | Answer/Activity checksum, open and disassembly passed; external-tests.txt, \*-dexdump.txt.                                   |
| Build without module proxy or CGO                     | Passed with GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 for the original standard-library-only snapshot.         |
| CLI with no tools in PATH                             | Generation/inspection passed with PATH=/nonexistent; offline-cli.txt.                                                        |
| Delivery ZIP                                          | Correct CRC; clean extraction rebuilt without module proxy; binary and Answer DEX byte-identical (cmp exit 0).               |
| Dependencies                                          | Original snapshot contained only koru and standard library; build-dependencies.txt. The current CLI additionally uses Cobra. |
| Windows amd64/macOS arm64                             | Cross-compilation only; cross-builds.txt; not executed on those hosts.                                                       |

Combined coverage uses `go test -coverpkg=./... -coverprofile=... ./...` and instruments inter-package calls. It measures test reach, not absence of defects.

## DEX measurements

Linux amd64; Go-reported CPU: AMD EPYC 9V74. Benchmarks use `-benchmem`, with IR prepared outside the measured region.

| Case                                               | Mean time    | Allocations | Allocated bytes/op | DEX size                   |
| -------------------------------------------------- | ------------ | ----------- | ------------------ | -------------------------- |
| One class with static answer()I                    | 6176 ns/op   | 70          | 3926 B/op          | 396 bytes in this fixture. |
| Public Activity build, including parser validation | 165560 ns/op | 1420        | 98460 B/op         | 1848 bytes.                |

`artifacts/answer.dex` also includes a constructor and is 476 bytes, a different fixture. The Linux binary uses trimpath, no stripping and no CGO; exact size and hash are in summary.json. Allocated bytes are not peak memory.

APK builds, installation, startup time, bridge performance and Capacitor comparisons were not measured at this stage. No performance superiority is claimed.

## Reproduction and provenance

The recorded session used `/workspace/tooling/go/bin` and `/workspace/tooling/android-validation/dexdump`. Those are environment-specific paths. The comparison tool is not distributed with koru or required by app builds. Tool origin/hashes are in tooling.json; consulted specifications in docs/specs/sources.json.

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 go build -trimpath -o bin/koru ./cmd/koru
PATH=/nonexistent ./bin/koru dex build examples/dex/answer.json --output answer.dex
PATH=/nonexistent ./bin/koru dex build --activity-experiment --output activity.dex
PATH=/nonexistent ./bin/koru dex inspect activity.dex
```

Rebuilding today's Cobra CLI without a proxy requires its modules to be cached first. The compiled binary's app build remains offline. The normal compiler does not launch external processes. dexdump validates structure/checksum and shows code; it does not execute methods in ART or validate Android rendering.

[Browse original records]({{base}}en/sources/)
