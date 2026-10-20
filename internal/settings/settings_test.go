package settings

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func envFileWith(t *testing.T, content string) string {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	t.Run("uses defaults when nothing is set", func(t *testing.T) {
		got, err := Load(filepath.Join(t.TempDir(), ".env"))

		if err != nil || got.PollInterval != 2*time.Minute || got.GamesFile != "games.json" {
			t.Fatalf("want defaults, got %+v (%v)", got, err)
		}
	})

	t.Run("lets a real environment variable beat the env file", func(t *testing.T) {
		t.Setenv(WebhookKey, "https://from-env")

		got, err := Load(envFileWith(t, WebhookKey+"=https://from-file\n"))

		if err != nil || got.DiscordWebhookURL != "https://from-env" {
			t.Fatalf("want env value, got %q (%v)", got.DiscordWebhookURL, err)
		}
	})

	t.Run("rejects polling faster than every 30 seconds", func(t *testing.T) {
		if _, err := Load(envFileWith(t, "POLL_INTERVAL_SECONDS=5\n")); err == nil {
			t.Fatal("want an error")
		}
	})
}

func TestSaveWebhook(t *testing.T) {
	t.Run("keeps existing settings in the env file", func(t *testing.T) {
		path := envFileWith(t, "POLL_INTERVAL_SECONDS=60\n")

		if err := SaveWebhook(path, "https://hook"); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)

		if err != nil || got.DiscordWebhookURL != "https://hook" || got.PollInterval != time.Minute {
			t.Fatalf("want both settings, got %+v (%v)", got, err)
		}
	})
}
