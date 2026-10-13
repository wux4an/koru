package main

import (
	"fmt"
	"io"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
)

func doctor(args []string, out io.Writer) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: koru doctor [project]")
	}
	_, e := fmt.Fprintln(out, "Build/verify: self-contained Go implementation; no Java, SDK or external build tools.")
	if e != nil {
		return e
	}
	project := "."
	if len(args) == 1 {
		project = args[0]
	} else {
		if _, e := os.Stat(filepath.Join(project, "koru.json")); os.IsNotExist(e) {
			_, e = fmt.Fprintln(out, "No project selected. Android device/emulator is optional and used only for execution tests.")
			return e
		}
	}
	c, e := config.Load(project)
	if e != nil {
		return e
	}
	key := c.SigningKey
	if !filepath.IsAbs(key) {
		key = filepath.Join(project, key)
	}
	if _, e = signing.LoadIdentity(key); e != nil {
		return e
	}
	root := c.WebDir
	if !filepath.IsAbs(root) {
		root = filepath.Join(project, root)
	}
	info, e := os.Lstat(filepath.Join(root, "index.html"))
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("web index must be a regular file")
	}
	if _, e = c.LoadPlugins(project); e != nil {
		return e
	}
	_, e = fmt.Fprintf(out, "Project %s: valid config, persistent signing identity, web index and plugin manifests.\n", c.AppID)
	return e
}
