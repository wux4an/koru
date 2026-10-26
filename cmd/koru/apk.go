package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
	"strings"
)

func apkBuild(project, output, key string, noCache, release bool, out io.Writer) error {
	c, err := config.Load(project)
	if err != nil {
		return err
	}
	if strings.ToLower(filepath.Ext(output)) != ".apk" {
		return errors.New("APK output must use the .apk extension")
	}
	webRoot := c.WebDir
	if !filepath.IsAbs(webRoot) {
		webRoot = filepath.Join(project, webRoot)
	}
	webRoot, err = filepath.Abs(webRoot)
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(webRoot, outputPath)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("APK output must be outside webDir")
	}
	keyPath := c.SigningKey
	if key != "" {
		keyPath = key
	}
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(project, keyPath)
	}
	if info, err := os.Stat(outputPath); err == nil {
		if keyInfo, err := os.Stat(keyPath); err == nil && os.SameFile(info, keyInfo) {
			return errors.New("APK output must not replace the signing identity")
		}
	}
	stats := app.BuildStats{}
	b, err := app.Build(project, app.Options{Release: release, Key: key, NoCache: noCache, Stats: &stats})
	if err != nil {
		return err
	}
	report, err := app.Verify(b)
	if err != nil {
		return err
	}
	if err = atomicWrite(output, b); err != nil {
		return err
	}
	if err = writeStatus(out, "Runtime cache hit:", fmt.Sprint(stats.RuntimeCacheHit), "iris"); err != nil {
		return err
	}
	return writeStatus(out, "Wrote", fmt.Sprintf("%s (%d bytes), %s, certificate %s", output, len(b), report.SignatureScheme, report.CertificateSHA256), "mint")
}

func initProject(args []string, out io.Writer) error {
	path := "."
	if len(args) == 1 {
		path = args[0]
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(path, "koru.json")
	if _, err := os.Lstat(configPath); !os.IsNotExist(err) {
		return errors.New("project configuration exists or cannot be inspected")
	}
	for _, dir := range []string{"www", ".koru"} {
		if err := os.MkdirAll(filepath.Join(path, dir), 0700); err != nil {
			return err
		}
	}
	if err := ignoreKoruDir(path); err != nil {
		return err
	}
	c := config.Default()
	c.Schema = config.SchemaURL
	if _, err := signing.CreateIdentity(filepath.Join(path, c.SigningKey)); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	html := `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Koru</title><style>body{font-family:system-ui;padding:2rem}button{padding:1rem;font:inherit}</style></head><body><h1>Koru</h1><p id="status">A standalone local app</p><button id="counter">Count: 0</button><script>let n=0;document.querySelector('#counter').onclick=e=>e.target.textContent='Count: '+(++n);</script></body></html>`
	entry := filepath.Join(path, "www/index.html")
	f, err = os.OpenFile(entry, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(html))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return writeStatus(out, "Initialized", fmt.Sprintf("%s; persistent signing identity: %s", path, c.SigningKey), "mint")
}

// ignoreKoruDir keeps the signing identity and cache out of version control,
// appending to an existing .gitignore instead of replacing it.
func ignoreKoruDir(project string) error {
	path := filepath.Join(project, ".gitignore")
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, line := range strings.Split(string(b), "\n") {
		switch strings.TrimSpace(line) {
		case ".koru", ".koru/", "/.koru", "/.koru/", "**/.koru/":
			return nil
		}
	}
	entry := ".koru/\n"
	if len(b) > 0 && b[len(b)-1] != '\n' {
		entry = "\n" + entry
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(entry)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

func keygen(path string, out io.Writer) error {
	if path == "" {
		return errors.New("an output path is required")
	}
	if _, err := signing.CreateIdentity(path); err != nil {
		return err
	}
	return writeStatus(out, "Created signing identity", path, "mint")
}
