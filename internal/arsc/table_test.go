package arsc

import "testing"

func TestTable(t *testing.T) {
	b, e := Build("com.example.koru")
	if e != nil {
		t.Fatal(e)
	}
	info, e := Parse(b)
	if e != nil || info.Package != "com.example.koru" || info.IconID != IconID || info.IconPath != "res/mipmap/ic_launcher.png" {
		t.Fatal(info, e)
	}
	for i := 0; i < len(b); i++ {
		if _, e := Parse(b[:i]); e == nil {
			t.Fatal(i)
		}
	}
	b[len(b)-1] = 255
	if _, e := Parse(b); e == nil {
		t.Fatal("corrupt table")
	}
}
func FuzzTable(f *testing.F) {
	b, _ := Build("com.example.demo")
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) { Parse(b) })
}
