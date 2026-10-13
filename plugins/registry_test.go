package plugins

import (
	"koru.dev/koru/pkg/dex"
	"testing"
)

func TestPluginABI(t *testing.T) {
	owner := "Lexample/Echo;"
	b := dex.NewCode(1).Emit(dex.Instruction{Op: "new-instance", Regs: []dex.Register{0}, Type: "Lorg/json/JSONObject;"}).Invoke("invoke-direct", dex.MethodRef{Owner: "Lorg/json/JSONObject;", Name: "<init>", Descriptor: "()V"}, 0).Emit(dex.Instruction{Op: "return-object", Regs: []dex.Register{0}})
	d := Descriptor{ID: "echo", ABI: ABI, Handlers: map[string]dex.MethodRef{"echo.ping": {Owner: owner, Name: "ping", Descriptor: HandlerDescriptor}}, Classes: []dex.Class{{Descriptor: owner, Super: "Ljava/lang/Object;", Access: dex.Public, Methods: []dex.Method{{Name: "ping", Descriptor: HandlerDescriptor, Access: dex.Public | dex.Static, Code: b.Build()}}}}}
	if e := d.Validate(); e != nil {
		t.Fatal(e)
	}
	d.ABI = 2
	if d.Validate() == nil {
		t.Fatal("unsupported ABI")
	}
	d.ABI = 1
	d.ID = "device"
	if d.Validate() == nil {
		t.Fatal("builtin replacement")
	}
}
