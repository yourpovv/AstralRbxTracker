// reads .env and the environment (github secrets > .env)
package settings

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	EnvFile         = ".env"
	WebhookKey      = "DISCORD_WEBHOOK_URL"
	minPollInterval = 30 * time.Second
)

type Settings struct {
	DiscordWebhookURL string
	GamesFile         string
	StateFile         string
	PollInterval      time.Duration
	RequestTimeout    time.Duration
}

func Load(envPath string) (Settings, error) {
	values, err := readDotEnv(envPath)
	if err != nil {
		return Settings{}, err
	}
	lookup := func(key, fallback string) string {
		if value, set := os.LookupEnv(key); set {
			return value
		}
		if value, set := values[key]; set {
			return value
		}
		return fallback
	}

	pollInterval, err := seconds(lookup("POLL_INTERVAL_SECONDS", "120"))
	if err != nil {
		return Settings{}, fmt.Errorf("POLL_INTERVAL_SECONDS: %w", err)
	}
	if pollInterval < minPollInterval {
		return Settings{}, fmt.Errorf("POLL_INTERVAL_SECONDS must be at least %d", int(minPollInterval.Seconds()))
	}
	requestTimeout, err := seconds(lookup("REQUEST_TIMEOUT_SECONDS", "10"))
	if err != nil {
		return Settings{}, fmt.Errorf("REQUEST_TIMEOUT_SECONDS: %w", err)
	}

	return Settings{
		DiscordWebhookURL: lookup(WebhookKey, ""),
		GamesFile:         lookup("GAMES_FILE", "games.json"),
		StateFile:         lookup("STATE_FILE", "state.json"),
		PollInterval:      pollInterval,
		RequestTimeout:    requestTimeout,
	}, nil
}

func seconds(value string) (time.Duration, error) {
	count, err := strconv.Atoi(value)
	if err != nil || count <= 0 {
		return 0, fmt.Errorf("%q is not a whole number of seconds", value)
	}
	return time.Duration(count) * time.Second, nil
}
