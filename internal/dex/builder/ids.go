package builder

import (
	"fmt"
	"koru.dev/koru/internal/dex/encoding"
	"koru.dev/koru/internal/dex/model"
	"sort"
	"unicode/utf8"
)

type proto struct {
	descriptor string
	p          model.Prototype
}
type pool struct {
	strings, types     []string
	protos             []proto
	fields             []model.FieldRef
	methods            []model.MethodRef
	si, ti, pi, fi, mi map[string]uint32
}

func collect(p model.Program) (*pool, error) {
	q := &pool{si: map[string]uint32{}, ti: map[string]uint32{}, pi: map[string]uint32{}, fi: map[string]uint32{}, mi: map[string]uint32{}}
	addString := func(s string) error {
		if !utf8.ValidString(s) {
			return fmt.Errorf("invalid UTF-8 string")
		}
		q.si[s] = 0
		return nil
	}
	addType := func(s string) error {
		if e := model.ValidateType(s, true); e != nil {
			return e
		}
		q.ti[s] = 0
		return addString(s)
	}
	addMethod := func(m model.MethodRef) error {
		if e := model.ValidateMethodRef(m); e != nil {
			return e
		}
		if e := addType(m.Owner); e != nil {
			return e
		}
		addString(m.Name)
		if _, ok := q.mi[model.MethodKey(m)]; !ok {
			q.methods = append(q.methods, m)
			q.mi[model.MethodKey(m)] = 0
		}
		if _, ok := q.pi[m.Descriptor]; !ok {
			pr, _ := model.ParsePrototype(m.Descriptor)
			q.protos = append(q.protos, proto{m.Descriptor, pr})
			q.pi[m.Descriptor] = 0
			addString(pr.Shorty())
			addType(pr.Return)
			for _, s := range pr.Parameters {
				addType(s)
			}
		}
		return nil
	}
	addField := func(f model.FieldRef) error {
		if e := model.ValidateFieldRef(f); e != nil {
			return e
		}
		addType(f.Owner)
		addType(f.Type)
		addString(f.Name)
		if _, ok := q.fi[model.FieldKey(f)]; !ok {
			q.fields = append(q.fields, f)
			q.fi[model.FieldKey(f)] = 0
		}
		return nil
	}
	for _, c := range p.Classes {
		if !model.IsClass(c.Descriptor) {
			return nil, fmt.Errorf("invalid class %q", c.Descriptor)
		}
		addType(c.Descriptor)
		if c.Super != "" {
			if !model.IsClass(c.Super) {
				return nil, fmt.Errorf("invalid superclass")
			}
			addType(c.Super)
		}
		for _, s := range c.Interfaces {
			if !model.IsClass(s) {
				return nil, fmt.Errorf("invalid interface")
			}
			addType(s)
		}
		if e := addString(c.Source); e != nil {
			return nil, e
		}
		for _, f := range c.Fields {
			if e := addField(model.FieldRef{Owner: c.Descriptor, Name: f.Name, Type: f.Type}); e != nil {
				return nil, e
			}
		}
		for _, m := range c.Methods {
			if e := addMethod(model.MethodRef{Owner: c.Descriptor, Name: m.Name, Descriptor: m.Descriptor}); e != nil {
				return nil, e
			}
			if m.Code != nil {
				for _, tr := range m.Code.Tries {
					if tr.Type != "" {
						if e := addType(tr.Type); e != nil {
							return nil, e
						}
					}
				}
				for _, i := range m.Code.Instructions {
					if i.Op == "const-string" {
						if e := addString(i.String); e != nil {
							return nil, e
						}
					}
					if i.Type != "" {
						if e := addType(i.Type); e != nil {
							return nil, e
						}
					}
					if i.Method != nil {
						if e := addMethod(*i.Method); e != nil {
							return nil, e
						}
					}
					if i.Field != nil {
						if e := addField(*i.Field); e != nil {
							return nil, e
						}
					}
				}
			}
		}
	}
	// Empty source file names are not emitted unless used as actual constants.
	usedEmpty := false
	for _, c := range p.Classes {
		for _, m := range c.Methods {
			if m.Code != nil {
				for _, tr := range m.Code.Tries {
					if tr.Type != "" {
						if e := addType(tr.Type); e != nil {
							return nil, e
						}
					}
				}
				for _, i := range m.Code.Instructions {
					usedEmpty = usedEmpty || (i.Op == "const-string" && i.String == "")
				}
			}
		}
	}
	if !usedEmpty {
		delete(q.si, "")
	}
	for s := range q.si {
		q.strings = append(q.strings, s)
	}
	sort.Slice(q.strings, func(i, j int) bool { return encoding.LessUTF16(q.strings[i], q.strings[j]) })
	for i, s := range q.strings {
		q.si[s] = uint32(i)
	}
	for s := range q.ti {
		q.types = append(q.types, s)
	}
	sort.Slice(q.types, func(i, j int) bool { return q.si[q.types[i]] < q.si[q.types[j]] })
	for i, s := range q.types {
		q.ti[s] = uint32(i)
	}
	sort.Slice(q.protos, func(i, j int) bool {
		a, b := q.protos[i].p, q.protos[j].p
		if q.ti[a.Return] != q.ti[b.Return] {
			return q.ti[a.Return] < q.ti[b.Return]
		}
		for k := 0; k < len(a.Parameters) && k < len(b.Parameters); k++ {
			if q.ti[a.Parameters[k]] != q.ti[b.Parameters[k]] {
				return q.ti[a.Parameters[k]] < q.ti[b.Parameters[k]]
			}
		}
		return len(a.Parameters) < len(b.Parameters)
	})
	for i, p := range q.protos {
		q.pi[p.descriptor] = uint32(i)
	}
	sort.Slice(q.fields, func(i, j int) bool {
		a, b := q.fields[i], q.fields[j]
		if q.ti[a.Owner] != q.ti[b.Owner] {
			return q.ti[a.Owner] < q.ti[b.Owner]
		}
		if q.si[a.Name] != q.si[b.Name] {
			return q.si[a.Name] < q.si[b.Name]
		}
		return q.ti[a.Type] < q.ti[b.Type]
	})
	for i, f := range q.fields {
		q.fi[model.FieldKey(f)] = uint32(i)
	}
	sort.Slice(q.methods, func(i, j int) bool {
		a, b := q.methods[i], q.methods[j]
		if q.ti[a.Owner] != q.ti[b.Owner] {
			return q.ti[a.Owner] < q.ti[b.Owner]
		}
		if q.si[a.Name] != q.si[b.Name] {
			return q.si[a.Name] < q.si[b.Name]
		}
		return q.pi[a.Descriptor] < q.pi[b.Descriptor]
	})
	for i, m := range q.methods {
		q.mi[model.MethodKey(m)] = uint32(i)
	}
	if len(q.types) > 65535 || len(q.protos) > 65535 || len(q.fields) > 65536 || len(q.methods) > 65536 {
		return nil, fmt.Errorf("DEX single-file identifier limit exceeded")
	}
	if uint64(len(q.strings)) > 0xffffffff {
		return nil, fmt.Errorf("string identifier overflow")
	}
	return q, nil
}
func (q *pool) String(s string) (uint32, error) {
	v, ok := q.si[s]
	if !ok {
		return 0, fmt.Errorf("unresolved string")
	}
	return v, nil
}
func (q *pool) Type(s string) (uint16, error) {
	v, ok := q.ti[s]
	if !ok {
		return 0, fmt.Errorf("unresolved type")
	}
	return uint16(v), nil
}
func (q *pool) Method(m model.MethodRef) (uint16, error) {
	v, ok := q.mi[model.MethodKey(m)]
	if !ok {
		return 0, fmt.Errorf("unresolved method")
	}
	return uint16(v), nil
}
func (q *pool) Field(f model.FieldRef) (uint16, error) {
	v, ok := q.fi[model.FieldKey(f)]
	if !ok {
		return 0, fmt.Errorf("unresolved field")
	}
	return uint16(v), nil
}
