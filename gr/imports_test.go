package gr_test

import (
	"os/exec"
	"strings"
	"testing"
)

// Plan9.3 gates [8][9]: no new-module imports of phys-gr or frozen phys.
// The gr package must resolve only to stdlib, phys-math, and phys-artifact.
func TestNoLegacyOrFrozenImports(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		switch {
		case dep == "github.com/PithomLabs/phys-gr",
			len(dep) > len("github.com/PithomLabs/phys-gr") &&
				strings.HasPrefix(dep, "github.com/PithomLabs/phys-gr/"):
			t.Fatalf("legacy phys-gr import: %s", dep)
		case dep == "github.com/PithomLabs/phys",
			len(dep) > len("github.com/PithomLabs/phys") &&
				strings.HasPrefix(dep, "github.com/PithomLabs/phys/"):
			t.Fatalf("frozen phys import in Phase 1: %s", dep)
		}
		// Allowed non-stdlib roots.
		first := strings.Split(dep, "/")[0]
		if strings.Contains(first, ".") {
			switch {
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-lib"):
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-math"):
			case strings.HasPrefix(dep, "github.com/PithomLabs/phys-artifact"):
			default:
				t.Fatalf("unexpected third-party dependency: %s", dep)
			}
		}
	}
}
