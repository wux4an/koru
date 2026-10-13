package dex_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"koru.dev/koru/pkg/dex"
	"testing"
)

func TestCrossClassesFieldsPrototypesAndUnicode(t *testing.T) {
	parent := dex.Class{Descriptor: "Lz/Parent;", Super: "Ljava/lang/Object;", Access: dex.Public, Methods: []dex.Method{{Name: "<init>", Descriptor: "()V", Access: dex.Public | dex.Constructor, Code: dex.NewCode(0).Invoke("invoke-direct", dex.MethodRef{Owner: "Ljava/lang/Object;", Name: "<init>", Descriptor: "()V"}, dex.P(0)).ReturnVoid().Build()}}}
	child := dex.Class{Descriptor: "La/Child;", Super: parent.Descriptor, Access: dex.Public, Fields: []dex.Field{{Name: "value", Type: "I", Access: dex.Public}, {Name: "message", Type: "Ljava/lang/String;", Access: dex.Static}}, Methods: []dex.Method{{Name: "<init>", Descriptor: "()V", Access: dex.Public | dex.Constructor, Code: dex.NewCode(0).Invoke("invoke-direct", dex.MethodRef{Owner: parent.Descriptor, Name: "<init>", Descriptor: "()V"}, dex.P(0)).ReturnVoid().Build()}, {Name: "text", Descriptor: "()Ljava/lang/String;", Access: dex.Public | dex.Static, Code: dex.NewCode(1).ConstString(dex.V(0), "😀").ConstString(dex.V(0), "\ue000").ConstString(dex.V(0), "á\x00中").Emit(dex.Instruction{Op: "return-object", Regs: []dex.Register{dex.V(0)}}).Build()}}}
	iface := dex.Class{Descriptor: "Li/Readable;", Super: "Ljava/lang/Object;", Access: dex.Public | dex.Interface | dex.Abstract, Methods: []dex.Method{{Name: "get", Descriptor: "([I)I", Access: dex.Public | dex.Abstract}}}
	child.Interfaces = []string{iface.Descriptor}
	child.Access |= dex.Abstract
	p := dex.Program{Classes: []dex.Class{child, parent, iface}}
	b, e := dex.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(b)
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Classes) != 3 || f.Types[f.Classes[2].Type] != child.Descriptor {
		t.Fatal(f.Classes)
	}
	emoji, bmp := -1, -1
	for i, s := range f.Strings {
		if s == "😀" {
			emoji = i
		}
		if s == "\ue000" {
			bmp = i
		}
	}
	if emoji < 0 || emoji >= bmp {
		t.Fatal("UTF-16 ordering", emoji, bmp)
	}
	p.Classes = []dex.Class{parent, iface, child}
	p.Classes[2].Methods[0], p.Classes[2].Methods[1] = p.Classes[2].Methods[1], p.Classes[2].Methods[0]
	b2, e := dex.Build(p)
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal("input declaration order changes output", e)
	}
}
func TestBuildRejectsBytecodeSemantics(t *testing.T) {
	for _, code := range []*dex.Code{dex.NewCode(1).Const(dex.V(0), 1).Build(), dex.NewCode(1).Emit(dex.Instruction{Op: "move-result", Regs: []dex.Register{0}}).ReturnVoid().Build()} {
		_, e := dex.Build(dex.Program{Classes: []dex.Class{{Descriptor: "Ltest/A;", Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{{Name: "f", Descriptor: "()V", Access: 9, Code: code}}}}})
		if e == nil {
			t.Fatal("accepted invalid bytecode")
		}
	}
}
func TestJSONRoundtrip(t *testing.T) {
	p := dex.Program{Classes: []dex.Class{{Descriptor: "Ltest/A;", Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{{Name: "f", Descriptor: "()V", Access: 9, Code: dex.NewCode(0).ReturnVoid().Build()}}}}}
	j, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	var p2 dex.Program
	if e = json.Unmarshal(j, &p2); e != nil {
		t.Fatal(e)
	}
	b, e := dex.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	b2, e := dex.Build(p2)
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal(e)
	}
}

