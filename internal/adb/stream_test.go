package adb

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"koru.dev/koru/internal/adb/adbtest"
)

type syncBuf struct {
	mu sync.Mutex
	sb strings.Builder
}

func (b *syncBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.Write(p)
}
func (b *syncBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

func TestStreamFollowsUntilCancelled(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	fake.SetRunning("com.example.app", 4242)
	// Larger than Exec's 1 MiB cap to prove Stream is unbounded.
	line := "I/App( 4242): " + strings.Repeat("x", 1000) + "\n"
	big := strings.Repeat(line, 1500)
	fake.SetLogcat(big, "")
	c := Client{Addr: fake.Addr()}
	ctx, cancel := context.WithCancel(context.Background())
	var buf syncBuf
	done := make(chan error, 1)
	go func() { done <- c.Stream(ctx, adbtest.Serial, "logcat --pid=4242 -v brief", &buf) }()
	deadline := time.Now().Add(5 * time.Second)
	for len(buf.String()) < len(big) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := len(buf.String()); got != len(big) {
		t.Fatalf("streamed %d of %d bytes", got, len(big))
	}
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stream did not stop on cancel")
	}
}

func TestStreamErrors(t *testing.T) {
	fake, err := adbtest.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fake.Close()
	c := Client{Addr: fake.Addr()}
	var buf syncBuf
	if err = c.Stream(context.Background(), "nope", "logcat", &buf); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unknown device: %v", err)
	}
	if err = c.Stream(context.Background(), adbtest.Serial, "bogus", &buf); err == nil || !strings.Contains(err.Error(), "unknown service") {
		t.Fatalf("unknown command: %v", err)
	}
	if err = (Client{Addr: "127.0.0.1:1"}).Stream(context.Background(), adbtest.Serial, "logcat", &buf); err == nil || !strings.Contains(err.Error(), "adb start-server") {
		t.Fatalf("no server: %v", err)
	}
}
