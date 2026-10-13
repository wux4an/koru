package parser

import (
	"fmt"
	"koru.dev/koru/internal/dex/model"
	"strings"
)

type decoded struct {
	pc, width                     int
	op                            byte
	target                        int
	branch, conditional, terminal bool
	result                        string
}

func (f *File) verifyCode(method Method) error {
	c := method.Code
	u := c.Units
	self := f.Methods[method.Index]
	returnType := f.Types[f.Prototypes[self.Prototype].Return]
	var instructions []decoded
	boundaries := map[int]int{}
	reg := func(v uint16) error {
		if v >= c.Registers {
			return fmt.Errorf("register v%d exceeds %d", v, c.Registers)
		}
		return nil
	}
	for pc := 0; pc < len(u); {
		word := u[pc]
		op := byte(word)
		d := decoded{pc: pc, width: 1, op: op}
		a := word >> 8
		switch {
		case op == 0:
			if word != 0 {
				return fmt.Errorf("payloads unsupported")
			}
		case op == 1 || op == 7:
			d.width = 1
		case op == 2 || op == 8:
			d.width = 2
		case op == 3 || op == 9:
			d.width = 3
			if a != 0 {
				return fmt.Errorf("reserved move/16 bits")
			}
		case op == 0xa || op == 0xc || op == 0xd || op == 0xf || op == 0x11 || op == 0xe || op == 0x12 || op == 0x81:
			d.width = 1
		case op == 0x13 || op == 0x1a || op == 0x1c || op == 0x1f || op == 0x22 || op == 0x38 || op == 0x39 || op == 0x33 || op == 0x34 || op == 0x35 || op == 0x20 || op == 0x23 || op == 0x46 || op == 0x4d:
			d.width = 2
		case op == 0x14 || op == 0x1b || op == 0x2a:
			d.width = 3
		case op == 0x28:
			d.width = 1
		case op == 0x29:
			d.width = 2
		case op >= 0x52 && op <= 0x6d:
			d.width = 2
		case op >= 0x6e && op <= 0x72 || op >= 0x74 && op <= 0x78:
			d.width = 3
		default:
			return fmt.Errorf("unsupported opcode 0x%x at %d", op, pc)
		}
		if d.width > len(u)-pc {
			return fmt.Errorf("truncated instruction at %d", pc)
		}
		var regs []uint16
		switch {
		case op == 1 || op == 7:
			regs = []uint16{a & 15, a >> 4}
		case op == 2 || op == 8:
			regs = []uint16{a, u[pc+1]}
		case op == 3 || op == 9:
			regs = []uint16{u[pc+1], u[pc+2]}
		case op == 0xa || op == 0xc || op == 0xd || op == 0xf || op == 0x11:
			regs = []uint16{a}
			if op == 0xf || op == 0x11 {
				d.terminal = true
				if op == 0xf && (returnType == "V" || returnType == "J" || returnType == "D" || strings.ContainsAny(returnType[:1], "L[")) {
					return fmt.Errorf("wrong scalar return opcode")
				}
				if op == 0x11 && !strings.ContainsAny(returnType[:1], "L[") {
					return fmt.Errorf("wrong object return opcode")
				}
			}
		case op == 0xe:
			if a != 0 || returnType != "V" {
				return fmt.Errorf("invalid return-void")
			}
			d.terminal = true
		case op == 0x12:
			regs = []uint16{a & 15}
		case op == 0x13 || op == 0x14:
			regs = []uint16{a}
		case op == 0x1a || op == 0x1b:
			regs = []uint16{a}
			idx := uint32(u[pc+1])
			if op == 0x1b {
				idx |= uint32(u[pc+2]) << 16
			}
			if idx >= uint32(len(f.Strings)) {
				return fmt.Errorf("const-string reference out of bounds")
			}
		case op == 0x1c || op == 0x1f || op == 0x22:
			regs = []uint16{a}
			idx := u[pc+1]
			if int(idx) >= len(f.Types) {
				return fmt.Errorf("type reference out of bounds")
			}
			s := f.Types[idx]
			if op == 0x22 && !model.IsClass(s) || op == 0x1f && !strings.ContainsAny(s[:1], "L[") {
				return fmt.Errorf("wrong type for opcode")
			}
		case op == 0x28 || op == 0x29 || op == 0x2a || op == 0x38 || op == 0x39 || op == 0x33 || op == 0x34 || op == 0x35:
			d.branch = true
			var delta int64
			switch op {
			case 0x28:
				delta = int64(int8(a))
			case 0x29:
				if a != 0 {
					return fmt.Errorf("reserved goto/16 bits")
				}
				delta = int64(int16(u[pc+1]))
			case 0x2a:
				if a != 0 {
					return fmt.Errorf("reserved goto/32 bits")
				}
				delta = int64(int32(uint32(u[pc+1]) | uint32(u[pc+2])<<16))
			default:
				regs = []uint16{a}
				if op == 0x33 || op == 0x34 || op == 0x35 {
					regs = []uint16{a & 15, a >> 4}
				}
				d.conditional = true
				delta = int64(int16(u[pc+1]))
			}
			target := int64(pc) + delta
			if delta == 0 || target < 0 || target >= int64(len(u)) {
				return fmt.Errorf("invalid branch target")
			}
			d.target = int(target)

		case op == 0x81:
			regs = []uint16{a & 15, (a & 15) + 1, a >> 4}
		case op == 0x20:
			regs = []uint16{a & 15, a >> 4}
			idx := u[pc+1]
			if int(idx) >= len(f.Types) || !strings.ContainsAny(f.Types[idx][:1], "L[") {
				return fmt.Errorf("instance-of type mismatch")
			}
		case op == 0x23:
			regs = []uint16{a & 15, a >> 4}
			idx := u[pc+1]
			if int(idx) >= len(f.Types) || !strings.HasPrefix(f.Types[idx], "[") {
				return fmt.Errorf("new-array type mismatch")
			}
		case op == 0x46 || op == 0x4d:
			regs = []uint16{a, u[pc+1] & 255, u[pc+1] >> 8}
		case op >= 0x52 && op <= 0x6d:
			idx := u[pc+1]
			if int(idx) >= len(f.Fields) {
				return fmt.Errorf("field reference out of bounds")
			}
			field := f.Fields[idx]
			instance := op <= 0x5f
			if instance {
				regs = []uint16{a & 15, a >> 4}
			} else {
				regs = []uint16{a}
			}
			base := byte(0x52)
			switch {
			case op >= 0x67:
				base = 0x67
			case op >= 0x60:
				base = 0x60
			case op >= 0x59:
				base = 0x59
			}
			variant := op - base
			s := f.Types[field.Type]
			want := byte(0)
			switch s[0] {
			case 'J', 'D':
				return fmt.Errorf("wide field instructions unsupported")
			case 'L', '[':
				want = 2
			case 'Z':
				want = 3
			case 'B':
				want = 4
			case 'C':
				want = 5
			case 'S':
				want = 6
			}
			if variant != want {
				return fmt.Errorf("field opcode/type mismatch")
			}
			for _, cl := range f.Classes {
				for _, v := range cl.Fields {
					if v.Index == uint32(idx) && (v.Access&model.Static == 0) != instance {
						return fmt.Errorf("field staticness mismatch")
					}
				}
			}
		case op >= 0x6e && op <= 0x72 || op >= 0x74 && op <= 0x78:
			idx := u[pc+1]
			if int(idx) >= len(f.Methods) {
				return fmt.Errorf("invoke method out of bounds")
			}
			target := f.Methods[idx]
			name := f.Strings[target.Name]
			rangeForm := op >= 0x74
			normal := op
			if rangeForm {
				normal -= 6
			}
			static := normal == 0x71
			pr := f.Prototypes[target.Prototype]
			words := 0
			if !static {
				words++
			}
			for _, v := range pr.Parameters {
				words++
				if f.Types[v] == "J" || f.Types[v] == "D" {
					words++
				}
			}
			if rangeForm {
				count := int(a)
				if int(u[pc+2])+count > int(c.Registers) {
					return fmt.Errorf("invoke/range register overflow")
				}
				for i := 0; i < count; i++ {
					regs = append(regs, u[pc+2]+uint16(i))
				}
			} else {
				count := int(word >> 12)
				if count > 5 {
					return fmt.Errorf("invoke-35c argument overflow")
				}
				packed := u[pc+2]
				all := []uint16{packed & 15, (packed >> 4) & 15, (packed >> 8) & 15, packed >> 12, a & 15}
				for _, v := range all[count:] {
					if v != 0 {
						return fmt.Errorf("nonzero unused invoke register")
					}
				}
				regs = all[:count]
			}
			if words != len(regs) || words > int(c.Outgoing) {
				return fmt.Errorf("invoke argument/outs mismatch")
			}
			arg := 0
			if !static {
				arg = 1
			}
			for _, v := range pr.Parameters {
				if f.Types[v] == "J" || f.Types[v] == "D" {
					if regs[arg] == 65535 || regs[arg+1] != regs[arg]+1 {
						return fmt.Errorf("nonconsecutive wide argument")
					}
					arg++
				}
				arg++
			}
			if name == "<clinit>" || name == "<init>" && normal != 0x70 {
				return fmt.Errorf("invalid initializer invocation")
			}
			for _, cl := range f.Classes {
				for _, m := range cl.Methods {
					if m.Index != uint32(idx) {
						continue
					}
					if (m.Access&model.Static != 0) != static {
						return fmt.Errorf("method staticness mismatch")
					}
					if normal == 0x70 && m.Access&(model.Private|model.Constructor) == 0 {
						return fmt.Errorf("invoke-direct requires private instance method or constructor")
					}
					if (normal == 0x6e || normal == 0x6f) && model.IsDirect(m.Access) {
						return fmt.Errorf("virtual/super invocation targets direct method")
					}
				}
			}
			d.result = f.Types[pr.Return]
		}
		for _, v := range regs {
			if e := reg(v); e != nil {
				return e
			}
		}
		boundaries[pc] = len(instructions)
		instructions = append(instructions, d)
		pc += d.width
	}

	handlerPCs := map[int]bool{}
	for _, tr := range c.Tries {
		_, startOK := boundaries[int(tr.Start)]
		_, endOK := boundaries[int(tr.Start+tr.Count)]
		if int(tr.Start+tr.Count) == len(u) {
			endOK = true
		}
		j, handlerOK := boundaries[int(tr.Handler.PC)]
		if !startOK || !endOK || !handlerOK || instructions[j].op != 0xd {
			return fmt.Errorf("invalid exception instruction boundaries")
		}
		if !tr.Handler.CatchAll && (tr.Handler.Type >= uint32(len(f.Types)) || !model.IsClass(f.Types[tr.Handler.Type])) {
			return fmt.Errorf("invalid catch type")
		}
		handlerPCs[int(tr.Handler.PC)] = true
	}
	for i, d := range instructions {
		if d.op == 0xd && !handlerPCs[d.pc] {
			return fmt.Errorf("move-exception outside handler")
		}
		if d.branch {
			if handlerPCs[d.target] {
				return fmt.Errorf("branch into exception handler")
			}
			j, ok := boundaries[d.target]
			if !ok {
				return fmt.Errorf("branch into instruction interior")
			}
			if instructions[j].op == 0xa || instructions[j].op == 0xc {
				return fmt.Errorf("branch into move-result")
			}
		}
		if d.op == 0xa || d.op == 0xc {
			if i == 0 || instructions[i-1].result == "" || instructions[i-1].result == "V" {
				return fmt.Errorf("move-result must follow nonvoid invoke")
			}
			result := instructions[i-1].result
			object := strings.ContainsAny(result[:1], "L[")
			if (d.op == 0xc) != object || result == "J" || result == "D" {
				return fmt.Errorf("move-result type mismatch")
			}
		}
	}
	// Check every reachable edge; a final nonterminal instruction is not executable
	// merely because all individual opcode encodings are valid.
	queue := []int{0}
	for pc := range handlerPCs {
		queue = append(queue, pc)
	}
	visited := map[int]bool{}
	for len(queue) > 0 {
		pc := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if visited[pc] {
			continue
		}
		visited[pc] = true
		d := instructions[boundaries[pc]]
		if d.branch {
			queue = append(queue, d.target)
		}
		if !d.terminal && (!d.branch || d.conditional) {
			next := d.pc + d.width
			if next >= len(u) {
				return fmt.Errorf("reachable execution falls past code end")
			}
			if handlerPCs[next] {
				return fmt.Errorf("normal flow into exception handler")
			}
			queue = append(queue, next)
		}
	}
	return nil
}
