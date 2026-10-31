package runtimedex

import (
	"strings"

	"github.com/wux4an/koru/internal/config"
	"github.com/wux4an/koru/pkg/dex"
)

const intent = "Landroid/content/Intent;"
const activityClass = "Landroid/app/Activity;"

// externalSchemes open in another app when a link leaves the app. Others,
// such as intent: or javascript:, are dropped.
var externalSchemes = []string{"https", "http", "mailto", "tel", "sms", "geo", "market"}

// UserAgentMarker tells core.js on a remote page that it runs inside a koru
// app and should wait for the native channel. Only apps with server.bridge
// add it.
const UserAgentMarker = " KoruApp/1"

// inAppMethod emits LocalClient.inApp(Uri): the app origin, or an https URL
// whose host is the server url host or in allowNavigation.
func inAppMethod(c config.Config) dex.Method {
	b := dex.NewCode(3)
	call(b, "invoke-static", client, "allowed", "("+uri+")Z", dex.P(0))
	scalar(b, 0)
	branch(b, "if-nez", "yes", 0)
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", dex.P(0))
	object(b, 1)
	branch(b, "if-eqz", "no", 1)
	b.ConstString(2, "https")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 2, 1)
	scalar(b, 0)
	branch(b, "if-eqz", "no", 0)
	call(b, "invoke-virtual", uri, "getHost", "()Ljava/lang/String;", dex.P(0))
	object(b, 1)
	branch(b, "if-eqz", "no", 1)
	for _, h := range c.Server.Hosts() {
		if suffix, ok := strings.CutPrefix(h, "*"); ok {
			// "*.example.com" matches subdomains only, as config.HostAllowed does.
			b.ConstString(2, suffix)
			call(b, "invoke-virtual", "Ljava/lang/String;", "endsWith", "(Ljava/lang/String;)Z", 1, 2)
		} else {
			b.ConstString(2, h)
			call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 2, 1)
		}
		scalar(b, 0)
		branch(b, "if-nez", "yes", 0)
	}
	b.Label("no").Const(0, 0)
	retScalar(b, 0)
	b.Label("yes").Const(0, 1)
	retScalar(b, 0)
	return dex.Method{Name: "inApp", Descriptor: "(" + uri + ")Z", Access: dex.Public | dex.Static, Code: b.Build()}
}

// openExternalMethod emits LocalClient.openExternal(Activity, Uri): hand the
// link to the app that handles it, such as the browser, mail or phone.
func openExternalMethod() dex.Method {
	b := dex.NewCode(3).Label("start")
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", dex.P(1))
	object(b, 0)
	branch(b, "if-eqz", "done", 0)
	for _, s := range externalSchemes {
		b.ConstString(1, s)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
		scalar(b, 2)
		branch(b, "if-nez", "open", 2)
	}
	b.Emit(dex.Instruction{Op: "goto", Target: "done"})
	b.Label("open")
	fresh(b, 0, intent)
	b.ConstString(1, "android.intent.action.VIEW")
	call(b, "invoke-direct", intent, "<init>", "(Ljava/lang/String;"+uri+")V", 0, 1, dex.P(1))
	b.ConstString(1, "android.intent.category.BROWSABLE")
	call(b, "invoke-virtual", intent, "addCategory", "(Ljava/lang/String;)"+intent, 0, 1)
	call(b, "invoke-virtual", activityClass, "startActivity", "("+intent+")V", dex.P(0), 0)
	b.Label("done").ReturnVoid()
	// No app for the link (ActivityNotFoundException) or a blocked intent: stay put.
	b.Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{0}}).ReturnVoid()
	b.Catch("start", "end", "catch", "Ljava/lang/Exception;")
	return dex.Method{Name: "openExternal", Descriptor: "(" + activityClass + uri + ")V", Access: dex.Public | dex.Static, Code: b.Build()}
}

