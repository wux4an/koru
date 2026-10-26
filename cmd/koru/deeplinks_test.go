package main

import (
	"bytes"
	"encoding/json"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/internal/config"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func deepLinkProject(t *testing.T, links []config.DeepLink) string {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	if e := run([]string{"init", dir}, &out); e != nil {
		t.Fatal(e)
	}
	c, e := config.Load(dir)
	if e != nil {
		t.Fatal(e)
	}
	c.AppID, c.DeepLinks = "com.example.links", links
	b, _ := json.Marshal(c)
	if e = os.WriteFile(filepath.Join(dir, "koru.json"), b, 0644); e != nil {
		t.Fatal(e)
	}
	return dir
}

func TestDeepLinksPrintsAssetLinksForBuiltAPK(t *testing.T) {
	dir := deepLinkProject(t, []config.DeepLink{{Scheme: "https", Host: "example.com", PathPrefix: "/app"}, {Scheme: "https", Host: "*.example.org"}, {Scheme: "myapp"}})
	apk := filepath.Join(dir, "out.apk")
	var out bytes.Buffer
	if e := run([]string{"build", dir, "--output", apk, "--release"}, &out); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(apk)
	if e != nil {
		t.Fatal(e)
	}
	report, e := app.Verify(b)
	if e != nil {
		t.Fatal(e)
	}
	out.Reset()
	if e = run([]string{"deeplinks", dir, "--json"}, &out); e != nil {
		t.Fatal(e)
	}
	var got []assetStatement
	if e = json.Unmarshal(out.Bytes(), &got); e != nil {
		t.Fatalf("%v: %s", e, out.String())
	}
	if len(got) != 1 || got[0].Relation[0] != "delegate_permission/common.handle_all_urls" || got[0].Target.Namespace != "android_app" || got[0].Target.PackageName != "com.example.links" {
		t.Fatalf("statement = %+v", got)
	}
	fp := got[0].Target.Fingerprints
	if len(fp) != 1 || !regexp.MustCompile(`^([0-9A-F]{2}:){31}[0-9A-F]{2}$`).MatchString(fp[0]) {
		t.Fatalf("fingerprint format = %v", fp)
	}
	// Independent check: the fingerprint must be the certificate hash that
	// verify reports for the signed APK.
	if strings.ToLower(strings.ReplaceAll(fp[0], ":", "")) != report.CertificateSHA256 {
		t.Fatalf("fingerprint %s != APK certificate %s", fp[0], report.CertificateSHA256)
	}
	if len(report.DeepLinks) != 3 {
		t.Fatalf("verify deep links = %v", report.DeepLinks)
	}

	out.Reset()
	if e = run([]string{"deeplinks", dir}, &out); e != nil {
		t.Fatal(e)
	}
	for _, want := range []string{"/.well-known/assetlinks.json", "example.com, example.org", fp[0], "initial page"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestDeepLinksErrors(t *testing.T) {
	var out bytes.Buffer
	if e := run([]string{"deeplinks", deepLinkProject(t, []config.DeepLink{{Scheme: "myapp"}})}, &out); e == nil || !strings.Contains(e.Error(), "no https deepLinks") {
		t.Fatalf("custom scheme only: %v", e)
	}
	if e := run([]string{"deeplinks", deepLinkProject(t, nil)}, &out); e == nil || !strings.Contains(e.Error(), "no https deepLinks") {
		t.Fatalf("no links: %v", e)
	}
	dir := deepLinkProject(t, []config.DeepLink{{Scheme: "https", Host: "example.com"}})
	os.Remove(filepath.Join(dir, ".koru/signing.pem"))
	if e := run([]string{"deeplinks", dir}, &out); e == nil || !strings.Contains(e.Error(), "signing identity") {
		t.Fatalf("missing key: %v", e)
	}
	if e := run([]string{"deeplinks", t.TempDir()}, &out); e == nil || !strings.Contains(e.Error(), "koru create") {
		t.Fatalf("missing project: %v", e)
	}
}

func TestCertFingerprintFormat(t *testing.T) {
	if got := certFingerprint([]byte("abc")); got != "BA:78:16:BF:8F:01:CF:EA:41:41:40:DE:5D:AE:22:23:B0:03:61:A3:96:17:7A:9C:B4:10:FF:61:F2:00:15:AD" {
		t.Fatal(got)
	}
}
