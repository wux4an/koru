package runtimedex

import "koru.dev/koru/pkg/dex"

const workerClient = "Lcom/koru/runtime/WorkerClient;"

// Workers receive Application context only; they never retain Activity/JNI locals.
func serviceWorkerClient() dex.Class {
	cl := dex.Class{Descriptor: workerClient, Super: "Landroid/webkit/ServiceWorkerClient;", Access: dex.Public, Fields: []dex.Field{{Name: "context", Type: context, Access: dex.Private}}}
	b := dex.NewCode(0)
	call(b, "invoke-direct", cl.Super, "<init>", "()V", dex.P(0))
	field(b, "iput", workerClient, "context", context, dex.P(1), dex.P(0))
	b.ReturnVoid()
	cl.Methods = append(cl.Methods, dex.Method{Name: "<init>", Descriptor: "(" + context + ")V", Access: dex.Public | dex.Constructor, Code: b.Build()})
	b = dex.NewCode(1)
	field(b, "iget", workerClient, "context", context, 0, dex.P(0))
	call(b, "invoke-static", client, "intercept", "("+context+"Landroid/webkit/WebResourceRequest;)"+response, 0, dex.P(1))
	object(b, 0)
	retObject(b, 0)
	cl.Methods = append(cl.Methods, dex.Method{Name: "shouldInterceptRequest", Descriptor: "(Landroid/webkit/WebResourceRequest;)" + response, Access: dex.Public, Code: b.Build()})
	return cl
}
