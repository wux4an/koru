package model

import "testing"

func TestDescriptors(t *testing.T) {
	p, e := ParsePrototype("(IJ[Ljava/lang/String;)V")
	if e != nil || p.Words() != 4 || p.Shorty() != "VIJL" {
		t.Fatal(p, e)
	}
	for _, s := range []string{"", "Vx", "[V", "Lbad//Name;", "Lbad.name;", "L;", "L space;"} {
		if ValidateType(s, false) == nil {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"()", "(V)V", "(I", "()VV", "I"} {
		if _, e := ParsePrototype(s); e == nil {
			t.Fatal(s)
		}
	}
}
