# Capture real Koru CLI commands. No prerecorded or fabricated command output.
Output artifacts/koru-demo.mp4

Set Shell bash
Set Width 1440
Set Height 900
Set FontSize 22
Set FontFamily "DejaVu Sans Mono"
Set Padding 40
Set WindowBar Rings
Set Framerate 25
Set TypingSpeed 32ms
Set Theme { "name": "Koru", "background": "#10201c", "foreground": "#e8f4ed", "black": "#10201c", "red": "#ff8f8f", "green": "#98efc4", "yellow": "#f5dda3", "blue": "#90c7f5", "magenta": "#c6afff", "cyan": "#78ded3", "white": "#e8f4ed", "brightBlack": "#70887d", "brightRed": "#ffaaaa", "brightGreen": "#b5ffdb", "brightYellow": "#ffe8b4", "brightBlue": "#acd6ff", "brightMagenta": "#dec8ff", "brightCyan": "#aff8ef", "brightWhite": "#ffffff", "cursor": "#98efc4", "selection": "#29483a" }

Hide
Type `export PS1='koru $ '; unset PROMPT_COMMAND; cd "$(mktemp -d /tmp/koru-vhs-XXXXXX)"; cp "$KORU_BIN" ./koru; cp -R "$KORU_WEB" web-demo; clear`
Enter
Sleep 1
Show
Type "# KORU: de HTML, CSS y JavaScript a un APK firmado"
Enter
Sleep 2
Type "./koru version"
Enter
Wait+Screen /koru 0.2.0-dev/
Sleep 2
Type "./koru init mi-app"
Enter
Wait+Screen /persistent signing identity/
Sleep 2
Type "cp web-demo/* mi-app/www/"
Enter
Type "ls mi-app/www"
Enter
Sleep 3

Hide
Type "clear"
Enter
Sleep 0.3
Show
Type "# Build sin Java, Gradle ni herramientas Android"
Enter
Sleep 2
Type "env -i PATH=/sin-herramientas ./koru build mi-app --release --output app.apk"
Enter
Wait+Screen /APK v2 RSA/
Sleep 4
Type "./koru verify app.apk | python3 -m json.tool"
Enter
Wait+Screen /signatureScheme/
Sleep 5

Hide
Type "clear"
Enter
Sleep 0.3
Show
Type "# Misma clave persistente. Mismo APK, byte a byte."
Enter
Sleep 2
Type "./koru build mi-app --release --output copia.apk"
Enter
Wait+Screen /runtime cache hit: true/
Sleep 2
Type "cmp app.apk copia.apk && echo 'APK identico: firma y contenido conservados'"
Enter
Wait+Screen /APK identico: firma y contenido conservados/
Sleep 4
Type "./koru doctor mi-app"
Enter
Wait+Screen /valid config/
Sleep 4
