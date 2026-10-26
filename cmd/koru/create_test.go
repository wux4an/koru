package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateBasicBuildsAndVerifies(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-app")
	var out bytes.Buffer
	if e := run([]string{"create", dir}, &out); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out.String(), "koru dev") || !strings.Contains(out.String(), "com.example.myapp") {
		t.Fatalf("create output: %s", out.String())
	}
	for _, name := range []string{"koru.json", "www/index.html", "www/app.js", "types/koru.d.mts", "jsconfig.json", ".gitignore", ".koru/signing.pem"} {
		if _, e := os.Stat(filepath.Join(dir, name)); e != nil {
			t.Fatal(e)
		}
	}
	apk := filepath.Join(t.TempDir(), "app.apk")
	if e := run([]string{"build", dir, "--output", apk}, &out); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"verify", apk}, &out); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"create", dir}, &out); e == nil {
		t.Fatal("create overwrote a project")
	}
}

func TestCreateViteExplainsMissingWebBuild(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "web")
	var out bytes.Buffer
	if e := run([]string{"create", dir, "-t", "vite", "--name", "Web", "--app-id", "com.example.web"}, &out); e != nil {
		t.Fatal(e)
	}
	e := run([]string{"build", dir, "--output", filepath.Join(t.TempDir(), "a.apk")}, &out)
	if e == nil || !strings.Contains(e.Error(), "npm run build") {
		t.Fatalf("build error = %v", e)
	}
	os.MkdirAll(filepath.Join(dir, "dist"), 0755)
	os.WriteFile(filepath.Join(dir, "dist/index.html"), []byte("<!doctype html><head></head>"), 0644)
	if e = run([]string{"build", dir, "--output", filepath.Join(t.TempDir(), "a.apk")}, &out); e != nil {
		t.Fatal(e)
	}
}

func TestCreatePromptsInTerminal(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "space-game")
	var out bytes.Buffer
	in := strings.NewReader(dir + "\n\ncom.example.space\n2\n")
	if e := createProject(createOptions{}, in, &out, true); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "koru.json"))
	for _, want := range []string{`"name": "Space Game"`, `"appId": "com.example.space"`, `"webDir": "dist"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("koru.json lacks %s: %s", want, b)
		}
	}
	if !strings.Contains(out.String(), "npm install") {
		t.Fatal(out.String())
	}
	if e := createProject(createOptions{}, strings.NewReader(""), &out, false); e == nil {
		t.Fatal("non-interactive create without directory accepted")
	}
}

func TestCreateFrameworkTemplatesPrintNodeSteps(t *testing.T) {
	for _, id := range []string{"react", "vue", "svelte", "vanilla-ts"} {
		dir := filepath.Join(t.TempDir(), id)
		var out bytes.Buffer
		if e := run([]string{"create", dir, "-t", id, "--name", "Web", "--app-id", "com.example.web"}, &out); e != nil {
			t.Fatal(id, e)
		}
		if !strings.Contains(out.String(), "npm install") {
			t.Fatalf("%s next steps: %s", id, out.String())
		}
	}
}
