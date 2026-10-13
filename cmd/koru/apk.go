package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
	"strings"
)

func apkBuild(args []string, out io.Writer) error {
	project := "."
	var flags []string
	found := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--output" || a == "--keystore" {
			if i+1 == len(args) {
				return fmt.Errorf("%s requires value", a)
			}
			flags = append(flags, a, args[i+1])
			i++
		} else if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
		} else {
			if found {
				return errors.New("one project directory required")
			}
			project = a
			found = true
		}
	}
	f := flag.NewFlagSet("build", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	output := f.String("output", "app.apk", "signed APK output")
	key := f.String("keystore", "", "existing PEM signing identity")
	noCache := f.Bool("no-cache", false, "disable runtime DEX cache")
	release := f.Bool("release", false, "disable debug mode, reuse persistent identity")
	if e := f.Parse(flags); e != nil {
		return e
	}
	c, e := config.Load(project)
	if e != nil {
		return e
	}
	if strings.ToLower(filepath.Ext(*output)) != ".apk" {
		return errors.New("APK output must use .apk extension")
	}
	webRoot := c.WebDir
	if !filepath.IsAbs(webRoot) {
		webRoot = filepath.Join(project, webRoot)
	}
	webRoot, e = filepath.Abs(webRoot)
	if e != nil {
		return e
	}
	outputPath, e := filepath.Abs(*output)
	if e != nil {
		return e
	}
	rel, e := filepath.Rel(webRoot, outputPath)
	if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("APK output must be outside webDir")
	}
	keyPath := c.SigningKey
	if *key != "" {
		keyPath = *key
	}
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(project, keyPath)
	}
	if info, e := os.Stat(outputPath); e == nil {
		if keyInfo, e := os.Stat(keyPath); e == nil && os.SameFile(info, keyInfo) {
			return errors.New("APK output must not replace signing identity")
		}
	}
	stats := app.BuildStats{}
	b, e := app.Build(project, app.Options{Release: *release, Key: *key, NoCache: *noCache, Stats: &stats})
	if e != nil {
		return e
	}
	if e = atomicWrite(*output, b); e != nil {
		return e
	}
	report, e := app.Verify(b)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(out, "runtime cache hit: %t\n", stats.RuntimeCacheHit)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(out, "wrote %s (%d bytes), %s, certificate %s\n", *output, len(b), report.SignatureScheme, report.CertificateSHA256)
	return e
}
func initProject(args []string, out io.Writer) error {
	if len(args) > 1 {
		return errors.New("usage: koru init [directory]")
	}
	path := "."
	if len(args) == 1 {
		path = args[0]
	}
	if e := os.MkdirAll(path, 0755); e != nil {
		return e
	}
	configPath := filepath.Join(path, "koru.json")
	if _, e := os.Lstat(configPath); !os.IsNotExist(e) {
		return errors.New("project configuration exists or cannot be inspected")
	}
	for _, dir := range []string{"www", ".koru"} {
		if e := os.MkdirAll(filepath.Join(path, dir), 0700); e != nil {
			return e
		}
	}
	c := config.Default()
	if _, e := signing.CreateIdentity(filepath.Join(path, c.SigningKey)); e != nil {
		return e
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(append(b, '\n'))
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	html := `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Koru</title><style>body{font-family:system-ui;padding:2rem}button{padding:1rem;font:inherit}</style></head><body><h1>Koru</h1><p id="status">Aplicación local independiente</p><button id="counter">Contador: 0</button><script>let n=0;document.querySelector('#counter').onclick=e=>e.target.textContent='Contador: '+(++n);</script></body></html>`
	entry := filepath.Join(path, "www/index.html")
	f, e = os.OpenFile(entry, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write([]byte(html))
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(out, "initialized %s; persistent signing identity: %s\n", path, c.SigningKey)
	return e
}
func keygen(args []string, out io.Writer) error {
	f := flag.NewFlagSet("keygen", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	path := f.String("output", "", "new identity path")
	if e := f.Parse(args); e != nil {
		return e
	}
	if *path == "" || f.NArg() != 0 {
		return errors.New("usage: koru keygen --output signing.pem")
	}
	if _, e := signing.CreateIdentity(*path); e != nil {
		return e
	}
	_, e := fmt.Fprintln(out, "created signing identity", *path)
	return e
}
