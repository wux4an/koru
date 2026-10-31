package runtimedex

import (
	"crypto/sha256"
	"fmt"
	"github.com/wux4an/koru/pkg/dex"
	"path/filepath"
	"testing"

	"github.com/wux4an/koru/internal/config"
	"github.com/wux4an/koru/internal/dex/verify"
	"github.com/wux4an/koru/plugins"
)

// Every runtime variant must pass the type checker: a method ART rejects
// crashes the app when its class loads.
func TestRuntimeVariantsVerify(t *testing.T) {
	toast, err := plugins.Load(filepath.Join("..", "..", "examples", "toast", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	assets := []Asset{{Path: "/index.html", MIME: "text/html"}, {Path: "/__koru/core.js", MIME: "text/javascript"}}
	for name, c := range runtimeVariants() {
		t.Run(name, func(t *testing.T) {
			if err := verify.Program(Runtime(c, assets)); err != nil {
				t.Fatal(err)
			}
			c.Plugins = append(c.Plugins, "toast")
			if err := verify.Program(Runtime(c, assets, toast)); err != nil {
				t.Fatal(err)
			}
		})
	}
	if err := verify.Program(ActivityExperiment()); err != nil {
		t.Fatal(err)
	}
}

func runtimeVariants() map[string]config.Config {
	all := config.Default()
	all.Plugins = []string{"device", "storage", "vibration", "app", "share", "notifications"}
	all.SPA = true
	network := all
	network.Permissions = []string{"android.permission.INTERNET"}
	none := config.Default()
	none.Plugins = nil
	remote := all
	remote.Server = &config.Server{URL: "https://example.com/app"}
	hybrid := all
	hybrid.Server = &config.Server{AllowNavigation: []string{"example.com", "*.example.org"}}
	bridged := all
	bridged.Server = &config.Server{URL: "https://example.com/", AllowNavigation: []string{"*.example.org"}, Bridge: []string{"https://example.com", "https://app.example.org"}}
	return map[string]config.Config{"default": config.Default(), "all plugins": all, "network": network, "no plugins": none,
		"server url": remote, "hybrid": hybrid, "bridge": bridged}
}

// Mutating the real runtime must be caught: swap every move-result kind, one
// at a time, and expect the verifier to reject each mutant.
func TestVerifierCatchesRuntimeMutations(t *testing.T) {
	c := runtimeVariants()["all plugins"]
	assets := []Asset{{Path: "/index.html", MIME: "text/html"}}
	p := Runtime(c, assets)
	swaps := map[string]string{"move-result": "move-result-object", "move-result-object": "move-result"}
	mutants, caught := 0, 0
	for ci := range p.Classes {
		for mi := range p.Classes[ci].Methods {
			code := p.Classes[ci].Methods[mi].Code
			if code == nil {
				continue
			}
			for ii := range code.Instructions {
				orig := code.Instructions[ii].Op
				swap, ok := swaps[orig]
				if !ok {
					continue
				}
				mutants++
				code.Instructions[ii].Op = swap
				if verify.Program(p) != nil {
					caught++
				}
				code.Instructions[ii].Op = orig
			}
		}
	}
	if mutants < 50 || caught != mutants {
		t.Fatalf("caught %d of %d move-result mutants", caught, mutants)
	}
}

// New runtime features are opt-in: without them the bytecode must stay
// byte-identical to the runtime validated on Android 11. Update these hashes
// only for a deliberate change to the default runtime, and validate it on a
// device first.
func TestDefaultRuntimeUnchanged(t *testing.T) {
	all := config.Default()
	all.Plugins = []string{"device", "storage", "vibration", "app", "share", "notifications"}
	network := all
	network.Permissions = []string{"android.permission.INTERNET"}
	want := map[string]config.Config{
		"af03b9f05626ca826a5b421cdc88cb24e02e6a5c1dcdc45c4b193b258dae2007": config.Default(),
		"8d0164ed947e686099741f8276ba00600e823dd827d72f66a4bfa5186f472e8f": all,
		"ebd01cd08a8c4a1f13ccfa7e6cbdbd871af4f21f2ae91020dcebddf19705da33": network,
	}
	for sum, c := range want {
		b, err := dex.Build(Runtime(c, []Asset{{Path: "/index.html", MIME: "text/html"}}))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != sum {
			t.Errorf("default runtime changed: %s, want %s", got, sum)
		}
	}
}
