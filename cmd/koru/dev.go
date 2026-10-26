package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/devserver"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func newDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev [project]",
		Short: "Preview the app in a browser with live reload",
		Long: `Serve the project to a browser with live reload. The koru JavaScript API runs on
web implementations, so the same code works before you build an APK.

With --proxy, koru forwards pages to your framework's dev server (Vite, Astro,
SvelteKit...) and adds the koru API to them, keeping that server's hot reload.`,
		Example: `  koru dev
  koru dev ./my-app --port 3000
  koru dev --host 0.0.0.0              # also reachable from phones on your network
  koru dev --proxy http://localhost:5173`,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().Int("port", 8100, "port to listen on")
	cmd.Flags().String("host", "127.0.0.1", "address to listen on; 0.0.0.0 exposes it to your network")
	cmd.Flags().String("proxy", "", "forward to this web dev server instead of serving webDir")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		proxy, _ := cmd.Flags().GetString("proxy")
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		return serveDev(ctx, devOptions{project: project, host: host, port: port, proxy: proxy}, cmd.OutOrStdout(), nil)
	}
	return cmd
}

type devOptions struct {
	project, host, proxy string
	port                 int
}

func serveDev(ctx context.Context, o devOptions, out io.Writer, listening func(addr string)) error {
	c, err := config.Load(o.project)
	if err != nil {
		return err
	}
	root := c.WebDir
	if !filepath.IsAbs(root) {
		root = filepath.Join(o.project, root)
	}
	opt := devserver.Options{Root: root, SPA: c.SPA, App: func() (devserver.AppInfo, error) {
		c, err := config.Load(o.project)
		return devserver.AppInfo{AppID: c.AppID, Name: c.Name, Version: c.Version, VersionCode: c.VersionCode, Plugins: c.Plugins}, err
	}}
	if o.proxy != "" {
		if opt.Proxy, err = url.Parse(o.proxy); err != nil {
			return fmt.Errorf("--proxy: %w", err)
		}
	}
	s, err := devserver.New(opt)
	if err != nil {
		return err
	}
	if o.port < 0 || o.port > 65535 {
		return errors.New("--port must be from 0 to 65535")
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(o.host, strconv.Itoa(o.port)))
	if err != nil {
		return fmt.Errorf("cannot listen on %s:%d (%w); pick another port with --port", o.host, o.port, err)
	}
	srv := &http.Server{Handler: s, ReadHeaderTimeout: 10 * time.Second}
	addr := ln.Addr().(*net.TCPAddr)
	if err = printDevURLs(out, o.host, addr.Port, o.proxy); err != nil {
		ln.Close()
		return err
	}
	if listening != nil {
		listening(ln.Addr().String())
	}
	if o.proxy == "" {
		go devserver.Watch(ctx, []string{o.project}, 300*time.Millisecond, func() {
			s.Reload()
			writeStatus(out, "Reloaded", "files changed", "iris")
		}, func(err error) {
			writeStatus(out, "Watch", err.Error(), "peach")
		})
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()
	select {
	case err = <-served:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
		return nil
	}
}

func printDevURLs(out io.Writer, host string, port int, proxy string) error {
	p := presentationFor(out)
	what := "live reload on save"
	if proxy != "" {
		what = "proxying " + proxy
	}
	if err := writeStatus(out, "koru dev", what+"; press Ctrl+C to stop", "mint"); err != nil {
		return err
	}
	local := host
	if host == "0.0.0.0" || host == "::" || host == "" {
		local = "127.0.0.1"
	}
	fmt.Fprintf(out, "  %s  http://%s\n", p.paint("Local:  ", "iris"), net.JoinHostPort(local, strconv.Itoa(port)))
	if local != host {
		for _, ip := range lanAddresses() {
			fmt.Fprintf(out, "  %s  http://%s\n", p.paint("Network:", "iris"), net.JoinHostPort(ip, strconv.Itoa(port)))
		}
	}
	_, err := fmt.Fprintln(out, "  The koru API uses web implementations here; run `koru run` for the native app.")
	return err
}

func lanAddresses() []string {
	var out []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && !n.IP.IsLoopback() {
			out = append(out, n.IP.String())
		}
	}
	return out
}
