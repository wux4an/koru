// Package devserver serves a koru project to desktop and phone browsers while
// it is being developed: web files, the koru JavaScript module running on web
// implementations, and live reload. It never takes part in an APK build.
package devserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/app"
	"koru.dev/koru/packages/javascript"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxProxiedHTML = 16 << 20

// AppInfo is what app.getInfo() and the plugin checks see in the browser.
type AppInfo struct {
	AppID       string   `json:"appId"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	VersionCode uint32   `json:"versionCode"`
	Plugins     []string `json:"plugins"`
}

type Options struct {
	// Root is the web directory served in static mode.
	Root string
	// Proxy forwards everything except /__koru/ to another dev server, such
	// as Vite, instead of serving Root.
	Proxy *url.URL
	SPA   bool
	// App is called for every page load so edits to koru.json apply on reload.
	App func() (AppInfo, error)
}

type Server struct {
	opt     Options
	proxy   *httputil.ReverseProxy
	mu      sync.Mutex
	clients map[chan struct{}]bool
}

func New(opt Options) (*Server, error) {
	if opt.App == nil {
		return nil, errors.New("devserver: App is required")
	}
	s := &Server{opt: opt, clients: map[chan struct{}]bool{}}
	if opt.Proxy != nil {
		if opt.Proxy.Scheme != "http" && opt.Proxy.Scheme != "https" || opt.Proxy.Host == "" {
			return nil, fmt.Errorf("--proxy must be an http(s) URL such as http://localhost:5173, not %q", opt.Proxy.String())
		}
		s.proxy = &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(opt.Proxy)
				r.Out.Host = r.In.Host
				// Drop the browser's encodings so the transport negotiates gzip
				// itself and hands back a decoded body that can be rewritten.
				r.Out.Header.Del("Accept-Encoding")
			},
			ModifyResponse: injectProxied,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, fmt.Sprintf("koru dev: cannot reach %s (%v). Start your web dev server first.", opt.Proxy, err), http.StatusBadGateway)
			},
		}
	} else {
		info, err := os.Stat(filepath.Join(opt.Root, "index.html"))
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%s has no index.html; set webDir in koru.json or use --proxy with your web dev server", opt.Root)
		}
	}
	return s, nil
}

const devTags = `<script src="/__koru/dev.js"></script><script src="/__koru/core.js"></script>`

func injectProxied(r *http.Response) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/html") || r.Header.Get("Content-Encoding") != "" {
		return nil
	}
	b, err := io.ReadAll(io.LimitReader(r.Body, maxProxiedHTML+1))
	r.Body.Close()
	if err != nil {
		return err
	}
	if len(b) > maxProxiedHTML {
		return errors.New("proxied HTML page is too large")
	}
	b = []byte(app.InjectHead(string(b), devTags))
	r.Body = io.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(len(b))
	r.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/__koru/core.js":
		serveBytes(w, "text/javascript; charset=utf-8", javascript.Core)
		return
	case "/__koru/core.mjs":
		serveBytes(w, "text/javascript; charset=utf-8", javascript.Module)
		return
	case "/__koru/dev.js":
		s.serveDevScript(w)
		return
	case "/__koru/events":
		s.serveEvents(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/__koru/") {
		http.NotFound(w, r)
		return
	}
	if s.proxy != nil {
		s.proxy.ServeHTTP(w, r)
		return
	}
	s.serveStatic(w, r)
}

func serveBytes(w http.ResponseWriter, contentType string, b []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Write(b)
}

func (s *Server) serveDevScript(w http.ResponseWriter) {
	info, err := s.opt.App()
	if err != nil {
		// Show configuration mistakes in the page's console instead of a blank page.
		msg, _ := json.Marshal("koru dev: " + err.Error())
		serveBytes(w, "text/javascript; charset=utf-8", []byte("console.error("+string(msg)+");\n"+reloadScript))
		return
	}
	b, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveBytes(w, "text/javascript; charset=utf-8", []byte("window.__KORU_APP__ = Object.freeze("+string(b)+");\n"+reloadScript))
}

const reloadScript = `(() => {
  if (typeof EventSource === "undefined") return;
  const events = new EventSource("/__koru/events");
  events.addEventListener("reload", () => location.reload());
})();
`

// Reload asks every connected page to reload.
func (s *Server) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.clients {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}

func (s *Server) serveEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	c := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients[c] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, c)
		s.mu.Unlock()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprint(w, ": koru dev\n\n")
	flusher.Flush()
	keepAlive := time.NewTicker(25 * time.Second)
	defer keepAlive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-c:
			fmt.Fprint(w, "event: reload\ndata: {}\n\n")
		case <-keepAlive.C:
			fmt.Fprint(w, ": ping\n\n")
		}
		flusher.Flush()
	}
}

// resolve maps a URL path to a regular file under Root, refusing hidden
// segments and symbolic links like the APK asset collector does.
func (s *Server) resolve(urlPath string) (string, bool) {
	clean := path.Clean("/" + urlPath)
	for _, seg := range strings.Split(clean, "/") {
		if strings.HasPrefix(seg, ".") || seg == "node_modules" {
			return "", false
		}
	}
	name := filepath.Join(s.opt.Root, filepath.FromSlash(clean))
	info, err := os.Lstat(name)
	if err == nil && info.IsDir() {
		name = filepath.Join(name, "index.html")
		info, err = os.Lstat(name)
	}
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return name, true
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name, ok := s.resolve(r.URL.Path)
	if !ok && s.opt.SPA && path.Ext(r.URL.Path) == "" {
		name, ok = s.resolve("/index.html")
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if strings.EqualFold(filepath.Ext(name), ".html") {
		b, err := os.ReadFile(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, app.InjectHead(string(b), devTags))
		return
	}
	http.ServeFile(w, r, name)
}
