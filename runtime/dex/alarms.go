package runtimedex

import (
	"github.com/wux4an/koru/internal/arsc"
	"github.com/wux4an/koru/pkg/dex"
)

// Scheduled notifications (the "alarms" plugin). Each alarm is stored as JSON
// in the private SharedPreferences "koru-alarms" and armed with AlarmManager.
// AlarmReceiver (not exported) shows the notification when it fires;
// BootReceiver re-arms the stored alarms after a reboot or an app update,
// which clear them. Delivery uses setAndAllowWhileIdle: Android may delay it
// by some minutes, especially in Doze; exact alarms need a permission the
// user grants in Settings, so they are not used.

const AlarmReceiverClass = "Lcom/koru/runtime/AlarmReceiver;"
const BootReceiverClass = "Lcom/koru/runtime/BootReceiver;"
const alarmManager = "Landroid/app/AlarmManager;"
const pendingIntent = "Landroid/app/PendingIntent;"
const prefs = "Landroid/content/SharedPreferences;"
const prefsEditor = "Landroid/content/SharedPreferences$Editor;"
const jsonArray = "Lorg/json/JSONArray;"

// PendingIntent flags: FLAG_IMMUTABLE 0x04000000, FLAG_UPDATE_CURRENT
// 0x08000000, FLAG_NO_CREATE 0x20000000.
const (
	piImmutableUpdate   = 0x0c000000
	piImmutableNoCreate = 0x24000000
)

// alarmRoutes are the dispatcher routes of the alarms plugin. The dispatcher
// holds params in v3 and the Activity in p0.
func alarmRoutes() []route {
	return []route{
		{"alarms", "alarms.schedule", func(b *dex.CodeBuilder) {
			paramInt(b, "id", 1, 4)
			paramString(b, "title", 9)
			paramString(b, "body", 12)
			b.ConstString(5, "at")
			call(b, "invoke-virtual", jsonObject, "opt", "(Ljava/lang/String;)Ljava/lang/Object;", 3, 5)
			object(b, 10)
			b.Emit(dex.Instruction{Op: "instance-of", Regs: []dex.Register{6, 10}, Type: "Ljava/lang/Number;"})
			branch(b, "if-eqz", "invalid", 6)
			call(b, "invoke-static", AlarmReceiverClass, "schedule", "("+context+jsonObject+")V", dex.P(0), 3)
			emptyResult(b)
			success(b)
		}},
		{"alarms", "alarms.cancel", func(b *dex.CodeBuilder) {
			paramInt(b, "id", 1, 4)
			call(b, "invoke-static", AlarmReceiverClass, "cancel", "("+context+"I)V", dex.P(0), 4)
			emptyResult(b)
			success(b)
		}},
		{"alarms", "alarms.list", func(b *dex.CodeBuilder) {
			call(b, "invoke-static", AlarmReceiverClass, "list", "("+context+")"+jsonArray, dex.P(0))
			object(b, 9)
			emptyResult(b)
			put(b, 8, "ids", 9, false)
			success(b)
		}},
	}
}

func alarmPrefs(b *dex.CodeBuilder, ctx, dst, tmp dex.Register) {
	b.ConstString(dst, "koru-alarms").Const(tmp, 0)
	call(b, "invoke-virtual", context, "getSharedPreferences", "(Ljava/lang/String;I)"+prefs, ctx, dst, tmp)
	object(b, dst)
}

func alarmIntent(b *dex.CodeBuilder, ctx, dst, tmp dex.Register) {
	fresh(b, dst, intent)
	b.Emit(dex.Instruction{Op: "const-class", Regs: []dex.Register{tmp}, Type: AlarmReceiverClass})
	call(b, "invoke-direct", intent, "<init>", "("+context+"Ljava/lang/Class;)V", dst, ctx, tmp)
}

func alarmService(b *dex.CodeBuilder, ctx, dst dex.Register) {
	b.ConstString(dst, "alarm")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", ctx, dst)
	object(b, dst)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{dst}, Type: alarmManager})
}

