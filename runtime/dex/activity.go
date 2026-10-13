// Package runtimedex generates Android runtime classes as Dalvik bytecode.
// ActivityExperiment preserves the initial file-asset experiment for comparison;
// normal APK builds use ControlledActivity and the HTTPS resource interceptor.
package runtimedex

import "koru.dev/koru/pkg/dex"

const ActivityDescriptor = "Lcom/koru/runtime/MainActivity;"

func ActivityExperiment() dex.Program {
	const activity = "Landroid/app/Activity;"
	const web = "Landroid/webkit/WebView;"
	const settings = "Landroid/webkit/WebSettings;"
	field := dex.FieldRef{Owner: ActivityDescriptor, Name: "web", Type: web}
	ref := func(owner, name, desc string) dex.MethodRef {
		return dex.MethodRef{Owner: owner, Name: name, Descriptor: desc}
	}
	invoke := func(b *dex.CodeBuilder, op, owner, name, desc string, regs ...dex.Register) {
		b.Invoke(op, ref(owner, name, desc), regs...)
	}
	v0, v1, v2, v3 := dex.V(0), dex.V(1), dex.V(2), dex.V(3)
	create := dex.NewCode(4)
	// p0 and p1 are resolved after the four local registers.
	invoke(create, "invoke-super", activity, "onCreate", "(Landroid/os/Bundle;)V", dex.P(0), dex.P(1))
	create.Emit(dex.Instruction{Op: "new-instance", Regs: []dex.Register{v0}, Type: web})
	invoke(create, "invoke-direct", web, "<init>", "(Landroid/content/Context;)V", v0, dex.P(0))
	create.Emit(dex.Instruction{Op: "iput", Regs: []dex.Register{v0, dex.P(0)}, Field: &field})
	invoke(create, "invoke-virtual", web, "getSettings", "()Landroid/webkit/WebSettings;", v0)
	create.Emit(dex.Instruction{Op: "move-result-object", Regs: []dex.Register{v1}})
	create.Emit(dex.Instruction{Op: "const/4", Regs: []dex.Register{v2}, Literal: 1})
	invoke(create, "invoke-virtual", settings, "setJavaScriptEnabled", "(Z)V", v1, v2)
	invoke(create, "invoke-virtual", settings, "setDomStorageEnabled", "(Z)V", v1, v2)
	create.Emit(dex.Instruction{Op: "const/4", Regs: []dex.Register{v2}, Literal: 0})
	// File assets remain readable when general file access is disabled, per WebSettings.
	for _, name := range []string{"setAllowFileAccess", "setAllowContentAccess", "setAllowFileAccessFromFileURLs", "setAllowUniversalAccessFromFileURLs"} {
		invoke(create, "invoke-virtual", settings, name, "(Z)V", v1, v2)
	}
	create.Emit(dex.Instruction{Op: "new-instance", Regs: []dex.Register{v3}, Type: "Landroid/webkit/WebViewClient;"})
	invoke(create, "invoke-direct", "Landroid/webkit/WebViewClient;", "<init>", "()V", v3)
	invoke(create, "invoke-virtual", web, "setWebViewClient", "(Landroid/webkit/WebViewClient;)V", v0, v3)
	invoke(create, "invoke-virtual", activity, "setContentView", "(Landroid/view/View;)V", dex.P(0), v0)
	create.ConstString(v1, "file:///android_asset/www/index.html")
	invoke(create, "invoke-virtual", web, "loadUrl", "(Ljava/lang/String;)V", v0, v1)
	create.ReturnVoid()
	ctor := dex.NewCode(0)
	invoke(ctor, "invoke-direct", activity, "<init>", "()V", dex.P(0))
	ctor.ReturnVoid()
	c := dex.Class{Descriptor: ActivityDescriptor, Super: activity, Access: dex.Public, Fields: []dex.Field{{Name: field.Name, Type: field.Type, Access: dex.Private}}, Methods: []dex.Method{{Name: "<init>", Descriptor: "()V", Access: dex.Public | dex.Constructor, Code: ctor.Build()}, {Name: "onCreate", Descriptor: "(Landroid/os/Bundle;)V", Access: dex.Protected, Code: create.Build()}}}
	for _, event := range []struct{ name, webMethod string }{{"onPause", "onPause"}, {"onResume", "onResume"}, {"onDestroy", "destroy"}} {
		b := dex.NewCode(1)
		if event.name != "onDestroy" {
			invoke(b, "invoke-super", activity, event.name, "()V", dex.P(0))
		}
		b.Emit(dex.Instruction{Op: "iget", Regs: []dex.Register{v0, dex.P(0)}, Field: &field}).Emit(dex.Instruction{Op: "if-eqz", Regs: []dex.Register{v0}, Target: "end"})
		if event.name == "onDestroy" {
			invoke(b, "invoke-virtual", web, "stopLoading", "()V", v0)
		}
		invoke(b, "invoke-virtual", web, event.webMethod, "()V", v0)
		b.Label("end")
		if event.name == "onDestroy" {
			invoke(b, "invoke-super", activity, event.name, "()V", dex.P(0))
		}
		b.ReturnVoid()
		c.Methods = append(c.Methods, dex.Method{Name: event.name, Descriptor: "()V", Access: dex.Protected, Code: b.Build()})
	}
	back := dex.NewCode(2)
	back.Emit(dex.Instruction{Op: "iget", Regs: []dex.Register{v0, dex.P(0)}, Field: &field}).Emit(dex.Instruction{Op: "if-eqz", Regs: []dex.Register{v0}, Target: "base"})
	invoke(back, "invoke-virtual", web, "canGoBack", "()Z", v0)
	back.Emit(dex.Instruction{Op: "move-result", Regs: []dex.Register{v1}}).Emit(dex.Instruction{Op: "if-eqz", Regs: []dex.Register{v1}, Target: "base"})
	invoke(back, "invoke-virtual", web, "goBack", "()V", v0)
	back.ReturnVoid().Label("base")
	invoke(back, "invoke-super", activity, "onBackPressed", "()V", dex.P(0))
	back.ReturnVoid()
	c.Methods = append(c.Methods, dex.Method{Name: "onBackPressed", Descriptor: "()V", Access: dex.Public, Code: back.Build()})
	return dex.Program{Classes: []dex.Class{c}}
}
