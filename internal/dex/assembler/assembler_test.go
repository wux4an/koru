package assembler

import (
	"koru.dev/koru/internal/dex/model"
	"reflect"
	"testing"
)

type resolver struct{}

func (resolver) String(string) (uint32, error)          { return 7, nil }
func (resolver) Type(string) (uint16, error)            { return 3, nil }
func (resolver) Method(model.MethodRef) (uint16, error) { return 5, nil }
func (resolver) Field(model.FieldRef) (uint16, error)   { return 2, nil }
func TestExecutableGolden(t *testing.T) {
	c := model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "const/4", Regs: []model.Register{0}, Literal: 7}, {Op: "return", Regs: []model.Register{0}}}}
	a, e := Assemble(c, "()I", model.Public|model.Static, resolver{})
	if e != nil || !reflect.DeepEqual(a.Units, []uint16{0x7012, 0x000f}) || a.Registers != 1 || a.Incoming != 0 {
		t.Fatal(a, e)
	}
}
func TestParameterAndInvoke(t *testing.T) {
	r := model.MethodRef{Owner: "Ljava/lang/Object;", Name: "<init>", Descriptor: "()V"}
	c := model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "invoke-direct", Regs: []model.Register{model.P(0)}, Method: &r}, {Op: "return-void"}}}
	a, e := Assemble(c, "()V", model.Public|model.Constructor, resolver{})
	if e != nil || a.Incoming != 1 || a.Registers != 2 || a.Outgoing != 1 || !reflect.DeepEqual(a.Units, []uint16{0x1070, 5, 1, 0x000e}) {
		t.Fatal(a, e)
	}
}
func TestBranches(t *testing.T) {
	c := model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "const/4", Regs: []model.Register{0}}, {Op: "if-eqz", Regs: []model.Register{0}, Target: "end"}, {Op: "nop"}, {Op: "label", Label: "end"}, {Op: "return-void"}}}
	a, e := Assemble(c, "()V", model.Static, resolver{})
	if e != nil || !reflect.DeepEqual(a.Units, []uint16{0x0012, 0x0038, 3, 0, 0xe}) {
		t.Fatal(a, e)
	}
}
func TestRejectInvalid(t *testing.T) {
	for _, i := range []model.Instruction{{Op: "unknown"}, {Op: "return", Regs: []model.Register{2}}, {Op: "const/4", Regs: []model.Register{0}, Literal: 8}, {Op: "goto", Target: "missing"}, {Op: "move", Regs: []model.Register{0, model.P(0)}}, {Op: "invoke-direct"}, {Op: "return-object", Regs: []model.Register{0}}} {
		_, e := Assemble(model.Code{Locals: 1, Instructions: []model.Instruction{i}}, "()V", model.Static, resolver{})
		if e == nil {
			t.Fatalf("accepted %+v", i)
		}
	}
}
