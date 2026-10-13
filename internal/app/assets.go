package app

import (
	"errors"
	"fmt"
	"io/fs"
	"koru.dev/koru/internal/apk"
	"os"
	"path/filepath"
	"strings"
)

const maxAssets = 256 << 20

func Assets(root string) ([]apk.Entry, error) {
	var out []apk.Entry
	total := 0
	index := false
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("asset symlink rejected: %s", path)
		}
		if d.IsDir() {
			if path != root && (d.Name() == ".git" || d.Name() == ".koru" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("non-regular asset: %s", path)
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		name := filepath.ToSlash(rel)
		if !apk.SafeName(name) {
			return errors.New("unsafe asset path")
		}
		if name == "index.html" {
			index = true
		}
		if name == "__koru/core.js" || name == "__koru/core.mjs" {
			return errors.New("reserved __koru asset path")
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if info.Size() > 64<<20 || info.Size() < 0 || int64(total)+info.Size() > maxAssets {
			return errors.New("web assets exceed build limits")
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if strings.Contains(string(b), "-----BEGIN PRIVATE KEY-----") || strings.Contains(string(b), "-----BEGIN RSA PRIVATE KEY-----") {
			return errors.New("private key found in web assets")
		}
		if strings.HasSuffix(strings.ToLower(name), ".html") {
			inject := `<script src="/__koru/core.js"></script>`
			s := string(b)
			lower := strings.ToLower(s)
			if p := strings.Index(lower, "</head>"); p >= 0 {
				s = s[:p] + inject + s[p:]
			} else {
				s = inject + s
			}
			b = []byte(s)
		}
		total += len(b)
		out = append(out, apk.Entry{Name: "assets/www/" + name, Data: b})
		if len(out) > 2000 {
			return errors.New("initial runtime supports at most 2000 web assets")
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	if !index {
		return nil, errors.New("webDir must contain index.html")
	}
	return out, nil
}
