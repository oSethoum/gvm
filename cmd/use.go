package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shell"
	"github.com/osethoum/gvm/internal/shim"
	"github.com/osethoum/gvm/internal/tui"
	"github.com/osethoum/gvm/internal/version"
)

var useCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "Switch active Go version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		paths.EnsureDirs()
		// list installed
		entries, _ := os.ReadDir(paths.VersionsDir())
		var installed []string
		for _, e := range entries {
			if e.IsDir() {
				installed = append(installed, e.Name())
			}
		}
		if len(installed) == 0 {
			return fmt.Errorf("no installed versions; run 'gvm install'")
		}
		var chosen string
		if len(args) == 0 {
			// try .go-version
			if data, err := os.ReadFile(".go-version"); err == nil {
				chosen = version.Normalize(string(data))
			}
		}
		if chosen == "" && len(args) > 0 {
			chosen = version.Normalize(args[0])
		}
		if chosen == "" {
			items := []tui.VersionListItem{}
			for _, v := range installed {
				status := "installed"
				if v == shim.ActiveVersion() {
					status = "active"
				}
				items = append(items, tui.VersionListItem{Version: v, Status: status})
			}
			sel, err := tui.RunVersionList(items)
			if err != nil {
				return err
			}
			if sel == "" {
				return nil
			}
			chosen = sel
		}
		// resolve partial against installed
		resolved, ok := version.MatchLatest(chosen, installed)
		if !ok {
			resolved = chosen
		}
		// validate installed
		found := false
		for _, v := range installed {
			if v == resolved || version.Normalize(v) == version.Normalize(resolved) {
				resolved = v
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("version %s not installed", chosen)
		}
		if err := shim.SetActive(resolved); err != nil {
			return err
		}
		_ = shell.ApplyEnv(resolved)
		fmt.Printf("Now using Go %s\nGOROOT=%s\nGOPATH=%s\n", resolved, paths.VersionDir(resolved), paths.GopathDir(resolved))
		fmt.Println("Restart terminals/IDEs for PATH changes to take effect.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
