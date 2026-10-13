package model

import "fmt"

func visibility(a uint32) error {
	v := a & 7
	if v != 0 && v != Public && v != Private && v != Protected {
		return fmt.Errorf("conflicting visibility flags 0x%x", a)
	}
	return nil
}
func ValidateClassAccess(a uint32) error {
	if a & ^uint32(Public|Final|Interface|Abstract|Synthetic|0x2000|0x4000) != 0 {
		return fmt.Errorf("unsupported class flags 0x%x", a)
	}
	if a&Final != 0 && a&Abstract != 0 {
		return fmt.Errorf("final abstract class")
	}
	if a&Interface != 0 && a&Abstract == 0 {
		return fmt.Errorf("interface must be abstract")
	}
	return nil
}
func ValidateMethodAccess(name, desc string, a uint32, hasCode bool) error {
	if e := visibility(a); e != nil {
		return e
	}
	if a & ^uint32(7|Static|Final|0x20|0x40|0x80|Native|Abstract|0x800|Synthetic|Constructor|0x20000) != 0 {
		return fmt.Errorf("unsupported method flags")
	}
	special := name == "<init>" || name == "<clinit>"
	if special != (a&Constructor != 0) {
		return fmt.Errorf("initializer requires constructor flag; ordinary method forbids it")
	}
	if name == "<init>" && (a&(Static|Native|Abstract) != 0) {
		return fmt.Errorf("invalid instance constructor flags")
	}
	if name == "<clinit>" && (a&Static == 0 || desc != "()V") {
		return fmt.Errorf("invalid class initializer")
	}
	if a&Abstract != 0 && a&(Private|Static|Final|Native|Constructor|0x20|0x800) != 0 {
		return fmt.Errorf("conflicting abstract method flags")
	}
	if a&0x20 != 0 && a&Native == 0 {
		return fmt.Errorf("ACC_SYNCHRONIZED requires native")
	}
	if hasCode == (a&(Native|Abstract) != 0) {
		return fmt.Errorf("code must be present exactly for concrete non-native methods")
	}
	return nil
}
func ValidateFieldAccess(a uint32) error {
	if e := visibility(a); e != nil {
		return e
	}
	if a & ^uint32(7|Static|Final|0x40|0x80|Synthetic|0x4000) != 0 {
		return fmt.Errorf("unsupported field flags")
	}
	if a&Final != 0 && a&0x40 != 0 {
		return fmt.Errorf("final volatile field")
	}
	return nil
}
func IsDirect(a uint32) bool { return a&(Static|Private|Constructor) != 0 }
