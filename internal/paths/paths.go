package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Root returns the gvm root directory, honoring GVM_ROOT.
func Root() string {
	if v := os.Getenv("GVM_ROOT"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".gvm")
}

func VersionsDir() string        { return filepath.Join(Root(), "versions") }
func VersionDir(v string) string { return filepath.Join(VersionsDir(), v) }
func GopathDir(v string) string  { return filepath.Join(Root(), "gopath", v) }
func ModCacheDir() string        { return filepath.Join(Root(), "modcache") }
func CacheDir() string           { return filepath.Join(Root(), "cache") }
func ShimsDir() string           { return filepath.Join(Root(), "shims") }
func ActiveFile() string         { return filepath.Join(Root(), "active") }
func ConfigFile() string         { return filepath.Join(Root(), "config.json") }
func CurrentLink() string        { return filepath.Join(Root(), "current") }

// EnsureDirs creates required directories.
func EnsureDirs() error {
	for _, d := range []string{VersionsDir(), filepath.Join(Root(), "gopath"), ModCacheDir(), CacheDir(), ShimsDir()} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// Equal reports whether two paths point at the same location.
// Comparison is case-insensitive on Windows.
func Equal(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// Within reports whether child is parent itself or lives inside it.
func Within(child, parent string) bool {
	if child == "" || parent == "" {
		return false
	}
	child, parent = filepath.Clean(child), filepath.Clean(parent)
	if Equal(child, parent) {
		return true
	}
	rel, err := filepath.Rel(parent, child)
	if err != nil || rel == ".." || filepath.IsAbs(rel) {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

// ReplaceRoot rewrites old as new in text, but only at path boundaries.
func ReplaceRoot(text, old, new string) string {
	if old == "" || new == "" || old == new {
		return text
	}
	var b strings.Builder
	for {
		i := strings.Index(text, old)
		if i < 0 {
			b.WriteString(text)
			return b.String()
		}
		b.WriteString(text[:i])
		end := i + len(old)
		if end == len(text) || isPathBoundary(text[end]) {
			b.WriteString(new)
		} else {
			b.WriteString(old)
		}
		text = text[end:]
	}
}

func isPathBoundary(c byte) bool {
	switch c {
	case '/', '\\', '"', '\'', ';', ':', ' ', '\t', '\n', '\r', ')', '}', ',':
		return true
	default:
		return false
	}
}

// ExeSuffix returns ".exe" on Windows.
func ExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
