package dex_test

import (
	"koru.dev/koru/pkg/dex"
	"testing"
)

func TestExceptionSectionsAndArrays(t *testing.T) {
	c := dex.NewCode(4).Label("try").Const(dex.V(0), 1).Emit(dex.Instruction{Op: "new-array", Regs: []dex.Register{dex.V(1), dex.V(0)}, Type: "[Ljava/lang/String;"}).ConstString(dex.V(2), "hello").Const(dex.V(0), 0).Emit(dex.Instruction{Op: "aput-object", Regs: []dex.Register{dex.V(2), dex.V(1), dex.V(0)}}).Emit(dex.Instruction{Op: "aget-object", Regs: []dex.Register{dex.V(3), dex.V(1), dex.V(0)}}).ReturnVoid().Label("end").Label("catch").Emit(dex.Instruction{Op: "move-exception", Regs: []dex.Register{dex.V(0)}}).ReturnVoid().Catch("try", "end", "catch", "Ljava/lang/Exception;")
	p := dex.Program{Classes: []dex.Class{{Descriptor: "Ltest/Exceptions;", Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{{Name: "f", Descriptor: "()V", Access: 9, Code: c.Build()}}}}}
	b, e := dex.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(b)
	if e != nil || len(f.Classes[0].Methods[0].Code.Tries) != 1 {
		t.Fatal(e)
	}
	p.Classes[0].Methods[0].Code.Tries[0].Handler = "try"
	if _, e = dex.Build(p); e == nil {
		t.Fatal("handler without move-exception")
	}
}
