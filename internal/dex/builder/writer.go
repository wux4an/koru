package builder

import (
	"encoding/binary"
	"fmt"
	"koru.dev/koru/internal/dex/encoding"
	"koru.dev/koru/internal/dex/format"
)

// A bounded append-only writer; table fixups never change already assigned offsets.
type writer struct {
	b        []byte
	err      error
	sections []format.MapItem
}

func (w *writer) pos() uint32 {
	if uint64(len(w.b)) > 0xffffffff {
		w.err = fmt.Errorf("DEX exceeds 4 GiB")
	}
	return uint32(len(w.b))
}
func (w *writer) bytes(b []byte) {
	if w.err != nil {
		return
	}
	if uint64(len(w.b))+uint64(len(b)) > 0xffffffff {
		w.err = fmt.Errorf("DEX exceeds 4 GiB")
		return
	}
	w.b = append(w.b, b...)
}
func (w *writer) u16(v uint16) { w.bytes([]byte{byte(v), byte(v >> 8)}) }
func (w *writer) u32(v uint32) { w.bytes([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}) }
func (w *writer) leb(v uint32) { w.bytes(encoding.ULEB(v)) }
func (w *writer) align() {
	for len(w.b)%4 != 0 {
		w.bytes([]byte{0})
		if w.err != nil {
			return
		}
	}
}
func (w *writer) put(p, v uint32) {
	if uint64(p)+4 > uint64(len(w.b)) {
		w.err = fmt.Errorf("invalid writer fixup")
		return
	}
	binary.LittleEndian.PutUint32(w.b[p:p+4], v)
}
func (w *writer) section(t uint16, n int) {
	if n > 0 {
		w.sections = append(w.sections, format.MapItem{Type: t, Size: uint32(n), Offset: w.pos()})
	}
}
func (w *writer) table(t uint16, n, width int) uint32 {
	if n == 0 {
		return 0
	}
	w.section(t, n)
	p := w.pos()
	size := uint64(n) * uint64(width)
	if size > 0xffffffff-uint64(p) {
		w.err = fmt.Errorf("table overflow")
		return 0
	}
	w.bytes(make([]byte, int(size)))
	return p
}
