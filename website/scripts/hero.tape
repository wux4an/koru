# Run from the repository root with KORU_BIN pointing to the compiled CLI.
# VHS, ttyd, FFmpeg and Chrome are documentation tools, never APK build dependencies.
Output website/assets/hero-demo.gif

Set Shell bash
Set Width 1280
Set Height 480
Set FontSize 19
Set FontFamily "DejaVu Sans Mono"
Set Padding 30
Set WindowBar Rings
Set Framerate 20
Set TypingSpeed 28ms
Set Theme { "name": "Koru clay", "background": "#171f32", "foreground": "#e2e8f5", "black": "#171f32", "red": "#f2b9a5", "green": "#bce4cc", "yellow": "#f0cf94", "blue": "#9fcaff", "magenta": "#c3b5f5", "cyan": "#a7d9dd", "white": "#e2e8f5", "brightBlack": "#8191ac", "brightRed": "#ffd0bf", "brightGreen": "#d4f5dd", "brightYellow": "#ffe2b3", "brightBlue": "#bbdeff", "brightMagenta": "#ded1ff", "brightCyan": "#c6eef0", "brightWhite": "#ffffff", "cursor": "#bce4cc", "selection": "#3a4768" }

Hide
Type `export PS1='koru $ ' TERM=xterm-256color COLORTERM=truecolor; unset PROMPT_COMMAND NO_COLOR; cd "$(mktemp -d /tmp/koru-hero-XXXXXX)"; cp "$KORU_BIN" ./koru; clear`
Enter
Sleep 0.6
Show
Type "./koru init my-app"
Enter
Wait+Screen /Initialized/
Sleep 1.2
Type "./koru build my-app --release --output app.apk"
Enter
Wait+Screen /APK v2 RSA/
Sleep 1.5
Type "./koru verify app.apk | python3 -m json.tool"
Enter
Wait+Screen /signatureScheme/
Sleep 2.5
