package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osethoum/gvm/internal/install"
)

type ghRelease struct {
	TagName string `json:"tag_name"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gvm to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Current version: %s\n", Version)

		resp, err := http.Get("https://api.github.com/repos/oSethoum/gvm/releases/latest")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("http %d", resp.StatusCode)
		}
		var rel ghRelease
		if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
			return err
		}
		latest := strings.TrimPrefix(rel.TagName, "v")
		if latest == "" {
			return fmt.Errorf("could not determine latest version")
		}
		if Version != "dev" && latest == strings.TrimPrefix(Version, "v") {
			fmt.Println("Already up to date.")
			return nil
		}
		fmt.Printf("Updating to %s...\n", latest)

		goos := runtime.GOOS
		goarch := runtime.GOARCH
		ext := ".tar.gz"
		if goos == "windows" {
			ext = ".zip"
		}
		asset := fmt.Sprintf("gvm_%s_%s_%s%s", latest, goos, goarch, ext)
		url := fmt.Sprintf("https://github.com/oSethoum/gvm/releases/download/v%s/%s", latest, asset)

		tmp, err := os.MkdirTemp("", "gvm-update-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)

		archive := filepath.Join(tmp, asset)
		if err := install.DownloadFile(url, archive, nil); err != nil {
			return err
		}

		// verify checksum if available
		checksumURL := fmt.Sprintf("https://github.com/oSethoum/gvm/releases/download/v%s/checksums.txt", latest)
		checksumPath := filepath.Join(tmp, "checksums.txt")
		if err := install.DownloadFile(checksumURL, checksumPath, nil); err == nil {
			data, _ := os.ReadFile(checksumPath)
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[1] == asset {
					if err := install.VerifySHA256(archive, fields[0]); err != nil {
						return err
					}
				}
			}
		}

		extractDir := filepath.Join(tmp, "extract")
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return err
		}
		if err := install.Extract(archive, extractDir); err != nil {
			return err
		}

		binName := "gvm"
		if goos == "windows" {
			binName = "gvm.exe"
		}
		var newBin string
		_ = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && filepath.Base(path) == binName {
				newBin = path
				return filepath.SkipAll
			}
			return nil
		})
		if newBin == "" {
			return fmt.Errorf("binary not found in release")
		}

		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			return err
		}
		_ = os.Chmod(newBin, 0755)

		if goos == "windows" {
			oldPath := exePath + ".old"
			_ = os.Remove(oldPath)
			if err := os.Rename(exePath, oldPath); err != nil {
				return err
			}
			if err := os.Rename(newBin, exePath); err != nil {
				// best-effort restore
				_ = os.Rename(oldPath, exePath)
				return err
			}
		} else {
			if err := os.Rename(newBin, exePath); err != nil {
				return err
			}
		}

		fmt.Printf("Updated gvm to %s\n", latest)
		fmt.Println("Restart your shell or re-run gvm to use the new version.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}