package main

import (
	"fmt"
	"time"

	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Config struct {
	HTTPServer httpserver.Config `koanf:"http_server"`
	Features   map[string]string `koanf:"features"`

	LeaderboardAPI struct {
		BaseURL string        `koanf:"base_url"`
		Timeout time.Duration `koanf:"timeout"`
	} `koanf:"leaderboard_api"`
	ProjectAPI struct {
		BaseURL string        `koanf:"base_url"`
		Timeout time.Duration `koanf:"timeout"`
	} `koanf:"project_api"`
}

func loadConfig(configPath string) (Config, error) {

	var cfg Config

	err := config.Load(config.Options{
		Prefix:       "WEB",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: configPath,
	}, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("web: load config: %w", err)
	}

	return cfg, nil
}
