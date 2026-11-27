package shell

import (
	"path/filepath"
	"testing"
)

func TestTranslatePath(t *testing.T) {
	oldRoot := filepath.Join(string(filepath.Separator), "tmp", "gvm")
	newRoot := filepath.Join(string(filepath.Separator), "data", "gvm")
	oldBin := filepath.Join(string(filepath.Separator), "tmp", "bin")
	newBin := filepath.Join(newRoot, "bin")

	cases := []struct {
		in, want string
	}{
		{oldRoot, newRoot},
		{filepath.Join(oldRoot, "shims"), filepath.Join(newRoot, "shims")},
		{oldBin, newBin},
		{filepath.Join(oldBin, "gvm"), filepath.Join(newBin, "gvm")},
		{filepath.Join(string(filepath.Separator), "usr", "bin"), filepath.Join(string(filepath.Separator), "usr", "bin")},
	}
	for _, tc := range cases {
		if got := translatePath(tc.in, oldRoot, newRoot, oldBin, newBin); got != tc.want {
			t.Errorf("translatePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
