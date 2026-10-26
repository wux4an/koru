//go:build android_external

package external_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"koru.dev/koru/internal/app"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
)

// The official aapt parses the configurable manifest the Go AXML writer emits.
// This checks that the attribute names resolve to the framework resource IDs;
// it does not prove the manifest behaves correctly on a device.
func TestOfficialAaptReadsConfigurableManifest(t *testing.T) {
	aapt := os.Getenv("KORU_TEST_AAPT")
	if aapt == "" {
		t.Fatal("set KORU_TEST_AAPT for external validation")
	}
	root := t.TempDir()
	c := config.Default()
	c.Orientation, c.Keyboard, c.AllowBackup = "portrait", "pan", true
	c.DeepLinks = []config.DeepLink{{Scheme: "https", Host: "example.com", PathPrefix: "/app"}, {Scheme: "myapp"}}
	for _, name := range []string{"www", ".koru"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string][]byte{"koru.json": data, "www/index.html": []byte("<!doctype html><html><head></head><body>Koru</body></html>")} {
		if err := os.WriteFile(filepath.Join(root, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := signing.CreateIdentity(filepath.Join(root, c.SigningKey)); err != nil {
		t.Fatal(err)
	}
	apk, err := app.Build(root, app.Options{Release: true})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "app.apk")
	if err := os.WriteFile(path, apk, 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(aapt, "dump", "xmltree", path, "AndroidManifest.xml").CombinedOutput()
	if err != nil {
		t.Fatalf("aapt dump xmltree: %v\n%s", err, out)
	}
	tree := string(out)
	for _, want := range []string{
		"android:screenOrientation(0x0101001e)=(type 0x10)0x1",
		"android:windowSoftInputMode(0x0101022b)=(type 0x10)0x20",
		"android:allowBackup(0x01010280)=(type 0x12)0xffffffff",
		"android:autoVerify(0x010104ee)=(type 0x12)0xffffffff",
		`android:name(0x01010003)="android.intent.action.VIEW"`,
		`android:name(0x01010003)="android.intent.category.BROWSABLE"`,
		`android:scheme(0x01010027)="https"`,
		`android:host(0x01010028)="example.com"`,
		`android:pathPrefix(0x0101002b)="/app"`,
		`android:scheme(0x01010027)="myapp"`,
	} {
		if !strings.Contains(tree, want) {
			t.Errorf("aapt xmltree lacks %s\n%s", want, tree)
		}
	}
}
