// Package verify type-checks symbolic DEX code the way ART's verifier does at
// install time, for the subset of opcodes koru emits. A method that ART would
// reject crashes the app with VerifyError as soon as its class loads, so the
// runtime is checked here, before any device runs it.
//
// It tracks the kind of every register (int, reference, null constant,
// uninitialized object, wide pair half) along all control-flow paths,
// including exception handlers, and checks each use: invoke arity and argument
// kinds, move-result after a matching invoke, field and return kinds, object
// initialization before use, and branch operands. It does not know the class
// hierarchy, so it cannot prove that a reference has the right class; that
// still needs a device.
package verify

import (
	"fmt"
	"strings"

	"github.com/wux4an/koru/internal/dex/model"
)

type kind int

const (
	undefined kind = iota
	integer        // int, boolean, byte, char, short, float bits
	zero           // const 0: valid as int and as null
	ref
	uninit // new-instance not yet initialized; site identifies the allocation
	wideLo
	wideHi
	conflict
)

func (k kind) String() string {
	return [...]string{"undefined", "int", "zero", "reference", "uninitialized", "wide-low", "wide-high", "conflicting"}[k]
}

type slot struct {
	kind kind
	site int // allocation site of an uninit value; -1 is `this` in a constructor
}

type state struct {
	regs []slot
	// result is the kind the previous instruction left for move-result:
	// 'V' none, 'I' scalar, 'L' reference, 'J' wide.
	result byte
}

func (s state) clone() state {
	return state{regs: append([]slot(nil), s.regs...), result: s.result}
}

func merge(a, b slot) slot {
	switch {
	case a == b:
		return a
	case a.kind == zero && b.kind == integer, a.kind == integer && b.kind == zero:
		return slot{kind: integer}
	case a.kind == zero && b.kind == ref, a.kind == ref && b.kind == zero:
		return slot{kind: ref}
	}
	return slot{kind: conflict}
}

// Program verifies every method with code.
func Program(p model.Program) error {
	for _, c := range p.Classes {
		for _, m := range c.Methods {
			if m.Code == nil {
				continue
			}
			if err := Method(c.Descriptor, m); err != nil {
				return err
			}
		}
	}
	return nil
}

