// Package assembler resolves symbolic registers/references into Dalvik code units.
// Source: https://source.android.com/docs/core/runtime/dalvik-bytecode
package assembler

import (
	"fmt"
	"koru.dev/koru/internal/dex/model"
	"math"
	"strings"
)

type Resolver interface {
	String(string) (uint32, error)
	Type(string) (uint16, error)
	Method(model.MethodRef) (uint16, error)
	Field(model.FieldRef) (uint16, error)
}
type Result struct {
	Registers, Incoming, Outgoing uint16
	Units                         []uint16
	Tries                         []ResolvedTry
}
type ResolvedTry struct {
	Start    uint32
	Count    uint16
	Handler  uint32
	Type     uint16
	CatchAll bool
}
type pending struct {
	pc          int
	target      string
	conditional bool
}

func Assemble(c model.Code, descriptor string, access uint32, r Resolver) (Result, error) {
	var out Result
	p, e := model.ParsePrototype(descriptor)
	if e != nil {
		return out, e
	}
	incoming := p.Words()
	if access&model.Static == 0 {
		incoming++
	}
	total := uint32(c.Locals) + incoming
	if total > math.MaxUint16 {
		return out, fmt.Errorf("register count exceeds 65535")
	}
	out.Registers = uint16(total)
	out.Incoming = uint16(incoming)
	labels := map[string]int{}
	var jumps []pending
	reg := func(v model.Register) (uint16, error) {
		if v >= 0 {
			if uint32(v) >= uint32(c.Locals) {
				return 0, fmt.Errorf("local register %d out of bounds", v)
			}
			return uint16(v), nil
		}
		index := -int64(v) - 1
		if index >= int64(incoming) {
			return 0, fmt.Errorf("parameter word %d out of bounds", index)
		}
		return c.Locals + uint16(index), nil
	}
	for pos, i := range c.Instructions {
		fail := func(s string) (Result, error) { return Result{}, fmt.Errorf("instruction %d (%s): %s", pos, i.Op, s) }
		if i.Op == "label" {
			if i.Label == "" {
				return fail("empty label")
			}
			if _, ok := labels[i.Label]; ok {
				return fail("duplicate label")
			}
			labels[i.Label] = len(out.Units)
			continue
		}
		rr := make([]uint16, len(i.Regs))
		for j, v := range i.Regs {
			rr[j], e = reg(v)
			if e != nil {
				return fail(e.Error())
			}
		}
		need := func(n int, limit uint16) bool {
			if len(rr) != n {
				return false
			}
			for _, v := range rr {
				if v > limit {
					return false
				}
			}
			return true
		}
		emit := func(v ...uint16) { out.Units = append(out.Units, v...) }
		switch i.Op {
		case "nop", "return-void":
			if !need(0, 0) {
				return fail("unexpected registers")
			}
			if i.Op == "nop" {
				emit(0)
			} else {
				if p.Return != "V" {
					return fail("return-void in nonvoid method")
				}
				emit(0xe)
			}
		case "move", "move-object":
			if !need(2, 65535) {
				return fail("requires two registers")
			}
			op := uint16(1)
			if i.Op == "move-object" {
				op = 7
			}
			if rr[0] < 16 && rr[1] < 16 {
				emit(op | rr[0]<<8 | rr[1]<<12)
			} else if rr[0] < 256 {
				emit(op+1|rr[0]<<8, rr[1])
			} else {
				emit(op+2, rr[0], rr[1])
			}
		case "move-exception", "move-result", "move-result-object", "return", "return-object":
			if !need(1, 255) {
				return fail("requires one 8-bit register")
			}
			ops := map[string]uint16{"move-exception": 0xd, "move-result": 0xa, "move-result-object": 0xc, "return": 0xf, "return-object": 0x11}
			if i.Op == "return" && (p.Return == "V" || p.Return == "J" || p.Return == "D" || strings.ContainsAny(p.Return[:1], "L[")) {
				return fail("wrong scalar return")
			}
			if i.Op == "return-object" && !strings.ContainsAny(p.Return[:1], "L[") {
				return fail("wrong object return")
			}
			emit(ops[i.Op] | rr[0]<<8)
		case "const", "const/4", "const/16":
			if !need(1, 255) {
				return fail("requires one 8-bit register")
			}
			if i.Op == "const/4" {
				if rr[0] > 15 || i.Literal < -8 || i.Literal > 7 {
					return fail("const/4 overflow")
				}
				emit(0x12 | rr[0]<<8 | (uint16(i.Literal)&15)<<12)
			} else if i.Op == "const/16" {
				if i.Literal < math.MinInt16 || i.Literal > math.MaxInt16 {
					return fail("const/16 overflow")
				}
				emit(0x13|rr[0]<<8, uint16(i.Literal))
			} else {
				emit(0x14|rr[0]<<8, uint16(i.Literal), uint16(uint32(i.Literal)>>16))
			}
		case "const-string":
			if !need(1, 255) {
				return fail("requires one 8-bit register")
			}
			idx, err := r.String(i.String)
			if err != nil {
				return fail(err.Error())
			}
			if idx <= 65535 {
				emit(0x1a|rr[0]<<8, uint16(idx))
			} else {
				emit(0x1b|rr[0]<<8, uint16(idx), uint16(idx>>16))
			}
		case "const-class", "new-instance", "check-cast":
			if !need(1, 255) {
				return fail("requires one 8-bit register")
			}
			if err := model.ValidateType(i.Type, true); err != nil {
				return fail(err.Error())
			}
			if i.Op == "new-instance" && !model.IsClass(i.Type) {
				return fail("new-instance requires class")
			}
			if i.Op == "check-cast" && !strings.ContainsAny(i.Type[:1], "L[") {
				return fail("check-cast requires reference")
			}
			idx, err := r.Type(i.Type)
			if err != nil {
				return fail(err.Error())
			}
			ops := map[string]uint16{"const-class": 0x1c, "check-cast": 0x1f, "new-instance": 0x22}
			emit(ops[i.Op]|rr[0]<<8, idx)
		case "invoke-interface", "invoke-direct", "invoke-virtual", "invoke-static", "invoke-super":
			if i.Method == nil {
				return fail("missing method reference")
			}
			if err := model.ValidateMethodRef(*i.Method); err != nil {
				return fail(err.Error())
			}
			proto, _ := model.ParsePrototype(i.Method.Descriptor)
			words := proto.Words()
			static := i.Op == "invoke-static"
			if !static {
				words++
			}
			if words != uint32(len(rr)) || words > 255 {
				return fail("argument word count mismatch or overflow")
			}
			arg := 0
			if !static {
				arg = 1
			}
			for _, s := range proto.Parameters {
				if s == "J" || s == "D" {
					if rr[arg] == 65535 || rr[arg+1] != rr[arg]+1 {
						return fail("wide parameter requires consecutive pair")
					}
					arg++
				}
				arg++
			}
			idx, err := r.Method(*i.Method)
			if err != nil {
				return fail(err.Error())
			}
			op := map[string]uint16{"invoke-virtual": 0x6e, "invoke-super": 0x6f, "invoke-direct": 0x70, "invoke-static": 0x71, "invoke-interface": 0x72}[i.Op]
			small := len(rr) <= 5
			for _, v := range rr {
				small = small && v < 16
			}
			if small {
				var a [5]uint16
				copy(a[:], rr)
				emit(op|uint16(len(rr))<<12|a[4]<<8, idx, a[0]|a[1]<<4|a[2]<<8|a[3]<<12)
			} else {
				for j := 1; j < len(rr); j++ {
					if rr[j] != rr[0]+uint16(j) {
						return fail("invoke/range requires consecutive registers")
					}
				}
				emit(op+6|uint16(len(rr))<<8, idx, rr[0])
			}
			if uint16(words) > out.Outgoing {
				out.Outgoing = uint16(words)
			}
		case "iget", "iput", "sget", "sput":
			if i.Field == nil {
				return fail("missing field reference")
			}
			if err := model.ValidateFieldRef(*i.Field); err != nil {
				return fail(err.Error())
			}
			if i.Field.Type == "J" || i.Field.Type == "D" {
				return fail("wide fields not supported yet")
			}
			instance := i.Op == "iget" || i.Op == "iput"
			if instance && !need(2, 15) || !instance && !need(1, 255) {
				return fail("field register format overflow")
			}
			idx, err := r.Field(*i.Field)
			if err != nil {
				return fail(err.Error())
			}
			op := map[string]uint16{"iget": 0x52, "iput": 0x59, "sget": 0x60, "sput": 0x67}[i.Op]
			switch i.Field.Type[0] {
			case 'L', '[':
				op += 2
			case 'Z':
				op += 3
			case 'B':
				op += 4
			case 'C':
				op += 5
			case 'S':
				op += 6
			}
			if instance {
				emit(op|rr[0]<<8|rr[1]<<12, idx)
			} else {
				emit(op|rr[0]<<8, idx)
			}
		case "instance-of":
			if !need(2, 15) || len(i.Type) == 0 || !strings.ContainsAny(i.Type[:1], "L[") {
				return fail("instance-of requires two 4-bit registers and reference type")
			}
			idx, err := r.Type(i.Type)
			if err != nil {
				return fail(err.Error())
			}
			emit(0x20|rr[0]<<8|rr[1]<<12, idx)
		case "new-array":
			if !need(2, 15) {
				return fail("new-array requires two 4-bit registers")
			}
			if !strings.HasPrefix(i.Type, "[") || model.ValidateType(i.Type, false) != nil {
				return fail("new-array requires array type")
			}
			idx, err := r.Type(i.Type)
			if err != nil {
				return fail(err.Error())
			}
			emit(0x23|rr[0]<<8|rr[1]<<12, idx)
		case "aget-object", "aput-object":
			if !need(3, 255) {
				return fail("array access requires three 8-bit registers")
			}
			op := uint16(0x46)
			if i.Op == "aput-object" {
				op = 0x4d
			}
			emit(op|rr[0]<<8, rr[1]|rr[2]<<8)
		case "int-to-long":
			if !need(2, 15) || rr[0]+1 >= out.Registers {
				return fail("invalid int-to-long registers")
			}
			emit(0x81 | rr[0]<<8 | rr[1]<<12)
		case "if-ne", "if-lt", "if-ge":
			if !need(2, 15) {
				return fail("if requires two 4-bit registers")
			}
			op := uint16(0x34)
			if i.Op == "if-ne" {
				op = 0x33
			}
			if i.Op == "if-ge" {
				op = 0x35
			}
			jumps = append(jumps, pending{len(out.Units), i.Target, true})
			emit(op|rr[0]<<8|rr[1]<<12, 0)
		case "if-eqz", "if-nez":
			if !need(1, 255) {
				return fail("requires one 8-bit register")
			}
			op := uint16(0x38)
			if i.Op == "if-nez" {
				op = 0x39
			}
			jumps = append(jumps, pending{len(out.Units), i.Target, true})
			emit(op|rr[0]<<8, 0)
		case "goto":
			if !need(0, 0) {
				return fail("unexpected registers")
			}
			jumps = append(jumps, pending{len(out.Units), i.Target, false})
			emit(0x2a, 0, 0) // Fixed goto/32 avoids iterative layout changes.
		default:
			return fail("unsupported opcode")
		}
	}
	if len(out.Units) == 0 {
		return Result{}, fmt.Errorf("empty concrete method")
	}
	for _, j := range jumps {
		target, ok := labels[j.target]
		if !ok || target == len(out.Units) {
			return Result{}, fmt.Errorf("undefined or terminal label %q", j.target)
		}
		d := int64(target) - int64(j.pc)
		if d == 0 {
			return Result{}, fmt.Errorf("zero branch offset")
		}
		if j.conditional {
			if d < math.MinInt16 || d > math.MaxInt16 {
				return Result{}, fmt.Errorf("conditional branch overflow")
			}
			out.Units[j.pc+1] = uint16(int16(d))
		} else {
			if d < math.MinInt32 || d > math.MaxInt32 {
				return Result{}, fmt.Errorf("branch overflow")
			}
			out.Units[j.pc+1] = uint16(d)
			out.Units[j.pc+2] = uint16(uint64(d) >> 16)
		}
	}
	if len(c.Tries) > 65535 {
		return Result{}, fmt.Errorf("too many try regions")
	}
	for _, tr := range c.Tries {
		start, ok := labels[tr.Start]
		end, ok2 := labels[tr.End]
		handler, ok3 := labels[tr.Handler]
		if !ok || !ok2 || !ok3 || start >= end || end-start > 65535 || handler >= len(out.Units) {
			return Result{}, fmt.Errorf("invalid try labels or size")
		}
		t := ResolvedTry{Start: uint32(start), Count: uint16(end - start), Handler: uint32(handler), CatchAll: tr.Type == ""}
		if !t.CatchAll {
			if !model.IsClass(tr.Type) {
				return Result{}, fmt.Errorf("invalid exception type")
			}
			t.Type, e = r.Type(tr.Type)
			if e != nil {
				return Result{}, e
			}
		}
		if len(out.Tries) > 0 {
			prev := out.Tries[len(out.Tries)-1]
			if t.Start < prev.Start+uint32(prev.Count) {
				return Result{}, fmt.Errorf("overlapping or unordered try regions")
			}
		}
		out.Tries = append(out.Tries, t)
	}
	return out, nil
}
