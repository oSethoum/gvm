package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/remote"
	"github.com/osethoum/gvm/internal/tui"
	"github.com/osethoum/gvm/internal/version"
)

var availableCmd = &cobra.Command{
	Use:   "available",
	Short: "List available Go versions (remote)",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths.EnsureDirs()
		var vs []remote.GoVersion
		err := tui.RunSpinner("Fetching...", func() error {
			v, e := remote.Fetch(false)
			vs = v
			return e
		})
		if err != nil {
			return err
		}
		entries, _ := os.ReadDir(paths.VersionsDir())
		installed := map[string]bool{}
		for _, e := range entries {
			if e.IsDir() {
				installed[e.Name()] = true
			}
		}
		fi, _ := os.Stdout.Stat()
		isTTY := (fi.Mode() & os.ModeCharDevice) != 0
		if isTTY {
			items := []tui.VersionListItem{}
			for _, v := range vs {
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
			if sel != "" {
				fmt.Printf("Selected: %s\n", sel)
			}
			return nil
		}
		for _, v := range vs {
			fmt.Println(version.Normalize(v.Version))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(availableCmd)
}
