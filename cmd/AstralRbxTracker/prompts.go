package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/yourpovv/AstralRbxTracker/internal/settings"
)

func askForWebhook(input *bufio.Scanner) error {
	config, err := settings.Load(settings.EnvFile)
	if err != nil {
		return err
	}
	if config.DiscordWebhookURL != "" {
		return nil
	}

	webhookURL, err := ask(input, "Discord webhook URL: ")
	if err != nil {
		return err
	}
	if err := settings.SaveWebhook(settings.EnvFile, webhookURL); err != nil {
		return err
	}
	fmt.Printf("Saved to %s\n\n", settings.EnvFile)
	return nil
}

func askForGames(ctx context.Context, app app, input *bufio.Scanner) error {
	for {
		hasGames, err := app.manager.HasGames()
		if err != nil || hasGames {
			return err
		}

		answer, err := ask(input, "Roblox place ID to track (the number in roblox.com/games/<id>): ")
		if err != nil {
			return err
		}
		placeID, err := strconv.ParseInt(answer, 10, 64)
		if err != nil {
			fmt.Println("Place ID must be a number")
			continue
		}
		if err := app.manager.Add(ctx, placeID, ""); err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Tracking place %d. Add more with: add <place_id>\n\n", placeID)
	}
}

func ask(input *bufio.Scanner, prompt string) (string, error) {
	fmt.Print(prompt)
	if !input.Scan() {
		if err := input.Err(); err != nil {
			return "", err
		}
		return "", io.ErrUnexpectedEOF
	}
	return strings.TrimSpace(input.Text()), nil
}
