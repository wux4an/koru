package runtimedex

import (
	"github.com/wux4an/koru/internal/config"
	"github.com/wux4an/koru/pkg/dex"
)

// The WebChromeClient lets pages use web APIs that a WebView blocks until the
// app handles them (webFeatures in koru.json): <input type=file>,
// navigator.geolocation and getUserMedia. Geolocation, camera and microphone
// ask Android for the permission and only answer the app origin and the
// server.bridge origins; every other page is denied.

const chromeClient = "Lcom/koru/runtime/ChromeClient;"
const valueCallback = "Landroid/webkit/ValueCallback;"
const fileParams = "Landroid/webkit/WebChromeClient$FileChooserParams;"
const geoCallback = "Landroid/webkit/GeolocationPermissions$Callback;"
const permissionRequest = "Landroid/webkit/PermissionRequest;"

// Request codes for startActivityForResult and requestPermissions. 100 is
// the notification permission.
const (
	fileRequest  = 4243
	geoRequest   = 4244
	mediaRequest = 4245
)

func usesChrome(c config.Config) bool {
	return c.HasWebFeature("files") || c.HasWebFeature("geolocation") || usesMedia(c)
}
func usesMedia(c config.Config) bool {
	return c.HasWebFeature("camera") || c.HasWebFeature("microphone")
}

// chromeFields are the Activity fields that hold a pending web request while
// Android shows its picker or permission dialog.
func chromeFields(c config.Config) []dex.Field {
	var f []dex.Field
	if c.HasWebFeature("files") {
		f = append(f, dex.Field{Name: "fileCallback", Type: valueCallback})
	}
	if c.HasWebFeature("geolocation") {
		f = append(f, dex.Field{Name: "geoCallback", Type: geoCallback}, dex.Field{Name: "geoOrigin", Type: "Ljava/lang/String;"})
	}
	if usesMedia(c) {
		f = append(f, dex.Field{Name: "mediaRequest", Type: permissionRequest})
	}
	return f
}

func chromeClass(c config.Config) dex.Class {
	cl := dex.Class{Descriptor: chromeClient, Super: "Landroid/webkit/WebChromeClient;", Access: dex.Public, Fields: []dex.Field{{Name: "activity", Type: ActivityDescriptor, Access: dex.Private}}}
	b := dex.NewCode(0)
	call(b, "invoke-direct", cl.Super, "<init>", "()V", dex.P(0))
	field(b, "iput", chromeClient, "activity", ActivityDescriptor, dex.P(1), dex.P(0))
	b.ReturnVoid()
	cl.Methods = append(cl.Methods, dex.Method{Name: "<init>", Descriptor: "(" + ActivityDescriptor + ")V", Access: dex.Public | dex.Constructor, Code: b.Build()})
	cl.Methods = append(cl.Methods, trustedMethod(c))
	if c.HasWebFeature("files") {
		cl.Methods = append(cl.Methods, fileChooserMethod())
	}
	if c.HasWebFeature("geolocation") {
		cl.Methods = append(cl.Methods, geolocationPromptMethod())
	}
	if usesMedia(c) {
		cl.Methods = append(cl.Methods, mediaRequestMethod(c))
	}
	return cl
}

// trustedMethod emits ChromeClient.trusted(Uri): whether the URI's origin is
// the app origin or a server.bridge origin.
func trustedMethod(c config.Config) dex.Method {
	b := dex.NewCode(4)
	branch(b, "if-eqz", "no", dex.P(0))
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", dex.P(0))
	object(b, 0)
	branch(b, "if-eqz", "no", 0)
	call(b, "invoke-virtual", uri, "getAuthority", "()Ljava/lang/String;", dex.P(0))
	object(b, 1)
	branch(b, "if-eqz", "no", 1)
	const builder = "Ljava/lang/StringBuilder;"
	fresh(b, 2, builder)
	call(b, "invoke-direct", builder, "<init>", "()V", 2)
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 2, 0)
	b.ConstString(3, "://")
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 2, 3)
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 2, 1)
	call(b, "invoke-virtual", builder, "toString", "()Ljava/lang/String;", 2)
	object(b, 0)
	for _, o := range append([]string{Origin}, c.Server.BridgeOrigins()...) {
		b.ConstString(1, o)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
		scalar(b, 2)
		branch(b, "if-nez", "yes", 2)
	}
	b.Label("no").Const(0, 0)
	retScalar(b, 0)
	b.Label("yes").Const(0, 1)
	retScalar(b, 0)
	return dex.Method{Name: "trusted", Descriptor: "(" + uri + ")Z", Access: dex.Public | dex.Static, Code: b.Build()}
}

