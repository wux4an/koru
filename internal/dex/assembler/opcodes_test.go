package assembler

import (
	"koru.dev/koru/internal/dex/model"
	"reflect"
	"testing"
)

func TestOpcodeGoldenWords(t *testing.T) {
	method := &model.MethodRef{Owner: "Ltest/A;", Name: "run", Descriptor: "(I)V"}
	field := &model.FieldRef{Owner: "Ltest/A;", Name: "value", Type: "I"}
	cases := []struct {
		op      string
		regs    []model.Register
		literal int32
		words   []uint16
		desc    string
	}{
		{"nop", nil, 0, []uint16{0}, "()V"},
		{"move", []model.Register{1, 2}, 0, []uint16{0x2101}, "()V"},
		{"move-object", []model.Register{1, 2}, 0, []uint16{0x2107}, "()V"},
		{"move-result", []model.Register{1}, 0, []uint16{0x010a}, "()V"},
		{"move-result-object", []model.Register{1}, 0, []uint16{0x010c}, "()V"},
		{"return-void", nil, 0, []uint16{0xe}, "()V"},
		{"return", []model.Register{1}, 0, []uint16{0x010f}, "()I"},
		{"return-object", []model.Register{1}, 0, []uint16{0x0111}, "()Ljava/lang/Object;"},
		{"const/4", []model.Register{1}, -8, []uint16{0x8112}, "()V"},
		{"const/16", []model.Register{1}, -32768, []uint16{0x0113, 0x8000}, "()V"},
		{"const", []model.Register{1}, -1, []uint16{0x0114, 0xffff, 0xffff}, "()V"},
		{"const-string", []model.Register{1}, 0, []uint16{0x011a, 7}, "()V"},
		{"const-class", []model.Register{1}, 0, []uint16{0x011c, 3}, "()V"},
		{"new-instance", []model.Register{1}, 0, []uint16{0x0122, 3}, "()V"},
		{"check-cast", []model.Register{1}, 0, []uint16{0x011f, 3}, "()V"},
		{"invoke-virtual", []model.Register{1, 2}, 0, []uint16{0x206e, 5, 0x21}, "()V"},
		{"invoke-super", []model.Register{1, 2}, 0, []uint16{0x206f, 5, 0x21}, "()V"},
		{"invoke-direct", []model.Register{1, 2}, 0, []uint16{0x2070, 5, 0x21}, "()V"},
		{"invoke-static", []model.Register{1}, 0, []uint16{0x1071, 5, 1}, "()V"},
		{"iget", []model.Register{1, 2}, 0, []uint16{0x2152, 2}, "()V"},
		{"iput", []model.Register{1, 2}, 0, []uint16{0x2159, 2}, "()V"},
		{"sget", []model.Register{1}, 0, []uint16{0x0160, 2}, "()V"},
		{"sput", []model.Register{1}, 0, []uint16{0x0167, 2}, "()V"},
	}
	for _, c := range cases {
		t.Run(c.op, func(t *testing.T) {
			i := model.Instruction{Op: c.op, Regs: c.regs, Literal: c.literal, Type: "Ljava/lang/Object;", Method: method, Field: field, String: "x"}
			a, e := Assemble(model.Code{Locals: 3, Instructions: []model.Instruction{i}}, c.desc, model.Static, resolver{})
			if e != nil || !reflect.DeepEqual(a.Units, c.words) {
				t.Fatalf("got %x want %x: %v", a.Units, c.words, e)
			}
		})
	}
}
func TestExtendedRegistersAndInvokeRange(t *testing.T) {
	for _, c := range []struct {
		i     model.Instruction
		words []uint16
	}{
		{model.Instruction{Op: "move", Regs: []model.Register{1, 256}}, []uint16{0x0102, 256}},
		{model.Instruction{Op: "move", Regs: []model.Register{256, 299}}, []uint16{3, 256, 299}},
		{model.Instruction{Op: "move-object", Regs: []model.Register{1, 256}}, []uint16{0x0108, 256}},
		{model.Instruction{Op: "move-object", Regs: []model.Register{256, 299}}, []uint16{9, 256, 299}},
		{model.Instruction{Op: "invoke-virtual", Regs: []model.Register{16, 17}, Method: &model.MethodRef{Owner: "Ltest/A;", Name: "f", Descriptor: "(I)V"}}, []uint16{0x0274, 5, 16}},
	} {
		a, e := Assemble(model.Code{Locals: 300, Instructions: []model.Instruction{c.i}}, "()V", model.Static, resolver{})
		if e != nil || !reflect.DeepEqual(a.Units, c.words) {
			t.Fatal(a, e, c.words)
		}
	}
}
func TestFieldVariants(t *testing.T) {
	for _, c := range []struct {
		typ   string
		delta uint16
	}{{"I", 0}, {"F", 0}, {"Ljava/lang/Object;", 2}, {"[I", 2}, {"Z", 3}, {"B", 4}, {"C", 5}, {"S", 6}} {
		i := model.Instruction{Op: "iget", Regs: []model.Register{1, 2}, Field: &model.FieldRef{Owner: "Ltest/A;", Name: "x", Type: c.typ}}
		a, e := Assemble(model.Code{Locals: 3, Instructions: []model.Instruction{i}}, "()V", model.Static, resolver{})
		if e != nil || a.Units[0] != 0x2152+c.delta {
			t.Fatal(c, a, e)
		}
	}
}
func TestWideIncomingWords(t *testing.T) {
	a, e := Assemble(model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "move", Regs: []model.Register{0, model.P(4)}}, {Op: "return-void"}}}, "(JD)V", model.Public, resolver{})
	if e != nil || a.Incoming != 5 || a.Registers != 6 || a.Units[0] != 0x5001 {
		t.Fatal(a, e)
	}
}
func TestRejectLimitsAndLabels(t *testing.T) {
	cases := []model.Code{
		{Locals: 65535, Instructions: []model.Instruction{{Op: "return-void"}}},
		{Locals: 1, Instructions: []model.Instruction{{Op: "label", Label: "a"}, {Op: "label", Label: "a"}, {Op: "return-void"}}},
		{Locals: 1, Instructions: []model.Instruction{{Op: "goto", Target: "end"}, {Op: "label", Label: "end"}}},
		{Locals: 20, Instructions: []model.Instruction{{Op: "invoke-virtual", Regs: []model.Register{16, 18}, Method: &model.MethodRef{Owner: "Ltest/A;", Name: "f", Descriptor: "(I)V"}}}},
	}
	for _, c := range cases {
		if _, e := Assemble(c, "()V", model.Public, resolver{}); e == nil {
			t.Fatal(c)
		}
	}
	c := model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "if-nez", Regs: []model.Register{0}, Target: "far"}}}
	for i := 0; i < 32768; i++ {
		c.Instructions = append(c.Instructions, model.Instruction{Op: "nop"})
	}
	c.Instructions = append(c.Instructions, model.Instruction{Op: "label", Label: "far"}, model.Instruction{Op: "return-void"})
	if _, e := Assemble(c, "()V", model.Static, resolver{}); e == nil {
		t.Fatal("accepted conditional branch overflow")
	}
}

type jumboResolver struct{ resolver }

func (jumboResolver) String(string) (uint32, error) { return 0x10000, nil }
func TestJumboString(t *testing.T) {
	a, e := Assemble(model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "const-string", Regs: []model.Register{0}, String: "x"}, {Op: "return-void"}}}, "()V", model.Static, jumboResolver{})
	if e != nil || !reflect.DeepEqual(a.Units, []uint16{0x1b, 0, 1, 0xe}) {
		t.Fatal(a, e)
	}
}
func TestBackwardGotoAndIfNez(t *testing.T) {
	c := model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "label", Label: "again"}, {Op: "const/4", Regs: []model.Register{0}, Literal: 1}, {Op: "if-nez", Regs: []model.Register{0}, Target: "again"}, {Op: "goto", Target: "again"}}}
	a, e := Assemble(c, "()V", model.Static, resolver{})
	if e != nil || !reflect.DeepEqual(a.Units, []uint16{0x1012, 0x39, 0xffff, 0x2a, 0xfffd, 0xffff}) {
		t.Fatal(a, e)
	}
}
