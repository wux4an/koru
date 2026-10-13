package encoding

import (
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

// StringData includes the UTF-16 unit count and terminator, not a byte count.
func StringData(s string) ([]byte, error) {
	if !utf8.ValidString(s) {
		return nil, fmt.Errorf("invalid UTF-8 input")
	}
	units := utf16.Encode([]rune(s))
	if uint64(len(units)) > 0xffffffff {
		return nil, fmt.Errorf("string too long")
	}
	b := ULEB(uint32(len(units)))
	for _, u := range units {
		switch {
		case u == 0:
			b = append(b, 0xc0, 0x80)
		case u < 128:
			b = append(b, byte(u))
		case u < 2048:
			b = append(b, 0xc0|byte(u>>6), 0x80|byte(u&63))
		default:
			b = append(b, 0xe0|byte(u>>12), 0x80|byte((u>>6)&63), 0x80|byte(u&63))
		}
	}
	return append(b, 0), nil
}

// ReadStringData preserves lone surrogate units: DEX strings are UTF-16, not Go UTF-8.
func ReadStringData(b []byte) ([]uint16, int, error) {
	count, p, e := ReadULEB(b)
	if e != nil {
		return nil, 0, e
	}
	if uint64(count) > uint64(len(b)-p) {
		return nil, 0, fmt.Errorf("UTF-16 length exceeds available data")
	}
	units := make([]uint16, 0, int(count))
	for p < len(b) {
		x := b[p]
		p++
		if x == 0 {
			if len(units) != int(count) {
				return nil, 0, fmt.Errorf("UTF-16 length mismatch")
			}
			return units, p, nil
		}
		var u uint16
		switch {
		case x < 128:
			u = uint16(x)
		case x >= 0xc0 && x <= 0xdf:
			if p >= len(b) || b[p]&0xc0 != 0x80 {
				return nil, 0, fmt.Errorf("invalid MUTF-8 continuation")
			}
			u = uint16(x&31)<<6 | uint16(b[p]&63)
			p++
			if u < 128 && !(x == 0xc0 && u == 0) {
				return nil, 0, fmt.Errorf("overlong MUTF-8")
			}
		case x >= 0xe0 && x <= 0xef:
			if len(b)-p < 2 || b[p]&0xc0 != 0x80 || b[p+1]&0xc0 != 0x80 {
				return nil, 0, fmt.Errorf("invalid MUTF-8 continuation")
			}
			u = uint16(x&15)<<12 | uint16(b[p]&63)<<6 | uint16(b[p+1]&63)
			p += 2
			if u < 2048 {
				return nil, 0, fmt.Errorf("overlong MUTF-8")
			}
		default:
			return nil, 0, fmt.Errorf("invalid MUTF-8 leading byte")
		}
		units = append(units, u)
		if len(units) > int(count) {
			return nil, 0, fmt.Errorf("UTF-16 length mismatch")
		}
	}
	return nil, 0, fmt.Errorf("unterminated MUTF-8 string")
}
func UTF16String(u []uint16) string { return string(utf16.Decode(u)) }
func LessUTF16(a, b string) bool {
	return CompareUTF16(utf16.Encode([]rune(a)), utf16.Encode([]rune(b))) < 0
}
func CompareUTF16(a, b []uint16) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}
