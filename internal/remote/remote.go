package remote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/osethoum/gvm/internal/config"
	"github.com/osethoum/gvm/internal/paths"
)

type GoFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

type GoVersion struct {
	Version string   `json:"version"` // "go1.27.4"
	Stable  bool     `json:"stable"`
	Files   []GoFile `json:"files"`
}

type cacheDoc struct {
	FetchedAt time.Time   `json:"fetchedAt"`
	Versions  []GoVersion `json:"versions"`
}

const ttl = 24 * time.Hour

func cachePath() string { return filepath.Join(paths.CacheDir(), "versions.json") }

func Fetch(refresh bool) ([]GoVersion, error) {
	if !refresh {
		if data, err := os.ReadFile(cachePath()); err == nil {
			var c cacheDoc
			if json.Unmarshal(data, &c) == nil && time.Since(c.FetchedAt) < ttl {
				return c.Versions, nil
			}
		}
	}
	cfg := config.Load()
	url := "https://go.dev/dl/?mode=json&include=all"
	// allow mirror override for the listing? listing always go.dev
	resp, err := http.Get(url)
	if err != nil {
		// try cache stale
		if data, err2 := os.ReadFile(cachePath()); err2 == nil {
			var c cacheDoc
			if json.Unmarshal(data, &c) == nil {
				return c.Versions, nil
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d fetching versions", resp.StatusCode)
	}
	var vs []GoVersion
	if err := json.NewDecoder(resp.Body).Decode(&vs); err != nil {
		return nil, err
	}
	_ = os.MkdirAll(paths.CacheDir(), 0755)
	data, _ := json.Marshal(cacheDoc{FetchedAt: time.Now(), Versions: vs})
	_ = os.WriteFile(cachePath(), data, 0644)
	_ = cfg
	return vs, nil
}

// ArchiveFor returns the archive file for current os/arch.
func ArchiveFor(v GoVersion) (GoFile, bool) {
	for _, f := range v.Files {
		if f.Kind == "archive" && f.OS == runtime.GOOS && f.Arch == runtime.GOARCH {
			return f, true
		}
	}
	return GoFile{}, false
}
