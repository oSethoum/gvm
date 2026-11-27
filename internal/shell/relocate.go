package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/osethoum/gvm/internal/paths"
)

// Relocate rewrites everything that stores an absolute gvm path: shell
// profiles and, on Windows, the user-scope Go environment variables and PATH.
func Relocate(oldRoot, newRoot, oldBinDir, newBinDir, active string) error {
	_ = os.Setenv("GVM_ROOT", newRoot)
	var errs []error
	if err := updateProfiles(oldRoot, newRoot, oldBinDir, newBinDir); err != nil {
		errs = append(errs, err)
	}
	if runtime.GOOS == "windows" {
		if err := updateWindowsEnv(oldRoot, newRoot, oldBinDir, newBinDir, active); err != nil {
			errs = append(errs, err)
		}
	}
	updateProcessEnv(oldRoot, newRoot, oldBinDir, newBinDir, active)
	if len(errs) > 0 {
		return fmt.Errorf("some environment locations could not be updated: %w", errs[0])
	}
	return nil
}

func updateProfiles(oldRoot, newRoot, oldBinDir, newBinDir string) error {
	var errs []error
	for _, sh := range []string{"powershell", "bash", "zsh", "fish"} {
		p := ProfileHint(sh)
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s := string(data)
		s = paths.ReplaceRoot(s, oldRoot, newRoot)
		if oldBinDir != "" && !paths.Within(oldBinDir, oldRoot) {
			s = paths.ReplaceRoot(s, oldBinDir, newBinDir)
		}
		if s != string(data) {
			if err := os.WriteFile(p, []byte(s), 0644); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func updateWindowsEnv(oldRoot, newRoot, oldBinDir, newBinDir, active string) error {
	parts := strings.Split(UserEnv("PATH"), ";")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		entry := strings.TrimSpace(strings.Trim(part, `"'`))
		if entry == "" {
			continue
		}
		entry = translatePath(entry, oldRoot, newRoot, oldBinDir, newBinDir)
		key := strings.ToLower(entry)
		if !seen[key] {
			seen[key] = true
			out = append(out, entry)
		}
	}
	setUserEnv("PATH", strings.Join(out, ";"))
	if active != "" {
		return ApplyEnv(active)
	}
	if !pathListContains(UserEnv("PATH"), filepath.Join(newRoot, "shims"), ';') {
		shims := filepath.Join(newRoot, "shims")
		cur := UserEnv("PATH")
		if cur == "" {
			setUserEnv("PATH", shims)
		} else {
			setUserEnv("PATH", shims+";"+cur)
		}
	}
	return nil
}

func translatePath(entry, oldRoot, newRoot, oldBinDir, newBinDir string) string {
	if oldBinDir != "" && paths.Equal(entry, oldBinDir) {
		return newBinDir
	}
	if paths.Equal(entry, oldRoot) {
		return newRoot
	}
	if paths.Within(entry, oldRoot) {
		rel, err := filepath.Rel(oldRoot, entry)
		if err == nil {
			return filepath.Join(newRoot, rel)
		}
	}
	if oldBinDir != "" && paths.Within(entry, oldBinDir) {
		rel, err := filepath.Rel(oldBinDir, entry)
		if err == nil {
			return filepath.Join(newBinDir, rel)
		}
	}
	return entry
}

func updateProcessEnv(oldRoot, newRoot, oldBinDir, newBinDir, active string) {
	_ = os.Setenv("GVM_ROOT", newRoot)
	if active != "" {
		goroot := filepath.Join(newRoot, "versions", active)
		gopath := filepath.Join(newRoot, "gopath", active)
		_ = os.Setenv("GOROOT", goroot)
		_ = os.Setenv("GOPATH", gopath)
		_ = os.Setenv("GOBIN", filepath.Join(gopath, "bin"))
		_ = os.Setenv("GOMODCACHE", filepath.Join(newRoot, "modcache"))
		_ = os.Setenv("GOTOOLCHAIN", "local")
	}
	sep := string(os.PathListSeparator)
	parts := strings.Split(os.Getenv("PATH"), sep)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, translatePath(part, oldRoot, newRoot, oldBinDir, newBinDir))
		}
	}
	_ = os.Setenv("PATH", strings.Join(out, sep))
}
