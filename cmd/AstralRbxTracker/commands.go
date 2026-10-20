package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/yourpovv/AstralRbxTracker/internal/manage"
	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

var commands = map[string]command{
	"add":     {"add <place_id> [--name NAME]", "track a game (name is fetched from Roblox)", runAdd},
	"remove":  {"remove <place_id>", "stop tracking a game and forget its state", placeCommand("removed", manage.Manager.Remove)},
	"enable":  {"enable <place_id>", "resume checking a game", placeCommand("enabled", manage.Manager.Enable)},
	"disable": {"disable <place_id>", "pause checking a game", placeCommand("disabled", manage.Manager.Disable)},
	"list":    {"list", "show tracked games", runList},
	"once":    {"once", "check every game one time, for GitHub Actions or cron", runOnce},
	"test":    {"test <place_id>", "send a sample alert with the game's live info", runTest},
}

type command struct {
	usage   string
	summary string
	run     func(ctx context.Context, app app, args []string) error
}

var errUsage = errors.New("run with no command to start tracking, or use one of the commands above")

func runCommand(ctx context.Context, args []string) error {
	chosen, found := commands[args[0]]
	if !found {
		printUsage()
		return errUsage
	}
	app, err := loadApp()
	if err != nil {
		return err
	}
	return chosen.run(ctx, app, args[1:])
}

func runAdd(ctx context.Context, app app, args []string) error {
	flags := flag.NewFlagSet("add", flag.ContinueOnError)
	name := flags.String("name", "", "alert name override (default: live Roblox name)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	placeID, err := parsePlaceID(flags.Args())
	if err != nil {
		return err
	}
	if err := flags.Parse(flags.Args()[1:]); err != nil {
		return err
	}

	if err := app.manager.Add(ctx, placeID, *name); err != nil {
		return err
	}
	fmt.Printf("Tracking place %d. First check saves a baseline.\n", placeID)
	return nil
}

func runList(_ context.Context, app app, _ []string) error {
	return app.manager.List(os.Stdout)
}

func runOnce(ctx context.Context, app app, _ []string) error {
	return checkOnce(ctx, app)
}

func runTest(ctx context.Context, app app, args []string) error {
	if err := requireWebhook(app); err != nil {
		return err
	}
	placeID, err := parsePlaceID(args)
	if err != nil {
		return err
	}
	universeID, err := app.manager.Places.UniverseID(ctx, placeID)
	if err != nil {
		return err
	}

	game := tracker.Game{PlaceID: placeID, UniverseID: universeID}
	if err := app.tracker.SendTest(ctx, game); err != nil {
		return err
	}
	fmt.Printf("Sent a test alert for place %d.\n", placeID)
	return nil
}

func placeCommand(
	result string, apply func(manage.Manager, int64) error,
) func(context.Context, app, []string) error {
	return func(_ context.Context, app app, args []string) error {
		placeID, err := parsePlaceID(args)
		if err != nil {
			return err
		}
		if err := apply(app.manager, placeID); err != nil {
			return err
		}
		fmt.Printf("Place %d: %s\n", placeID, result)
		return nil
	}
}

func parsePlaceID(args []string) (int64, error) {
	if len(args) == 0 {
		return 0, errors.New("missing place ID, the number in roblox.com/games/<place_id>")
	}
	placeID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || placeID <= 0 {
		return 0, fmt.Errorf("%q isn't a place ID, it must be a number", args[0])
	}
	return placeID, nil
}

func printUsage() {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Commands:")
	for _, name := range names {
		fmt.Printf("  %-32s %s\n", commands[name].usage, commands[name].summary)
	}
}
