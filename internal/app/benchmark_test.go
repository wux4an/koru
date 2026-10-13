package app

import (
	"encoding/json"
	"koru.dev/koru/internal/config"
	"koru.dev/koru/internal/signing"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkAPKBuildCached(b *testing.B) {
	p := b.TempDir()
	os.MkdirAll(filepath.Join(p, "www"), 0700)
	os.MkdirAll(filepath.Join(p, ".koru"), 0700)
	c := config.Default()
	data, _ := json.Marshal(c)
	os.WriteFile(filepath.Join(p, "koru.json"), data, 0600)
	os.WriteFile(filepath.Join(p, "www/index.html"), []byte("<head></head><body>benchmark</body>"), 0600)
	if _, e := signing.CreateIdentity(filepath.Join(p, c.SigningKey)); e != nil {
		b.Fatal(e)
	}
	if _, e := Build(p, Options{}); e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := Build(p, Options{}); e != nil {
			b.Fatal(e)
		}
	}
}
