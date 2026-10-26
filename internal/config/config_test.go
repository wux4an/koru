package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	c := Default()
	if e := c.Validate(); e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(*Config){func(c *Config) { c.AppID = "../x" }, func(c *Config) { c.MinSDK = 23 }, func(c *Config) { c.TargetSDK = 99 }, func(c *Config) { c.Plugins = []string{"unknown"} }, func(c *Config) { c.Permissions = []string{"BAD"} }} {
		c := Default()
		edit(&c)
		if c.Validate() == nil {
			t.Fatal(c)
		}
	}
}

func TestSchemaMatchesConfigFields(t *testing.T) {
	var schema struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if e := json.Unmarshal(JSONSchema, &schema); e != nil {
		t.Fatal(e)
	}
	typ := reflect.TypeOf(Config{})
	fields := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")
		fields[tag[0]] = true
		if _, ok := schema.Properties[tag[0]]; !ok {
			t.Errorf("schema lacks %s", tag[0])
		}
		if len(tag) == 1 && !slices.Contains(schema.Required, tag[0]) {
			t.Errorf("schema should require %s", tag[0])
		}
	}
	for name := range schema.Properties {
		if !fields[name] {
			t.Errorf("schema property %s is not a Config field", name)
		}
	}
}

func TestLoadAcceptsSchemaAndExplainsErrors(t *testing.T) {
	dir := t.TempDir()
	if _, e := Load(dir); e == nil || !strings.Contains(e.Error(), "koru create") {
		t.Fatalf("missing config error = %v", e)
	}
	c := Default()
	c.Schema = SchemaURL
	b, _ := json.Marshal(c)
	os.WriteFile(filepath.Join(dir, "koru.json"), b, 0644)
	if _, e := Load(dir); e != nil {
		t.Fatal(e)
	}
	c.AppID = "myapp"
	b, _ = json.Marshal(c)
	os.WriteFile(filepath.Join(dir, "koru.json"), b, 0644)
	if _, e := Load(dir); e == nil || !strings.Contains(e.Error(), "com.example.myapp") {
		t.Fatalf("appId error = %v", e)
	}
	os.WriteFile(filepath.Join(dir, "koru.json"), []byte(`{"appid":"x"}`), 0644)
	if _, e := Load(dir); e == nil || !strings.HasPrefix(e.Error(), "koru.json: ") {
		t.Fatalf("unknown field error = %v", e)
	}
}
