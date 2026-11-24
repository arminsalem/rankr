package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeWebFlags(t *testing.T) {
	// A bad port identifies server creation without opening a socket. A bad
	// API URL identifies leaderboard client creation before server creation.
	const config = `
http_server:
  port: 0
leaderboard_api:
  base_url: not-a-url
  timeout: 1s
`
	const invalidPort = "web: create server: invalid port: 0"
	const invalidURL = "web: create leaderboard client:"
	tests := []struct {
		name     string
		features string
		env      map[string]string
		args     []string
		want     string
	}{
		{name: "both off by default"},
		{name: "project alone", features: "  web_project: on\n"},
		{
			name:     "project starts only with web enabled",
			features: "  web_enabled: on\n  web_project: on\n",
			want:     "web: create project client:",
		},
		{
			name:     "project in shadow",
			features: "  web_enabled: on\n  web_project: shadow\n",
			want:     invalidPort,
		},
		{name: "leaderboard alone", features: "  web_leaderboard: on\n"},
		{name: "web in shadow", features: "  web_enabled: shadow\n", args: []string{"--preview"}},
		{name: "web only", features: "  web_enabled: on\n", want: invalidPort},
		{
			name: "leaderboard in shadow", features: "  web_enabled: on\n  web_leaderboard: shadow\n",
			want: invalidPort,
		},
		{
			name: "web and leaderboard", features: "  web_enabled: on\n  web_leaderboard: on\n",
			want: invalidURL,
		},
		{
			name:     "environment enables web and leaderboard",
			features: "  web_enabled: off\n  web_leaderboard: off\n",
			env:      map[string]string{"RANKR_FEATURE_WEB_ENABLED": "on", "RANKR_FEATURE_WEB_LEADERBOARD": "on"},
			want:     invalidURL,
		},
		{
			name:     "environment disables leaderboard",
			features: "  web_enabled: on\n  web_leaderboard: on\n",
			env:      map[string]string{"RANKR_FEATURE_WEB_LEADERBOARD": "off"},
			args:     []string{"--leaderboard"},
			want:     invalidPort,
		},
		{
			name:     "explicit web off overrides preview and leaderboard switches",
			features: "  web_enabled: on\n  web_leaderboard: on\n",
			env:      map[string]string{"RANKR_FEATURE_WEB_ENABLED": "off"},
			args:     []string{"--preview", "--leaderboard"},
		},
		{
			name:     "configured leaderboard stays on without CLI opt-in",
			features: "  web_enabled: on\n  web_leaderboard: on\n",
			args:     []string{"--leaderboard=false"},
			want:     invalidURL,
		},
		{
			name:     "invalid flag value",
			features: "  web_enabled: standby\n",
			want:     `web: resolve feature flags: featureflag: features.web_enabled: invalid value "standby"`,
		},
		{name: "legacy preview", args: []string{"--preview"}, want: invalidPort},
		{name: "legacy leaderboard", args: []string{"--leaderboard"}, want: invalidURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := config
			if tt.features != "" {
				input += "features:\n" + tt.features
			}
			configPath := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(configPath, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CONFIG_PATH", configPath)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			root := newRootCommand()
			root.SetArgs(append([]string{"serve"}, tt.args...))
			root.SilenceUsage = true
			root.SilenceErrors = true

			err := root.ExecuteContext(t.Context())
			if tt.want == "" {
				if err != nil {
					t.Fatalf("disabled web should exit successfully, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}
