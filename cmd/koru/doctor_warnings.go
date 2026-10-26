package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"koru.dev/koru/internal/config"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	internetPermission = "android.permission.INTERNET"
	minIconSize        = 192
	maxScanFiles       = 2000
	maxScanBytes       = 2 << 20
)

var viewportPattern = regexp.MustCompile(`(?i)<meta\s[^>]*name\s*=\s*["']?viewport`)

// Heuristic: a fetch call whose first argument is a literal absolute http(s) URL.
var remoteFetchPattern = regexp.MustCompile("fetch\\(\\s*[\"'`]https?://")

// doctorWarnings returns advisory findings for a valid project. They never fail the check.
func doctorWarnings(project, webRoot string, c config.Config) ([]string, error) {
	var out []string
	w, err := ignoreFileWarning(project)
	if err != nil {
		return nil, err
	}
	if w != "" {
		out = append(out, w)
	}
	index, err := readLimit(filepath.Join(webRoot, "index.html"), maxScanBytes)
	if err != nil {
		return nil, err
	}
	if !viewportPattern.Match(index) {
		out = append(out, `index.html has no <meta name="viewport" content="width=device-width,initial-scale=1">; the page renders at desktop width on phones. Add it inside <head>.`)
	}
	if !c.HasPermission(internetPermission) {
		file, err := findRemoteFetch(webRoot)
		if err != nil {
			return nil, err
		}
		if file != "" {
			out = append(out, fmt.Sprintf("%s calls fetch() with an https:// URL but %s is not in permissions (heuristic text match, not a full analysis). Add it to \"permissions\" in koru.json if the app needs the network.", file, internetPermission))
		}
	}
	if c.Icon != "" {
		iw, err := iconWarning(project, c.Icon)
		if err != nil {
			return nil, err
		}
		out = append(out, iw...)
	}
	if c.TargetSDK != 30 {
		out = append(out, fmt.Sprintf("targetSdk is %d; the recorded Android execution matrix covers API 30 only (docs/results/full). Test on a device before shipping.", c.TargetSDK))
	}
	return out, nil
}

// ignoreFileWarning checks that the version-control ignore file excludes .koru/.
// It only applies to projects that have a repository or an ignore file.
func ignoreFileWarning(project string) (string, error) {
	_, repoErr := os.Stat(filepath.Join(project, ".git"))
	f, ignoreErr := os.Open(filepath.Join(project, ".gitignore"))
	if os.IsNotExist(repoErr) && os.IsNotExist(ignoreErr) {
		return "", nil
	}
	if repoErr != nil && !os.IsNotExist(repoErr) {
		return "", repoErr
	}
	if ignoreErr == nil {
		defer f.Close()
		sc := bufio.NewScanner(io.LimitReader(f, 1<<20))
		for sc.Scan() {
			switch strings.TrimSpace(sc.Text()) {
			case ".koru", ".koru/", "/.koru", "/.koru/", ".koru/*", "/.koru/*":
				return "", nil
			}
		}
		if err := sc.Err(); err != nil {
			return "", err
		}
	} else if !os.IsNotExist(ignoreErr) {
		return "", ignoreErr
	}
	return ".koru/ is not listed in .gitignore; it holds your signing key. Add a line \".koru/\" to .gitignore and keep a private backup of the key.", nil
}

func findRemoteFetch(webRoot string) (string, error) {
	var found string
	count := 0
	err := filepath.WalkDir(webRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found != "" || !d.Type().IsRegular() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".js", ".mjs", ".html":
		default:
			return nil
		}
		if count++; count > maxScanFiles {
			return fs.SkipAll
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		b, err := io.ReadAll(io.LimitReader(f, maxScanBytes))
		if err != nil {
			return err
		}
		if remoteFetchPattern.Match(b) {
			rel, err := filepath.Rel(webRoot, path)
			if err != nil {
				rel = path
			}
			found = filepath.ToSlash(rel)
		}
		return nil
	})
	return found, err
}

func iconWarning(project, icon string) ([]string, error) {
	path := icon
	if !filepath.IsAbs(path) {
		path = filepath.Join(project, path)
	}
	b, err := readLimit(path, maxInputSize)
	if err != nil {
		return nil, fmt.Errorf("icon %q: %w; fix the \"icon\" path in koru.json", icon, err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("icon %q must be a PNG file: %w", icon, err)
	}
	var out []string
	if cfg.Width != cfg.Height {
		out = append(out, fmt.Sprintf("icon %s is %dx%d; use a square PNG or Android will distort it.", icon, cfg.Width, cfg.Height))
	}
	if cfg.Width < minIconSize || cfg.Height < minIconSize {
		out = append(out, fmt.Sprintf("icon %s is %dx%d; use at least %dx%d px so it stays sharp on high-density screens.", icon, cfg.Width, cfg.Height, minIconSize, minIconSize))
	}
	return out, nil
}
