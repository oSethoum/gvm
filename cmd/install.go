package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/config"
	"github.com/osethoum/gvm/internal/install"
	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/remote"
	"github.com/osethoum/gvm/internal/shell"
	"github.com/osethoum/gvm/internal/shim"
	"github.com/osethoum/gvm/internal/tui"
	"github.com/osethoum/gvm/internal/version"
)

var installPre bool
var installDefault bool

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Install a Go version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		paths.EnsureDirs()
		var chosen string
		if len(args) == 0 {
			var vs []remote.GoVersion
			err := tui.RunSpinner("Fetching versions...", func() error {
				v, e := remote.Fetch(false)
				vs = v
				return e
			})
			if err != nil {
				return err
			}
			// installed set
			entries, _ := os.ReadDir(paths.VersionsDir())
			installed := map[string]bool{}
			for _, e := range entries {
				if e.IsDir() {
					installed[e.Name()] = true
				}
			}
			items := []tui.VersionListItem{}
			for _, v := range vs {
				if !v.Stable && !installPre && !cfg.PreRelease {
					continue
				}
				if _, ok := remote.ArchiveFor(v); !ok {
					continue
				}
				ver := version.Normalize(v.Version)
				if installed[ver] || installed[v.Version] {
					continue
				}
				items = append(items, tui.VersionListItem{Version: ver, Status: "available"})
			}
			sel, err := tui.RunVersionList(items)
			if err != nil {
				return err
			}
			if sel == "" {
				return nil
			}
			chosen = sel
		} else {
			chosen = version.Normalize(args[0])
		}

		// resolve partial
		vs, err := remote.Fetch(false)
		if err != nil {
			return err
		}
		var candidates []string
		for _, v := range vs {
			if !v.Stable && !installPre && !cfg.PreRelease {
				continue
			}
			candidates = append(candidates, version.Normalize(v.Version))
		}
		resolved, ok := version.MatchLatest(chosen, candidates)
		if !ok {
			// maybe chosen is full version present
			resolved = chosen
		}
		// find remote entry
		var rv remote.GoVersion
		found := false
		for _, v := range vs {
			if version.Normalize(v.Version) == resolved {
				rv = v
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("version %s not found", chosen)
		}
		af, ok := remote.ArchiveFor(rv)
		if !ok {
			return fmt.Errorf("no archive for %s/%s", "current os/arch", resolved)
		}

		ver := version.Normalize(rv.Version)
		dest := paths.VersionDir(ver)
		if _, err := os.Stat(dest); err == nil {
			// already installed
			var reinstall bool
			confirm := huh.NewConfirm().Title(fmt.Sprintf("%s already installed. Reinstall?", ver)).Value(&reinstall)
			if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil {
				return err
			}
			if !reinstall {
				fmt.Println("Skipped.")
				return nil
			}
			os.RemoveAll(dest)
		}

		url := install.MirrorURL(af.Filename)
		cacheFile := filepath.Join(paths.CacheDir(), af.Filename)
		fmt.Printf("Installing %s...\n", ver)
		fi, _ := os.Stdout.Stat()
		isTTY := (fi.Mode() & os.ModeCharDevice) != 0
		if isTTY {
			p := tea.NewProgram(tui.NewDownloadModel())
			go func() {
				err := install.DownloadFile(url, cacheFile, func(d, t int64) {
					p.Send(tui.DownloadProgressMsg{Downloaded: d, Total: af.Size})
				})
				if err == nil {
					err = install.VerifySHA256(cacheFile, af.SHA256)
				}
				p.Send(tui.DownloadDoneMsg{Err: err})
			}()
			m, err := p.Run()
			if err != nil {
				return err
			}
			dm := m.(tui.DownloadModel)
			if dm.Err != nil {
				return dm.Err
			}
		} else {
			err := install.DownloadFile(url, cacheFile, func(_, _ int64) {
				fmt.Fprint(os.Stderr, "\rDownloading...")
			})
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return err
			}
			if err := install.VerifySHA256(cacheFile, af.SHA256); err != nil {
				return err
			}
		}
		// extract
		fmt.Println("Extracting...")
		os.RemoveAll(dest)
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		if err := install.Extract(cacheFile, dest); err != nil {
			return err
		}
		// gopath
		gp := paths.GopathDir(ver)
		os.MkdirAll(filepath.Join(gp, "bin"), 0755)
		os.MkdirAll(filepath.Join(gp, "src"), 0755)
		os.MkdirAll(filepath.Join(gp, "pkg"), 0755)

		fmt.Printf("Installed %s\n", ver)

		// auto-set active if none
		if installDefault || shim.ActiveVersion() == "" {
			if err := shim.SetActive(ver); err != nil {
				return err
			}
			_ = shell.ApplyEnv(ver)
			fmt.Printf("Set as active: %s\n", ver)
			fmt.Println("Restart terminals/IDEs for PATH changes to take effect.")
		} else {
			fmt.Printf("Run 'gvm use %s' to switch.\n", ver)
		}
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&installPre, "pre", false, "include pre-release versions")
	installCmd.Flags().BoolVar(&installDefault, "default", false, "set as active after install")
	rootCmd.AddCommand(installCmd)
}
