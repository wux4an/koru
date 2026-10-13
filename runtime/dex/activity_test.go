package runtimedex

import (
	"bytes"
	"koru.dev/koru/pkg/dex"
	"testing"
)

func TestActivityStructure(t *testing.T) {
	b, e := dex.Build(ActivityExperiment())
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(b)
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Classes) != 1 || f.Types[f.Classes[0].Type] != ActivityDescriptor || f.Types[f.Classes[0].Super] != "Landroid/app/Activity;" || len(f.Classes[0].Methods) != 6 {
		t.Fatal(f.Classes)
	}
	found := false
	for _, m := range f.Classes[0].Methods {
		if f.Strings[f.Methods[m.Index].Name] == "onCreate" {
			found = true
			if m.Code.Incoming != 2 || m.Code.Registers != 6 || byte(m.Code.Units[0]) != 0x6f {
				t.Fatal(m)
			}
		}
	}
	if !found {
		t.Fatal("onCreate missing")
	}
	b2, e := dex.Build(ActivityExperiment())
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal("nondeterminism", e)
	}
}
func BenchmarkActivity(b *testing.B) {
	p := ActivityExperiment()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := dex.Build(p); e != nil {
			b.Fatal(e)
		}
	}
}
