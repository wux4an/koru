package main

import (
	"io"
	"os"
	"strings"

	"rsc.io/qr"
)

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

const quietZone = 4 // light modules around the code, as the QR specification requires

// qrMatrix encodes text and returns the module matrix with a quiet zone;
// true is a dark module.
func qrMatrix(text string) ([][]bool, error) {
	code, err := qr.Encode(text, qr.L)
	if err != nil {
		return nil, err
	}
	n := code.Size + 2*quietZone
	m := make([][]bool, n)
	for y := range m {
		m[y] = make([]bool, n)
		for x := range m[y] {
			m[y][x] = code.Black(x-quietZone, y-quietZone)
		}
	}
	return m, nil
}

// renderQR draws text as a QR code for a terminal. In a UTF-8 locale it packs
// two module rows per line with ▀▄█; otherwise it falls back to "##" and
// spaces. Terminals are usually dark, so light modules are drawn filled (a
// light code on a dark background, with its quiet zone, which phone scanners
// accept); on a light background (p.light) dark modules are filled instead.
// Scanning with a physical camera has not been tested.
func renderQR(text string, p presentation) (string, error) {
	m, err := qrMatrix(text)
	if err != nil {
		return "", err
	}
	fill := func(x, y int) bool {
		if y >= len(m) {
			return !p.light // odd row count: pad with a light module
		}
		return m[y][x] == p.light
	}
	var b strings.Builder
	if !p.utf8 {
		for y := range m {
			b.WriteString("  ")
			for x := range m[y] {
				if fill(x, y) {
					b.WriteString("##")
				} else {
					b.WriteString("  ")
				}
			}
			b.WriteByte('\n')
		}
		return b.String(), nil
	}
	for y := 0; y < len(m); y += 2 {
		b.WriteString("  ")
		for x := range m[y] {
			top, bottom := fill(x, y), fill(x, y+1)
			switch {
			case top && bottom:
				b.WriteString("█")
			case top:
				b.WriteString("▀")
			case bottom:
				b.WriteString("▄")
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
