package builder

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"hash/adler32"
	"koru.dev/koru/internal/dex/model"
	"testing"
)

func simple() model.Program {
	return model.Program{Classes: []model.Class{{Descriptor: "Ldemo/Answer;", Super: "Ljava/lang/Object;", Access: model.Public, Methods: []model.Method{{Name: "answer", Descriptor: "()I", Access: model.Public | model.Static, Code: &model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "const/16", Regs: []model.Register{0}, Literal: 42}, {Op: "return", Regs: []model.Register{0}}}}}}}}}
}
func TestBuildHeaderSectionsAndIntegrity(t *testing.T) {
	b, e := Build(simple())
	if e != nil {
		t.Fatal(e)
	}
	u := func(p int) uint32 { return binary.LittleEndian.Uint32(b[p:]) }
	if string(b[:8]) != "dex\n035\x00" || u(32) != uint32(len(b)) || u(36) != 112 || u(40) != 0x12345678 {
		t.Fatal("header")
	}
	sha := sha1.Sum(b[32:])
	if !bytes.Equal(sha[:], b[12:32]) || adler32.Checksum(b[12:]) != u(8) {
		t.Fatal("integrity")
	}
	for _, p := range []int{60, 68, 76, 84, 92, 100, 108} {
		if u(p)%4 != 0 {
			t.Fatal("alignment", p)
		}
	}
	if u(96) != 1 || u(88) != 1 || u(64) != 3 {
		t.Fatal("tables")
	}
	m := u(52)
	if m%4 != 0 || u(int(m)) != 10 {
		t.Fatal("map", m, u(int(m)))
	}
	b2, e := Build(simple())
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal("nondeterministic", e)
	}
}
func TestRejectBadDefinitions(t *testing.T) {
	for _, modify := range []func(*model.Program){func(p *model.Program) { p.Classes = append(p.Classes, p.Classes[0]) }, func(p *model.Program) { p.Classes[0].Methods = append(p.Classes[0].Methods, p.Classes[0].Methods[0]) }, func(p *model.Program) { p.Classes[0].Descriptor = "bad" }, func(p *model.Program) { p.Classes[0].Super = p.Classes[0].Descriptor }, func(p *model.Program) { p.Classes[0].Methods[0].Access = model.Public | model.Protected }, func(p *model.Program) { p.Classes[0].Methods[0].Code = nil }, func(p *model.Program) {
		p.Classes[0].Methods[0].Code.Locals = 65535
		p.Classes[0].Methods[0].Access = model.Public
	}, func(p *model.Program) { p.Classes[0].Methods[0].Name = "<init>" }} {
		p := simple()
		modify(&p)
		if _, e := Build(p); e == nil {
			t.Fatal("accepted invalid definitions")
		}
	}
}
func BenchmarkBuild(b *testing.B) {
	p := simple()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := Build(p); e != nil {
			b.Fatal(e)
		}
	}
}
