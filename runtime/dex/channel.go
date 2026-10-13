package runtimedex

import "koru.dev/koru/pkg/dex"

const dispatcherClass = "Lcom/koru/runtime/Dispatcher;"

func pageFinished() dex.Method {
	b := dex.NewCode(9)
	b.ConstString(7, "Koru").ConstString(8, "onPageFinished")
	call(b, "invoke-static", "Landroid/util/Log;", "i", "(Ljava/lang/String;Ljava/lang/String;)I", 7, 8)
	call(b, "invoke-super", "Landroid/webkit/WebViewClient;", "onPageFinished", "("+webview+"Ljava/lang/String;)V", dex.P(0), dex.P(1), dex.P(2))
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, dex.P(2))
	object(b, 0)
	call(b, "invoke-static", client, "allowed", "("+uri+")Z", 0)
	scalar(b, 1)
	branch(b, "if-eqz", "end", 1)
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
	b.ConstString(6, Origin)
	call(b, "invoke-static", uri, "parse", "(Ljava/lang/String;)"+uri, 6)
	object(b, 6)
	call(b, "invoke-virtual", webview, "postWebMessage", "("+message+uri+")V", dex.P(1), 0, 6)
	b.ConstString(7, "Koru").ConstString(8, "native port transferred")
	call(b, "invoke-static", "Landroid/util/Log;", "i", "(Ljava/lang/String;Ljava/lang/String;)I", 7, 8)
	b.Label("end").ReturnVoid()
	return dex.Method{Name: "onPageFinished", Descriptor: "(" + webview + "Ljava/lang/String;)V", Access: dex.Public, Code: b.Build()}
}
func portCallback() dex.Class {
	cl := dex.Class{Descriptor: callback, Super: "Landroid/webkit/WebMessagePort$WebMessageCallback;", Access: dex.Public, Fields: []dex.Field{{Name: "activity", Type: ActivityDescriptor, Access: dex.Private}}}
	b := dex.NewCode(0)
	call(b, "invoke-direct", cl.Super, "<init>", "()V", dex.P(0))
	field(b, "iput", callback, "activity", ActivityDescriptor, dex.P(1), dex.P(0))
	b.ReturnVoid()
	cl.Methods = append(cl.Methods, dex.Method{Name: "<init>", Descriptor: "(" + ActivityDescriptor + ")V", Access: dex.Public | dex.Constructor, Code: b.Build()})
	b = dex.NewCode(3).Label("start")
	call(b, "invoke-virtual", message, "getData", "()Ljava/lang/String;", dex.P(2))
	object(b, 0)
	field(b, "iget", callback, "activity", ActivityDescriptor, 1, dex.P(0))
	call(b, "invoke-static", dispatcherClass, "dispatch", "("+ActivityDescriptor+"Ljava/lang/String;)Ljava/lang/String;", 1, 0)
	object(b, 0)
	branch(b, "if-eqz", "done", 0)
	fresh(b, 1, message)
	call(b, "invoke-direct", message, "<init>", "(Ljava/lang/String;)V", 1, 0)
	call(b, "invoke-virtual", port, "postMessage", "("+message+")V", dex.P(1), 1)
	b.Label("done").ReturnVoid().Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{2}}).ReturnVoid().Catch("start", "end", "catch", "Ljava/lang/Exception;")
	cl.Methods = append(cl.Methods, dex.Method{Name: "onMessage", Descriptor: "(" + port + message + ")V", Access: dex.Public, Code: b.Build()})
	return cl
}
