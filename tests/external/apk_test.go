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

// Official tools validate only the finished artifact, never generate it.
func TestOfficialAPKValidation(t *testing.T) {
	aapt, java, signer := os.Getenv("KORU_TEST_AAPT"), os.Getenv("KORU_TEST_JAVA"), os.Getenv("KORU_TEST_APKSIGNER_JAR")
	if aapt == "" || java == "" || signer == "" {
		t.Fatal("set KORU_TEST_AAPT, KORU_TEST_JAVA and KORU_TEST_APKSIGNER_JAR for external validation")
	}
	root := t.TempDir()
	c := config.Default()
	c.Plugins = []string{"device", "storage", "vibration", "app", "share", "notifications"}
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
	data, err = app.Build(root, app.Options{Release: true})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "app.apk")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(java, "-jar", signer, "verify", "--verbose", path).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "Verified using v2 scheme (APK Signature Scheme v2): true") {
		t.Fatalf("official signature verifier: %v\n%s", err, out)
	}
	out, err = exec.Command(aapt, "dump", "badging", path).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "com.koru.runtime.MainActivity") {
		t.Fatalf("official resource parser: %v\n%s", err, out)
	}
	t.Logf("Go-built APK accepted by official resource and signature validators (%d bytes)", len(data))
}
