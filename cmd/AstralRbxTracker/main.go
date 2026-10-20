package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/discord"
	"github.com/yourpovv/AstralRbxTracker/internal/manage"
	"github.com/yourpovv/AstralRbxTracker/internal/roblox"
	"github.com/yourpovv/AstralRbxTracker/internal/settings"
	"github.com/yourpovv/AstralRbxTracker/internal/store"
	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
	"github.com/yourpovv/AstralRbxTracker/internal/wait"
)

var errNoWebhook = errors.New(settings.WebhookKey + " is not set")

type app struct {
	settings settings.Settings
	manager  manage.Manager
	tracker  *tracker.Tracker
	log      *slog.Logger
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := run(ctx, os.Args[1:])
	if errors.Is(err, context.Canceled) {
		fmt.Println("\nStopped.")
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return runCommand(ctx, args)
	}

	input := bufio.NewScanner(os.Stdin)
	if err := askForWebhook(input); err != nil {
		return err
	}
	app, err := loadApp()
	if err != nil {
		return err
	}
	if err := askForGames(ctx, app, input); err != nil {
		return err
	}
	return trackForever(ctx, app)
}

func loadApp() (app, error) {
	config, err := settings.Load(settings.EnvFile)
	if err != nil {
		return app{}, err
	}

	httpClient := &http.Client{Timeout: config.RequestTimeout}
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	robloxClient := &roblox.Client{
		HTTP:      httpClient,
		Endpoints: roblox.PublicEndpoints,
		Retry:     roblox.DefaultRetry,
		Log:       log,
	}
	games := store.GameFile{Path: config.GamesFile}
	state := store.StateFile{Path: config.StateFile}

	return app{
		settings: config,
		manager:  manage.Manager{Games: games, State: state, Places: robloxClient},
		tracker: &tracker.Tracker{
			Games:    games,
			State:    state,
			Source:   robloxClient,
			Notifier: discord.Webhook{URL: config.DiscordWebhookURL, HTTP: httpClient},
			Log:      log,
			Now:      time.Now,
		},
		log: log,
	}, nil
}

func checkOnce(ctx context.Context, app app) error {
	if app.settings.DiscordWebhookURL == "" {
		return errNoWebhook
	}
	return app.tracker.Check(ctx)
}

func trackForever(ctx context.Context, app app) error {
	if app.settings.DiscordWebhookURL == "" {
		return errNoWebhook
	}
	fmt.Printf("Checking every %s. Ctrl+C to stop.\n\n", app.settings.PollInterval)
	for {
		if err := app.tracker.Check(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			app.log.Error("check failed", "err", err)
		}
		if err := wait.For(ctx, app.settings.PollInterval); err != nil {
			return err
		}
	}
}
