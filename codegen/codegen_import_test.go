package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"peddle/source"
)

func writeImportTestSource(t *testing.T, root string, name string, src string) string {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestCodegenProgramLoadedFromImports(t *testing.T) {
	root := t.TempDir()

	mainPath := writeImportTestSource(t, root, "main.ped", `
import "/lib/math"
import "lib/output"

fn main() {
    var x int

    x = answer()
    show(x)
}
`)
	writeImportTestSource(t, root, "lib/math.ped", `
fn answer() int {
    return 42
}
`)
	writeImportTestSource(t, root, "lib/output.ped", `
fn show(x int) {
    if x == 42 {
        print("OK")
    }
}
`)

	src, err := source.LoadWithImports(mainPath)
	if err != nil {
		t.Fatalf("load imports: %v", err)
	}

	asm := compileSource(t, src)
	requireASM(t, asm,
		"jsr answer",
		"jsr show",
		"answer_return:",
	)
	requireASMAssemblesWith64tass(t, asm)
}
