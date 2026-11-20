package config

import (
	"encoding/json"
	"os"

	"github.com/osethoum/gvm/internal/paths"
)

type Config struct {
	Mirror         string `json:"mirror"`
	GopathMode     string `json:"gopathMode"` // per-version | shared
	SharedModCache bool   `json:"sharedModCache"`
	PreRelease     bool   `json:"preRelease"`
}

func Default() Config {
	return Config{
		Mirror:         "https://go.dev/dl/",
		GopathMode:     "per-version",
		SharedModCache: true,
		PreRelease:     false,
	}
}

func Load() Config {
	cfg := Default()
	data, err := os.ReadFile(paths.ConfigFile())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	if cfg.Mirror == "" {
		cfg.Mirror = Default().Mirror
	}
	if cfg.GopathMode == "" {
		cfg.GopathMode = Default().GopathMode
	}
	return cfg
}

func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ConfigFile(), data, 0644)
}
