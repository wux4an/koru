package app

import (
	"errors"
	"fmt"
	"io/fs"
	"koru.dev/koru/internal/apk"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxAssets = 256 << 20

var privateKeyPattern = regexp.MustCompile(`-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----`)

// InjectHead inserts snippet before the first </head>, or at the start of a
// document without one.
func InjectHead(html, snippet string) string {
	if p := indexFoldASCII(html, "</head>"); p >= 0 {
		return html[:p] + snippet + html[p:]
	}
	return snippet + html
}

// indexFoldASCII finds sub in s ignoring ASCII case only, so the offset is
// valid for s; strings.ToLower can change byte lengths of non-ASCII runes.
func indexFoldASCII(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		j := 0
		for ; j < len(sub); j++ {
			c := s[i+j]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != sub[j] {
				break
			}
		}
		if j == len(sub) {
			return i
		}
	}
	return -1
}

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
		if privateKeyPattern.Match(b) {
			return errors.New("private key found in web assets")
		}
		if strings.HasSuffix(strings.ToLower(name), ".html") {
			b = []byte(InjectHead(string(b), `<script src="/__koru/core.js"></script>`))
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