// fileChooserMethod emits onShowFileChooser: open Android's picker for the
// <input type=file>, with multiple selection when the input asks for it.
func fileChooserMethod() dex.Method {
	b := dex.NewCode(4)
	field(b, "iget", chromeClient, "activity", ActivityDescriptor, 0, dex.P(0))
	// A picker still open from an earlier input is answered with "nothing".
	field(b, "iget", ActivityDescriptor, "fileCallback", valueCallback, 1, 0)
	branch(b, "if-eqz", "fresh", 1)
	b.Const(2, 0)
	call(b, "invoke-interface", valueCallback, "onReceiveValue", "(Ljava/lang/Object;)V", 1, 2)
	b.Label("fresh")
	field(b, "iput", ActivityDescriptor, "fileCallback", valueCallback, dex.P(2), 0)
	b.Label("start")
	call(b, "invoke-virtual", fileParams, "createIntent", "()"+intent, dex.P(3))
	object(b, 1)
	call(b, "invoke-virtual", fileParams, "getMode", "()I", dex.P(3))
	scalar(b, 2)
	b.Const(3, 1) // FileChooserParams.MODE_OPEN_MULTIPLE
	branch(b, "if-ne", "single", 2, 3)
	b.ConstString(2, "android.intent.extra.ALLOW_MULTIPLE")
	call(b, "invoke-virtual", intent, "putExtra", "(Ljava/lang/String;Z)"+intent, 1, 2, 3)
	b.Label("single")
	b.Const(2, fileRequest)
	call(b, "invoke-virtual", activityClass, "startActivityForResult", "("+intent+"I)V", 0, 1, 2)
	b.Label("end")
	b.Const(1, 1)
	retScalar(b, 1)
	// No picker app: answer the page with "nothing" so the input works again.
	b.Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{1}})
	b.Const(1, 0)
	field(b, "iput", ActivityDescriptor, "fileCallback", valueCallback, 1, 0)
	call(b, "invoke-interface", valueCallback, "onReceiveValue", "(Ljava/lang/Object;)V", dex.P(2), 1)
	b.Const(1, 1)
	retScalar(b, 1)
	b.Catch("start", "end", "catch", "Ljava/lang/Exception;")
	return dex.Method{Name: "onShowFileChooser", Descriptor: "(" + webview + valueCallback + fileParams + ")Z", Access: dex.Public, Code: b.Build()}
}

// activityResultMethod emits MainActivity.onActivityResult: hand the picked
// files (one, several through ClipData, or none) back to the page.
func activityResultMethod() dex.Method {
	b := dex.NewCode(6)
	call(b, "invoke-super", activityClass, "onActivityResult", "(II"+intent+")V", dex.P(0), dex.P(1), dex.P(2), dex.P(3))
	b.Const(0, fileRequest)
	branch(b, "if-ne", "end", dex.P(1), 0)
	field(b, "iget", ActivityDescriptor, "fileCallback", valueCallback, 1, dex.P(0))
	branch(b, "if-eqz", "end", 1)
	b.Const(0, 0)
	field(b, "iput", ActivityDescriptor, "fileCallback", valueCallback, 0, dex.P(0))
	b.Const(2, 0)
	b.Const(0, -1) // Activity.RESULT_OK
	branch(b, "if-ne", "deliver", dex.P(2), 0)
	branch(b, "if-eqz", "deliver", dex.P(3))
	const clip = "Landroid/content/ClipData;"
	call(b, "invoke-virtual", intent, "getClipData", "()"+clip, dex.P(3))
	object(b, 3)
	branch(b, "if-eqz", "single", 3)
	call(b, "invoke-virtual", clip, "getItemCount", "()I", 3)
	scalar(b, 4)
	b.Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{2, 4}, Type: "[" + uri})
	b.Const(5, 0)
	b.Label("loop")
	branch(b, "if-ge", "deliver", 5, 4)
	call(b, "invoke-virtual", clip, "getItemAt", "(I)Landroid/content/ClipData$Item;", 3, 5)
	object(b, 0)
	call(b, "invoke-virtual", "Landroid/content/ClipData$Item;", "getUri", "()"+uri, 0)
	object(b, 0)
	b.Emit(dex.Instruction{Op: "aput-object", Regs: []dex.Register{0, 2, 5}})
	b.Emit(dex.Instruction{Op: "add-int/lit8", Regs: []dex.Register{5, 5}, Literal: 1})
	b.Emit(dex.Instruction{Op: "goto", Target: "loop"})
	b.Label("single")
	call(b, "invoke-static", fileParams, "parseResult", "(I"+intent+")["+uri, dex.P(2), dex.P(3))
	object(b, 2)
	b.Label("deliver")
	call(b, "invoke-interface", valueCallback, "onReceiveValue", "(Ljava/lang/Object;)V", 1, 2)
	b.Label("end").ReturnVoid()
	return dex.Method{Name: "onActivityResult", Descriptor: "(II" + intent + ")V", Access: dex.Protected, Code: b.Build()}
}

