// koru compiles and inspects DEX without executing any external build tools.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/pkg/dex"
	runtimedex "koru.dev/koru/runtime/dex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "0.2.0-dev"
const maxInputSize = 64 << 20

func main() {
	if e := run(os.Args[1:], os.Stdout); e != nil {
		fmt.Fprintln(os.Stderr, "koru:", e)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: koru dex build <ir.json> --output classes.dex | dex inspect <file> | doctor | version")
	}
	switch args[0] {
	case "version":
		if len(args) != 1 {
			return errors.New("version accepts no arguments")
		}
		_, e := fmt.Fprintf(out, "koru %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return e
	case "doctor":
		return doctor(args[1:], out)
	case "build":
		return apkBuild(args[1:], out)
	case "init":
		return initProject(args[1:], out)
	case "keygen":
		return keygen(args[1:], out)
	case "verify":
		if len(args) != 2 {
			return errors.New("usage: koru verify app.apk")
		}
		b, e := readLimit(args[1], 320<<20)
		if e != nil {
			return e
		}
		report, e := app.Verify(b)
		if e != nil {
			return e
		}
		return json.NewEncoder(out).Encode(report)
	case "dex":
		if len(args) < 2 {
			return errors.New("expected dex build or dex inspect")
		}
		switch args[1] {
		case "build":
			return build(args[2:], out)
		case "inspect":
			if len(args) != 3 {
				return errors.New("usage: koru dex inspect <file.dex>")
			}
			b, e := readBounded(args[2])
			if e != nil {
				return e
			}
			f, e := dex.Inspect(b)
			if e != nil {
				return e
			}
			encoder := json.NewEncoder(out)
			encoder.SetIndent("", "  ")
			return encoder.Encode(f)
		default:
			return fmt.Errorf("unknown dex command %q", args[1])
		}
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
func readBounded(path string) ([]byte, error) { return readLimit(path, maxInputSize) }
func readLimit(path string, limit int64) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !s.Mode().IsRegular() {
		return nil, errors.New("input must be a regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("input exceeds CLI %d byte limit", limit)
	}
	return b, nil
}
func build(args []string, out io.Writer) error {
	// Go flag stops at the first positional argument; support the advertised
	// `koru dex build input.json --output result.dex` syntax explicitly.
	var source string
	flags := []string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--output" || a == "-output" {
			if i+1 == len(args) {
				return errors.New("--output requires a path")
			}
			flags = append(flags, a, args[i+1])
			i++
			continue
		}
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)
		} else {
			if source != "" {
				return errors.New("only one IR file is allowed")
			}
			source = a
		}
	}
	fs := flag.NewFlagSet("dex build", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	output := fs.String("output", "classes.dex", "output path")
	experiment := fs.Bool("activity-experiment", false, "emit experimental WebView Activity")
	if e := fs.Parse(flags); e != nil {
		return e
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected argument")
	}
	var p dex.Program
	if *experiment {
		if source != "" {
			return errors.New("activity experiment takes no IR input")
		}
		p = runtimedex.ActivityExperiment()
	} else {
		if source == "" {
			return errors.New("usage: koru dex build <ir.json> --output classes.dex")
		}
		b, e := readBounded(source)
		if e != nil {
			return e
		}
		dec := json.NewDecoder(strings.NewReader(string(b)))
		dec.DisallowUnknownFields()
		if e = dec.Decode(&p); e != nil {
			return e
		}
		var extra any
		if e = dec.Decode(&extra); e != io.EOF {
			return errors.New("trailing JSON input")
		}
	}
	if source != "" {
		a, e := filepath.Abs(source)
		if e != nil {
			return e
		}
		b, e := filepath.Abs(*output)
		if e != nil {
			return e
		}
		if a == b {
			return errors.New("output must differ from input")
		}
		sourceInfo, e := os.Stat(source)
		if e != nil {
			return e
		}
		if target, e := os.Stat(*output); e == nil && os.SameFile(sourceInfo, target) {
			return errors.New("output aliases input")
		}
	}
	b, e := dex.Build(p)
	if e != nil {
		return e
	}
	if e = atomicWrite(*output, b); e != nil {
		return e
	}
	_, e = fmt.Fprintf(out, "wrote %s (%d bytes), DEX 035 structurally verified\n", *output, len(b))
	return e
}

// Write the complete validated file before replacing an existing output.
func atomicWrite(path string, b []byte) error {
	tmp, e := os.CreateTemp(filepath.Dir(path), ".koru-dex-*")
	if e != nil {
		return e
	}
	defer os.Remove(tmp.Name())
	if _, e = tmp.Write(b); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Chmod(0644); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Sync(); e != nil {
		tmp.Close()
		return e
	}
	if e = tmp.Close(); e != nil {
		return e
	}
	return os.Rename(tmp.Name(), path)
}
