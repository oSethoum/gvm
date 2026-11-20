package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shell"
	"github.com/osethoum/gvm/internal/shim"
)

var initInstall bool

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Print shell integration snippet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := args[0]
		fmt.Print(shell.Snippet(sh))
		// ensure shims exist
		paths.EnsureDirs()
		if err := shim.GenerateShims(); err != nil {
			return err
		}
		if initInstall {
			profile := shell.ProfileHint(sh)
			if profile == "" {
				return fmt.Errorf("unknown shell")
			}
			f, err := os.OpenFile(profile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.WriteString("\n" + shell.Snippet(sh) + "\n")
			return err
		}
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initInstall, "install", false, "append snippet to profile")
	rootCmd.AddCommand(initCmd)
}
