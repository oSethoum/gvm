package paths

import (
	"os"
	"path/filepath"
	"runtime"
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

// ExeSuffix returns ".exe" on Windows.
func ExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
