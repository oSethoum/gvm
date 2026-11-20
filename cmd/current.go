package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shim"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active Go version",
	Run: func(cmd *cobra.Command, args []string) {
		v := shim.ActiveVersion()
		if v == "" {
			fmt.Println("No active version")
			return
		}
		fmt.Printf("%s\nGOROOT=%s\nGOPATH=%s\n", v, paths.VersionDir(v), paths.GopathDir(v))
		_ = os.Args
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
