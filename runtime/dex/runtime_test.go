package runtimedex

import (
	"koru.dev/koru/internal/config"
	"koru.dev/koru/pkg/dex"
	"testing"
)

func TestControlledOriginRuntimeAllPlugins(t *testing.T) {
	c := config.Default()
	c.Plugins = []string{"device", "storage", "vibration", "app", "share", "notifications"}
	c.SPA = true
	p := Runtime(c, []Asset{{Path: "/index.html", MIME: "text/html"}, {Path: "/__koru/core.js", MIME: "text/javascript"}})
	b, e := dex.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(b)
	if e != nil || len(f.Classes) != 5 {
		t.Fatal(e)
	}
	for _, s := range f.Strings {
		if s == "file:///android_asset/www/index.html" || s == "addJavascriptInterface" {
			t.Fatal("unsafe bridge", s)
		}
	}
}
