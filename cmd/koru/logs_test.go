package main

import (
	"bytes"
	"context"
	"koru.dev/koru/internal/adb"
	"koru.dev/koru/internal/adb/adbtest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testPidLog = "I/chromium( 4242): [INFO:CONSOLE(7)] \"hello from js\", source: https://app.koru.invalid/ (7)\n" +
		"W/ActivityManager( 4242): slow operation\n"
	// 4242 repeats the console line (must be dropped); 9999 is another process.
	testChromiumLog = "I/chromium( 4242): [INFO:CONSOLE(7)] \"hello from js\", source: https://app.koru.invalid/ (7)\n" +
		"E/chromium( 9999): [ERROR:CONSOLE(3)] \"other webview\"\n"
)

func waitFor(t *testing.T, out *lockedBuffer, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(out.String(), want) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("output lacks %q:\n%s", want, out.String())
	}
}

func TestFollowLogs(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	client := adb.Client{Addr: fake.Addr()}

	// App not open: explain how to open it.
	e := followLogs(context.Background(), client, adbtest.Serial, "com.example.app", 1, &lockedBuffer{})
	if e == nil || !strings.Contains(e.Error(), "koru run") {
		t.Fatalf("not running error = %v", e)
	}
	if e = followLogs(context.Background(), client, adbtest.Serial, "bad; rm -rf", 1, &lockedBuffer{}); e == nil || !strings.Contains(e.Error(), "package name") {
		t.Fatalf("bad appId error = %v", e)
	}

	fake.SetRunning("com.example.app", 4242)
	fake.SetLogcat(testPidLog, testChromiumLog)
	ctx, cancel := context.WithCancel(context.Background())
	out := &lockedBuffer{}
	done := make(chan error, 1)
	go func() { done <- followLogs(ctx, client, adbtest.Serial, "com.example.app", 1, out) }()
	waitFor(t, out, "W/ActivityManager( 4242): slow operation")
	waitFor(t, out, "other webview")
	cancel()
	select {
	case e = <-done:
		if e != nil {
			t.Fatalf("Ctrl+C must end cleanly: %v", e)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("followLogs did not stop")
	}
	if n := strings.Count(out.String(), "hello from js"); n != 1 {
		t.Fatalf("console line printed %d times:\n%s", n, out.String())
	}
	cmds := strings.Join(fake.Commands(), "\n")
	for _, want := range []string{"exec:pidof com.example.app", "exec:logcat --pid=4242 -v brief", "exec:logcat -v brief -s chromium:V"} {
		if !strings.Contains(cmds, want) {
			t.Fatalf("adb commands lack %q:\n%s", want, cmds)
		}
	}
}

func TestLineWriterJoinsPartialLines(t *testing.T) {
	var dst bytes.Buffer
	w := &lineWriter{dst: &dst, keep: func(l []byte) bool { return !bytes.HasPrefix(l, []byte("drop")) }}
	w.Write([]byte("one\r\ndr"))
	w.Write([]byte("op me\ntw"))
	w.Write([]byte("o"))
	if err := w.flush(); err != nil {
		t.Fatal(err)
	}
	if dst.String() != "one\ntwo\n" {
		t.Fatalf("got %q", dst.String())
	}
}

func TestLogsCommandReportsClosedApp(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	t.Setenv("ADB_SERVER_SOCKET", "tcp:"+fake.Addr())
	t.Setenv("ANDROID_SERIAL", "")
	dir := filepath.Join(t.TempDir(), "logapp")
	var out bytes.Buffer
	if e := run([]string{"create", dir}, &out); e != nil {
		t.Fatal(e)
	}
	e := run([]string{"logs", dir}, &out)
	if e == nil || !strings.Contains(e.Error(), "com.example.logapp is not running") || !strings.Contains(e.Error(), "koru run") {
		t.Fatalf("error = %v", e)
	}
	if e = run([]string{"logs", dir, "--device", "other"}, &out); e == nil || !strings.Contains(e.Error(), "koru devices") {
		t.Fatalf("error = %v", e)
	}
}

func TestRunWithLogsFollowsAfterOpening(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	dir := filepath.Join(t.TempDir(), "phone-app")
	var created bytes.Buffer
	if e := run([]string{"create", dir}, &created); e != nil {
		t.Fatal(e)
	}
	fake.SetRunning("com.example.phoneapp", 4242)
	fake.SetLogcat(testPidLog, "")
	ctx, cancel := context.WithCancel(context.Background())
	out := &lockedBuffer{}
	done := make(chan error, 1)
	go func() { done <- runOnDevice(ctx, adb.Client{Addr: fake.Addr()}, dir, "", false, false, true, out) }()
	waitFor(t, out, "hello from js")
	if !strings.Contains(out.String(), "Opened") {
		t.Fatal(out.String())
	}
	cancel()
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}
