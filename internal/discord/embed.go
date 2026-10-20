package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourpovv/AstralRbxTracker/internal/tracker"
)

const (
	updateColor  = 0xE74C3C
	detailsColor = 0x5865F2
	footerText   = "Astral Event Watchdog"
	gameURL      = "https://www.roblox.com/games/%d"
	fieldLimit   = 1024

	noPublicChanges = "Likely a script/map update"
)

type message struct {
	Username string  `json:"username"`
	Embeds   []embed `json:"embeds"`
}

type embed struct {
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Description string     `json:"description"`
	Color       int        `json:"color"`
	Fields      []field    `json:"fields"`
	Thumbnail   *thumbnail `json:"thumbnail,omitempty"`
	Footer      footer     `json:"footer"`
	Timestamp   string     `json:"timestamp"`
}

type field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type thumbnail struct {
	URL string `json:"url"`
}

type footer struct {
	Text string `json:"text"`
}

func newEmbed(update tracker.Update) embed {
	result := headline(update)
	result.URL = fmt.Sprintf(gameURL, update.Game.PlaceID)
	result.Fields = fields(update)
	result.Footer = footer{Text: footerText}
	result.Timestamp = update.DetectedAt.UTC().Format(time.RFC3339)
	if update.Current.IconURL != "" {
		result.Thumbnail = &thumbnail{URL: update.Current.IconURL}
	}
	return result
}

func headline(update tracker.Update) embed {
	name := update.DisplayName()
	if update.Published() {
		return embed{
			Title:       "🚨 [ASTRAL UPDATE DETECTED] " + name,
			Description: fmt.Sprintf("**%s** has just been updated on Roblox!", name),
			Color:       updateColor,
		}
	}
	return embed{
		Title:       "🛠️ [ASTRAL DETAILS CHANGED] " + name,
		Description: fmt.Sprintf("The public details for **%s** changed. No new update was published.", name),
		Color:       detailsColor,
	}
}

func fields(update tracker.Update) []field {
	result := []field{
		{Name: "👥 Active Players", Value: withThousands(update.Current.Playing)},
		{Name: "📦 What Changed", Value: whatChanged(update.Changes)},
	}
	if update.Published() {
		result = append(result,
			field{Name: "📅 Old Timestamp", Value: "`" + update.Previous.Updated + "`"},
			field{Name: "🚀 New Timestamp", Value: "`" + update.Current.Updated + "`"},
		)
	}
	return result
}

func whatChanged(changes []tracker.Change) string {
	if len(changes) == 0 {
		return noPublicChanges
	}
	summaries := make([]string, len(changes))
	for i, change := range changes {
		summaries[i] = change.Summary
	}
	return truncate(strings.Join(summaries, "\n"), fieldLimit)
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit])
}

func withThousands(number int) string {
	digits := fmt.Sprint(number)
	var grouped strings.Builder
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(digit)
	}
	return grouped.String()
}
