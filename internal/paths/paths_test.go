package paths

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestWithinAndEqual(t *testing.T) {
	base := filepath.Join(string(filepath.Separator), "tmp", "gvm")
	if !Within(base, base) {
		t.Fatal("a path should be within itself")
	}
	if !Within(filepath.Join(base, "cache"), base) {
		t.Fatal("child path should be within parent")
	}
	if Within(base, filepath.Join(base, "cache")) {
		t.Fatal("parent should not be within child")
	}
	if !Equal(base, base+string(filepath.Separator)) {
		t.Fatal("trailing separator should be ignored")
	}
	if runtime.GOOS == "windows" && !Equal(`C:\Gvm`, `c:\gvm`) {
		t.Fatal("Windows paths should compare case-insensitively")
	}
}

func TestReplaceRoot(t *testing.T) {
	old := filepath.Join(string(filepath.Separator), "tmp", "gvm")
	new := filepath.Join(string(filepath.Separator), "data", "gvm")
	in := "export GVM_ROOT=\"" + old + "\"\nPATH=\"" + old + "/shims:$PATH\"\nleave=" + old + "more\n"
	want := "export GVM_ROOT=\"" + new + "\"\nPATH=\"" + new + "/shims:$PATH\"\nleave=" + old + "more\n"
	if got := ReplaceRoot(in, old, new); got != want {
		t.Fatalf("ReplaceRoot mismatch:\n got: %q\nwant: %q", got, want)
	}
}