// Method verifies one method of class owner.
func Method(owner string, m model.Method) error {
	where := owner + "->" + m.Name + m.Descriptor
	fail := func(i int, format string, args ...any) error {
		return fmt.Errorf("verify %s: instruction %d: %s", where, i, fmt.Sprintf(format, args...))
	}
	proto, err := model.ParsePrototype(m.Descriptor)
	if err != nil {
		return fmt.Errorf("verify %s: %w", where, err)
	}
	code := m.Code
	static := m.Access&model.Static != 0
	words := int(proto.Words())
	if !static {
		words++
	}
	locals := int(code.Locals)
	total := locals + words
	index := func(r model.Register) (int, bool) {
		var n int
		if r >= 0 {
			n = int(r)
		} else {
			n = locals + int(-r-1)
		}
		return n, n >= 0 && n < total
	}

	entry := state{regs: make([]slot, total), result: 'V'}
	w := locals
	if !static {
		this := slot{kind: ref}
		if m.Name == "<init>" {
			this = slot{kind: uninit, site: -1}
		}
		entry.regs[w] = this
		w++
	}
	for _, p := range proto.Parameters {
		switch p[0] {
		case 'J', 'D':
			entry.regs[w], entry.regs[w+1] = slot{kind: wideLo}, slot{kind: wideHi}
			w += 2
		case 'L', '[':
			entry.regs[w] = slot{kind: ref}
			w++
		default:
			entry.regs[w] = slot{kind: integer}
			w++
		}
	}

	ins := code.Instructions
	labels := map[string]int{}
	for i, in := range ins {
		if in.Op == "label" {
			labels[in.Label] = i
		}
	}
	target := func(i int, name string) (int, error) {
		t, ok := labels[name]
		if !ok {
			return 0, fail(i, "unknown label %q", name)
		}
		return t, nil
	}
	var handlers []handler
	for _, t := range code.Tries {
		s, ok1 := labels[t.Start]
		e, ok2 := labels[t.End]
		h, ok3 := labels[t.Handler]
		if !ok1 || !ok2 || !ok3 {
			return fmt.Errorf("verify %s: try uses an unknown label", where)
		}
		handlers = append(handlers, handler{s, e, h})
	}

	states := make([]*state, len(ins)+1)
	var work []int
	visit := func(at int, s state) error {
		if at >= len(ins) {
			return fail(at, "execution falls off the end of the method")
		}
		if states[at] == nil {
			c := s.clone()
			states[at] = &c
			work = append(work, at)
			return nil
		}
		old := states[at]
		changed := false
		for r := range old.regs {
			if v := merge(old.regs[r], s.regs[r]); v != old.regs[r] {
				old.regs[r] = v
				changed = true
			}
		}
		if old.result != s.result && old.result != 0 {
			old.result = 0 // differs between paths: no move-result is valid here
			changed = true
		}
		if changed {
			work = append(work, at)
		}
		return nil
	}
	if err := visit(0, entry); err != nil {
		return err
	}

	for len(work) > 0 {
		i := work[len(work)-1]
		work = work[:len(work)-1]
		in := ins[i]
		s := states[i].clone()

		// Exception handlers see the registers as they are before a throwing
		// instruction; move-exception then defines its own register.
		if in.Op != "label" {
			for _, h := range handlers {
				if i > h.start && i < h.end || i == h.start {
					hs := s.clone()
					hs.result = 'V'
					if err := visit(h.at, hs); err != nil {
						return err
					}
				}
			}
		}

		reg := func(n int) (int, error) {
			if n >= len(in.Regs) {
				return 0, fail(i, "%s is missing a register", in.Op)
			}
			r, ok := index(in.Regs[n])
			if !ok {
				return 0, fail(i, "%s uses register %d outside the method's %d registers", in.Op, in.Regs[n], total)
			}
			return r, nil
		}
		use := func(n int, want ...kind) (int, error) {
			r, err := reg(n)
			if err != nil {
				return 0, err
			}
			got := s.regs[r].kind
			for _, k := range want {
				if got == k {
					return r, nil
				}
			}
			names := make([]string, len(want))
			for j, k := range want {
				names[j] = k.String()
			}
			return 0, fail(i, "%s reads register %d as %s, but it holds %s", in.Op, in.Regs[n], strings.Join(names, " or "), got)
		}
		def := func(n int, v slot) error {
			r, err := reg(n)
			if err != nil {
				return err
			}
			// Overwriting half of a wide pair invalidates the other half.
			if s.regs[r].kind == wideLo && r+1 < total {
				s.regs[r+1] = slot{kind: undefined}
			}
			if s.regs[r].kind == wideHi && r > 0 {
				s.regs[r-1] = slot{kind: undefined}
			}
			s.regs[r] = v
			return nil
		}
		scalarKind := func(t string) []kind {
			if t[0] == 'L' || t[0] == '[' {
				return []kind{ref, zero}
			}
			return []kind{integer, zero}
		}
		valueOf := func(t string) slot {
			if t[0] == 'L' || t[0] == '[' {
				return slot{kind: ref}
			}
			return slot{kind: integer}
		}
		next := true
		result := byte('V')
		var err error
		switch in.Op {
		case "label", "nop":
			result = s.result
		case "return-void":
			next = false
		case "return":
			_, err = use(0, integer, zero)
			next = false
		case "return-object":
			_, err = use(0, ref, zero)
			next = false
		case "move":
			var r int
			if r, err = use(1, integer, zero); err == nil {
				err = def(0, s.regs[r])
			}
		case "move-object":
			var r int
			if r, err = use(1, ref, zero); err == nil {
				err = def(0, s.regs[r])
			}
		case "move-result", "move-result-object":
			want := map[string]byte{"move-result": 'I', "move-result-object": 'L'}[in.Op]
			if s.result != want {
				return fail(i, "%s does not follow an invoke that returns a %s (previous result: %c)", in.Op, map[byte]string{'I': "scalar", 'L': "reference"}[want], s.result)
			}
			if want == 'L' {
				err = def(0, slot{kind: ref})
			} else {
				err = def(0, slot{kind: integer})
			}
		case "move-exception":
			err = def(0, slot{kind: ref})
		case "const", "const/4", "const/16":
			if in.Literal == 0 {
				err = def(0, slot{kind: zero})
			} else {
				err = def(0, slot{kind: integer})
			}
		case "const-string", "const-class":
			err = def(0, slot{kind: ref})
		case "new-instance":
			err = def(0, slot{kind: uninit, site: i})
		case "check-cast":
			_, err = use(0, ref, zero)
		case "instance-of":
			if _, err = use(1, ref, zero); err == nil {
				err = def(0, slot{kind: integer})
			}
		case "new-array":
			if _, err = use(1, integer, zero); err == nil {
				err = def(0, slot{kind: ref})
			}
		case "aget-object":
			if _, err = use(1, ref, zero); err == nil {
				if _, err = use(2, integer, zero); err == nil {
					err = def(0, slot{kind: ref})
				}
			}
		case "aput-object":
			if _, err = use(0, ref, zero); err == nil {
				if _, err = use(1, ref, zero); err == nil {
					_, err = use(2, integer, zero)
				}
			}
		case "int-to-long":
			var r int
			if _, err = use(1, integer, zero); err == nil {
				if r, err = reg(0); err == nil {
					if r+1 >= total {
						return fail(i, "int-to-long needs a register pair")
					}
					s.regs[r], s.regs[r+1] = slot{kind: wideLo}, slot{kind: wideHi}
				}
			}
		case "if-eqz", "if-nez":
			if _, err = use(0, integer, zero, ref); err == nil {
				var t int
				if t, err = target(i, in.Target); err == nil {
					err = visit(t, s)
				}
			}
		case "if-ne":
			var a, b int
			if a, err = use(0, integer, zero, ref); err == nil {
				if b, err = use(1, integer, zero, ref); err == nil {
					ka, kb := s.regs[a].kind, s.regs[b].kind
					if (ka == ref) != (kb == ref) && ka != zero && kb != zero {
						return fail(i, "if-ne compares a reference with an int")
					}
					var t int
					if t, err = target(i, in.Target); err == nil {
						err = visit(t, s)
					}
				}
			}
		case "if-lt", "if-ge":
			if _, err = use(0, integer, zero); err == nil {
				if _, err = use(1, integer, zero); err == nil {
					var t int
					if t, err = target(i, in.Target); err == nil {
						err = visit(t, s)
					}
				}
			}
		case "goto":
			var t int
			if t, err = target(i, in.Target); err == nil {
				err = visit(t, s)
			}
			next = false
		case "iget", "sget", "iput", "sput":
			if in.Field == nil {
				return fail(i, "%s without a field", in.Op)
			}
			n := 0
			if in.Op == "iget" || in.Op == "iput" {
				if _, err = use(1, ref, zero); err != nil {
					break
				}
			}
			if in.Op == "iget" || in.Op == "sget" {
				err = def(n, valueOf(in.Field.Type))
			} else {
				_, err = use(n, scalarKind(in.Field.Type)...)
			}
		case "invoke-virtual", "invoke-super", "invoke-direct", "invoke-static", "invoke-interface":
			result, err = invoke(i, in, &s, index, total, fail)
		default:
			return fail(i, "opcode %q is not supported by the verifier", in.Op)
		}
		if err != nil {
			return err
		}
		s.result = result
		if next {
			if err := visit(i+1, s); err != nil {
				return err
			}
		}
	}
	return nil
}

