package config

import "testing"

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
