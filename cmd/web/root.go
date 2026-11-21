package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/gocasters/rankr/pkg/featureflag"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "web",
	}

	preview := root.PersistentFlags().Bool(
		"preview", false, "Enable a local preview with sample data",
	)
	leaderboard := root.PersistentFlags().Bool(
		"leaderboard", false, "Enable the API-backed leaderboard page",
	)

	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.Getenv("CONFIG_PATH")

			if configPath == "" {
				projectRoot, err := path.PathProjectRoot()
				if err != nil {
					return fmt.Errorf("web: find project root: %w", err)
				}

				configPath = filepath.Join(
					projectRoot, "deploy", "web", "development", "config.yml",
				)
			}

			cfg, err := loadConfig(configPath)
			if err != nil {
				return err
			}

			// The CLI switches provide local defaults for unconfigured flags.
			if *preview || *leaderboard {
				if cfg.Features == nil {
					cfg.Features = make(map[string]string)
				}
				if _, configured := cfg.Features["web_enabled"]; !configured {
					cfg.Features["web_enabled"] = featureflag.On.String()
				}
				if _, configured := cfg.Features["web_leaderboard"]; !configured && *leaderboard {
					cfg.Features["web_leaderboard"] = featureflag.On.String()
				}
			}

			flags, err := featureflag.New(cfg.Features)
			if err != nil {
				return fmt.Errorf("web: resolve feature flags: %w", err)
			}
			if !flags.Enabled("web_enabled") {
				return nil
			}
			leaderboardEnabled := flags.Enabled("web_leaderboard")

			if *preview {
				cfg.HTTPServer.PublicPaths = append(
					cfg.HTTPServer.PublicPaths, "/preview",
				)
			}

			var client *leaderboardstat.HTTPClient
			if leaderboardEnabled {
				client, err = leaderboardstat.NewHTTPClient(
					cfg.LeaderboardAPI.BaseURL,
					cfg.LeaderboardAPI.Timeout,
				)
				if err != nil {
					return fmt.Errorf("web: create leaderboard client: %w", err)
				}

				cfg.HTTPServer.PublicPaths = append(
					cfg.HTTPServer.PublicPaths, "/leaderboard",
				)
			}

			server, err := newServer(cfg.HTTPServer)
			if err != nil {
				return err
			}

			registerLeaderboardPreview(server.GetRouter(), *preview)
			registerLeaderboard(
				server.GetRouter(), client, leaderboardEnabled,
			)

			return serve(cmd.Context(), server)
		},
	})

	return root
}
