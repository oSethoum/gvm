package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/config"
	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/remote"
	"github.com/osethoum/gvm/internal/shim"
	"github.com/osethoum/gvm/internal/tui"
	"github.com/osethoum/gvm/internal/version"
)

var listInstalled bool
var listRemote bool
var listPre bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Go versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		paths.EnsureDirs()
		// installed set
		entries, _ := os.ReadDir(paths.VersionsDir())
		installed := map[string]bool{}
		var installedList []string
		for _, e := range entries {
			if e.IsDir() {
				installed[e.Name()] = true
				installedList = append(installedList, e.Name())
			}
		}
		active := shim.ActiveVersion()

		fi, _ := os.Stdout.Stat()
		isTTY := (fi.Mode() & os.ModeCharDevice) != 0
		if isTTY && !listInstalled && !listRemote && !listPre {
			var vs []remote.GoVersion
			_ = tui.RunSpinner("Fetching available versions...", func() error {
				v, e := remote.Fetch(false)
				vs = v
				return e
			})
			items := []tui.VersionListItem{}
			for _, v := range installedList {
				status := "installed"
				if v == active {
					status = "active"
				}
				items = append(items, tui.VersionListItem{Version: v, Status: status})
			}
			for _, v := range vs {
				ver := version.Normalize(v.Version)
				if installed[ver] || installed[v.Version] {
					continue
				}
				if !v.Stable && !cfg.PreRelease {
					continue
				}
				items = append(items, tui.VersionListItem{Version: ver, Status: "available"})
			}
			sel, err := tui.RunVersionList(items)
			if err != nil {
				return err
			}
			if sel != "" {
				fmt.Printf("Selected: %s\n", sel)
			}
			return nil
		}

		showRemote := !listInstalled
		if listRemote {
			showRemote = true
			listInstalled = false
		}
		if listInstalled {
			showRemote = false
		}

		if !listRemote {
			fmt.Println(tui.RenderTitle("Installed"))
			if len(installedList) == 0 {
				fmt.Println(tui.StyleHint.Render("  none"))
			}
			for _, v := range installedList {
				mark := "  "
				if v == active {
					mark = "* "
				}
				line := fmt.Sprintf("%s%s", mark, v)
				if v == active {
					fmt.Println(tui.StyleActive.Render(line))
				} else {
					fmt.Println(tui.StyleInstalled.Render(line))
				}
			}
			fmt.Println()
		}

		if !showRemote {
			return nil
		}

		var vs []remote.GoVersion
		err := tui.RunSpinner("Fetching available versions...", func() error {
			v, e := remote.Fetch(false)
			vs = v
			return e
		})
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(tui.RenderTitle("Available (not installed)"))
		count := 0
		for _, v := range vs {
			ver := version.Normalize(v.Version)
			if installed[ver] || installed[v.Version] {
				continue
			}
			if !v.Stable && !listPre && !cfg.PreRelease {
				continue
			}
			fmt.Println(tui.StyleAvailable.Render("  " + ver))
			count++
		}
		if count == 0 {
			fmt.Println(tui.StyleHint.Render("  none"))
		}
		_ = strings.TrimSpace
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listInstalled, "installed", false, "show installed only")
	listCmd.Flags().BoolVar(&listRemote, "remote", false, "show remote only")
	listCmd.Flags().BoolVar(&listPre, "pre", false, "include pre-release")
	rootCmd.AddCommand(listCmd)
}
