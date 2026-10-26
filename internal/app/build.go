package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"koru.dev/koru/internal/apk"
	"koru.dev/koru/internal/arsc"
	"koru.dev/koru/internal/axml"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"koru.dev/koru/packages/javascript"
	"koru.dev/koru/pkg/dex"
	runtimedex "koru.dev/koru/runtime/dex"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Release bool
	Key     string
	NoCache bool
	Stats   *BuildStats
}

func defaultIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 96, 96))
	for y := 0; y < 96; y++ {
		for x := 0; x < 96; x++ {
			c := color.RGBA{23, 107, 73, 255}
			if x >= 26 && x <= 34 && y >= 22 && y <= 74 || x >= 34 && x <= 66 && (abs(x+y-86) < 5 || abs(x-y) < 5) {
				c = color.RGBA{227, 255, 231, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func Build(project string, opt Options) ([]byte, error) {
	c, e := config.Load(project)
	if e != nil {
		return nil, e
	}
	keyPath := c.SigningKey
	if opt.Key != "" {
		keyPath = opt.Key
	}
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(project, keyPath)
	}
	id, e := signing.LoadIdentity(keyPath)
	if e != nil {
		return nil, fmt.Errorf("load persistent signing identity: %w", e)
	}
	custom, e := c.LoadPlugins(project)
	if e != nil {
		return nil, e
	}
	permissions := c.EffectivePermissions()
	for _, p := range custom {
		for _, permission := range p.Permissions {
			found := false
			for _, s := range permissions {
				found = found || s == permission
			}
			if !found {
				permissions = append(permissions, permission)
			}
		}
	}
	root := c.WebDir
	if !filepath.IsAbs(root) {
		root = filepath.Join(project, root)
	}
	if _, e := os.Stat(filepath.Join(root, "index.html")); errors.Is(e, fs.ErrNotExist) {
		return nil, fmt.Errorf("webDir %q has no index.html; if a bundler writes it, run your web build first (for example npm run build)", c.WebDir)
	}
	assets, e := Assets(root)
	if e != nil {
		return nil, e
	}
	if opt.Stats != nil {
		opt.Stats.Assets = len(assets) + 2
	}
	assets = append(assets, apk.Entry{Name: "assets/www/__koru/core.js", Data: javascript.Core}, apk.Entry{Name: "assets/www/__koru/core.mjs", Data: javascript.Module})
	runtimeAssets := make([]runtimedex.Asset, 0, len(assets))
	for _, a := range assets {
		path := strings.TrimPrefix(a.Name, "assets/www")
		runtimeAssets = append(runtimeAssets, runtimedex.Asset{Path: path, MIME: assetMIME(path)})
	}
	dexBytes, e := runtimeDEX(project, c, runtimeAssets, custom, opt.Release, opt.NoCache, opt.Stats)
	if e != nil {
		return nil, e
	}
	links := make([]axml.DeepLink, 0, len(c.DeepLinks))
	for _, l := range c.DeepLinks {
		links = append(links, axml.DeepLink{Scheme: l.Scheme, Host: l.Host, PathPrefix: l.PathPrefix})
	}
	manifest, e := axml.Build(axml.Manifest{Package: c.AppID, Name: c.Name, Version: c.Version, VersionCode: c.VersionCode, MinSDK: c.MinSDK, TargetSDK: c.TargetSDK, Activity: "com.koru.runtime.MainActivity", Permissions: permissions, Debuggable: !opt.Release, Icon: arsc.IconID, Orientation: c.Orientation, AllowBackup: c.AllowBackup, Keyboard: c.Keyboard, DeepLinks: links})
	if e != nil {
		return nil, e
	}
	resources, e := arsc.Build(c.AppID)
	if e != nil {
		return nil, e
	}
	icon := defaultIcon()
	if c.Icon != "" {
		p := c.Icon
		if !filepath.IsAbs(p) {
			p = filepath.Join(project, p)
		}
		icon, e = os.ReadFile(p)
		if e != nil {
			return nil, e
		}
		if _, e = png.DecodeConfig(bytes.NewReader(icon)); e != nil {
			return nil, errors.New("icon must be PNG")
		}
	}
	entries := append(assets, apk.Entry{Name: "AndroidManifest.xml", Data: manifest, Store: true, Alignment: 4}, apk.Entry{Name: "classes.dex", Data: dexBytes, Store: true, Alignment: 4}, apk.Entry{Name: "resources.arsc", Data: resources, Store: true, Alignment: 4}, apk.Entry{Name: "res/mipmap/ic_launcher.png", Data: icon})
	native, e := nativeEntries(project, c.NativeLibraries)
	if e != nil {
		return nil, e
	}
	entries = append(entries, native...)
	total := uint64(0)
	for _, entry := range entries {
		total += uint64(len(entry.Data))
	}
	if total > 256<<20 {
		return nil, errors.New("APK content exceeds 256 MiB limit")
	}
	unsigned, e := apk.Pack(entries)
	if e != nil {
		return nil, e
	}
	signed, e := signing.SignV2(unsigned, id)
	if e != nil {
		return nil, e
	}
	if _, e = Verify(signed); e != nil {
		return nil, fmt.Errorf("verify generated APK: %w", e)
	}
	return signed, nil
}

type Report struct {
	Package           string `json:"appId"`
	Entries           int    `json:"entries"`
	Bytes             int    `json:"bytes"`
	CertificateSHA256 string `json:"certificateSha256"`
	SignatureScheme   string `json:"signatureScheme"`
	// DeepLinks lists the manifest URL patterns as scheme://host/pathPrefix.
	DeepLinks []string `json:"deepLinks,omitempty"`
}

func Verify(b []byte) (Report, error) {
	var out Report
	files, cert, e := apk.ReadVerified(b)
	if e != nil {
		return out, e
	}
	for _, s := range []string{"AndroidManifest.xml", "classes.dex", "assets/www/index.html", "resources.arsc"} {
		if len(files[s]) == 0 {
			return out, fmt.Errorf("APK missing %s", s)
		}
	}
	els, e := axml.Parse(files["AndroidManifest.xml"])
	if e != nil {
		return out, e
	}
	for _, a := range els[0].Attributes {
		if a.Name == "package" {
			out.Package = a.Text
		}
	}
	if out.Package == "" {
		return out, errors.New("manifest package missing")
	}
	for _, l := range axml.Links(els) {
		out.DeepLinks = append(out.DeepLinks, l.Scheme+"://"+l.Host+l.PathPrefix)
	}
	resources, e := arsc.Parse(files["resources.arsc"])
	if e != nil {
		return out, e
	}
	if resources.Package != out.Package || len(files[resources.IconPath]) == 0 {
		return out, errors.New("resource package/icon mismatch")
	}
	if _, e = dex.Inspect(files["classes.dex"]); e != nil {
		return out, e
	}
	sum := sha256.Sum256(cert)
	out.Entries = len(files)
	out.Bytes = len(b)
	out.CertificateSHA256 = hex.EncodeToString(sum[:])
	out.SignatureScheme = "APK v2 RSA/SHA-256"
	return out, nil
}

func assetMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm":
		return "text/html"
	case ".js", ".mjs":
		return "text/javascript"
	case ".css":
		return "text/css"
	case ".json", ".map":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".wasm":
		return "application/wasm"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
