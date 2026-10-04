package classifications_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

var requiredKeys = []string{"id", "source", "class", "target", "status"}

var requiredSections = []string{
	"Why it is theory-agnostic:",
	"Known/current consumers:",
	"Potential future consumers:",
	"Theory-specific assumptions:",
	"Stable interface:",
	"Dependencies:",
	"Why it does not belong in phys-math:",
	"Why it does not belong in phys-lib/gr:",
}

var validClasses = map[string]bool{
	"PHYS-MATH": true, "PHYS-LIB/CORE": true, "PHYS-LIB/GR": true,
	"SPLIT": true, "REPLACE": true, "DELETE": true, "DEFER": true,
}

// parseRecord splits front-matter (--- delimited) from the body.
func parseRecord(t *testing.T, path string) (map[string]string, string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		t.Fatalf("%s: missing front-matter", path)
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		t.Fatalf("%s: unterminated front-matter", path)
	}
	front := text[4 : 4+end]
	body := text[4+end+4:]
	fields := map[string]string{}
	for _, line := range strings.Split(front, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("%s: malformed front-matter line %q", path, line)
		}
		fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return fields, body
}

func sectionText(body, heading string) string {
	i := strings.Index(body, heading)
	if i < 0 {
		return ""
	}
	rest := body[i+len(heading):]
	// Cut at the next heading (a line ending with ':').
	lines := strings.Split(rest, "\n")
	var out []string
	for _, l := range lines {
		trim := strings.TrimSpace(l)
		if trim != "" && strings.HasSuffix(trim, ":") && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			break
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// TestRecordsComplete enforces D19: every record parses; CORE records carry
// full non-empty evidence. (No per-abstraction records exist yet in Phase 4;
// this test governs Phase 5 additions.)
func TestRecordsComplete(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") || name == "README.md" {
			continue
		}
		count++
		fields, body := parseRecord(t, name)
		for _, k := range requiredKeys {
			if fields[k] == "" {
				t.Fatalf("%s: front-matter key %q missing or empty", name, k)
			}
		}
		if !validClasses[fields["class"]] {
			t.Fatalf("%s: invalid class %q", name, fields["class"])
		}
		if fields["status"] != "adopted" {
			t.Fatalf("%s: status must be adopted", name)
		}
		for _, h := range requiredSections {
			if !strings.Contains(body, h) {
				t.Fatalf("%s: body missing section %q", name, h)
			}
		}
		if fields["class"] == "PHYS-LIB/CORE" {
			for _, h := range requiredSections {
				if sectionText(body, h) == "" {
					t.Fatalf("%s: CORE record section %q is empty", name, h)
				}
			}
		}
	}
	t.Logf("%d classification records checked", count)
}

// TestCoreEmpty asserts package core exports zero identifiers (D20).
func TestCoreEmpty(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir("../core")
	if err != nil {
		t.Fatal(err)
	}
	goFiles := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		goFiles++
		f, err := parser.ParseFile(fset, filepath.Join("../core", name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if f.Name.Name != "core" {
			t.Fatalf("core file %s declares package %q", name, f.Name.Name)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if unicode.IsUpper(rune(d.Name.Name[0])) {
					t.Fatalf("core exports %q", d.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if unicode.IsUpper(rune(s.Name.Name[0])) {
							t.Fatalf("core exports %q", s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, id := range s.Names {
							if unicode.IsUpper(rune(id.Name[0])) {
								t.Fatalf("core exports %q", id.Name)
							}
						}
					}
				}
			}
		}
	}
	if goFiles == 0 {
		t.Fatal("core has no Go files (doc.go required for a compilable empty package)")
	}
}

// TestNoProductionConformanceImports reserves internal/conformance for the
// Phase-8 test fixture: no production file may import it.
func TestNoProductionConformanceImports(t *testing.T) {
	root := ".."
	fset := token.NewFileSet()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "internal" || info.Name() == "classifications" {
				return nil // reserved area / governance tests
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(p, "/internal/") {
				t.Fatalf("production file %s imports %s", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
