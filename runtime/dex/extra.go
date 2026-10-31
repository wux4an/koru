package runtimedex

import "github.com/wux4an/koru/pkg/dex"

// Small native plugins handled inline by the dispatcher: browser (open a link
// in another app), screen (keep it awake) and clipboard. The dispatcher holds
// params in v3 and the Activity in p0, and runs on the main thread.

func extraRoutes() []route {
	return []route{
		{"browser", "browser.open", browserOpen},
		{"screen", "screen.keepAwake", keepAwake},
		{"clipboard", "clipboard.writeText", clipboardWrite},
		{"clipboard", "clipboard.readText", clipboardRead},
	}
}

func browserOpen(b *dex.CodeBuilder) {
	ok := b.UniqueLabel("scheme")
	paramString(b, "url", 9)
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, 9)
	object(b, 10)
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", 10)
	object(b, 4)
	branch(b, "if-eqz", "invalid", 4)
	for _, s := range externalSchemes {
		b.ConstString(5, s)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 5, 4)
		scalar(b, 6)
		branch(b, "if-nez", ok, 6)
	}
	b.Emit(dex.Instruction{Op: "goto", Target: "invalid"})
	b.Label(ok)
	fresh(b, 7, intent)
	b.ConstString(5, "android.intent.action.VIEW")
	call(b, "invoke-direct", intent, "<init>", "(Ljava/lang/String;"+uri+")V", 7, 5, 10)
	b.ConstString(5, "android.intent.category.BROWSABLE")
	call(b, "invoke-virtual", intent, "addCategory", "(Ljava/lang/String;)"+intent, 7, 5)
	call(b, "invoke-virtual", activityClass, "startActivity", "("+intent+")V", dex.P(0), 7)
	emptyResult(b)
	success(b)
}

func keepAwake(b *dex.CodeBuilder) {
	off := b.UniqueLabel("awakeOff")
	done := b.UniqueLabel("awakeDone")
	b.ConstString(5, "enabled")
	call(b, "invoke-virtual", jsonObject, "opt", "(Ljava/lang/String;)Ljava/lang/Object;", 3, 5)
	object(b, 9)
	b.Emit(dex.Instruction{Op: "instance-of", Regs: []dex.Register{6, 9}, Type: "Ljava/lang/Boolean;"})
	branch(b, "if-eqz", "invalid", 6)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{9}, Type: "Ljava/lang/Boolean;"})
	call(b, "invoke-virtual", "Ljava/lang/Boolean;", "booleanValue", "()Z", 9)
	scalar(b, 6)
	call(b, "invoke-virtual", activityClass, "getWindow", "()Landroid/view/Window;", dex.P(0))
	object(b, 7)
	b.Const(4, 128) // WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON
	branch(b, "if-eqz", off, 6)
	call(b, "invoke-virtual", "Landroid/view/Window;", "addFlags", "(I)V", 7, 4)
	b.Emit(dex.Instruction{Op: "goto", Target: done})
	b.Label(off)
	call(b, "invoke-virtual", "Landroid/view/Window;", "clearFlags", "(I)V", 7, 4)
	b.Label(done)
	emptyResult(b)
	success(b)
}

const clipboardManager = "Landroid/content/ClipboardManager;"
const clipData = "Landroid/content/ClipData;"

func clipboardService(b *dex.CodeBuilder) {
	b.ConstString(5, "clipboard")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", dex.P(0), 5)
	object(b, 7)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{7}, Type: clipboardManager})
}

func clipboardWrite(b *dex.CodeBuilder) {
	paramString(b, "text", 9)
	clipboardService(b)
	b.ConstString(5, "koru")
	call(b, "invoke-static", clipData, "newPlainText", "(Ljava/lang/CharSequence;Ljava/lang/CharSequence;)"+clipData, 5, 9)
	object(b, 10)
	call(b, "invoke-virtual", clipboardManager, "setPrimaryClip", "("+clipData+")V", 7, 10)
	emptyResult(b)
	success(b)
}

// clipboardRead answers {value} with the clipboard as text, or {} when it is
// empty. Android 10 and later only let the app in the foreground read it.
func clipboardRead(b *dex.CodeBuilder) {
	none := b.UniqueLabel("clipEmpty")
	clipboardService(b)
	call(b, "invoke-virtual", clipboardManager, "getPrimaryClip", "()"+clipData, 7)
	object(b, 10)
	emptyResult(b)
	branch(b, "if-eqz", none, 10)
	call(b, "invoke-virtual", clipData, "getItemCount", "()I", 10)
	scalar(b, 6)
	branch(b, "if-eqz", none, 6)
	b.Const(6, 0)
	call(b, "invoke-virtual", clipData, "getItemAt", "(I)Landroid/content/ClipData$Item;", 10, 6)
	object(b, 10)
	call(b, "invoke-virtual", "Landroid/content/ClipData$Item;", "coerceToText", "("+context+")Ljava/lang/CharSequence;", 10, dex.P(0))
	object(b, 10)
	call(b, "invoke-interface", "Ljava/lang/CharSequence;", "toString", "()Ljava/lang/String;", 10)
	object(b, 10)
	put(b, 8, "value", 10, false)
	b.Label(none)
	success(b)
}
