package androidbin

import (
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
)

func TestStringPool(t *testing.T) {
	s := []string{"", "hello", "中文😀", string(make([]byte, 129))}
	b, e := Pool(s)
	if e != nil {
		t.Fatal(e)
	}
	out, e := ReadPool(b)
	if e != nil || !reflect.DeepEqual(out, s) {
		t.Fatal(out, e)
	}
	for i := 0; i < len(b); i++ {
		if _, e := ReadPool(b[:i]); e == nil {
			t.Fatal(i)
		}
	}
}

// A compact malicious pool can alias one long string hundreds of times.
// Bound decoded work/memory rather than trusting compressed/input size alone.
func TestAliasedPoolDecodedBudget(t *testing.T) {
	const count = 700
	seed, err := Pool([]string{strings.Repeat("x", 32767)})
	if err != nil {
		t.Fatal(err)
	}
	data := seed[32:]
	b := make([]byte, 28+4*count+len(data))
	copy(b, seed[:28])
	binary.LittleEndian.PutUint32(b[4:], uint32(len(b)))
	binary.LittleEndian.PutUint32(b[8:], count)
	binary.LittleEndian.PutUint32(b[20:], 28+4*count)
	copy(b[28+4*count:], data)
	if _, err = ReadPool(b); err == nil || !strings.Contains(err.Error(), "decoded size limit") {
		t.Fatalf("alias amplification accepted: %v", err)
	}
}