func TestCodeBuilderOwnsInstructionSnapshots(t *testing.T) {
	regs := []dex.Register{0}
	m := dex.MethodRef{Owner: "Ltest/A;", Name: "f", Descriptor: "()V"}
	field := dex.FieldRef{Owner: "Ltest/A;", Name: "x", Type: "I"}
	b := dex.NewCode(1).Emit(dex.Instruction{Op: "nop", Regs: regs, Method: &m, Field: &field})
	regs[0] = 12
	m.Name = "changed"
	field.Name = "changed"
	first := b.Build()
	if first.Instructions[0].Regs[0] != 0 || first.Instructions[0].Method.Name != "f" || first.Instructions[0].Field.Name != "x" {
		t.Fatal("Emit retains caller-owned mutable state")
	}
	first.Instructions[0].Regs[0] = 13
	first.Instructions[0].Method.Name = "changed again"
	second := b.Build()
	if second.Instructions[0].Regs[0] != 0 || second.Instructions[0].Method.Name != "f" {
		t.Fatal("Build aliases previous snapshot")
	}
}
func TestRejectWrongInvocationKind(t *testing.T) {
	owner := "Ltest/A;"
	p := dex.Program{Classes: []dex.Class{{Descriptor: owner, Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{
		{Name: "f", Descriptor: "()V", Access: dex.Public, Code: dex.NewCode(0).ReturnVoid().Build()},
		{Name: "caller", Descriptor: "()V", Access: dex.Public, Code: dex.NewCode(0).Invoke("invoke-direct", dex.MethodRef{Owner: owner, Name: "f", Descriptor: "()V"}, dex.P(0)).ReturnVoid().Build()},
	}}}}
	if _, e := dex.Build(p); e == nil {
		t.Fatal("invoke-direct cannot target ordinary virtual method")
	}
}

func TestFieldsAndInvokeReferences(t *testing.T) {
	owner := "Ltest/Fields;"
	instance := dex.FieldRef{Owner: owner, Name: "value", Type: "I"}
	static := dex.FieldRef{Owner: owner, Name: "shared", Type: "I"}
	b := dex.NewCode(2).Const(dex.V(0), 42).Emit(dex.Instruction{Op: "iput", Regs: []dex.Register{dex.V(0), dex.P(0)}, Field: &instance}).Emit(dex.Instruction{Op: "iget", Regs: []dex.Register{dex.V(0), dex.P(0)}, Field: &instance}).Emit(dex.Instruction{Op: "sput", Regs: []dex.Register{dex.V(0)}, Field: &static}).Emit(dex.Instruction{Op: "sget", Regs: []dex.Register{dex.V(1)}, Field: &static}).Emit(dex.Instruction{Op: "return", Regs: []dex.Register{dex.V(1)}})
	p := dex.Program{Classes: []dex.Class{{Descriptor: owner, Super: "Ljava/lang/Object;", Access: 1, Fields: []dex.Field{{Name: instance.Name, Type: "I", Access: 1}, {Name: static.Name, Type: "I", Access: dex.Public | dex.Static}}, Methods: []dex.Method{{Name: "fields", Descriptor: "()I", Access: dex.Public, Code: b.Build()}}}}}
	out, e := dex.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	f, e := dex.Inspect(out)
	if e != nil || len(f.Fields) != 2 || len(f.Classes[0].Fields) != 2 {
		t.Fatal(f, e)
	}
	p.Classes[0].Fields[0].Access |= dex.Static
	if _, e = dex.Build(p); e == nil {
		t.Fatal("instance access of static field accepted")
	}
}
func TestIdentifierLimit(t *testing.T) {
	c := dex.NewCode(1)
	for i := 0; i < 65536; i++ {
		c.ConstString(dex.V(0), "x").Emit(dex.Instruction{Op: "const-class", Regs: []dex.Register{dex.V(0)}, Type: fmt.Sprintf("Llimit/T%05d;", i)})
	}
	c.ReturnVoid()
	p := dex.Program{Classes: []dex.Class{{Descriptor: "Ltest/A;", Super: "Ljava/lang/Object;", Access: 1, Methods: []dex.Method{{Name: "f", Descriptor: "()V", Access: 9, Code: c.Build()}}}}}
	if _, e := dex.Build(p); e == nil {
		t.Fatal("more than 65535 types accepted")
	}
}
