package app

import (
	"encoding/json"
	"koru.dev/koru/internal/apk"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/pkg/dex"
	"koru.dev/koru/plugins"
	"os"
	"path/filepath"
	"testing"
)

// TestToastExamplePlugin builds a complete APK with examples/toast/plugin.json and
// checks that the class and handler method reach classes.dex. It does not run ART.
func TestToastExamplePlugin(t *testing.T) {
	manifest, e := filepath.Abs("../../examples/toast/plugin.json")
	if e != nil {
		t.Fatal(e)
	}
	d, e := plugins.Load(manifest)
	if e != nil {
		t.Fatal(e)
	}
	if d.ID != "toast" || len(d.Permissions) != 0 {
		t.Fatalf("unexpected descriptor %q permissions %v", d.ID, d.Permissions)
	}
	p := project(t)
	c := config.Default()
	c.Plugins = append(c.Plugins, "toast")
	c.PluginManifests = []string{manifest}
	data, e := json.Marshal(c)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(p, "koru.json"), data, 0644); e != nil {
		t.Fatal(e)
	}
	built, e := Build(p, Options{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Verify(built); e != nil {
		t.Fatal(e)
	}
	files, _, e := apk.ReadVerified(built)
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(files["classes.dex"])
	if e != nil {
		t.Fatal(e)
	}
	const owner = "Lcom/example/toast/ToastPlugin;"
	ownerType := -1
	for i, s := range f.Types {
		if s == owner {
			ownerType = i
		}
	}
	if ownerType < 0 {
		t.Fatalf("%s missing from classes.dex", owner)
	}
	found := false
	for _, m := range f.Methods {
		if int(m.Owner) == ownerType && f.Strings[m.Name] == "show" {
			found = true
		}
	}
	if !found {
		t.Fatal("ToastPlugin.show missing from classes.dex")
	}
	hasToast := false
	for _, s := range f.Types {
		hasToast = hasToast || s == "Landroid/widget/Toast;"
	}
	if !hasToast {
		t.Fatal("Toast reference missing")
	}
	// The dispatcher must route toast.show to the plugin.
	routed := false
	for _, s := range f.Strings {
		routed = routed || s == "toast.show"
	}
	if !routed {
		t.Fatal("dispatcher does not route toast.show")
	}
}
