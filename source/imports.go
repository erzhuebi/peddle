package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LoadWithImports expands Peddle import declarations before lexing/parsing.
// Import paths are always resolved from the directory of the entry file.
func LoadWithImports(entryPath string) (string, error) {
	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return "", fmt.Errorf("resolve entry path: %w", err)
	}

	root := filepath.Clean(filepath.Dir(entryAbs))
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve project root %s: %w", root, err)
	}

	l := &loader{
		root:     root,
		rootReal: filepath.Clean(rootReal),
		seen:     map[string]bool{},
	}

	return l.loadFile(entryAbs)
}

type loader struct {
	root     string
	rootReal string
	seen     map[string]bool
}

func (l *loader) loadFile(path string) (string, error) {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	realPath = filepath.Clean(realPath)

	if !withinDir(l.rootReal, realPath) {
		return "", fmt.Errorf("entry file %s is outside project root %s", path, l.root)
	}

	if l.seen[realPath] {
		return "", nil
	}
	l.seen[realPath] = true

	data, err := os.ReadFile(realPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", realPath, err)
	}

	return l.expandFile(realPath, string(data))
}

func (l *loader) expandFile(filePath string, src string) (string, error) {
	lines := strings.SplitAfter(src, "\n")
	depth := 0

	var imports strings.Builder
	var body strings.Builder

	for i, line := range lines {
		lineNo := i + 1
		spec, isImport, err := parseImportLine(line)
		if err != nil {
			return "", fmt.Errorf("%s:%d: %w", filePath, lineNo, err)
		}

		if isImport {
			if depth != 0 {
				return "", fmt.Errorf("%s:%d: import declarations are only allowed at top level", filePath, lineNo)
			}

			target, err := l.resolveImport(spec, filePath, lineNo)
			if err != nil {
				return "", err
			}

			chunk, err := l.loadFile(target)
			if err != nil {
				return "", err
			}
			appendChunk(&imports, chunk)
			continue
		}

		body.WriteString(line)
		depth += braceDelta(line)
		if depth < 0 {
			depth = 0
		}
	}

	out := imports.String() + body.String()
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out, nil
}

func (l *loader) resolveImport(spec string, importer string, line int) (string, error) {
	if spec == "" {
		return "", fmt.Errorf("%s:%d: empty import path", importer, line)
	}
	if strings.ContainsAny(spec, "\x00\r\n") {
		return "", fmt.Errorf("%s:%d: import path %q contains an invalid character", importer, line, spec)
	}
	if vol := filepath.VolumeName(spec); vol != "" {
		return "", fmt.Errorf("%s:%d: host absolute import paths are not allowed: %q", importer, line, spec)
	}

	importPath := strings.TrimLeft(spec, "/")
	if importPath == "" {
		return "", fmt.Errorf("%s:%d: empty import path", importer, line)
	}
	if !strings.HasSuffix(importPath, ".ped") {
		importPath += ".ped"
	}

	target := filepath.Clean(filepath.Join(l.root, filepath.FromSlash(importPath)))
	if !withinDir(l.root, target) {
		return "", fmt.Errorf("%s:%d: import %q escapes project root", importer, line, spec)
	}

	realTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", fmt.Errorf("%s:%d: import %q not found at %s", importer, line, spec, target)
	}
	realTarget = filepath.Clean(realTarget)

	if !withinDir(l.rootReal, realTarget) {
		return "", fmt.Errorf("%s:%d: import %q escapes project root through symlink", importer, line, spec)
	}

	return realTarget, nil
}

func parseImportLine(line string) (string, bool, error) {
	trimmed := strings.TrimSpace(line)
	if !startsWithImportKeyword(trimmed) {
		return "", false, nil
	}

	rest := strings.TrimSpace(trimmed[len("import"):])
	if rest == "" || rest[0] != '"' {
		return "", true, fmt.Errorf("invalid import declaration: expected import \"path\"")
	}

	for i := 1; i < len(rest); i++ {
		switch rest[i] {
		case '\\':
			i++
		case '"':
			lit := rest[:i+1]
			spec, err := strconv.Unquote(lit)
			if err != nil {
				return "", true, fmt.Errorf("invalid import string %s", lit)
			}

			tail := strings.TrimSpace(rest[i+1:])
			if tail != "" && !strings.HasPrefix(tail, "#") {
				return "", true, fmt.Errorf("invalid import declaration: unexpected text after import path")
			}

			return spec, true, nil
		}
	}

	return "", true, fmt.Errorf("invalid import declaration: unterminated import string")
}

func startsWithImportKeyword(s string) bool {
	if !strings.HasPrefix(s, "import") {
		return false
	}
	if len(s) == len("import") {
		return true
	}
	next := s[len("import")]
	return next == ' ' || next == '\t' || next == '"'
}

func appendChunk(b *strings.Builder, chunk string) {
	if chunk == "" {
		return
	}
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(chunk)
	if !strings.HasSuffix(chunk, "\n") {
		b.WriteByte('\n')
	}
}

func withinDir(root string, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func braceDelta(line string) int {
	delta := 0
	inString := false
	inChar := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if inChar {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '\'' {
				inChar = false
			}
			continue
		}

		switch ch {
		case '#':
			return delta
		case '"':
			inString = true
		case '\'':
			inChar = true
		case '{':
			delta++
		case '}':
			delta--
		}
	}

	return delta
}
