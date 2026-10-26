package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"koru.dev/koru/internal/adb"
	"koru.dev/koru/internal/config"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [project]",
		Short: "Stream the app's logs from a device or emulator",
		Long: `Stream the logs of the running app: its logcat output plus the "chromium" lines
that carry console.log, console.warn and console.error from the WebView.

The app must be open (see ` + "`koru run`" + `). koru talks to the adb server through its
network protocol; it does not run logcat locally. Press Ctrl+C to stop.`,
		Example: `  koru logs
  koru logs ./my-app --device emulator-5554`,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("device", "", "device serial from `koru devices` (default: the only connected device, or $ANDROID_SERIAL)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		project := "."
		if len(args) == 1 {
			project = args[0]
		}
		serial, _ := cmd.Flags().GetString("device")
		if serial == "" {
			serial = os.Getenv("ANDROID_SERIAL")
		}
		addr, err := adb.Address()
		if err != nil {
			return err
		}
		c, err := config.Load(project)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		client := adb.Client{Addr: addr}
		devices, err := client.Devices(ctx)
		if err != nil {
			return err
		}
		device, err := adb.Pick(devices, serial)
		if err != nil {
			return err
		}
		return followLogs(ctx, client, device.Serial, c.AppID, 1, cmd.OutOrStdout())
	}
	return cmd
}

var validAppID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)

// briefChromium matches the start of a logcat "brief" line from the WebView:
// `I/chromium( 1234): ...`. Group 1 is the pid.
var briefChromium = regexp.MustCompile(`^[VDIWEFA]/chromium\s*\(\s*(\d+)\)`)

// appPid returns the pid of appID using `pidof`, trying up to attempts times
// 300 ms apart (right after `am start` the process can still be registering).
func appPid(ctx context.Context, client adb.Client, serial, appID string, attempts int) (int, error) {
	if !validAppID.MatchString(appID) {
		return 0, fmt.Errorf("appId %q is not a valid package name", appID)
	}
	for i := 0; ; i++ {
		out, err := client.Exec(ctx, serial, "pidof "+appID, nil)
		if err != nil {
			return 0, err
		}
		// pidof may print several pids; the first is the oldest process.
		if f := strings.Fields(out); len(f) > 0 {
			if pid, perr := strconv.Atoi(f[0]); perr == nil && pid > 0 {
				return pid, nil
			}
			return 0, fmt.Errorf("unexpected pidof output %q", strings.TrimSpace(out))
		}
		if i+1 >= attempts {
			return 0, fmt.Errorf("%s is not running on %s: open it with `koru run` first", appID, serial)
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// followLogs prints the app's logs until ctx is cancelled (returns nil) or the
// app's process ends (returns an error). It starts two logcat streams:
//
//	logcat --pid=<pid> -v brief       everything the app process logs
//	logcat -v brief -s chromium:V     WebView console messages
//
// Console messages are logged by the app's own process, so the second stream
// repeats them; lines whose pid is the app's are dropped from it. It only adds
// "chromium" lines from other pids (for example a separate WebView process).
//
// NOTE: the exact logcat output format has not been validated on a real
// device; tests use a simulated server. The brief format and the pidof/logcat
// flags follow the Android documentation (Android 7+ toolbox pidof, logcat
// --pid).
func followLogs(ctx context.Context, client adb.Client, serial, appID string, pidAttempts int, out io.Writer) error {
	pid, err := appPid(ctx, client, serial, appID, pidAttempts)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	if err = writeStatus(out, "Logs", fmt.Sprintf("%s (pid %d) on %s; press Ctrl+C to stop", appID, pid, serial), "mint"); err != nil {
		return err
	}
	shared := &syncWriter{w: out}
	parent := ctx
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	streams := []struct {
		cmd  string
		keep func(line []byte) bool
	}{
		{fmt.Sprintf("logcat --pid=%d -v brief", pid), nil},
		{"logcat -v brief -s chromium:V", func(line []byte) bool {
			m := briefChromium.FindSubmatch(line)
			return m != nil && string(m[1]) != strconv.Itoa(pid)
		}},
	}
	errs := make(chan error, len(streams))
	for _, s := range streams {
		go func() {
			lw := &lineWriter{dst: shared, keep: s.keep}
			err := client.Stream(ctx, serial, s.cmd, lw)
			if ferr := lw.flush(); err == nil {
				err = ferr
			}
			errs <- err
		}()
	}
	// The first stream to stop decides the outcome; the other one is cancelled.
	first := <-errs
	cancel()
	<-errs
	if parent.Err() != nil {
		return nil // Ctrl+C is the normal way out
	}
	if first != nil {
		return first
	}
	return fmt.Errorf("log stream from %s ended: the app may have closed or the device disconnected", serial)
}

type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

// lineWriter forwards whole lines that pass keep (all lines when keep is nil),
// so two concurrent streams never interleave inside a line.
type lineWriter struct {
	dst  io.Writer
	keep func([]byte) bool
	buf  []byte
}

func (l *lineWriter) Write(p []byte) (int, error) {
	l.buf = append(l.buf, p...)
	for {
		i := bytes.IndexByte(l.buf, '\n')
		if i < 0 {
			return len(p), nil
		}
		line := bytes.TrimRight(l.buf[:i], "\r")
		if err := l.emit(line); err != nil {
			return 0, err
		}
		l.buf = l.buf[i+1:]
	}
}

func (l *lineWriter) flush() error {
	if len(l.buf) == 0 {
		return nil
	}
	line := l.buf
	l.buf = nil
	return l.emit(bytes.TrimRight(line, "\r"))
}

func (l *lineWriter) emit(line []byte) error {
	if l.keep != nil && !l.keep(line) {
		return nil
	}
	_, err := l.dst.Write(append(append([]byte(nil), line...), '\n'))
	return err
}
