//go:build android_external

// Package external_test runs optional official Android validation, never builds.
package external_test

import (
	"koru.dev/koru/pkg/dex"
	runtimedex "koru.dev/koru/runtime/dex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAOSPDEXDump(t *testing.T) {
	tool := os.Getenv("KORU_TEST_DEXDUMP")
	if tool == "" {
		t.Fatal("set KORU_TEST_DEXDUMP to the official dexdump executable")
	}
	simple := dex.Program{Classes: []dex.Class{{Descriptor: "Lexample/Answer;", Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{{Name: "answer", Descriptor: "()I", Access: 9, Code: dex.NewCode(1).Const(dex.V(0), 42).Emit(dex.Instruction{Op: "return", Regs: []dex.Register{0}}).Build()}}}}}
	for _, c := range []struct {
		name string
		p    dex.Program
	}{{"answer", simple}, {"activity", runtimedex.ActivityExperiment()}} {
		t.Run(c.name, func(t *testing.T) {
			b, e := dex.Build(c.p)
			if e != nil {
				t.Fatal(e)
			}
			path := filepath.Join(t.TempDir(), c.name+".dex")
			if e = os.WriteFile(path, b, 0644); e != nil {
				t.Fatal(e)
			}
			for _, args := range [][]string{{"-c", path}, {"-d", "-f", path}} {
				out, e := exec.Command(tool, args...).CombinedOutput()
				if e != nil {
					t.Fatalf("dexdump: %v\n%s", e, out)
				}
			}
		})
	}
}
