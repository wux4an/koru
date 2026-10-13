// Package encoding implements the DEX 32-bit LEB128 and UTF-16 based MUTF-8 formats.
package encoding

import "fmt"

func ULEB(v uint32) []byte {
	var b []byte
	for {
		x := byte(v & 127)
		v >>= 7
		if v != 0 {
			x |= 128
		}
		b = append(b, x)
		if v == 0 {
			return b
		}
	}
}
func SLEB(v int32) []byte {
	var b []byte
	for {
		x := byte(v & 127)
		v >>= 7
		done := (v == 0 && x&64 == 0) || (v == -1 && x&64 != 0)
		if !done {
			x |= 128
		}
		b = append(b, x)
		if done {
			return b
		}
	}
}
func ReadULEB(b []byte) (uint32, int, error) {
	var v uint32
	for i := 0; i < 5; i++ {
		if i >= len(b) {
			return 0, 0, fmt.Errorf("truncated ULEB128")
		}
		x := b[i]
		if i == 4 && x > 15 {
			return 0, 0, fmt.Errorf("ULEB128 overflow")
		}
		v |= uint32(x&127) << uint(7*i)
		if x&128 == 0 {
			return v, i + 1, nil
		}
	}
	return 0, 0, fmt.Errorf("ULEB128 overflow")
}
func ReadSLEB(b []byte) (int32, int, error) {
	var v int64
	for i := 0; i < 5; i++ {
		if i >= len(b) {
			return 0, 0, fmt.Errorf("truncated SLEB128")
		}
		x := b[i]
		v |= int64(x&127) << uint(7*i)
		if x&128 == 0 {
			if x&64 != 0 {
				v |= -(int64(1) << uint(7*(i+1)))
			}
			if v < -1<<31 || v > 1<<31-1 {
				return 0, 0, fmt.Errorf("SLEB128 overflow")
			}
			return int32(v), i + 1, nil
		}
	}
	return 0, 0, fmt.Errorf("SLEB128 overflow")
}