// overrideMethods emits shouldOverrideUrlLoading for apps with a server
// block: in-app hosts load in the WebView, frames load where they are, and
// any other main-frame link opens outside the app.
func overrideMethods() []dex.Method {
	var out []dex.Method
	for _, request := range []bool{true, false} {
		b := dex.NewCode(2)
		desc := "(" + webview + "Ljava/lang/String;)Z"
		if request {
			desc = "(" + webview + "Landroid/webkit/WebResourceRequest;)Z"
			call(b, "invoke-interface", "Landroid/webkit/WebResourceRequest;", "isForMainFrame", "()Z", dex.P(2))
			scalar(b, 0)
			branch(b, "if-eqz", "allow", 0)
			call(b, "invoke-interface", "Landroid/webkit/WebResourceRequest;", "getUrl", "()"+uri, dex.P(2))
			object(b, 0)
		} else {
			call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, dex.P(2))
			object(b, 0)
		}
		call(b, "invoke-static", client, "inApp", "("+uri+")Z", 0)
		scalar(b, 1)
		branch(b, "if-nez", "allow", 1)
		field(b, "iget", client, "activity", ActivityDescriptor, 1, dex.P(0))
		call(b, "invoke-static", client, "openExternal", "("+activityClass+uri+")V", 1, 0)
		b.Const(1, 1)
		retScalar(b, 1)
		b.Label("allow").Const(1, 0)
		retScalar(b, 1)
		out = append(out, dex.Method{Name: "shouldOverrideUrlLoading", Descriptor: desc, Access: dex.Public, Code: b.Build()})
	}
	return out
}

// bridgePageFinished replaces onPageFinished when server.bridge is set: the
// app origin and each bridge origin get the native channel, posted to exactly
// that origin; any other page is told "koru-none" so core.js stops waiting
// and uses its web implementations.
func bridgePageFinished(origins []string) dex.Method {
	b := dex.NewCode(9)
	b.ConstString(7, "Koru").ConstString(8, "onPageFinished")
	call(b, "invoke-static", "Landroid/util/Log;", "i", "(Ljava/lang/String;Ljava/lang/String;)I", 7, 8)
	call(b, "invoke-super", "Landroid/webkit/WebViewClient;", "onPageFinished", "("+webview+"Ljava/lang/String;)V", dex.P(0), dex.P(1), dex.P(2))
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, dex.P(2))
	object(b, 0)
	call(b, "invoke-static", client, "allowed", "("+uri+")Z", 0)
	scalar(b, 1)
	branch(b, "if-eqz", "remote", 1)
	b.ConstString(7, Origin)
	b.Emit(dex.Instruction{Op: "goto", Target: "channel"})
	b.Label("remote")
	call(b, "invoke-virtual", uri, "getScheme", "()Ljava/lang/String;", 0)
	object(b, 1)
	branch(b, "if-eqz", "none", 1)
	call(b, "invoke-virtual", uri, "getAuthority", "()Ljava/lang/String;", 0)
	object(b, 2)
	branch(b, "if-eqz", "none", 2)
	const builder = "Ljava/lang/StringBuilder;"
	fresh(b, 3, builder)
	call(b, "invoke-direct", builder, "<init>", "()V", 3)
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 3, 1)
	b.ConstString(4, "://")
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 3, 4)
	call(b, "invoke-virtual", builder, "append", "(Ljava/lang/String;)"+builder, 3, 2)
	call(b, "invoke-virtual", builder, "toString", "()Ljava/lang/String;", 3)
	object(b, 7)
	for _, o := range origins {
		b.ConstString(1, o)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 7)
		scalar(b, 2)
		branch(b, "if-nez", "channel", 2)
	}
	b.Label("none")
	fresh(b, 0, message)
	b.ConstString(1, "koru-none")
	call(b, "invoke-direct", message, "<init>", "(Ljava/lang/String;)V", 0, 1)
	b.ConstString(1, "*")
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, 1)
	object(b, 1)
	call(b, "invoke-virtual", webview, "postWebMessage", "("+message+uri+")V", dex.P(1), 0, 1)
	b.ReturnVoid()
	b.Label("channel")
	call(b, "invoke-virtual", webview, "createWebMessageChannel", "()[Landroid/webkit/WebMessagePort;", dex.P(1))
	object(b, 0)
	b.Const(1, 0).Emit(dex.Instruction{Op: "aget-object", Regs: []dex.Register{2, 0, 1}})
	field(b, "iget", client, "activity", ActivityDescriptor, 3, dex.P(0))
	field(b, "iget", ActivityDescriptor, "port", port, 4, 3)
	branch(b, "if-eqz", "fresh", 4)
	call(b, "invoke-virtual", port, "close", "()V", 4)
	b.Label("fresh")
	field(b, "iput", ActivityDescriptor, "port", port, 2, 3)
	fresh(b, 4, callback)
	call(b, "invoke-direct", callback, "<init>", "("+ActivityDescriptor+")V", 4, 3)
	call(b, "invoke-virtual", port, "setWebMessageCallback", "(Landroid/webkit/WebMessagePort$WebMessageCallback;)V", 2, 4)
	b.Const(1, 1).Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{5, 1}, Type: "[Landroid/webkit/WebMessagePort;"}).Emit(dex.Instruction{Op: "aget-object", Regs: []dex.Register{6, 0, 1}}).Const(1, 0).Emit(dex.Instruction{Op: "aput-object", Regs: []dex.Register{6, 5, 1}})
	fresh(b, 0, message)
	b.ConstString(1, "koru-ready")
	call(b, "invoke-direct", message, "<init>", "(Ljava/lang/String;[Landroid/webkit/WebMessagePort;)V", 0, 1, 5)
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, 7)
	object(b, 6)
	call(b, "invoke-virtual", webview, "postWebMessage", "("+message+uri+")V", dex.P(1), 0, 6)
	b.ConstString(7, "Koru").ConstString(8, "native port transferred")
	call(b, "invoke-static", "Landroid/util/Log;", "i", "(Ljava/lang/String;Ljava/lang/String;)I", 7, 8)
	b.ReturnVoid()
	return dex.Method{Name: "onPageFinished", Descriptor: "(" + webview + "Ljava/lang/String;)V", Access: dex.Public, Code: b.Build()}
}

