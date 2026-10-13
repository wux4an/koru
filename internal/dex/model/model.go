// Package model contains the symbolic DEX input representation, independent of offsets.
package model

const (
	Public      uint32 = 1
	Private     uint32 = 2
	Protected   uint32 = 4
	Static      uint32 = 8
	Final       uint32 = 0x10
	Native      uint32 = 0x100
	Interface   uint32 = 0x200
	Abstract    uint32 = 0x400
	Synthetic   uint32 = 0x1000
	Constructor uint32 = 0x10000
)

// Register is a local slot (>=0) or a parameter word (<0). Parameters always
// occupy the last registers of a code_item; P(0) is this for instance methods.
type Register int32

func V(n uint16) Register { return Register(n) }
func P(n uint16) Register { return -Register(n) - 1 }

type Program struct {
	Classes []Class `json:"classes"`
}
type Class struct {
	Descriptor string   `json:"descriptor"`
	Super      string   `json:"super"`
	Access     uint32   `json:"access"`
	Interfaces []string `json:"interfaces,omitempty"`
	Source     string   `json:"source,omitempty"`
	Fields     []Field  `json:"fields,omitempty"`
	Methods    []Method `json:"methods"`
}
type FieldRef struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}
type MethodRef struct {
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	Descriptor string `json:"descriptor"`
}
type Field struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Access uint32 `json:"access"`
}
type Method struct {
	Name       string `json:"name"`
	Descriptor string `json:"descriptor"`
	Access     uint32 `json:"access"`
	Code       *Code  `json:"code,omitempty"`
}
type Code struct {
	Locals       uint16        `json:"locals"`
	Tries        []Try         `json:"tries,omitempty"`
	Instructions []Instruction `json:"instructions"`
}

// Op accepts the documented Dalvik subset; label is a pseudo instruction.
// Method/Field references are resolved after identifier sorting.
type Instruction struct {
	Op      string     `json:"op"`
	Regs    []Register `json:"regs,omitempty"`
	Literal int32      `json:"literal,omitempty"`
	String  string     `json:"string,omitempty"`
	Type    string     `json:"type,omitempty"`
	Method  *MethodRef `json:"method,omitempty"`
	Field   *FieldRef  `json:"field,omitempty"`
	Label   string     `json:"label,omitempty"`
	Target  string     `json:"target,omitempty"`
}

// Try specifies a half-open protected region and one typed or catch-all handler.
// An empty Type catches all throwables. Handler starts with move-exception.
type Try struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Handler string `json:"handler"`
	Type    string `json:"type,omitempty"`
}
