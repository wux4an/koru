package main

import (
	"bytes"
	"koru.dev/koru/packages/javascript"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSWritesModuleAndRefusesOverwrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "src/koru")
	var out bytes.Buffer
	if e := run([]string{"js", dir}, &out); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "core.d.mts"))
	if !bytes.Equal(b, javascript.Types) {
		t.Fatal("types differ")
	}
	os.WriteFile(filepath.Join(dir, "core.mjs"), []byte("old"), 0644)
	if e := run([]string{"js", dir}, &out); e == nil || !strings.Contains(e.Error(), "--force") {
		t.Fatalf("overwrite error = %v", e)
	}
	if e := run([]string{"js", dir, "--force"}, &out); e != nil {
		t.Fatal(e)
	}
	if b, _ = os.ReadFile(filepath.Join(dir, "core.mjs")); !bytes.Equal(b, javascript.Module) {
		t.Fatal("module not replaced")
	}
}
