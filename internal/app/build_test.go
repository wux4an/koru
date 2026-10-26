package app

import (
	"bytes"
	"encoding/json"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
	"testing"
)

func project(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	os.MkdirAll(filepath.Join(p, "www"), 0755)
	os.MkdirAll(filepath.Join(p, ".koru"), 0700)
	c := config.Default()
	data, _ := json.Marshal(c)
	os.WriteFile(filepath.Join(p, "koru.json"), data, 0644)
	os.WriteFile(filepath.Join(p, "www/index.html"), []byte("<!doctype html><head></head><body>Koru</body>"), 0644)
	if _, e := signing.CreateIdentity(filepath.Join(p, c.SigningKey)); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestFullAPKBuildVerifyAndKeyReuse(t *testing.T) {
	p := project(t)
	a, e := Build(p, Options{})
	if e != nil {
		t.Fatal(e)
	}
	r, e := Verify(a)
	if e != nil || r.Package != "com.example.koru" {
		t.Fatal(r, e)
	}
	b, e := Build(p, Options{})
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("nondeterminism", e)
	}
	release, e := Build(p, Options{Release: true})
	if e != nil {
		t.Fatal(e)
	}
	rr, e := Verify(release)
	if e != nil || rr.CertificateSHA256 != r.CertificateSHA256 {
		t.Fatal("release rotates key", e)
	}
	a[64] ^= 1
	if _, e = Verify(a); e == nil {
		t.Fatal("corrupt APK accepted")
	}
}
func TestMissingSigningKeyDoesNotRegenerate(t *testing.T) {
	p := project(t)
	key := filepath.Join(p, ".koru/signing.pem")
	os.Remove(key)
	if _, e := Build(p, Options{}); e == nil {
		t.Fatal("missing key accepted")
	}
	if _, e := os.Stat(key); !os.IsNotExist(e) {
		t.Fatal("key regenerated")
	}
}
func TestAssetsRejectSymlinksAndPrivateKeys(t *testing.T) {
	p := project(t)
	root := filepath.Join(p, "www")
	os.Symlink("/etc/passwd", filepath.Join(root, "leak"))
	if _, e := Assets(root); e == nil {
		t.Fatal("symlink allowed")
	}
	os.Remove(filepath.Join(root, "leak"))
	os.WriteFile(filepath.Join(root, "key.pem"), []byte("-----BEGIN RSA PRIVATE KEY-----"), 0600)
	if _, e := Assets(root); e == nil {
		t.Fatal("key leak allowed")
	}
}

func TestWebOnlyChangeUsesCachedRuntime(t *testing.T) {
	p := project(t)
	first := BuildStats{}
	if _, e := Build(p, Options{Stats: &first}); e != nil {
		t.Fatal(e)
	}
	if first.RuntimeCacheHit {
		t.Fatal("first cache hit")
	}
	os.WriteFile(filepath.Join(p, "www/index.html"), []byte("<head></head><body>changed web content</body>"), 0644)
	second := BuildStats{}
	if _, e := Build(p, Options{Stats: &second}); e != nil {
		t.Fatal(e)
	}
	if !second.RuntimeCacheHit {
		t.Fatal("web content rebuilt runtime")
	}
}
func TestCorruptCacheFailsExplicitly(t *testing.T) {
	p := project(t)
	if _, e := Build(p, Options{}); e != nil {
		t.Fatal(e)
	}
	files, e := filepath.Glob(filepath.Join(p, ".koru/cache/*.dex"))
	if e != nil || len(files) != 1 {
		t.Fatal(files, e)
	}
	os.WriteFile(files[0], []byte("corrupt"), 0600)
	if _, e = Build(p, Options{}); e == nil {
		t.Fatal("corrupt cache silently regenerated")
	}
}

func TestHTMLInjectionOffsetSurvivesNonASCII(t *testing.T) {
	p := project(t)
	root := filepath.Join(p, "www")
	html := "<!doctype html><HEAD><title>İİİİ</title></HEAD><body></body>"
	os.WriteFile(filepath.Join(root, "index.html"), []byte(html), 0644)
	entries, e := Assets(root)
	if e != nil {
		t.Fatal(e)
	}
	want := "<!doctype html><HEAD><title>İİİİ</title><script src=\"/__koru/core.js\"></script></HEAD><body></body>"
	if got := string(entries[0].Data); got != want {
		t.Fatalf("injected HTML = %q", got)
	}
}

func TestAssetsRejectOtherPrivateKeyFormats(t *testing.T) {
	for _, header := range []string{"EC PRIVATE KEY", "ENCRYPTED PRIVATE KEY", "OPENSSH PRIVATE KEY"} {
		p := project(t)
		root := filepath.Join(p, "www")
		os.WriteFile(filepath.Join(root, "key.txt"), []byte("-----BEGIN "+header+"-----"), 0600)
		if _, e := Assets(root); e == nil {
			t.Fatalf("%s leak allowed", header)
		}
	}
}
