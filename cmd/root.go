package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "gvm",
	Short: "Go version manager",
	Long:  "gvm - a polished Go version manager powered by Charmbracelet.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			return
		}
		sel, err := tui.RunMenu()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if sel == "" || sel == "quit" {
			return
		}
		// sub-command args
		if sel == "init" || sel == "install" || sel == "use" || sel == "uninstall" {
			// let cobra handle; re-run with selected command
		}
		cmd.SetArgs([]string{sel})
		_ = cmd.Execute()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("no-color", false, "disable color output")
}
