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

func TestManifestFieldsValidation(t *testing.T) {
	c := Default()
	c.Orientation, c.Keyboard, c.AllowBackup = "sensor", "pan", true
	c.DeepLinks = []DeepLink{{Scheme: "https", Host: "example.com", PathPrefix: "/app"}, {Scheme: "https", Host: "*.example.org"}, {Scheme: "myapp"}, {Scheme: "myapp", Host: "open", PathPrefix: "/x"}}
	if e := c.Validate(); e != nil {
		t.Fatal(e)
	}
	if got := c.HTTPSHosts(); !slices.Equal(got, []string{"example.com", "example.org"}) {
		t.Fatal(got)
	}
	https := func(h, p string) DeepLink { return DeepLink{Scheme: "https", Host: h, PathPrefix: p} }
	cases := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"orientation", func(c *Config) { c.Orientation = "Portrait" }, `use "portrait"`},
		{"keyboard", func(c *Config) { c.Keyboard = "none" }, `use "resize"`},
		{"http", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "http", Host: "example.com"}} }, "plain HTTP"},
		{"javascript", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "javascript"}} }, "forbidden"},
		{"file", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "file"}} }, "forbidden"},
		{"content", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "content"}} }, "forbidden"},
		{"uppercase scheme", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "MyApp"}} }, "lowercase"},
		{"empty scheme", func(c *Config) { c.DeepLinks = []DeepLink{{}} }, "scheme"},
		{"https without host", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "https"}} }, "full domain"},
		{"https single label", func(c *Config) { c.DeepLinks = []DeepLink{https("localhost", "")} }, "full domain"},
		{"host with scheme", func(c *Config) { c.DeepLinks = []DeepLink{https("https://example.com", "")} }, "invalid deepLinks host"},
		{"host with port", func(c *Config) { c.DeepLinks = []DeepLink{https("example.com:8080", "")} }, "invalid deepLinks host"},
		{"uppercase host", func(c *Config) { c.DeepLinks = []DeepLink{https("Example.com", "")} }, "lowercase"},
		{"path without slash", func(c *Config) { c.DeepLinks = []DeepLink{https("example.com", "app")} }, `start with "/"`},
		{"path with query", func(c *Config) { c.DeepLinks = []DeepLink{https("example.com", "/a?b")} }, "remove spaces"},
		{"path with space", func(c *Config) { c.DeepLinks = []DeepLink{https("example.com", "/a b")} }, "remove spaces"},
		{"path without host", func(c *Config) { c.DeepLinks = []DeepLink{{Scheme: "myapp", PathPrefix: "/x"}} }, "needs a host"},
		{"duplicate", func(c *Config) { c.DeepLinks = []DeepLink{https("example.com", ""), https("example.com", "")} }, "duplicate"},
		{"too many", func(c *Config) {
			for i := 0; i <= MaxDeepLinks; i++ {
				c.DeepLinks = append(c.DeepLinks, DeepLink{Scheme: "myapp", Host: "h", PathPrefix: "/" + strings.Repeat("a", i+1)})
			}
		}, "too many"},
	}
	for _, tc := range cases {
		c := Default()
		tc.edit(&c)
		if e := c.Validate(); e == nil || !strings.Contains(e.Error(), tc.want) {
			t.Errorf("%s: error = %v, want %q", tc.name, e, tc.want)
		}
	}
}

func TestLoadRejectsUnknownDeepLinkField(t *testing.T) {
	dir := t.TempDir()
	b, _ := json.Marshal(Default())
	b = append(b[:len(b)-1], []byte(`,"deepLinks":[{"scheme":"myapp","path":"/x"}]}`)...)
	os.WriteFile(filepath.Join(dir, "koru.json"), b, 0644)
	if _, e := Load(dir); e == nil || !strings.Contains(e.Error(), "unknown field") {
		t.Fatal(e)
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
