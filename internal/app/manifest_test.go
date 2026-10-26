package app

import (
	"encoding/json"
	"koru.dev/koru/internal/apk"
	"koru.dev/koru/internal/axml"
	"koru.dev/koru/internal/config"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildEmitsConfigurableManifest(t *testing.T) {
	p := project(t)
	c := config.Default()
	c.Orientation, c.Keyboard, c.AllowBackup = "portrait", "pan", true
	c.DeepLinks = []config.DeepLink{{Scheme: "https", Host: "example.com", PathPrefix: "/app"}, {Scheme: "myapp"}}
	data, _ := json.Marshal(c)
	os.WriteFile(filepath.Join(p, "koru.json"), data, 0644)
	b, e := Build(p, Options{Release: true})
	if e != nil {
		t.Fatal(e)
	}
	r, e := Verify(b)
	if e != nil {
		t.Fatal(e)
	}
	if want := []string{"https://example.com/app", "myapp://"}; !reflect.DeepEqual(r.DeepLinks, want) {
		t.Fatalf("deep links = %v", r.DeepLinks)
	}
	files, _, e := apk.ReadVerified(b)
	if e != nil {
		t.Fatal(e)
	}
	els, e := axml.Parse(files["AndroidManifest.xml"])
	if e != nil {
		t.Fatal(e)
	}
	got := map[string]uint32{}
	for _, el := range els {
		for _, a := range el.Attributes {
			switch a.Name {
			case "screenOrientation", "windowSoftInputMode", "allowBackup":
				got[a.Name] = a.Data
			}
		}
	}
	if want := (map[string]uint32{"screenOrientation": 1, "windowSoftInputMode": 0x20, "allowBackup": 0xffffffff}); !reflect.DeepEqual(got, want) {
		t.Fatalf("attributes = %v", got)
	}
}

func TestVerifyReportsNoDeepLinksByDefault(t *testing.T) {
	b, e := Build(project(t), Options{})
	if e != nil {
		t.Fatal(e)
	}
	if r, e := Verify(b); e != nil || len(r.DeepLinks) != 0 {
		t.Fatal(r, e)
	}
}
