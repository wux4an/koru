package encoding

import (
	"bytes"
	"math"
	"testing"
)

func TestULEBGolden(t *testing.T) {
	for _, c := range []struct {
		v uint32
		b []byte
	}{{0, []byte{0}}, {127, []byte{127}}, {128, []byte{0x80, 1}}, {624485, []byte{0xe5, 0x8e, 0x26}}, {math.MaxUint32, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}}} {
		b := ULEB(c.v)
		if !bytes.Equal(b, c.b) {
			t.Fatalf("%d: %x != %x", c.v, b, c.b)
		}
		v, n, e := ReadULEB(b)
		if e != nil || v != c.v || n != len(b) {
			t.Fatal(v, n, e)
		}
	}
}
func TestSLEBGolden(t *testing.T) {
	for _, c := range []struct {
		v int32
		b []byte
	}{{0, []byte{0}}, {-1, []byte{0x7f}}, {-128, []byte{0x80, 0x7f}}, {-624485, []byte{0x9b, 0xf1, 0x59}}, {math.MinInt32, []byte{0x80, 0x80, 0x80, 0x80, 0x78}}, {math.MaxInt32, []byte{0xff, 0xff, 0xff, 0xff, 7}}} {
		b := SLEB(c.v)
		if !bytes.Equal(b, c.b) {
			t.Fatalf("%d: %x", c.v, b)
		}
		v, n, e := ReadSLEB(b)
		if e != nil || v != c.v || n != len(b) {
			t.Fatal(v, n, e)
		}
	}
}
func TestLEBRejectsOverflowAndTruncation(t *testing.T) {
	for _, b := range [][]byte{{}, {0x80}, {0xff, 0xff, 0xff, 0xff, 0x10}, {0x80, 0x80, 0x80, 0x80, 0x80}} {
		if _, _, e := ReadULEB(b); e == nil {
			t.Fatalf("accepted %x", b)
		}
	}
	for _, b := range [][]byte{{}, {0x80}, {0xff, 0xff, 0xff, 0xff, 8}, {0x80, 0x80, 0x80, 0x80, 0x70}} {
		if _, _, e := ReadSLEB(b); e == nil {
			t.Fatalf("accepted %x", b)
		}
	}
}
func TestMUTF8Golden(t *testing.T) {
	for _, c := range []struct {
		s string
		b []byte
	}{{"", []byte{0, 0}}, {"A\x00", []byte{2, 65, 0xc0, 0x80, 0}}, {"é", []byte{1, 0xc3, 0xa9, 0}}, {"😀", []byte{2, 0xed, 0xa0, 0xbd, 0xed, 0xb8, 0x80, 0}}} {
		b, e := StringData(c.s)
		if e != nil || !bytes.Equal(b, c.b) {
			t.Fatal(c.s, b, e)
		}
		u, n, e := ReadStringData(b)
		if e != nil || n != len(b) || UTF16String(u) != c.s {
			t.Fatal(u, n, e)
		}
	}
}
func TestMUTF8Malformed(t *testing.T) {
	for _, b := range [][]byte{{}, {1, 65}, {2, 65, 0}, {1, 0xc1, 0x81, 0}, {1, 0xf0, 0x90, 0x80, 0x80, 0}, {1, 0xe0, 0x80, 0x80, 0}, {1, 0x80, 0}} {
		if _, _, e := ReadStringData(b); e == nil {
			t.Fatalf("accepted %x", b)
		}
	}
	if _, e := StringData(string([]byte{0xff})); e == nil {
		t.Fatal("invalid UTF-8")
	}
}
func TestUnpairedSurrogatePreserved(t *testing.T) {
	u, n, e := ReadStringData([]byte{1, 0xed, 0xa0, 0x80, 0})
	if e != nil || n != 5 || len(u) != 1 || u[0] != 0xd800 {
		t.Fatal(u, n, e)
	}
}
func FuzzLEB(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 15})
	f.Fuzz(func(t *testing.T, b []byte) { ReadULEB(b); ReadSLEB(b) })
}
func FuzzStringData(f *testing.F) {
	f.Add([]byte{1, 65, 0})
	f.Fuzz(func(t *testing.T, b []byte) { ReadStringData(b) })
}