func receiverClasses() []dex.Class {
	alarm := ctor(AlarmReceiverClass, "Landroid/content/BroadcastReceiver;", nil)

	// arm(Context, JSONObject alarm): register the alarm with AlarmManager.
	b := dex.NewCode(6)
	b.ConstString(0, "id")
	call(b, "invoke-virtual", jsonObject, "optInt", "(Ljava/lang/String;)I", dex.P(1), 0)
	scalar(b, 1)
	alarmIntent(b, dex.P(0), 2, 3)
	b.ConstString(0, "alarm")
	call(b, "invoke-virtual", jsonObject, "toString", "()Ljava/lang/String;", dex.P(1))
	object(b, 3)
	call(b, "invoke-virtual", intent, "putExtra", "(Ljava/lang/String;Ljava/lang/String;)"+intent, 2, 0, 3)
	b.Const(0, piImmutableUpdate)
	call(b, "invoke-static", pendingIntent, "getBroadcast", "("+context+"I"+intent+"I)"+pendingIntent, dex.P(0), 1, 2, 0)
	object(b, 2)
	alarmService(b, dex.P(0), 3)
	b.ConstString(0, "at")
	call(b, "invoke-virtual", jsonObject, "optLong", "(Ljava/lang/String;)J", dex.P(1), 0)
	b.Emit(dex.Instruction{Op: "move-result-wide", Regs: []dex.Register{4}})
	b.Const(0, 0) // AlarmManager.RTC_WAKEUP
	call(b, "invoke-virtual", alarmManager, "setAndAllowWhileIdle", "(IJ"+pendingIntent+")V", 3, 0, 4, 5, 2)
	b.ReturnVoid()
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "arm", Descriptor: "(" + context + jsonObject + ")V", Access: dex.Public | dex.Static, Code: b.Build()})

	// schedule(Context, JSONObject alarm): store it, then arm it.
	b = dex.NewCode(4)
	alarmPrefs(b, dex.P(0), 0, 1)
	call(b, "invoke-interface", prefs, "edit", "()"+prefsEditor, 0)
	object(b, 0)
	b.ConstString(1, "id")
	call(b, "invoke-virtual", jsonObject, "optInt", "(Ljava/lang/String;)I", dex.P(1), 1)
	scalar(b, 2)
	call(b, "invoke-static", "Ljava/lang/String;", "valueOf", "(I)Ljava/lang/String;", 2)
	object(b, 2)
	call(b, "invoke-virtual", jsonObject, "toString", "()Ljava/lang/String;", dex.P(1))
	object(b, 3)
	call(b, "invoke-interface", prefsEditor, "putString", "(Ljava/lang/String;Ljava/lang/String;)"+prefsEditor, 0, 2, 3)
	call(b, "invoke-interface", prefsEditor, "commit", "()Z", 0)
	call(b, "invoke-static", AlarmReceiverClass, "arm", "("+context+jsonObject+")V", dex.P(0), dex.P(1))
	b.ReturnVoid()
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "schedule", Descriptor: "(" + context + jsonObject + ")V", Access: dex.Public | dex.Static, Code: b.Build()})

	// forget(Context, int id): drop the stored alarm.
	b = dex.NewCode(2)
	alarmPrefs(b, dex.P(0), 0, 1)
	call(b, "invoke-interface", prefs, "edit", "()"+prefsEditor, 0)
	object(b, 0)
	call(b, "invoke-static", "Ljava/lang/String;", "valueOf", "(I)Ljava/lang/String;", dex.P(1))
	object(b, 1)
	call(b, "invoke-interface", prefsEditor, "remove", "(Ljava/lang/String;)"+prefsEditor, 0, 1)
	call(b, "invoke-interface", prefsEditor, "commit", "()Z", 0)
	b.ReturnVoid()
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "forget", Descriptor: "(" + context + "I)V", Access: dex.Public | dex.Static, Code: b.Build()})

	// cancel(Context, int id): disarm and forget.
	b = dex.NewCode(3)
	alarmIntent(b, dex.P(0), 0, 1)
	b.Const(1, piImmutableNoCreate)
	call(b, "invoke-static", pendingIntent, "getBroadcast", "("+context+"I"+intent+"I)"+pendingIntent, dex.P(0), dex.P(1), 0, 1)
	object(b, 0)
	branch(b, "if-eqz", "forget", 0)
	alarmService(b, dex.P(0), 2)
	call(b, "invoke-virtual", alarmManager, "cancel", "("+pendingIntent+")V", 2, 0)
	call(b, "invoke-virtual", pendingIntent, "cancel", "()V", 0)
	b.Label("forget")
	call(b, "invoke-static", AlarmReceiverClass, "forget", "("+context+"I)V", dex.P(0), dex.P(1))
	b.ReturnVoid()
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "cancel", Descriptor: "(" + context + "I)V", Access: dex.Public | dex.Static, Code: b.Build()})

	// list(Context): the ids of the stored alarms.
	b = dex.NewCode(4)
	fresh(b, 0, jsonArray)
	call(b, "invoke-direct", jsonArray, "<init>", "()V", 0)
	alarmPrefs(b, dex.P(0), 1, 2)
	call(b, "invoke-interface", prefs, "getAll", "()Ljava/util/Map;", 1)
	object(b, 1)
	call(b, "invoke-interface", "Ljava/util/Map;", "keySet", "()Ljava/util/Set;", 1)
	object(b, 1)
	call(b, "invoke-interface", "Ljava/util/Set;", "iterator", "()Ljava/util/Iterator;", 1)
	object(b, 1)
	b.Label("loop")
	call(b, "invoke-interface", "Ljava/util/Iterator;", "hasNext", "()Z", 1)
	scalar(b, 2)
	branch(b, "if-eqz", "done", 2)
	call(b, "invoke-interface", "Ljava/util/Iterator;", "next", "()Ljava/lang/Object;", 1)
	object(b, 2)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{2}, Type: "Ljava/lang/String;"})
	call(b, "invoke-static", "Ljava/lang/Integer;", "parseInt", "(Ljava/lang/String;)I", 2)
	scalar(b, 3)
	call(b, "invoke-virtual", jsonArray, "put", "(I)"+jsonArray, 0, 3)
	b.Emit(dex.Instruction{Op: "goto", Target: "loop"})
	b.Label("done")
	retObject(b, 0)
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "list", Descriptor: "(" + context + ")" + jsonArray, Access: dex.Public | dex.Static, Code: b.Build()})

	// rearmAll(Context): after a reboot, arm every stored alarm again. Alarms
	// already due fire right away.
	b = dex.NewCode(3).Label("start")
	alarmPrefs(b, dex.P(0), 0, 1)
	call(b, "invoke-interface", prefs, "getAll", "()Ljava/util/Map;", 0)
	object(b, 0)
	call(b, "invoke-interface", "Ljava/util/Map;", "values", "()Ljava/util/Collection;", 0)
	object(b, 0)
	call(b, "invoke-interface", "Ljava/util/Collection;", "iterator", "()Ljava/util/Iterator;", 0)
	object(b, 0)
	b.Label("loop")
	call(b, "invoke-interface", "Ljava/util/Iterator;", "hasNext", "()Z", 0)
	scalar(b, 1)
	branch(b, "if-eqz", "done", 1)
	call(b, "invoke-interface", "Ljava/util/Iterator;", "next", "()Ljava/lang/Object;", 0)
	object(b, 1)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{1}, Type: "Ljava/lang/String;"})
	fresh(b, 2, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "(Ljava/lang/String;)V", 2, 1)
	call(b, "invoke-static", AlarmReceiverClass, "arm", "("+context+jsonObject+")V", dex.P(0), 2)
	b.Emit(dex.Instruction{Op: "goto", Target: "loop"})
	b.Label("done").ReturnVoid()
	b.Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{0}}).ReturnVoid()
	b.Catch("start", "end", "catch", "Ljava/lang/Exception;")
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "rearmAll", Descriptor: "(" + context + ")V", Access: dex.Public | dex.Static, Code: b.Build()})

	// onReceive: show the notification and forget the alarm. Tapping it opens
	// the app.
	const builder = "Landroid/app/Notification$Builder;"
	b = dex.NewCode(8).Label("start")
	b.ConstString(0, "alarm")
	call(b, "invoke-virtual", intent, "getStringExtra", "(Ljava/lang/String;)Ljava/lang/String;", dex.P(2), 0)
	object(b, 1)
	branch(b, "if-eqz", "end", 1)
	fresh(b, 2, jsonObject)
	call(b, "invoke-direct", jsonObject, "<init>", "(Ljava/lang/String;)V", 2, 1)
	b.ConstString(0, "id")
	call(b, "invoke-virtual", jsonObject, "optInt", "(Ljava/lang/String;)I", 2, 0)
	scalar(b, 3)
	b.ConstString(0, "title")
	call(b, "invoke-virtual", jsonObject, "optString", "(Ljava/lang/String;)Ljava/lang/String;", 2, 0)
	object(b, 4)
	b.ConstString(0, "body")
	call(b, "invoke-virtual", jsonObject, "optString", "(Ljava/lang/String;)Ljava/lang/String;", 2, 0)
	object(b, 5)
	call(b, "invoke-static", AlarmReceiverClass, "forget", "("+context+"I)V", dex.P(1), 3)
	b.ConstString(0, "notification")
	call(b, "invoke-virtual", context, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;", dex.P(1), 0)
	object(b, 6)
	b.Emit(dex.Instruction{Op: "check-cast", Regs: []dex.Register{6}, Type: "Landroid/app/NotificationManager;"})
	fresh(b, 7, "Landroid/app/NotificationChannel;")
	b.ConstString(0, "koru-local").ConstString(1, "Local notifications").Const(2, 3)
	call(b, "invoke-direct", "Landroid/app/NotificationChannel;", "<init>", "(Ljava/lang/String;Ljava/lang/CharSequence;I)V", 7, 0, 1, 2)
	call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "createNotificationChannel", "(Landroid/app/NotificationChannel;)V", 6, 7)
	fresh(b, 2, intent)
	b.Emit(dex.Instruction{Op: "const-class", Regs: []dex.Register{1}, Type: ActivityDescriptor})
	call(b, "invoke-direct", intent, "<init>", "("+context+"Ljava/lang/Class;)V", 2, dex.P(1), 1)
	b.Const(1, 0x04000000) // FLAG_IMMUTABLE
	call(b, "invoke-static", pendingIntent, "getActivity", "("+context+"I"+intent+"I)"+pendingIntent, dex.P(1), 3, 2, 1)
	object(b, 2)
	fresh(b, 7, builder)
	call(b, "invoke-direct", builder, "<init>", "("+context+"Ljava/lang/String;)V", 7, dex.P(1), 0)
	b.Const(1, int32(arsc.IconID))
	call(b, "invoke-virtual", builder, "setSmallIcon", "(I)"+builder, 7, 1)
	call(b, "invoke-virtual", builder, "setContentTitle", "(Ljava/lang/CharSequence;)"+builder, 7, 4)
	call(b, "invoke-virtual", builder, "setContentText", "(Ljava/lang/CharSequence;)"+builder, 7, 5)
	call(b, "invoke-virtual", builder, "setContentIntent", "("+pendingIntent+")"+builder, 7, 2)
	b.Const(1, 1)
	call(b, "invoke-virtual", builder, "setAutoCancel", "(Z)"+builder, 7, 1)
	call(b, "invoke-virtual", builder, "build", "()Landroid/app/Notification;", 7)
	object(b, 7)
	call(b, "invoke-virtual", "Landroid/app/NotificationManager;", "notify", "(ILandroid/app/Notification;)V", 6, 3, 7)
	b.Label("end").ReturnVoid()
	// A denied notification permission or a malformed alarm must not crash.
	b.Label("tryEnd").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{0}}).ReturnVoid()
	b.Catch("start", "tryEnd", "catch", "Ljava/lang/Exception;")
	alarm.Methods = append(alarm.Methods, dex.Method{Name: "onReceive", Descriptor: "(" + context + intent + ")V", Access: dex.Public, Code: b.Build()})

	// BootReceiver acts only on the protected system broadcasts, which other
	// apps cannot send, and ignores any extras.
	boot := ctor(BootReceiverClass, "Landroid/content/BroadcastReceiver;", nil)
	b = dex.NewCode(2)
	call(b, "invoke-virtual", intent, "getAction", "()Ljava/lang/String;", dex.P(2))
	object(b, 0)
	branch(b, "if-eqz", "end", 0)
	for _, action := range BootActions {
		b.ConstString(1, action)
		call(b, "invoke-virtual", "Ljava/lang/String;", "equals", "(Ljava/lang/Object;)Z", 1, 0)
		scalar(b, 1)
		branch(b, "if-nez", "rearm", 1)
	}
	b.Emit(dex.Instruction{Op: "goto", Target: "end"})
	b.Label("rearm")
	call(b, "invoke-static", AlarmReceiverClass, "rearmAll", "("+context+")V", dex.P(1))
	b.Label("end").ReturnVoid()
	boot.Methods = append(boot.Methods, dex.Method{Name: "onReceive", Descriptor: "(" + context + intent + ")V", Access: dex.Public, Code: b.Build()})
	return []dex.Class{alarm, boot}
}

// BootActions clear AlarmManager alarms; BootReceiver listens to them.
var BootActions = []string{"android.intent.action.BOOT_COMPLETED", "android.intent.action.MY_PACKAGE_REPLACED"}
