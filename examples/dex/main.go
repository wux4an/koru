// A real static method and constructor, generated without Java source.
package main

import (
	"koru.dev/koru/pkg/dex"
	"log"
	"os"
)

func main() {
	c := dex.Class{Descriptor: "Lexample/Answer;", Super: "Ljava/lang/Object;", Access: dex.Public}
	c.Methods = []dex.Method{{Name: "<init>", Descriptor: "()V", Access: dex.Public | dex.Constructor, Code: dex.NewCode(0).Invoke("invoke-direct", dex.MethodRef{Owner: c.Super, Name: "<init>", Descriptor: "()V"}, dex.P(0)).ReturnVoid().Build()}, {Name: "answer", Descriptor: "()I", Access: dex.Public | dex.Static, Code: dex.NewCode(1).Const(dex.V(0), 42).Emit(dex.Instruction{Op: "return", Regs: []dex.Register{dex.V(0)}}).Build()}}
	b, e := dex.Build(dex.Program{Classes: []dex.Class{c}})
	if e != nil {
		log.Fatal(e)
	}
	if e = os.WriteFile("answer.dex", b, 0644); e != nil {
		log.Fatal(e)
	}
}
