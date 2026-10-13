package parser

import (
	"encoding/binary"
	"fmt"
	"koru.dev/koru/internal/dex/format"
	"koru.dev/koru/internal/dex/model"
)

func (f *File) classes(b []byte) error {
	u := func(p uint32) uint32 { return binary.LittleEndian.Uint32(b[p : p+4]) }
	count, off := u(96), u(100)
	defined := map[uint32]int{}
	usedData, usedCode := map[uint32]bool{}, map[uint32]bool{}
	for i := uint32(0); i < count; i++ {
		base := off + i*32
		c := Class{Type: u(base), Access: u(base + 4), Super: u(base + 8), Source: u(base + 16)}
		if c.Type >= uint32(len(f.Types)) || !model.IsClass(f.Types[c.Type]) {
			return fmt.Errorf("invalid class type")
		}
		if _, ok := defined[c.Type]; ok {
			return fmt.Errorf("duplicate class definition")
		}
		defined[c.Type] = int(i)
		if e := model.ValidateClassAccess(c.Access); e != nil {
			return e
		}
		if c.Super == format.NoIndex {
			if f.Types[c.Type] != "Ljava/lang/Object;" {
				return fmt.Errorf("missing superclass")
			}
		} else if c.Super >= uint32(len(f.Types)) || !model.IsClass(f.Types[c.Super]) || c.Super == c.Type {
			return fmt.Errorf("invalid superclass")
		}
		if c.Source != format.NoIndex && c.Source >= uint32(len(f.Strings)) {
			return fmt.Errorf("invalid source index")
		}
		if u(base+20) != 0 || u(base+28) != 0 {
			return fmt.Errorf("annotations/static values unsupported")
		}
		if io := u(base + 12); io != 0 {
			var ok bool
			c.Interfaces, ok = f.lists[io]
			if !ok {
				return fmt.Errorf("invalid interfaces offset")
			}
			for j, v := range c.Interfaces {
				if !model.IsClass(f.Types[v]) || j > 0 && v <= c.Interfaces[j-1] || uint32(v) == c.Type {
					return fmt.Errorf("invalid or duplicate interface")
				}
			}
		}
		do := u(base + 24)
		if do != 0 {
			d, ok := f.classData[do]
			if !ok || usedData[do] {
				return fmt.Errorf("invalid/shared class_data offset")
			}
			usedData[do] = true
			seenFields, seenMethods := map[uint32]bool{}, map[uint32]bool{}
			for group, fields := range d.fields {
				for _, v := range fields {
					if seenFields[v.Index] || uint32(f.Fields[v.Index].Owner) != c.Type {
						return fmt.Errorf("invalid field owner or duplicate")
					}
					seenFields[v.Index] = true
					if e := model.ValidateFieldAccess(v.Access); e != nil {
						return e
					}
					if (v.Access&model.Static != 0) != (group == 0) {
						return fmt.Errorf("incorrect field classification")
					}
					c.Fields = append(c.Fields, v)
				}
			}
			for group, methods := range d.methods {
				for _, v := range methods {
					if seenMethods[v.index] || uint32(f.Methods[v.index].Owner) != c.Type {
						return fmt.Errorf("invalid method owner or duplicate")
					}
					seenMethods[v.index] = true
					m := f.Methods[v.index]
					name := f.Strings[m.Name]
					if e := model.ValidateMethodAccess(name, f.descriptor(m.Prototype), v.access, v.offset != 0); e != nil {
						return e
					}
					if model.IsDirect(v.access) != (group == 0) {
						return fmt.Errorf("incorrect method classification")
					}
					if v.access&model.Abstract != 0 && c.Access&model.Abstract == 0 {
						return fmt.Errorf("abstract method in concrete class")
					}
					if c.Access&model.Interface != 0 && name != "<clinit>" && v.access&(model.Public|model.Abstract) != (model.Public|model.Abstract) {
						return fmt.Errorf("DEX 035 interface method flags")
					}
					cm := Method{Index: v.index, Access: v.access}
					if v.offset != 0 {
						code, ok := f.codes[v.offset]
						if !ok || usedCode[v.offset] {
							return fmt.Errorf("invalid/shared code offset")
						}
						usedCode[v.offset] = true
						pr, _ := model.ParsePrototype(f.descriptor(m.Prototype))
						words := pr.Words()
						if v.access&model.Static == 0 {
							words++
						}
						if words != uint32(code.Incoming) {
							return fmt.Errorf("incoming parameter words mismatch")
						}
						cm.Code = code
					}
					c.Methods = append(c.Methods, cm)
				}
			}
		}
		f.Classes = append(f.Classes, c)
	}
	for i, c := range f.Classes {
		dependencies := []uint32{c.Super}
		for _, v := range c.Interfaces {
			dependencies = append(dependencies, uint32(v))
		}
		for _, v := range dependencies {
			if j, ok := defined[v]; ok && j >= i {
				return fmt.Errorf("class dependency order/cycle")
			}
		}
	}
	if len(usedData) != len(f.classData) || len(usedCode) != len(f.codes) {
		return fmt.Errorf("unreferenced code or class data")
	}
	for _, c := range f.Classes {
		for _, m := range c.Methods {
			if m.Code != nil {
				if e := f.verifyCode(m); e != nil {
					return fmt.Errorf("%s->%s: %w", f.Types[c.Type], f.Strings[f.Methods[m.Index].Name], e)
				}
			}
		}
	}
	return nil
}
