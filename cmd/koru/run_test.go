package main

import (
	"bytes"
	"context"
	"koru.dev/koru/internal/adb"
	"koru.dev/koru/internal/adb/adbtest"
	"koru.dev/koru/internal/app"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuildsInstallsAndOpens(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	dir := filepath.Join(t.TempDir(), "phone-app")
	var out bytes.Buffer
	if e := run([]string{"create", dir}, &out); e != nil {
		t.Fatal(e)
	}
	client := adb.Client{Addr: fake.Addr()}
	out.Reset()
	if e := runOnDevice(context.Background(), client, dir, "", false, false, false, &out); e != nil {
		t.Fatal(e)
	}
	report, e := app.Verify(fake.Installed())
	if e != nil || report.Package != "com.example.phoneapp" {
		t.Fatalf("installed APK: %+v %v", report, e)
	}
	cmds := strings.Join(fake.Commands(), "\n")
	if !strings.Contains(cmds, "exec:am start -W -n com.example.phoneapp/com.koru.runtime.MainActivity") {
		t.Fatal(cmds)
	}
	for _, want := range []string{"Installed", "sdk gphone x86 64 (emulator-5554)", "chrome://inspect"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q: %s", want, out.String())
		}
	}
	fake.SetReply("Failure [INSTALL_FAILED_UPDATE_INCOMPATIBLE]\n")
	if e = runOnDevice(context.Background(), client, dir, "", true, false, false, &out); e == nil || !strings.Contains(e.Error(), "Uninstall") {
		t.Fatal(e)
	}
	if e = runOnDevice(context.Background(), client, dir, "missing", false, false, false, &out); e == nil || !strings.Contains(e.Error(), "koru devices") {
		t.Fatal(e)
	}
	out.Reset()
	if e = listDevices(context.Background(), client, &out); e != nil || !strings.Contains(out.String(), "emulator-5554") {
		t.Fatal(out.String(), e)
	}
}