type handler struct{ start, end, at int }

// invoke checks the receiver and arguments of one call and returns the kind of
// its result for a following move-result.
func invoke(i int, in model.Instruction, s *state, index func(model.Register) (int, bool), total int, fail func(int, string, ...any) error) (byte, error) {
	if in.Method == nil {
		return 0, fail(i, "%s without a method", in.Op)
	}
	m := *in.Method
	proto, err := model.ParsePrototype(m.Descriptor)
	if err != nil {
		return 0, fail(i, "%v", err)
	}
	var want []byte // expected kind per register: 'R' receiver, 'L', 'I', 'J' low, 'j' high
	if in.Op != "invoke-static" {
		want = append(want, 'R')
	}
	for _, p := range proto.Parameters {
		switch p[0] {
		case 'J', 'D':
			want = append(want, 'J', 'j')
		case 'L', '[':
			want = append(want, 'L')
		default:
			want = append(want, 'I')
		}
	}
	if len(in.Regs) != len(want) {
		return 0, fail(i, "%s %s passes %d registers, the descriptor needs %d", in.Op, model.MethodKey(m), len(in.Regs), len(want))
	}
	regs := make([]int, len(in.Regs))
	for n, r := range in.Regs {
		x, ok := index(r)
		if !ok {
			return 0, fail(i, "%s uses register %d outside the method's %d registers", in.Op, r, total)
		}
		regs[n] = x
	}
	for n, w := range want {
		got := s.regs[regs[n]]
		bad := false
		switch w {
		case 'R':
			if m.Name == "<init>" {
				bad = got.kind != uninit || in.Op != "invoke-direct"
			} else {
				bad = got.kind != ref && got.kind != zero
			}
		case 'L':
			bad = got.kind != ref && got.kind != zero
		case 'I':
			bad = got.kind != integer && got.kind != zero
		case 'J':
			bad = got.kind != wideLo || n+1 >= len(regs) || regs[n+1] != regs[n]+1
		case 'j':
			bad = got.kind != wideHi
		}
		if bad {
			return 0, fail(i, "%s %s: argument %d (register %d) holds %s", in.Op, model.MethodKey(m), n, in.Regs[n], got.kind)
		}
	}
	if m.Name == "<init>" {
		site := s.regs[regs[0]].site
		for r := range s.regs {
			if s.regs[r].kind == uninit && s.regs[r].site == site {
				s.regs[r] = slot{kind: ref}
			}
		}
	}
	switch proto.Return[0] {
	case 'V':
		return 'V', nil
	case 'L', '[':
		return 'L', nil
	case 'J', 'D':
		return 'J', nil
	}
	return 'I', nil
}
