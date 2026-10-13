package runtimedex

import (
	"fmt"
	"koru.dev/koru/internal/arsc"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/pkg/dex"
	"koru.dev/koru/plugins"
	"sort"
)

const jsonObject = "Lorg/json/JSONObject;"

func put(b *dex.CodeBuilder, obj dex.Register, key string, value dex.Register, integer bool) {
	b.ConstString(5, key)
	desc := "(Ljava/lang/String;Ljava/lang/Object;)" + jsonObject
	if integer {
		desc = "(Ljava/lang/String;I)" + jsonObject
	}
	call(b, "invoke-virtual", jsonObject, "put", desc, obj, 5, value)
}
func text(b *dex.CodeBuilder, obj dex.Register, key, value string) {
	b.ConstString(7, value)
	put(b, obj, key, 7, false)
}
func success(b *dex.CodeBuilder) {
	call(b, "invoke-static", dispatcherClass, "success", "(Ljava/lang/String;"+jsonObject+")Ljava/lang/String;", 1, 8)
	object(b, 0)
	retObject(b, 0)
}
func nativeError(b *dex.CodeBuilder, code, msg string) {
	b.ConstString(5, code).ConstString(7, msg)
	call(b, "invoke-static", dispatcherClass, "error", "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;", 1, 5, 7)
	object(b, 0)
	retObject(b, 0)
}
func emptyResult(b *dex.CodeBuilder) {
	fresh(b, 8, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "()V", 8)
}
func paramString(b *dex.CodeBuilder, key string, dst dex.Register) {
	b.ConstString(5, key)
	call(b, "invoke-virtual", jsonObject, "opt", "(Ljava/lang/String;)Ljava/lang/Object;", 3, 5)
	object(b, dst)
	b.Emit(dex.Instruction{Op: "instance-of", Regs: []dex.Register{6, dst}, Type: "Ljava/lang/String;"})
	branch(b, "if-eqz", "invalid", 6)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{dst}, Type: "Ljava/lang/String;"})
}
func paramInt(b *dex.CodeBuilder, key string, def int32, dst dex.Register) {
	dflt := b.UniqueLabel("default")
	done := b.UniqueLabel("integer")
	b.ConstString(5, key)
	call(b, "invoke-virtual", jsonObject, "opt", "(Ljava/lang/String;)Ljava/lang/Object;", 3, 5)
	object(b, dst)
	branch(b, "if-eqz", dflt, dst)
	b.Emit(dex.Instruction{Op: "instance-of", Regs: []dex.Register{11, dst}, Type: "Ljava/lang/Integer;"})
	branch(b, "if-eqz", "invalid", 11)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{dst}, Type: "Ljava/lang/Integer;"})
	call(b, "invoke-virtual", "Ljava/lang/Integer;", "intValue", "()I", dst)
	scalar(b, dst)
	branch(b, "goto", done)
	b.Label(dflt).Const(dst, def).Label(done)
}
func dispatcher(c config.Config, custom []plugins.Descriptor) dex.Class {
	cl := dex.Class{Descriptor: dispatcherClass, Super: "Ljava/lang/Object;", Access: dex.Public, Methods: []dex.Method{successMethod(), errorMethod()}}
	b := dex.NewCode(13).ConstString(1, "").Label("start")
	branch(b, "if-eqz", "invalid", dex.P(1))
	call(b, "invoke-virtual", "Ljava/lang/String;", "length", "()I", dex.P(1))
	scalar(b, 6)
	b.Const(12, 65536)
	branch(b, "if-ge", "invalid", 6, 12)
	fresh(b, 0, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "(Ljava/lang/String;)V", 0, dex.P(1))
	b.ConstString(5, "id")
	call(b, "invoke-virtual", jsonObject, "optString", "(Ljava/lang/String;)Ljava/lang/String;", 0, 5)
	object(b, 1)
	b.ConstString(5, "[0-9]{1,10}")
	call(b, "invoke-virtual", "Ljava/lang/String;", "matches", "(Ljava/lang/String;)Z", 1, 5)
	scalar(b, 6)
	branch(b, "if-eqz", "invalid", 6)
	b.ConstString(5, "method")
	call(b, "invoke-virtual", jsonObject, "optString", "(Ljava/lang/String;)Ljava/lang/String;", 0, 5)
	object(b, 2)
	b.ConstString(5, "params")
	call(b, "invoke-virtual", jsonObject, "optJSONObject", "(Ljava/lang/String;)"+jsonObject, 0, 5)
	object(b, 3)
	branch(b, "if-eqz", "invalid", 3)
	routes := []struct {
		plugin, method string
		emit           func(*dex.CodeBuilder)
	}{
		{"device", "device.info", deviceInfo}, {"app", "app.info", func(b *dex.CodeBuilder) {
			emptyResult(b)
			text(b, 8, "appId", c.AppID)
			text(b, 8, "name", c.Name)
			text(b, 8, "version", c.Version)
			b.Const(7, int32(c.VersionCode))
			put(b, 8, "versionCode", 7, true)
			success(b)
		}},
		{"storage", "storage.get", func(b *dex.CodeBuilder) { storageOp(b, "get") }}, {"storage", "storage.set", func(b *dex.CodeBuilder) { storageOp(b, "set") }}, {"storage", "storage.remove", func(b *dex.CodeBuilder) { storageOp(b, "remove") }}, {"storage", "storage.clear", func(b *dex.CodeBuilder) { storageOp(b, "clear") }},
		{"vibration", "vibration.vibrate", vibrate}, {"share", "share.text", shareText},
		{"notifications", "notifications.show", func(b *dex.CodeBuilder) { notify(b, "show") }}, {"notifications", "notifications.cancel", func(b *dex.CodeBuilder) { notify(b, "cancel") }}, {"notifications", "notifications.permission", func(b *dex.CodeBuilder) { notificationPermission(b, false) }}, {"notifications", "notifications.requestPermission", func(b *dex.CodeBuilder) { notificationPermission(b, true) }},
	}
	for i, r := range routes {
		if !c.HasPlugin(r.plugin) {
			continue
		}
		b.ConstString(5, r.method)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 5, 2)
		scalar(b, 6)
		label := fmt.Sprintf("route%d", i)
		branch(b, "if-eqz", label, 6)
		r.emit(b)
		b.Label(label)
	}

	for _, plugin := range custom {
		names := make([]string, 0, len(plugin.Handlers))
		for name := range plugin.Handlers {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			label := b.UniqueLabel("plugin")
			b.ConstString(5, name)
			call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 5, 2)
			scalar(b, 6)
			branch(b, "if-eqz", label, 6)
			b.Invoke("invoke-static", plugin.Handlers[name], dex.P(0), 3)
			object(b, 8)
			branch(b, "if-eqz", "invalid", 8)
			success(b)
			b.Label(label)
		}
	}
	nativeError(b, "METHOD_NOT_FOUND", "Method unavailable or plugin disabled")
	b.Label("invalid")
	nativeError(b, "INVALID_REQUEST", "Invalid native request")
	b.Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{7}})
	nativeError(b, "NATIVE_ERROR", "Android operation failed")
	b.Catch("start", "end", "catch", "Ljava/lang/Throwable;")
	cl.Methods = append(cl.Methods, dex.Method{Name: "dispatch", Descriptor: "(" + ActivityDescriptor + "Ljava/lang/String;)Ljava/lang/String;", Access: dex.Public | dex.Static, Code: b.Build()})
	return cl
}
func successMethod() dex.Method {
	b := dex.NewCode(2)
	fresh(b, 0, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "()V", 0)
	b.ConstString(1, "id")
	call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 0, 1, dex.P(0))
	b.ConstString(1, "result")
	call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 0, 1, dex.P(1))
	call(b, "invoke-virtual", jsonObject, "toString", "()Ljava/lang/String;", 0)
	object(b, 0)
	retObject(b, 0)
	return dex.Method{Name: "success", Descriptor: "(Ljava/lang/String;" + jsonObject + ")Ljava/lang/String;", Access: dex.Public | dex.Static, Code: b.Build()}
}
func errorMethod() dex.Method {
	b := dex.NewCode(3)
	fresh(b, 0, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "()V", 0)
	fresh(b, 1, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "()V", 1)
	for _, x := range []struct {
		k string
		r dex.Register
	}{{"code", dex.P(1)}, {"message", dex.P(2)}} {
		b.ConstString(2, x.k)
		call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 1, 2, x.r)
	}
	b.ConstString(2, "id")
	call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 0, 2, dex.P(0))
	b.ConstString(2, "error")
	call(b, "invoke-virtual", jsonObject, "put", "(Ljava/lang/String;Ljava/lang/Object;)"+jsonObject, 0, 2, 1)
	call(b, "invoke-virtual", jsonObject, "toString", "()Ljava/lang/String;", 0)
	object(b, 0)
	retObject(b, 0)
	return dex.Method{Name: "error", Descriptor: "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;", Access: dex.Public | dex.Static, Code: b.Build()}
}
func deviceInfo(b *dex.CodeBuilder) {
	emptyResult(b)
	for _, f := range []struct{ owner, name, key string }{{"Landroid/os/Build;", "MODEL", "model"}, {"Landroid/os/Build;", "MANUFACTURER", "manufacturer"}, {"Landroid/os/Build$VERSION;", "RELEASE", "osVersion"}} {
		field(b, "sget", f.owner, f.name, "Ljava/lang/String;", 7)
		put(b, 8, f.key, 7, false)
	}
	field(b, "sget", "Landroid/os/Build$VERSION;", "SDK_INT", "I", 7)
	put(b, 8, "sdk", 7, true)
	text(b, 8, "platform", "android")
	success(b)
}
func storageOp(b *dex.CodeBuilder, op string) {
	if op != "clear" {
		paramString(b, "key", 9)
		b.ConstString(5, "[A-Za-z0-9_.:-]{1,128}")
		call(b, "invoke-virtual", "Ljava/lang/String;", "matches", "(Ljava/lang/String;)Z", 9, 5)
		scalar(b, 6)
		branch(b, "if-eqz", "invalid", 6)
	}
	b.ConstString(5, "koru").Const(6, 0)
	call(b, "invoke-virtual", context, "getSharedPreferences", "(Ljava/lang/String;I)Landroid/content/SharedPreferences;", dex.P(0), 5, 6)
	object(b, 7)
	emptyResult(b)
	if op == "get" {
		b.Const(6, 0)
		call(b, "invoke-interface", "Landroid/content/SharedPreferences;", "getString", "(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/String;", 7, 9, 6)
		object(b, 7)
		branch(b, "if-nez", "hasValue", 7)
		field(b, "sget", jsonObject, "NULL", "Ljava/lang/Object;", 7)
		b.Label("hasValue")
		put(b, 8, "value", 7, false)
		success(b)
		return
	}
	call(b, "invoke-interface", "Landroid/content/SharedPreferences;", "edit", "()Landroid/content/SharedPreferences$Editor;", 7)
	object(b, 7)
	switch op {
	case "set":
		paramString(b, "value", 10)
		call(b, "invoke-interface", "Landroid/content/SharedPreferences$Editor;", "putString", "(Ljava/lang/String;Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;", 7, 9, 10)
	case "remove":
		call(b, "invoke-interface", "Landroid/content/SharedPreferences$Editor;", "remove", "(Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;", 7, 9)
	case "clear":
		call(b, "invoke-interface", "Landroid/content/SharedPreferences$Editor;", "clear", "()Landroid/content/SharedPreferences$Editor;", 7)
	}
	call(b, "invoke-interface", "Landroid/content/SharedPreferences$Editor;", "commit", "()Z", 7)
	scalar(b, 6)
	label := "stored" + op
	branch(b, "if-nez", label, 6)
	nativeError(b, "STORAGE_ERROR", "Cannot persist storage")
	b.Label(label)
	success(b)
}
func vibrate(b *dex.CodeBuilder) {
	paramInt(b, "duration", 100, 6)
	b.Const(12, 1)
	branch(b, "if-lt", "invalid", 6, 12)
	b.Const(12, 10001)
	branch(b, "if-ge", "invalid", 6, 12)
	b.Emit(dex.Instruction{Op: "int-to-long", Regs: []dex.Register{10, 6}})
	b.Const(12, -1)
	call(b, "invoke-static", "Landroid/os/VibrationEffect;", "createOneShot", "(JI)Landroid/os/VibrationEffect;", 10, 11, 12)
	object(b, 9)
	b.ConstString(5, "vibrator")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", dex.P(0), 5)
	object(b, 7)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{7}, Type: "Landroid/os/Vibrator;"})
	call(b, "invoke-virtual", "Landroid/os/Vibrator;", "vibrate", "(Landroid/os/VibrationEffect;)V", 7, 9)
	emptyResult(b)
	success(b)
}
func shareText(b *dex.CodeBuilder) {
	paramString(b, "text", 9)
	fresh(b, 7, "Landroid/content/Intent;")
	b.ConstString(5, "android.intent.action.SEND")
	call(b, "invoke-direct", "Landroid/content/Intent;", "<init>", "(Ljava/lang/String;)V", 7, 5)
	b.ConstString(5, "text/plain")
	call(b, "invoke-virtual", "Landroid/content/Intent;", "setType", "(Ljava/lang/String;)Landroid/content/Intent;", 7, 5)
	b.ConstString(5, "android.intent.extra.TEXT")
	call(b, "invoke-virtual", "Landroid/content/Intent;", "putExtra", "(Ljava/lang/String;Ljava/lang/String;)Landroid/content/Intent;", 7, 5, 9)
	b.ConstString(5, "Share")
	call(b, "invoke-static", "Landroid/content/Intent;", "createChooser", "(Landroid/content/Intent;Ljava/lang/CharSequence;)Landroid/content/Intent;", 7, 5)
	object(b, 7)
	call(b, "invoke-virtual", "Landroid/app/Activity;", "startActivity", "(Landroid/content/Intent;)V", dex.P(0), 7)
	emptyResult(b)
	success(b)
}
func notificationPermission(b *dex.CodeBuilder, request bool) {
	granted := b.UniqueLabel("granted")
	settings := b.UniqueLabel("settings")
	start := b.UniqueLabel("request")
	call(b, "invoke-virtual", ActivityDescriptor, "getNotificationStatus", "()Ljava/lang/String;", dex.P(0))
	object(b, 9)
	b.ConstString(5, "granted")
	call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 5, 9)
	scalar(b, 6)
	branch(b, "if-nez", granted, 6)
	if !request {
		emptyResult(b)
		put(b, 8, "status", 9, false)
		success(b)
	} else {
		field(b, "iget", ActivityDescriptor, "permissionRequestId", "Ljava/lang/String;", 9, dex.P(0))
		branch(b, "if-eqz", start, 9)
		nativeError(b, "BUSY", "Notification permission request already pending")
		b.Label(start)
		field(b, "iput", ActivityDescriptor, "permissionRequestId", "Ljava/lang/String;", 1, dex.P(0))
		field(b, "iget", ActivityDescriptor, "port", port, 9, dex.P(0))
		field(b, "iput", ActivityDescriptor, "permissionPort", port, 9, dex.P(0))
		field(b, "sget", "Landroid/os/Build$VERSION;", "SDK_INT", "I", 6)
		b.Const(12, 33)
		branch(b, "if-lt", settings, 6, 12)
		b.Const(6, 1).Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{7, 6}, Type: "[Ljava/lang/String;"}).Const(6, 0).ConstString(5, "android.permission.POST_NOTIFICATIONS").Emit(dex.Instruction{Op: "aput-object", Regs: []dex.Register{5, 7, 6}}).Const(6, 100)
		call(b, "invoke-virtual", "Landroid/app/Activity;", "requestPermissions", "([Ljava/lang/String;I)V", dex.P(0), 7, 6)
		b.Const(0, 0)
		retObject(b, 0)
		b.Label(settings).Const(6, 1)
		field(b, "iput", ActivityDescriptor, "permissionSettings", "Z", 6, dex.P(0))
		fresh(b, 7, "Landroid/content/Intent;")
		b.ConstString(5, "android.settings.APP_NOTIFICATION_SETTINGS")
		call(b, "invoke-direct", "Landroid/content/Intent;", "<init>", "(Ljava/lang/String;)V", 7, 5)
		call(b, "invoke-virtual", context, "getPackageName", "()Ljava/lang/String;", dex.P(0))
		object(b, 9)
		b.ConstString(5, "android.provider.extra.APP_PACKAGE")
		call(b, "invoke-virtual", "Landroid/content/Intent;", "putExtra", "(Ljava/lang/String;Ljava/lang/String;)Landroid/content/Intent;", 7, 5, 9)
		call(b, "invoke-virtual", "Landroid/app/Activity;", "startActivity", "(Landroid/content/Intent;)V", dex.P(0), 7)
		b.Const(0, 0)
		retObject(b, 0)
	}
	b.Label(granted)
	emptyResult(b)
	text(b, 8, "status", "granted")
	success(b)
}
func notify(b *dex.CodeBuilder, op string) {
	if op == "show" {
		call(b, "invoke-virtual", ActivityDescriptor, "getNotificationStatus", "()Ljava/lang/String;", dex.P(0))
		object(b, 9)
		b.ConstString(5, "granted")
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 5, 9)
		scalar(b, 6)
		label := b.UniqueLabel("enabled")
		branch(b, "if-nez", label, 6)
		nativeError(b, "PERMISSION_DENIED", "Notifications are disabled")
		b.Label(label)
	}
	b.ConstString(5, "notification")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", dex.P(0), 5)
	object(b, 7)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{7}, Type: "Landroid/app/NotificationManager;"})
	paramInt(b, "id", 1, 6)
	if op == "cancel" {
		call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "cancel", "(I)V", 7, 6)
		emptyResult(b)
		success(b)
		return
	}
	field(b, "sget", "Landroid/os/Build$VERSION;", "SDK_INT", "I", 6)
	b.Const(12, 33)
	branch(b, "if-lt", "notifyGranted", 6, 12)
	b.ConstString(5, "android.permission.POST_NOTIFICATIONS")
	call(b, "invoke-virtual", context, "checkSelfPermission", "(Ljava/lang/String;)I", dex.P(0), 5)
	scalar(b, 6)
	branch(b, "if-eqz", "notifyGranted", 6)
	nativeError(b, "PERMISSION_DENIED", "Notification permission required")
	b.Label("notifyGranted")
	fresh(b, 9, "Landroid/app/NotificationChannel;")
	b.ConstString(5, "koru-local").ConstString(10, "Local notifications").Const(12, 3)
	call(b, "invoke-direct", "Landroid/app/NotificationChannel;", "<init>", "(Ljava/lang/String;Ljava/lang/CharSequence;I)V", 9, 5, 10, 12)
	call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "createNotificationChannel", "(Landroid/app/NotificationChannel;)V", 7, 9)
	fresh(b, 9, "Landroid/app/Notification$Builder;")
	call(b, "invoke-direct", "Landroid/app/Notification$Builder;", "<init>", "("+context+"Ljava/lang/String;)V", 9, dex.P(0), 5)
	b.Const(6, int32(arsc.IconID))
	call(b, "invoke-virtual", "Landroid/app/Notification$Builder;", "setSmallIcon", "(I)Landroid/app/Notification$Builder;", 9, 6)
	paramString(b, "title", 10)
	call(b, "invoke-virtual", "Landroid/app/Notification$Builder;", "setContentTitle", "(Ljava/lang/CharSequence;)Landroid/app/Notification$Builder;", 9, 10)
	paramString(b, "body", 10)
	call(b, "invoke-virtual", "Landroid/app/Notification$Builder;", "setContentText", "(Ljava/lang/CharSequence;)Landroid/app/Notification$Builder;", 9, 10)
	call(b, "invoke-virtual", "Landroid/app/Notification$Builder;", "build", "()Landroid/app/Notification;", 9)
	object(b, 9)
	paramInt(b, "id", 1, 6)
	call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "notify", "(ILandroid/app/Notification;)V", 7, 6, 9)
	emptyResult(b)
	success(b)
}
