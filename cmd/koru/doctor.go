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
	if e := writeStatus(out, "Build and verification", "use the self-contained Go implementation; no Java, SDK, or external build tools are required.", "iris"); e != nil {
		return e
	}
	project := "."
	if len(args) == 1 {
		project = args[0]
	} else {
		if _, e := os.Stat(filepath.Join(project, "koru.json")); os.IsNotExist(e) {
			return writeStatus(out, "No project selected.", "An Android device or emulator is optional and used only for execution tests.", "peach")
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
	return writeStatus(out, "Project", fmt.Sprintf("%s: configuration, persistent signing identity, web entry point, and plugin manifests are valid.", c.AppID), "mint")
}
