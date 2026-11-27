package relocate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "gvm")
	if err := os.MkdirAll(filepath.Join(src, "cache"), 0755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(base, "new-gvm")

	if err := Validate(src, dst); err != nil {
		t.Fatalf("valid destination rejected: %v", err)
	}
	if err := Validate(src, src); err == nil {
		t.Fatal("same directory should be rejected")
	}
	if err := Validate(src, filepath.Join(src, "sub")); err == nil {
		t.Fatal("nested destination should be rejected")
	}
	if err := Validate(src, base); err == nil {
		t.Fatal("parent destination should be rejected")
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "file"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Validate(src, dst); err == nil {
		t.Fatal("non-empty destination should be rejected")
	}
}

func TestCopyTreeAndVerify(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "gvm")
	dst := filepath.Join(base, "new-gvm")
	content := []byte("hello gvm")
	if err := os.MkdirAll(filepath.Join(src, "cache"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "cache", "archive.zip"), content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "active"), []byte("go1.23.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// "current" is always recreated after the move.
	if err := os.WriteFile(filepath.Join(src, "current"), []byte("old-link"), 0644); err != nil {
		t.Fatal(err)
	}

	total, err := TotalSize(src)
	if err != nil {
		t.Fatal(err)
	}
	if total != int64(len(content)+len("go1.23.0\n")) {
		t.Fatalf("unexpected total size: %d", total)
	}

	var copies []int64
	if err := CopyTree(context.Background(), src, dst, func(p Progress) { copies = append(copies, p.Copied) }); err != nil {
		t.Fatal(err)
	}
	if len(copies) == 0 || copies[len(copies)-1] != total {
		t.Fatalf("progress did not reach total: %v (total %d)", copies, total)
	}
	got, err := os.ReadFile(filepath.Join(dst, "cache", "archive.zip"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("copied content mismatch: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dst, "current")); !os.IsNotExist(err) {
		t.Fatalf("current link should not be copied: %v", err)
	}
	if err := Verify(src, dst); err != nil {
		t.Fatalf("verification failed for identical trees: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dst, "active"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(src, dst); err == nil {
		t.Fatal("verification should detect a modified file")
	}
}

func TestCopyTreeCancel(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "gvm")
	dst := filepath.Join(base, "staging")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "big"), make([]byte, 512*1024), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelled := false
	err := CopyTree(ctx, src, dst, func(p Progress) {
		if !cancelled && p.Copied > 0 {
			cancelled = true
			cancel()
		}
	})
	if err == nil {
		t.Fatal("cancelled copy should fail")
	}
	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Fatalf("partial copy should be rolled back: %v", statErr)
	}
}

func TestStagingCleanupAndFinalize(t *testing.T) {
	base := t.TempDir()
	dst := filepath.Join(base, "gvm")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	staging := StagingDir(dst)
	if filepath.Dir(staging) != filepath.Dir(dst) {
		t.Fatal("staging must be beside the destination")
	}
	if !strings.Contains(filepath.Base(staging), ".gvm-moving-") {
		t.Fatalf("unexpected staging name: %s", staging)
	}

	stale := filepath.Join(base, "."+filepath.Base(dst)+".gvm-moving-999")
	keep := filepath.Join(base, "keep")
	if err := os.MkdirAll(stale, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(keep, 0755); err != nil {
		t.Fatal(err)
	}
	if err := CleanupStale(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("stale staging folder was not removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatal("unrelated folder was removed")
	}

	if err := os.MkdirAll(staging, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "active"), []byte("go1.23.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Finalize(staging, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "active")); err != nil {
		t.Fatalf("finalized tree is missing active file: %v", err)
	}
	if err := RollbackFinalize(staging, dst); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("rollback should leave the destination empty")
	}
}
