// Package adb is a client for the Android Debug Bridge server's smart-socket
// protocol. It talks to an adb server that is already running (normally on
// 127.0.0.1:5037); it never starts processes. Protocol reference:
// https://android.googlesource.com/platform/packages/modules/adb/+/refs/heads/main/docs/dev/services.md
package adb

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultAddress = "127.0.0.1:5037"
const maxReply = 1 << 20

// Address returns the adb server address, honoring ADB_SERVER_SOCKET
// (tcp:host:port) and ANDROID_ADB_SERVER_PORT like the adb tool does.
func Address() (string, error) {
	if s := os.Getenv("ADB_SERVER_SOCKET"); s != "" {
		rest, ok := strings.CutPrefix(s, "tcp:")
		if !ok {
			return "", fmt.Errorf("ADB_SERVER_SOCKET %q: only tcp:host:port is supported", s)
		}
		if !strings.Contains(rest, ":") {
			rest = "127.0.0.1:" + rest
		}
		return rest, nil
	}
	if p := os.Getenv("ANDROID_ADB_SERVER_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return "", fmt.Errorf("ANDROID_ADB_SERVER_PORT %q is not a port", p)
		}
		return "127.0.0.1:" + p, nil
	}
	return DefaultAddress, nil
}

type Client struct {
	Addr    string
	Timeout time.Duration
}

// ErrNoServer means nothing answered at the adb server address.
var ErrNoServer = errors.New("adb server is not running")

func (c Client) dial(ctx context.Context) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return nil, fmt.Errorf("%w at %s (start it with `adb start-server`, or open Android Studio's Device Manager): %v", ErrNoServer, c.Addr, err)
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	conn.SetDeadline(time.Now().Add(timeout))
	return conn, nil
}

func send(conn net.Conn, request string) error {
	if len(request) > 0xffff {
		return errors.New("adb request too long")
	}
	if _, err := fmt.Fprintf(conn, "%04x%s", len(request), request); err != nil {
		return err
	}
	status := make([]byte, 4)
	if _, err := io.ReadFull(conn, status); err != nil {
		return fmt.Errorf("adb %s: %w", request, err)
	}
	switch string(status) {
	case "OKAY":
		return nil
	case "FAIL":
		msg, err := readBlock(conn)
		if err != nil {
			return err
		}
		return fmt.Errorf("adb: %s", msg)
	}
	return fmt.Errorf("adb %s: unexpected reply %q", request, status)
}

func readBlock(r io.Reader) (string, error) {
	size := make([]byte, 4)
	if _, err := io.ReadFull(r, size); err != nil {
		return "", err
	}
	n, err := strconv.ParseUint(string(size), 16, 16)
	if err != nil {
		return "", fmt.Errorf("adb: bad length %q", size)
	}
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	return string(b), err
}

// Version returns the adb server protocol version.
func (c Client) Version(ctx context.Context) (int, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	if err = send(conn, "host:version"); err != nil {
		return 0, err
	}
	s, err := readBlock(conn)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 16, 32)
	return int(v), err
}

type Device struct {
	Serial string
	// State is "device" when usable; also "unauthorized", "offline"...
	State string
	Model string
}

func (c Client) Devices(ctx context.Context) ([]Device, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err = send(conn, "host:devices-l"); err != nil {
		return nil, err
	}
	s, err := readBlock(conn)
	if err != nil {
		return nil, err
	}
	var out []Device
	for _, line := range strings.Split(s, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		d := Device{Serial: f[0], State: f[1]}
		for _, kv := range f[2:] {
			if m, ok := strings.CutPrefix(kv, "model:"); ok {
				d.Model = strings.ReplaceAll(m, "_", " ")
			}
		}
		out = append(out, d)
	}
	return out, nil
}

// Pick chooses the device to use: serial if given, otherwise the only
// connected device. It explains how to fix every other situation.
func Pick(devices []Device, serial string) (Device, error) {
	if serial != "" {
		for _, d := range devices {
			if d.Serial == serial {
				if d.State != "device" {
					return d, stateError(d)
				}
				return d, nil
			}
		}
		return Device{}, fmt.Errorf("no device %q; run `koru devices` to list serials", serial)
	}
	var usable []Device
	for _, d := range devices {
		if d.State == "device" {
			usable = append(usable, d)
		}
	}
	switch {
	case len(usable) == 1:
		return usable[0], nil
	case len(usable) == 0 && len(devices) > 0:
		return devices[0], stateError(devices[0])
	case len(usable) == 0:
		return Device{}, errors.New("no Android device found: connect a phone with USB debugging enabled, or start an emulator")
	}
	var names []string
	for _, d := range usable {
		names = append(names, d.Serial)
	}
	return Device{}, fmt.Errorf("several devices are connected (%s); choose one with --device", strings.Join(names, ", "))
}

func stateError(d Device) error {
	switch d.State {
	case "unauthorized":
		return fmt.Errorf("%s is unauthorized: unlock it and accept the \"Allow USB debugging\" prompt", d.Serial)
	case "offline":
		return fmt.Errorf("%s is offline: reconnect the cable or restart the emulator", d.Serial)
	}
	return fmt.Errorf("%s is %s, not ready", d.Serial, d.State)
}

// Exec runs a command on the device without a terminal, streaming stdin to
// it, and returns its combined output.
func (c Client) Exec(ctx context.Context, serial, command string, stdin io.Reader) (string, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err = send(conn, "host:transport:"+serial); err != nil {
		return "", err
	}
	if err = send(conn, "exec:"+command); err != nil {
		return "", err
	}
	if stdin != nil {
		if _, err = io.Copy(conn, stdin); err != nil {
			return "", fmt.Errorf("streaming to %s: %w", serial, err)
		}
	}
	var out bytes.Buffer
	if _, err = io.Copy(&out, io.LimitReader(conn, maxReply)); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

// Install streams an APK to the package manager (`cmd package install -S`).
func (c Client) Install(ctx context.Context, serial string, apk []byte) error {
	out, err := c.Exec(ctx, serial, fmt.Sprintf("cmd package install -r -S %d", len(apk)), bytes.NewReader(apk))
	if err != nil {
		return err
	}
	out = strings.TrimSpace(out)
	if strings.Contains(out, "Success") {
		return nil
	}
	return installError(out)
}

func installError(out string) error {
	line := out
	if s := bufio.NewScanner(strings.NewReader(out)); s.Scan() {
		line = s.Text()
	}
	switch {
	case strings.Contains(out, "INSTALL_FAILED_UPDATE_INCOMPATIBLE"):
		return fmt.Errorf("install refused: an app with this appId is installed with a different signing key. Uninstall it from the device first (%s)", line)
	case strings.Contains(out, "INSTALL_FAILED_VERSION_DOWNGRADE"):
		return fmt.Errorf("install refused: the device has a newer versionCode; increase versionCode in koru.json (%s)", line)
	case strings.Contains(out, "INSTALL_FAILED_OLDER_SDK"):
		return fmt.Errorf("install refused: the device's Android version is older than minSdk in koru.json (%s)", line)
	case out == "":
		return errors.New("install failed: the device returned no result")
	}
	return fmt.Errorf("install failed: %s", line)
}

// Start launches an Activity component such as com.example.app/.MainActivity.
func (c Client) Start(ctx context.Context, serial, component string) error {
	out, err := c.Exec(ctx, serial, "am start -W -n "+component, nil)
	if err != nil {
		return err
	}
	if strings.Contains(out, "Error") || strings.Contains(out, "Exception") {
		return fmt.Errorf("launch failed: %s", strings.TrimSpace(out))
	}
	return nil
}
