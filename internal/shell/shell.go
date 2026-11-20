package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/osethoum/gvm/internal/paths"
)

// ApplyEnv updates user-level environment variables (no admin required).
func ApplyEnv(v string) error {
	if runtime.GOOS != "windows" {
		// On Unix, env is managed via shell init snippets.
		return nil
	}
	goroot := paths.VersionDir(v)
	gopath := paths.GopathDir(v)
	gobin := filepath.Join(gopath, "bin")
	modcache := paths.ModCacheDir()
	shims := paths.ShimsDir()
	pairs := map[string]string{
		"GOROOT":      goroot,
		"GOPATH":      gopath,
		"GOBIN":       gobin,
		"GOMODCACHE":  modcache,
		"GOTOOLCHAIN": "local",
	}
	for k, val := range pairs {
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("[Environment]::SetEnvironmentVariable('%s','%s','User')", k, val))
		_ = cmd.Run()
	}
	// ensure shims on user PATH
	out, _ := exec.Command("powershell", "-NoProfile", "-Command",
		"[Environment]::GetEnvironmentVariable('PATH','User')").Output()
	cur := strings.TrimSpace(string(out))
	if !strings.Contains(cur, shims) {
		newPath := shims
		if cur != "" {
			newPath = shims + ";" + cur
		}
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("[Environment]::SetEnvironmentVariable('PATH','%s','User')", newPath))
		_ = cmd.Run()
	}
	return nil
}

func Snippet(sh string) string {
	shims := paths.ShimsDir()
	root := paths.Root()
	switch sh {
	case "powershell", "pwsh":
		return fmt.Sprintf(`# gvm
$env:GVM_ROOT = "%s"
$env:PATH = "%s;" + $env:PATH
$gvmBin = if (Get-Command gvm.exe -ErrorAction SilentlyContinue) { (Get-Command gvm.exe).Source } else { "gvm" }
function gvm {
    & $gvmBin @args
    if ($LASTEXITCODE -eq 0 -and ($args -contains "use" -or $args -contains "install")) {
        $env:GOROOT = (& $gvmBin env 2>$null | Where-Object { $_ -match '^GOROOT=' }) -replace '^GOROOT=',''
        $env:GOPATH = (& $gvmBin env 2>$null | Where-Object { $_ -match '^GOPATH=' }) -replace '^GOPATH=',''
        $env:GOBIN = (& $gvmBin env 2>$null | Where-Object { $_ -match '^GOBIN=' }) -replace '^GOBIN=',''
        $env:GOMODCACHE = (& $gvmBin env 2>$null | Where-Object { $_ -match '^GOMODCACHE=' }) -replace '^GOMODCACHE=',''
        $env:GOTOOLCHAIN = "local"
    }
}
`, root, shims)
	case "bash", "zsh":
		return fmt.Sprintf(`# gvm
export GVM_ROOT="%s"
export PATH="%s:$PATH"
gvm() {
    "%s/gvm" "$@"
    if [ $? -eq 0 ] && echo "$*" | grep -q "use"; then
        eval $("${GVM_ROOT}/gvm" env)
    fi
}
`, root, shims, root)
	case "fish":
		return fmt.Sprintf(`# gvm
set -gx GVM_ROOT "%s"
set -gx PATH "%s" $PATH
`, root, shims)
	default:
		return fmt.Sprintf("# unsupported shell %s\nexport PATH=\"%s:$PATH\"\n", sh, shims)
	}
}

func ProfileHint(sh string) string {
	switch sh {
	case "powershell", "pwsh":
		return filepath.Join(osHome(), "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	case "bash":
		return filepath.Join(osHome(), ".bashrc")
	case "zsh":
		return filepath.Join(osHome(), ".zshrc")
	case "fish":
		return filepath.Join(osHome(), ".config", "fish", "config.fish")
	default:
		return ""
	}
}

func osHome() string {
	h, _ := os.UserHomeDir()
	return h
}
