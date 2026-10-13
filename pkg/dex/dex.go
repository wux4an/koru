// Package dex exposes koru's experimental programmatic DEX 035 API.
// API stability is not promised until v1; format and supported subset are explicit.
package dex

import (
	"fmt"
	"koru.dev/koru/internal/dex/builder"
	"koru.dev/koru/internal/dex/model"
	"koru.dev/koru/internal/dex/parser"
)

type Program = model.Program
type Class = model.Class
type Field = model.Field
type Method = model.Method
type MethodRef = model.MethodRef
type FieldRef = model.FieldRef
type Code = model.Code
type Try = model.Try
type Instruction = model.Instruction
type Register = model.Register

// Inspection is the validated, decoded representation of a supported DEX file.
type Inspection = parser.File

const (
	Public      = model.Public
	Private     = model.Private
	Protected   = model.Protected
	Static      = model.Static
	Final       = model.Final
	Native      = model.Native
	Interface   = model.Interface
	Abstract    = model.Abstract
	Constructor = model.Constructor
)

func V(n uint16) Register { return model.V(n) }
func P(n uint16) Register { return model.P(n) }

// Build checks both the symbolic input and the emitted structure before returning.
func Build(p Program) ([]byte, error) {
	b, e := builder.Build(p)
	if e != nil {
		return nil, e
	}
	if _, e = parser.Parse(b); e != nil {
		return nil, e
	}
	return b, nil
}

// Inspect validates checksums, sections, identifiers, references and instructions.
// It rejects sections and opcodes outside the currently supported subset.
func Inspect(b []byte) (*Inspection, error) { return parser.Parse(b) }

// CodeBuilder is an optional convenience over the serializable Code IR.
type CodeBuilder struct {
	code      Code
	nextLabel uint32
}

func NewCode(locals uint16) *CodeBuilder { return &CodeBuilder{code: Code{Locals: locals}} }
func (b *CodeBuilder) Emit(i Instruction) *CodeBuilder {
	b.code.Instructions = append(b.code.Instructions, cloneInstruction(i))
	return b
}
func (b *CodeBuilder) Label(name string) *CodeBuilder {
	return b.Emit(Instruction{Op: "label", Label: name})
}
func (b *CodeBuilder) ReturnVoid() *CodeBuilder { return b.Emit(Instruction{Op: "return-void"}) }
func (b *CodeBuilder) Const(dst Register, n int32) *CodeBuilder {
	return b.Emit(Instruction{Op: "const", Regs: []Register{dst}, Literal: n})
}
func (b *CodeBuilder) ConstString(dst Register, s string) *CodeBuilder {
	return b.Emit(Instruction{Op: "const-string", Regs: []Register{dst}, String: s})
}
func (b *CodeBuilder) Invoke(op string, m MethodRef, args ...Register) *CodeBuilder {
	return b.Emit(Instruction{Op: op, Method: &m, Regs: append([]Register(nil), args...)})
}
func (b *CodeBuilder) Build() *Code {
	out := b.code
	out.Tries = append([]Try(nil), b.code.Tries...)
	out.Instructions = make([]Instruction, len(b.code.Instructions))
	for j, i := range b.code.Instructions {
		out.Instructions[j] = cloneInstruction(i)
	}
	return &out
}

func cloneInstruction(i Instruction) Instruction {
	i.Regs = append([]Register(nil), i.Regs...)
	if i.Method != nil {
		v := *i.Method
		i.Method = &v
	}
	if i.Field != nil {
		v := *i.Field
		i.Field = &v
	}
	return i
}

func (b *CodeBuilder) Catch(start, end, handler, typ string) *CodeBuilder {
	b.code.Tries = append(b.code.Tries, Try{Start: start, End: end, Handler: handler, Type: typ})
	return b
}

func (b *CodeBuilder) UniqueLabel(prefix string) string {
	b.nextLabel++
	return fmt.Sprintf("__koru_%s_%d", prefix, b.nextLabel)
}
