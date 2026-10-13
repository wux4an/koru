package axml

import (
	"bytes"
	"testing"
)

func TestManifest(t *testing.T) {
	m := Manifest{Package: "com.example.koru", Name: "Koru 中 😀", Version: "1.0", VersionCode: 1, MinSDK: 26, TargetSDK: 36, Activity: "com.koru.runtime.MainActivity", Icon: 0x7f010000}
	b, e := Build(m)
	if e != nil {
		t.Fatal(e)
	}
	els, e := Parse(b)
	if e != nil {
		t.Fatal(e)
	}
	if els[0].Name != "manifest" || els[0].Attributes[0].Text != m.Package {
		t.Fatal(els)
	}
	found := false
	for _, el := range els {
		for _, a := range el.Attributes {
			if a.Name == "label" && a.Text == m.Name {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("label")
	}
	b2, e := Build(m)
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal(e)
	}
	for i := 0; i < len(b); i++ {
		if _, e := Parse(b[:i]); e == nil {
			t.Fatal(i)
		}
	}
}
func FuzzXML(f *testing.F) {
	b, _ := Build(Manifest{Package: "com.example.demo", Name: "Demo", VersionCode: 1, MinSDK: 26, TargetSDK: 36, Activity: "com.koru.runtime.MainActivity"})
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) { Parse(b) })
}
