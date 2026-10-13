package model

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Prototype struct {
	Return     string
	Parameters []string
}

func ValidName(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '$' || r == '-' || r == '_' || (r >= 0xa1 && r <= 0x1fff) || (r >= 0x2010 && r <= 0x2027) || (r >= 0x2030 && r <= 0xd7ff) || (r >= 0xe000 && r <= 0xffef) || (r >= 0x10000 && r <= 0x10ffff)) {
			return false
		}
	}
	return true
}
func typePrefix(s string, allowVoid bool) (int, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty type")
	}
	p := 0
	for p < len(s) && s[p] == '[' {
		p++
	}
	if p > 255 || p == len(s) {
		return 0, fmt.Errorf("invalid array descriptor %q", s)
	}
	switch s[p] {
	case 'V':
		if p != 0 || !allowVoid {
			return 0, fmt.Errorf("void is not a value type")
		}
		return 1, nil
	case 'Z', 'B', 'S', 'C', 'I', 'J', 'F', 'D':
		return p + 1, nil
	case 'L':
		end := strings.IndexByte(s[p:], ';')
		if end < 0 {
			return 0, fmt.Errorf("unterminated class type %q", s)
		}
		end += p
		for _, part := range strings.Split(s[p+1:end], "/") {
			if !ValidName(part) {
				return 0, fmt.Errorf("invalid class name %q", s)
			}
		}
		return end + 1, nil
	}
	return 0, fmt.Errorf("invalid descriptor %q", s)
}
func ValidateType(s string, allowVoid bool) error {
	n, e := typePrefix(s, allowVoid)
	if e != nil {
		return e
	}
	if n != len(s) {
		return fmt.Errorf("trailing descriptor data %q", s)
	}
	return nil
}
func IsClass(s string) bool { return len(s) > 2 && s[0] == 'L' && ValidateType(s, false) == nil }
func ParsePrototype(s string) (Prototype, error) {
	var p Prototype
	if len(s) < 3 || s[0] != '(' {
		return p, fmt.Errorf("invalid method descriptor %q", s)
	}
	pos := 1
	for pos < len(s) && s[pos] != ')' {
		n, e := typePrefix(s[pos:], false)
		if e != nil {
			return p, e
		}
		p.Parameters = append(p.Parameters, s[pos:pos+n])
		pos += n
	}
	if pos >= len(s) || s[pos] != ')' {
		return p, fmt.Errorf("missing ')' in %q", s)
	}
	p.Return = s[pos+1:]
	if e := ValidateType(p.Return, true); e != nil {
		return p, e
	}
	return p, nil
}
func shortType(s string) byte {
	if s[0] == 'L' || s[0] == '[' {
		return 'L'
	}
	return s[0]
}
func (p Prototype) Shorty() string {
	b := []byte{shortType(p.Return)}
	for _, s := range p.Parameters {
		b = append(b, shortType(s))
	}
	return string(b)
}
func (p Prototype) Words() uint32 {
	var n uint32
	for _, s := range p.Parameters {
		n++
		if s == "J" || s == "D" {
			n++
		}
	}
	return n
}
func MethodKey(m MethodRef) string { return m.Owner + "->" + m.Name + m.Descriptor }
func FieldKey(f FieldRef) string   { return f.Owner + "->" + f.Name + ":" + f.Type }
func ValidateMethodRef(m MethodRef) error {
	if !(IsClass(m.Owner) || (strings.HasPrefix(m.Owner, "[") && ValidateType(m.Owner, false) == nil)) {
		return fmt.Errorf("invalid method owner %q", m.Owner)
	}
	if !ValidName(m.Name) && m.Name != "<init>" && m.Name != "<clinit>" {
		return fmt.Errorf("invalid method name %q", m.Name)
	}
	p, e := ParsePrototype(m.Descriptor)
	if e != nil {
		return e
	}
	if (m.Name == "<init>" || m.Name == "<clinit>") && p.Return != "V" {
		return fmt.Errorf("initializer must return void")
	}
	if m.Name == "<clinit>" && len(p.Parameters) > 0 {
		return fmt.Errorf("class initializer has parameters")
	}
	return nil
}
func ValidateFieldRef(f FieldRef) error {
	if !IsClass(f.Owner) || !ValidName(f.Name) {
		return fmt.Errorf("invalid field %q", FieldKey(f))
	}
	return ValidateType(f.Type, false)
}
