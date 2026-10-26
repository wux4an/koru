// Package templates writes starter projects for `koru create`.
package templates

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/packages/javascript"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

//go:embed all:files
var files embed.FS

type Template struct {
	ID          string
	Description string
	WebDir      string
	Plugins     []string
	// Vendor copies the koru JavaScript module into the project, for bundlers
	// that resolve imports at build time.
	Vendor string
}

var List = []Template{
	{ID: "basic", Description: "HTML, CSS and JavaScript, no build step or Node.js needed", WebDir: "www", Plugins: []string{"device", "storage", "app", "vibration", "share"}},
	{ID: "vite", Description: "Vite project with hot reload (needs Node.js)", WebDir: "dist", Plugins: []string{"device", "storage", "app"}, Vendor: "src/koru"},
}

func Find(id string) (Template, error) {
	var ids []string
	for _, t := range List {
		if t.ID == id {
			return t, nil
		}
		ids = append(ids, t.ID)
	}
	return Template{}, fmt.Errorf("unknown template %q (available: %s)", id, strings.Join(ids, ", "))
}

type Options struct {
	Template string
	Name     string
	AppID    string
}

// Slug turns a display name into a lowercase file and package name.
func Slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
			dash = false
		} else if b.Len() > 0 && !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	s := strings.TrimRight(b.String(), "-")
	if s == "" {
		return "app"
	}
	return s
}

// AppID derives a valid Android package name from a display name.
func AppID(name string) string {
	segment := strings.ReplaceAll(Slug(name), "-", "")
	if segment[0] >= '0' && segment[0] <= '9' {
		segment = "app" + segment
	}
	return "com.example." + segment
}

// DisplayName turns a directory name such as my-app into "My App".
func DisplayName(dir string) string {
	words := strings.FieldsFunc(filepath.Base(dir), func(r rune) bool { return r == '-' || r == '_' || r == ' ' || r == '.' })
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	if len(words) == 0 {
		return "My App"
	}
	return strings.Join(words, " ")
}

// Write creates the project files in dir, which must be missing or empty.
// It does not create the signing identity.
func Write(dir string, o Options) (config.Config, error) {
	t, err := Find(o.Template)
	if err != nil {
		return config.Config{}, err
	}
	c := config.Default()
	c.Schema = config.SchemaURL
	c.Name = o.Name
	c.AppID = o.AppID
	c.WebDir = t.WebDir
	c.Plugins = append([]string(nil), t.Plugins...)
	if err = c.Validate(); err != nil {
		return c, err
	}
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return c, fmt.Errorf("%s already exists and is not empty; choose a new directory", dir)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return c, err
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		return c, err
	}
	slug := Slug(o.Name)
	root := path.Join("files", t.ID)
	err = fs.WalkDir(files, root, func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(name, root+"/")
		if rel == "gitignore" {
			rel = ".gitignore"
		}
		appName := o.Name
		switch path.Ext(rel) {
		case ".html":
			appName = html.EscapeString(o.Name)
		case ".json":
			quoted, _ := json.Marshal(o.Name)
			appName = string(quoted[1 : len(quoted)-1])
		}
		text := strings.NewReplacer("__APP_NAME__", appName, "__APP_SLUG__", slug).Replace(string(data))
		return writeNew(filepath.Join(dir, filepath.FromSlash(rel)), []byte(text))
	})
	if err != nil {
		return c, err
	}
	if t.Vendor != "" {
		for name, data := range map[string][]byte{"core.js": javascript.Core, "core.mjs": javascript.Module, "core.d.mts": javascript.Types} {
			if err = writeNew(filepath.Join(dir, filepath.FromSlash(t.Vendor), name), data); err != nil {
				return c, err
			}
		}
	} else if err = writeNew(filepath.Join(dir, "types", "koru.d.mts"), javascript.Types); err != nil {
		return c, err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return c, err
	}
	return c, writeNew(filepath.Join(dir, "koru.json"), append(b, '\n'))
}

func writeNew(name string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
