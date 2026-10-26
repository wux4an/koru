package adb

import (
	"context"
	"errors"
	"koru.dev/koru/internal/adb/adbtest"
	"net"
	"strings"
	"testing"
)

func newFake(t *testing.T) *adbtest.Server {
	f, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func client(f *adbtest.Server) Client { return Client{Addr: f.Addr()} }

func TestVersionDevicesInstallStart(t *testing.T) {
	f := newFake(t)
	c := client(f)
	ctx := context.Background()
	if v, err := c.Version(ctx); err != nil || v != 41 {
		t.Fatal(v, err)
	}
	devices, err := c.Devices(ctx)
	if err != nil || len(devices) != 1 || devices[0].Model != "sdk gphone x86 64" {
		t.Fatalf("%+v %v", devices, err)
	}
	apk := []byte(strings.Repeat("apk!", 50000))
	if err = c.Install(ctx, "emulator-5554", apk); err != nil {
		t.Fatal(err)
	}
	if string(f.Installed()) != string(apk) {
		t.Fatal("APK bytes differ")
	}
	if err = c.Start(ctx, "emulator-5554", "com.example.app/com.koru.runtime.MainActivity"); err != nil {
		t.Fatal(err)
	}
	if err = c.Start(ctx, "emulator-5554", "com.example.app/.Missing"); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatal(err)
	}
	if err = c.Install(ctx, "other", apk); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatal(err)
	}
}

func TestInstallFailuresExplainTheFix(t *testing.T) {
	f := newFake(t)
	for out, want := range map[string]string{
		"Failure [INSTALL_FAILED_UPDATE_INCOMPATIBLE: signatures do not match]\n": "different signing key",
		"Failure [INSTALL_FAILED_VERSION_DOWNGRADE]\n":                            "increase versionCode",
		"Failure [INSTALL_FAILED_OLDER_SDK]\n":                                    "minSdk",
		"Failure [INSTALL_PARSE_FAILED_NO_CERTIFICATES]\n":                        "INSTALL_PARSE_FAILED_NO_CERTIFICATES",
		"": "no result",
	} {
		f.SetReply(out)
		err := client(f).Install(context.Background(), "emulator-5554", []byte("x"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%q: %v", out, err)
		}
	}
}

func TestPick(t *testing.T) {
	one := []Device{{Serial: "a", State: "device"}}
	two := []Device{{Serial: "a", State: "device"}, {Serial: "b", State: "device"}}
	if d, err := Pick(one, ""); err != nil || d.Serial != "a" {
		t.Fatal(d, err)
	}
	if _, err := Pick(two, ""); err == nil || !strings.Contains(err.Error(), "--device") {
		t.Fatal(err)
	}
	if d, err := Pick(two, "b"); err != nil || d.Serial != "b" {
		t.Fatal(d, err)
	}
	if _, err := Pick(two, "c"); err == nil {
		t.Fatal("unknown serial accepted")
	}
	if _, err := Pick(nil, ""); err == nil || !strings.Contains(err.Error(), "USB debugging") {
		t.Fatal(err)
	}
	if _, err := Pick([]Device{{Serial: "z", State: "unauthorized"}}, ""); err == nil || !strings.Contains(err.Error(), "Allow USB debugging") {
		t.Fatal(err)
	}
	if _, err := Pick([]Device{{Serial: "z", State: "offline"}, {Serial: "y", State: "device"}}, ""); err != nil {
		t.Fatal(err)
	}
}

func TestNoServerAndAddress(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	ln.Close()
	_, err := Client{Addr: addr}.Devices(context.Background())
	if !errors.Is(err, ErrNoServer) || !strings.Contains(err.Error(), "adb start-server") {
		t.Fatal(err)
	}
	t.Setenv("ADB_SERVER_SOCKET", "")
	t.Setenv("ANDROID_ADB_SERVER_PORT", "")
	if a, _ := Address(); a != DefaultAddress {
		t.Fatal(a)
	}
	t.Setenv("ANDROID_ADB_SERVER_PORT", "5038")
	if a, _ := Address(); a != "127.0.0.1:5038" {
		t.Fatal(a)
	}
	t.Setenv("ANDROID_ADB_SERVER_PORT", "x")
	if _, err := Address(); err == nil {
		t.Fatal("bad port accepted")
	}
	t.Setenv("ADB_SERVER_SOCKET", "tcp:10.0.2.2:5037")
	if a, _ := Address(); a != "10.0.2.2:5037" {
		t.Fatal(a)
	}
	t.Setenv("ADB_SERVER_SOCKET", "localfilesystem:/x")
	if _, err := Address(); err == nil {
		t.Fatal("unix socket accepted")
	}
}
