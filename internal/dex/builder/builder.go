// Package builder serializes symbolic DEX input to DEX 035, using Go only.
package builder

import (
	"crypto/sha1"
	"fmt"
	"hash/adler32"
	"koru.dev/koru/internal/dex/assembler"
	"koru.dev/koru/internal/dex/encoding"
	"koru.dev/koru/internal/dex/format"
	"koru.dev/koru/internal/dex/model"
	"sort"
)

func classOrder(p model.Program) ([]model.Class, error) {
	defs := map[string]model.Class{}
	for _, c := range p.Classes {
		if _, ok := defs[c.Descriptor]; ok {
			return nil, fmt.Errorf("duplicate class %s", c.Descriptor)
		}
		defs[c.Descriptor] = c
	}
	names := make([]string, 0, len(defs))
	for s := range defs {
		names = append(names, s)
	}
	sort.Slice(names, func(i, j int) bool { return encoding.LessUTF16(names[i], names[j]) })
	var out []model.Class
	state := map[string]int{}
	var visit func(string) error
	visit = func(s string) error {
		c, ok := defs[s]
		if !ok {
			return nil
		}
		if state[s] == 1 {
			return fmt.Errorf("class dependency cycle at %s", s)
		}
		if state[s] == 2 {
			return nil
		}
		state[s] = 1
		if e := visit(c.Super); e != nil {
			return e
		}
		interfaces := append([]string(nil), c.Interfaces...)
		sort.Strings(interfaces)
		seen := map[string]bool{}
		for _, v := range interfaces {
			if seen[v] {
				return fmt.Errorf("duplicate interface")
			}
			seen[v] = true
			if e := visit(v); e != nil {
				return e
			}
		}
		state[s] = 2
		out = append(out, c)
		return nil
	}
	for _, s := range names {
		if e := visit(s); e != nil {
			return nil, e
		}
	}
	return out, nil
}
func validate(classes []model.Class) error {
	for _, c := range classes {
		if e := model.ValidateClassAccess(c.Access); e != nil {
			return e
		}
		if c.Super == "" && c.Descriptor != "Ljava/lang/Object;" {
			return fmt.Errorf("only Object may omit superclass")
		}
		seen := map[string]bool{}
		for _, f := range c.Fields {
			k := f.Name + ":" + f.Type
			if seen[k] {
				return fmt.Errorf("duplicate field %s", k)
			}
			seen[k] = true
			if e := model.ValidateFieldAccess(f.Access); e != nil {
				return e
			}
		}
		seen = map[string]bool{}
		for _, m := range c.Methods {
			k := m.Name + m.Descriptor
			if seen[k] {
				return fmt.Errorf("duplicate method %s", k)
			}
			seen[k] = true
			if e := model.ValidateMethodAccess(m.Name, m.Descriptor, m.Access, m.Code != nil); e != nil {
				return fmt.Errorf("%s->%s: %w", c.Descriptor, k, e)
			}
			if m.Access&model.Abstract != 0 && c.Access&model.Abstract == 0 {
				return fmt.Errorf("abstract method in concrete class")
			}
			if c.Access&model.Interface != 0 && m.Name != "<clinit>" && m.Access&(model.Public|model.Abstract) != (model.Public|model.Abstract) {
				return fmt.Errorf("DEX 035 interface requires public abstract methods")
			}
		}
	}
	return nil
}
func Build(p model.Program) ([]byte, error) {
	classes, e := classOrder(p)
	if e != nil {
		return nil, e
	}
	if e = validate(classes); e != nil {
		return nil, e
	}
	q, e := collect(p)
	if e != nil {
		return nil, e
	}
	w := writer{}
	w.section(format.HeaderItem, 1)
	w.bytes(make([]byte, format.HeaderSize))
	copy(w.b, format.Magic)
	so := w.table(format.StringID, len(q.strings), 4)
	to := w.table(format.TypeID, len(q.types), 4)
	po := w.table(format.ProtoID, len(q.protos), 12)
	fo := w.table(format.FieldID, len(q.fields), 8)
	mo := w.table(format.MethodID, len(q.methods), 8)
	co := w.table(format.ClassDef, len(classes), 32)
	w.align()
	dataOff := w.pos()
	w.section(format.StringData, len(q.strings))
	for i, s := range q.strings {
		w.put(so+uint32(i*4), w.pos())
		b, e := encoding.StringData(s)
		if e != nil {
			return nil, e
		}
		w.bytes(b)
	}
	// Share parameter/interface lists by their exact ordered type indices.
	listKeys := []string{}
	lists := map[string][]uint16{}
	makeList := func(types []string, ordered bool) string {
		if len(types) == 0 {
			return ""
		}
		ids := make([]uint16, len(types))
		for i, s := range types {
			ids[i] = uint16(q.ti[s])
		}
		if ordered {
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		}
		k := ""
		for _, v := range ids {
			k += fmt.Sprintf("%04x", v)
		}
		if _, ok := lists[k]; !ok {
			listKeys = append(listKeys, k)
			lists[k] = ids
		}
		return k
	}
	pk := make([]string, len(q.protos))
	for i, p := range q.protos {
		pk[i] = makeList(p.p.Parameters, false)
	}
	ck := make([]string, len(classes))
	for i, c := range classes {
		ck[i] = makeList(c.Interfaces, true)
	}
	offsets := map[string]uint32{}
	sort.Strings(listKeys)
	w.align()
	w.section(format.TypeList, len(listKeys))
	for _, k := range listKeys {
		w.align()
		offsets[k] = w.pos()
		w.u32(uint32(len(lists[k])))
		for _, v := range lists[k] {
			w.u16(v)
		}
	}
	codes := map[string]uint32{}
	codeCount := 0
	for _, c := range classes {
		for _, m := range c.Methods {
			if m.Code != nil {
				codeCount++
			}
		}
	}
	w.align()
	w.section(format.CodeItem, codeCount)
	for _, c := range classes {
		methods := append([]model.Method(nil), c.Methods...)
		sort.Slice(methods, func(i, j int) bool {
			return q.mi[model.MethodKey(model.MethodRef{Owner: c.Descriptor, Name: methods[i].Name, Descriptor: methods[i].Descriptor})] < q.mi[model.MethodKey(model.MethodRef{Owner: c.Descriptor, Name: methods[j].Name, Descriptor: methods[j].Descriptor})]
		})
		for _, m := range methods {
			if m.Code == nil {
				continue
			}
			a, e := assembler.Assemble(*m.Code, m.Descriptor, m.Access, q)
			if e != nil {
				return nil, fmt.Errorf("%s->%s: %w", c.Descriptor, m.Name, e)
			}
			w.align()
			key := model.MethodKey(model.MethodRef{Owner: c.Descriptor, Name: m.Name, Descriptor: m.Descriptor})
			codes[key] = w.pos()
			w.u16(a.Registers)
			w.u16(a.Incoming)
			w.u16(a.Outgoing)
			w.u16(uint16(len(a.Tries)))
			w.u32(0)
			w.u32(uint32(len(a.Units)))
			for _, v := range a.Units {
				w.u16(v)
			}
			if len(a.Tries) > 0 {
				if len(a.Units)%2 != 0 {
					w.u16(0)
				}
				handlers := encoding.ULEB(uint32(len(a.Tries)))
				offsets := make([]uint16, len(a.Tries))
				for i, tr := range a.Tries {
					if len(handlers) > 65535 {
						return nil, fmt.Errorf("catch handler offset overflow")
					}
					offsets[i] = uint16(len(handlers))
					if tr.CatchAll {
						handlers = append(handlers, encoding.SLEB(0)...)
					} else {
						handlers = append(handlers, encoding.SLEB(1)...)
						handlers = append(handlers, encoding.ULEB(uint32(tr.Type))...)
					}
					handlers = append(handlers, encoding.ULEB(tr.Handler)...)
				}
				for i, tr := range a.Tries {
					w.u32(tr.Start)
					w.u16(tr.Count)
					w.u16(offsets[i])
				}
				w.bytes(handlers)
			}
		}
	}
	classData := map[string]uint32{}
	dataCount := 0
	for _, c := range classes {
		if len(c.Fields)+len(c.Methods) > 0 {
			dataCount++
		}
	}
	w.section(format.ClassData, dataCount)
	type encoded struct{ idx, access, code uint32 }
	for _, c := range classes {
		if len(c.Fields)+len(c.Methods) == 0 {
			continue
		}
		classData[c.Descriptor] = w.pos()
		var sf, inf, dm, vm []encoded
		for _, f := range c.Fields {
			v := encoded{idx: q.fi[model.FieldKey(model.FieldRef{Owner: c.Descriptor, Name: f.Name, Type: f.Type})], access: f.Access}
			if f.Access&model.Static != 0 {
				sf = append(sf, v)
			} else {
				inf = append(inf, v)
			}
		}
		for _, m := range c.Methods {
			k := model.MethodKey(model.MethodRef{Owner: c.Descriptor, Name: m.Name, Descriptor: m.Descriptor})
			v := encoded{q.mi[k], m.Access, codes[k]}
			if model.IsDirect(m.Access) {
				dm = append(dm, v)
			} else {
				vm = append(vm, v)
			}
		}
		for _, v := range [][]encoded{sf, inf, dm, vm} {
			w.leb(uint32(len(v)))
		}
		for i, v := range [][]encoded{sf, inf, dm, vm} {
			sort.Slice(v, func(i, j int) bool { return v[i].idx < v[j].idx })
			var prev uint32
			for _, x := range v {
				w.leb(x.idx - prev)
				w.leb(x.access)
				if i >= 2 {
					w.leb(x.code)
				}
				prev = x.idx
			}
		}
	}
	for i, s := range q.types {
		w.put(to+uint32(i*4), q.si[s])
	}
	for i, p := range q.protos {
		base := po + uint32(i*12)
		w.put(base, q.si[p.p.Shorty()])
		w.put(base+4, q.ti[p.p.Return])
		w.put(base+8, offsets[pk[i]])
	}
	for i, f := range q.fields {
		base := fo + uint32(i*8)
		w.put(base, q.ti[f.Owner]|q.ti[f.Type]<<16)
		w.put(base+4, q.si[f.Name])
	}
	for i, m := range q.methods {
		base := mo + uint32(i*8)
		w.put(base, q.ti[m.Owner]|q.pi[m.Descriptor]<<16)
		w.put(base+4, q.si[m.Name])
	}
	for i, c := range classes {
		base := co + uint32(i*32)
		w.put(base, q.ti[c.Descriptor])
		w.put(base+4, c.Access)
		super := format.NoIndex
		if c.Super != "" {
			super = q.ti[c.Super]
		}
		w.put(base+8, super)
		w.put(base+12, offsets[ck[i]])
		source := format.NoIndex
		if c.Source != "" {
			source = q.si[c.Source]
		}
		w.put(base+16, source)
		w.put(base+24, classData[c.Descriptor])
	}
	w.align()
	mapOff := w.pos()
	w.section(format.MapList, 1)
	w.u32(uint32(len(w.sections)))
	for _, s := range w.sections {
		w.u16(s.Type)
		w.u16(0)
		w.u32(s.Size)
		w.u32(s.Offset)
	}
	w.put(32, w.pos())
	w.put(36, format.HeaderSize)
	w.put(40, format.Endian)
	w.put(52, mapOff)
	for i, t := range []struct {
		n   int
		off uint32
	}{{len(q.strings), so}, {len(q.types), to}, {len(q.protos), po}, {len(q.fields), fo}, {len(q.methods), mo}, {len(classes), co}} {
		w.put(56+uint32(i*8), uint32(t.n))
		w.put(60+uint32(i*8), t.off)
	}
	w.put(104, w.pos()-dataOff)
	w.put(108, dataOff)
	if w.err != nil {
		return nil, w.err
	}
	sig := sha1.Sum(w.b[32:])
	copy(w.b[12:32], sig[:])
	w.put(8, adler32.Checksum(w.b[12:]))
	return w.b, nil
}
