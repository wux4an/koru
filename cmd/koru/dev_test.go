package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}
func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func TestDevServesCreatedProject(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo")
	var created bytes.Buffer
	if e := run([]string{"create", dir}, &created); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	addr := make(chan string, 1)
	out := &lockedBuffer{}
	done := make(chan error, 1)
	go func() {
		done <- serveDev(ctx, devOptions{project: dir, host: "127.0.0.1", port: 0}, out, func(a string) { addr <- a })
	}()
	var base string
	select {
	case a := <-addr:
		base = "http://" + a
	case e := <-done:
		t.Fatal(e)
	}
	res, err := http.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if !strings.Contains(string(body), "/__koru/dev.js") || !strings.Contains(string(body), "<title>Demo</title>") {
		t.Fatal(string(body))
	}
	res, _ = http.Get(base + "/__koru/dev.js")
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if !strings.Contains(string(body), `"appId":"com.example.demo"`) {
		t.Fatal(string(body))
	}
	os.WriteFile(filepath.Join(dir, "www/style.css"), []byte("body{}"), 0644)
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(out.String(), "Reloaded") && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !strings.Contains(out.String(), "Reloaded") || !strings.Contains(out.String(), "Local:") {
		t.Fatal(out.String())
	}
	cancel()
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}

func TestDevRejectsBadInput(t *testing.T) {
	if e := serveDev(context.Background(), devOptions{project: t.TempDir()}, io.Discard, nil); e == nil || !strings.Contains(e.Error(), "koru create") {
		t.Fatalf("missing project error = %v", e)
	}
}
