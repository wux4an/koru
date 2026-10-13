package runtimedex

import "koru.dev/koru/pkg/dex"

func notificationStatusMethod() dex.Method {
	b := dex.NewCode(2)
	b.ConstString(0, "notification")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", dex.P(0), 0)
	object(b, 0)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{0}, Type: "Landroid/app/NotificationManager;"})
	call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "areNotificationsEnabled", "()Z", 0)
	scalar(b, 0)
	branch(b, "if-eqz", "denied", 0)
	field(b, "sget", "Landroid/os/Build$VERSION;", "SDK_INT", "I", 0)
	b.Const(1, 33)
	branch(b, "if-lt", "granted", 0, 1)
	b.ConstString(1, "android.permission.POST_NOTIFICATIONS")
	call(b, "invoke-virtual", context, "checkSelfPermission", "(Ljava/lang/String;)I", dex.P(0), 1)
	scalar(b, 0)
	branch(b, "if-nez", "denied", 0)
	b.Label("granted").ConstString(0, "granted")
	retObject(b, 0)
	b.Label("denied").ConstString(0, "denied")
	retObject(b, 0)
	return dex.Method{Name: "getNotificationStatus", Descriptor: "()Ljava/lang/String;", Access: dex.Public, Code: b.Build()}
}
func completePermissionMethod() dex.Method {
	b := dex.NewCode(6)
	field(b, "iget", ActivityDescriptor, "permissionRequestId", "Ljava/lang/String;", 1, dex.P(0))
	branch(b, "if-eqz", "end", 1)
	field(b, "iget", ActivityDescriptor, "permissionPort", port, 4, dex.P(0))
	b.Const(0, 0)
	field(b, "iput", ActivityDescriptor, "permissionRequestId", "Ljava/lang/String;", 0, dex.P(0))
	field(b, "iput", ActivityDescriptor, "permissionPort", port, 0, dex.P(0))
	field(b, "iput", ActivityDescriptor, "permissionSettings", "Z", 0, dex.P(0))
	branch(b, "if-eqz", "end", 4)
	b.Label("start")
	call(b, "invoke-virtual", ActivityDescriptor, "getNotificationStatus", "()Ljava/lang/String;", dex.P(0))
	object(b, 3)
	fresh(b, 2, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "()V", 2)
	b.ConstString(5, "status")
	call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 2, 5, 3)
	call(b, "invoke-static", dispatcherClass, "success", "(Ljava/lang/String;"+jsonObject+")Ljava/lang/String;", 1, 2)
	object(b, 0)
	fresh(b, 2, message)
	call(b, "invoke-direct", message, "<init>", "(Ljava/lang/String;)V", 2, 0)
	call(b, "invoke-virtual", port, "postMessage", "("+message+")V", 4, 2)
	b.ReturnVoid().Label("tryEnd").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{0}}).ReturnVoid().Label("end").ReturnVoid().Catch("start", "tryEnd", "catch", "Ljava/lang/Exception;")
	return dex.Method{Name: "completeNotificationPermission", Descriptor: "()V", Access: dex.Public, Code: b.Build()}
}
func permissionResultMethod() dex.Method {
	b := dex.NewCode(1)
	call(b, "invoke-super", "Landroid/app/Activity;", "onRequestPermissionsResult", "(I[Ljava/lang/String;[I)V", dex.P(0), dex.P(1), dex.P(2), dex.P(3))
	b.Const(0, 100)
	branch(b, "if-ne", "end", dex.P(1), 0)
	call(b, "invoke-virtual", ActivityDescriptor, "completeNotificationPermission", "()V", dex.P(0))
	b.Label("end").ReturnVoid()
	return dex.Method{Name: "onRequestPermissionsResult", Descriptor: "(I[Ljava/lang/String;[I)V", Access: dex.Public, Code: b.Build()}
}