// requestPermissions asks Android for perms with the given request code;
// already granted permissions come back granted without a dialog.
func requestPermissions(b *dex.CodeBuilder, activity dex.Register, perms []string, code int32, arr, idx, tmp dex.Register) {
	b.Const(idx, int32(len(perms)))
	b.Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{arr, idx}, Type: "[Ljava/lang/String;"})
	for n, p := range perms {
		b.Const(idx, int32(n)).ConstString(tmp, p)
		b.Emit(dex.Instruction{Op: "aput-object", Regs: []dex.Register{tmp, arr, idx}})
	}
	b.Const(idx, code)
	call(b, "invoke-virtual", activityClass, "requestPermissions", "([Ljava/lang/String;I)V", activity, arr, idx)
}

// geolocationPromptMethod emits onGeolocationPermissionsShowPrompt.
func geolocationPromptMethod() dex.Method {
	b := dex.NewCode(5)
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, dex.P(1))
	object(b, 0)
	call(b, "invoke-static", chromeClient, "trusted", "("+uri+")Z", 0)
	scalar(b, 1)
	branch(b, "if-eqz", "deny", 1)
	field(b, "iget", chromeClient, "activity", ActivityDescriptor, 0, dex.P(0))
	field(b, "iget", ActivityDescriptor, "geoCallback", geoCallback, 1, 0)
	branch(b, "if-eqz", "fresh", 1)
	field(b, "iget", ActivityDescriptor, "geoOrigin", "Ljava/lang/String;", 2, 0)
	b.Const(3, 0)
	call(b, "invoke-interface", geoCallback, "invoke", "(Ljava/lang/String;ZZ)V", 1, 2, 3, 3)
	b.Label("fresh")
	field(b, "iput", ActivityDescriptor, "geoCallback", geoCallback, dex.P(2), 0)
	field(b, "iput", ActivityDescriptor, "geoOrigin", "Ljava/lang/String;", dex.P(1), 0)
	requestPermissions(b, 0, []string{"android.permission.ACCESS_FINE_LOCATION", "android.permission.ACCESS_COARSE_LOCATION"}, geoRequest, 2, 3, 4)
	b.ReturnVoid()
	b.Label("deny").Const(1, 0)
	call(b, "invoke-interface", geoCallback, "invoke", "(Ljava/lang/String;ZZ)V", dex.P(2), dex.P(1), 1, 1)
	b.ReturnVoid()
	return dex.Method{Name: "onGeolocationPermissionsShowPrompt", Descriptor: "(Ljava/lang/String;" + geoCallback + ")V", Access: dex.Public, Code: b.Build()}
}

