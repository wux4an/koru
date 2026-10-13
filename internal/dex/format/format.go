// Package format defines DEX 035 binary constants. Source:
// https://source.android.com/docs/core/runtime/dex-format
package format

const (
	Magic             = "dex\n035\x00"
	HeaderSize        = 0x70
	Endian            = 0x12345678
	NoIndex    uint32 = 0xffffffff
	HeaderItem uint16 = 0
	StringID   uint16 = 1
	TypeID     uint16 = 2
	ProtoID    uint16 = 3
	FieldID    uint16 = 4
	MethodID   uint16 = 5
	ClassDef   uint16 = 6
	MapList    uint16 = 0x1000
	TypeList   uint16 = 0x1001
	ClassData  uint16 = 0x2000
	CodeItem   uint16 = 0x2001
	StringData uint16 = 0x2002
)

type MapItem struct {
	Type         uint16
	Size, Offset uint32
}
