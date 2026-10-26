package main

import (
	"context"
	"fmt"
	"io"
	"koru.dev/koru/internal/adb"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
	"time"
)

func doctor(args []string, out io.Writer) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: koru doctor [project]")
	}
	if e := writeStatus(out, "Build and verification", "use the self-contained Go implementation; no Java, SDK, or external build tools are required.", "iris"); e != nil {
		return e
	}
	if e := doctorDevices(out); e != nil {
		return e
	}
	project := "."
	if len(args) == 1 {
		project = args[0]
	} else {
		if _, e := os.Stat(filepath.Join(project, "koru.json")); os.IsNotExist(e) {
			return writeStatus(out, "No project selected.", "Run koru doctor inside a project, or create one with koru create.", "peach")
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
	if e = writeStatus(out, "Project", fmt.Sprintf("%s: configuration, persistent signing identity, web entry point, and plugin manifests are valid.", c.AppID), "mint"); e != nil {
		return e
	}
	warnings, e := doctorWarnings(project, root, c)
	if e != nil {
		return e
	}
	for _, w := range warnings {
		if e = writeStatus(out, "Warning", w, "peach"); e != nil {
			return e
		}
	}
	// The default targetSdk differs from the recorded matrix, so this is a
	// note rather than a warning every new project would show.
	if c.TargetSDK != 30 {
		return writeStatus(out, "Note", fmt.Sprintf("recorded Android execution covers API 30 (targetSdk here is %d); test on your devices before shipping.", c.TargetSDK), "iris")
	}
	return nil
}

// doctorDevices reports the adb server and devices. They are only needed by
// koru run, so their absence is information, not an error.
func doctorDevices(out io.Writer) error {
	addr, err := adb.Address()
	if err != nil {
		return writeStatus(out, "Devices", err.Error(), "peach")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	devices, err := adb.Client{Addr: addr, Timeout: 3 * time.Second}.Devices(ctx)
	if err != nil {
		return writeStatus(out, "Devices", "no adb server at "+addr+"; only koru run needs one (adb start-server or Android Studio)", "peach")
	}
	ready := 0
	for _, d := range devices {
		if d.State == "device" {
			ready++
		}
	}
	return writeStatus(out, "Devices", fmt.Sprintf("%d ready of %d connected (adb server %s); see koru devices", ready, len(devices), addr), "mint")
}
