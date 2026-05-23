package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var versionRE = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func main() {
	var quiet bool
	flag.BoolVar(&quiet, "quiet", false, "print only the new version")
	flag.BoolVar(&quiet, "q", false, "print only the new version (shorthand)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-quiet|-q] <version-file> <patch|minor|major>\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s .version patch\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -quiet .version minor\n", os.Args[0])
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(2)
	}

	versionFile := strings.TrimSpace(args[0])
	mode := strings.ToLower(strings.TrimSpace(args[1]))

	oldVersion, err := readVersionFile(versionFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	newVersion, err := bumpVersion(oldVersion, mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := writeVersionFile(versionFile, newVersion); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if quiet {
		fmt.Println(newVersion)
	} else {
		fmt.Printf("%s -> %s\n", oldVersion, newVersion)
	}
}

func readVersionFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("version file path must not be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read version file %q: %w", path, err)
	}

	version := strings.TrimSpace(string(data))
	if !versionRE.MatchString(version) {
		return "", fmt.Errorf("invalid version in %q: %q (expected X.Y.Z)", path, version)
	}

	return version, nil
}

func writeVersionFile(path, version string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("version file path must not be empty")
	}
	if !versionRE.MatchString(version) {
		return fmt.Errorf("invalid version %q (expected X.Y.Z)", version)
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create parent dir for %q: %w", path, err)
		}
	}

	if err := os.WriteFile(path, []byte(version+"\n"), 0o644); err != nil {
		return fmt.Errorf("write version file %q: %w", path, err)
	}
	return nil
}

func bumpVersion(version, mode string) (string, error) {
	m := versionRE.FindStringSubmatch(strings.TrimSpace(version))
	if m == nil {
		return "", fmt.Errorf("invalid version %q (expected X.Y.Z)", version)
	}

	major, err := strconv.Atoi(m[1])
	if err != nil {
		return "", fmt.Errorf("parse major version: %w", err)
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return "", fmt.Errorf("parse minor version: %w", err)
	}
	patch, err := strconv.Atoi(m[3])
	if err != nil {
		return "", fmt.Errorf("parse patch version: %w", err)
	}

	switch mode {
	case "patch":
		patch++
	case "minor":
		minor++
		patch = 0
	case "major":
		major++
		minor = 0
		patch = 0
	default:
		return "", fmt.Errorf("unknown bump mode %q (want patch, minor, or major)", mode)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
