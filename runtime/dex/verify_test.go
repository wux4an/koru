package runtimedex

import (
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
	return map[string]config.Config{"default": config.Default(), "all plugins": all, "network": network, "no plugins": none}
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
