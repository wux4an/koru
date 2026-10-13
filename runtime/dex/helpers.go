package runtimedex

import "koru.dev/koru/pkg/dex"

func call(b *dex.CodeBuilder, op, owner, name, descriptor string, regs ...dex.Register) {
	b.Invoke(op, dex.MethodRef{Owner: owner, Name: name, Descriptor: descriptor}, regs...)
}
func object(b *dex.CodeBuilder, r dex.Register) {
	b.Emit(dex.Instruction{Op: "move-result-object", Regs: []dex.Register{r}})
}
func scalar(b *dex.CodeBuilder, r dex.Register) {
	b.Emit(dex.Instruction{Op: "move-result", Regs: []dex.Register{r}})
}
func retObject(b *dex.CodeBuilder, r dex.Register) {
	b.Emit(dex.Instruction{Op: "return-object", Regs: []dex.Register{r}})
}
func retScalar(b *dex.CodeBuilder, r dex.Register) {
	b.Emit(dex.Instruction{Op: "return", Regs: []dex.Register{r}})
}
func fresh(b *dex.CodeBuilder, r dex.Register, typ string) {
	b.Emit(dex.Instruction{Op: "new-instance", Regs: []dex.Register{r}, Type: typ})
}
func branch(b *dex.CodeBuilder, op string, target string, regs ...dex.Register) {
	b.Emit(dex.Instruction{Op: op, Regs: regs, Target: target})
}
func moveObject(b *dex.CodeBuilder, d, s dex.Register) {
	b.Emit(dex.Instruction{Op: "move-object", Regs: []dex.Register{d, s}})
}
func field(b *dex.CodeBuilder, op, owner, name, typ string, regs ...dex.Register) {
	r := dex.FieldRef{Owner: owner, Name: name, Type: typ}
	b.Emit(dex.Instruction{Op: op, Field: &r, Regs: regs})
}
func ctor(owner, super string, extraFields []dex.Field) dex.Class {
	b := dex.NewCode(0)
	call(b, "invoke-direct", super, "<init>", "()V", dex.P(0))
	b.ReturnVoid()
	return dex.Class{Descriptor: owner, Super: super, Access: dex.Public, Fields: extraFields, Methods: []dex.Method{{Name: "<init>", Descriptor: "()V", Access: dex.Public | dex.Constructor, Code: b.Build()}}}
}
