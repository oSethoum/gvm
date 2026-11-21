package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shim"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose gvm setup",
	Run: func(cmd *cobra.Command, args []string) {
		paths.EnsureDirs()
		fmt.Printf("Root: %s\n", paths.Root())
		fmt.Printf("Active: %s\n", shim.ActiveVersion())
		fmt.Printf("Shims: %s\n", paths.ShimsDir())
		pathEnv := os.Getenv("PATH")
		if !strings.Contains(pathEnv, paths.ShimsDir()) {
			fmt.Println("WARNING: shims dir not in PATH; run 'gvm init <shell>'")
		} else {
			fmt.Println("OK: shims on PATH")
		}
		v := shim.ActiveVersion()
		if v != "" {
			gobin := paths.GopathDir(v) + string(os.PathSeparator) + "bin"
			if !strings.Contains(pathEnv, gobin) {
				fmt.Println("HINT: GOBIN not on PATH; run 'gvm init <shell>' or add manually")
			} else {
				fmt.Println("OK: GOBIN on PATH")
			}
			if _, err := os.Stat(paths.VersionDir(v)); err != nil {
				fmt.Println("ERROR: active version dir missing")
			} else {
				fmt.Println("OK: active version installed")
			}
		}
		fmt.Println("Refresh current shell: gvm init <shell>  (then source profile), or eval 'gvm env'")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
