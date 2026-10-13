package app

import (
	"encoding/binary"
	"errors"
	"fmt"
	"koru.dev/koru/internal/apk"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func nativeEntries(project string, libs map[string][]string) ([]apk.Entry, error) {
	var entries []apk.Entry
	seen := map[string]bool{}
	total := int64(0)
	abis := make([]string, 0, len(libs))
	for abi := range libs {
		abis = append(abis, abi)
	}
	sort.Strings(abis)
	for _, abi := range abis {
		class := byte(2)
		machine := uint16(0)
		switch abi {
		case "arm64-v8a":
			machine = 183
		case "x86_64":
			machine = 62
		case "armeabi-v7a":
			class = 1
			machine = 40
		case "x86":
			class = 1
			machine = 3
		default:
			return nil, fmt.Errorf("unsupported native ABI %s", abi)
		}
		if len(libs[abi]) > 16 {
			return nil, errors.New("native library count limit")
		}
		for _, path := range libs[abi] {
			nameKey := abi + "/" + filepath.Base(path)
			if seen[nameKey] {
				return nil, errors.New("duplicate native library")
			}
			seen[nameKey] = true
			if !filepath.IsAbs(path) {
				path = filepath.Join(project, path)
			}
			info, e := os.Lstat(path)
			if e != nil {
				return nil, e
			}
			if !info.Mode().IsRegular() || info.Size() > 32<<20 {
				return nil, errors.New("invalid native library file")
			}
			total += info.Size()
			if total > 64<<20 {
				return nil, errors.New("native library total size limit")
			}
			b, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			name := filepath.Base(path)
			if !strings.HasPrefix(name, "lib") || !strings.HasSuffix(name, ".so") || !apk.SafeName(name) || len(b) < 64 || string(b[:4]) != "\x7fELF" || b[4] != class || b[5] != 1 || binary.LittleEndian.Uint16(b[16:]) != 3 || binary.LittleEndian.Uint16(b[18:]) != machine {
				return nil, errors.New("native library is not an ELF shared object for configured ABI")
			}
			entries = append(entries, apk.Entry{Name: "lib/" + abi + "/" + name, Data: b, Store: true, Alignment: 16384})
		}
	}
	return entries, nil
}
