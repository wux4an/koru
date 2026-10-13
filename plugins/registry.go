// Package plugins defines the stable v1 DEX plugin handler ABI. A plugin receives
// Activity + validated JSONObject params, and returns a JSONObject result.
package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/pkg/dex"
	"os"
	"regexp"
	"strings"
)

const ABI = 1
const HandlerDescriptor = "(Landroid/app/Activity;Lorg/json/JSONObject;)Lorg/json/JSONObject;"

type Descriptor struct {
	ID          string                   `json:"id"`
	ABI         uint32                   `json:"abi"`
	Permissions []string                 `json:"permissions"`
	Handlers    map[string]dex.MethodRef `json:"handlers"`
	Classes     []dex.Class              `json:"classes"`
}

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
var methodPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}\.[a-zA-Z][a-zA-Z0-9_]{0,63}$`)
var permissionPattern = regexp.MustCompile(`^android\.permission\.[A-Z][A-Z0-9_]*$`)

func Builtin(id string) bool {
	switch id {
	case "device", "app", "storage", "vibration", "share", "notifications":
		return true
	}
	return false
}
func Load(path string) (Descriptor, error) {
	var d Descriptor
	f, e := os.Open(path)
	if e != nil {
		return d, e
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 4<<20))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&d); e != nil {
		return d, e
	}
	var x any
	if dec.Decode(&x) != io.EOF {
		return d, errors.New("trailing plugin input")
	}
	return d, d.Validate()
}
func (d Descriptor) Validate() error {
	if !idPattern.MatchString(d.ID) || Builtin(d.ID) || d.ABI != ABI || len(d.Handlers) == 0 {
		return errors.New("invalid/custom plugin identity or ABI")
	}
	if len(d.Classes) == 0 {
		return errors.New("plugin classes required")
	}
	classes := map[string]dex.Class{}
	for _, c := range d.Classes {
		if strings.HasPrefix(c.Descriptor, "Landroid/") || strings.HasPrefix(c.Descriptor, "Ljava/") || strings.HasPrefix(c.Descriptor, "Lcom/koru/runtime/") {
			return errors.New("plugin cannot replace platform/runtime classes")
		}
		if _, ok := classes[c.Descriptor]; ok {
			return errors.New("duplicate plugin class")
		}
		classes[c.Descriptor] = c
	}
	for name, ref := range d.Handlers {
		if !methodPattern.MatchString(name) || !strings.HasPrefix(name, d.ID+".") || ref.Descriptor != HandlerDescriptor {
			return fmt.Errorf("invalid plugin handler %s", name)
		}
		c, ok := classes[ref.Owner]
		found := false
		if ok && c.Access&dex.Public != 0 {
			for _, m := range c.Methods {
				if m.Name == ref.Name && m.Descriptor == ref.Descriptor && m.Access&(dex.Public|dex.Static) == dex.Public|dex.Static {
					found = true
				}
			}
		}
		if !found {
			return errors.New("handler must resolve to a public static plugin method")
		}
	}
	seen := map[string]bool{}
	for _, p := range d.Permissions {
		if !permissionPattern.MatchString(p) || seen[p] {
			return errors.New("invalid plugin permissions")
		}
		seen[p] = true
	}
	if _, e := dex.Build(dex.Program{Classes: d.Classes}); e != nil {
		return fmt.Errorf("plugin DEX: %w", e)
	}
	return nil
}
