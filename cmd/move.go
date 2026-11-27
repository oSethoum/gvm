package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/relocate"
	"github.com/osethoum/gvm/internal/shell"
	"github.com/osethoum/gvm/internal/shim"
	"github.com/osethoum/gvm/internal/tui"
)

var moveDir string
var moveYes bool

var moveCmd = &cobra.Command{
	Use:     "move [new-directory]",
	Aliases: []string{"relocate", "set-root", "change-dir"},
	Short:   "Move gvm (versions, caches, modules and env) to another directory",
	Long: "Move the entire gvm tree - Go versions, GOPATHs, module cache, download cache,\n" +
		"shims and configuration - to a new directory and update all environment\n" +
		"references. With no argument an interactive folder browser is shown.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stopSignals()
		ctx, cancel := context.WithCancel(baseCtx)
		defer cancel()

		src := filepath.Clean(paths.Root())
		dst := moveDir
		if dst == "" && len(args) > 0 {
			dst = args[0]
		}
		fi, statErr := os.Stdout.Stat()
		isTTY := statErr == nil && (fi.Mode()&os.ModeCharDevice) != 0
		if dst == "" {
			if !isTTY {
				return fmt.Errorf("choose a destination with --dir (interactive browsing needs a TTY)")
			}
			sel, err := tui.RunDirPicker(src)
			if err != nil {
				return err
			}
			if sel == "" {
				fmt.Println("Cancelled.")
				return nil
			}
			dst = sel
		}
		if !filepath.IsAbs(dst) {
			abs, err := filepath.Abs(dst)
			if err != nil {
				return err
			}
			dst = abs
		}
		dst = filepath.Clean(dst)
		if err := relocate.Validate(src, dst); err != nil {
			return err
		}

		if isTTY && !moveYes {
			ok := false
			confirm := huh.NewConfirm().
				Title(fmt.Sprintf("Move %s to %s?", src, dst)).
				Description("The old directory is deleted only after the copy has been verified.").
				Value(&ok)
			if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil {
				return err
			}
			if !ok {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := relocate.CleanupStale(dst); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not clean up an interrupted copy: %v\n", err)
		}
		staging := relocate.StagingDir(dst)
		total, err := relocate.TotalSize(src)
		if err != nil {
			return err
		}
		active := shim.ActiveVersion()
		exe, exeErr := os.Executable()
		oldBinDir := ""
		if exeErr == nil {
			oldBinDir = filepath.Dir(exe)
		}

		var prog *tea.Program
		if isTTY {
			prog = tea.NewProgram(tui.NewMoveModel(dst, cancel))
		}
		onProgress := func(p relocate.Progress) {
			if prog != nil {
				prog.Send(tui.MoveProgressMsg{Copied: p.Copied, Total: total})
				return
			}
			fmt.Fprint(os.Stderr, "\rMoving gvm...")
		}

		// Nothing outside the staging folder is touched until the copy is done
		// and verified. Cancellation rolls the staging folder back.
		copyErr := runCopy(ctx, prog, src, staging, onProgress)
		if errors.Is(copyErr, context.Canceled) {
			fmt.Println("Cancelled; gvm was not changed.")
			return nil
		}
		if copyErr != nil {
			return copyErr
		}
		if err := ctx.Err(); err != nil {
			_ = os.RemoveAll(staging)
			fmt.Println("Cancelled; gvm was not changed.")
			return nil
		}
		if err := relocate.Verify(src, staging); err != nil {
			_ = os.RemoveAll(staging)
			return fmt.Errorf("copy verification failed; gvm was not changed: %w", err)
		}

		// The gvm executable may live outside the gvm tree (GVM_INSTALL_DIR).
		// Add it to the staging tree so the destination is complete at swap time.
		newBinDir := filepath.Join(dst, "bin")
		copiedBinary := false
		if exeErr == nil {
			if paths.Within(oldBinDir, src) {
				if rel, err := filepath.Rel(src, oldBinDir); err == nil {
					newBinDir = filepath.Join(dst, rel)
				}
			} else {
				if err := relocate.CopyFile(exe, filepath.Join(staging, "bin", filepath.Base(exe))); err != nil {
					_ = os.RemoveAll(staging)
					return fmt.Errorf("copy finished, but the executable could not be copied: %w", err)
				}
				copiedBinary = true
			}
		}
		if err := ctx.Err(); err != nil {
			_ = os.RemoveAll(staging)
			fmt.Println("Cancelled; gvm was not changed.")
			return nil
		}
		if err := relocate.Finalize(staging, dst); err != nil {
			return fmt.Errorf("copy is complete but could not be put in place: %w", err)
		}
		if err := ctx.Err(); err != nil {
			if rbErr := relocate.RollbackFinalize(staging, dst); rbErr != nil {
				fmt.Fprintf(os.Stderr, "warning: interrupted swap could not be rolled back: %v\n", rbErr)
			}
			fmt.Println("Cancelled; gvm was not changed.")
			return nil
		}

		// From here on all helpers resolve paths against the new root.
		_ = os.Setenv("GVM_ROOT", dst)
		if active != "" {
			if err := shim.SetActive(active); err != nil {
				return fmt.Errorf("gvm tree was copied, but shims could not be updated: %w", err)
			}
		} else if err := shim.GenerateShims(); err != nil {
			return fmt.Errorf("gvm tree was copied, but shims could not be updated: %w", err)
		}

		if err := shell.Relocate(src, dst, oldBinDir, newBinDir, active); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			fmt.Fprintln(os.Stderr, "You may need to update your shell profile manually.")
		}

		removed := true
		if err := relocate.RemoveSource(src); err != nil {
			removed = false
			fmt.Fprintf(os.Stderr, "warning: old directory %s could not be removed: %v\n", src, err)
		}
		if copiedBinary {
			if err := relocate.RetireOldBinary(exe); err != nil {
				fmt.Fprintf(os.Stderr, "warning: old gvm executable %s could not be removed: %v\n", exe, err)
			}
		}

		fmt.Printf("gvm moved to %s\n", dst)
		if active != "" {
			fmt.Printf("Active version: %s\n", active)
		}
		fmt.Println("Restart your terminal/IDE for the new environment to take effect.")
		if !removed {
			fmt.Println("The old directory still exists; delete it manually once no process uses it.")
		}
		return nil
	},
}

func runCopy(ctx context.Context, prog *tea.Program, src, staging string, onProgress func(relocate.Progress)) error {
	if prog == nil {
		err := relocate.CopyTree(ctx, src, staging, onProgress)
		fmt.Fprintln(os.Stderr)
		return err
	}
	done := make(chan error, 1)
	go func() {
		err := relocate.CopyTree(ctx, src, staging, onProgress)
		done <- err
		prog.Send(tui.MoveDoneMsg{Err: err})
	}()
	final, err := prog.Run()
	copyErr := <-done
	if err != nil {
		return err
	}
	if model, ok := final.(tui.MoveModel); ok && model.Err != nil {
		return model.Err
	}
	return copyErr
}

func init() {
	moveCmd.Flags().StringVar(&moveDir, "dir", "", "destination directory (skip the folder browser)")
	moveCmd.Flags().BoolVarP(&moveYes, "yes", "y", false, "skip confirmation")
	rootCmd.AddCommand(moveCmd)
}