// applyServer rewires a runtime for server.url, allowNavigation and bridge.
// Apps without a server block never reach it, so their bytecode is unchanged.
func applyServer(c config.Config, p *dex.Program) {
	s := c.Server
	for ci := range p.Classes {
		cl := &p.Classes[ci]
		switch cl.Descriptor {
		case ActivityDescriptor:
			for mi := range cl.Methods {
				m := &cl.Methods[mi]
				if m.Name != "onCreate" {
					continue
				}
				for k := range m.Code.Instructions {
					i := &m.Code.Instructions[k]
					if s.URL != "" && i.Op == "const-string" && i.String == Origin+"/index.html" {
						i.String = s.URL
					}
				}
				if len(s.Bridge) > 0 {
					markUserAgent(m)
				}
			}
		case client:
			kept := cl.Methods[:0]
			for _, m := range cl.Methods {
				if m.Name == "shouldOverrideUrlLoading" || (m.Name == "onPageFinished" && len(s.Bridge) > 0) {
					continue
				}
				kept = append(kept, m)
			}
			cl.Methods = append(kept, inAppMethod(c), openExternalMethod())
			cl.Methods = append(cl.Methods, overrideMethods()...)
			if len(s.Bridge) > 0 {
				cl.Methods = append(cl.Methods, bridgePageFinished(s.BridgeOrigins()))
			}
		}
	}
}

// markUserAgent appends UserAgentMarker right after the WebSettings are read
// in onCreate (v1 holds them, v2 is free scratch).
func markUserAgent(m *dex.Method) {
	for k, i := range m.Code.Instructions {
		if i.Method == nil || i.Method.Name != "setDomStorageEnabled" {
			continue
		}
		extra := dex.NewCode(0)
		call(extra, "invoke-virtual", "Landroid/webkit/WebSettings;", "getUserAgentString", "()Ljava/lang/String;", 1)
		object(extra, 2)
		extra.ConstString(3, UserAgentMarker)
		call(extra, "invoke-virtual", "Ljava/lang/String;", "concat", "(Ljava/lang/String;)Ljava/lang/String;", 2, 3)
		object(extra, 2)
		call(extra, "invoke-virtual", "Landroid/webkit/WebSettings;", "setUserAgentString", "(Ljava/lang/String;)V", 1, 2)
		// setDomStorageEnabled read v2 as the boolean true; the code after it
		// reloads v2 with const before using it again.
		rest := append([]dex.Instruction(nil), m.Code.Instructions[k+1:]...)
		m.Code.Instructions = append(append(m.Code.Instructions[:k+1], extra.Build().Instructions...), rest...)
		return
	}
	panic("runtime: onCreate has no setDomStorageEnabled")
}
