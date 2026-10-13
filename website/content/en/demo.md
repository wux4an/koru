# Record the koru CLI with VHS

The recording uses [Charm VHS](https://github.com/charmbracelet/vhs), version 0.12.1, and actually executes the koru binary. Command output is not simulated.

It demonstrates init, sample HTML/CSS/JS, a release build with no tools in PATH, v2 verification, cached rebuilding, byte-identical output and doctor. Python in the verification scene only formats JSON for readability; it does not compile or sign.

## Reproduce

Install VHS, ttyd, FFmpeg and Chromium, then compile koru. From the repository root:

```sh
mkdir -p artifacts
KORU_BIN="$PWD/bin/koru" \
KORU_WEB="$PWD/examples/web" \
VHS_BROWSER_PATH=/usr/bin/chromium \
vhs docs/demo/koru.tape
```

The tape creates a fresh temporary directory at `/tmp/koru-vhs-*` and a signing identity through init. Later builds of that app reuse it. Keys never appear on screen or enter the recording/repository. The temporary directory is private test material.

Output: `artifacts/koru-demo.mp4`, 1440×900 at 25 fps. Extract the signature-verification poster after recording:

```sh
ffmpeg -y -ss 30 -i artifacts/koru-demo.mp4 -frames:v 1 artifacts/koru-demo-poster.png
```

The delivered MP4 was remuxed with `-c copy -movflags +faststart` for progressive playback. All frames decoded without errors, and the recording's APKs were verified. Evidence: [validation.json]({{base}}original/docs/demo/validation.json ":ignore"). This records the CLI; Android installation/execution evidence is separate.

VHS and its dependencies are documentation tools, not `koru build` requirements. The recording is not automatically published to external services. It predates the Cobra help presentation; new recordings use the current English CLI.
