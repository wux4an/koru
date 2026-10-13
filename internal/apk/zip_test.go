package apk

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestZipAlignmentDeterminismAndPayload(t *testing.T) {
	entries := []Entry{{Name: "classes.dex", Data: []byte("dex"), Store: true, Alignment: 4}, {Name: "lib/arm64-v8a/test.so", Data: []byte("native"), Store: true, Alignment: 16384}, {Name: "assets/www/index.html", Data: bytes.Repeat([]byte("hello"), 200)}}
	b, e := Pack(entries)
	if e != nil {
		t.Fatal(e)
	}
	r, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range r.File {
		off, e := f.DataOffset()
		if e != nil {
			t.Fatal(e)
		}
		for _, entry := range entries {
			if entry.Name == f.Name {
				if entry.Alignment > 0 && off%int64(entry.Alignment) != 0 {
					t.Fatal("alignment", f.Name, off)
				}
				rd, e := f.Open()
				if e != nil {
					t.Fatal(e)
				}
				data, e := io.ReadAll(rd)
				rd.Close()
				if e != nil || !bytes.Equal(data, entry.Data) {
					t.Fatal(e)
				}
			}
		}
	}
	entries[0], entries[2] = entries[2], entries[0]
	b2, e := Pack(entries)
	if e != nil || !bytes.Equal(b, b2) {
		t.Fatal("determinism", e)
	}
}
func TestRejectNames(t *testing.T) {
	for _, name := range []string{"../bad", "/bad", "x/../bad", "x\\bad", "a//b", "a\x00b"} {
		if _, e := Pack([]Entry{{Name: name}}); e == nil {
			t.Fatal(name)
		}
	}
	if _, e := Pack([]Entry{{Name: "a"}, {Name: "a"}}); e == nil {
		t.Fatal("duplicates")
	}
}
