package runtimedex

import (
	"fmt"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/pkg/dex"
	"koru.dev/koru/plugins"
	"sort"
)

const Origin = "https://app.koru.invalid"
const client = "Lcom/koru/runtime/LocalClient;"
const port = "Landroid/webkit/WebMessagePort;"
const callback = "Lcom/koru/runtime/PortCallback;"
const response = "Landroid/webkit/WebResourceResponse;"
const context = "Landroid/content/Context;"
const uri = "Landroid/net/Uri;"
const webview = "Landroid/webkit/WebView;"
const message = "Landroid/webkit/WebMessage;"

type Asset struct{ Path, MIME string }

// Runtime emits a controlled-origin WebView runtime. Paths are a compile-time allowlist.
func Runtime(c config.Config, assets []Asset, custom ...plugins.Descriptor) dex.Program {
	p := ActivityExperiment()
	a := &p.Classes[0]
	a.Fields = append(a.Fields, dex.Field{Name: "port", Type: port, Access: dex.Private})
	for j := range a.Methods {
		m := &a.Methods[j]
		if m.Name == "onCreate" {
			for k := range m.Code.Instructions {
				i := &m.Code.Instructions[k]
				if i.Type == "Landroid/webkit/WebViewClient;" {
					i.Type = client
				}
				if i.Method != nil && i.Method.Owner == "Landroid/webkit/WebViewClient;" && i.Method.Name == "<init>" {
					i.Method.Owner = client
					i.Method.Descriptor = "(" + ActivityDescriptor + ")V"
					i.Regs = append(i.Regs, dex.P(0))
				}
				if i.Op == "const-string" && i.String == "file:///android_asset/www/index.html" {
					i.String = Origin + "/index.html"
				}
			}
		}
		if m.Name == "onBackPressed" {
			back := dex.NewCode(3)
			field(back, "iget", ActivityDescriptor, "web", webview, 0, dex.P(0))
			branch(back, "if-eqz", "base", 0)
			call(back, "invoke-virtual", webview, "canGoBack", "()Z", 0)
			scalar(back, 1)
			branch(back, "if-nez", "native", 1)
			// Chromium can skip history entries added without activation in its UI API.
			// The actual history index still distinguishes a document transition from root.
			call(back, "invoke-virtual", webview, "copyBackForwardList", "()Landroid/webkit/WebBackForwardList;", 0)
			object(back, 1)
			call(back, "invoke-virtual", "Landroid/webkit/WebBackForwardList;", "getCurrentIndex", "()I", 1)
			scalar(back, 1)
			back.Const(2, 1)
			branch(back, "if-lt", "base", 1, 2)
			back.ConstString(1, "history.back();").Const(2, 0)
			call(back, "invoke-virtual", webview, "evaluateJavascript", "(Ljava/lang/String;Landroid/webkit/ValueCallback;)V", 0, 1, 2)
			back.ReturnVoid().Label("native")
			call(back, "invoke-virtual", webview, "goBack", "()V", 0)
			back.ReturnVoid().Label("base")
			call(back, "invoke-super", "Landroid/app/Activity;", "onBackPressed", "()V", dex.P(0))
			back.ReturnVoid()
			m.Code = back.Build()
		}
		if m.Name == "onDestroy" {
			b := dex.NewCode(1)
			field(b, "iget", ActivityDescriptor, "port", port, dex.V(0), dex.P(0))
			branch(b, "if-eqz", "web", dex.V(0))
			call(b, "invoke-virtual", port, "close", "()V", dex.V(0))
			b.Label("web")
			for _, i := range m.Code.Instructions {
				b.Emit(i)
			}
			m.Code = b.Build()
		}
	}
	// LocalClient/PortCallback hold an Activity, so those generated classes need access
	// to the port field; use package visibility instead of private cross-class access.
	a.Fields[len(a.Fields)-1].Access = 0
	if c.HasPlugin("notifications") {
		a.Fields = append(a.Fields, dex.Field{Name: "permissionRequestId", Type: "Ljava/lang/String;"}, dex.Field{Name: "permissionPort", Type: port}, dex.Field{Name: "permissionSettings", Type: "Z"})
		a.Methods = append(a.Methods, notificationStatusMethod(), completePermissionMethod(), permissionResultMethod())
		for j := range a.Methods {
			m := &a.Methods[j]
			if m.Name == "onResume" {
				m.Code.Instructions = m.Code.Instructions[:len(m.Code.Instructions)-1]
				resume := dex.NewCode(0)
				field(resume, "iget", ActivityDescriptor, "permissionSettings", "Z", 0, dex.P(0))
				branch(resume, "if-eqz", "resumeDone", 0)
				call(resume, "invoke-virtual", ActivityDescriptor, "completeNotificationPermission", "()V", dex.P(0))
				resume.Label("resumeDone").ReturnVoid()
				m.Code.Instructions = append(m.Code.Instructions, resume.Build().Instructions...)
			}
		}
	}
	// Consume Back before a focused WebView can process the same hardware event.
	keys := dex.NewCode(2)
	call(keys, "invoke-virtual", "Landroid/view/KeyEvent;", "getKeyCode", "()I", dex.P(1))
	scalar(keys, 0)
	keys.Const(1, 4)
	branch(keys, "if-ne", "delegate", 0, 1)
	call(keys, "invoke-virtual", "Landroid/view/KeyEvent;", "getAction", "()I", dex.P(1))
	scalar(keys, 0)
	keys.Const(1, 1)
	branch(keys, "if-ne", "consume", 0, 1)
	call(keys, "invoke-virtual", "Landroid/view/KeyEvent;", "isCanceled", "()Z", dex.P(1))
	scalar(keys, 0)
	branch(keys, "if-nez", "consume", 0)
	call(keys, "invoke-virtual", ActivityDescriptor, "onBackPressed", "()V", dex.P(0))
	keys.Label("consume").Const(0, 1)
	retScalar(keys, 0)
	keys.Label("delegate")
	call(keys, "invoke-super", "Landroid/app/Activity;", "dispatchKeyEvent", "(Landroid/view/KeyEvent;)Z", dex.P(0), dex.P(1))
	scalar(keys, 0)
	retScalar(keys, 0)
	a.Methods = append(a.Methods, dex.Method{Name: "dispatchKeyEvent", Descriptor: "(Landroid/view/KeyEvent;)Z", Access: dex.Public, Code: keys.Build()})
	// Register a worker client backed by Application context, not by this Activity.
	for j := range a.Methods {
		m := &a.Methods[j]
		if m.Name != "onCreate" {
			continue
		}
		for k, i := range m.Code.Instructions {
			if i.Method != nil && i.Method.Name == "setWebViewClient" {
				extra := dex.NewCode(0)
				call(extra, "invoke-virtual", context, "getApplicationContext", "()"+context, dex.P(0))
				object(extra, 2)
				fresh(extra, 1, workerClient)
				call(extra, "invoke-direct", workerClient, "<init>", "("+context+")V", 1, 2)
				call(extra, "invoke-static", "Landroid/webkit/ServiceWorkerController;", "getInstance", "()Landroid/webkit/ServiceWorkerController;")
				object(extra, 2)
				call(extra, "invoke-virtual", "Landroid/webkit/ServiceWorkerController;", "setServiceWorkerClient", "(Landroid/webkit/ServiceWorkerClient;)V", 2, 1)
				call(extra, "invoke-virtual", "Landroid/webkit/ServiceWorkerController;", "getServiceWorkerWebSettings", "()Landroid/webkit/ServiceWorkerWebSettings;", 2)
				object(extra, 2)
				extra.Const(1, 0)
				for _, name := range []string{"setAllowFileAccess", "setAllowContentAccess"} {
					call(extra, "invoke-virtual", "Landroid/webkit/ServiceWorkerWebSettings;", name, "(Z)V", 2, 1)
				}
				block := int32(1)
				if c.HasPermission("android.permission.INTERNET") {
					block = 0
				}
				extra.Const(1, block)
				call(extra, "invoke-virtual", "Landroid/webkit/ServiceWorkerWebSettings;", "setBlockNetworkLoads", "(Z)V", 2, 1)
				prefix := append([]dex.Instruction(nil), m.Code.Instructions[:k+1]...)
				m.Code.Instructions = append(prefix, append(extra.Build().Instructions, m.Code.Instructions[k+1:]...)...)
				break
			}
		}
	}
	p.Classes = append(p.Classes, localClient(c, assets), portCallback(), dispatcher(c, custom), serviceWorkerClient())
	for _, plugin := range custom {
		p.Classes = append(p.Classes, plugin.Classes...)
	}
	return p
}
func localClient(c config.Config, assets []Asset) dex.Class {
	cl := dex.Class{Descriptor: client, Super: "Landroid/webkit/WebViewClient;", Access: dex.Public, Fields: []dex.Field{{Name: "activity", Type: ActivityDescriptor, Access: dex.Private}}}
	b := dex.NewCode(0)
	call(b, "invoke-direct", cl.Super, "<init>", "()V", dex.P(0))
	field(b, "iput", client, "activity", ActivityDescriptor, dex.P(1), dex.P(0))
	b.ReturnVoid()
	cl.Methods = append(cl.Methods, dex.Method{Name: "<init>", Descriptor: "(" + ActivityDescriptor + ")V", Access: dex.Public | dex.Constructor, Code: b.Build()})
	b = dex.NewCode(3)
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", dex.P(0))
	object(b, 0)
	b.ConstString(1, "https")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
	scalar(b, 2)
	branch(b, "if-eqz", "no", 2)
	call(b, "invoke-virtual", uri, "getAuthority", "()Ljava/lang/String;", dex.P(0))
	object(b, 0)
	b.ConstString(1, "app.koru.invalid")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
	scalar(b, 2)
	retScalar(b, 2)
	b.Label("no").Const(2, 0)
	retScalar(b, 2)
	cl.Methods = append(cl.Methods, dex.Method{Name: "allowed", Descriptor: "(" + uri + ")Z", Access: dex.Public | dex.Static, Code: b.Build()})
	for _, request := range []bool{true, false} {
		b = dex.NewCode(2)
		desc := "(" + webview + "Ljava/lang/String;)Z"
		if request {
			desc = "(" + webview + "Landroid/webkit/WebResourceRequest;)Z"
			call(b, "invoke-interface", "Landroid/webkit/WebResourceRequest;", "getUrl", "()"+uri, dex.P(2))
			object(b, 0)
		} else {
			call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, dex.P(2))
			object(b, 0)
		}
		call(b, "invoke-static", client, "allowed", "("+uri+")Z", 0)
		scalar(b, 0)
		branch(b, "if-eqz", "blocked", 0)
		b.Const(1, 0)
		retScalar(b, 1)
		b.Label("blocked").Const(1, 1)
		retScalar(b, 1)
		cl.Methods = append(cl.Methods, dex.Method{Name: "shouldOverrideUrlLoading", Descriptor: desc, Access: dex.Public, Code: b.Build()})
	}
	cl.Methods = append(cl.Methods, fallbackResponse(), openResponse())
	b = dex.NewCode(6)
	call(b, "invoke-interface", "Landroid/webkit/WebResourceRequest;", "getUrl", "()"+uri, dex.P(2))
	object(b, 0)
	call(b, "invoke-static", client, "allowed", "("+uri+")Z", 0)
	scalar(b, 1)
	branch(b, "if-nez", "local", 1)
	if c.HasPermission("android.permission.INTERNET") {
		b.Const(0, 0)
		retObject(b, 0)
	} else {
		call(b, "invoke-static", client, "notFound", "()"+response)
		object(b, 0)
		retObject(b, 0)
	}
	b.Label("local")
	call(b, "invoke-virtual", uri, "getPath", "()Ljava/lang/String;", 0)
	object(b, 2)
	field(b, "iget", client, "activity", ActivityDescriptor, 5, dex.P(0))
	assets = append([]Asset(nil), assets...)
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	assets = append([]Asset{{Path: "/", MIME: "text/html"}}, assets...)
	for i, a := range assets {
		b.ConstString(3, a.Path)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 3, 2)
		scalar(b, 4)
		label := fmt.Sprintf("asset%d", i)
		branch(b, "if-eqz", label, 4)
		asset := a.Path
		if asset == "/" {
			asset = "/index.html"
		}
		b.ConstString(3, "www"+asset).ConstString(4, a.MIME)
		call(b, "invoke-static", client, "open", "("+context+"Ljava/lang/String;Ljava/lang/String;)"+response, 5, 3, 4)
		object(b, 0)
		retObject(b, 0)
		b.Label(label)
	}
	if c.SPA {
		b.ConstString(3, ".")
		call(b, "invoke-virtual", "Ljava/lang/String;", "contains", "(Ljava/lang/CharSequence;)Z", 2, 3)
		scalar(b, 4)
		branch(b, "if-nez", "missing", 4)
		b.ConstString(3, "www/index.html").ConstString(4, "text/html")
		call(b, "invoke-static", client, "open", "("+context+"Ljava/lang/String;Ljava/lang/String;)"+response, 5, 3, 4)
		object(b, 0)
		retObject(b, 0)
	}
	b.Label("missing")
	call(b, "invoke-static", client, "notFound", "()"+response)
	object(b, 0)
	retObject(b, 0)
	intercept := b.Build()
	for j, i := range intercept.Instructions {
		if i.Op == "iget" && i.Field != nil && i.Field.Name == "activity" {
			intercept.Instructions[j] = dex.Instruction{Op: "move-object", Regs: []dex.Register{5, dex.P(0)}}
		}
		if i.Method != nil && i.Method.Owner == "Landroid/webkit/WebResourceRequest;" && i.Method.Name == "getUrl" {
			intercept.Instructions[j].Regs = []dex.Register{dex.P(1)}
		}
	}
	cl.Methods = append(cl.Methods, dex.Method{Name: "intercept", Descriptor: "(" + context + "Landroid/webkit/WebResourceRequest;)" + response, Access: dex.Public | dex.Static, Code: intercept})
	wrapper := dex.NewCode(1)
	field(wrapper, "iget", client, "activity", ActivityDescriptor, 0, dex.P(0))
	call(wrapper, "invoke-static", client, "intercept", "("+context+"Landroid/webkit/WebResourceRequest;)"+response, 0, dex.P(2))
	object(wrapper, 0)
	retObject(wrapper, 0)
	cl.Methods = append(cl.Methods, dex.Method{Name: "shouldInterceptRequest", Descriptor: "(" + webview + "Landroid/webkit/WebResourceRequest;)" + response, Access: dex.Public, Code: wrapper.Build()})
	cl.Methods = append(cl.Methods, pageFinished())
	return cl
}
func fallbackResponse() dex.Method {
	b := dex.NewCode(5)
	fresh(b, 0, response)
	b.ConstString(1, "text/plain").ConstString(2, "UTF-8").Const(3, 0)
	b.Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{4, 3}, Type: "[B"})
	fresh(b, 3, "Ljava/io/ByteArrayInputStream;")
	call(b, "invoke-direct", "Ljava/io/ByteArrayInputStream;", "<init>", "([B)V", 3, 4)
	call(b, "invoke-direct", response, "<init>", "(Ljava/lang/String;Ljava/lang/String;Ljava/io/InputStream;)V", 0, 1, 2, 3)
	b.Const(1, 404).ConstString(2, "Not Found")
	call(b, "invoke-virtual", response, "setStatusCodeAndReasonPhrase", "(ILjava/lang/String;)V", 0, 1, 2)
	retObject(b, 0)
	return dex.Method{Name: "notFound", Descriptor: "()" + response, Access: dex.Public | dex.Static, Code: b.Build()}
}
func openResponse() dex.Method {
	b := dex.NewCode(8).Label("start")
	call(b, "invoke-virtual", context, "getAssets", "()Landroid/content/res/AssetManager;", dex.P(0))
	object(b, 6)
	call(b, "invoke-virtual", "Landroid/content/res/AssetManager;", "open", "(Ljava/lang/String;)Ljava/io/InputStream;", 6, dex.P(1))
	object(b, 6)
	fresh(b, 5, "Ljava/util/HashMap;")
	call(b, "invoke-direct", "Ljava/util/HashMap;", "<init>", "()V", 5)
	b.ConstString(7, "Content-Security-Policy").ConstString(1, "default-src 'self' data: blob:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' https:; frame-src 'self'; object-src 'none'; base-uri 'self'")
	call(b, "invoke-virtual", "Ljava/util/HashMap;", "put", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;", 5, 7, 1)
	b.ConstString(7, "Cache-Control").ConstString(1, "no-store")
	call(b, "invoke-virtual", "Ljava/util/HashMap;", "put", "(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;", 5, 7, 1)
	fresh(b, 0, response)
	moveObject(b, 1, dex.P(2))
	b.ConstString(2, "UTF-8").Const(3, 200).ConstString(4, "OK")
	call(b, "invoke-direct", response, "<init>", "(Ljava/lang/String;Ljava/lang/String;ILjava/lang/String;Ljava/util/Map;Ljava/io/InputStream;)V", 0, 1, 2, 3, 4, 5, 6)
	retObject(b, 0)
	b.Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{0}})
	call(b, "invoke-static", client, "notFound", "()"+response)
	object(b, 0)
	retObject(b, 0)
	b.Catch("start", "end", "catch", "Ljava/lang/Exception;")
	return dex.Method{Name: "open", Descriptor: "(" + context + "Ljava/lang/String;Ljava/lang/String;)" + response, Access: dex.Public | dex.Static, Code: b.Build()}
}
