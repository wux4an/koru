package templates

import (
	"encoding/json"
	"koru.dev/koru/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNames(t *testing.T) {
	for _, c := range []struct{ in, slug, id, display string }{
		{"My App", "my-app", "com.example.myapp", "My App"},
		{"  Café <Ñ> 2 ", "caf-2", "com.example.caf2", "Café <Ñ> 2"},
		{"9 lives", "9-lives", "com.example.app9lives", "9 Lives"},
		{"***", "app", "com.example.app", "***"},
	} {
		if got := Slug(c.in); got != c.slug {
			t.Errorf("Slug(%q) = %q", c.in, got)
		}
		if got := AppID(c.in); got != c.id {
			t.Errorf("AppID(%q) = %q", c.in, got)
		}
		cfg := config.Default()
		cfg.AppID = AppID(c.in)
		if e := cfg.Validate(); e != nil {
			t.Errorf("AppID(%q) invalid: %v", c.in, e)
		}
	}
	if got := DisplayName("/tmp/my-cool_app"); got != "My Cool App" {
		t.Fatal(got)
	}
}

func TestWriteEveryTemplate(t *testing.T) {
	for _, tpl := range List {
		dir := filepath.Join(t.TempDir(), "app")
		c, err := Write(dir, Options{Template: tpl.ID, Name: `Tom & "Jerry"`, AppID: "com.example.tom"})
		if err != nil {
			t.Fatal(tpl.ID, err)
		}
		loaded, err := config.Load(dir)
		if err != nil {
			t.Fatal(tpl.ID, err)
		}
		if loaded.Name != c.Name || loaded.Schema != config.SchemaURL || loaded.WebDir != tpl.WebDir {
			t.Fatalf("%s: %+v", tpl.ID, loaded)
		}
		ignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
		if err != nil || !strings.Contains(string(ignore), ".koru/") {
			t.Fatalf("%s .gitignore: %q %v", tpl.ID, ignore, err)
		}
		index, _ := os.ReadFile(filepath.Join(dir, "index.html"))
		if tpl.ID == "basic" {
			index, _ = os.ReadFile(filepath.Join(dir, "www/index.html"))
		}
		if !strings.Contains(string(index), "<title>Tom &amp; &#34;Jerry&#34;</title>") || strings.Contains(string(index), "__APP_") {
			t.Fatalf("%s index.html placeholders: %s", tpl.ID, index)
		}
		if tpl.ID == "vite" {
			var pkg map[string]any
			b, _ := os.ReadFile(filepath.Join(dir, "package.json"))
			if err = json.Unmarshal(b, &pkg); err != nil || pkg["name"] != "tom-jerry" {
				t.Fatalf("package.json: %v %v", pkg, err)
			}
			if _, err = os.Stat(filepath.Join(dir, "src/koru/core.d.mts")); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestWriteRefusesNonEmptyDirectoryAndBadInput(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0644)
	if _, err := Write(dir, Options{Template: "basic", Name: "A", AppID: "com.example.a"}); err == nil {
		t.Fatal("non-empty directory accepted")
	}
	if _, err := Write(t.TempDir(), Options{Template: "nope", Name: "A", AppID: "com.example.a"}); err == nil || !strings.Contains(err.Error(), "basic, vite") {
		t.Fatalf("template error = %v", err)
	}
	if _, err := Write(t.TempDir(), Options{Template: "basic", Name: "A", AppID: "bad"}); err == nil {
		t.Fatal("bad appId accepted")
	}
}
