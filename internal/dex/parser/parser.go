// Package parser independently walks DEX sections and validates the supported
// DEX 035 subset. It does not import the assembler, builder, or writer.
// Unsupported sections are errors, never silently accepted.
package parser

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"koru.dev/koru/internal/dex/encoding"
	"koru.dev/koru/internal/dex/format"
)

type File struct {
	Size        uint32           `json:"size"`
	Strings     []string         `json:"strings"`
	Types       []string         `json:"types"`
	Prototypes  []Prototype      `json:"prototypes"`
	Fields      []FieldID        `json:"fields"`
	Methods     []MethodID       `json:"methods"`
	Classes     []Class          `json:"classes"`
	Map         []format.MapItem `json:"map"`
	units       [][]uint16
	typeIndices []uint32
	lists       map[uint32][]uint16
	codes       map[uint32]*Code
	classData   map[uint32]data
}
type Prototype struct {
	Shorty, Return uint32
	Parameters     []uint16
}
type FieldID struct {
	Owner, Type uint16
	Name        uint32
}
type MethodID struct {
	Owner, Prototype uint16
	Name             uint32
}
type Class struct {
	Type, Access, Super, Source uint32
	Interfaces                  []uint16
	Fields                      []Field
	Methods                     []Method
}
type Field struct{ Index, Access uint32 }
type Method struct {
	Index, Access uint32
	Code          *Code
}
type Code struct {
	Offset                        uint32
	Registers, Incoming, Outgoing uint16
	Units                         []uint16
	Tries                         []Try
}
type Catch struct {
	Type, PC uint32
	CatchAll bool
}
type Try struct {
	Start, Count uint32
	Handler      Catch
}
type data struct {
	fields  [2][]Field
	methods [2][]encodedMethod
}
type encodedMethod struct{ index, access, offset uint32 }
type reader struct {
	b   []byte
	pos uint32
	end uint32
	err error
}

