package verify

import (
	"strings"
	"testing"

	"github.com/wux4an/koru/internal/dex/model"
)

func method(desc string, access uint32, locals uint16, tries []model.Try, ins ...model.Instruction) model.Method {
	return model.Method{Name: "m", Descriptor: desc, Access: access, Code: &model.Code{Locals: locals, Tries: tries, Instructions: ins}}
}
func call(op, owner, name, desc string, regs ...model.Register) model.Instruction {
	return model.Instruction{Op: op, Method: &model.MethodRef{Owner: owner, Name: name, Descriptor: desc}, Regs: regs}
}
func op(name string, regs ...model.Register) model.Instruction {
	return model.Instruction{Op: name, Regs: regs}
}

const str = "Ljava/lang/String;"

func TestAcceptsValidCode(t *testing.T) {
	cases := map[string]model.Method{
		"string length": method("("+str+")I", model.Public|model.Static, 1, nil,
			call("invoke-virtual", str, "length", "()I", model.P(0)),
			op("move-result", 0),
			op("return", 0)),
		"new object": method("()Ljava/lang/Object;", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "new-instance", Regs: []model.Register{0}, Type: "Ljava/lang/Object;"},
			call("invoke-direct", "Ljava/lang/Object;", "<init>", "()V", 0),
			op("return-object", 0)),
		"null as reference and int": method("()V", model.Public|model.Static, 2, nil,
			model.Instruction{Op: "const", Regs: []model.Register{0}},
			model.Instruction{Op: "const", Regs: []model.Register{1}},
			call("invoke-static", "Lx/Y;", "f", "("+str+"I)V", 0, 1),
			op("return-void")),
		"branches merge": method("(I)"+str, model.Public|model.Static, 1, nil,
			model.Instruction{Op: "const-string", Regs: []model.Register{0}, String: "a"},
			model.Instruction{Op: "if-eqz", Regs: []model.Register{model.P(0)}, Target: "x"},
			model.Instruction{Op: "const-string", Regs: []model.Register{0}, String: "b"},
			model.Instruction{Op: "label", Label: "x"},
			op("return-object", 0)),
		"constructor": {Name: "<init>", Descriptor: "()V", Access: model.Public | model.Constructor, Code: &model.Code{Instructions: []model.Instruction{
			call("invoke-direct", "Ljava/lang/Object;", "<init>", "()V", model.P(0)), op("return-void")}}},
		"handler": method("()V", model.Public|model.Static, 1, []model.Try{{Start: "s", End: "e", Handler: "h"}},
			model.Instruction{Op: "label", Label: "s"},
			call("invoke-static", "Lx/Y;", "f", "()V"),
			model.Instruction{Op: "label", Label: "e"},
			op("return-void"),
			model.Instruction{Op: "label", Label: "h"},
			op("move-exception", 0),
			op("return-void")),
		"long argument": method("(I)V", model.Public|model.Static, 2, nil,
			op("int-to-long", 0, model.P(0)),
			call("invoke-static", "Ljava/lang/Thread;", "sleep", "(J)V", 0, 1),
			op("return-void")),
	}
	for name, m := range cases {
		if err := Method("Lt/T;", m); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestRejectsInvalidCode(t *testing.T) {
	cases := map[string]struct {
		m    model.Method
		want string
	}{
		"move-result-object after an int invoke": {method("("+str+")V", model.Public|model.Static, 1, nil,
			call("invoke-virtual", str, "length", "()I", model.P(0)), op("move-result-object", 0), op("return-void")), "does not follow"},
		"move-result without invoke": {method("()V", model.Public|model.Static, 1, nil,
			op("move-result", 0), op("return-void")), "does not follow"},
		"undefined register": {method("()V", model.Public|model.Static, 1, nil,
			call("invoke-static", "Lx/Y;", "f", "("+str+")V", 0), op("return-void")), "holds undefined"},
		"int passed as object": {method("()V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "const", Regs: []model.Register{0}, Literal: 7},
			call("invoke-static", "Lx/Y;", "f", "("+str+")V", 0), op("return-void")), "holds int"},
		"wrong arity": {method("()V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "const-string", Regs: []model.Register{0}, String: "x"},
			call("invoke-virtual", str, "length", "()I"), op("return-void")), "needs 1"},
		"uninitialized object used": {method("()V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "new-instance", Regs: []model.Register{0}, Type: "Ljava/lang/Object;"},
			call("invoke-virtual", "Ljava/lang/Object;", "hashCode", "()I", 0), op("return-void")), "holds uninitialized"},
		"conflicting merge": {method("(I)V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "const-string", Regs: []model.Register{0}, String: "a"},
			model.Instruction{Op: "if-eqz", Regs: []model.Register{model.P(0)}, Target: "x"},
			model.Instruction{Op: "const", Regs: []model.Register{0}, Literal: 3},
			model.Instruction{Op: "label", Label: "x"},
			call("invoke-static", "Lx/Y;", "f", "("+str+")V", 0), op("return-void")), "holds conflicting"},
		"falls off the end": {method("()V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "const", Regs: []model.Register{0}}), "falls off"},
		"object field read as int": {method("()I", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "sget", Regs: []model.Register{0}, Field: &model.FieldRef{Owner: "Lx/Y;", Name: "s", Type: str}},
			op("return", 0)), "holds reference"},
		"register reused after the handler": {method("()V", model.Public|model.Static, 1, []model.Try{{Start: "s", End: "e", Handler: "h"}},
			model.Instruction{Op: "label", Label: "s"},
			model.Instruction{Op: "const", Regs: []model.Register{0}, Literal: 5},
			call("invoke-static", "Lx/Y;", "f", "()V"),
			model.Instruction{Op: "const-string", Regs: []model.Register{0}, String: "x"},
			model.Instruction{Op: "label", Label: "e"},
			model.Instruction{Op: "label", Label: "h"},
			call("invoke-static", "Lx/Y;", "g", "("+str+")V", 0), op("return-void")), "holds"},
		"wide half passed alone": {method("(I)V", model.Public|model.Static, 2, nil,
			op("int-to-long", 0, model.P(0)),
			call("invoke-static", "Lx/Y;", "g", "(I)V", 0), op("return-void")), "holds wide-low"},
		"constructor called twice": {method("()V", model.Public|model.Static, 1, nil,
			model.Instruction{Op: "new-instance", Regs: []model.Register{0}, Type: "Ljava/lang/Object;"},
			call("invoke-direct", "Ljava/lang/Object;", "<init>", "()V", 0),
			call("invoke-direct", "Ljava/lang/Object;", "<init>", "()V", 0), op("return-void")), "holds reference"},
	}
	for name, c := range cases {
		err := Method("Lt/T;", c.m)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want an error containing %q", name, err, c.want)
		}
	}
}
