package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestProject(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "app")
	if e := run([]string{"init", dir}, &bytes.Buffer{}); e != nil {
		t.Fatal(e)
	}
	return dir
}

func readKoruJSON(t *testing.T, dir string) string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(dir, "koru.json"))
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}

func TestBumpLevels(t *testing.T) {
	dir := newTestProject(t)
	for _, tc := range []struct {
		args    []string
		version string
		code    string
	}{
		{[]string{"bump", dir}, `"version": "1.0.1"`, `"versionCode": 2`},
		{[]string{"bump", "minor", dir}, `"version": "1.1.0"`, `"versionCode": 3`},
		{[]string{"bump", "major", dir}, `"version": "2.0.0"`, `"versionCode": 4`},
		{[]string{"bump", "code", dir}, `"version": "2.0.0"`, `"versionCode": 5`},
		{[]string{"bump", "patch", dir}, `"version": "2.0.1"`, `"versionCode": 6`},
	} {
		if e := run(tc.args, &bytes.Buffer{}); e != nil {
			t.Fatal(tc.args, e)
		}
		got := readKoruJSON(t, dir)
		if !strings.Contains(got, tc.version) || !strings.Contains(got, tc.code) {
			t.Fatalf("%v: %s", tc.args, got)
		}
	}
	got := readKoruJSON(t, dir)
	if !strings.HasSuffix(got, "}\n") || strings.Index(got, `"appId"`) > strings.Index(got, `"versionCode"`) {
		t.Fatalf("field order or trailing newline lost:\n%s", got)
	}
}

func TestBumpNonSemverOnlyBumpsCode(t *testing.T) {
	dir := newTestProject(t)
	path := filepath.Join(dir, "koru.json")
	b, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(b), `"1.0.0"`, `"2024-beta"`, 1)), 0644)
	var out bytes.Buffer
	if e := run([]string{"bump", dir}, &out); e != nil {
		t.Fatal(e)
	}
	got := readKoruJSON(t, dir)
	if !strings.Contains(got, `"2024-beta"`) || !strings.Contains(got, `"versionCode": 2`) || !strings.Contains(out.String(), "Warning") {
		t.Fatalf("%s\n%s", got, out.String())
	}
}

func TestBumpRejectsOverflowAndMissingProject(t *testing.T) {
	dir := newTestProject(t)
	path := filepath.Join(dir, "koru.json")
	b, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(b), `"versionCode": 1`, `"versionCode": 2100000000`, 1)), 0644)
	if e := run([]string{"bump", dir}, &bytes.Buffer{}); e == nil || !strings.Contains(e.Error(), "limit") {
		t.Fatal(e)
	}
	if e := run([]string{"bump", filepath.Join(dir, "missing")}, &bytes.Buffer{}); e == nil {
		t.Fatal("missing project accepted")
	}
	if e := run([]string{"bump", "huge", dir}, &bytes.Buffer{}); e == nil {
		t.Fatal("unknown level accepted")
	}
}

func TestBuildBump(t *testing.T) {
	dir := newTestProject(t)
	apk := filepath.Join(t.TempDir(), "app.apk")
	if e := run([]string{"build", dir, "--bump", "--output", apk}, &bytes.Buffer{}); e != nil {
		t.Fatal(e)
	}
	if got := readKoruJSON(t, dir); !strings.Contains(got, `"version": "1.0.1"`) || !strings.Contains(got, `"versionCode": 2`) {
		t.Fatal(got)
	}
}

func TestPluginAddRemoveList(t *testing.T) {
	dir := newTestProject(t)
	var out bytes.Buffer
	if e := run([]string{"plugin", "add", "vibration", "share", "device", "--project", dir}, &out); e != nil {
		t.Fatal(e)
	}
	got := readKoruJSON(t, dir)
	if strings.Count(got, `"device"`) != 1 || !strings.Contains(got, `"vibration"`) || !strings.Contains(got, `"share"`) {
		t.Fatal(got)
	}
	out.Reset()
	if e := run([]string{"plugin", "list", dir}, &out); e != nil {
		t.Fatal(e)
	}
	list := out.String()
	for _, want := range []string{"on   vibration", "android.permission.VIBRATE", "off  notifications", "android.permission.POST_NOTIFICATIONS"} {
		if !strings.Contains(list, want) {
			t.Fatalf("list lacks %q:\n%s", want, list)
		}
	}
	if e := run([]string{"plugin", "remove", "vibration", "--project", dir}, &bytes.Buffer{}); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(readKoruJSON(t, dir), "vibration") {
		t.Fatal("plugin not removed")
	}
	before := readKoruJSON(t, dir)
	out.Reset()
	if e := run([]string{"plugin", "remove", "vibration", "--project", dir}, &out); e != nil || !strings.Contains(out.String(), "Unchanged") || readKoruJSON(t, dir) != before {
		t.Fatal(e, out.String())
	}
}

