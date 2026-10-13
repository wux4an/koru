package parser

import (
	"encoding/binary"
	"fmt"
	"koru.dev/koru/internal/dex/encoding"
	"koru.dev/koru/internal/dex/model"
	"strings"
)

func (f *File) tables(b []byte) error {
	u := func(p uint32) uint32 { return binary.LittleEndian.Uint32(b[p : p+4]) }
	sCount, sOff := u(56), u(60)
	if int(sCount) != len(f.Strings) {
		return fmt.Errorf("string table/data count mismatch")
	}
	byOff := map[uint32][]uint16{}
	for i, off := range f.typeIndices {
		byOff[off] = f.units[i]
	}
	f.units = nil
	f.Strings = nil
	used := map[uint32]bool{}
	for i := uint32(0); i < sCount; i++ {
		off := u(sOff + i*4)
		units, ok := byOff[off]
		if !ok || used[off] {
			return fmt.Errorf("invalid string_data offset")
		}
		used[off] = true
		if i > 0 && encoding.CompareUTF16(f.units[i-1], units) >= 0 {
			return fmt.Errorf("strings not strictly UTF-16 ordered")
		}
		f.units = append(f.units, units)
		f.Strings = append(f.Strings, encoding.UTF16String(units))
	}
	f.typeIndices = nil
	tCount, tOff := u(64), u(68)
	var prev uint32
	for i := uint32(0); i < tCount; i++ {
		idx := u(tOff + i*4)
		if idx >= sCount || i > 0 && idx <= prev {
			return fmt.Errorf("invalid or unordered type identifier")
		}
		s := f.Strings[idx]
		if e := model.ValidateType(s, true); e != nil {
			return e
		}
		f.Types = append(f.Types, s)
		f.typeIndices = append(f.typeIndices, idx)
		prev = idx
	}
	for _, list := range f.lists {
		for _, idx := range list {
			if int(idx) >= len(f.Types) || f.Types[idx] == "V" {
				return fmt.Errorf("invalid type_list reference")
			}
		}
	}
	pCount, pOff := u(72), u(76)
	for i := uint32(0); i < pCount; i++ {
		base := pOff + i*12
		p := Prototype{Shorty: u(base), Return: u(base + 4)}
		off := u(base + 8)
		if p.Shorty >= sCount || p.Return >= tCount {
			return fmt.Errorf("invalid prototype index")
		}
		if off != 0 {
			var ok bool
			p.Parameters, ok = f.lists[off]
			if !ok {
				return fmt.Errorf("prototype does not reference type_list start")
			}
		}
		pr := model.Prototype{Return: f.Types[p.Return]}
		for _, idx := range p.Parameters {
			pr.Parameters = append(pr.Parameters, f.Types[idx])
		}
		if f.Strings[p.Shorty] != pr.Shorty() {
			return fmt.Errorf("shorty mismatch")
		}
		if i > 0 && !protoLess(f.Prototypes[i-1], p) {
			return fmt.Errorf("duplicate or unordered prototype")
		}
		f.Prototypes = append(f.Prototypes, p)
	}
	fCount, fOff := u(80), u(84)
	for i := uint32(0); i < fCount; i++ {
		base := fOff + i*8
		packed := u(base)
		v := FieldID{uint16(packed), uint16(packed >> 16), u(base + 4)}
		if uint32(v.Owner) >= tCount || uint32(v.Type) >= tCount || v.Name >= sCount {
			return fmt.Errorf("invalid field identifier")
		}
		if e := model.ValidateFieldRef(model.FieldRef{Owner: f.Types[v.Owner], Name: f.Strings[v.Name], Type: f.Types[v.Type]}); e != nil {
			return e
		}
		if i > 0 {
			a := f.Fields[i-1]
			if !(a.Owner < v.Owner || a.Owner == v.Owner && (a.Name < v.Name || a.Name == v.Name && a.Type < v.Type)) {
				return fmt.Errorf("duplicate or unordered field")
			}
		}
		f.Fields = append(f.Fields, v)
	}
	mCount, mOff := u(88), u(92)
	for i := uint32(0); i < mCount; i++ {
		base := mOff + i*8
		packed := u(base)
		v := MethodID{uint16(packed), uint16(packed >> 16), u(base + 4)}
		if uint32(v.Owner) >= tCount || uint32(v.Prototype) >= pCount || v.Name >= sCount {
			return fmt.Errorf("invalid method identifier")
		}
		if e := model.ValidateMethodRef(model.MethodRef{Owner: f.Types[v.Owner], Name: f.Strings[v.Name], Descriptor: f.descriptor(v.Prototype)}); e != nil {
			return e
		}
		if i > 0 {
			a := f.Methods[i-1]
			if !(a.Owner < v.Owner || a.Owner == v.Owner && (a.Name < v.Name || a.Name == v.Name && a.Prototype < v.Prototype)) {
				return fmt.Errorf("duplicate or unordered method")
			}
		}
		f.Methods = append(f.Methods, v)
	}
	return nil
}
func protoLess(a, b Prototype) bool {
	if a.Return != b.Return {
		return a.Return < b.Return
	}
	for i := 0; i < len(a.Parameters) && i < len(b.Parameters); i++ {
		if a.Parameters[i] != b.Parameters[i] {
			return a.Parameters[i] < b.Parameters[i]
		}
	}
	return len(a.Parameters) < len(b.Parameters)
}
func (f *File) descriptor(idx uint16) string {
	p := f.Prototypes[idx]
	var b strings.Builder
	b.WriteByte('(')
	for _, v := range p.Parameters {
		b.WriteString(f.Types[v])
	}
	b.WriteByte(')')
	b.WriteString(f.Types[p.Return])
	return b.String()
}