// mediaRequestMethod emits onPermissionRequest for getUserMedia: only camera
// and microphone, only the features turned on, only trusted origins.
func mediaRequestMethod(c config.Config) dex.Method {
	camera, mic := c.HasWebFeature("camera"), c.HasWebFeature("microphone")
	b := dex.NewCode(8)
	call(b, "invoke-virtual", permissionRequest, "getOrigin", "()"+uri, dex.P(1))
	object(b, 0)
	call(b, "invoke-static", chromeClient, "trusted", "("+uri+")Z", 0)
	scalar(b, 1)
	branch(b, "if-eqz", "deny", 1)
	call(b, "invoke-virtual", permissionRequest, "getResources", "()[Ljava/lang/String;", dex.P(1))
	object(b, 2)
	branch(b, "if-eqz", "deny", 2)
	b.Emit(dex.Instruction{Op: "array-length", Regs: []dex.Register{3, 2}})
	branch(b, "if-eqz", "deny", 3)
	b.Const(4, 0).Const(5, 0).Const(6, 0) // index, wants camera, wants microphone
	b.Label("loop")
	branch(b, "if-ge", "ask", 4, 3)
	b.Emit(dex.Instruction{Op: "aget-object", Regs: []dex.Register{0, 2, 4}})
	b.ConstString(1, "android.webkit.resource.VIDEO_CAPTURE")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
	scalar(b, 7)
	branch(b, "if-eqz", "notCamera", 7)
	if !camera {
		b.Emit(dex.Instruction{Op: "goto", Target: "deny"})
	} else {
		b.Const(5, 1)
		b.Emit(dex.Instruction{Op: "goto", Target: "next"})
	}
	b.Label("notCamera")
	b.ConstString(1, "android.webkit.resource.AUDIO_CAPTURE")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
	scalar(b, 7)
	branch(b, "if-eqz", "deny", 7) // MIDI, protected media and others are refused
	if !mic {
		b.Emit(dex.Instruction{Op: "goto", Target: "deny"})
	} else {
		b.Const(6, 1)
	}
	b.Label("next")
	b.Emit(dex.Instruction{Op: "add-int/lit8", Regs: []dex.Register{4, 4}, Literal: 1})
	b.Emit(dex.Instruction{Op: "goto", Target: "loop"})
	b.Label("ask")
	field(b, "iget", chromeClient, "activity", ActivityDescriptor, 0, dex.P(0))
	field(b, "iget", ActivityDescriptor, "mediaRequest", permissionRequest, 1, 0)
	branch(b, "if-eqz", "fresh", 1)
	call(b, "invoke-virtual", permissionRequest, "deny", "()V", 1)
	b.Label("fresh")
	field(b, "iput", ActivityDescriptor, "mediaRequest", permissionRequest, dex.P(1), 0)
	ask := func(label string, perms ...string) {
		b.Label(label)
		requestPermissions(b, 0, perms, mediaRequest, 2, 3, 4)
		b.ReturnVoid()
	}
	// v5 camera, v6 microphone: ask exactly for what the page wants.
	branch(b, "if-eqz", "micOnly", 5)
	branch(b, "if-eqz", "cameraOnly", 6)
	ask("both", "android.permission.CAMERA", "android.permission.RECORD_AUDIO")
	ask("cameraOnly", "android.permission.CAMERA")
	ask("micOnly", "android.permission.RECORD_AUDIO")
	b.Label("deny")
	call(b, "invoke-virtual", permissionRequest, "deny", "()V", dex.P(1))
	b.ReturnVoid()
	return dex.Method{Name: "onPermissionRequest", Descriptor: "(" + permissionRequest + ")V", Access: dex.Public, Code: b.Build()}
}

// grantedMethod emits MainActivity.granted([I, Z): with all true, whether
// every result is PERMISSION_GRANTED; otherwise whether any is. An empty
// array (a cancelled dialog) is never granted.
func grantedMethod() dex.Method {
	b := dex.NewCode(3)
	branch(b, "if-eqz", "no", dex.P(0))
	b.Emit(dex.Instruction{Op: "array-length", Regs: []dex.Register{0, dex.P(0)}})
	branch(b, "if-eqz", "no", 0)
	b.Const(1, 0)
	b.Label("loop")
	branch(b, "if-ge", "done", 1, 0)
	b.Emit(dex.Instruction{Op: "aget", Regs: []dex.Register{2, dex.P(0), 1}})
	branch(b, "if-eqz", "grantedOne", 2)
	branch(b, "if-nez", "no", dex.P(1)) // all required and this one was denied
	b.Emit(dex.Instruction{Op: "goto", Target: "next"})
	b.Label("grantedOne")
	branch(b, "if-eqz", "yes", dex.P(1)) // any is enough
	b.Label("next")
	b.Emit(dex.Instruction{Op: "add-int/lit8", Regs: []dex.Register{1, 1}, Literal: 1})
	b.Emit(dex.Instruction{Op: "goto", Target: "loop"})
	b.Label("done")
	// Every result was granted (all) or none was (any).
	branch(b, "if-nez", "yes", dex.P(1))
	b.Label("no").Const(0, 0)
	retScalar(b, 0)
	b.Label("yes").Const(0, 1)
	retScalar(b, 0)
	return dex.Method{Name: "granted", Descriptor: "([IZ)Z", Access: dex.Public | dex.Static, Code: b.Build()}
}

