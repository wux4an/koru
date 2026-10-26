package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/plugins"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// JSONSchema is the JSON Schema for koru.json, published at SchemaURL.
//
//go:embed koru.schema.json
var JSONSchema []byte

// SchemaURL is where the published JSON Schema for koru.json lives.
const SchemaURL = "https://koru.js.org/schema/koru.schema.json"

type Config struct {
	Schema          string              `json:"$schema,omitempty"`
	AppID           string              `json:"appId"`
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	VersionCode     uint32              `json:"versionCode"`
	WebDir          string              `json:"webDir"`
	MinSDK          uint32              `json:"minSdk"`
	TargetSDK       uint32              `json:"targetSdk"`
	Permissions     []string            `json:"permissions"`
	Plugins         []string            `json:"plugins"`
	SigningKey      string              `json:"signingKey"`
	Icon            string              `json:"icon,omitempty"`
	SPA             bool                `json:"spa,omitempty"`
	PluginManifests []string            `json:"pluginManifests,omitempty"`
	NativeLibraries map[string][]string `json:"nativeLibraries,omitempty"`
}

var appPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)+$`)
var permissionPattern = regexp.MustCompile(`^android\.permission\.[A-Z][A-Z0-9_]*$`)
var pluginPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
var supported = map[string]bool{"device": true, "storage": true, "vibration": true, "app": true, "share": true, "notifications": true}

func Default() Config {
	return Config{AppID: "com.example.koru", Name: "Koru", Version: "1.0.0", VersionCode: 1, WebDir: "www", MinSDK: 26, TargetSDK: 36, Permissions: []string{}, Plugins: []string{"device", "storage", "app"}, SigningKey: ".koru/signing.pem"}
}
func (c Config) Validate() error {
	if !appPattern.MatchString(c.AppID) || len(c.AppID) > 127 {
		return fmt.Errorf("invalid appId %q: use reverse-domain form such as com.example.myapp (letters, digits and _; at least two segments)", c.AppID)
	}
	if strings.TrimSpace(c.Name) == "" || len(c.Name) > 256 {
		return errors.New("invalid name: set a display name of 1 to 256 characters")
	}
	if c.Version == "" || len(c.Version) > 128 {
		return errors.New("invalid version: set a version string such as 1.0.0")
	}
	if c.VersionCode == 0 || c.VersionCode > 2100000000 {
		return errors.New("invalid versionCode: use a whole number from 1 to 2100000000, and increase it for every update")
	}
	if c.MinSDK < 26 || c.TargetSDK < c.MinSDK || c.TargetSDK > 36 {
		return errors.New("supported SDK range: 26 <= minSdk <= targetSdk <= 36")
	}
	if c.WebDir == "" || c.SigningKey == "" {
		return errors.New("webDir and signingKey required")
	}
	seen := map[string]bool{}
	for _, p := range c.Permissions {
		if !permissionPattern.MatchString(p) || seen[p] {
			return fmt.Errorf("invalid or duplicate permission %q", p)
		}
		seen[p] = true
	}
	seen = map[string]bool{}
	for _, p := range c.Plugins {
		if (!supported[p] && len(c.PluginManifests) == 0) || seen[p] || !pluginPattern.MatchString(p) {
			return fmt.Errorf("unsupported or duplicate plugin %q (built-in plugins: app, device, notifications, share, storage, vibration)", p)
		}
		seen[p] = true
	}
	return nil
}
func Load(project string) (Config, error) {
	var c Config
	f, e := os.Open(filepath.Join(project, "koru.json"))
	if os.IsNotExist(e) {
		return c, fmt.Errorf("no koru.json in %s: create a project with `koru create` or `koru init`", project)
	}
	if e != nil {
		return c, e
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return c, fmt.Errorf("koru.json: %w", e)
	}
	var x any
	if e = d.Decode(&x); e != io.EOF {
		return c, errors.New("trailing config content")
	}
	if e = c.Validate(); e != nil {
		return c, fmt.Errorf("koru.json: %w", e)
	}
	return c, nil
}
func (c Config) HasPlugin(name string) bool {
	for _, p := range c.Plugins {
		if p == name {
			return true
		}
	}
	return false
}
func (c Config) HasPermission(name string) bool {
	for _, p := range c.Permissions {
		if p == name {
			return true
		}
	}
	return false
}
func (c Config) EffectivePermissions() []string {
	out := append([]string(nil), c.Permissions...)
	add := func(s string) {
		for _, v := range out {
			if v == s {
				return
			}
		}
		out = append(out, s)
	}
	if c.HasPlugin("vibration") {
		add("android.permission.VIBRATE")
	}
	if c.HasPlugin("notifications") {
		add("android.permission.POST_NOTIFICATIONS")
	}
	return out
}

func (c Config) LoadPlugins(project string) ([]plugins.Descriptor, error) {
	registry := map[string]plugins.Descriptor{}
	for _, path := range c.PluginManifests {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		d, e := plugins.Load(path)
		if e != nil {
			return nil, e
		}
		if _, ok := registry[d.ID]; ok {
			return nil, errors.New("duplicate plugin manifest")
		}
		registry[d.ID] = d
	}
	var out []plugins.Descriptor
	for _, id := range c.Plugins {
		if plugins.Builtin(id) {
			continue
		}
		d, ok := registry[id]
		if !ok {
			return nil, fmt.Errorf("plugin %q has no manifest", id)
		}
		out = append(out, d)
	}
	return out, nil
}