func TestPluginRejectsUnknown(t *testing.T) {
	dir := newTestProject(t)
	before := readKoruJSON(t, dir)
	e := run([]string{"plugin", "add", "vibration", "camera", "--project", dir}, &bytes.Buffer{})
	if e == nil || !strings.Contains(e.Error(), "valid plugins: device, storage, vibration, app, share, notifications") {
		t.Fatal(e)
	}
	if readKoruJSON(t, dir) != before {
		t.Fatal("config changed after a rejected add")
	}
}

func TestConfigCommand(t *testing.T) {
	dir := newTestProject(t)
	if e := run([]string{"plugin", "add", "vibration", "--project", dir}, &bytes.Buffer{}); e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	if e := run([]string{"config", dir}, &out); e != nil {
		t.Fatal(e)
	}
	var report struct {
		Config               map[string]any `json:"config"`
		EffectivePermissions []string       `json:"effectivePermissions"`
		WebRoot              string         `json:"webRoot"`
		SigningKey           string         `json:"signingKey"`
	}
	if e := json.Unmarshal(out.Bytes(), &report); e != nil {
		t.Fatal(e, out.String())
	}
	abs, _ := filepath.Abs(dir)
	if report.Config["appId"] != "com.example.koru" || report.WebRoot != filepath.Join(abs, "www") || report.SigningKey != filepath.Join(abs, ".koru", "signing.pem") {
		t.Fatalf("%+v", report)
	}
	if len(report.EffectivePermissions) != 1 || report.EffectivePermissions[0] != "android.permission.VIBRATE" {
		t.Fatalf("%v", report.EffectivePermissions)
	}
	if strings.Contains(out.String(), "PRIVATE KEY") {
		t.Fatal("key content printed")
	}
}

func doctorOutput(t *testing.T, dir string) string {
	t.Helper()
	var out bytes.Buffer
	if e := run([]string{"doctor", dir}, &out); e != nil {
		t.Fatal(e)
	}
	return out.String()
}

func TestDoctorWarnings(t *testing.T) {
	dir := newTestProject(t)
	// A fresh init project (no repository, viewport present) only warns about targetSdk.
	base := doctorOutput(t, dir)
	if !strings.Contains(base, "targetSdk is 36") || strings.Count(base, "Warning") != 1 {
		t.Fatal(base)
	}
	if strings.Contains(base, ".gitignore") || strings.Contains(base, "viewport") {
		t.Fatal(base)
	}

	// Ignore file without .koru/.
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n"), 0644)
	if !strings.Contains(doctorOutput(t, dir), ".koru/ is not listed in .gitignore") {
		t.Fatal("missing gitignore warning")
	}
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n.koru/\n"), 0644)
	if strings.Contains(doctorOutput(t, dir), ".gitignore") {
		t.Fatal("unexpected gitignore warning")
	}

	// Missing viewport and an https fetch without INTERNET.
	html := `<!doctype html><html><head><title>x</title></head><body><script>fetch("https://api.example.com/x")</script></body></html>`
	os.WriteFile(filepath.Join(dir, "www", "index.html"), []byte(html), 0644)
	got := doctorOutput(t, dir)
	if !strings.Contains(got, "no <meta name=\"viewport\"") || !strings.Contains(got, "index.html calls fetch()") || !strings.Contains(got, "heuristic") {
		t.Fatal(got)
	}
	path := filepath.Join(dir, "koru.json")
	b, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(b), `"permissions": []`, `"permissions": ["android.permission.INTERNET"]`, 1)), 0644)
	if strings.Contains(doctorOutput(t, dir), "calls fetch()") {
		t.Fatal("INTERNET declared but still warned")
	}

	// Non-square, small icon.
	var buf bytes.Buffer
	png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 64, 32)))
	os.WriteFile(filepath.Join(dir, "icon.png"), buf.Bytes(), 0644)
	b, _ = os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(b), `"signingKey"`, `"icon": "icon.png",`+"\n  "+`"signingKey"`, 1)), 0644)
	got = doctorOutput(t, dir)
	if !strings.Contains(got, "64x32; use a square PNG") || !strings.Contains(got, "at least 192x192") {
		t.Fatal(got)
	}

	// An invalid icon is an error, not a warning.
	os.WriteFile(filepath.Join(dir, "icon.png"), []byte("not a png"), 0644)
	if e := run([]string{"doctor", dir}, &bytes.Buffer{}); e == nil || !strings.Contains(e.Error(), "PNG") {
		t.Fatal(e)
	}
}
