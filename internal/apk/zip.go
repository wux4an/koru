// Package apk implements deterministic ZIP32 packaging with Android alignment.
package apk

import (
	"bytes"
	"compress/flate"
	"errors"
	"hash/crc32"
	"koru.dev/koru/internal/androidbin"
	"sort"
	"strings"
	"unicode/utf8"
)

type Entry struct {
	Name      string
	Data      []byte
	Store     bool
	Alignment uint32
}

func SafeName(s string) bool {
	if s == "" || !utf8.ValidString(s) || strings.ContainsAny(s, "\\\x00:") || strings.HasPrefix(s, "/") || strings.HasSuffix(s, "/") {
		return false
	}
	for _, p := range strings.Split(s, "/") {
		if p == "" || p == "." || p == ".." {
			return false
		}
	}
	return true
}
func Pack(entries []Entry) ([]byte, error) {
	entries = append([]Entry(nil), entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	if len(entries) >= 65535 {
		return nil, errors.New("ZIP32 entry limit")
	}
	var result, central []byte
	seen := map[string]bool{}
	for _, e := range entries {
		if !SafeName(e.Name) || len(e.Name) > 65535 || seen[e.Name] {
			return nil, errors.New("unsafe or duplicate APK entry")
		}
		seen[e.Name] = true
		if uint64(len(e.Data)) > 0xffffffff {
			return nil, errors.New("ZIP32 size limit")
		}
		raw := e.Data
		method := uint16(0)
		if !e.Store {
			method = 8
			var buf bytes.Buffer
			w, err := flate.NewWriter(&buf, flate.DefaultCompression)
			if err != nil {
				return nil, err
			}
			if _, err = w.Write(raw); err != nil {
				return nil, err
			}
			if err = w.Close(); err != nil {
				return nil, err
			}
			raw = buf.Bytes()
		}
		align := e.Alignment
		if align == 0 {
			align = 1
		}
		if align > 16384 || align&(align-1) != 0 || !e.Store && align > 1 {
			return nil, errors.New("invalid ZIP alignment")
		}
		extra := []byte{}
		if align > 1 {
			pad := (align - (uint32(len(result))+30+uint32(len(e.Name))+6)%align) % align
			extra = androidbin.U16(extra, 0xd935)
			extra = androidbin.U16(extra, uint16(pad+2))
			extra = androidbin.U16(extra, uint16(align))
			extra = append(extra, make([]byte, pad)...)
		}
		offset := uint32(len(result))
		crc := crc32.ChecksumIEEE(e.Data)
		h := androidbin.U32(nil, 0x04034b50)
		h = androidbin.U16(h, 20)
		h = androidbin.U16(h, 0x800)
		h = androidbin.U16(h, method)
		h = androidbin.U16(h, 0)
		h = androidbin.U16(h, 33)
		h = androidbin.U32(h, crc)
		h = androidbin.U32(h, uint32(len(raw)))
		h = androidbin.U32(h, uint32(len(e.Data)))
		h = androidbin.U16(h, uint16(len(e.Name)))
		h = androidbin.U16(h, uint16(len(extra)))
		result = append(result, h...)
		result = append(result, []byte(e.Name)...)
		result = append(result, extra...)
		result = append(result, raw...)
		c := androidbin.U32(nil, 0x02014b50)
		c = androidbin.U16(c, 0x0314)
		c = androidbin.U16(c, 20)
		c = androidbin.U16(c, 0x800)
		c = androidbin.U16(c, method)
		c = androidbin.U16(c, 0)
		c = androidbin.U16(c, 33)
		c = androidbin.U32(c, crc)
		c = androidbin.U32(c, uint32(len(raw)))
		c = androidbin.U32(c, uint32(len(e.Data)))
		c = androidbin.U16(c, uint16(len(e.Name)))
		c = androidbin.U16(c, 0)
		c = androidbin.U16(c, 0)
		c = androidbin.U16(c, 0)
		c = androidbin.U16(c, 0)
		c = androidbin.U32(c, 0100644<<16)
		c = androidbin.U32(c, offset)
		central = append(central, c...)
		central = append(central, []byte(e.Name)...)
		if uint64(len(result))+uint64(len(central))+22 > 0xffffffff {
			return nil, errors.New("ZIP32 archive limit")
		}
	}
	cd := uint32(len(result))
	result = append(result, central...)
	end := androidbin.U32(nil, 0x06054b50)
	end = androidbin.U16(end, 0)
	end = androidbin.U16(end, 0)
	end = androidbin.U16(end, uint16(len(entries)))
	end = androidbin.U16(end, uint16(len(entries)))
	end = androidbin.U32(end, uint32(len(central)))
	end = androidbin.U32(end, cd)
	end = androidbin.U16(end, 0)
	return append(result, end...), nil
}
