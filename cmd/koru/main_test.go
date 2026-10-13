package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildInspectCLI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "answer.dex")
	var stdout bytes.Buffer
	if e := run([]string{"dex", "build", "../../examples/dex/answer.json", "--output", out}, &stdout); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(stdout.String(), "verified") {
		t.Fatal(stdout.String())
	}
	stdout.Reset()
	if e := run([]string{"dex", "inspect", out}, &stdout); e != nil {
		t.Fatal(e)
	}
	if !json.Valid(stdout.Bytes()) || strings.Contains(stdout.String(), "\x1b[") {
		t.Fatal("DEX inspection output must remain plain JSON")
	}
	if !strings.Contains(stdout.String(), "Lexample/Answer;") {
		t.Fatal(stdout.String())
	}
	b, e := os.ReadFile(out)
	if e != nil || string(b[:8]) != "dex\n035\x00" {
		t.Fatal(e)
	}
}
func TestActivityCLI(t *testing.T) {
	var out bytes.Buffer
	if e := run([]string{"dex", "build", "--activity-experiment", "--output", filepath.Join(t.TempDir(), "activity.dex")}, &out); e != nil {
		t.Fatal(e)
	}
}
func TestErrorsPreserveOutput(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.dex")
	input := filepath.Join(dir, "bad.json")
	os.WriteFile(output, []byte("keep"), 0644)
	os.WriteFile(input, []byte(`{"classes":[],"unknown":true}`), 0644)
	if e := run([]string{"dex", "build", input, "--output", output}, &bytes.Buffer{}); e == nil {
		t.Fatal("invalid JSON accepted")
	}
	b, _ := os.ReadFile(output)
	if string(b) != "keep" {
		t.Fatal("output overwritten")
	}
}
func TestUnsupportedCommandsAndArguments(t *testing.T) {
	for _, args := range [][]string{{"build", "one", "two"}, {"verify", "x.apk"}, {"dex", "inspect"}, {"dex", "build"}, {"dex", "build", "--output"}, {"dex", "build", "--activity-experiment", "input.json"}, {"unknown"}, {"version", "extra"}, {"doctor", "extra"}} {
		if e := run(args, &bytes.Buffer{}); e == nil {
			t.Fatal(args)
		}
	}
}

func TestCobraHelpIsEnglishAndListsCommands(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "Commands:", "Build and sign an Android APK", "Check a project's configuration", "version"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help is missing %q:\n%s", want, help)
		}
	}
	for _, unwanted := range []string{"Uso:", "Comandos:", "inicializado", "Aplicación"} {
		if strings.Contains(help, unwanted) {
			t.Fatalf("help contains non-English text %q:\n%s", unwanted, help)
		}
	}
}
func TestDoctorVersion(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"doctor"}} {
		var b bytes.Buffer
		if e := run(args, &b); e != nil || b.Len() == 0 {
			t.Fatal(e)
		}
	}
}
func TestInitBuildReleaseAndVerify(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if e := run([]string{"init", dir}, &out); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(dir, "out.apk")
	if e := run([]string{"build", dir, "--output", path, "--release"}, &out); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"verify", path}, &out); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"init", dir}, &out); e == nil {
		t.Fatal("init overwrites existing project")
	}
}
