package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestQRMatrixShape(t *testing.T) {
	m, err := qrMatrix("http://192.168.1.20:8100")
	if err != nil {
		t.Fatal(err)
	}
	n := len(m)
	if size := n - 2*quietZone; size < 21 || (size-17)%4 != 0 {
		t.Fatalf("size %d is not a QR version size", size)
	}
	for y := range m {
		if len(m[y]) != n {
			t.Fatalf("row %d has %d modules, want %d", y, len(m[y]), n)
		}
	}
	// Quiet zone is light; the three finder patterns are a 7x7 dark ring with
	// a 3x3 dark center.
	for i := 0; i < n; i++ {
		if m[0][i] || m[i][0] || m[n-1][i] || m[i][n-1] {
			t.Fatal("quiet zone has dark modules")
		}
	}
	for _, o := range [][2]int{{quietZone, quietZone}, {n - quietZone - 7, quietZone}, {quietZone, n - quietZone - 7}} {
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				want := dx == 0 || dx == 6 || dy == 0 || dy == 6 || (dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4)
				if m[o[1]+dy][o[0]+dx] != want {
					t.Fatalf("finder pattern at %v differs at %d,%d", o, dx, dy)
				}
			}
		}
	}
}

func TestRenderQR(t *testing.T) {
	m, _ := qrMatrix("http://10.0.0.5:8100")
	n := len(m)
	art, err := renderQR("http://10.0.0.5:8100", presentation{utf8: true})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(art, "\n"), "\n")
	if len(lines) != (n+1)/2 {
		t.Fatalf("%d lines for %d module rows", len(lines), n)
	}
	for _, l := range lines {
		if got := len([]rune(l)); got != n+2 {
			t.Fatalf("line width %d, want %d", got, n+2)
		}
	}
	for _, g := range []string{"▀", "▄", "█"} {
		if !strings.Contains(art, g) {
			t.Fatalf("no %q in output", g)
		}
	}
	// The first line is all quiet zone: light modules, drawn filled on dark terminals.
	if strings.TrimSpace(lines[0]) != strings.Repeat("█", n) {
		t.Fatalf("first line %q", lines[0])
	}
	light, _ := renderQR("http://10.0.0.5:8100", presentation{utf8: true, light: true})
	if light == art {
		t.Fatal("light terminals must get the inverted drawing")
	}
	ascii, _ := renderQR("http://10.0.0.5:8100", presentation{})
	if !strings.Contains(ascii, "##") || strings.ContainsAny(ascii, "▀▄█") || strings.Count(ascii, "\n") != n {
		t.Fatalf("ascii fallback: %q", ascii)
	}
}

func TestDevURLsPrintQR(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	if err := printDevURLs(&out, "192.0.2.10", 8100, "", true); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Scan to open http://192.0.2.10:8100", "█", "▀"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatal("NO_COLOR output has escape codes")
	}
	// Not a terminal and no --qr: no code. Loopback: never, even with --qr.
	out.Reset()
	if err := printDevURLs(&out, "192.0.2.10", 8100, "", false); err != nil || strings.Contains(out.String(), "Scan to open") {
		t.Fatal(out.String(), err)
	}
	out.Reset()
	if err := printDevURLs(&out, "127.0.0.1", 8100, "", true); err != nil || strings.Contains(out.String(), "Scan to open") {
		t.Fatal(out.String(), err)
	}
}
