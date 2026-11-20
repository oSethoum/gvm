package shim

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/version"
)

func ActiveVersion() string {
	data, err := os.ReadFile(paths.ActiveFile())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SetActive(v string) error {
	v = version.Normalize(v)
	if err := os.MkdirAll(paths.Root(), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(paths.ActiveFile(), []byte(v+"\n"), 0644); err != nil {
		return err
	}
	if err := GenerateShims(); err != nil {
		return err
	}
	_ = createCurrentLink(v)
	return nil
}

func GenerateShims() error {
	if err := os.MkdirAll(paths.ShimsDir(), 0755); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		goCmd := "@echo off\r\n" +
			"setlocal\r\n" +
			"set \"GVM_HOME=" + paths.Root() + "\"\r\n" +
			"for /f \"usebackq delims=\" %%v in (\"%GVM_HOME%\\active\") do set \"GVM_VER=%%v\"\r\n" +
			"set \"GOROOT=%GVM_HOME%\\versions\\%GVM_VER%\"\r\n" +
			"set \"GOPATH=%GVM_HOME%\\gopath\\%GVM_VER%\"\r\n" +
			"set \"GOBIN=%GOPATH%\\bin\"\r\n" +
			"set \"GOMODCACHE=%GVM_HOME%\\modcache\"\r\n" +
			"set \"GOTOOLCHAIN=local\"\r\n" +
			"set \"PATH=%GOROOT%\\bin;%GOBIN%;%PATH%\"\r\n" +
			"\"%GOROOT%\\bin\\go.exe\" %*\r\n"
		gofmtCmd := "@echo off\r\n" +
			"setlocal\r\n" +
			"set \"GVM_HOME=" + paths.Root() + "\"\r\n" +
			"for /f \"usebackq delims=\" %%v in (\"%GVM_HOME%\\active\") do set \"GVM_VER=%%v\"\r\n" +
			"set \"GOROOT=%GVM_HOME%\\versions\\%GVM_VER%\"\r\n" +
			"\"%GOROOT%\\bin\\gofmt.exe\" %*\r\n"
		if err := os.WriteFile(filepath.Join(paths.ShimsDir(), "go.cmd"), []byte(goCmd), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(paths.ShimsDir(), "gofmt.cmd"), []byte(gofmtCmd), 0755); err != nil {
			return err
		}
		// also provide extensionless copies for Git Bash/MSYS
		shGo := "#!/bin/sh\n" +
			"GVM_HOME=\"" + paths.Root() + "\"\n" +
			"GVM_VER=$(cat \"$GVM_HOME/active\")\n" +
			"GOROOT=\"$GVM_HOME/versions/$GVM_VER\"\n" +
			"GOPATH=\"$GVM_HOME/gopath/$GVM_VER\"\n" +
			"export GOROOT GOPATH GOBIN=\"$GOPATH/bin\" GOMODCACHE=\"$GVM_HOME/modcache\" GOTOOLCHAIN=local\n" +
			"export PATH=\"$GOROOT/bin:$GOBIN:$PATH\"\n" +
			"exec \"$GOROOT/bin/go\" \"$@\"\n"
		shGofmt := "#!/bin/sh\n" +
			"GVM_HOME=\"" + paths.Root() + "\"\n" +
			"GVM_VER=$(cat \"$GVM_HOME/active\")\n" +
			"GOROOT=\"$GVM_HOME/versions/$GVM_VER\"\n" +
			"exec \"$GOROOT/bin/gofmt\" \"$@\"\n"
		_ = os.WriteFile(filepath.Join(paths.ShimsDir(), "go"), []byte(shGo), 0755)
		_ = os.WriteFile(filepath.Join(paths.ShimsDir(), "gofmt"), []byte(shGofmt), 0755)
	} else {
		shGo := "#!/bin/sh\n" +
			"GVM_HOME=\"" + paths.Root() + "\"\n" +
			"GVM_VER=$(cat \"$GVM_HOME/active\")\n" +
			"GOROOT=\"$GVM_HOME/versions/$GVM_VER\"\n" +
			"GOPATH=\"$GVM_HOME/gopath/$GVM_VER\"\n" +
			"export GOROOT GOPATH GOBIN=\"$GOPATH/bin\" GOMODCACHE=\"$GVM_HOME/modcache\" GOTOOLCHAIN=local\n" +
			"export PATH=\"$GOROOT/bin:$GOBIN:$PATH\"\n" +
			"exec \"$GOROOT/bin/go\" \"$@\"\n"
		shGofmt := "#!/bin/sh\n" +
			"GVM_HOME=\"" + paths.Root() + "\"\n" +
			"GVM_VER=$(cat \"$GVM_HOME/active\")\n" +
			"GOROOT=\"$GVM_HOME/versions/$GVM_VER\"\n" +
			"exec \"$GOROOT/bin/gofmt\" \"$@\"\n"
		if err := os.WriteFile(filepath.Join(paths.ShimsDir(), "go"), []byte(shGo), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(paths.ShimsDir(), "gofmt"), []byte(shGofmt), 0755); err != nil {
			return err
		}
	}
	return nil
}

func createCurrentLink(v string) error {
	link := paths.CurrentLink()
	_ = os.Remove(link)
	target := paths.VersionDir(v)
	if runtime.GOOS == "windows" {
		// try junction
		cmd := exec.Command("cmd", "/c", "mklink", "/J", link, target)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}
	return os.Symlink(target, link)
}

func GOROOTActive() string {
	v := ActiveVersion()
	if v == "" {
		return ""
	}
	return paths.VersionDir(v)
}
func GOPATHActive() string {
	v := ActiveVersion()
	if v == "" {
		return ""
	}
	return paths.GopathDir(v)
}