// permissionsResultMethod replaces onRequestPermissionsResult when web
// features are on, dispatching by request code.
func permissionsResultMethod(c config.Config) dex.Method {
	b := dex.NewCode(4)
	call(b, "invoke-super", activityClass, "onRequestPermissionsResult", "(I[Ljava/lang/String;[I)V", dex.P(0), dex.P(1), dex.P(2), dex.P(3))
	if c.HasPlugin("notifications") {
		b.Const(0, 100)
		branch(b, "if-ne", "geo", dex.P(1), 0)
		call(b, "invoke-virtual", ActivityDescriptor, "completeNotificationPermission", "()V", dex.P(0))
		b.ReturnVoid()
	}
	b.Label("geo")
	if c.HasWebFeature("geolocation") {
		b.Const(0, geoRequest)
		branch(b, "if-ne", "media", dex.P(1), 0)
		field(b, "iget", ActivityDescriptor, "geoCallback", geoCallback, 1, dex.P(0))
		branch(b, "if-eqz", "end", 1)
		field(b, "iget", ActivityDescriptor, "geoOrigin", "Ljava/lang/String;", 2, dex.P(0))
		b.Const(0, 0)
		field(b, "iput", ActivityDescriptor, "geoCallback", geoCallback, 0, dex.P(0))
		field(b, "iput", ActivityDescriptor, "geoOrigin", "Ljava/lang/String;", 0, dex.P(0))
		b.Const(3, 0) // any location permission, precise or approximate
		call(b, "invoke-static", ActivityDescriptor, "granted", "([IZ)Z", dex.P(3), 3)
		scalar(b, 0)
		call(b, "invoke-interface", geoCallback, "invoke", "(Ljava/lang/String;ZZ)V", 1, 2, 0, 3)
		b.ReturnVoid()
	}
	b.Label("media")
	if usesMedia(c) {
		b.Const(0, mediaRequest)
		branch(b, "if-ne", "end", dex.P(1), 0)
		field(b, "iget", ActivityDescriptor, "mediaRequest", permissionRequest, 1, dex.P(0))
		branch(b, "if-eqz", "end", 1)
		b.Const(0, 0)
		field(b, "iput", ActivityDescriptor, "mediaRequest", permissionRequest, 0, dex.P(0))
		b.Const(3, 1) // every permission the page asked for
		call(b, "invoke-static", ActivityDescriptor, "granted", "([IZ)Z", dex.P(3), 3)
		scalar(b, 0)
		branch(b, "if-eqz", "denyMedia", 0)
		call(b, "invoke-virtual", permissionRequest, "getResources", "()[Ljava/lang/String;", 1)
		object(b, 2)
		call(b, "invoke-virtual", permissionRequest, "grant", "([Ljava/lang/String;)V", 1, 2)
		b.ReturnVoid()
		b.Label("denyMedia")
		call(b, "invoke-virtual", permissionRequest, "deny", "()V", 1)
		b.ReturnVoid()
	}
	b.Label("end").ReturnVoid()
	return dex.Method{Name: "onRequestPermissionsResult", Descriptor: "(I[Ljava/lang/String;[I)V", Access: dex.Public, Code: b.Build()}
}

// applyWebFeatures adds the WebChromeClient and the Activity callbacks.
// Apps without webFeatures never reach it, so their bytecode is unchanged.
func applyWebFeatures(c config.Config, p *dex.Program) {
	a := &p.Classes[0]
	a.Fields = append(a.Fields, chromeFields(c)...)
	kept := a.Methods[:0]
	for _, m := range a.Methods {
		if m.Name != "onRequestPermissionsResult" {
			kept = append(kept, m)
		}
	}
	a.Methods = kept
	if c.HasWebFeature("files") {
		a.Methods = append(a.Methods, activityResultMethod())
	}
	if c.HasWebFeature("geolocation") || usesMedia(c) {
		a.Methods = append(a.Methods, grantedMethod())
	}
	if c.HasWebFeature("geolocation") || usesMedia(c) || c.HasPlugin("notifications") {
		a.Methods = append(a.Methods, permissionsResultMethod(c))
	}
	for j := range a.Methods {
		m := &a.Methods[j]
		if m.Name != "onCreate" {
			continue
		}
		for k, i := range m.Code.Instructions {
			if i.Method == nil || i.Method.Name != "setWebViewClient" {
				continue
			}
			// v0 is the WebView; v3 held the WebViewClient and is free again.
			extra := dex.NewCode(0)
			fresh(extra, 3, chromeClient)
			call(extra, "invoke-direct", chromeClient, "<init>", "("+ActivityDescriptor+")V", 3, dex.P(0))
			call(extra, "invoke-virtual", webview, "setWebChromeClient", "(Landroid/webkit/WebChromeClient;)V", 0, 3)
			rest := append([]dex.Instruction(nil), m.Code.Instructions[k+1:]...)
			m.Code.Instructions = append(append(m.Code.Instructions[:k+1], extra.Build().Instructions...), rest...)
			break
		}
	}
	p.Classes = append(p.Classes, chromeClass(c))
}
