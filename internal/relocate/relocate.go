// Package relocate moves the whole gvm tree to a new directory.
package relocate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/osethoum/gvm/internal/paths"
)

// Progress reports bytes copied during a move.
type Progress struct {
	Copied int64
	Total  int64
}

// Validate checks that src can be moved into dst without clobbering data.
func Validate(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("both source and destination are required")
	}
	if !filepath.IsAbs(dst) {
		return fmt.Errorf("destination must be an absolute path: %s", dst)
	}
	dst = filepath.Clean(dst)
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot read current gvm directory %s: %w", src, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("current gvm directory %s is not a directory", src)
	}
	if paths.Equal(src, dst) {
		return fmt.Errorf("destination is already the gvm directory: %s", dst)
	}
	if paths.Within(dst, src) {
		return fmt.Errorf("destination %s is inside the current gvm directory", dst)
	}
	if paths.Within(src, dst) {
		return fmt.Errorf("destination %s contains the current gvm directory", dst)
	}
	if existing, err := os.Stat(dst); err == nil {
		if !existing.IsDir() {
			return fmt.Errorf("destination exists and is not a directory: %s", dst)
		}
		entries, err := os.ReadDir(dst)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("destination is not empty: %s", dst)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// TotalSize returns the number of bytes that need to be copied.
func TotalSize(src string) (int64, error) {
	var total int64
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skip(p, src, d) || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// CopyTree copies src into dst, reporting progress after each chunk.
// If ctx is cancelled, the partially copied destination is removed and the
// context error is returned. It never deletes src.
func CopyTree(ctx context.Context, src, dst string, onProgress func(Progress)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	created := false
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
		created = true
	}

	var copied int64
	buf := make([]byte, 128*1024)
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel == "." || skip(p, src, d) {
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(target, info.Mode().Perm()|0700)
		}
		if d.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			in.Close()
			return err
		}
		for {
			if err := ctx.Err(); err != nil {
				out.Close()
				in.Close()
				return err
			}
			n, readErr := in.Read(buf)
			if n > 0 {
				if _, err := out.Write(buf[:n]); err != nil {
					out.Close()
					in.Close()
					return err
				}
				copied += int64(n)
				if onProgress != nil {
					onProgress(Progress{Copied: copied})
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				out.Close()
				in.Close()
				return readErr
			}
		}
		if err := out.Close(); err != nil {
			in.Close()
			return err
		}
		if err := in.Close(); err != nil {
			return err
		}
		return os.Chmod(target, info.Mode().Perm())
	})
	if err != nil {
		rollback(dst, created)
		return err
	}
	if onProgress != nil {
		onProgress(Progress{Copied: copied})
	}
	return nil
}

type treeEntry struct {
	size  int64
	mode  os.FileMode
	link  string
	isDir bool
	hash  string
}

func snapshot(root string) (map[string]treeEntry, error) {
	entries := map[string]treeEntry{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if rel == "." || skip(p, root, d) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		e := treeEntry{size: info.Size(), mode: info.Mode(), isDir: d.IsDir()}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			e.link = link
			e.size = 0
		} else if !e.isDir && info.Size() <= 1<<20 {
			h := sha256.New()
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			_, err = io.Copy(h, f)
			f.Close()
			if err != nil {
				return err
			}
			e.hash = hex.EncodeToString(h.Sum(nil))
		}
		entries[filepath.ToSlash(rel)] = e
		return nil
	})
	return entries, err
}

// Verify compares the copied tree with the source before anything is switched.
func Verify(src, dst string) error {
	want, err := snapshot(src)
	if err != nil {
		return err
	}
	got, err := snapshot(dst)
	if err != nil {
		return err
	}
	if len(got) != len(want) {
		return fmt.Errorf("copy is incomplete: %d of %d entries present", len(got), len(want))
	}
	for rel, w := range want {
		g, ok := got[rel]
		if !ok {
			return fmt.Errorf("copy is missing %s", rel)
		}
		if g.isDir != w.isDir || g.link != w.link || g.size != w.size || g.hash != w.hash {
			return fmt.Errorf("copy differs at %s", rel)
		}
	}
	return nil
}

// StagingDir returns a sibling staging directory for dst. Keeping it beside
// the destination makes the final rename atomic (same filesystem).
func StagingDir(dst string) string {
	return filepath.Join(filepath.Dir(dst), "."+filepath.Base(dst)+".gvm-moving-"+strconv.Itoa(os.Getpid()))
}

// CleanupStale removes staging folders left by interrupted earlier runs.
func CleanupStale(dst string) error {
	parent := filepath.Dir(dst)
	prefix := "." + filepath.Base(dst) + ".gvm-moving-"
	entries, err := os.ReadDir(parent)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			if err := os.RemoveAll(filepath.Join(parent, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// Finalize swaps a fully verified staging tree into the final destination.
func Finalize(staging, dst string) error {
	if info, err := os.Stat(dst); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("destination exists and is not a directory: %s", dst)
		}
		entries, err := os.ReadDir(dst)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("destination is no longer empty: %s", dst)
		}
		if err := os.Remove(dst); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(staging, dst)
}

// RollbackFinalize undoes Finalize, leaving the destination empty again.
func RollbackFinalize(staging, dst string) error {
	if err := os.Rename(dst, staging); err != nil {
		return err
	}
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	return os.MkdirAll(dst, 0755)
}

// CopyFile copies a single file, preserving its mode.
func CopyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode().Perm())
}

// RemoveSource deletes the old tree. On Windows a running executable cannot be
// deleted, so it is renamed out of the way first (best effort).
func RemoveSource(src string) error {
	if runtime.GOOS == "windows" {
		if exe, err := os.Executable(); err == nil && paths.Within(exe, src) {
			if err := os.Remove(exe); err != nil {
				old := exe + ".old"
				_ = os.Remove(old)
				if err := os.Rename(exe, old); err != nil {
					return fmt.Errorf("old executable is locked: %w", err)
				}
			}
		}
	}
	return os.RemoveAll(src)
}

// RetireOldBinary removes (or, on Windows, renames) a binary that was copied
// to the new location. Failure is returned but is not fatal.
func RetireOldBinary(exe string) error {
	if runtime.GOOS == "windows" {
		old := exe + ".old"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return err
		}
		return nil
	}
	return os.Remove(exe)
}

func skip(p, src string, d fs.DirEntry) bool {
	// "current" is an absolute symlink/junction and is recreated after the move.
	if rel, err := filepath.Rel(src, p); err == nil && rel == "current" {
		return true
	}
	return false
}

func rollback(dst string, created bool) {
	if created {
		_ = os.RemoveAll(dst)
		return
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		return
	}
	for _, e := range entries {
		_ = os.RemoveAll(filepath.Join(dst, e.Name()))
	}
}
