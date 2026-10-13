package apk

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/signing"
	"strings"
)

func ReadVerified(b []byte) (map[string][]byte, []byte, error) {
	cert, e := signing.VerifyV2(b)
	if e != nil {
		return nil, nil, e
	}
	cd, _, e := signing.Layout(b)
	if e != nil {
		return nil, nil, e
	}
	blockSize := binary.LittleEndian.Uint64(b[cd-24:])
	limit := uint64(cd) - blockSize - 8
	r, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		return nil, nil, e
	}
	files := map[string][]byte{}
	total := uint64(0)
	type interval struct{ start, end uint64 }
	var spans []interval
	for _, f := range r.File {
		if !SafeName(f.Name) || files[f.Name] != nil || f.Flags != 0x800 || (f.Method != 0 && f.Method != 8) || f.UncompressedSize64 > 64<<20 {
			return nil, nil, errors.New("unsupported, unsafe or duplicate APK ZIP entry")
		}
		total += f.UncompressedSize64
		if total > 256<<20 {
			return nil, nil, errors.New("APK inflated size limit")
		}
		off, e := f.DataOffset()
		if e != nil || off < 30 || uint64(off)+f.CompressedSize64 > limit {
			return nil, nil, errors.New("ZIP data overlaps signing block")
		}
		if f.Method == 0 && off%4 != 0 {
			return nil, nil, fmt.Errorf("stored entry %s is not aligned", f.Name)
		}
		if strings.HasPrefix(f.Name, "lib/") && f.Method == 0 && off%16384 != 0 {
			return nil, nil, errors.New("native library lacks 16 KiB alignment")
		}
		spans = append(spans, interval{uint64(off), uint64(off) + f.CompressedSize64})
		rd, e := f.Open()
		if e != nil {
			return nil, nil, e
		}
		data, e := io.ReadAll(io.LimitReader(rd, int64(f.UncompressedSize64)+1))
		closeErr := rd.Close()
		if e != nil {
			return nil, nil, e
		}
		if closeErr != nil {
			return nil, nil, closeErr
		}
		if uint64(len(data)) != f.UncompressedSize64 {
			return nil, nil, errors.New("ZIP uncompressed size mismatch")
		}
		files[f.Name] = data
	}
	for i, a := range spans {
		for _, c := range spans[i+1:] {
			if a.start < c.end && c.start < a.end {
				return nil, nil, errors.New("overlapping ZIP entries")
			}
		}
	}
	return files, cert, nil
}
