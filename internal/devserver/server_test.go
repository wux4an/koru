package devserver

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func info() (AppInfo, error) {
	return AppInfo{AppID: "com.example.dev", Name: "Dev <App>", Version: "1.0.0", VersionCode: 3, Plugins: []string{"device"}}, nil
}

func site(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "index.html"), []byte("<html><HEAD><title>x</title></HEAD><body>home</body></html>"), 0644)
	os.MkdirAll(filepath.Join(root, "docs"), 0755)
	os.WriteFile(filepath.Join(root, "docs/index.html"), []byte("<p>docs</p>"), 0644)
	os.WriteFile(filepath.Join(root, "app.js"), []byte("console.log(1)"), 0644)
	os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=1"), 0644)
	return root
}

func get(t *testing.T, h http.Handler, target string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec.Code, rec.Body.String()
}

func TestStaticServingInjectsBridge(t *testing.T) {
	root := site(t)
	s, err := New(Options{Root: root, App: info})
	if err != nil {
		t.Fatal(err)
	}
	code, body := get(t, s, "/")
	if code != 200 || !strings.Contains(body, devTags+"</HEAD>") {
		t.Fatalf("%d %s", code, body)
	}
	if _, body = get(t, s, "/docs/"); !strings.HasPrefix(body, devTags+"<p>docs</p>") {
		t.Fatal(body)
	}
	if _, body = get(t, s, "/app.js"); body != "console.log(1)" {
		t.Fatal(body)
	}
	for _, hidden := range []string{"/.env", "/missing", "/../index.html/..%2f.env", "/__koru/other.js"} {
		if code, _ = get(t, s, hidden); code != 404 {
			t.Fatalf("%s: %d", hidden, code)
		}
	}
	if _, body = get(t, s, "/__koru/core.js"); !strings.Contains(body, "koru v1 protocol") {
		t.Fatal("core.js not served")
	}
	if _, body = get(t, s, "/__koru/core.mjs"); !strings.Contains(body, "export const device") {
		t.Fatal("core.mjs not served")
	}
	_, body = get(t, s, "/__koru/dev.js")
	if !strings.Contains(body, `"appId":"com.example.dev"`) || !strings.Contains(body, `"name":"Dev `+"\\"+"u003cApp") || !strings.Contains(body, "EventSource") {
		t.Fatal(body)
	}
	os.Symlink(filepath.Join(root, ".env"), filepath.Join(root, "leak.txt"))
	if code, _ = get(t, s, "/leak.txt"); code != 404 {
		t.Fatal("symlink served")
	}
}

func TestSPAFallbackAndMissingIndex(t *testing.T) {
	root := site(t)
	s, _ := New(Options{Root: root, SPA: true, App: info})
	if code, body := get(t, s, "/settings/profile"); code != 200 || !strings.Contains(body, "home") {
		t.Fatalf("%d %s", code, body)
	}
	if code, _ := get(t, s, "/missing.png"); code != 404 {
		t.Fatal("asset fallback")
	}
	if _, err := New(Options{Root: t.TempDir(), App: info}); err == nil || !strings.Contains(err.Error(), "--proxy") {
		t.Fatalf("missing index error = %v", err)
	}
}

func TestConfigErrorsReachTheConsole(t *testing.T) {
	s, _ := New(Options{Root: site(t), App: func() (AppInfo, error) { return AppInfo{}, os.ErrNotExist }})
	if _, body := get(t, s, "/__koru/dev.js"); !strings.Contains(body, "console.error") {
		t.Fatal(body)
	}
}

func TestReloadEvent(t *testing.T) {
	s, _ := New(Options{Root: site(t), App: info})
	srv := httptest.NewServer(s)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/__koru/events", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	r := bufio.NewReader(res.Body)
	if line, _ := r.ReadString('\n'); line != ": koru dev\n" {
		t.Fatalf("%q", line)
	}
	r.ReadString('\n')
	s.Reload()
	if line, _ := r.ReadString('\n'); line != "event: reload\n" {
		t.Fatalf("%q", line)
	}
}

func TestProxyInjectsIntoHTMLOnly(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/main.js" {
			w.Header().Set("Content-Type", "text/javascript")
			io.WriteString(w, "</head>")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<head></head>vite")
	}))
	defer upstream.Close()
	u, _ := url.Parse(upstream.URL)
	s, err := New(Options{Proxy: u, App: info})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(s)
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	req.Header.Set("Accept-Encoding", "br")
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if string(b) != "<head>"+devTags+"</head>vite" {
		t.Fatal(string(b))
	}
	res, _ = http.Get(srv.URL + "/main.js")
	b, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if string(b) != "</head>" {
		t.Fatal(string(b))
	}
	if _, err = New(Options{Proxy: &url.URL{Scheme: "ftp", Host: "x"}, App: info}); err == nil {
		t.Fatal("ftp proxy accepted")
	}
	down, _ := url.Parse("http://127.0.0.1:1")
	s, _ = New(Options{Proxy: down, App: info})
	if code, body := get(t, s, "/"); code != http.StatusBadGateway || !strings.Contains(body, "Start your web dev server") {
		t.Fatalf("%d %s", code, body)
	}
}

func TestWatchReportsChanges(t *testing.T) {
	root := site(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go Watch(ctx, []string{root}, 10*time.Millisecond, func() { changed <- struct{}{} }, func(error) {})
	time.Sleep(40 * time.Millisecond)
	os.WriteFile(filepath.Join(root, "app.js"), []byte("console.log(22)"), 0644)
	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("no change detected")
	}
	// A save can be seen half-written; let it settle before the next check.
	time.Sleep(60 * time.Millisecond)
	for len(changed) > 0 {
		<-changed
	}
	os.MkdirAll(filepath.Join(root, "node_modules"), 0755)
	os.WriteFile(filepath.Join(root, "node_modules/x.js"), []byte("x"), 0644)
	select {
	case <-changed:
		t.Fatal("node_modules change triggered reload")
	case <-time.After(100 * time.Millisecond):
	}
}
