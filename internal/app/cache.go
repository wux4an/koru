package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/pkg/dex"
	"koru.dev/koru/plugins"
	runtimedex "koru.dev/koru/runtime/dex"
	"os"
	"path/filepath"
	"strings"
)

type BuildStats struct {
	RuntimeCacheHit bool `json:"runtimeCacheHit"`
	DEXBytes        int  `json:"dexBytes"`
	Assets          int  `json:"assets"`
}

func runtimeDEX(project string, c config.Config, assets []runtimedex.Asset, custom []plugins.Descriptor, release bool, noCache bool, stats *BuildStats) ([]byte, error) {
	exe, e := os.Executable()
	if e != nil {
		return nil, e
	}
	tool, e := os.ReadFile(exe)
	if e != nil {
		return nil, e
	}
	toolSHA := sha256.Sum256(tool)
	fingerprint, e := json.Marshal(struct {
		Config  config.Config
		Assets  []runtimedex.Asset
		Plugins []plugins.Descriptor
		Release bool
		ToolSHA [32]byte
	}{c, assets, custom, release, toolSHA})
	if e != nil {
		return nil, e
	}
	sum := sha256.Sum256(fingerprint)
	cacheDir := filepath.Join(project, ".koru", "cache")
	path := filepath.Join(cacheDir, hex.EncodeToString(sum[:])+".dex")
	if !noCache {
		for _, dir := range []string{filepath.Join(project, ".koru"), cacheDir} {
			if info, e := os.Lstat(dir); e == nil {
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return nil, errors.New("runtime cache directory must not be a symlink")
				}
			} else if !os.IsNotExist(e) {
				return nil, e
			}
		}
		if info, e := os.Lstat(path); e == nil {
			if !info.Mode().IsRegular() || info.Size() > 64<<20 {
				return nil, errors.New("unsafe runtime cache")
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			checkInfo, e := os.Lstat(path + ".sha256")
			if e != nil || !checkInfo.Mode().IsRegular() || checkInfo.Size() > 128 {
				return nil, errors.New("invalid runtime cache checksum file")
			}
			expected, e := os.ReadFile(path + ".sha256")
			if e != nil {
				return nil, errors.New("runtime cache checksum missing; remove the cache explicitly")
			}
			hash := sha256.Sum256(data)
			if strings.TrimSpace(string(expected)) != hex.EncodeToString(hash[:]) {
				return nil, errors.New("runtime cache checksum mismatch; remove the cache explicitly")
			}
			if _, e = dex.Inspect(data); e != nil {
				return nil, e
			}
			if stats != nil {
				stats.RuntimeCacheHit = true
				stats.DEXBytes = len(data)
			}
			return data, nil
		} else if !os.IsNotExist(e) {
			return nil, e
		}
	}
	program := runtimedex.Runtime(c, assets, custom...)
	for j := range program.Classes[0].Methods {
		m := &program.Classes[0].Methods[j]
		if m.Name == "onCreate" {
			debug := int32(0)
			if !release {
				debug = 1
			}
			inject := []dex.Instruction{{Op: "const/4", Regs: []dex.Register{2}, Literal: debug}, {Op: "invoke-static", Regs: []dex.Register{2}, Method: &dex.MethodRef{Owner: "Landroid/webkit/WebView;", Name: "setWebContentsDebuggingEnabled", Descriptor: "(Z)V"}}}
			m.Code.Instructions = append(m.Code.Instructions[:1], append(inject, m.Code.Instructions[1:]...)...)
		}
	}
	data, e := dex.Build(program)
	if e != nil {
		return nil, e
	}
	if stats != nil {
		stats.DEXBytes = len(data)
	}
	if !noCache {
		if e = os.MkdirAll(cacheDir, 0700); e != nil {
			return nil, e
		}
		if e = os.WriteFile(path, data, 0600); e != nil {
			return nil, e
		}
		hash := sha256.Sum256(data)
		if e = os.WriteFile(path+".sha256", []byte(hex.EncodeToString(hash[:])), 0600); e != nil {
			return nil, e
		}
	}
	return data, nil
}
