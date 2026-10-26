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
	"slices"
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
	// Orientation locks the Activity: "portrait", "landscape" or "sensor".
	// Empty leaves Android's default and emits no attribute.
	Orientation string `json:"orientation,omitempty"`
	// AllowBackup opts the app into Android auto backup (default false).
	AllowBackup bool `json:"allowBackup,omitempty"`
	// Keyboard is the soft keyboard policy: "resize" (default) or "pan".
	Keyboard string `json:"keyboard,omitempty"`
	// DeepLinks make links open the app. The link opens the app on its
	// initial page: the URL is not delivered to JavaScript yet.
	DeepLinks []DeepLink `json:"deepLinks,omitempty"`
}

// DeepLink is one URL pattern the app handles: an https link such as
// {"scheme":"https","host":"example.com","pathPrefix":"/app"} or a custom
// scheme such as {"scheme":"myapp"}.
type DeepLink struct {
	Scheme     string `json:"scheme"`
	Host       string `json:"host,omitempty"`
	PathPrefix string `json:"pathPrefix,omitempty"`
}

// MaxDeepLinks bounds the number of deep link intent filters in the manifest.
const MaxDeepLinks = 32

var schemePattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]{0,31}$`)
var hostPattern = regexp.MustCompile(`^(\*\.)?[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)
var forbiddenSchemes = map[string]string{
	"http":       "plain HTTP is not allowed: use \"https\"",
	"javascript": "it would let any page run script in the app",
	"file":       "it would expose local files",
	"content":    "it would expose content providers",
}

// Validate checks one link and explains how to fix it.
func (l DeepLink) Validate() error {
	if !schemePattern.MatchString(l.Scheme) {
		return fmt.Errorf("invalid deepLinks scheme %q: use lowercase letters, digits, + . - starting with a letter, for example \"https\" or \"myapp\"", l.Scheme)
	}
	if why, bad := forbiddenSchemes[l.Scheme]; bad {
		return fmt.Errorf("forbidden deepLinks scheme %q: %s", l.Scheme, why)
	}
	if l.Host != "" && (len(l.Host) > 253 || !hostPattern.MatchString(l.Host)) {
		return fmt.Errorf("invalid deepLinks host %q: use a lowercase host name such as example.com (or *.example.com), without scheme, port or path", l.Host)
	}
	if l.Scheme == "https" && (l.Host == "" || !strings.Contains(strings.TrimPrefix(l.Host, "*."), ".")) {
		return fmt.Errorf("deepLinks https link needs a full domain in host, for example \"example.com\" (got %q)", l.Host)
	}
	if l.PathPrefix != "" {
		if l.Host == "" {
			return fmt.Errorf("deepLinks pathPrefix %q needs a host: Android ignores paths without one", l.PathPrefix)
		}
		if !strings.HasPrefix(l.PathPrefix, "/") || len(l.PathPrefix) > 256 {
			return fmt.Errorf("invalid deepLinks pathPrefix %q: it must start with \"/\", for example \"/app\"", l.PathPrefix)
		}
		for _, r := range l.PathPrefix {
			if r <= ' ' || r == 0x7f || r == '?' || r == '#' {
				return fmt.Errorf("invalid deepLinks pathPrefix %q: remove spaces, control characters, \"?\" and \"#\"", l.PathPrefix)
			}
		}
	}
	return nil
}

// HTTPSHosts returns the distinct domains of https deep links in order, with
// any "*." wildcard removed: Digital Asset Links are published per domain.
func (c Config) HTTPSHosts() []string {
	var out []string
	for _, l := range c.DeepLinks {
		h := strings.TrimPrefix(l.Host, "*.")
		if l.Scheme == "https" && !slices.Contains(out, h) {
			out = append(out, h)
		}
	}
	return out
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
	switch c.Orientation {
	case "", "portrait", "landscape", "sensor":
	default:
		return fmt.Errorf("invalid orientation %q: use \"portrait\", \"landscape\" or \"sensor\", or remove the field to keep Android's default", c.Orientation)
	}
	switch c.Keyboard {
	case "", "resize", "pan":
	default:
		return fmt.Errorf("invalid keyboard %q: use \"resize\" (shrink the page) or \"pan\" (slide the page up), or remove the field", c.Keyboard)
	}
	if len(c.DeepLinks) > MaxDeepLinks {
		return fmt.Errorf("too many deepLinks (%d): at most %d are supported; use pathPrefix to cover several pages", len(c.DeepLinks), MaxDeepLinks)
	}
	links := map[DeepLink]bool{}
	for _, l := range c.DeepLinks {
		if e := l.Validate(); e != nil {
			return e
		}
		if links[l] {
			return fmt.Errorf("duplicate deepLinks entry %+v: remove the repeated link", l)
		}
		links[l] = true
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
