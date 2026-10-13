// Package androidbin implements Android resource chunk primitives from ResourceTypes.h.
package androidbin

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"
	"unicode/utf8"
)

func U16(b []byte, v uint16) []byte          { return append(b, byte(v), byte(v>>8)) }
func U32(b []byte, v uint32) []byte          { return append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24)) }
func Header(t, h uint16, size uint32) []byte { return U32(U16(U16(nil, t), h), size) }
func Pad(b []byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}
func length(b []byte, n int) []byte {
	if n >= 128 {
		b = append(b, byte(n>>8)|128)
	}
	return append(b, byte(n))
}

const maxPoolDecodeCost = 64 << 20

func Pool(strings []string) ([]byte, error) {
	offsets := []byte{}
	data := []byte{}
	var cost uint64
	for _, s := range strings {
		n := len(utf16.Encode([]rune(s)))
		if !utf8.ValidString(s) || len(s) > 32767 || n > 32767 {
			return nil, errors.New("Android pool string too long or invalid")
		}
		cost += uint64(64 + len(s) + 2*n)
		if cost > maxPoolDecodeCost {
			return nil, errors.New("string pool decoded size limit")
		}
		offsets = U32(offsets, uint32(len(data)))
		data = length(data, n)
		data = length(data, len(s))
		data = append(data, []byte(s)...)
		data = append(data, 0)
	}
	data = Pad(data)
	b := Header(1, 28, uint32(28+len(offsets)+len(data)))
	b = U32(b, uint32(len(strings)))
	b = U32(b, 0)
	b = U32(b, 0x100)
	b = U32(b, uint32(28+len(offsets)))
	b = U32(b, 0)
	b = append(b, offsets...)
	return append(b, data...), nil
}
func ReadPool(b []byte) ([]string, error) {
	if len(b) < 28 || binary.LittleEndian.Uint16(b) != 1 || binary.LittleEndian.Uint16(b[2:]) != 28 || int(binary.LittleEndian.Uint32(b[4:])) != len(b) {
		return nil, errors.New("invalid string pool header")
	}
	n := binary.LittleEndian.Uint32(b[8:])
	start := binary.LittleEndian.Uint32(b[20:])
	if uint64(n)*64 > maxPoolDecodeCost || n > uint32((len(b)-28)/4) || uint64(28)+uint64(n)*4 > uint64(start) || start > uint32(len(b)) || binary.LittleEndian.Uint32(b[12:]) != 0 || binary.LittleEndian.Uint32(b[16:]) != 0x100 {
		return nil, errors.New("unsupported string pool")
	}
	var cost uint64
	var out []string
	readLen := func(p *int) (int, error) {
		if *p >= len(b) {
			return 0, errors.New("truncated pool length")
		}
		v := int(b[*p])
		*p++
		if v&128 != 0 {
			if *p >= len(b) {
				return 0, errors.New("truncated pool length")
			}
			v = (v&127)<<8 | int(b[*p])
			*p++
		}
		return v, nil
	}
	for i := uint32(0); i < n; i++ {
		off := uint64(start) + uint64(binary.LittleEndian.Uint32(b[28+i*4:]))
		if off >= uint64(len(b)) {
			return nil, errors.New("string offset out of bounds")
		}
		p := int(off)
		count, e := readLen(&p)
		if e != nil {
			return nil, e
		}
		size, e := readLen(&p)
		if e != nil || size >= len(b)-p {
			return nil, errors.New("truncated string")
		}
		cost += uint64(64 + size + 2*count)
		if cost > maxPoolDecodeCost {
			return nil, errors.New("string pool decoded size limit")
		}
		s := string(b[p : p+size])
		if b[p+size] != 0 || !utf8.ValidString(s) || len(utf16.Encode([]rune(s))) != count {
			return nil, errors.New("invalid pool string")
		}
		out = append(out, s)
	}
	return out, nil
}
