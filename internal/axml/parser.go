package axml

import (
	"encoding/binary"
	"errors"
	"fmt"
	"koru.dev/koru/internal/androidbin"
)

type Value struct {
	Namespace, Name, Text string
	Kind                  byte
	Data                  uint32
}
type Element struct {
	Name       string
	Attributes []Value
}

// Link is a <data> URL pattern found in a parsed manifest.
type Link struct {
	Scheme, Host, PathPrefix string
	AutoVerify               bool
}

// Links lists the <data> elements that declare a scheme, in document order.
// AutoVerify reports android:autoVerify on the enclosing intent-filter.
func Links(elements []Element) []Link {
	var out []Link
	verify := false
	for _, el := range elements {
		switch el.Name {
		case "intent-filter":
			verify = false
			for _, a := range el.Attributes {
				if a.Namespace == android && a.Name == "autoVerify" && a.Kind == 0x12 && a.Data != 0 {
					verify = true
				}
			}
		case "data":
			var l Link
			for _, a := range el.Attributes {
				if a.Namespace != android || a.Kind != 3 {
					continue
				}
				switch a.Name {
				case "scheme":
					l.Scheme = a.Text
				case "host":
					l.Host = a.Text
				case "pathPrefix":
					l.PathPrefix = a.Text
				}
			}
			if l.Scheme != "" {
				l.AutoVerify = verify
				out = append(out, l)
			}
		}
	}
	return out
}

func Parse(b []byte) ([]Element, error) {
	if len(b) < 8 || binary.LittleEndian.Uint16(b) != 3 || binary.LittleEndian.Uint16(b[2:]) != 8 || int(binary.LittleEndian.Uint32(b[4:])) != len(b) {
		return nil, errors.New("invalid binary XML")
	}
	var pool []string
	var elements []Element
	var stack []string
	namespaces := 0
	at := func(i uint32) (string, error) {
		if i == none {
			return "", nil
		}
		if int(i) >= len(pool) {
			return "", errors.New("XML string index out of bounds")
		}
		return pool[i], nil
	}
	for pos := 8; pos < len(b); {
		if len(b)-pos < 8 {
			return nil, errors.New("truncated XML chunk")
		}
		kind := binary.LittleEndian.Uint16(b[pos:])
		header := int(binary.LittleEndian.Uint16(b[pos+2:]))
		size := int(binary.LittleEndian.Uint32(b[pos+4:]))
		if size < header || header < 8 || size%4 != 0 || size > len(b)-pos {
			return nil, errors.New("invalid XML chunk size")
		}
		c := b[pos : pos+size]
		u := func(p int) uint32 { return binary.LittleEndian.Uint32(c[p:]) }
		switch kind {
		case 1:
			if pool != nil {
				return nil, errors.New("duplicate XML pool")
			}
			var e error
			pool, e = androidbin.ReadPool(c)
			if e != nil {
				return nil, e
			}
		case 0x180:
			if header != 8 || (size-8)%4 != 0 || (size-8)/4 > len(pool) {
				return nil, errors.New("invalid XML resource map")
			}
		case 0x100, 0x101:
			if header != 16 || size != 24 {
				return nil, errors.New("invalid namespace node")
			}
			if _, e := at(u(16)); e != nil {
				return nil, e
			}
			if _, e := at(u(20)); e != nil {
				return nil, e
			}
			if kind == 0x100 {
				namespaces++
			} else {
				namespaces--
				if namespaces < 0 {
					return nil, errors.New("unbalanced namespaces")
				}
			}
		case 0x102:
			if header != 16 || size < 36 || binary.LittleEndian.Uint16(c[24:]) != 20 || binary.LittleEndian.Uint16(c[26:]) != 20 {
				return nil, errors.New("invalid start element")
			}
			name, e := at(u(20))
			if e != nil {
				return nil, e
			}
			count := int(binary.LittleEndian.Uint16(c[28:]))
			if count > (size-36)/20 || 36+count*20 != size {
				return nil, errors.New("invalid attribute count")
			}
			el := Element{Name: name}
			for i := 0; i < count; i++ {
				p := 36 + i*20
				namespace, e := at(u(p))
				if e != nil {
					return nil, e
				}
				name, e := at(u(p + 4))
				if e != nil {
					return nil, e
				}
				if binary.LittleEndian.Uint16(c[p+12:]) != 8 || c[p+14] != 0 {
					return nil, errors.New("invalid typed value")
				}
				v := Value{Namespace: namespace, Name: name, Kind: c[p+15], Data: u(p + 16)}
				if v.Kind == 3 {
					v.Text, e = at(v.Data)
					if e != nil {
						return nil, e
					}
				}
				if _, e = at(u(p + 8)); e != nil {
					return nil, e
				}
				el.Attributes = append(el.Attributes, v)
			}
			elements = append(elements, el)
			stack = append(stack, name)
		case 0x103:
			if header != 16 || size != 24 {
				return nil, errors.New("invalid end element")
			}
			name, e := at(u(20))
			if e != nil {
				return nil, e
			}
			if len(stack) == 0 || stack[len(stack)-1] != name {
				return nil, errors.New("unbalanced XML elements")
			}
			stack = stack[:len(stack)-1]
		default:
			return nil, fmt.Errorf("unsupported XML chunk %x", kind)
		}
		pos += size
	}
	if len(stack) != 0 || namespaces != 0 || len(elements) == 0 || elements[0].Name != "manifest" {
		return nil, errors.New("incomplete manifest XML")
	}
	return elements, nil
}
