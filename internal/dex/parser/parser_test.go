package parser_test

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"hash/adler32"
	"koru.dev/koru/internal/dex/builder"
	"koru.dev/koru/internal/dex/model"
	. "koru.dev/koru/internal/dex/parser"
	"koru.dev/koru/pkg/dex"
	runtimedex "koru.dev/koru/runtime/dex"
	"testing"
)

func fixture(t testing.TB) []byte {
	t.Helper()
	b, e := builder.Build(model.Program{Classes: []model.Class{{Descriptor: "Ldemo/Answer;", Super: "Ljava/lang/Object;", Access: 1, Methods: []model.Method{{Name: "answer", Descriptor: "()I", Access: 9, Code: &model.Code{Locals: 1, Instructions: []model.Instruction{{Op: "const/16", Regs: []model.Register{0}, Literal: 42}, {Op: "return", Regs: []model.Register{0}}}}}}}}})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func rehash(b []byte) {
	s := sha1.Sum(b[32:])
	copy(b[12:32], s[:])
	binary.LittleEndian.PutUint32(b[8:], adler32.Checksum(b[12:]))
}
func TestParseExecutable(t *testing.T) {
	b := fixture(t)
	f, e := Parse(b)
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Classes) != 1 || f.Types[f.Classes[0].Type] != "Ldemo/Answer;" || len(f.Classes[0].Methods) != 1 {
		t.Fatal(f)
	}
	m := f.Classes[0].Methods[0]
	if m.Code == nil || m.Code.Registers != 1 || !bytes.Equal([]byte{byte(m.Code.Units[0]), byte(m.Code.Units[0] >> 8)}, []byte{0x13, 0}) {
		t.Fatal(m)
	}
}
func TestRejectStructuralCorruptionWithValidHashes(t *testing.T) {
	orig := fixture(t)
	u := func(b []byte, p int) uint32 { return binary.LittleEndian.Uint32(b[p:]) }
	for _, c := range []struct {
		name string
		edit func([]byte)
	}{
		{"file size", func(b []byte) { binary.LittleEndian.PutUint32(b[32:], 1) }},
		{"header alignment", func(b []byte) { binary.LittleEndian.PutUint32(b[60:], 113) }},
		{"string offset", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 60):], uint32(len(b)-1)) }},
		{"type index", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 68):], 0xffffffff) }},
		{"prototype shorty", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 76):], 0xffffffff) }},
		{"map overlap", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 52)+4+12+8:], 0) }},
		{"class reference", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 100):], 0xffffffff) }},
		{"map count overflow", func(b []byte) { binary.LittleEndian.PutUint32(b[u(b, 52):], 0xffffffff) }},
		{"bad opcode", func(b []byte) {
			f, e := Parse(b)
			if e != nil {
				t.Fatal(e)
			}
			off := f.Classes[0].Methods[0].Code.Offset
			b[off+16] = 0xff
		}},
		{"bad register", func(b []byte) {
			f, e := Parse(b)
			if e != nil {
				t.Fatal(e)
			}
			off := f.Classes[0].Methods[0].Code.Offset
			b[off+17] = 1
		}},
		{"fall through", func(b []byte) {
			f, e := Parse(b)
			if e != nil {
				t.Fatal(e)
			}
			off := f.Classes[0].Methods[0].Code.Offset
			b[off+20] = 0
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := append([]byte(nil), orig...)
			c.edit(b)
			rehash(b)
			if _, e := Parse(b); e == nil {
				t.Fatal("accepted corrupt DEX")
			}
		})
	}
}
func TestRejectTruncationAndBadHashes(t *testing.T) {
	b := fixture(t)
	for i := 0; i < len(b); i++ {
		if _, e := Parse(b[:i]); e == nil {
			t.Fatalf("truncation %d", i)
		}
	}
	b[len(b)-1] ^= 1
	if _, e := Parse(b); e == nil {
		t.Fatal("checksum")
	}
}
func FuzzParse(f *testing.F) {
	f.Add(fixture(f))
	activity, e := dex.Build(runtimedex.ActivityExperiment())
	if e != nil {
		f.Fatal(e)
	}
	f.Add(activity)
	f.Add([]byte("dex\n035\x00"))
	f.Fuzz(func(t *testing.T, b []byte) {
		Parse(b)
		if len(b) >= 112 && len(b) < 1<<20 {
			x := append([]byte(nil), b...)
			rehash(x)
			Parse(x)
		}
	})
}

func TestRejectUnpairedSurrogateInTypeName(t *testing.T) {
	p := model.Program{Classes: []model.Class{{Descriptor: "Ldemo/\ue000;", Super: "Ljava/lang/Object;", Access: 1}}}
	b, e := builder.Build(p)
	if e != nil {
		t.Fatal(e)
	}
	pos := bytes.Index(b, []byte{0xee, 0x80, 0x80})
	if pos < 0 {
		t.Fatal("fixture missing name")
	}
	copy(b[pos:pos+3], []byte{0xed, 0xa0, 0x80})
	rehash(b)
	if _, e := Parse(b); e == nil {
		t.Fatal("accepted a lone surrogate in class name")
	}
}

func TestSHA1CheckedIndependentlyOfAdler32(t *testing.T) {
	b := fixture(t)
	b[12] ^= 1
	binary.LittleEndian.PutUint32(b[8:], adler32.Checksum(b[12:]))
	if _, e := Parse(b); e == nil {
		t.Fatal("SHA-1 corruption accepted")
	}
}
func TestMapAndDataCorruption(t *testing.T) {
	orig := fixture(t)
	f, e := Parse(orig)
	if e != nil {
		t.Fatal(e)
	}
	mapOffset := binary.LittleEndian.Uint32(orig[52:])
	var classData, codeEntry uint32
	for i, s := range f.Map {
		switch s.Type {
		case 0x2000:
			classData = s.Offset
		case 0x2001:
			codeEntry = mapOffset + 4 + uint32(i*12)
		}
	}
	cases := []struct {
		name string
		edit func([]byte)
	}{
		{"unknown section", func(b []byte) { binary.LittleEndian.PutUint16(b[codeEntry:], 0xffff) }},
		{"unsupported tries", func(b []byte) { binary.LittleEndian.PutUint16(b[f.Classes[0].Methods[0].Code.Offset+6:], 1) }},
		{"wrong parameter count", func(b []byte) { binary.LittleEndian.PutUint16(b[f.Classes[0].Methods[0].Code.Offset+2:], 1) }},
		{"method count", func(b []byte) { b[classData+2] = 0x7f }},
		{"wrong method classification", func(b []byte) { b[classData+2] = 0; b[classData+3] = 1 }},
		{"code size", func(b []byte) { binary.LittleEndian.PutUint32(b[f.Classes[0].Methods[0].Code.Offset+12:], 0xffffffff) }},
		{"string count", func(b []byte) {
			for _, s := range f.Map {
				if s.Type == 0x2002 {
					b[s.Offset] = 127
					break
				}
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := append([]byte(nil), orig...)
			c.edit(b)
			rehash(b)
			if _, e := Parse(b); e == nil {
				t.Fatal("accepted corrupt DEX")
			}
		})
	}
}