func (r *reader) take(n uint32) []byte {
	if r.err != nil {
		return nil
	}
	if uint64(r.pos)+uint64(n) > uint64(r.end) {
		r.err = fmt.Errorf("truncated section at 0x%x", r.pos)
		return nil
	}
	b := r.b[r.pos : r.pos+n]
	r.pos += n
	return b
}
func (r *reader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}
func (r *reader) u32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}
func (r *reader) leb() uint32 {
	if r.err != nil {
		return 0
	}
	v, n, e := encoding.ReadULEB(r.b[r.pos:r.end])
	if e != nil {
		r.err = e
		return 0
	}
	r.pos += uint32(n)
	return v
}
func (r *reader) padding() {
	for r.pos%4 != 0 && r.err == nil {
		b := r.take(1)
		if b != nil && b[0] != 0 {
			r.err = fmt.Errorf("nonzero alignment padding")
		}
	}
}
func (r *reader) finish() error {
	if r.err != nil {
		return r.err
	}
	for _, b := range r.b[r.pos:r.end] {
		if b != 0 {
			return fmt.Errorf("unparsed section bytes at 0x%x", r.pos)
		}
	}
	if r.end-r.pos > 3 {
		return fmt.Errorf("excessive padding")
	}
	return nil
}
func Parse(b []byte) (*File, error) {
	if len(b) < 112 || string(b[:8]) != format.Magic {
		return nil, fmt.Errorf("requires DEX 035 header")
	}
	if uint64(len(b)) > 0xffffffff {
		return nil, fmt.Errorf("oversized DEX")
	}
	u := func(p int) uint32 { return binary.LittleEndian.Uint32(b[p : p+4]) }
	if u(32) != uint32(len(b)) || u(36) != 112 || u(40) != format.Endian {
		return nil, fmt.Errorf("invalid size, header, or endian")
	}
	if u(44) != 0 || u(48) != 0 {
		return nil, fmt.Errorf("linked DEX unsupported")
	}
	if adler32.Checksum(b[12:]) != u(8) {
		return nil, fmt.Errorf("Adler-32 mismatch")
	}
	sig := sha1.Sum(b[32:])
	if !bytes.Equal(sig[:], b[12:32]) {
		return nil, fmt.Errorf("SHA-1 mismatch")
	}
	dataOff := u(108)
	if dataOff < 112 || dataOff%4 != 0 || uint64(dataOff)+uint64(u(104)) != uint64(len(b)) {
		return nil, fmt.Errorf("invalid data range")
	}
	mapOff := u(52)
	if mapOff < dataOff || mapOff%4 != 0 || uint64(mapOff)+4 > uint64(len(b)) {
		return nil, fmt.Errorf("invalid map offset")
	}
	mr := reader{b: b, pos: mapOff, end: uint32(len(b))}
	n := mr.u32()
	if n == 0 || n > (mr.end-mr.pos)/12 {
		return nil, fmt.Errorf("invalid map size")
	}
	f := &File{Size: uint32(len(b)), lists: map[uint32][]uint16{}, codes: map[uint32]*Code{}, classData: map[uint32]data{}}
	seen := map[uint16]bool{}
	for i := uint32(0); i < n; i++ {
		t := mr.u16()
		unused := mr.u16()
		size, off := mr.u32(), mr.u32()
		if unused != 0 || size == 0 || seen[t] || off >= uint32(len(b)) || (len(f.Map) > 0 && off <= f.Map[len(f.Map)-1].Offset) {
			return nil, fmt.Errorf("invalid or overlapping map entry")
		}
		seen[t] = true
		f.Map = append(f.Map, format.MapItem{Type: t, Size: size, Offset: off})
	}
	if mr.err != nil {
		return nil, mr.err
	}
	if f.Map[0] != (format.MapItem{Type: 0, Size: 1, Offset: 0}) {
		return nil, fmt.Errorf("missing header map item")
	}
	fixedSizes := map[uint16]uint32{0: 112, 1: 4, 2: 4, 3: 12, 4: 8, 5: 8, 6: 32}
	fixedEnd := uint32(112)
	for i, s := range f.Map {
		end := uint32(len(b))
		if i+1 < len(f.Map) {
			end = f.Map[i+1].Offset
		}
		if width, ok := fixedSizes[s.Type]; ok {
			if s.Offset%4 != 0 || uint64(s.Offset)+uint64(width)*uint64(s.Size) > uint64(end) {
				return nil, fmt.Errorf("overlapping fixed section")
			}
			if s.Type > 0 {
				if s.Size != u(56+int(s.Type-1)*8) || s.Offset != u(60+int(s.Type-1)*8) {
					return nil, fmt.Errorf("header/map mismatch")
				}
				finish := s.Offset + s.Size*width
				if finish > fixedEnd {
					fixedEnd = finish
				}
			}
		} else {
			if s.Offset < dataOff {
				return nil, fmt.Errorf("data section outside data area")
			}
		}
	}
	for t := uint16(1); t <= 6; t++ {
		size, off := u(56+int(t-1)*8), u(60+int(t-1)*8)
		if size == 0 && off != 0 || size > 0 && !seen[t] {
			return nil, fmt.Errorf("missing or empty table inconsistency")
		}
		if (t == 2 || t == 3) && size > 65535 || (t == 4 || t == 5) && size > 65536 {
			return nil, fmt.Errorf("identifier limit exceeded")
		}
	}
	if fixedEnd > dataOff {
		return nil, fmt.Errorf("data overlaps identifier tables")
	}
	if !seen[format.MapList] {
		return nil, fmt.Errorf("missing map_list")
	}
	// Parse physical data before following table offsets. References must point to
	// exact item starts, never merely somewhere within a plausible byte range.
	for i, s := range f.Map {
		end := uint32(len(b))
		if i+1 < len(f.Map) {
			end = f.Map[i+1].Offset
		}
		r := reader{b: b, pos: s.Offset, end: end}
		switch s.Type {
		case format.HeaderItem, format.StringID, format.TypeID, format.ProtoID, format.FieldID, format.MethodID, format.ClassDef:
			continue
		case format.MapList:
			if s.Size != 1 || s.Offset != mapOff || mr.pos > end {
				return nil, fmt.Errorf("map_list inconsistency")
			}
			r.pos = mr.pos
		case format.StringData:
			if s.Size > end-s.Offset {
				return nil, fmt.Errorf("string section count overflow")
			}
			for j := uint32(0); j < s.Size; j++ {
				off := r.pos
				units, n, e := encoding.ReadStringData(b[r.pos:end])
				if e != nil {
					return nil, e
				}
				r.pos += uint32(n)
				f.units = append(f.units, units)
				f.Strings = append(f.Strings, encoding.UTF16String(units))
				f.typeIndices = append(f.typeIndices, off)
			}
		case format.TypeList:
			if s.Offset%4 != 0 || s.Size > (end-s.Offset)/4 {
				return nil, fmt.Errorf("invalid type list section")
			}
			for j := uint32(0); j < s.Size; j++ {
				r.padding()
				off := r.pos
				count := r.u32()
				if count == 0 || count > (end-r.pos)/2 {
					return nil, fmt.Errorf("invalid type list size")
				}
				list := make([]uint16, count)
				for k := range list {
					list[k] = r.u16()
				}
				f.lists[off] = list
			}
		case format.CodeItem:
			if s.Offset%4 != 0 || s.Size > (end-s.Offset)/16 {
				return nil, fmt.Errorf("invalid code section")
			}
			for j := uint32(0); j < s.Size; j++ {
				r.padding()
				c := &Code{Offset: r.pos}
				c.Registers = r.u16()
				c.Incoming = r.u16()
				c.Outgoing = r.u16()
				tries := r.u16()
				debug := r.u32()
				count := r.u32()
				if r.err != nil {
					return nil, r.err
				}
				if debug != 0 {
					return nil, fmt.Errorf("exception/debug data unsupported")
				}
				if count == 0 || count > (end-r.pos)/2 || c.Incoming > c.Registers {
					return nil, fmt.Errorf("invalid code size or registers")
				}
				c.Units = make([]uint16, count)
				for k := range c.Units {
					c.Units[k] = r.u16()
				}
				if tries > 0 {
					if count%2 != 0 && r.u16() != 0 {
						return nil, fmt.Errorf("nonzero try padding")
					}
					if uint32(tries) > (r.end-r.pos)/8 {
						return nil, fmt.Errorf("truncated tries")
					}
					type rawTry struct {
						start, count uint32
						offset       uint16
					}
					raw := make([]rawTry, tries)
					for k := range raw {
						raw[k] = rawTry{r.u32(), uint32(r.u16()), r.u16()}
					}
					handlerStart := r.pos
					n := r.leb()
					if n == 0 || n > uint32(tries) {
						return nil, fmt.Errorf("invalid handler count")
					}
					handlers := map[uint16]Catch{}
					for k := uint32(0); k < n; k++ {
						relative := r.pos - handlerStart
						if relative > 65535 {
							return nil, fmt.Errorf("handler offset overflow")
						}
						sz, used, e := encoding.ReadSLEB(b[r.pos:r.end])
						if e != nil {
							return nil, e
						}
						r.pos += uint32(used)
						h := Catch{}
						if sz == 0 {
							h.CatchAll = true
						} else if sz == 1 {
							h.Type = r.leb()
						} else {
							return nil, fmt.Errorf("multiple catch clauses unsupported")
						}
						h.PC = r.leb()
						if r.err != nil {
							return nil, r.err
						}
						handlers[uint16(relative)] = h
					}
					for _, tr := range raw {
						h, ok := handlers[tr.offset]
						if !ok || tr.count == 0 || uint64(tr.start)+uint64(tr.count) > uint64(len(c.Units)) || h.PC >= uint32(len(c.Units)) {
							return nil, fmt.Errorf("invalid exception region")
						}
						if len(c.Tries) > 0 {
							prev := c.Tries[len(c.Tries)-1]
							if tr.start < prev.Start+prev.Count {
								return nil, fmt.Errorf("overlapping tries")
							}
						}
						c.Tries = append(c.Tries, Try{tr.start, tr.count, h})
					}
				}
				f.codes[c.Offset] = c
			}
		case format.ClassData:
			if s.Size > (end-s.Offset)/4 {
				return nil, fmt.Errorf("invalid class data count")
			}
			for j := uint32(0); j < s.Size; j++ {
				off := r.pos
				var d data
				var counts [4]uint32
				for k := range counts {
					counts[k] = r.leb()
				}
				for k, count := range counts {
					limit := u(80)
					if k >= 2 {
						limit = u(88)
					}
					if count > limit || uint64(count)*2 > uint64(end-r.pos) {
						return nil, fmt.Errorf("class member count overflow")
					}
					var idx uint32
					for m := uint32(0); m < count; m++ {
						diff := r.leb()
						if m > 0 && diff == 0 || uint64(idx)+uint64(diff) >= uint64(limit) {
							return nil, fmt.Errorf("invalid class member delta")
						}
						idx += diff
						a := r.leb()
						if k < 2 {
							d.fields[k] = append(d.fields[k], Field{idx, a})
						} else {
							d.methods[k-2] = append(d.methods[k-2], encodedMethod{idx, a, r.leb()})
						}
					}
				}
				f.classData[off] = d
			}
		default:
			return nil, fmt.Errorf("unsupported DEX section 0x%x", s.Type)
		}
		if e := r.finish(); e != nil {
			return nil, fmt.Errorf("section 0x%x: %w", s.Type, e)
		}
	}
	if e := f.tables(b); e != nil {
		return nil, e
	}
	if e := f.classes(b); e != nil {
		return nil, e
	}
	return f, nil
}
