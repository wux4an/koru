package axml

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

func baseManifest() Manifest {
	return Manifest{Package: "com.example.koru", Name: "Koru", Version: "1.0", VersionCode: 1, MinSDK: 26, TargetSDK: 36, Activity: "com.koru.runtime.MainActivity", Icon: 0x7f010000, Permissions: []string{"android.permission.VIBRATE"}}
}

func featureManifest() Manifest {
	m := baseManifest()
	m.Orientation, m.Keyboard, m.AllowBackup = "landscape", "pan", true
	m.DeepLinks = []DeepLink{{"https", "example.com", "/app"}, {"https", "*.example.org", ""}, {"myapp", "", ""}, {"myapp", "open", "/x"}}
	return m
}

// render prints the whole parsed tree, one element per line, so a golden diff
// shows exactly which attribute changed.
func render(els []Element) string {
	var b strings.Builder
	for _, el := range els {
		b.WriteString(el.Name)
		for _, a := range el.Attributes {
			ns := ""
			if a.Namespace != "" {
				ns = "android:"
			}
			switch a.Kind {
			case 3:
				fmt.Fprintf(&b, " %s%s=%q", ns, a.Name, a.Text)
			case 0x12:
				fmt.Fprintf(&b, " %s%s=%t", ns, a.Name, a.Data != 0)
			default:
				fmt.Fprintf(&b, " %s%s=#%x:0x%x", ns, a.Name, a.Kind, a.Data)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := "testdata/" + name
	if *update {
		if e := os.WriteFile(path, got, 0644); e != nil {
			t.Fatal(e)
		}
	}
	want, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s differs from the golden file (run go test ./internal/axml -update after reviewing)", name)
	}
}

// resourceMap returns the attribute IDs of the XML resource map chunk.
func resourceMap(t *testing.T, b []byte) []uint32 {
	t.Helper()
	for pos := 8; pos+8 <= len(b); {
		size := int(binary.LittleEndian.Uint32(b[pos+4:]))
		if binary.LittleEndian.Uint16(b[pos:]) == 0x180 {
			var out []uint32
			for p := pos + 8; p < pos+size; p += 4 {
				out = append(out, binary.LittleEndian.Uint32(b[p:]))
			}
			return out
		}
		pos += size
	}
	t.Fatal("no resource map")
	return nil
}

func TestDefaultManifestUnchanged(t *testing.T) {
	b, e := Build(baseManifest())
	if e != nil {
		t.Fatal(e)
	}
	// Generated before the configurable fields existed: omitting them must not
	// change a single byte.
	golden(t, "default.golden", b)
}

func TestFeatureManifestGolden(t *testing.T) {
	b, e := Build(featureManifest())
	if e != nil {
		t.Fatal(e)
	}
	els, e := Parse(b)
	if e != nil {
		t.Fatal(e)
	}
	golden(t, "features.golden", b)
	golden(t, "features.tree", []byte(render(els)))
	if !slices.IsSorted(resourceMap(t, b)) {
		t.Fatal("resource map must be sorted by attribute ID")
	}
	for _, id := range []uint32{0x0101001e, 0x01010027, 0x01010028, 0x0101002b, 0x010104ee, 0x0101022b, 0x01010280} {
		if !slices.Contains(resourceMap(t, b), id) {
			t.Errorf("resource map lacks 0x%08x", id)
		}
	}
	want := []Link{{"https", "example.com", "/app", true}, {"https", "*.example.org", "", true}, {"myapp", "", "", false}, {"myapp", "open", "/x", false}}
	if got := Links(els); !reflect.DeepEqual(got, want) {
		t.Fatalf("links = %+v", got)
	}
}

func attrOf(els []Element, element, name string) (Value, bool) {
	for _, el := range els {
		if el.Name != element {
			continue
		}
		for _, a := range el.Attributes {
			if a.Name == name {
				return a, true
			}
		}
	}
	return Value{}, false
}

func TestOrientationKeyboardAndBackupValues(t *testing.T) {
	for orientation, want := range map[string]uint32{"portrait": 1, "landscape": 0, "sensor": 4} {
		m := baseManifest()
		m.Orientation = orientation
		b, e := Build(m)
		if e != nil {
			t.Fatal(e)
		}
		els, _ := Parse(b)
		if a, ok := attrOf(els, "activity", "screenOrientation"); !ok || a.Kind != 0x10 || a.Data != want {
			t.Errorf("%s: %+v %v", orientation, a, ok)
		}
	}
	b, _ := Build(baseManifest())
	els, _ := Parse(b)
	if _, ok := attrOf(els, "activity", "screenOrientation"); ok {
		t.Error("orientation must be omitted by default")
	}
	if a, _ := attrOf(els, "activity", "windowSoftInputMode"); a.Data != 0x10 {
		t.Errorf("default keyboard = 0x%x", a.Data)
	}
	if a, _ := attrOf(els, "application", "allowBackup"); a.Data != 0 {
		t.Error("allowBackup must default to false")
	}
	if len(Links(els)) != 0 {
		t.Error("no deep links by default")
	}
	for keyboard, want := range map[string]uint32{"resize": 0x10, "pan": 0x20} {
		m := baseManifest()
		m.Keyboard, m.AllowBackup = keyboard, true
		b, _ := Build(m)
		els, _ := Parse(b)
		if a, _ := attrOf(els, "activity", "windowSoftInputMode"); a.Data != want {
			t.Errorf("%s = 0x%x", keyboard, a.Data)
		}
		if a, _ := attrOf(els, "application", "allowBackup"); a.Data == 0 {
			t.Error("allowBackup true")
		}
	}
}

func TestInvalidFeatureInputs(t *testing.T) {
	for name, edit := range map[string]func(*Manifest){
		"orientation": func(m *Manifest) { m.Orientation = "diagonal" },
		"keyboard":    func(m *Manifest) { m.Keyboard = "hide" },
		"scheme":      func(m *Manifest) { m.DeepLinks = []DeepLink{{Host: "example.com"}} },
	} {
		m := baseManifest()
		edit(&m)
		if _, e := Build(m); e == nil {
			t.Errorf("%s accepted", name)
		}
	}
	b, _ := Build(featureManifest())
	for i := 0; i < len(b); i++ {
		if _, e := Parse(b[:i]); e == nil {
			t.Fatalf("truncated at %d accepted", i)
		}
	}
	// A string index past the pool must be rejected, not read.
	bad := append([]byte(nil), b...)
	for i := len(bad) - 40; i+4 <= len(bad); i += 4 {
		binary.LittleEndian.PutUint32(bad[i:], 0xfffffff0)
	}
	if _, e := Parse(bad); e == nil {
		t.Error("corrupt string indexes accepted")
	}
}
