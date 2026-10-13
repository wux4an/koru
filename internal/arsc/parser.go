package arsc

import (
	"encoding/binary"
	"errors"
	"koru.dev/koru/internal/androidbin"
	"unicode/utf16"
)

type Info struct {
	Package, IconPath string
	IconID            uint32
}

func Parse(b []byte) (Info, error) {
	var out Info
	if len(b) < 12 || binary.LittleEndian.Uint16(b) != 2 || binary.LittleEndian.Uint16(b[2:]) != 12 || int(binary.LittleEndian.Uint32(b[4:])) != len(b) || binary.LittleEndian.Uint32(b[8:]) != 1 {
		return out, errors.New("unsupported ARSC table")
	}
	read := func(data []byte, p int) ([]byte, error) {
		if p < 0 || len(data)-p < 8 {
			return nil, errors.New("truncated resource chunk")
		}
		size := uint64(binary.LittleEndian.Uint32(data[p+4:]))
		head := binary.LittleEndian.Uint16(data[p+2:])
		if size < uint64(head) || head < 8 || size%4 != 0 || size > uint64(len(data)-p) {
			return nil, errors.New("invalid resource chunk")
		}
		return data[p : p+int(size)], nil
	}
	global, e := read(b, 12)
	if e != nil {
		return out, e
	}
	strings, e := androidbin.ReadPool(global)
	if e != nil {
		return out, e
	}
	pkg, e := read(b, 12+len(global))
	if e != nil {
		return out, e
	}
	if 12+len(global)+len(pkg) != len(b) || len(pkg) < 288 || binary.LittleEndian.Uint16(pkg) != 0x200 || binary.LittleEndian.Uint16(pkg[2:]) != 288 || binary.LittleEndian.Uint32(pkg[8:]) != 0x7f {
		return out, errors.New("invalid resource package")
	}
	var name []uint16
	for p := 12; p < 268; p += 2 {
		v := binary.LittleEndian.Uint16(pkg[p:])
		if v == 0 {
			break
		}
		name = append(name, v)
	}
	out.Package = string(utf16.Decode(name))
	tp := int(binary.LittleEndian.Uint32(pkg[268:]))
	kp := int(binary.LittleEndian.Uint32(pkg[276:]))
	types, e := read(pkg, tp)
	if e != nil {
		return out, e
	}
	typeNames, e := androidbin.ReadPool(types)
	if e != nil {
		return out, e
	}
	keys, e := read(pkg, kp)
	if e != nil {
		return out, e
	}
	keyNames, e := androidbin.ReadPool(keys)
	if e != nil {
		return out, e
	}
	if tp != 288 || kp != 288+len(types) || len(typeNames) != 1 || typeNames[0] != "mipmap" || len(keyNames) != 1 || keyNames[0] != "ic_launcher" {
		return out, errors.New("unsupported resource symbols")
	}
	spec, e := read(pkg, kp+len(keys))
	if e != nil {
		return out, e
	}
	if len(spec) != 20 || binary.LittleEndian.Uint16(spec) != 0x202 || binary.LittleEndian.Uint16(spec[2:]) != 16 || spec[8] != 1 || binary.LittleEndian.Uint32(spec[12:]) != 1 {
		return out, errors.New("invalid type spec")
	}
	typ, e := read(pkg, kp+len(keys)+len(spec))
	if e != nil {
		return out, e
	}
	if kp+len(keys)+len(spec)+len(typ) != len(pkg) || len(typ) != 104 || binary.LittleEndian.Uint16(typ) != 0x201 || binary.LittleEndian.Uint16(typ[2:]) != 84 || typ[8] != 1 || typ[9] != 0 || binary.LittleEndian.Uint32(typ[12:]) != 1 || binary.LittleEndian.Uint32(typ[16:]) != 88 || binary.LittleEndian.Uint32(typ[20:]) != 64 || binary.LittleEndian.Uint32(typ[84:]) != 0 || binary.LittleEndian.Uint16(typ[88:]) != 8 || binary.LittleEndian.Uint16(typ[90:]) != 0 || binary.LittleEndian.Uint32(typ[92:]) != 0 || binary.LittleEndian.Uint16(typ[96:]) != 8 || typ[98] != 0 || typ[99] != 3 {
		return out, errors.New("invalid icon resource entry")
	}
	idx := binary.LittleEndian.Uint32(typ[100:])
	if uint64(idx) >= uint64(len(strings)) {
		return out, errors.New("invalid icon path index")
	}
	out.IconID = IconID
	out.IconPath = strings[idx]
	return out, nil
}
