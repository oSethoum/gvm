package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/paths"
	"github.com/osethoum/gvm/internal/shim"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print environment variables for the active version",
	Run: func(cmd *cobra.Command, args []string) {
		v := shim.ActiveVersion()
		if v == "" {
			fmt.Println("# no active version")
			return
		}
		goroot := paths.VersionDir(v)
		gopath := paths.GopathDir(v)
		gobin := gopath + string(os.PathSeparator) + "bin"
		modcache := paths.ModCacheDir()
		if os.Getenv("GOOS") == "windows" || (len(os.Args) > 0) {
			// print shell-agnostic KEY=VALUE
		}
		fmt.Printf("GOROOT=%s\n", goroot)
		fmt.Printf("GOPATH=%s\n", gopath)
		fmt.Printf("GOBIN=%s\n", gobin)
		fmt.Printf("GOMODCACHE=%s\n", modcache)
		fmt.Printf("GOTOOLCHAIN=local\n")
		fmt.Printf("PATH=%s%s%s\n", goroot+string(os.PathSeparator)+"bin", string(os.PathListSeparator), gobin)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}
