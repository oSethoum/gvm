package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shim"
	"github.com/osethoum/gvm/internal/tui"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [version]",
	Short: "Uninstall a Go version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		paths.EnsureDirs()
		entries, _ := os.ReadDir(paths.VersionsDir())
		var installed []string
		for _, e := range entries {
			if e.IsDir() {
				installed = append(installed, e.Name())
			}
		}
		if len(installed) == 0 {
			return fmt.Errorf("no installed versions")
		}
		var chosen string
		if len(args) == 0 {
			items := []tui.VersionListItem{}
			for _, v := range installed {
				items = append(items, tui.VersionListItem{Version: v, Status: "installed"})
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
			chosen = args[0]
		}
		// find
		found := ""
		for _, v := range installed {
			if v == chosen {
				found = v
				break
			}
		}
		if found == "" {
			return fmt.Errorf("version %s not installed", chosen)
		}
		var confirm bool
		c := huh.NewConfirm().Title(fmt.Sprintf("Uninstall %s?", found)).Value(&confirm)
		if err := huh.NewForm(huh.NewGroup(c)).Run(); err != nil {
			return err
		}
		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}
		os.RemoveAll(paths.VersionDir(found))
		os.RemoveAll(paths.GopathDir(found))
		if shim.ActiveVersion() == found {
			os.Remove(paths.ActiveFile())
			fmt.Println("Active version removed; run 'gvm use' to select another.")
		}
		_ = filepath.Base
		fmt.Printf("Uninstalled %s\n", found)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
