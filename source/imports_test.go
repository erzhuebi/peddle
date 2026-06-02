package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSourceFile(t *testing.T, root string, name string, src string) string {
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

func TestLoadWithImportsSimpleNestedDedupAndRootRelative(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
import "game/player"
import "./shared/constants"

fn main() {
}
`)
	writeSourceFile(t, root, "game/player.ped", `
import "shared/constants"

fn player_init() {
}
`)
	writeSourceFile(t, root, "shared/constants.ped", `
const ANSWER = 42
`)

	src, err := LoadWithImports(mainPath)
	if err != nil {
		t.Fatalf("load imports: %v", err)
	}

	if strings.Count(src, "const ANSWER = 42") != 1 {
		t.Fatalf("expected shared import once, got source:\n%s", src)
	}
	if !strings.Contains(src, "fn player_init()") || !strings.Contains(src, "fn main()") {
		t.Fatalf("expected imported and entry declarations, got source:\n%s", src)
	}
	if strings.Contains(src, "import ") {
		t.Fatalf("expected import lines to be removed, got source:\n%s", src)
	}
}

func TestLoadWithImportsLeadingSlashAndDotDot(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
import "/game/../shared/constants"

fn main() {
}
`)
	writeSourceFile(t, root, "shared/constants.ped", `
const ROOTED = 1
`)

	src, err := LoadWithImports(mainPath)
	if err != nil {
		t.Fatalf("load imports: %v", err)
	}
	if !strings.Contains(src, "const ROOTED = 1") {
		t.Fatalf("expected rooted import, got source:\n%s", src)
	}
}

func TestLoadWithImportsCycleDoesNotLoop(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
import "a"

fn main() {
}
`)
	writeSourceFile(t, root, "a.ped", `
import "b"

fn a() {
}
`)
	writeSourceFile(t, root, "b.ped", `
import "a"

fn b() {
}
`)

	src, err := LoadWithImports(mainPath)
	if err != nil {
		t.Fatalf("load imports: %v", err)
	}
	if strings.Count(src, "fn a()") != 1 || strings.Count(src, "fn b()") != 1 {
		t.Fatalf("expected cyclic imports once each, got source:\n%s", src)
	}
}

func TestLoadWithImportsRejectsOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")

	mainPath := writeSourceFile(t, root, "main.ped", `
import "../outside/secret"

fn main() {
}
`)
	writeSourceFile(t, outside, "secret.ped", `
fn secret() {
}
`)

	_, err := LoadWithImports(mainPath)
	if err == nil {
		t.Fatalf("expected outside-root import error")
	}
	if !strings.Contains(err.Error(), "escapes project root") {
		t.Fatalf("expected outside-root error, got %v", err)
	}
}

func TestLoadWithImportsRejectsSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")

	mainPath := writeSourceFile(t, root, "main.ped", `
import "linked/secret"

fn main() {
}
`)
	writeSourceFile(t, outside, "secret.ped", `
fn secret() {
}
`)

	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	_, err := LoadWithImports(mainPath)
	if err == nil {
		t.Fatalf("expected symlink escape error")
	}
	if !strings.Contains(err.Error(), "escapes project root through symlink") {
		t.Fatalf("expected symlink escape error, got %v", err)
	}
}

func TestLoadWithImportsMissingFileIncludesImporterAndLine(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
import "missing"

fn main() {
}
`)

	_, err := LoadWithImports(mainPath)
	if err == nil {
		t.Fatalf("expected missing import error")
	}
	if !strings.Contains(err.Error(), "main.ped:2") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected importer line and missing-file message, got %v", err)
	}
}

func TestLoadWithImportsRejectsBadImportSyntax(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
import game/player

fn main() {
}
`)

	_, err := LoadWithImports(mainPath)
	if err == nil {
		t.Fatalf("expected bad import syntax error")
	}
	if !strings.Contains(err.Error(), "main.ped:2") || !strings.Contains(err.Error(), "expected import") {
		t.Fatalf("expected syntax error with importer line, got %v", err)
	}
}

func TestLoadWithImportsRejectsImportInsideFunction(t *testing.T) {
	root := t.TempDir()

	mainPath := writeSourceFile(t, root, "main.ped", `
fn main() {
    import "helper"
}
`)

	_, err := LoadWithImports(mainPath)
	if err == nil {
		t.Fatalf("expected nested import error")
	}
	if !strings.Contains(err.Error(), "main.ped:3") || !strings.Contains(err.Error(), "top level") {
		t.Fatalf("expected top-level import error, got %v", err)
	}
}
