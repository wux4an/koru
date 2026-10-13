// Package arsc creates a default configuration with one launcher icon resource.
package arsc

import (
	"errors"
	"koru.dev/koru/internal/androidbin"
	"unicode/utf16"
)

const IconID uint32 = 0x7f010000

func Build(appID string) ([]byte, error) {
	name := utf16.Encode([]rune(appID))
	if len(name) > 127 {
		return nil, errors.New("resource package name too long")
	}
	global, e := androidbin.Pool([]string{"res/mipmap/ic_launcher.png"})
	if e != nil {
		return nil, e
	}
	types, _ := androidbin.Pool([]string{"mipmap"})
	keys, _ := androidbin.Pool([]string{"ic_launcher"})
	spec := androidbin.Header(0x202, 16, 20)
	spec = append(spec, 1, 0, 0, 0)
	spec = androidbin.U32(spec, 1)
	spec = androidbin.U32(spec, 0)
	typ := androidbin.Header(0x201, 84, 104)
	typ = append(typ, 1, 0, 0, 0)
	typ = androidbin.U32(typ, 1)
	typ = androidbin.U32(typ, 88)
	typ = androidbin.U32(typ, 64)
	typ = append(typ, make([]byte, 60)...)
	typ = androidbin.U32(typ, 0)
	typ = androidbin.U16(typ, 8)
	typ = androidbin.U16(typ, 0)
	typ = androidbin.U32(typ, 0)
	typ = androidbin.U16(typ, 8)
	typ = append(typ, 0, 3)
	typ = androidbin.U32(typ, 0)
	size := 288 + len(types) + len(keys) + len(spec) + len(typ)
	pkg := androidbin.Header(0x200, 288, uint32(size))
	pkg = androidbin.U32(pkg, 0x7f)
	for i := 0; i < 128; i++ {
		v := uint16(0)
		if i < len(name) {
			v = name[i]
		}
		pkg = androidbin.U16(pkg, v)
	}
	pkg = androidbin.U32(pkg, 288)
	pkg = androidbin.U32(pkg, 1)
	pkg = androidbin.U32(pkg, uint32(288+len(types)))
	pkg = androidbin.U32(pkg, 1)
	pkg = androidbin.U32(pkg, 0)
	for _, b := range [][]byte{types, keys, spec, typ} {
		pkg = append(pkg, b...)
	}
	table := androidbin.Header(2, 12, uint32(12+len(global)+len(pkg)))
	table = androidbin.U32(table, 1)
	table = append(table, global...)
	return append(table, pkg...), nil
}
